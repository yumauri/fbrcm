package firebase

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"golang.org/x/oauth2"

	coreenv "github.com/yumauri/fbrcm/core/env"
)

func TestNewServiceWithAccessTokenAuthenticatesInMemory(t *testing.T) {
	const (
		accessToken  = "test-access-token"
		quotaProject = "billing-project"
	)
	t.Setenv(coreenv.GoogleCloudQuotaProject, quotaProject)

	svc, err := NewServiceWithAccessToken(context.Background(), accessToken)
	if err != nil {
		t.Fatal(err)
	}
	resilient, ok := svc.httpClient.Transport.(*resilientTransport)
	if !ok {
		t.Fatalf("transport = %T, want resilient transport", svc.httpClient.Transport)
	}
	oauthTransport, ok := resilient.base.(*oauth2.Transport)
	if !ok {
		t.Fatalf("base transport = %T, want OAuth transport", resilient.base)
	}

	requestCount := 0
	oauthTransport.Base = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		requestCount++
		if got := req.Header.Get("Authorization"); got != "Bearer "+accessToken {
			t.Fatalf("Authorization = %q, want bearer token", got)
		}
		if got := req.Header.Get("X-Goog-User-Project"); got != quotaProject {
			t.Fatalf("X-Goog-User-Project = %q, want %q", got, quotaProject)
		}
		if logged := formatHeaders(req.Header); strings.Contains(logged, accessToken) {
			t.Fatalf("formatted headers leaked access token: %q", logged)
		}
		return jsonHTTPResponse(http.StatusOK, `{}`, "etag"), nil
	})

	if _, _, err := svc.GetRemoteConfig(context.Background(), "target-project"); err != nil {
		t.Fatal(err)
	}
	if requestCount != 1 {
		t.Fatalf("request count = %d, want 1", requestCount)
	}
}

func TestNewServiceWithAccessTokenUsesTargetProjectForQuota(t *testing.T) {
	const targetProject = "target-project"
	t.Setenv(coreenv.GoogleCloudQuotaProject, "")

	svc, err := NewServiceWithAccessToken(context.Background(), "test-access-token")
	if err != nil {
		t.Fatal(err)
	}
	resilient, ok := svc.httpClient.Transport.(*resilientTransport)
	if !ok {
		t.Fatalf("transport = %T, want resilient transport", svc.httpClient.Transport)
	}
	oauthTransport, ok := resilient.base.(*oauth2.Transport)
	if !ok {
		t.Fatalf("base transport = %T, want OAuth transport", resilient.base)
	}
	oauthTransport.Base = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if got := req.Header.Get("X-Goog-User-Project"); got != targetProject {
			t.Fatalf("X-Goog-User-Project = %q, want %q", got, targetProject)
		}
		return jsonHTTPResponse(http.StatusOK, `{}`, "etag"), nil
	})

	if _, _, err := svc.GetRemoteConfig(context.Background(), targetProject); err != nil {
		t.Fatal(err)
	}
}

func TestNewServiceWithAccessTokenRejectsInvalidToken(t *testing.T) {
	for _, test := range []struct {
		name  string
		token string
	}{
		{name: "empty"},
		{name: "spaces", token: "   "},
		{name: "embedded space", token: "secret token"},
		{name: "newline", token: "secret\ntoken"},
		{name: "control", token: "secret\x00token"},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewServiceWithAccessToken(context.Background(), test.token)
			var authentication *AuthenticationError
			if !errors.As(err, &authentication) || authentication.Kind != AuthenticationCredentialsInvalid || authentication.AuthType != accessTokenAuthType || authentication.Operation != "load_credentials" {
				t.Fatalf("error = %#v, want typed access-token credential error", err)
			}
			if strings.Contains(err.Error(), "secret") {
				t.Fatalf("error leaked access token: %q", err)
			}
		})
	}
}
