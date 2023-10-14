package twitch_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jaaanko/twitch-clip-compilation-tool/internal/twitch"
)

func TestGetClipURLs_DoesNotReturnStatusOK(t *testing.T) {
	authSvr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{
			"access_token": "testtoken123",
			"expires_in": 5513382,
			"token_type": "bearer"
		}`))
	}))
	defer authSvr.Close()

	apiSvr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer apiSvr.Close()

	twitchSvc, err := twitch.NewService("client_id", "client_secret", authSvr.URL, apiSvr.URL)
	if err != nil {
		t.Error(err)
	}

	_, err = twitchSvc.GetClipURLs("0", "2023-10-05T00:00:00Z", 10)
	if err == nil {
		t.Error("expected an error")
	}
}

func TestGetClipURLs_ReturnsStatusOK(t *testing.T) {
	authSvr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{
			"access_token": "testtoken123",
			"expires_in": 5513382,
			"token_type": "bearer"
		}`))

	}))
	defer authSvr.Close()

	testClipID := "testClipID1"
	testURL := fmt.Sprintf("https://clips.twitch.tv/%s", testClipID)
	testDuration := 25
	commonPath := "test1/12345-offset-20320"
	testThumbnailURL := fmt.Sprintf("https://clips-media-assets2.twitch.tv/%s-preview-480x272.jpg", commonPath)

	apiSvr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(fmt.Sprintf(`{
			"data": [
				{
					"id": "%s",
					"url": "%s",
					"duration": %d,
					"thumbnail_url": "%s"
				}
			]
		}`, testClipID, testURL, testDuration, testThumbnailURL)))
	}))
	defer apiSvr.Close()

	twitchSvc, err := twitch.NewService("client_id", "client_secret", authSvr.URL, apiSvr.URL)
	if err != nil {
		t.Error(err)
	}

	clips, err := twitchSvc.GetClipURLs("0", "2023-10-05T00:00:00Z", 10)
	if err != nil {
		t.Error(err)
	}

	if len(clips) == 0 {
		t.Error("no clips returned")
	}

	expected := fmt.Sprintf("https://clips-media-assets2.twitch.tv/%s.mp4", commonPath)
	if clips[0] != expected {
		t.Errorf("expected: %s, got: %s", expected, clips[0])
	}
}
