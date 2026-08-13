package firebase

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

func TestRecoverRejectedOAuthTokenReauthorizesWhenRefreshTokenIsInvalid(t *testing.T) {
	var grantTypes []string
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if err := req.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		grantType := req.Form.Get("grant_type")
		grantTypes = append(grantTypes, grantType)
		w.Header().Set("Content-Type", "application/json")
		if grantType == "refresh_token" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = io.WriteString(w, `{"error":"invalid_grant"}`)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "recovered-access",
			"refresh_token": "recovered-refresh",
			"token_type":    "Bearer",
			"expires_in":    3600,
		})
	}))
	defer provider.Close()

	cfg := &oauth2.Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		Endpoint: oauth2.Endpoint{
			AuthURL:  provider.URL + "/authorize",
			TokenURL: provider.URL,
		},
	}
	var events []OAuthAuthorizationEvent
	ctx := WithOAuthTerminalOutput(context.Background(), false)
	ctx = WithOAuthAuthorizationObserver(ctx, func(event OAuthAuthorizationEvent) {
		events = append(events, event)
		if event.Done {
			return
		}
		authorizationURL, err := url.Parse(event.URL)
		if err != nil {
			t.Errorf("parse authorization URL: %v", err)
			return
		}
		callbackURL, err := url.Parse(authorizationURL.Query().Get("redirect_uri"))
		if err != nil {
			t.Errorf("parse callback URL: %v", err)
			return
		}
		query := callbackURL.Query()
		query.Set("code", "authorization-code")
		query.Set("state", authorizationURL.Query().Get("state"))
		callbackURL.RawQuery = query.Encode()
		callbackRequest, err := http.NewRequestWithContext(ctx, http.MethodGet, callbackURL.String(), nil)
		if err != nil {
			t.Errorf("create callback request: %v", err)
			return
		}
		go func() {
			response, callbackErr := (&http.Client{Timeout: 2 * time.Second}).Do(callbackRequest)
			if callbackErr == nil {
				_ = response.Body.Close()
			}
		}()
	})
	tokenPath := filepath.Join(t.TempDir(), "token.json")
	cached := &oauth2.Token{
		AccessToken:  "broken-access",
		RefreshToken: "broken-refresh",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(time.Hour),
	}

	_, recovered, err := recoverRejectedOAuthToken(ctx, cfg, cached, tokenPath, true, false)
	if err != nil {
		t.Fatalf("recoverRejectedOAuthToken = %v", err)
	}
	if recovered.AccessToken != "recovered-access" || recovered.RefreshToken != "recovered-refresh" {
		t.Fatalf("recovered token = %#v", recovered)
	}
	if len(events) != 2 || events[0].URL == "" || events[0].Cancel == nil || !events[1].Done || events[1].Err != nil {
		t.Fatalf("authorization events = %#v, want successful start/completion", events)
	}
	if len(grantTypes) < 2 || grantTypes[0] != "refresh_token" || grantTypes[len(grantTypes)-1] != "authorization_code" {
		t.Fatalf("grant types = %#v, want refresh followed by authorization code", grantTypes)
	}
	data, err := os.ReadFile(tokenPath)
	if err != nil {
		t.Fatalf("read recovered token: %v", err)
	}
	if !strings.Contains(string(data), "recovered-access") || !strings.Contains(string(data), "recovered-refresh") {
		t.Fatalf("persisted recovered token = %s", data)
	}
}

func TestRecoverRejectedOAuthTokenPreservesTransientRefreshFailure(t *testing.T) {
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = io.WriteString(w, `{"error":"temporarily_unavailable"}`)
	}))
	defer provider.Close()

	cfg := &oauth2.Config{ClientID: "client-id", ClientSecret: "client-secret", Endpoint: oauth2.Endpoint{TokenURL: provider.URL}}
	cached := &oauth2.Token{AccessToken: "expired", RefreshToken: "refresh", Expiry: time.Unix(0, 0)}
	ctx := WithOAuthInteractionAllowed(context.Background(), false)

	_, _, err := recoverRejectedOAuthToken(ctx, cfg, cached, filepath.Join(t.TempDir(), "token.json"), true, false)
	if errors.Is(err, ErrOAuthInteractionRequired) {
		t.Fatalf("transient refresh failure became interaction required: %v", err)
	}
	var authentication *AuthenticationError
	if !errors.As(err, &authentication) || authentication.Kind != AuthenticationRequestFailed || authentication.HTTPStatus != http.StatusServiceUnavailable || !authentication.Retryable {
		t.Fatalf("recoverRejectedOAuthToken error = %#v, %v", authentication, err)
	}
}

func TestOAuthUnauthorizedTransportRecoversAndRetriesOnce(t *testing.T) {
	tokens := newRotatingOAuthTokenSource(
		oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "broken"}),
		&oauth2.Token{AccessToken: "broken"},
	)
	var authorizations []string
	ctx := context.WithValue(
		context.Background(),
		oauth2.HTTPClient,
		&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			authorizations = append(authorizations, req.Header.Get("Authorization"))
			if req.Header.Get("Authorization") == "Bearer broken" {
				return jsonHTTPResponse(http.StatusUnauthorized, `{"error":"invalid token"}`, ""), nil
			}
			return jsonHTTPResponse(http.StatusOK, `{}`, ""), nil
		})},
	)
	recoveries := 0
	client, err := newRecoveringOAuthClient(
		ctx,
		tokens,
		func(*oauth2.Token) (oauth2.TokenSource, *oauth2.Token, error) {
			recoveries++
			token := &oauth2.Token{AccessToken: "recovered"}
			return oauth2.StaticTokenSource(token), token, nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	client = wrapAuthHTTPClient(client)

	response, err := client.Get("https://example.test/config")
	if err != nil {
		t.Fatalf("first request = %v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.StatusCode)
	}
	_ = response.Body.Close()

	response, err = client.Get("https://example.test/config")
	if err != nil {
		t.Fatalf("second request = %v", err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("second status = %d, want 200", response.StatusCode)
	}
	if recoveries != 1 {
		t.Fatalf("recoveries = %d, want 1", recoveries)
	}
	wantAuthorizations := []string{"Bearer broken", "Bearer recovered", "Bearer recovered"}
	if strings.Join(authorizations, ",") != strings.Join(wantAuthorizations, ",") {
		t.Fatalf("authorizations = %#v, want %#v", authorizations, wantAuthorizations)
	}
}

func TestOAuthUnauthorizedTransportDoesNotLoopAfterRetried401(t *testing.T) {
	tokens := newRotatingOAuthTokenSource(
		oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "broken"}),
		&oauth2.Token{AccessToken: "broken"},
	)
	attempts := 0
	base := &oauth2.Transport{
		Source: tokens,
		Base: roundTripFunc(func(*http.Request) (*http.Response, error) {
			attempts++
			return jsonHTTPResponse(http.StatusUnauthorized, `{"error":"still unauthorized"}`, ""), nil
		}),
	}
	recoveries := 0
	transport := newOAuthUnauthorizedTransport(
		base,
		tokens,
		func(*oauth2.Token) (oauth2.TokenSource, *oauth2.Token, error) {
			recoveries++
			token := &oauth2.Token{AccessToken: "recovered"}
			return oauth2.StaticTokenSource(token), token, nil
		},
	)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://example.test/config", nil)
	if err != nil {
		t.Fatal(err)
	}

	response, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip = %v", err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", response.StatusCode)
	}
	if attempts != 2 || recoveries != 1 {
		t.Fatalf("attempts/recoveries = %d/%d, want 2/1", attempts, recoveries)
	}
}

func TestOAuthUnauthorizedTransportDoesNotReplayNonReplayableBody(t *testing.T) {
	tokens := newRotatingOAuthTokenSource(
		oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "broken"}),
		&oauth2.Token{AccessToken: "broken"},
	)
	base := &oauth2.Transport{
		Source: tokens,
		Base: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return jsonHTTPResponse(http.StatusUnauthorized, `{"error":"invalid token"}`, ""), nil
		}),
	}
	recoveries := 0
	transport := newOAuthUnauthorizedTransport(
		base,
		tokens,
		func(*oauth2.Token) (oauth2.TokenSource, *oauth2.Token, error) {
			recoveries++
			return nil, nil, nil
		},
	)
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPut,
		"https://example.test/config",
		io.NopCloser(strings.NewReader(`{}`)),
	)
	if err != nil {
		t.Fatal(err)
	}

	response, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip = %v", err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusUnauthorized || recoveries != 0 {
		t.Fatalf("status/recoveries = %d/%d, want 401/0", response.StatusCode, recoveries)
	}
}
