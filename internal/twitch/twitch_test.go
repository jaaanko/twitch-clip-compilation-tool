package twitch_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/jaaanko/twitch-clip-compilation-tool/internal/twitch"
)

type result struct {
	urls     []string
	hasError bool
}

func TestGetClipURLs(t *testing.T) {
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{
			"access_token": "testtoken123",
			"expires_in": 5513382,
			"token_type": "bearer"
		}`))
	}))
	defer authServer.Close()

	apiFailServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer apiFailServer.Close()

	response := map[string][]struct {
		ID                  string  `json:"id"`
		URL                 string  `json:"url"`
		Duration            float32 `json:"duration"`
		ThumbnailURL        string  `json:"thumbnail_url"`
		expectedDownloadURL string
	}{
		"data": {
			{
				ID:                  "testClipID1",
				URL:                 "https://clips.twitch.tv/testClipID1",
				Duration:            25.0,
				ThumbnailURL:        "https://clips-media-assets2.twitch.tv/12345-offset-20320-preview-480x272.jpg",
				expectedDownloadURL: "https://clips-media-assets2.twitch.tv/12345-offset-20320.mp4",
			},
			{
				ID:                  "testClipID2",
				URL:                 "https://clips.twitch.tv/testClipID2",
				Duration:            10.0,
				ThumbnailURL:        "https://clips-media-assets2.twitch.tv/6789-offset-41256-preview-480x272.jpg",
				expectedDownloadURL: "https://clips-media-assets2.twitch.tv/6789-offset-41256.mp4",
			},
		},
	}

	apiSuccessServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(&response)
	}))
	defer apiSuccessServer.Close()

	successURLs := []string{response["data"][0].expectedDownloadURL, response["data"][1].expectedDownloadURL}
	tests := map[string]struct {
		authServer *httptest.Server
		apiServer  *httptest.Server
		count      int
		want       result
	}{
		"resource server returns error": {
			authServer: authServer,
			apiServer:  apiFailServer,
			count:      2,
			want: result{
				urls:     nil,
				hasError: true,
			},
		},
		"resource server returns normally": {
			authServer: authServer,
			apiServer:  apiSuccessServer,
			count:      2,
			want: result{
				urls:     successURLs,
				hasError: false,
			},
		},
		"resource server has less clips than requested": {
			authServer: authServer,
			apiServer:  apiSuccessServer,
			count:      3,
			want: result{
				urls:     successURLs,
				hasError: false,
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			twitchSvc, err := twitch.NewService("client_id", "client_secret", tc.authServer.URL, tc.apiServer.URL)
			if err != nil {
				t.Fatal(err)
			}

			urls, err := twitchSvc.GetClipURLs("0", "2023-10-05T00:00:00Z", tc.count)
			hasError := err != nil
			got := result{urls: urls, hasError: hasError}

			if !reflect.DeepEqual(tc.want, got) {
				if tc.want.hasError != got.hasError {
					t.Fatalf("expected: %#v, got: %#v, error: %v", tc.want, got, err)
				} else {
					t.Fatalf("expected: %#v, got: %#v", tc.want, got)
				}
			}
		})
	}
}
