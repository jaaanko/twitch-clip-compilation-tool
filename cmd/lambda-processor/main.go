package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/jaaanko/twitch-clip-compilation-tool/internal/compiler"
	"github.com/jaaanko/twitch-clip-compilation-tool/internal/downloader"
	"github.com/jaaanko/twitch-clip-compilation-tool/internal/twitch"
)

type request struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Start    string `json:"start"`
	End      string `json:"end"`
	Count    int    `json:"count"`
	UserID   string `json:"user_id"`
}

type item struct {
	ID       string  `dynamodbav:"id"`
	URL      *string `dynamodbav:"url"`
	Status   int     `dynamodbav:"status"`
	ErrorMsg *string `dynamodbav:"error_msg"`
}

const (
	outputDir  = "/tmp"
	ffmpegPath = "/opt/ffmpeg"
)

func handle(ctx context.Context, event *events.SQSEvent) error {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return err
	}

	var req request
	err = json.Unmarshal([]byte(event.Records[0].Body), &req)
	if err != nil {
		return err
	}
	tableName := os.Getenv("DYNAMODB_TABLE_NAME")
	dbClient := dynamodb.NewFromConfig(cfg)

	defer func() {
		if err != nil {
			errorMsg := err.Error()
			item := item{ID: req.ID, Status: 1, ErrorMsg: &errorMsg}
			mapItem, marshalErr := attributevalue.MarshalMap(item)
			if marshalErr != nil {
				err = errors.Join(err, marshalErr)
			} else {
				dbClient.PutItem(context.TODO(), &dynamodb.PutItemInput{
					Item:      mapItem,
					TableName: &tableName,
				})
			}
		}
	}()

	clientId := os.Getenv("TWITCH_CLIENT_ID")
	clientSecret := os.Getenv("TWITCH_CLIENT_SECRET")
	authBaseURL := os.Getenv("TWITCH_AUTH_BASE_URL")
	apiBaseURL := os.Getenv("TWITCH_API_BASE_URL")

	twitchSvc, err := twitch.NewService(clientId, clientSecret, authBaseURL, apiBaseURL)
	if err != nil {
		return fmt.Errorf("error initializing twitch service: %v", err)
	}

	urls, err := twitchSvc.GetClipURLs(req.UserID, req.Start, req.End, min(req.Count, 10))
	if err != nil {
		return fmt.Errorf("error fetching clips: %v", err)
	} else if len(urls) == 0 {
		return fmt.Errorf("no clips found within the specified date range")
	}

	downloadedClips, err := downloader.Run(outputDir, urls)
	if errors.Is(err, downloader.ErrCreateOutputDir) {
		return err
	}

	outputFileName := fmt.Sprintf("%v-%v.mp4", req.Username, uuid.New().String())
	compiler := compiler.New(outputDir, outputFileName, ffmpegPath, true)
	if err = compiler.Run(downloadedClips); err != nil {
		return err
	}

	s3Client := s3.NewFromConfig(cfg)
	uploader := manager.NewUploader(s3Client)
	file, err := os.Open(filepath.Join(outputDir, outputFileName))
	if err != nil {
		return err
	}
	defer file.Close()

	bucketName := os.Getenv("DEST_S3_BUCKET_NAME")
	_, err = uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket: &bucketName,
		Key:    &outputFileName,
		Body:   file,
	})
	if err != nil {
		return err
	}

	presignClient := s3.NewPresignClient(s3Client)
	presignedUrl, err := presignClient.PresignGetObject(context.TODO(),
		&s3.GetObjectInput{
			Bucket: &bucketName,
			Key:    &outputFileName,
		},
		s3.WithPresignExpires(time.Hour*1),
	)
	if err != nil {
		return err
	}

	if err := os.Remove(filepath.Join(outputDir, outputFileName)); err != nil {
		return err
	}

	item := item{ID: req.ID, URL: &presignedUrl.URL, Status: 0}
	mapItem, err := attributevalue.MarshalMap(item)
	if err != nil {
		return err
	}
	_, err = dbClient.PutItem(context.TODO(), &dynamodb.PutItemInput{
		Item:      mapItem,
		TableName: &tableName,
	})
	if err != nil {
		return err
	}

	return nil
}

func main() {
	lambda.Start(handle)
}
