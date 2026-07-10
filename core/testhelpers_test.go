package core

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
	"github.com/yumauri/fbrcm/core/firebase"
)

func setupCoreTestEnv(t *testing.T) *Core {
	t.Helper()
	root := t.TempDir()
	t.Setenv(env.ConfigDir, filepath.Join(root, "config"))
	t.Setenv(env.CacheDir, filepath.Join(root, "cache"))
	if err := config.SwitchProfile(config.DefaultProfileName); err != nil {
		t.Fatalf("SwitchProfile returned error: %v", err)
	}
	svc, err := NewService(context.Background())
	if err != nil {
		t.Fatalf("NewService returned error: %v", err)
	}
	return svc
}

func seedAuthAndProject(t *testing.T, svc *Core, authID, projectID string) {
	t.Helper()
	if _, err := svc.AddGCloudAuth(authID, authID); err != nil {
		t.Fatalf("AddGCloudAuth returned error: %v", err)
	}
	projects := []config.Project{{
		Name:         "Demo",
		ProjectID:    projectID,
		AuthID:       authID,
		DiscoveredBy: []string{authID},
	}}
	if err := config.SaveProjects(projects, time.Now().UTC()); err != nil {
		t.Fatalf("SaveProjects returned error: %v", err)
	}
}

func injectFirebaseService(t *testing.T, svc *Core, authID string, client *firebase.Service) {
	t.Helper()
	svc.InjectFirebaseService(authID, client)
}

func saveStaleParametersCache(t *testing.T, projectID, version string) *ParametersCache {
	t.Helper()
	cache := &config.ParametersCache{
		ETag:         "etag-stale",
		CachedAt:     time.Now().UTC().Add(-15 * time.Minute),
		RemoteConfig: remoteConfigRaw(version, map[string]string{"flag": "cached"}),
	}
	if err := config.SaveParametersCache(projectID, cache); err != nil {
		t.Fatalf("SaveParametersCache returned error: %v", err)
	}
	return cache
}

func writeCorruptParametersCache(t *testing.T, projectID string) {
	t.Helper()
	path := config.GetParametersCachePath(projectID)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir cache dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(`{"etag":"e","cached_at":"2026-06-21T12:00:00Z","remote_config":"not-json"}`), 0o600); err != nil {
		t.Fatalf("write corrupt cache: %v", err)
	}
}

func assertRemoteConfigVersion(t *testing.T, raw json.RawMessage, want string) {
	t.Helper()
	cfg, err := firebase.ParseRemoteConfig(raw)
	if err != nil {
		t.Fatalf("ParseRemoteConfig returned error: %v", err)
	}
	if cfg.Version.VersionNumber != want {
		t.Fatalf("version = %q, want %q", cfg.Version.VersionNumber, want)
	}
}
