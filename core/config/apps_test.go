package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yumauri/fbrcm/core/env"
)

func TestAppsCacheIsProfileScopedAndPrivate(t *testing.T) {
	root := t.TempDir()
	t.Setenv(env.ConfigDir, filepath.Join(root, "config"))
	t.Setenv(env.CacheDir, filepath.Join(root, "cache"))
	if err := SwitchProfile(DefaultProfileName); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	if err := SaveAppsIndexCache("demo", now, []byte(`[]`)); err != nil {
		t.Fatal(err)
	}
	path := GetAppsIndexCachePath("demo")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != PrivateFileMode {
		t.Fatalf("file mode = %o", info.Mode().Perm())
	}
	if dirInfo, err := os.Stat(filepath.Dir(path)); err != nil || dirInfo.Mode().Perm() != PrivateDirMode {
		t.Fatalf("dir info = %#v, err = %v", dirInfo, err)
	}
	if err := SwitchProfile("other"); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadAppsIndexCache("demo"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("other profile load error = %v", err)
	}
}

func TestAppConfigCacheUsesHashedAppIDAndExactArtifact(t *testing.T) {
	root := t.TempDir()
	t.Setenv(env.ConfigDir, filepath.Join(root, "config"))
	t.Setenv(env.CacheDir, filepath.Join(root, "cache"))
	if err := SwitchProfile(DefaultProfileName); err != nil {
		t.Fatal(err)
	}
	appID := "1:123:web:opaque/value"
	contents := []byte("{\"apiKey\":\"secret\"}\n")
	if err := SaveAppConfigCache("demo", appID, time.Now().UTC(), "firebase-config.json", "application/json", []byte(`{"app_id":"`+appID+`"}`), contents); err != nil {
		t.Fatal(err)
	}
	path := GetAppConfigCachePath("demo", appID)
	if filepath.Base(path) == appID+".json" || filepath.Base(path) == "value.json" {
		t.Fatalf("app ID was not hashed in %q", path)
	}
	record, got, err := LoadAppConfigCache("demo", appID)
	if err != nil {
		t.Fatal(err)
	}
	if record.AppID != appID || string(got) != string(contents) {
		t.Fatalf("record = %#v, contents = %q", record, got)
	}
}

func TestAppsCacheRejectsMissingTimestamp(t *testing.T) {
	root := t.TempDir()
	t.Setenv(env.ConfigDir, filepath.Join(root, "config"))
	t.Setenv(env.CacheDir, filepath.Join(root, "cache"))
	if err := SwitchProfile(DefaultProfileName); err != nil {
		t.Fatal(err)
	}
	record := AppsCacheRecord{FormatVersion: AppsCacheFormatVersion, Kind: "index", ProjectID: "demo", Payload: json.RawMessage(`[]`)}
	data, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	path := GetAppsIndexCachePath("demo")
	if err := EnsurePrivateDir(filepath.Dir(path)); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, PrivateFileMode); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadAppsIndexCache("demo"); err == nil || !strings.Contains(err.Error(), "invalid index cache metadata") {
		t.Fatalf("load error = %v", err)
	}
	if _, err := ListAppsCacheEntries(); err == nil || !strings.Contains(err.Error(), "invalid cache metadata") {
		t.Fatalf("list error = %v", err)
	}
}
