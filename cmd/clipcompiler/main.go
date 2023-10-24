package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jaaanko/twitch-clip-compilation-tool/internal/downloader"
	"github.com/jaaanko/twitch-clip-compilation-tool/internal/twitch"
	"github.com/joho/godotenv"
)

const outputDir = "out"
const concatListFileName = "list.txt"

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

	fileNames, err := find(outputDir, ".mp4")
	if err != nil {
		log.Fatal(err)
	}

	file, err := os.Create(filepath.Join(outputDir, concatListFileName))
	if err != nil {
		log.Fatalf("failed to create file: %v", err)
	}

	filesModified := makeTimebaseEqual(fileNames)
	err = writeFileNames(filesModified, file)

	if err != nil {
		log.Println(err)
	}

	file.Close()

	remove(fileNames...)
	cmd := exec.Command(
		"ffmpeg", "-y", "-f", "concat", "-i",
		filepath.Join(outputDir, concatListFileName), "-c", "copy", filepath.Join(outputDir, "compilation.mp4"),
	)
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	remove(concatListFileName)
	remove(filesModified...)
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

func makeTimebaseEqual(fileNames []string) []string {
	var filesModified []string
	for _, fileName := range fileNames {
		destFileName := fmt.Sprintf("%v_new.mp4", strings.TrimSuffix(fileName, filepath.Ext(fileName)))
		cmd := exec.Command(
			"ffmpeg", "-i",
			filepath.Join(outputDir, fileName), "-c", "copy",
			"-video_track_timescale", "15360", filepath.Join(outputDir, destFileName),
		)
		if err := cmd.Run(); err != nil {
			fmt.Println(err)
		}
		filesModified = append(filesModified, destFileName)
	}
	return filesModified
}

func remove(fileNames ...string) {
	for _, fileName := range fileNames {
		err := os.Remove(filepath.Join(outputDir, fileName))
		if err != nil {
			fmt.Println(err)
		}
	}
}

func find(root, ext string) ([]string, error) {
	var fileNames []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(d.Name()) == ext {
			fileNames = append(fileNames, d.Name())
		}
		return nil
	})
	return fileNames, err
}
