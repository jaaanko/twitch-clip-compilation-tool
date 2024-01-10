package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jaaanko/twitch-clip-compilation-tool/internal/lambdaprocessor"
)

func main() {
	handler := lambdaprocessor.NewHandler()
	lambda.Start(handler)
}
