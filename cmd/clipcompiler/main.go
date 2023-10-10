package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"

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
	apiBaseUrl := os.Getenv("API_BASE_URL")

	twitchSvc, err := twitch.NewService(clientId, clientSecret, authBaseUrl, apiBaseUrl)
	if err != nil {
		log.Fatal("Error initializing twitch service")
	}

	fmt.Println("Awaiting input...")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	broadcasterId := scanner.Text()

	scanner.Scan()
	startDate := scanner.Text()

	scanner.Scan()
	count, err := strconv.Atoi(scanner.Text())
	if err != nil {
		log.Fatal("Not a valid integer")
	}

	clipLinks, err := twitchSvc.GetClipURLs(broadcasterId, startDate, count)

	if err != nil {
		log.Fatal("Cannot fetch clips")
	}

	fmt.Println(clipLinks)
}
