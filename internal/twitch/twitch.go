package twitch

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type twitchService struct {
	clientId     string
	clientSecret string
	apiBaseUrl   string
	authBaseUrl  string
	accessToken  accessToken
}

type accessToken struct {
	Value     string `json:"access_token"`
	ExpiresIn uint   `json:"expires_in"`
	Type      string `json:"token_type"`
}

type clipQueryResponse struct {
	Clips []clip `json:"data"`
}

type clip struct {
	Duration     float32 `json:"duration"`
	Url          string  `json:"url"`
	ThumbnailUrl string  `json:"thumbnail_url"`
}

var errUnsupportedThumbnailURL = errors.New("unable to generate direct URL from given thumbnail URL")

func NewService(clientId, clientSecret, authBaseUrl, apiBaseUrl string) (*twitchService, error) {
	token, err := getAccessToken(clientId, clientSecret, authBaseUrl)
	if err != nil {
		return nil, err
	}

	return &twitchService{
		clientId:     clientId,
		clientSecret: clientSecret,
		apiBaseUrl:   apiBaseUrl,
		authBaseUrl:  authBaseUrl,
		accessToken:  token,
	}, nil
}

func getAccessToken(clientId, clientSecret, authBaseUrl string) (accessToken, error) {
	data := url.Values{}
	data.Set("client_id", clientId)
	data.Set("client_secret", clientSecret)
	data.Set("grant_type", "client_credentials")
	authUrl, err := url.JoinPath(authBaseUrl, "oauth2/token")
	if err != nil {
		return accessToken{}, err
	}

	res, err := http.PostForm(authUrl, data)
	if err != nil {
		return accessToken{}, err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		return accessToken{}, err
	}

	var token accessToken
	err = json.NewDecoder(res.Body).Decode(&token)
	if err != nil {
		return accessToken{}, err
	}

	return token, nil
}

func (twitchSvc twitchService) GetClipURLs(broadcasterId, startDate string, count int) ([]string, error) {
	apiUrl, err := url.JoinPath(twitchSvc.apiBaseUrl, "clips")
	if err != nil {
		return nil, err
	}
	client := &http.Client{}

	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", twitchSvc.accessToken.Value))
	req.Header.Add("Client-Id", twitchSvc.clientId)

	query := req.URL.Query()
	query.Add("broadcaster_id", broadcasterId)
	query.Add("started_at", startDate)
	query.Add("first", strconv.Itoa(count))
	req.URL.RawQuery = query.Encode()

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, err
	}

	var clipQueryRes clipQueryResponse
	err = json.NewDecoder(res.Body).Decode(&clipQueryRes)
	if err != nil {
		return nil, err
	}

	var directURLs []string
	for _, clip := range clipQueryRes.Clips {
		if clip.Duration >= 10.0 {
			directURL, err := generateDirectURL(clip.ThumbnailUrl)
			if err != errUnsupportedThumbnailURL {
				directURLs = append(directURLs, directURL)
			}
		}
	}

	return directURLs, nil
}

func generateDirectURL(thumbnailURL string) (string, error) {
	i := strings.LastIndex(thumbnailURL, "-preview")
	if i == -1 {
		return "", errUnsupportedThumbnailURL
	}
	return thumbnailURL[:i] + ".mp4", nil
}
