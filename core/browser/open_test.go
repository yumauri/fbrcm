package browser

import (
	"strings"
	"testing"
)

func TestRedactedURLMasksOAuthClientID(t *testing.T) {
	got := redactedURL("https://accounts.google.com/o/oauth2/auth?client_id=semi-private-client-id&state=secret-state&scope=openid")
	for _, forbidden := range []string{"semi-private-client-id", "secret-state"} {
		if strings.Contains(got, forbidden) {
			t.Fatalf("redacted URL leaked %q: %s", forbidden, got)
		}
	}
	if !strings.Contains(got, "client_id=%5BREDACTED%5D") || !strings.Contains(got, "state=%5BREDACTED%5D") {
		t.Fatalf("redacted URL = %q", got)
	}
}
