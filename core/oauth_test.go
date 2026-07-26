package core

import (
	"context"
	"testing"
)

func TestOAuthAuthorizationContextPreservesDefaultBehaviorUntilConfigured(t *testing.T) {
	svc := &Core{}
	ctx := context.Background()

	gotCtx, autoOpen := svc.oauthAuthorizationContext(ctx, "default", true)

	if gotCtx != ctx {
		t.Fatal("unconfigured OAuth context was replaced")
	}
	if !autoOpen {
		t.Fatal("unconfigured OAuth unexpectedly disabled browser opening")
	}
}

func TestOAuthAuthorizationContextUsesHostPresentationSettings(t *testing.T) {
	svc := &Core{}
	svc.ConfigureOAuthAuthorization(false, func(OAuthAuthorizationEvent) {})

	ctx, autoOpen := svc.oauthAuthorizationContext(context.Background(), "work", true)
	if autoOpen {
		t.Fatal("configured OAuth did not disable browser opening")
	}
	if err := ctx.Err(); err != nil {
		t.Fatalf("configured OAuth context error = %v, want a live context", err)
	}
}
