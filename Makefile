build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap ./cmd/clip-compiler-lambda/main.go
	zip twitch-clip-compiler-lambda.zip bootstrap
	rm bootstrap