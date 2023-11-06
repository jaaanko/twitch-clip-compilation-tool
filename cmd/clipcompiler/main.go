package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/jaaanko/twitch-clip-compilation-tool/internal/compiler"
	"github.com/jaaanko/twitch-clip-compilation-tool/internal/downloader"
	"github.com/jaaanko/twitch-clip-compilation-tool/internal/twitch"
	"github.com/joho/godotenv"
)

const outputDir = "out"

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	clientId := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	authBaseURL := os.Getenv("AUTH_BASE_URL")
	apiBaseURL := os.Getenv("API_BASE_URL")

	twitchSvc, err := twitch.NewService(clientId, clientSecret, authBaseURL, apiBaseURL)
	if err != nil {
		log.Fatal("Error initializing twitch service")
	}

	scanner := bufio.NewScanner(os.Stdin)
	var broadcasterId string
	usernamePrompt := "Enter username of streamer: "

	for fmt.Print(usernamePrompt); scanner.Scan(); fmt.Print(usernamePrompt) {
		broadcasterId, err = twitchSvc.GetBroadcasterID(scanner.Text())
		if err == twitch.ErrUserNotFound {
			fmt.Printf("Error: %v", err)
		} else if err != nil {
			log.Fatal(err)
		} else {
			break
		}
	}

	fmt.Print("Enter start date (example: 2023-04-26): ")
	scanner.Scan()
	date, err := time.Parse("2006-01-02", scanner.Text())
	if err != nil {
		log.Fatal("error parsing date")
	}
	startDate := date.Format(time.RFC3339)

	fmt.Print("Enter number of clips to fetch (max of 20 will be retrieved): ")
	scanner.Scan()
	count, err := strconv.Atoi(scanner.Text())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print("Downloading clips...")

	urls, err := twitchSvc.GetClipURLs(broadcasterId, startDate, count)
	if err != nil {
		log.Fatalf("cannot fetch clips: %v", err)
	}

	downloadedClips, err := downloader.Run(outputDir, urls)
	if errors.Is(err, downloader.ErrMkdir) {
		log.Fatal(err)
	} else if err != nil {
		fmt.Println(err)
	}

	fmt.Print("Compiling downloaded clips...")

	compiler := compiler.New(outputDir, "compilation.mp4")
	if err = compiler.Run(downloadedClips); err != nil {
		log.Fatal(err)
	}
}
