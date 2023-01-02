package gdocs

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/docs/v1"
)

const localCredentialPath = "credentials.json"

func WatchGdoc(docID, credentialPath string) {
	if credentialPath == "" {
		credentialPath = localCredentialPath
	}

	ctx := context.Background()

	docsService, err := docs.NewService(ctx)
	if err != nil {
		log.Printf("failed to initialize docs service: %s", err.Error())
		return
	}

	token, err := getToken(ctx, credentialPath)
	if err != nil {
		log.Printf("failed to get oauth2 token: %s", err.Error())
		return
	}

}

func getToken(ctx context.Context, credentialPath string) (*oauth2.Token, error) {

	// try local file first
	b, readErr := os.ReadFile(credentialPath)

	if readErr != nil && !errors.Is(readErr, os.ErrNotExist) {
		// file was found, but could not be read
		return nil, fmt.Errorf("failed to open credential file: %w", readErr)
	}

	if readErr == nil { // file found
		creds, err := google.CredentialsFromJSON(ctx, b)

		if err != nil {
			return nil, fmt.Errorf("failed to parse local credential file: %w", err)
		}

		return creds.TokenSource.Token()
	}

	// local file not found: try default credentials
	creds, err := google.FindDefaultCredentials(ctx)
	if err == nil {
		// success
		return creds.TokenSource.Token()
	}

	if !isDefaultCredentialsNotFoundErr(err) {
		return nil, fmt.Errorf("error while retrieving default credentials: %w", err)
	}

	return nil, fmt.Errorf("no credentials are available.")
}

// this might break at some point, but unfortunately there are no error types or variables to compare against.
func isDefaultCredentialsNotFoundErr(err error) bool {
	return strings.Contains(err.Error(), "google: could not find default credentials")
}
