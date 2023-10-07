package twitch

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type twitchService struct {
	clientId     string
	clientSecret string
	authBaseUrl  string
	accessToken  accessToken
}

type accessToken struct {
	Value     string `json:"access_token"`
	ExpiresIn uint   `json:"expires_in"`
	Type      string `json:"token_type"`
}

const authBaseUrl = ""

func NewService(clientId, clientSecret, authBaseUrl string) (*twitchService, error) {
	token, err := getAccessToken(clientId, clientSecret, authBaseUrl)
	if err != nil {
		return nil, err
	}

	return &twitchService{
		clientId:     clientId,
		clientSecret: clientSecret,
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
		fmt.Printf("Status code not 200, but %v\n", res.StatusCode)
		return accessToken{}, err
	}

	var token accessToken
	err = json.NewDecoder(res.Body).Decode(&token)
	if err != nil {
		return accessToken{}, err
	}

	return token, nil
}
