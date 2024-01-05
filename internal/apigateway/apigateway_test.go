package apigateway_test

import (
	"errors"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/jaaanko/twitch-clip-compilation-tool/internal/apigateway"
)

func TestNewErrorJSONString(t *testing.T) {
	tests := map[string]struct {
		input error
		want  string
	}{
		"simple case": {
			input: errors.New("test error"),
			want:  "{\"error_message\":\"test error\"}",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := apigateway.NewErrorJSONString(tc.input)
			if got != tc.want {
				t.Fatalf("expected: %v, got: %v", tc.want, got)
			}
		})
	}
}

func TestNewResponse(t *testing.T) {
	tests := map[string]struct {
		statusCode int
		json       string
		want       *events.APIGatewayV2HTTPResponse
	}{
		"simple case": {
			statusCode: 202,
			json:       "{\"id\":\"test-123\"}",
			want: &events.APIGatewayV2HTTPResponse{
				StatusCode: 202,
				Headers:    map[string]string{"Content-Type": "application/json"},
				Body:       "{\"id\":\"test-123\"}",
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := apigateway.NewResponse(tc.statusCode, tc.json)
			if got.StatusCode != tc.want.StatusCode {
				t.Fatalf("expected status code: %v, got: %v", tc.want.StatusCode, got.StatusCode)
			}
			if got.Headers["Content-Type"] != tc.want.Headers["Content-Type"] {
				t.Fatalf("expected Content-Type header value: %v, got: %v",
					tc.want.Headers["Content-Type"], got.Headers["Content-Type"],
				)
			}
			if got.Body != tc.want.Body {
				t.Fatalf("expected body: %v, got: %v", got.Body, tc.want.Body)
			}
		})
	}
}
