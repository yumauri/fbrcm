package core

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/core/config"
)

func TestGoogleAuthAvailableRequiresCompleteBuiltInClient(t *testing.T) {
	tests := []struct {
		name         string
		clientID     string
		clientSecret string
		want         bool
	}{
		{name: "missing both"},
		{name: "missing secret", clientID: "client-id"},
		{name: "missing id", clientSecret: "client-secret"},
		{name: "blank values", clientID: "  ", clientSecret: "\t"},
		{name: "complete", clientID: "client-id", clientSecret: "client-secret", want: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			svc, err := NewService(
				context.Background(),
				WithGoogleOAuthClientCredentials(test.clientID, test.clientSecret),
			)
			if err != nil {
				t.Fatal(err)
			}
			if got := svc.GoogleAuthAvailable(); got != test.want {
				t.Fatalf("GoogleAuthAvailable() = %v, want %v", got, test.want)
			}
		})
	}

	var nilService *Core
	if nilService.GoogleAuthAvailable() {
		t.Fatal("nil Core reports Google auth available")
	}
}

func TestLoadRequiredAuthMissingIncludesSetupGuidance(t *testing.T) {
	_ = setupCoreTestEnv(t)

	_, err := loadRequiredAuth()
	if err == nil || !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("loadRequiredAuth missing = %v, want ErrNotExist", err)
	}
	if !strings.HasPrefix(err.Error(), "read auth config:") || !strings.Contains(err.Error(), "\n\n"+authSetupHint) {
		t.Fatalf("loadRequiredAuth missing = %q, want original error followed by setup guidance", err)
	}
}

func TestLoadRequiredAuthEmptyIncludesSetupGuidance(t *testing.T) {
	_ = setupCoreTestEnv(t)
	if err := config.SaveAuth(&config.AuthFile{Version: config.AuthConfigVersion}); err != nil {
		t.Fatalf("SaveAuth empty = %v", err)
	}

	_, err := loadRequiredAuth()
	if err == nil || err.Error() != "no auth identities configured\n\n"+authSetupHint {
		t.Fatalf("loadRequiredAuth empty = %q, want setup guidance", err)
	}
}

func TestLoadRequiredAuthCorruptKeepsOriginalError(t *testing.T) {
	_ = setupCoreTestEnv(t)
	if err := os.WriteFile(config.GetAuthFilePath(), []byte("{"), config.PrivateFileMode); err != nil {
		t.Fatalf("write corrupt auth config = %v", err)
	}

	_, err := loadRequiredAuth()
	if err == nil || !strings.Contains(err.Error(), "decode auth config") {
		t.Fatalf("loadRequiredAuth corrupt = %v, want decode error", err)
	}
	if strings.Contains(err.Error(), authSetupHint) {
		t.Fatalf("loadRequiredAuth corrupt = %q, want no setup guidance", err)
	}
}

func TestListAuthEmpty(t *testing.T) {
	svc := setupCoreTestEnv(t)

	entries, defaultID, err := svc.ListAuth()
	if err != nil {
		t.Fatalf("ListAuth = %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("entries = %+v, want empty", entries)
	}
	if defaultID != "" {
		t.Fatalf("defaultID = %q, want empty", defaultID)
	}
}

func TestAddGCloudAuthAndAuthPaths(t *testing.T) {
	svc := setupCoreTestEnv(t)

	entry, err := svc.AddGCloudAuth("main", "Main GCloud")
	if err != nil {
		t.Fatalf("AddGCloudAuth = %v", err)
	}
	if entry.Type != config.AuthTypeGCloud {
		t.Fatalf("type = %q, want gcloud", entry.Type)
	}

	auth, paths, err := svc.AuthPaths("main")
	if err != nil {
		t.Fatalf("AuthPaths = %v", err)
	}
	if auth.ID != "main" {
		t.Fatalf("auth ID = %q, want main", auth.ID)
	}
	if paths.AuthConfigPath == "" || paths.ProfileConfigPath == "" {
		t.Fatalf("paths = %+v, want config paths set", paths)
	}
	if paths.ClientSecretPath != "" || paths.ServiceAccountPath != "" {
		t.Fatalf("gcloud paths should not include secret files: %+v", paths)
	}
}

func TestAuthQuotaProjectPersistsAndSurvivesCredentialReplacement(t *testing.T) {
	svc := setupCoreTestEnv(t)

	entry, err := svc.AddGCloudAuthWithQuotaProject("main", "Main", "billing-project")
	if err != nil {
		t.Fatal(err)
	}
	if entry.QuotaProjectID != "billing-project" {
		t.Fatalf("quota project = %q", entry.QuotaProjectID)
	}
	entry, err = svc.AddGCloudAuth("main", "Renamed")
	if err != nil {
		t.Fatal(err)
	}
	if entry.QuotaProjectID != "billing-project" {
		t.Fatalf("quota project after replacement = %q", entry.QuotaProjectID)
	}

	entry, previous, changed, err := svc.SetAuthQuotaProject("main", "other-billing-project")
	if err != nil {
		t.Fatal(err)
	}
	if !changed || previous != "billing-project" || entry.QuotaProjectID != "other-billing-project" {
		t.Fatalf("set result = %+v, previous %q, changed %v", entry, previous, changed)
	}
	entry, previous, changed, err = svc.SetAuthQuotaProject("main", "")
	if err != nil {
		t.Fatal(err)
	}
	if !changed || previous != "other-billing-project" || entry.QuotaProjectID != "" {
		t.Fatalf("unset result = %+v, previous %q, changed %v", entry, previous, changed)
	}

	authFile, err := config.LoadAuth()
	if err != nil {
		t.Fatal(err)
	}
	if authFile.Version != 1 {
		t.Fatalf("auth config version = %d, want 1", authFile.Version)
	}
}

func TestAddOAuthAuthWritesSecretAndListAuth(t *testing.T) {
	svc := setupCoreTestEnv(t)

	entry, err := svc.AddOAuthAuth("oauth", "OAuth", validOAuthClientSecret())
	if err != nil {
		t.Fatalf("AddOAuthAuth = %v", err)
	}
	if entry.Type != config.AuthTypeOAuth {
		t.Fatalf("type = %q, want oauth", entry.Type)
	}

	_, paths, err := svc.AuthPaths("oauth")
	if err != nil {
		t.Fatalf("AuthPaths = %v", err)
	}
	if _, err := os.Stat(paths.ClientSecretPath); err != nil {
		t.Fatalf("client secret missing: %v", err)
	}

	entries, _, err := svc.ListAuth()
	if err != nil {
		t.Fatalf("ListAuth = %v", err)
	}
	if len(entries) != 1 || entries[0].ID != "oauth" {
		t.Fatalf("entries = %+v, want oauth", entries)
	}
}

func TestAddGoogleAuthUsesBuiltInClientWithoutPersistingIt(t *testing.T) {
	_ = setupCoreTestEnv(t)
	svc, err := NewService(context.Background(), WithGoogleOAuthClientCredentials("test-client-id", "test-client-secret"))
	if err != nil {
		t.Fatal(err)
	}

	entry, err := svc.AddGoogleAuthWithQuotaProject("google", "Google", "billing-project")
	if err != nil {
		t.Fatalf("AddGoogleAuthWithQuotaProject = %v", err)
	}
	if entry.Type != config.AuthTypeGoogle || entry.QuotaProjectID != "billing-project" {
		t.Fatalf("entry = %+v", entry)
	}
	if entry.ClientSecretPath != "" || entry.ServiceAccountPath != "" || entry.TokenPath == "" {
		t.Fatalf("google entry paths = %+v, want token only", entry)
	}

	auth, paths, err := svc.AuthPaths("google")
	if err != nil {
		t.Fatal(err)
	}
	if auth.Type != config.AuthTypeGoogle || paths.TokenPath == "" || paths.ClientSecretPath != "" || paths.ServiceAccountPath != "" {
		t.Fatalf("AuthPaths = auth:%+v paths:%+v", auth, paths)
	}
	raw, err := os.ReadFile(config.GetAuthFilePath())
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"test-client-id", "test-client-secret", "client_secret_path", "service_account_path"} {
		if strings.Contains(string(raw), forbidden) {
			t.Fatalf("auth config persisted %q: %s", forbidden, raw)
		}
	}
}

func TestAddGoogleAuthWithoutBuiltInClientFailsBeforePersistence(t *testing.T) {
	svc := setupCoreTestEnv(t)

	_, err := svc.AddGoogleAuth("google", "Google")
	var authErr *AuthError
	if !errors.As(err, &authErr) || authErr.Kind != "configuration" {
		t.Fatalf("AddGoogleAuth error = %#v, want configuration AuthError", err)
	}
	if !strings.Contains(err.Error(), "unavailable in this build") {
		t.Fatalf("AddGoogleAuth error = %q", err)
	}
	if _, statErr := os.Stat(config.GetAuthFilePath()); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("auth config exists after failed add: %v", statErr)
	}
}

func TestAddGoogleAuthReplacesImportedOAuthFiles(t *testing.T) {
	_ = setupCoreTestEnv(t)
	svc, err := NewService(context.Background(), WithGoogleOAuthClientCredentials("test-client-id", "test-client-secret"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AddOAuthAuth("main", "Imported", validOAuthClientSecret()); err != nil {
		t.Fatal(err)
	}
	_, oldPaths, err := svc.AuthPaths("main")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(oldPaths.TokenPath), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(oldPaths.TokenPath, []byte(`{"access_token":"old"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := svc.AddGoogleAuth("main", "Built-in"); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{oldPaths.ClientSecretPath, oldPaths.TokenPath} {
		if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("replaced OAuth file still exists at %s: %v", path, err)
		}
	}
}

func TestDeleteAuthRemovesRegistryEntry(t *testing.T) {
	svc := setupCoreTestEnv(t)
	if _, err := svc.AddGCloudAuth("main", "Main"); err != nil {
		t.Fatalf("AddGCloudAuth = %v", err)
	}

	auth, _, err := svc.DeleteAuth("main")
	if err != nil {
		t.Fatalf("DeleteAuth = %v", err)
	}
	if auth.ID != "main" {
		t.Fatalf("deleted auth = %+v, want main", auth)
	}

	entries, _, err := svc.ListAuth()
	if err != nil {
		t.Fatalf("ListAuth = %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("entries after delete = %+v, want empty", entries)
	}
}

func TestDeleteAuthMissing(t *testing.T) {
	svc := setupCoreTestEnv(t)
	if err := config.SaveAuth(&config.AuthFile{Version: config.AuthConfigVersion}); err != nil {
		t.Fatalf("SaveAuth empty = %v", err)
	}

	_, _, err := svc.DeleteAuth("missing")
	if err == nil || !strings.Contains(err.Error(), `auth "missing" is not configured`) {
		t.Fatalf("DeleteAuth = %v, want not configured error", err)
	}
	var authErr *AuthError
	if !errors.As(err, &authErr) || authErr.Kind != "setup_required" {
		t.Fatalf("DeleteAuth error type = %#v", err)
	}
}

func TestDeleteAuthMissingWithoutAuthFile(t *testing.T) {
	svc := setupCoreTestEnv(t)

	_, _, err := svc.DeleteAuth("missing")
	if err == nil || !strings.Contains(err.Error(), `auth "missing" is not configured`) {
		t.Fatalf("DeleteAuth = %v, want not configured error", err)
	}
}

func TestAddOAuthAuthRejectsInvalidAuthID(t *testing.T) {
	svc := setupCoreTestEnv(t)

	_, err := svc.AddOAuthAuth("../bad", "Bad", []byte("{}"))
	if err == nil {
		t.Fatal("AddOAuthAuth invalid id = nil, want error")
	}
}

func TestAddOAuthAuthRejectsInvalidSecretWithoutPersistingIdentity(t *testing.T) {
	svc := setupCoreTestEnv(t)

	if _, err := svc.AddOAuthAuth("oauth", "OAuth", []byte(`{"installed":{}}`)); err == nil {
		t.Fatal("AddOAuthAuth invalid secret = nil, want error")
	}
	entries, _, err := svc.ListAuth()
	if err != nil {
		t.Fatalf("ListAuth = %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("entries = %+v, want no persisted invalid identity", entries)
	}
}

func validOAuthClientSecret() []byte {
	return []byte(`{"installed":{"client_id":"client-id","client_secret":"client-secret","auth_uri":"https://accounts.google.com/o/oauth2/auth","token_uri":"https://oauth2.googleapis.com/token","redirect_uris":["http://localhost"]}}`)
}
