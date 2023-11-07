package downloader_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/jaaanko/twitch-clip-compilation-tool/internal/downloader"
)

func TestRun(t *testing.T) {
	clip1 := "example1.mp4"
	clip2 := "example2.mp4"
	partialFailServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if path.Base(r.URL.Path) == clip1 {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.Write([]byte("clip data"))
		}
	}))
	defer partialFailServer.Close()

	fullFailServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer fullFailServer.Close()

	successServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("clip data"))
	}))
	defer successServer.Close()

	type result struct {
		downloaded map[string]bool
		hasError   bool
	}

	tests := map[string]struct {
		server *httptest.Server
		want   result
	}{
		"download partial fail": {
			server: partialFailServer,
			want:   result{downloaded: map[string]bool{clip2: true}, hasError: true},
		},
		"download full fail": {
			server: fullFailServer,
			want:   result{downloaded: map[string]bool{}, hasError: true}},
		"download success": {
			server: successServer,
			want:   result{downloaded: map[string]bool{clip1: true, clip2: true}, hasError: false},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			clipURL1, _ := url.JoinPath(tc.server.URL, clip1)
			clipURL2, _ := url.JoinPath(tc.server.URL, clip2)
			urls := []string{clipURL1, clipURL2}
			tempDir := t.TempDir()

			downloaded, errDownload := downloader.Run(tempDir, urls)
			hasError := errDownload != nil

			fileNames := map[string]bool{}
			for _, path := range downloaded {
				fileNames[filepath.Base(path)] = true
			}

			got := result{downloaded: fileNames, hasError: hasError}
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
