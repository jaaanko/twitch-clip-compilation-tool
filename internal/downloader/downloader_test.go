package downloader_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"testing"

	"github.com/jaaanko/twitch-clip-compilation-tool/internal/downloader"
)

type result struct {
	numDownloaded int
	hasError      bool
}

func TestRun(t *testing.T) {
	clip1 := "/example1.mp4"
	clip2 := "/example2.mp4"
	partialFailServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == clip1 {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.Write([]byte("clip data"))
		}
	}))

	fullFailServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	successServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("clip data"))
	}))

	tests := map[string]struct {
		server *httptest.Server
		want   result
	}{
		"download partial fail": {server: partialFailServer, want: result{numDownloaded: 1, hasError: true}},
		"download full fail":    {server: fullFailServer, want: result{numDownloaded: 0, hasError: true}},
		"download success":      {server: successServer, want: result{numDownloaded: 2, hasError: false}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			clipURL1, _ := url.JoinPath(tc.server.URL, clip1)
			clipURL2, _ := url.JoinPath(tc.server.URL, clip2)
			urls := []string{clipURL1, clipURL2}
			tempDir := t.TempDir()

			errDownload := downloader.Run(tempDir, urls)
			hasError := errDownload != nil

			numFiles, err := numberOfFiles(tempDir)
			if err != nil {
				t.Fatal(err)
			}

			got := result{numDownloaded: numFiles, hasError: hasError}
			if !reflect.DeepEqual(tc.want, got) {
				if tc.want.hasError != got.hasError {
					t.Fatalf("expected: %#v, got: %#v, error: %v", tc.want, got, errDownload)
				} else {
					t.Fatalf("expected: %#v, got: %#v", tc.want, got)
				}
			}
		})
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
