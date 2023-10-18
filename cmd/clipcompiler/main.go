package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"

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

	clipLinks, err := twitchSvc.GetClipURLs(broadcasterId, startDate, count)
	if err != nil {
		log.Fatal("Cannot fetch clips")
	}

	fmt.Println(clipLinks)
	var wg sync.WaitGroup
	var paths []string

	err = os.MkdirAll(outputDir, 0750)
	if err != nil {
		log.Fatal(err)
	}

	for i, link := range clipLinks {
		path := fmt.Sprintf("clip%d.mp4", i)
		wg.Add(1)
		go func(path, link string) {
			defer wg.Done()
			// TODO: better error handling
			saveClip(path, link)
			paths = append(paths, path)
		}(path, link)
	}

	wg.Wait()

	file, err := os.Create(filepath.Join(outputDir, "list.txt"))
	if err != nil {
		log.Fatalf("failed to create file: %v", err)
	}

	defer file.Close()
	err = writeFileNames(paths, file)
	if err != nil {
		log.Println(err)
	}

}

func saveClip(path, link string) {
	res, err := http.Get(link)
	if err != nil {
		log.Println(err)
		return
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Printf("expected 200, got: %v", res.Status)
		return
	}

	file, err := os.Create(filepath.Join(outputDir, path))
	if err != nil {
		log.Printf("failed to create file: %v", err)
		return
	}

	defer file.Close()

	_, err = io.Copy(file, res.Body)
	if err != nil {
		log.Println(err)
		return
	}
}

func writeFileNames(fileNames []string, dest io.Writer) error {
	var errFailedToWrite error
	for _, fileName := range fileNames {
		_, err := dest.Write([]byte(fmt.Sprintf("file '%v'\n", fileName)))
		if err != nil {
			errFailedToWrite = errors.Join(errFailedToWrite, err)
		}
	}
	return errFailedToWrite
}
