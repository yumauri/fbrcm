package firebase

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func TestAuthorizeDesktopClientHonorsCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	ctx = WithOAuthTerminalOutput(ctx, false)
	var events []OAuthAuthorizationEvent
	ctx = WithOAuthAuthorizationObserver(ctx, func(event OAuthAuthorizationEvent) {
		events = append(events, event)
	})
	cfg := &oauth2.Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		Endpoint:     google.Endpoint,
	}

	started := time.Now()
	_, err := authorizeDesktopClient(ctx, cfg, false, false)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("authorizeDesktopClient = %v, want context canceled", err)
	}
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("canceled authorization took %v, want under one second", elapsed)
	}
	if len(events) != 2 {
		t.Fatalf("authorization events = %#v, want start and completion", events)
	}
	if events[0].URL == "" || events[0].Done || events[0].Cancel == nil {
		t.Fatalf("start event = %#v, want authorization URL", events[0])
	}
	if !events[1].Done || !errors.Is(events[1].Err, context.Canceled) {
		t.Fatalf("completion event = %#v, want canceled completion", events[1])
	}
}

func TestAuthorizeDesktopClientRejectsNonInteractiveContextBeforeListening(t *testing.T) {
	ctx := WithOAuthInteractionAllowed(t.Context(), false)
	_, err := authorizeDesktopClient(ctx, &oauth2.Config{}, true, true)
	if !errors.Is(err, ErrOAuthInteractionRequired) {
		t.Fatalf("authorizeDesktopClient = %v, want OAuth interaction required", err)
	}
}

func TestOAuthListenerExistsBeforePresentationAndClosesAfterTimeout(t *testing.T) {
	ctx := WithOAuthTimeout(WithOAuthTerminalOutput(t.Context(), false), 50*time.Millisecond)
	var callback string
	ctx = WithOAuthAuthorizationObserver(ctx, func(event OAuthAuthorizationEvent) {
		if event.Done {
			return
		}
		parsed, err := url.Parse(event.URL)
		if err != nil {
			t.Error(err)
			return
		}
		callback = parsed.Query().Get("redirect_uri")
		if !strings.HasPrefix(callback, "http://127.0.0.1:") {
			t.Errorf("not loopback: %s", callback)
		}
		// A request to a different route proves the listener is already serving
		// without sending a callback that would end authorization.
		listener, err := url.Parse(callback)
		if err != nil {
			t.Error(err)
			return
		}
		listener.Path = "/ready"
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, listener.String(), nil)
		if err != nil {
			t.Error(err)
			return
		}
		response, err := (&http.Client{Timeout: time.Second}).Do(request)
		if err != nil {
			t.Errorf("listener was not ready when URL was presented: %v", err)
			return
		}
		_ = response.Body.Close()
	})
	_, err := authorizeDesktopClient(ctx, &oauth2.Config{ClientID: "test", Endpoint: google.Endpoint}, false, false)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("timeout: %v", err)
	}
	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, callback, nil)
	if err != nil {
		t.Fatal(err)
	}
	response, err := (&http.Client{Timeout: time.Second}).Do(request)
	if response != nil {
		_ = response.Body.Close()
	}
	if err == nil {
		t.Fatal("listener remained open after timeout")
	}
}
