package twitch_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/jaaanko/twitch-clip-compilation-tool/internal/twitch"
)

func TestGetBroadcasterID_ReturnsOKStatusCode(t *testing.T) {
	authServer := testAuthServer()
	defer authServer.Close()

	testID := "1234"
	testLogin := "test1"
	testDisplayName := "TesT1"

	type result struct {
		id  string
		err error
	}

	response := map[string][]struct {
		ID          string `json:"id"`
		Login       string `json:"login"`
		DisplayName string `json:"display_name"`
	}{
		"data": {
			{
				ID:          testID,
				Login:       testLogin,
				DisplayName: testDisplayName,
			},
		},
	}

	apiSuccessServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username := r.URL.Query().Get("login")
		if username != testLogin {
			w.Write([]byte(`{"data": []}`))
		} else {
			json.NewEncoder(w).Encode(&response)
		}

	}))
	defer apiSuccessServer.Close()

	tests := map[string]struct {
		authServer *httptest.Server
		apiServer  *httptest.Server
		username   string
		want       result
	}{
		"resource server returns user": {
			authServer: authServer,
			apiServer:  apiSuccessServer,
			username:   testLogin,
			want:       result{id: testID, err: nil},
		},
		"resource server cannot find user": {
			authServer: authServer,
			apiServer:  apiSuccessServer,
			username:   "test2",
			want:       result{id: "", err: twitch.ErrUserNotFound},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			twitchSvc, err := twitch.NewService("client_id", "client_secret", tc.authServer.URL, tc.apiServer.URL)
			if err != nil {
				t.Fatal(err)
			}

			id, err := twitchSvc.GetBroadcasterID(tc.username)

			if tc.want.err != err {
				t.Fatalf("expected error: %v, got: %v", tc.want.err, err)
			}

			if tc.want.id != id {
				t.Fatalf("expected id: %v, got: %v", tc.want.id, id)
			}
		})
	}
}

func TestGetBroadcasterID_ReturnsNonOKStatusCode(t *testing.T) {
	authServer := testAuthServer()
	defer authServer.Close()

	apiFailServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer apiFailServer.Close()

	twitchSvc, err := twitch.NewService("client_id", "client_secret", authServer.URL, apiFailServer.URL)
	if err != nil {
		t.Fatal(err)
	}

	id, err := twitchSvc.GetBroadcasterID("test")
	if err == nil {
		t.Fatal("expected an error")
	}

	if id != "" {
		t.Fatalf("expected id to be an empty string, got: %v", id)
	}
}

func TestGetClipURLs(t *testing.T) {
	authServer := testAuthServer()
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
	type result struct {
		urls     []string
		hasError bool
	}

	tests := map[string]struct {
		authServer *httptest.Server
		apiServer  *httptest.Server
		count      int
		want       result
	}{
		"resource server returns a non-successful status code": {
			authServer: authServer,
			apiServer:  apiFailServer,
			count:      2,
			want: result{
				urls:     nil,
				hasError: true,
			},
		},
		"resource server returns successfully": {
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

			if tc.want.hasError != hasError {
				if tc.want.hasError {
					t.Fatal("expected an error")
				} else {
					t.Fatalf("expected no error, got: %v", err)
				}
			}

			if !reflect.DeepEqual(tc.want.urls, urls) {
				t.Fatalf("expected: %v, got: %v", tc.want.urls, urls)
			}

		})
	}
}

func testAuthServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{
			"access_token": "testtoken123",
			"expires_in": 5513382,
			"token_type": "bearer"
		}`))
	}))
}
