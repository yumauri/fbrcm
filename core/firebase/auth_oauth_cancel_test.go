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
}
