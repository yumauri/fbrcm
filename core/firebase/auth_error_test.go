package firebase

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

func TestAuthenticationRequestErrorClassifiesOAuthResponses(t *testing.T) {
	tests := []struct {
		name          string
		status        int
		remoteCode    string
		wantKind      string
		wantRetryable bool
		wantAuthorize bool
	}{
		{name: "invalid grant", status: http.StatusBadRequest, remoteCode: "invalid_grant", wantKind: AuthenticationCredentialsInvalid, wantAuthorize: true},
		{name: "invalid client", status: http.StatusUnauthorized, remoteCode: "invalid_client", wantKind: AuthenticationCredentialsInvalid},
		{name: "service unavailable", status: http.StatusServiceUnavailable, remoteCode: "temporarily_unavailable", wantKind: AuthenticationRequestFailed, wantRetryable: true},
		{name: "rate limited", status: http.StatusTooManyRequests, remoteCode: "slow_down", wantKind: AuthenticationRequestFailed, wantRetryable: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			retrieve := &oauth2.RetrieveError{Response: &http.Response{StatusCode: test.status, Header: make(http.Header)}, ErrorCode: test.remoteCode}
			err := authenticationRequestError("oauth", "refresh_token", retrieve)
			var authentication *AuthenticationError
			if !errors.As(err, &authentication) || authentication.Kind != test.wantKind || authentication.HTTPStatus != test.status || authentication.RemoteCode != test.remoteCode || authentication.Retryable != test.wantRetryable {
				t.Fatalf("authentication error = %#v", authentication)
			}
			if got := oauthRefreshRequiresAuthorization(err, true); got != test.wantAuthorize {
				t.Fatalf("requires authorization = %t, want %t", got, test.wantAuthorize)
			}
		})
	}
}

func TestOAuthHTTPClientPreservesTransientRefreshFailure(t *testing.T) {
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = io.WriteString(w, `{"error":"temporarily_unavailable"}`)
	}))
	defer provider.Close()

	dir := t.TempDir()
	secretPath := filepath.Join(dir, "client.json")
	tokenPath := filepath.Join(dir, "token.json")
	secret := map[string]any{"installed": map[string]any{
		"client_id": "client-id", "client_secret": "client-secret", "auth_uri": provider.URL + "/authorize", "token_uri": provider.URL, "redirect_uris": []string{"http://localhost"},
	}}
	secretRaw, err := json.Marshal(secret)
	if err != nil {
		t.Fatal(err)
	}
	tokenRaw, err := json.Marshal(&oauth2.Token{AccessToken: "expired", RefreshToken: "refresh", Expiry: time.Unix(0, 0)})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(secretPath, secretRaw, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tokenPath, tokenRaw, 0o600); err != nil {
		t.Fatal(err)
	}

	ctx := WithOAuthInteractionAllowed(context.Background(), false)
	_, err = oauthHTTPClient(ctx, secretPath, tokenPath, false)
	if errors.Is(err, ErrOAuthInteractionRequired) {
		t.Fatalf("transient refresh failure became interaction required: %v", err)
	}
	var authentication *AuthenticationError
	if !errors.As(err, &authentication) || authentication.Kind != AuthenticationRequestFailed || authentication.HTTPStatus != http.StatusServiceUnavailable || !authentication.Retryable {
		t.Fatalf("oauthHTTPClient error = %#v, %v", authentication, err)
	}
}

func TestCredentialParsersReturnTypedAuthenticationFailures(t *testing.T) {
	path := filepath.Join(t.TempDir(), "credentials.json")
	if err := os.WriteFile(path, []byte(`{"broken":true}`), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name string
		load func() error
	}{
		{name: "oauth", load: func() error {
			_, err := oauthHTTPClient(context.Background(), path, filepath.Join(t.TempDir(), "token.json"), false)
			return err
		}},
		{name: "service account", load: func() error { _, err := serviceAccountHTTPClient(context.Background(), path); return err }},
	} {
		t.Run(test.name, func(t *testing.T) {
			var authentication *AuthenticationError
			if err := test.load(); !errors.As(err, &authentication) || authentication.Kind != AuthenticationCredentialsInvalid {
				t.Fatalf("credential error = %#v, %v", authentication, err)
			}
		})
	}
}
