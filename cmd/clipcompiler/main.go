package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/jaaanko/twitch-clip-compilation-tool/internal/compiler"
	"github.com/jaaanko/twitch-clip-compilation-tool/internal/downloader"
	"github.com/jaaanko/twitch-clip-compilation-tool/internal/twitch"
)

const usageString = `

Usage: %v [options] username start_date
			 
Arguments

	username      :   Unique twitch username of the user you wish to watch clips from. [required]
	start_date    :   Only clips created between this date and a week after will be fetched. 
	                  YY-MM-DD format (example: 2023-04-26). [required]

Options

	--max	      :   Maximum number of clips to fetch, no more than 20. Default is 10.
	--output-dir  :   Name of the directory where the final .mp4 file and any temporary files will be placed. 
	                  A default folder named "out" will be created in the current directory if not specified.
	--output-file :   Name of the final .mp4 file. Default is "compilation.mp4".
	--help        :   Displays this message and exits the program.

`

func main() {
	clientId := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	authBaseURL := os.Getenv("AUTH_BASE_URL")
	apiBaseURL := os.Getenv("API_BASE_URL")

	programName := filepath.Base(os.Args[0])
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usageString, programName)
	}

	max := flag.Int("max", 10, "")
	outputDir := flag.String("output-dir", "out", "")
	outputFileName := flag.String("output-file", "compilation.mp4", "")
	flag.Parse()
	args := flag.Args()
	var username, startDate string

	switch len(args) {
	case 0:
		log.Fatal("no arguments provided")
	case 1:
		log.Fatal("insufficient arguments provided")
	case 2:
		username = args[0]
		startDate = args[1]
	default:
		log.Fatal("more than 2 arguments provided")
	}

	twitchSvc, err := twitch.NewService(clientId, clientSecret, authBaseURL, apiBaseURL)
	if err != nil {
		log.Fatalf("error initializing twitch service: %v", err)
	}

	broadcasterId, err := twitchSvc.GetBroadcasterID(username)
	if err != nil {
		log.Fatalf("error getting broadcaster id of %v: %v", username, err)
	}

	date, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		log.Fatal(err)
	}
	startDate = date.Format(time.RFC3339)

	fmt.Println("Downloading clips...")

	urls, err := twitchSvc.GetClipURLs(broadcasterId, startDate, *max)
	if err != nil {
		log.Fatalf("error fetching clips: %v", err)
	}

	downloadedClips, err := downloader.Run(*outputDir, urls)
	if errors.Is(err, downloader.ErrMkdir) {
		log.Fatal(err)
	} else if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Compiling downloaded clips...")

	compiler := compiler.New(*outputDir, *outputFileName)
	if err = compiler.Run(downloadedClips); err != nil {
		log.Fatal(err)
	}
}
