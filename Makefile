build:
	export GOOS=linux 
	export GOARCH=amd64 
	export CGO_ENABLED=0 
	go build -o bootstrap ./cmd/lambda-processor/main.go
	zip lambda-processor.zip bootstrap
	go build -o bootstrap ./cmd/lambda-api/main.go
	zip lambda-api.zip bootstrap
	rm bootstrap