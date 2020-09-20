package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type Getter interface {
	Get(path string) (*http.Response, error)
}

type Client struct {
	cli Getter
}

func NewClient(client Getter) *Client {
	return &Client{
		cli: client,
	}
}

func (c *Client) GetUserEmailFromAccessToken(accessToken string) (string, error) {
	f, err := url.Parse(fmt.Sprintf("https://www.googleapis.com/oauth2/v2/userinfo?access_token=%s", accessToken))
	if err != nil {
		return "", err
	}

	resp, err := c.cli.Get(f.String())
	if err != nil {
		return "", fmt.Errorf("getting user information: %v", err)
	}

	defer resp.Body.Close()

	var u struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return "", fmt.Errorf("decoding user information from google: %v", err)
	}

	return u.Email, nil
}
