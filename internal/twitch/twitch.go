package twitch

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type twitchService struct {
	clientId     string
	clientSecret string
	apiBaseURL   string
	authBaseURL  string
	accessToken  accessToken
}

type accessToken struct {
	Value     string `json:"access_token"`
	ExpiresIn uint   `json:"expires_in"`
	Type      string `json:"token_type"`
}

type unexpectedStatusCodeError struct {
	expected int
	got      int
}

func (e unexpectedStatusCodeError) Error() string {
	return fmt.Sprintf("expected status code: %v, got: %v", e.expected, e.got)
}

var errCreateDownloadURL = errors.New("unable to create download URL")
var errUserNotFound = errors.New("user does not exist on twitch")

func NewService(clientId, clientSecret, authBaseURL, apiBaseURL string) (*twitchService, error) {
	token, err := getAccessToken(clientId, clientSecret, authBaseURL)
	if err != nil {
		return nil, err
	}

	return &twitchService{
		clientId:     clientId,
		clientSecret: clientSecret,
		apiBaseURL:   apiBaseURL,
		authBaseURL:  authBaseURL,
		accessToken:  token,
	}, nil
}

func getAccessToken(clientId, clientSecret, authBaseURL string) (accessToken, error) {
	data := url.Values{}
	data.Set("client_id", clientId)
	data.Set("client_secret", clientSecret)
	data.Set("grant_type", "client_credentials")
	authURL, err := url.JoinPath(authBaseURL, "oauth2/token")
	if err != nil {
		return accessToken{}, err
	}

	res, err := http.PostForm(authURL, data)
	if err != nil {
		return accessToken{}, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return accessToken{}, unexpectedStatusCodeError{expected: http.StatusOK, got: res.StatusCode}
	}

	var token accessToken
	err = json.NewDecoder(res.Body).Decode(&token)
	if err != nil {
		return accessToken{}, err
	}

	return token, nil
}

func (twitchSvc twitchService) GetClipURLs(broadcasterId, startDate, endDate string, count int) ([]string, error) {
	apiURL, err := url.JoinPath(twitchSvc.apiBaseURL, "clips")
	if err != nil {
		return nil, err
	}
	client := &http.Client{}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", twitchSvc.accessToken.Value))
	req.Header.Add("Client-Id", twitchSvc.clientId)

	query := req.URL.Query()
	query.Add("broadcaster_id", broadcasterId)
	query.Add("started_at", startDate)
	query.Add("ended_at", endDate)
	query.Add("first", strconv.Itoa(count))
	req.URL.RawQuery = query.Encode()

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, unexpectedStatusCodeError{expected: http.StatusOK, got: res.StatusCode}
	}

	clipQueryRes := struct {
		Data []struct {
			ThumbnailURL string `json:"thumbnail_url"`
		} `json:"data"`
	}{}

	err = json.NewDecoder(res.Body).Decode(&clipQueryRes)
	if err != nil {
		return nil, err
	}

	var downloadURLs []string
	for _, clip := range clipQueryRes.Data {
		downloadURL, err := createDownloadURL(clip.ThumbnailURL)
		if !errors.Is(err, errCreateDownloadURL) {
			downloadURLs = append(downloadURLs, downloadURL)
		} else {
			log.Printf("%v: skipping %v", err, clip.ThumbnailURL)
		}
	}

	return downloadURLs, nil
}

func (twitchSvc twitchService) GetBroadcasterID(username string) (string, error) {
	apiURL, err := url.JoinPath(twitchSvc.apiBaseURL, "users")
	if err != nil {
		return "", err
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "nil", err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", twitchSvc.accessToken.Value))
	req.Header.Add("Client-Id", twitchSvc.clientId)

	query := req.URL.Query()
	query.Add("login", username)
	req.URL.RawQuery = query.Encode()

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", unexpectedStatusCodeError{expected: http.StatusOK, got: res.StatusCode}
	}

	userQueryResponse := struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}{}

	err = json.NewDecoder(res.Body).Decode(&userQueryResponse)
	if err != nil {
		return "", err
	}

	if len(userQueryResponse.Data) == 0 {
		return "", errUserNotFound
	}

	return userQueryResponse.Data[0].ID, nil
}

func createDownloadURL(thumbnailURL string) (string, error) {
	i := strings.LastIndex(thumbnailURL, "-preview")
	if i == -1 {
		return "", errCreateDownloadURL
	}
	return thumbnailURL[:i] + ".mp4", nil
}
