package auth

import (
	"reflect"
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
)

func TestAuthPathPayloadAndLines(t *testing.T) {
	auth := config.AuthEntry{ID: "main", Type: config.AuthTypeOAuth}
	paths := core.AuthPaths{
		AuthConfigPath:    "/auth/config.json",
		ProfileConfigPath: "/profile/config.json",
		ClientSecretPath:  "/auth/client.json",
		TokenPath:         "/auth/token.json",
	}

	payload := authPathPayload(auth, paths)
	if payload.ID != "main" || payload.Type != config.AuthTypeOAuth {
		t.Fatalf("payload identity = %#v, want main/oauth", payload)
	}
	if payload.ClientSecretPath != "/auth/client.json" || payload.TokenPath != "/auth/token.json" {
		t.Fatalf("payload paths = %#v, want oauth paths", payload)
	}
	if payload.ServiceAccountPath != "" {
		t.Fatalf("payload includes service account path for oauth: %#v", payload)
	}

	if got, want := authPathLines(auth, paths), []string{"/auth/client.json", "/auth/token.json"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("oauth path lines = %#v, want %#v", got, want)
	}
	service := config.AuthEntry{ID: "svc", Type: config.AuthTypeServiceAccount}
	if got, want := authPathLines(service, core.AuthPaths{ServiceAccountPath: "/auth/service.json"}), []string{"/auth/service.json"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("service account path lines = %#v, want %#v", got, want)
	}
	if got := authPathLines(config.AuthEntry{Type: config.AuthTypeGCloud}, paths); got != nil {
		t.Fatalf("gcloud path lines = %#v, want nil", got)
	}
}

func TestRenderAuthTablePlainText(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	table := renderAuthTable([]config.AuthEntry{
		{ID: "main", Type: config.AuthTypeOAuth, Label: "Main"},
		{ID: "svc", Type: config.AuthTypeServiceAccount, Label: "Service"},
	}, "main", 120)

	for _, want := range []string{"Auth", "main", "oauth", "Main", "✓", "svc", "service-account"} {
		if !strings.Contains(table, want) {
			t.Fatalf("renderAuthTable = %q, want substring %q", table, want)
		}
	}
}

func TestRenderAuthTableCropsFlexibleColumnsAtNarrowWidth(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	output := renderAuthTable([]config.AuthEntry{{
		ID: "main", Type: config.AuthTypeOAuth,
		Label: "A deliberately long identity label", QuotaProjectID: "a-deliberately-long-quota-project-id",
	}}, "main", 64)

	for line := range strings.SplitSeq(output, "\n") {
		if width := len([]rune(line)); width > 64 {
			t.Fatalf("rendered line width = %d, want <= 64: %q", width, line)
		}
	}
	if !strings.Contains(output, "…") {
		t.Fatalf("renderAuthTable = %q, want cropped flexible content", output)
	}
}

func TestNonEmptyStrings(t *testing.T) {
	if got, want := nonEmptyStrings("", "a", "", "b"), []string{"a", "b"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("nonEmptyStrings = %#v, want %#v", got, want)
	}
}
