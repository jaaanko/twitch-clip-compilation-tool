package twitch

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/jaaanko/twitch-clip-compilation-tool/internal/httpext"
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

var errCreateDownloadURL = errors.New("unable to create download URL")
var errUserNotFound = errors.New("user does not exist on twitch")

func NewService(clientId, clientSecret, authBaseURL, apiBaseURL string) (*twitchService, error) {
	svc := &twitchService{
		clientId:     clientId,
		clientSecret: clientSecret,
		apiBaseURL:   apiBaseURL,
		authBaseURL:  authBaseURL,
	}

	err := svc.refreshToken()
	if err != nil {
		return nil, err
	}

	return svc, nil
}

func (twitchSvc *twitchService) refreshToken() error {
	data := url.Values{}
	data.Set("client_id", twitchSvc.clientId)
	data.Set("client_secret", twitchSvc.clientSecret)
	data.Set("grant_type", "client_credentials")
	authURL, err := url.JoinPath(twitchSvc.authBaseURL, "oauth2/token")
	if err != nil {
		return err
	}

	res, err := http.PostForm(authURL, data)
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
		return fmt.Errorf("unable to get a new access token: %v %v", res.StatusCode, errMsg)
	}

	err = json.NewDecoder(res.Body).Decode(&twitchSvc.accessToken)
	if err != nil {
		return err
	}

	return nil
}

func retryIfTokenExpired(twitchSvc *twitchService) httpext.Decorator {
	return func(c httpext.Client) httpext.Client {
		return httpext.ClientFunc(func(req *http.Request) (*http.Response, error) {
			res, err := c.Do(req)
			if err != nil && res.StatusCode == http.StatusUnauthorized {
				err = twitchSvc.refreshToken()
				if err != nil {
					return nil, err
				}
				req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", twitchSvc.accessToken.Value))
				return c.Do(req)
			}
			return res, err
		})

	}
}

func (twitchSvc *twitchService) GetClipURLs(broadcasterId, startDate, endDate string, count int) ([]string, error) {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, err
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, err
	}
	end = end.Add(time.Hour*time.Duration(23) +
		time.Minute*time.Duration(59) +
		time.Second*time.Duration(59))

	apiURL, err := url.JoinPath(twitchSvc.apiBaseURL, "clips")
	if err != nil {
		return nil, err
	}
	client := httpext.Decorate(&http.Client{}, retryIfTokenExpired(twitchSvc))

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", twitchSvc.accessToken.Value))
	req.Header.Add("Client-Id", twitchSvc.clientId)

	query := req.URL.Query()
	query.Add("broadcaster_id", broadcasterId)
	query.Add("started_at", start.Format(time.RFC3339))
	query.Add("ended_at", end.Format(time.RFC3339))
	query.Add("first", strconv.Itoa(count))
	req.URL.RawQuery = query.Encode()

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, err := io.ReadAll(res.Body)
		var errMsg string
		if err != nil {
			errMsg = err.Error()
		} else {
			errMsg = string(body)
		}
		return nil, fmt.Errorf("unable to get clips: %v %v", res.StatusCode, errMsg)
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

func (twitchSvc *twitchService) GetBroadcasterID(username string) (string, error) {
	apiURL, err := url.JoinPath(twitchSvc.apiBaseURL, "users")
	if err != nil {
		return "", err
	}

	client := httpext.Decorate(&http.Client{}, retryIfTokenExpired(twitchSvc))
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
		body, err := io.ReadAll(res.Body)
		var errMsg string
		if err != nil {
			errMsg = err.Error()
		} else {
			errMsg = string(body)
		}
		return "", fmt.Errorf("unable to get user information: %v %v", res.StatusCode, errMsg)
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
