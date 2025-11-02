package timely

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"golang.org/x/oauth2"
)

// LocalCallback is the OAuth2 callback URL for local applications
const LocalCallback = "urn:ietf:wg:oauth:2.0:oob"

func (c *Client) setupOAuthConfig() error {

	if c.CallbackURL != LocalCallback {
		return fmt.Errorf("the local callback is currently the only supported callback URL")
	}

	c.oauthConfig = &oauth2.Config{
		ClientID:     c.ApplicationID,
		ClientSecret: c.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  c.timelyEndpoint("1.1", "oauth/authorize"),
			TokenURL: c.timelyEndpoint("1.1", "oauth/token"),
		},
		RedirectURL: c.CallbackURL,
	}
	return nil
}

func (c *Client) getToken(ctx context.Context) error {
	// option 1: load from storage
	err := c.loadTokenFromFile()
	if err != nil {
		return fmt.Errorf("loadToken: %w", err)
	}

	// option 2: no token stored yet - need to authenticate
	if c.token == nil {
		err = c.oAuthExchange(ctx)
		if err != nil {
			return fmt.Errorf("oAuthExchange: %w", err)
		}
	}

	if c.needTokenRefresh(time.Now()) {
		err = c.refreshToken()
		if err != nil {
			return fmt.Errorf("refreshToken: %w", err)
		}
	}

	return nil
}

func (c *Client) oAuthExchange(ctx context.Context) error {
	state := "placeholder" // TODO: generate and store a state string
	authURL := c.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	fmt.Printf("Please visit this auth URL: %s\n", authURL)

	var authCode string
	fmt.Print("Enter the authorization code: ")
	_, err := fmt.Scanln(&authCode)
	if err != nil {
		return fmt.Errorf("fmt.Scanln: %w", err)
	}
	fmt.Println("Attempting exchange..")

	token, err := c.oauthConfig.Exchange(ctx, authCode)
	if err != nil {
		return fmt.Errorf("oauthConfig.Exchange: %w", err)
	}
	fmt.Println("Success!")

	c.token = token

	err = c.storeTokenToFile()
	if err != nil {
		return fmt.Errorf("storeTokenToFile: %w", err)
	}

	return nil
}

func (c *Client) loadTokenFromFile() error {
	path, err := c.getTokenStoragePath()
	if err != nil {
		return fmt.Errorf("getTokenStoragePath: %w", err)
	}

	finfo, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		// No token stored yet
		return nil
	}

	if err != nil {
		return fmt.Errorf("os.Stat: %w", err)
	}

	if finfo.IsDir() {
		return fmt.Errorf("token storage path %s is a directory", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("os.ReadFile: %w", err)
	}

	var token oauth2.Token
	err = json.Unmarshal(data, &token)
	if err != nil {
		return fmt.Errorf("json.Unmarshal: %w", err)
	}

	c.token = &token
	fmt.Println("Timely oAuth token loaded from local storage.")
	return nil
}

func (c *Client) storeTokenToFile() error {
	if c.token == nil {
		return fmt.Errorf("no token to store")
	}

	path, err := c.getTokenStoragePath()
	if err != nil {
		return fmt.Errorf("getTokenStoragePath: %w", err)
	}

	data, err := json.Marshal(c.token)
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}

	err = os.WriteFile(path, data, 0600)
	if err != nil {
		return fmt.Errorf("os.WriteFile: %w", err)
	}
	return nil
}

func (c *Client) getTokenStoragePath() (string, error) {
	tokenStorageDir, err := c.ensureTokenStorageDir()
	if err != nil {
		return "", fmt.Errorf("ensureTokenStorageDir: %w", err)
	}

	tokenStoragePath := tokenStorageDir + "/timely_token.json"
	return tokenStoragePath, nil
}

func (c *Client) ensureTokenStorageDir() (string, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("os.UserHomeDir: %w", err)
	}
	tokenStorageDir := homedir + "/.clocker"

	finfo, err := os.Stat(tokenStorageDir)
	if err == nil && !finfo.IsDir() {
		return "", fmt.Errorf("token storage path %s is not a directory", tokenStorageDir)
	}

	if err != nil && os.IsNotExist(err) {
		err = os.Mkdir(tokenStorageDir, 0700)
		if err != nil {
			return "", fmt.Errorf("os.Mkdir(%s): %w", tokenStorageDir, err)
		}
	}

	if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("os.Stat: %w", err)
	}

	return tokenStorageDir, nil
}

func (c *Client) refreshToken() error {
	// TODO: implement token refresh
	// Timely tends to return everlasting tokens, so this is not urgent.
	return nil
}

func (c *Client) needTokenRefresh(now time.Time) bool {
	if c.token == nil {
		return true
	}

	if c.token.Expiry.IsZero() {
		return false // token never expires
	}

	return c.token.Expiry.Before(now.Add(168 * time.Hour)) // token expires in less than 7 days
}
