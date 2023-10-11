package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

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

	clipLinks, err := twitchSvc.GetClipURLs(broadcasterId, startDate, count)
	// clipLinks, err := twitchSvc.GetClipURLs("0", "2023-10-01T00:00:00Z", 10)
	if err != nil {
		log.Fatal("Cannot fetch clips")
	}

	fmt.Println(clipLinks)
	var wg sync.WaitGroup

	for i, link := range clipLinks {
		path := fmt.Sprintf("clip%d.mp4", i)
		wg.Add(1)
		go func(path, link string) {
			defer wg.Done()
			saveClip(path, link)
		}(path, link)
	}

	wg.Wait()
}

func saveClip(path, link string) {
	res, err := http.Get(link)
	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Fatal("expected 200, got: ", res.Status)
	}

	file, err := os.Create(path)
	if err != nil {
		log.Fatal("failed to create file")
	}

	defer file.Close()

	_, err = io.Copy(file, res.Body)
	if err != nil {
		log.Fatal(err)
	}
}
