.PHONY: deploy-lambda
deploy-lambda:
	export GOOS=linux 
	export GOARCH=amd64 
	export CGO_ENABLED=0 
	go build -o bootstrap ./cmd/lambdaprocessor/main.go
	zip lambda-processor.zip bootstrap
	go build -o bootstrap ./cmd/lambdaapi/main.go
	zip lambda-api.zip bootstrap
	rm bootstrap
	aws lambda update-function-code \
		--no-cli-pager --function-name  ${API_FUNCTION_NAME} \
		--zip-file fileb://./lambda-api.zip
	aws lambda update-function-code \
		--no-cli-pager --function-name  ${PROCESSOR_FUNCTION_NAME} \
		--zip-file fileb://./lambda-processor.zip