package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/core/env"
)

func TestFindLocalConfigUsesNearestAncestor(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "repo", "one", "two")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	parent := filepath.Join(root, LocalConfigFileName)
	nearest := filepath.Join(root, "repo", LocalConfigFileName)
	writeFile(t, parent, "profile = \"parent\"\n", 0o644)
	writeFile(t, nearest, "profile = \"nearest\"\n", 0o644)

	path, found, err := FindLocalConfig(nested)
	if err != nil {
		t.Fatal(err)
	}
	if !found || path != nearest {
		t.Fatalf("FindLocalConfig = %q, %v; want %q, true", path, found, nearest)
	}
}

func TestFindLocalConfigReturnsCurrentDirectoryCandidate(t *testing.T) {
	root := t.TempDir()
	path, found, err := FindLocalConfig(root)
	if err != nil {
		t.Fatal(err)
	}
	if found || path != filepath.Join(root, LocalConfigFileName) {
		t.Fatalf("FindLocalConfig = %q, %v", path, found)
	}
}

func TestResolveAppConfigDeeplyOverlaysLocalValues(t *testing.T) {
	setupTestDirs(t)
	root := t.TempDir()
	withWorkingDirectory(t, root)
	enabled := true
	if err := SaveAppConfig(&AppConfig{
		Profile:         "global",
		PowerlineGlyphs: &enabled,
		Keys: map[string]map[string][]string{
			"projects": {"refresh": {"r"}, "delete": {"d"}},
		},
	}); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, LocalConfigFileName), `profile = "local"
powerline_glyphs = false

[keys.projects]
delete = ["D"]
`, 0o644)

	resolved, err := ResolveAppConfig()
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Effective.Profile != "local" || resolved.Effective.PowerlineGlyphs == nil || *resolved.Effective.PowerlineGlyphs {
		t.Fatalf("effective scalar config = %+v", resolved.Effective)
	}
	if got := resolved.Effective.Keys["projects"]["refresh"]; !reflect.DeepEqual(got, []string{"r"}) {
		t.Fatalf("refresh = %v", got)
	}
	if got := resolved.Effective.Keys["projects"]["delete"]; !reflect.DeepEqual(got, []string{"D"}) {
		t.Fatalf("delete = %v", got)
	}
	if resolved.Global.Config.Profile != "global" || resolved.Local.Config.Profile != "local" {
		t.Fatalf("stored layers changed: global=%+v local=%+v", resolved.Global.Config, resolved.Local.Config)
	}
}

func TestResolveAppConfigCanDisableLocalOverlay(t *testing.T) {
	setupTestDirs(t)
	root := t.TempDir()
	withWorkingDirectory(t, root)
	if err := SaveAppConfig(&AppConfig{Profile: "global"}); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, LocalConfigFileName), "profile = \"local\"\n", 0o644)
	SetLocalConfigDisabled(true)
	t.Cleanup(func() { SetLocalConfigDisabled(false) })

	resolved, err := ResolveAppConfig()
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Effective.Profile != "global" || resolved.Local.Exists {
		t.Fatalf("resolution = %+v", resolved)
	}
}

func TestResolveAppConfigReportsLocalDecodePath(t *testing.T) {
	setupTestDirs(t)
	root := t.TempDir()
	withWorkingDirectory(t, root)
	path := filepath.Join(root, LocalConfigFileName)
	writeFile(t, path, "unknown = true\n", 0o644)

	_, err := ResolveAppConfig()
	if err == nil || !strings.Contains(err.Error(), "decode local config") || !strings.Contains(err.Error(), filepath.Base(path)) || !strings.Contains(err.Error(), "unknown") {
		t.Fatalf("ResolveAppConfig error = %v", err)
	}
}

func TestSaveActiveProfileNeverCopiesLocalValuesToGlobal(t *testing.T) {
	setupTestDirs(t)
	root := t.TempDir()
	withWorkingDirectory(t, root)
	writeFile(t, filepath.Join(root, LocalConfigFileName), `profile = "local"

[keys.projects]
refresh = ["u"]
`, 0o644)

	if err := SwitchProfile("global"); err != nil {
		t.Fatal(err)
	}
	stored, err := LoadGlobalAppConfig()
	if err != nil {
		t.Fatal(err)
	}
	if stored.Profile != "global" || len(stored.Keys) != 0 {
		t.Fatalf("global config = %+v", stored)
	}
}

func TestLocalProfileSelectsExistingProfileBelowEnvironmentOverride(t *testing.T) {
	setupTestDirs(t)
	root := t.TempDir()
	withWorkingDirectory(t, root)
	if err := SwitchProfile("global"); err != nil {
		t.Fatal(err)
	}
	if err := SwitchProfile("local"); err != nil {
		t.Fatal(err)
	}
	if err := SwitchProfile("global"); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, LocalConfigFileName), "profile = \"local\"\n", 0o644)
	resetPaths()
	if err := EnsureActiveProfile(); err != nil {
		t.Fatal(err)
	}
	if got := GetActiveProfileName(); got != "local" {
		t.Fatalf("local active profile = %q", got)
	}

	t.Setenv(env.Profile, "global")
	resetPaths()
	if err := EnsureActiveProfile(); err != nil {
		t.Fatal(err)
	}
	if got := GetActiveProfileName(); got != "global" {
		t.Fatalf("environment active profile = %q", got)
	}
}

func TestMissingLocalProfileReportsConfigPath(t *testing.T) {
	setupTestDirs(t)
	root := t.TempDir()
	withWorkingDirectory(t, root)
	path := filepath.Join(root, LocalConfigFileName)
	writeFile(t, path, "profile = \"company\"\n", 0o644)
	resetPaths()

	err := EnsureActiveProfile()
	if err == nil || !strings.Contains(err.Error(), "company") || !strings.Contains(err.Error(), "does not exist") || !strings.Contains(err.Error(), filepath.Base(path)) {
		t.Fatalf("EnsureActiveProfile error = %v", err)
	}
}

func TestDeleteRefusesGlobalAndLocallySelectedProfiles(t *testing.T) {
	setupTestDirs(t)
	root := t.TempDir()
	withWorkingDirectory(t, root)
	if err := SwitchProfile("local"); err != nil {
		t.Fatal(err)
	}
	if err := SwitchProfile("global"); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, LocalConfigFileName), "profile = \"local\"\n", 0o644)
	resetPaths()

	for _, profile := range []string{"global", "local"} {
		if err := EnsureProfileCanDelete(profile); err == nil || !strings.Contains(err.Error(), "active profile") {
			t.Fatalf("EnsureProfileCanDelete(%q) = %v", profile, err)
		}
	}
}

func withWorkingDirectory(t *testing.T, dir string) {
	t.Helper()
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previous); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})
}
