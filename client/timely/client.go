package timely

import (
	"fmt"
	"time"

	"github.com/sporadisk/clocker/client"
	"golang.org/x/oauth2"
)

type Client struct {
	// Configuration
	Endpoint      string
	ApplicationID string
	ClientSecret  string
	CallbackURL   string

	// State
	HttpClient *client.HttpClient
	token      *oauth2.Token
	accountID  int64
}

func (c *Client) Init() error {
	c.HttpClient = client.NewHttpClient(10 * time.Second)
	if c.Endpoint == "" {
		c.Endpoint = "https://api.timelyapp.com/1.1/"
	}
	return nil
}

func (c *Client) LoadToken(token *oauth2.Token) {
	c.token = token
}

func (c *Client) GetToken() *oauth2.Token {
	return c.token
}

func (c *Client) HasToken() bool {
	return c.token != nil && c.token.Valid()
}

func (c *Client) verifyToken() (bool, error) {
	if !c.HasToken() {
		return false, nil
	}

	if c.needTokenRefresh(time.Now()) {
		err := c.refreshToken()
		if err != nil {
			return false, fmt.Errorf("failed to refresh token: %w", err)
		}
	}

	// Assume the token is valid if we got this far.
	// We'll find out soon enough if it's not.
	return true, nil
}

// prep needs to be run prior to making API calls, in order to verify that the
// token is valid and up to date.
func (c *Client) prep() error {
	ok, err := c.verifyToken()
	if err != nil {
		return fmt.Errorf("c.verifyToken(): %w", err)
	}

	if !ok {
		return fmt.Errorf("no valid token available")
	}

	return nil
}

func (c *Client) refreshToken() error {
	// TODO: implement token refresh
	return nil
}

func (c *Client) needTokenRefresh(now time.Time) bool {
	if c.token == nil {
		return true
	}
	return c.token.Expiry.Before(now.Add(168 * time.Hour)) // token expires in less than 7 days
}
