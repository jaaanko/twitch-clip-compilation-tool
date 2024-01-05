package apigateway

import (
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
)

func NewErrorJSONString(err error) string {
	e := struct {
		ErrMsg string `json:"error_message"`
	}{
		ErrMsg: err.Error(),
	}

	b, _ := json.Marshal(e)
	return string(b)
}

func NewResponse(statusCode int, json string) *events.APIGatewayV2HTTPResponse {
	return &events.APIGatewayV2HTTPResponse{
		StatusCode: statusCode,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       json,
	}
}
