package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/yumauri/fbrcm/core/env"
)

func setupTestDirs(t *testing.T) {
	t.Helper()
	root := t.TempDir()
	t.Setenv(env.ConfigDir, filepath.Join(root, "config"))
	t.Setenv(env.CacheDir, filepath.Join(root, "cache"))
	t.Setenv(env.Profile, "")
	if err := SetProfileOverride(""); err != nil {
		t.Fatalf("clear profile override: %v", err)
	}
	resetPaths()
}

func writeFile(t *testing.T, path string, content string, mode os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func assertFileMode(t *testing.T, path string, want os.FileMode) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	if got := info.Mode().Perm(); got != want {
		t.Fatalf("%s mode = %o, want %o", path, got, want)
	}
}

func TestEnsurePrivateDirAndFile(t *testing.T) {
	setupTestDirs(t)
	dir := filepath.Join(t.TempDir(), "nested", "private")
	path := filepath.Join(dir, "secret.json")

	if err := EnsurePrivateDir(dir); err != nil {
		t.Fatalf("EnsurePrivateDir returned error: %v", err)
	}
	assertFileMode(t, dir, PrivateDirMode)

	writeFile(t, path, "{}", 0o644)
	if err := EnsurePrivateFile(path); err != nil {
		t.Fatalf("EnsurePrivateFile returned error: %v", err)
	}
	assertFileMode(t, path, PrivateFileMode)
}

func TestLoadAppConfigMissingCorruptAndRoundTrip(t *testing.T) {
	setupTestDirs(t)

	_, err := LoadAppConfig()
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("LoadAppConfig missing = %v, want ErrNotExist", err)
	}

	writeFile(t, GetGlobalConfigFilePath(), "profile = [", PrivateFileMode)
	_, err = LoadAppConfig()
	if err == nil || !strings.Contains(err.Error(), "decode global config") {
		t.Fatalf("LoadAppConfig corrupt = %v, want decode error", err)
	}

	disabled := false
	cfg := &AppConfig{Profile: "work", PowerlineGlyphs: &disabled, Keys: map[string]map[string][]string{"global": {"quit": {"q"}}}}
	if err := SaveAppConfig(cfg); err != nil {
		t.Fatalf("SaveAppConfig returned error: %v", err)
	}
	assertFileMode(t, GetGlobalConfigFilePath(), PrivateFileMode)

	loaded, err := LoadAppConfig()
	if err != nil {
		t.Fatalf("LoadAppConfig returned error: %v", err)
	}
	if loaded.Profile != "work" {
		t.Fatalf("profile = %q, want work", loaded.Profile)
	}
	if loaded.PowerlineGlyphs == nil || *loaded.PowerlineGlyphs {
		t.Fatalf("powerline_glyphs = %v, want false", loaded.PowerlineGlyphs)
	}
}

func TestDecodeAppConfigStrictRejectsUnknownFields(t *testing.T) {
	_, err := DecodeAppConfig([]byte("powerline_glyphs = true\nunknown = true\n"), true)
	if err == nil || !strings.Contains(err.Error(), "unknown") {
		t.Fatalf("DecodeAppConfig strict error = %v, want unknown field", err)
	}

	cfg, err := DecodeAppConfig([]byte("powerline_glyphs = true\nunknown = true\n"), false)
	if err != nil {
		t.Fatalf("DecodeAppConfig non-strict = %v", err)
	}
	if cfg.PowerlineGlyphs == nil || !*cfg.PowerlineGlyphs {
		t.Fatalf("decoded config = %+v", cfg)
	}
}

func TestSaveAppConfigAtomicallyReplacesPrivateFile(t *testing.T) {
	setupTestDirs(t)
	enabled := true
	if err := SaveAppConfig(&AppConfig{PowerlineGlyphs: &enabled}); err != nil {
		t.Fatal(err)
	}
	disabled := false
	if err := SaveAppConfig(&AppConfig{PowerlineGlyphs: &disabled}); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadAppConfigStrict()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.PowerlineGlyphs == nil || *loaded.PowerlineGlyphs {
		t.Fatalf("replaced config = %+v", loaded)
	}
	assertFileMode(t, GetGlobalConfigFilePath(), PrivateFileMode)
}

func TestSwitchProfileWritesConfigToml(t *testing.T) {
	setupTestDirs(t)

	if err := SwitchProfile("staging"); err != nil {
		t.Fatalf("SwitchProfile returned error: %v", err)
	}
	if got := GetActiveProfileName(); got != "staging" {
		t.Fatalf("active profile = %q, want staging", got)
	}

	loaded, err := LoadAppConfig()
	if err != nil {
		t.Fatalf("LoadAppConfig returned error: %v", err)
	}
	if loaded.Profile != "staging" {
		t.Fatalf("config.toml profile = %q, want staging", loaded.Profile)
	}
	assertFileMode(t, profileConfigDir("staging"), PrivateDirMode)
	assertFileMode(t, profileCacheDir("staging"), PrivateDirMode)
}

func TestValidateProfileNameAndAuthID(t *testing.T) {
	t.Parallel()

	for _, id := range []string{"", " ", "..", "a/b", `a\b`} {
		if err := ValidateProfileName(id); err == nil {
			t.Fatalf("ValidateProfileName(%q) = nil, want error", id)
		}
		if err := ValidateAuthID(id); err == nil {
			t.Fatalf("ValidateAuthID(%q) = nil, want error", id)
		}
	}
	if err := ValidateProfileName("valid-name"); err != nil {
		t.Fatalf("ValidateProfileName valid = %v", err)
	}
	if err := ValidateAuthID("valid-id"); err != nil {
		t.Fatalf("ValidateAuthID valid = %v", err)
	}
}

func TestLoadAuthMissingCorruptAndRoundTrip(t *testing.T) {
	setupTestDirs(t)
	if err := SwitchProfile(DefaultProfileName); err != nil {
		t.Fatalf("SwitchProfile returned error: %v", err)
	}

	_, err := LoadAuth()
	if err == nil || !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("LoadAuth missing = %v, want ErrNotExist", err)
	}

	empty, err := LoadAuthOrEmpty()
	if err != nil {
		t.Fatalf("LoadAuthOrEmpty missing = %v", err)
	}
	if empty.Version != AuthConfigVersion || len(empty.Auth) != 0 {
		t.Fatalf("LoadAuthOrEmpty = %+v, want empty v1", empty)
	}

	writeFile(t, GetAuthFilePath(), "{", PrivateFileMode)
	_, err = LoadAuth()
	if err == nil || !strings.Contains(err.Error(), "decode auth config") {
		t.Fatalf("LoadAuth corrupt = %v, want decode error", err)
	}

	entry := DefaultOAuthAuthEntry("main", "Main OAuth")
	file := UpsertAuthEntry(nil, entry)
	if err := SaveAuth(file); err != nil {
		t.Fatalf("SaveAuth returned error: %v", err)
	}
	assertFileMode(t, GetAuthFilePath(), PrivateFileMode)

	loaded, err := LoadAuth()
	if err != nil {
		t.Fatalf("LoadAuth returned error: %v", err)
	}
	got, ok := loaded.FindAuth("main")
	if !ok || got.Label != "Main OAuth" {
		t.Fatalf("FindAuth = %+v, ok=%v", got, ok)
	}
}

func TestLoadProjectsMissingEmptyCorruptAndRoundTrip(t *testing.T) {
	setupTestDirs(t)
	if err := SwitchProfile(DefaultProfileName); err != nil {
		t.Fatalf("SwitchProfile returned error: %v", err)
	}

	_, err := LoadProjects()
	if err == nil || !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("LoadProjects missing = %v, want ErrNotExist", err)
	}

	writeFile(t, GetProjectsFilePath(), "   \n", PrivateFileMode)
	_, err = LoadProjects()
	if !errors.Is(err, ErrEmptyProjectsFile) {
		t.Fatalf("LoadProjects empty = %v, want ErrEmptyProjectsFile", err)
	}

	writeFile(t, GetProjectsFilePath(), "{", PrivateFileMode)
	_, err = LoadProjects()
	if err == nil || !strings.Contains(err.Error(), "decode projects config") {
		t.Fatalf("LoadProjects corrupt = %v, want decode error", err)
	}

	projects := []Project{{
		Name:      "Demo",
		ProjectID: "demo",
		AuthID:    "main",
		Disabled:  true,
	}}
	if err := SaveProjects(projects, time.Now()); err != nil {
		t.Fatalf("SaveProjects returned error: %v", err)
	}
	assertFileMode(t, GetProjectsFilePath(), PrivateFileMode)

	loaded, err := LoadProjects()
	if err != nil {
		t.Fatalf("LoadProjects returned error: %v", err)
	}
	if len(loaded) != 1 || loaded[0].ProjectID != "demo" || !loaded[0].Disabled {
		t.Fatalf("LoadProjects = %+v, want demo project", loaded)
	}
}

func TestLoadParametersCacheMissingCorruptAndRoundTrip(t *testing.T) {
	setupTestDirs(t)
	if err := SwitchProfile(DefaultProfileName); err != nil {
		t.Fatalf("SwitchProfile returned error: %v", err)
	}

	_, err := LoadParametersCache("demo")
	if err == nil || !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("LoadParametersCache missing = %v, want ErrNotExist", err)
	}

	path := GetParametersCachePath("demo")
	writeFile(t, path, "{", PrivateFileMode)
	_, err = LoadParametersCache("demo")
	if err == nil || !strings.Contains(err.Error(), "decode parameters cache") {
		t.Fatalf("LoadParametersCache corrupt = %v, want decode error", err)
	}

	now := time.Now().UTC()
	cache := &ParametersCache{
		ETag:         "etag-1",
		CachedAt:     now,
		RemoteConfig: json.RawMessage(`{"version":{"versionNumber":"1"}}`),
	}
	if err := SaveParametersCache("demo", cache); err != nil {
		t.Fatalf("SaveParametersCache returned error: %v", err)
	}
	versionPath := GetParametersCacheVersionPath("demo", "1")
	assertFileMode(t, versionPath, PrivateFileMode)
	info, err := os.Lstat(path)
	if err != nil {
		t.Fatalf("lstat current pointer: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("current cache path is not a symlink: mode=%s", info.Mode())
	}
	target, err := os.Readlink(path)
	if err != nil {
		t.Fatalf("readlink current pointer: %v", err)
	}
	if target != filepath.Base(versionPath) {
		t.Fatalf("pointer target = %q, want %q", target, filepath.Base(versionPath))
	}

	loaded, err := LoadParametersCache("demo")
	if err != nil {
		t.Fatalf("LoadParametersCache returned error: %v", err)
	}
	if loaded.ETag != "etag-1" {
		t.Fatalf("etag = %q, want etag-1", loaded.ETag)
	}
}

func TestParametersCacheVersionsArePreservedAndLatestWins(t *testing.T) {
	setupTestDirs(t)
	if err := SwitchProfile(DefaultProfileName); err != nil {
		t.Fatalf("SwitchProfile returned error: %v", err)
	}

	if err := SaveParametersCache("demo", &ParametersCache{
		ETag:         "etag-2",
		CachedAt:     time.Now().UTC(),
		RemoteConfig: json.RawMessage(`{"version":{"versionNumber":"2"}}`),
	}); err != nil {
		t.Fatalf("SaveParametersCache v2 = %v", err)
	}
	if err := SaveParametersCache("demo", &ParametersCache{
		ETag:         "etag-10",
		CachedAt:     time.Now().UTC(),
		RemoteConfig: json.RawMessage(`{"version":{"versionNumber":"10"}}`),
	}); err != nil {
		t.Fatalf("SaveParametersCache v10 = %v", err)
	}

	if _, err := os.Stat(GetParametersCacheVersionPath("demo", "2")); err != nil {
		t.Fatalf("v2 snapshot missing: %v", err)
	}
	loaded, err := LoadParametersCache("demo")
	if err != nil {
		t.Fatalf("LoadParametersCache = %v", err)
	}
	if loaded.ETag != "etag-10" {
		t.Fatalf("loaded etag = %q, want etag-10", loaded.ETag)
	}
}

func TestSaveParametersCacheSnapshotDoesNotMoveCurrentPointer(t *testing.T) {
	setupTestDirs(t)
	if err := SwitchProfile(DefaultProfileName); err != nil {
		t.Fatal(err)
	}
	current := &ParametersCache{ETag: "etag-10", CachedAt: time.Now().UTC(), RemoteConfig: json.RawMessage(`{"version":{"versionNumber":"10"}}`)}
	if err := SaveParametersCache("demo", current); err != nil {
		t.Fatal(err)
	}
	historical := &ParametersCache{ETag: "etag-3", CachedAt: time.Now().UTC(), RemoteConfig: json.RawMessage(`{"version":{"versionNumber":"3"}}`)}
	if err := SaveParametersCacheSnapshot("demo", historical); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadParametersCache("demo")
	if err != nil {
		t.Fatal(err)
	}
	if loaded.ETag != "etag-10" {
		t.Fatalf("current pointer moved to %q", loaded.ETag)
	}
	if _, err := LoadParametersCacheVersion("demo", "3"); err != nil {
		t.Fatalf("historical snapshot missing: %v", err)
	}
}

func TestLoadParametersCacheLegacyFile(t *testing.T) {
	setupTestDirs(t)
	if err := SwitchProfile(DefaultProfileName); err != nil {
		t.Fatalf("SwitchProfile returned error: %v", err)
	}

	writeFile(t, GetParametersCachePath("demo"), `{"etag":"legacy","cached_at":"2026-06-21T12:00:00Z","remote_config":{"version":{"versionNumber":"1"}}}`, PrivateFileMode)
	loaded, err := LoadParametersCache("demo")
	if err != nil {
		t.Fatalf("LoadParametersCache legacy = %v", err)
	}
	if loaded.ETag != "legacy" {
		t.Fatalf("legacy etag = %q", loaded.ETag)
	}
}

func TestListParametersCacheSnapshotsSkipsPointer(t *testing.T) {
	setupTestDirs(t)
	if err := SwitchProfile(DefaultProfileName); err != nil {
		t.Fatalf("SwitchProfile returned error: %v", err)
	}
	if err := SaveParametersCache("demo", &ParametersCache{
		ETag:         "etag-1",
		CachedAt:     time.Now().UTC(),
		RemoteConfig: json.RawMessage(`{"version":{"versionNumber":"1"}}`),
	}); err != nil {
		t.Fatalf("SaveParametersCache = %v", err)
	}

	snapshots, err := ListParametersCacheSnapshots()
	if err != nil {
		t.Fatalf("ListParametersCacheSnapshots = %v", err)
	}
	if len(snapshots) != 1 {
		t.Fatalf("snapshot count = %d, want 1: %+v", len(snapshots), snapshots)
	}
	if snapshots[0].ProjectID != "demo" || snapshots[0].Version != "1" {
		t.Fatalf("snapshot = %+v, want demo v1", snapshots[0])
	}
}

func TestSaveParametersCachePointerCopyFallback(t *testing.T) {
	setupTestDirs(t)
	if err := SwitchProfile(DefaultProfileName); err != nil {
		t.Fatalf("SwitchProfile returned error: %v", err)
	}
	original := createParametersCacheSymlink
	createParametersCacheSymlink = func(string, string) error {
		return errors.New("no symlink")
	}
	t.Cleanup(func() { createParametersCacheSymlink = original })

	if err := SaveParametersCache("demo", &ParametersCache{
		ETag:         "etag-1",
		CachedAt:     time.Now().UTC(),
		RemoteConfig: json.RawMessage(`{"version":{"versionNumber":"1"}}`),
	}); err != nil {
		t.Fatalf("SaveParametersCache = %v", err)
	}
	info, err := os.Lstat(GetParametersCachePath("demo"))
	if err != nil {
		t.Fatalf("lstat pointer copy: %v", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		t.Fatalf("pointer fallback is symlink, want regular file")
	}
	loaded, err := LoadParametersCache("demo")
	if err != nil {
		t.Fatalf("LoadParametersCache fallback = %v", err)
	}
	if loaded.ETag != "etag-1" {
		t.Fatalf("fallback etag = %q", loaded.ETag)
	}
}

func TestParametersCacheIsFresh(t *testing.T) {
	now := time.Now()
	fresh := &ParametersCache{CachedAt: now.Add(-5 * time.Minute)}
	stale := &ParametersCache{CachedAt: now.Add(-11 * time.Minute)}

	if !fresh.IsFresh(now) {
		t.Fatal("fresh cache reported stale")
	}
	if stale.IsFresh(now) {
		t.Fatal("stale cache reported fresh")
	}
	if (*ParametersCache)(nil).IsFresh(now) {
		t.Fatal("nil cache reported fresh")
	}
}

func TestDraftLoadSaveDeleteAndList(t *testing.T) {
	setupTestDirs(t)
	if err := SwitchProfile(DefaultProfileName); err != nil {
		t.Fatalf("SwitchProfile returned error: %v", err)
	}

	_, err := LoadDraft("demo")
	if err == nil || !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("LoadDraft missing = %v, want ErrNotExist", err)
	}

	raw := json.RawMessage(`{"version":{"versionNumber":"1"},"parameters":{"flag":{"defaultValue":{"value":"draft"}}}}`)
	now := time.Now().UTC()
	stored := &Draft{FormatVersion: DraftFormatVersion, ProjectID: "demo", BaseVersion: "1", BaseETag: "etag-1", CreatedAt: now, UpdatedAt: now, BaseRemoteConfig: raw, RemoteConfig: raw}
	if err := SaveDraft(stored); err != nil {
		t.Fatalf("SaveDraft returned error: %v", err)
	}
	assertFileMode(t, GetDraftPath("demo"), PrivateFileMode)

	loaded, err := LoadDraft("demo")
	if err != nil {
		t.Fatalf("LoadDraft returned error: %v", err)
	}
	var gotJSON, wantJSON any
	_ = json.Unmarshal(loaded.RemoteConfig, &gotJSON)
	_ = json.Unmarshal(raw, &wantJSON)
	if !reflect.DeepEqual(gotJSON, wantJSON) || loaded.BaseVersion != "1" {
		t.Fatalf("LoadDraft = %+v, want remote config and base version 1", loaded)
	}

	ids, err := ListDraftProjectIDs()
	if err != nil {
		t.Fatalf("ListDraftProjectIDs returned error: %v", err)
	}
	if len(ids) != 1 || ids[0] != "demo" {
		t.Fatalf("ListDraftProjectIDs = %v, want [demo]", ids)
	}

	if err := DeleteDraft("demo"); err != nil {
		t.Fatalf("DeleteDraft returned error: %v", err)
	}
	_, err = LoadDraft("demo")
	if err == nil || !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("LoadDraft after delete = %v, want ErrNotExist", err)
	}
}

func TestListDraftProjectIDsMissingDir(t *testing.T) {
	setupTestDirs(t)
	if err := SwitchProfile(DefaultProfileName); err != nil {
		t.Fatalf("SwitchProfile returned error: %v", err)
	}

	ids, err := ListDraftProjectIDs()
	if err != nil {
		t.Fatalf("ListDraftProjectIDs returned error: %v", err)
	}
	if ids != nil {
		t.Fatalf("ListDraftProjectIDs = %v, want nil", ids)
	}
}

func TestLoadDraftRejectsPlainRemoteConfigFormat(t *testing.T) {
	setupTestDirs(t)
	if err := SwitchProfile(DefaultProfileName); err != nil {
		t.Fatalf("SwitchProfile returned error: %v", err)
	}
	if err := EnsurePrivateDir(GetDraftsDirPath()); err != nil {
		t.Fatalf("EnsurePrivateDir returned error: %v", err)
	}
	writeFile(t, GetDraftPath("demo"), `{"version":{"versionNumber":"1"}}`, PrivateFileMode)
	if _, err := LoadDraft("demo"); err == nil || !strings.Contains(err.Error(), "unsupported draft format") {
		t.Fatalf("LoadDraft legacy format error = %v", err)
	}
}

func TestListProfilesAndEnsureActiveProfile(t *testing.T) {
	setupTestDirs(t)

	profiles, err := ListProfiles()
	if err != nil {
		t.Fatalf("ListProfiles empty root = %v", err)
	}
	if len(profiles) != 0 {
		t.Fatalf("ListProfiles = %v, want empty", profiles)
	}

	if err := SwitchProfile("work"); err != nil {
		t.Fatalf("SwitchProfile work = %v", err)
	}
	if err := SwitchProfile("staging"); err != nil {
		t.Fatalf("SwitchProfile staging = %v", err)
	}

	profiles, err = ListProfiles()
	if err != nil {
		t.Fatalf("ListProfiles = %v", err)
	}
	if len(profiles) != 2 {
		t.Fatalf("ListProfiles = %v, want [staging work]", profiles)
	}
	if profiles[0] != "staging" || profiles[1] != "work" {
		t.Fatalf("ListProfiles = %v, want sorted [staging work]", profiles)
	}

	if err := EnsureActiveProfile(); err != nil {
		t.Fatalf("EnsureActiveProfile = %v", err)
	}
	if got := GetActiveProfileName(); got != "staging" {
		t.Fatalf("active profile = %q, want staging", got)
	}
}

func TestProfileOverrideDoesNotMutatePersistedProfile(t *testing.T) {
	setupTestDirs(t)
	if err := SwitchProfile("active"); err != nil {
		t.Fatal(err)
	}
	if err := SwitchProfile("automation"); err != nil {
		t.Fatal(err)
	}
	if err := SwitchProfile("active"); err != nil {
		t.Fatal(err)
	}

	if err := SetProfileOverride("automation"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = SetProfileOverride("") })
	if err := EnsureActiveProfile(); err != nil {
		t.Fatalf("EnsureActiveProfile override = %v", err)
	}
	if got := GetActiveProfileName(); got != "automation" {
		t.Fatalf("effective profile = %q, want automation", got)
	}
	loaded, err := LoadAppConfig()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Profile != "active" {
		t.Fatalf("persisted profile = %q, want active", loaded.Profile)
	}
}

func TestProfileEnvironmentOverrideAndMissingProfile(t *testing.T) {
	setupTestDirs(t)
	if err := SwitchProfile("active"); err != nil {
		t.Fatal(err)
	}
	if err := SwitchProfile("automation"); err != nil {
		t.Fatal(err)
	}
	if err := SwitchProfile("active"); err != nil {
		t.Fatal(err)
	}
	t.Setenv(env.Profile, "automation")
	resetPaths()
	if err := EnsureActiveProfile(); err != nil {
		t.Fatalf("EnsureActiveProfile env override = %v", err)
	}
	if got := GetActiveProfileName(); got != "automation" {
		t.Fatalf("environment profile = %q, want automation", got)
	}

	t.Setenv(env.Profile, "missing")
	resetPaths()
	if err := EnsureActiveProfile(); err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("missing environment profile error = %v", err)
	}
}

func TestRenameProfileUpdatesActiveProfile(t *testing.T) {
	setupTestDirs(t)
	if err := SwitchProfile("alpha"); err != nil {
		t.Fatalf("SwitchProfile alpha = %v", err)
	}

	if err := RenameProfile("alpha", "beta"); err != nil {
		t.Fatalf("RenameProfile = %v", err)
	}
	if got := GetActiveProfileName(); got != "beta" {
		t.Fatalf("active profile = %q, want beta", got)
	}

	profiles, err := ListProfiles()
	if err != nil {
		t.Fatalf("ListProfiles = %v", err)
	}
	if len(profiles) != 1 || profiles[0] != "beta" {
		t.Fatalf("ListProfiles = %v, want [beta]", profiles)
	}
}

func TestRenameProfileRejectsExistingTarget(t *testing.T) {
	setupTestDirs(t)
	if err := SwitchProfile("one"); err != nil {
		t.Fatalf("SwitchProfile one = %v", err)
	}
	if err := SwitchProfile("two"); err != nil {
		t.Fatalf("SwitchProfile two = %v", err)
	}

	err := RenameProfile("one", "two")
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("RenameProfile = %v, want already exists error", err)
	}
}

func TestDeleteProfileRejectsActiveProfile(t *testing.T) {
	setupTestDirs(t)
	if err := SwitchProfile("keep"); err != nil {
		t.Fatalf("SwitchProfile keep = %v", err)
	}

	err := DeleteProfile("keep")
	if err == nil || !strings.Contains(err.Error(), "cannot delete active profile") {
		t.Fatalf("DeleteProfile active = %v, want rejection", err)
	}
}

func TestDeleteProfileRemovesInactiveProfile(t *testing.T) {
	setupTestDirs(t)
	if err := SwitchProfile("active"); err != nil {
		t.Fatalf("SwitchProfile active = %v", err)
	}
	if err := SwitchProfile("temp"); err != nil {
		t.Fatalf("SwitchProfile temp = %v", err)
	}
	if err := SwitchProfile("active"); err != nil {
		t.Fatalf("SwitchProfile active again = %v", err)
	}

	if err := DeleteProfile("temp"); err != nil {
		t.Fatalf("DeleteProfile temp = %v", err)
	}
	profiles, err := ListProfiles()
	if err != nil {
		t.Fatalf("ListProfiles = %v", err)
	}
	if len(profiles) != 1 || profiles[0] != "active" {
		t.Fatalf("ListProfiles = %v, want [active]", profiles)
	}
}

func TestGetProfileDirPathsValidateName(t *testing.T) {
	setupTestDirs(t)

	if _, err := GetProfileConfigDirPath("valid"); err != nil {
		t.Fatalf("GetProfileConfigDirPath valid = %v", err)
	}
	if _, err := GetProfileCacheDirPath("valid"); err != nil {
		t.Fatalf("GetProfileCacheDirPath valid = %v", err)
	}
	if _, err := GetProfileConfigDirPath("../bad"); err == nil {
		t.Fatal("GetProfileConfigDirPath invalid = nil, want error")
	}
}
