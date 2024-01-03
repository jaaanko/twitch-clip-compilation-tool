package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/google/uuid"
	"github.com/jaaanko/twitch-clip-compilation-tool/internal/twitch"
)

type request struct {
	Username string `json:"username"`
	Start    string `json:"start"`
	End      string `json:"end"`
	Count    int    `json:"count"`
}

type message struct {
	request
	ID     string `json:"id"`
	UserID string `json:"user_id"`
}

type response struct {
	ID string `json:"id"`
}

func handle(ctx context.Context, event *events.APIGatewayV2HTTPRequest) (*response, error) {
	var req request
	err := json.Unmarshal([]byte(event.Body), &req)
	if err != nil {
		return nil, err
	}

	clientId := os.Getenv("TWITCH_CLIENT_ID")
	clientSecret := os.Getenv("TWITCH_CLIENT_SECRET")
	authBaseURL := os.Getenv("TWITCH_AUTH_BASE_URL")
	apiBaseURL := os.Getenv("TWITCH_API_BASE_URL")

	twitchSvc, err := twitch.NewService(clientId, clientSecret, authBaseURL, apiBaseURL)
	if err != nil {
		return nil, fmt.Errorf("error initializing twitch service: %v", err)
	}

	broadcasterId, err := twitchSvc.GetBroadcasterID(req.Username)
	if err != nil {
		return nil, fmt.Errorf("error getting broadcaster id of %v: %v", req.Username, err)
	}

	messageID := fmt.Sprintf("%v-%v", req.Username, uuid.New().String())
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}

	sqsClient := sqs.NewFromConfig(cfg)
	queueName := os.Getenv("SQS_QUEUE_NAME")
	output, err := sqsClient.GetQueueUrl(context.TODO(), &sqs.GetQueueUrlInput{
		QueueName: &queueName,
	})
	if err != nil {
		return nil, err
	}

	msg := message{ID: messageID, UserID: broadcasterId, request: req}
	b, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	body := string(b)
	_, err = sqsClient.SendMessage(
		context.TODO(),
		&sqs.SendMessageInput{
			QueueUrl:    output.QueueUrl,
			MessageBody: &body,
		},
	)
	if err != nil {
		return nil, err
	}

	return &response{ID: messageID}, nil
}

func main() {
	lambda.Start(handle)
}
