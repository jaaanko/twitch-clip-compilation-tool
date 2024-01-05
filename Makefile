deploy-lambda:
	export GOOS=linux 
	export GOARCH=amd64 
	export CGO_ENABLED=0 
	go build -o bootstrap ./cmd/lambda-processor/main.go
	zip lambda-processor.zip bootstrap
	go build -o bootstrap ./cmd/lambda-api/main.go
	zip lambda-api.zip bootstrap
	rm bootstrap
	aws lambda update-function-code \
		--function-name  ${api-function-name} \
		--zip-file fileb://./lambda-api.zip
	aws lambda update-function-code \
		--function-name  ${processor-function-name} \
		--zip-file fileb://./lambda-processor.zip