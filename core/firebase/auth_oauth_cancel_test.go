package firebase

import (
	"context"
	"errors"
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
