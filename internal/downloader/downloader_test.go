package downloader_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/jaaanko/twitch-clip-compilation-tool/internal/downloader"
)

func TestRun_ReturnsErrorWhenDownloadPartiallyFails(t *testing.T) {
	clip1 := "/example1.mp4"
	clip2 := "/example2.mp4"
	clipServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == clip1 {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.Write([]byte("clip data"))
		}
	}))

	clipURL1, _ := url.JoinPath(clipServer.URL, clip1)
	clipURL2, _ := url.JoinPath(clipServer.URL, clip2)
	urls := []string{
		clipURL1,
		clipURL2,
	}

	tempDir := t.TempDir()
	if err := downloader.Run(tempDir, urls); err == nil {
		t.Fatalf("expected an error")
	}

	numFiles, err := numberOfFiles(tempDir)
	if err != nil {
		t.Error(err)
	}

	if numFiles != 1 {
		t.Errorf("expected one file to be downloaded, got %v", numFiles)
	}
}

func TestRun_ReturnsErrorWhenDownloadFullyFails(t *testing.T) {
	clip1 := "/example1.mp4"
	clip2 := "/example2.mp4"
	clipServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	clipURL1, _ := url.JoinPath(clipServer.URL, clip1)
	clipURL2, _ := url.JoinPath(clipServer.URL, clip2)
	urls := []string{
		clipURL1,
		clipURL2,
	}

	tempDir := t.TempDir()
	if err := downloader.Run(tempDir, urls); err == nil {
		t.Fatalf("expected an error")
	}

	numFiles, err := numberOfFiles(tempDir)
	if err != nil {
		t.Error(err)
	}

	if numFiles != 0 {
		t.Errorf("expected 0 files to be downloaded, got %v", numFiles)
	}
}

func TestRun_ReturnsNoErrorWhenDownloadSucceeds(t *testing.T) {
	clip1 := "/example1.mp4"
	clip2 := "/example2.mp4"
	clipServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("clip data"))
	}))

	clipURL1, _ := url.JoinPath(clipServer.URL, clip1)
	clipURL2, _ := url.JoinPath(clipServer.URL, clip2)
	urls := []string{
		clipURL1,
		clipURL2,
	}

	tempDir := t.TempDir()
	if err := downloader.Run(tempDir, urls); err != nil {
		t.Fatalf("expected no errors")
	}

	numFiles, err := numberOfFiles(tempDir)
	if err != nil {
		t.Error(err)
	}

	if numFiles != 2 {
		t.Errorf("expected 2 files to be downloaded, got %v", numFiles)
	}
}

func numberOfFiles(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	names, err := f.Readdirnames(0)
	return len(names), err
}
