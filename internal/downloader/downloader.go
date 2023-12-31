package downloader

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sync"
)

var ErrCreateOutputDir = errors.New("failed to create output directory")

func Run(outputPath string, urls []string) ([]string, error) {
	err := os.MkdirAll(outputPath, 0750)
	if err != nil {
		return nil, errors.Join(ErrCreateOutputDir, err)
	}

	var wg sync.WaitGroup
	errs := make(chan error, len(urls))
	paths := make(chan string, len(urls))
	for _, url := range urls {
		path := filepath.Join(outputPath, path.Base(url))
		url := url
		wg.Add(1)
		go func() {
			defer wg.Done()
			err = download(path, url)
			if err != nil {
				errs <- err
			} else {
				paths <- path
			}
		}()
	}

	go func() {
		wg.Wait()
		close(errs)
		close(paths)
	}()

	var joinedErrors error
	for err := range errs {
		joinedErrors = errors.Join(joinedErrors, err)
	}

	var downloaded []string
	for path := range paths {
		downloaded = append(downloaded, path)
	}
	return downloaded, joinedErrors
}

func download(path, url string) error {
	res, err := http.Get(url)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, err := io.ReadAll(res.Body)
		var errMsg string
		if err != nil {
			errMsg = "unable to read response body"
		} else {
			errMsg = string(body)
		}
		return fmt.Errorf("unable to download clip: %v %v", res.StatusCode, errMsg)
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}

	defer file.Close()

	_, err = io.Copy(file, res.Body)
	if err != nil {
		return err
	}

	return nil
}
