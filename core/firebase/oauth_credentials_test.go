package firebase

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/core/config"
)

func TestOAuthClientCredentialsConfig(t *testing.T) {
	credentials := OAuthClientCredentials{ClientID: "client-id", ClientSecret: "client-secret"}
	cfg, err := credentials.config()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ClientID != credentials.ClientID || cfg.ClientSecret != credentials.ClientSecret {
		t.Fatalf("config credentials = %#v", cfg)
	}
	if cfg.Endpoint.AuthURL == "" || cfg.Endpoint.TokenURL == "" || len(cfg.Scopes) != 1 || cfg.Scopes[0] != cloudPlatformScope {
		t.Fatalf("config OAuth settings = %#v", cfg)
	}
}

func TestGoogleServiceRejectsMissingBuiltInCredentialsAsGoogleAuthFailure(t *testing.T) {
	_, err := NewServiceForAuth(context.Background(), config.AuthEntry{ID: "google", Type: config.AuthTypeGoogle}, false)
	var authentication *AuthenticationError
	if !errors.As(err, &authentication) || authentication.AuthType != config.AuthTypeGoogle || authentication.Kind != AuthenticationCredentialsInvalid || authentication.Operation != "load_credentials" {
		t.Fatalf("NewServiceForAuth error = %#v", err)
	}
}

func TestPKCEChallengeMatchesRFC7636Example(t *testing.T) {
	const verifier = "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	const want = "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"
	if got := pkceChallenge(verifier); got != want {
		t.Fatalf("pkceChallenge = %q, want %q", got, want)
	}
}

func TestOAuthClientCredentialsValidationDoesNotLeakPartialValues(t *testing.T) {
	for _, credentials := range []OAuthClientCredentials{
		{ClientID: "semi-private-client-id"},
		{ClientSecret: "semi-private-client-secret"},
		{},
	} {
		err := credentials.Validate()
		if err == nil || !strings.Contains(err.Error(), "unavailable in this build") {
			t.Fatalf("Validate(%+v) = %v", credentials, err)
		}
		if strings.Contains(err.Error(), credentials.ClientID) && credentials.ClientID != "" {
			t.Fatalf("error leaked client ID: %q", err)
		}
		if strings.Contains(err.Error(), credentials.ClientSecret) && credentials.ClientSecret != "" {
			t.Fatalf("error leaked client secret: %q", err)
		}
	}
}
