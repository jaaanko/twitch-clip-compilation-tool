package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

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

	urls, err := twitchSvc.GetClipURLs(broadcasterId, startDate, count)
	if err != nil {
		log.Fatal("Cannot fetch clips")
	}

	err = downloader.Run(outputDir, urls)
	if errors.Is(err, downloader.ErrMkdir) {
		log.Fatal(err)
	} else if err != nil {
		fmt.Println(err)
	}

	compiler := compiler.New(outputDir, outputDir, "compilation.mp4")
	if err = compiler.Run(); err != nil {
		log.Fatal(err)
	}
}
