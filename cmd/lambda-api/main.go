package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/google/uuid"
	"github.com/jaaanko/twitch-clip-compilation-tool/internal/apigateway"
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

func handle(ctx context.Context, event *events.APIGatewayV2HTTPRequest) (*events.APIGatewayV2HTTPResponse, error) {
	var req request
	err := json.Unmarshal([]byte(event.Body), &req)
	if err != nil {
		return apigateway.NewResponse(
			http.StatusBadRequest, apigateway.NewErrorJSONString(err),
		), err
	}

	clientId := os.Getenv("TWITCH_CLIENT_ID")
	clientSecret := os.Getenv("TWITCH_CLIENT_SECRET")
	authBaseURL := os.Getenv("TWITCH_AUTH_BASE_URL")
	apiBaseURL := os.Getenv("TWITCH_API_BASE_URL")

	twitchSvc, err := twitch.NewService(clientId, clientSecret, authBaseURL, apiBaseURL)
	if err != nil {
		err = fmt.Errorf("error initializing twitch service: %w", err)
		return apigateway.NewResponse(
			http.StatusInternalServerError, apigateway.NewErrorJSONString(err),
		), err
	}

	broadcasterId, err := twitchSvc.GetBroadcasterID(req.Username)
	if err != nil {
		err = fmt.Errorf("error getting broadcaster id of %v: %w", req.Username, err)
		return apigateway.NewResponse(
			http.StatusBadRequest, apigateway.NewErrorJSONString(err),
		), err
	}

	messageID := fmt.Sprintf("%v-%v", req.Username, uuid.New().String())
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return apigateway.NewResponse(
			http.StatusInternalServerError, apigateway.NewErrorJSONString(err),
		), err
	}

	sqsClient := sqs.NewFromConfig(cfg)
	queueName := os.Getenv("SQS_QUEUE_NAME")
	output, err := sqsClient.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: &queueName,
	})
	if err != nil {
		return apigateway.NewResponse(
			http.StatusInternalServerError, apigateway.NewErrorJSONString(err),
		), err
	}

	msg := message{ID: messageID, UserID: broadcasterId, request: req}
	b, err := json.Marshal(msg)
	if err != nil {
		return apigateway.NewResponse(
			http.StatusInternalServerError, apigateway.NewErrorJSONString(err),
		), err
	}

	messageBody := string(b)
	_, err = sqsClient.SendMessage(
		ctx,
		&sqs.SendMessageInput{
			QueueUrl:    output.QueueUrl,
			MessageBody: &messageBody,
		},
	)
	if err != nil {
		return apigateway.NewResponse(
			http.StatusInternalServerError, apigateway.NewErrorJSONString(err),
		), err
	}

	resp := response{ID: messageID}
	b, err = json.Marshal(resp)
	if err != nil {
		return apigateway.NewResponse(
			http.StatusInternalServerError, apigateway.NewErrorJSONString(err),
		), err
	}

	return apigateway.NewResponse(http.StatusAccepted, string(b)), nil
}

func main() {
	lambda.Start(handle)
}
