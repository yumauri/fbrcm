package firebase

import (
	"fmt"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// OAuthClientCredentials contains a Desktop OAuth client's build-scoped
// identity. Official builds inject fbrcm's Google client; profile state never
// persists these values.
type OAuthClientCredentials struct {
	ClientID     string
	ClientSecret string
}

// Validate reports whether both required Desktop OAuth values are available.
func (c OAuthClientCredentials) Validate() error {
	if strings.TrimSpace(c.ClientID) == "" || strings.TrimSpace(c.ClientSecret) == "" {
		return fmt.Errorf("built-in Google OAuth client credentials are unavailable in this build")
	}
	return nil
}

func (c OAuthClientCredentials) config() (*oauth2.Config, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &oauth2.Config{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{cloudPlatformScope},
	}, nil
}
