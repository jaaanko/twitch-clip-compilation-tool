package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jaaanko/twitch-clip-compilation-tool/internal/lambdaapi"
)

func main() {
	handler := lambdaapi.NewHandler()
	lambda.Start(handler)
}
