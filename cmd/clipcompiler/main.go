package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jaaanko/twitch-clip-compilation-tool/internal/twitch"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	clientId := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	authBaseUrl := os.Getenv("AUTH_BASE_URL")

	twitchSvc, err := twitch.NewService(clientId, clientSecret, authBaseUrl)
	if err != nil {
		log.Fatal("Error initializing twitch service")
	}

	fmt.Println(twitchSvc)
}
