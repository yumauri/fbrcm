package harness

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckExpectedStateFileSnapshots(t *testing.T) {
	root := t.TempDir()
	configDir := filepath.Join(root, "config")
	scenarioDir := filepath.Join(root, "scenario")
	statePath := filepath.Join(configDir, "profiles", "default", "projects.json")
	if err := os.MkdirAll(filepath.Dir(statePath), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(statePath, []byte("{\n  \"synced_at\": \"2026-08-15T10:00:00Z\",\n  \"path\": \"/tmp/runtime\"\n}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	scenario := Scenario{
		Directory: scenarioDir,
		ExpectedStateFiles: []StateFileExpectation{{
			Root: "config", Path: filepath.Join("profiles", "default", "projects.json"),
			JSONReplacements: map[string]string{"/synced_at": "<E2E_SYNCED_AT>"},
		}},
	}
	environment := Environment{Variables: []string{"FBRCM_CONFIG_DIR=" + configDir, "FBRCM_CACHE_DIR=" + filepath.Join(root, "cache")}}
	replacements := []SnapshotReplacement{{Old: "/tmp/runtime", New: "<E2E_RUNTIME_PATH>"}}
	changes, err := checkExpectedStateFileSnapshots(scenario, environment, ModeRecordMissing, replacements)
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 1 || !changes[0].Created {
		t.Fatalf("create changes = %#v", changes)
	}
	snapshotPath := filepath.Join(scenarioDir, "state", "config", "profiles", "default", "projects.json.golden")
	raw, err := os.ReadFile(snapshotPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"synced_at": "<E2E_SYNCED_AT>"`) || !strings.Contains(string(raw), `"path": "<E2E_RUNTIME_PATH>"`) {
		t.Fatalf("canonical state snapshot = %s", raw)
	}
	if _, err := checkExpectedStateFileSnapshots(scenario, environment, ModeReplay, replacements); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(statePath, []byte("{\"synced_at\":\"2026-08-16T10:00:00Z\",\"path\":\"changed\"}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := checkExpectedStateFileSnapshots(scenario, environment, ModeReplay, replacements); err == nil {
		t.Fatal("state snapshot replay accepted changed non-canonical bytes")
	}
}

func TestJSONPointerSnapshotReplacementsRejectsMissingPointer(t *testing.T) {
	_, err := jsonPointerSnapshotReplacements([]byte(`{"present":true}`), map[string]string{"/missing": "<MISSING>"})
	if err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("missing pointer error = %v", err)
	}
}

func TestCheckExpectedAbsentStatePathsAcceptsFilesAndDirectoriesRemoved(t *testing.T) {
	root := t.TempDir()
	configDir := filepath.Join(root, "config")
	cacheDir := filepath.Join(root, "cache")
	environment := Environment{Variables: []string{"FBRCM_CONFIG_DIR=" + configDir, "FBRCM_CACHE_DIR=" + cacheDir}}
	scenario := Scenario{ExpectedAbsentStatePaths: []StatePathExpectation{
		{Root: "config", Path: "removed.json"},
		{Root: "cache", Path: "removed-directory"},
	}}
	if err := checkExpectedAbsentStatePaths(scenario, environment); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(cacheDir, "removed-directory")
	if err := os.MkdirAll(path, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := checkExpectedAbsentStatePaths(scenario, environment); err == nil || !strings.Contains(err.Error(), "directory still exists") {
		t.Fatalf("existing directory error = %v", err)
	}
}
