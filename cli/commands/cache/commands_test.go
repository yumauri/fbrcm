package cache

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
)

func setupCacheTest(t *testing.T) {
	t.Helper()
	root := t.TempDir()
	t.Setenv(env.ConfigDir, filepath.Join(root, "config"))
	t.Setenv(env.CacheDir, filepath.Join(root, "cache"))
	if err := config.SwitchProfile(config.DefaultProfileName); err != nil {
		t.Fatal(err)
	}
}

func TestCachePathCommand(t *testing.T) {
	setupCacheTest(t)
	pathCmd, _, err := New().Find([]string{"path"})
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	pathCmd.SetOut(&out)
	if err := pathCmd.RunE(pathCmd, nil); err != nil {
		t.Fatalf("cache path = %v", err)
	}
	if got, want := strings.TrimSpace(out.String()), config.GetParametersCacheDirPath(); got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestCachePathJSON(t *testing.T) {
	setupCacheTest(t)
	pathCmd, _, err := New().Find([]string{"path"})
	if err != nil {
		t.Fatal(err)
	}
	if err := pathCmd.Flags().Set("json", "true"); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	pathCmd.SetOut(&out)
	if err := pathCmd.RunE(pathCmd, nil); err != nil {
		t.Fatalf("cache path json = %v", err)
	}
	if !strings.Contains(out.String(), `"path": "`+config.GetParametersCacheDirPath()+`"`) {
		t.Fatalf("output = %q, want snapshots directory", out.String())
	}
}

func TestCachePurgeWithYes(t *testing.T) {
	setupCacheTest(t)
	if err := config.SaveParametersCache("demo", &config.ParametersCache{
		RemoteConfig: []byte(`{"version":{"versionNumber":"1"}}`),
	}); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	raw := []byte(`{"version":{"versionNumber":"1"}}`)
	if err := config.SaveDraft(&config.Draft{FormatVersion: config.DraftFormatVersion, ProjectID: "demo", BaseVersion: "1", BaseETag: "etag-1", CreatedAt: now, UpdatedAt: now, BaseRemoteConfig: raw, RemoteConfig: raw}); err != nil {
		t.Fatal(err)
	}

	purgeCmd, _, err := New().Find([]string{"purge"})
	if err != nil {
		t.Fatal(err)
	}
	if err := purgeCmd.Flags().Set("yes", "true"); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	purgeCmd.SetOut(&out)
	if err := purgeCmd.RunE(purgeCmd, nil); err != nil {
		t.Fatalf("cache purge = %v", err)
	}
	if !strings.Contains(out.String(), "purged caches") {
		t.Fatalf("output = %q", out.String())
	}
	if _, err := config.LoadDraft("demo"); err != nil {
		t.Fatalf("cache purge removed draft: %v", err)
	}
}

func TestCachePurgeEmptyDoesNotPrompt(t *testing.T) {
	setupCacheTest(t)
	purgeCmd, _, err := New().Find([]string{"purge"})
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	purgeCmd.SetOut(&out)
	if err := purgeCmd.RunE(purgeCmd, nil); err != nil {
		t.Fatalf("empty cache purge = %v", err)
	}
	if !strings.Contains(out.String(), "Nothing to purge") {
		t.Fatalf("output = %q", out.String())
	}
}
