package timely

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"time"

	"github.com/sporadisk/clocker/client"
	"golang.org/x/oauth2"
)

type Client struct {
	// Configuration
	ApiURL        string
	ApplicationID string
	ClientSecret  string
	CallbackURL   string
	AccountID     int
	ProjectID     int
	DebugMode     bool // Will occasionally get used for debugging

	// State
	HttpClient  *client.HttpClient
	token       *oauth2.Token
	oauthConfig *oauth2.Config
	apiURL      *url.URL
	labels      []Label
	projects    []Project
	tags        map[string]Label
	user        User
}

func (c *Client) Init(ctx context.Context) error {
	c.HttpClient = client.NewHttpClient(10 * time.Second)
	if c.ApiURL == "" {
		c.ApiURL = "https://api.timelyapp.com/"
	}

	apiURL, err := url.Parse(c.ApiURL)
	if err != nil {
		return fmt.Errorf("invalid ApiURL: %w", err)
	}
	c.apiURL = apiURL

	err = c.setupOAuthConfig()
	if err != nil {
		return fmt.Errorf("setupOAuthConfig: %w", err)
	}

	err = c.getToken(ctx)
	if err != nil {
		return fmt.Errorf("getToken: %w", err)
	}

	err = c.SelectAccount()
	if err != nil {
		return fmt.Errorf("SelectAccount: %w", err)
	}

	fmt.Printf("Timely account ID selected: %d\n", c.AccountID)

	err = c.GetAllLabels()
	if err != nil {
		return fmt.Errorf("GetAllLabels: %w", err)
	}

	err = c.GetProjects()
	if err != nil {
		return fmt.Errorf("GetProjects: %w", err)
	}

	projectList, err := c.ListProjects()
	if err != nil {
		return fmt.Errorf("ListProjects: %w", err)
	}

	if c.ProjectID == 0 {
		fmt.Println("\nprojectId param has not been specified; please select one from the following list:")
		fmt.Println(projectList)
		return fmt.Errorf("projectId parameter is required")
	}

	project, ok := c.GetProjectByID(c.ProjectID)
	if !ok || !project.Active {
		fmt.Println("\nAvailable projects:")
		fmt.Println(projectList)
		return fmt.Errorf("projectId %d not found among available projects", c.ProjectID)
	}

	err = c.mapTags()
	if err != nil {
		return fmt.Errorf("mapTags: %w", err)
	}

	fmt.Printf("Project: %d (%s) / %d tags.\n", project.ID, project.Name, len(c.tags))

	user, err := c.GetCurrentUser()
	if err != nil {
		return fmt.Errorf("GetCurrentUser: %w", err)
	}
	c.user = user
	fmt.Printf("Current user: %d (%s)\n", user.ID, user.Name)

	return nil
}

func (c *Client) timelyEndpoint(version, endpoint string) string {
	u := *c.apiURL // make a copy of the base URL
	u.Path = path.Join(version, endpoint)
	return u.String()
}

func (c *Client) verifyToken() (bool, error) {
	if c.token == nil || !c.token.Valid() {
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
