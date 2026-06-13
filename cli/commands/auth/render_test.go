package auth

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"

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
	if payload["id"] != "main" || payload["type"] != config.AuthTypeOAuth {
		t.Fatalf("payload identity = %#v, want main/oauth", payload)
	}
	if payload["client_secret_path"] != "/auth/client.json" || payload["token_path"] != "/auth/token.json" {
		t.Fatalf("payload paths = %#v, want oauth paths", payload)
	}
	if _, ok := payload["service_account_path"]; ok {
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
	}, "main")

	for _, want := range []string{"Auth", "main", "oauth", "Main", "✓", "svc", "service-account"} {
		if !strings.Contains(table, want) {
			t.Fatalf("renderAuthTable = %q, want substring %q", table, want)
		}
	}
}

func TestReadJSONFileInputFromPath(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "secret-*.json")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err := file.WriteString(`{"client_id":"demo"}`); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close temp file: %v", err)
	}

	got, err := readJSONFileInput(&cobra.Command{}, file.Name(), "client secret")
	if err != nil {
		t.Fatalf("readJSONFileInput returned error: %v", err)
	}
	if string(got) != `{"client_id":"demo"}` {
		t.Fatalf("readJSONFileInput = %q, want temp file content", string(got))
	}
}

func TestNonEmptyStrings(t *testing.T) {
	if got, want := nonEmptyStrings("", "a", "", "b"), []string{"a", "b"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("nonEmptyStrings = %#v, want %#v", got, want)
	}
}
