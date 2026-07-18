package core

import (
	"os"
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/core/config"
)

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

func TestPurgeAuthRemovesRegistryEntry(t *testing.T) {
	svc := setupCoreTestEnv(t)
	if _, err := svc.AddGCloudAuth("main", "Main"); err != nil {
		t.Fatalf("AddGCloudAuth = %v", err)
	}

	auth, _, err := svc.PurgeAuth("main")
	if err != nil {
		t.Fatalf("PurgeAuth = %v", err)
	}
	if auth.ID != "main" {
		t.Fatalf("purged auth = %+v, want main", auth)
	}

	entries, _, err := svc.ListAuth()
	if err != nil {
		t.Fatalf("ListAuth = %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("entries after purge = %+v, want empty", entries)
	}
}

func TestPurgeAuthMissing(t *testing.T) {
	svc := setupCoreTestEnv(t)
	if err := config.SaveAuth(&config.AuthFile{Version: config.AuthConfigVersion}); err != nil {
		t.Fatalf("SaveAuth empty = %v", err)
	}

	_, _, err := svc.PurgeAuth("missing")
	if err == nil || !strings.Contains(err.Error(), `auth "missing" is not configured`) {
		t.Fatalf("PurgeAuth = %v, want not configured error", err)
	}
}

func TestPurgeAuthMissingWithoutAuthFile(t *testing.T) {
	svc := setupCoreTestEnv(t)

	_, _, err := svc.PurgeAuth("missing")
	if err == nil || !strings.Contains(err.Error(), `auth "missing" is not configured`) {
		t.Fatalf("PurgeAuth = %v, want not configured error", err)
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
