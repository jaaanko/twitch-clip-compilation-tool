package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/jaaanko/twitch-clip-compilation-tool/internal/compiler"
	"github.com/jaaanko/twitch-clip-compilation-tool/internal/downloader"
	"github.com/jaaanko/twitch-clip-compilation-tool/internal/twitch"
)

type event struct {
	Username  string `json:username`
	StartDate string `json:start_date`
	EndDate   string `json:end_date`
	Count     int    `json:count`
}

type response struct {
	URL string `json:url`
}

const (
	outputDir  = "tmp"
	bucketName = "twitch-compiled-clips"
)

func handle(ctx context.Context, event *event) (*response, error) {
	if event == nil {
		return nil, fmt.Errorf("received nil event")
	}

	clientId := os.Getenv("TWITCH_CLIENT_ID")
	clientSecret := os.Getenv("TWITCH_CLIENT_SECRET")
	authBaseURL := os.Getenv("TWITCH_AUTH_BASE_URL")
	apiBaseURL := os.Getenv("TWITCH_API_BASE_URL")
	awsBucketRegion := os.Getenv("AWS_BUCKET_REGION")
	// awsAccessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	// awsSecret := os.Getenv("AWS_SECRET_ACCESS_KEY")

	twitchSvc, err := twitch.NewService(clientId, clientSecret, authBaseURL, apiBaseURL)
	if err != nil {
		return nil, fmt.Errorf("error initializing twitch service: %v", err)
	}

	broadcasterId, err := twitchSvc.GetBroadcasterID(event.Username)
	if err != nil {
		return nil, fmt.Errorf("error getting broadcaster id of %v: %v", event.Username, err)
	}

	urls, err := twitchSvc.GetClipURLs(broadcasterId, event.StartDate, event.EndDate, min(event.Count, 10))
	if err != nil {
		return nil, fmt.Errorf("error fetching clips: %v", err)
	} else if len(urls) == 0 {
		return nil, fmt.Errorf("no clips found within the specified date range")
	}

	downloadedClips, err := downloader.Run(outputDir, urls)
	if errors.Is(err, downloader.ErrCreateOutputDir) {
		return nil, err
	}

	outputFileName := fmt.Sprintf("%v-%v.mp4", event.Username, uuid.New().String())
	compiler := compiler.New(outputDir, outputFileName, true)
	if err = compiler.Run(downloadedClips); err != nil {
		return nil, err
	}

	//creds := credentials.NewStaticCredentialsProvider(awsAccessKeyID, awsSecret, "")
	cfg := aws.Config{
		Region: *aws.String(awsBucketRegion),
	}

	client := s3.NewFromConfig(cfg)
	uploader := manager.NewUploader(client)
	file, err := os.Open(filepath.Join(outputDir, outputFileName))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	_, err = uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(outputFileName),
		Body:   file,
	})
	if err != nil {
		return nil, err
	}

	presignClient := s3.NewPresignClient(client)
	presignedUrl, err := presignClient.PresignGetObject(context.Background(),
		&s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(outputFileName),
		},
		s3.WithPresignExpires(time.Hour*1),
	)

	if err != nil {
		return nil, err
	}

	return &response{URL: presignedUrl.URL}, nil
}

func main() {
	lambda.Start(handle)
}
