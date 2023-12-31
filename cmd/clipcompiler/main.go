package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/jaaanko/twitch-clip-compilation-tool/internal/compiler"
	"github.com/jaaanko/twitch-clip-compilation-tool/internal/downloader"
	"github.com/jaaanko/twitch-clip-compilation-tool/internal/twitch"
)

const usageString = `

Usage: %v [options] username start_date end_date
			 
Arguments

	username      :   Unique twitch username of the user you wish to watch clips from. [required]
	start_date    :   Start date in YY-MM-DD format (example: 2023-04-26). [required]
	end_date      :   End date in YY-MM-DD format (example: 2023-04-26). [required]
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
	var username, start, end string

	switch len(args) {
	case 0:
		log.Fatal("no arguments provided")
	case 1, 2:
		log.Fatal("insufficient arguments provided")
	case 3:
		username = args[0]
		start = args[1]
		end = args[2]
	default:
		log.Fatal("more than 3 arguments provided")
	}

	twitchSvc, err := twitch.NewService(clientId, clientSecret, authBaseURL, apiBaseURL)
	if err != nil {
		log.Fatalf("error initializing twitch service: %v", err)
	}

	broadcasterId, err := twitchSvc.GetBroadcasterID(username)
	if err != nil {
		log.Fatalf("error getting broadcaster id of %v: %v", username, err)
	}

	fmt.Println("Downloading clips...")

	urls, err := twitchSvc.GetClipURLs(broadcasterId, start, end, *max)
	if err != nil {
		log.Fatalf("error fetching clips: %v", err)
	} else if len(urls) == 0 {
		fmt.Println("No clips found within the specified date range.")
		return
	}

	downloadedClips, err := downloader.Run(*outputDir, urls)
	if errors.Is(err, downloader.ErrCreateOutputDir) {
		log.Fatal(err)
	} else if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Compiling downloaded clips...")

	compiler := compiler.New(*outputDir, *outputFileName, "ffmpeg", true)
	if err = compiler.Run(downloadedClips); err != nil {
		log.Fatal(err)
	}
}
