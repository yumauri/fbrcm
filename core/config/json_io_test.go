package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteJSONFileAtomicallyReplacesPrivateFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")
	writeFile(t, path, "previous contents", 0o644)
	before, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	value := struct {
		Value string `json:"value"`
	}{Value: "new contents"}
	if err := writeJSONFile(path, value); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if want := "{\n  \"value\": \"new contents\"\n}\n"; string(got) != want {
		t.Fatalf("saved JSON = %q, want %q", got, want)
	}
	after, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if os.SameFile(before, after) {
		t.Fatal("save overwrote the original file instead of replacing it")
	}
	assertFileMode(t, path, PrivateFileMode)
	assertOnlyJSONDestination(t, dir, "state.json")
}

func TestWriteJSONFilePreservesPreviousOnEncodingFailure(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")
	previous := "{\"value\":\"previous contents\"}\n"
	writeFile(t, path, previous, PrivateFileMode)
	if err := writeJSONFile(path, json.RawMessage(`{`)); !isEncodeError(err) {
		t.Fatalf("save error = %v, want encoding error", err)
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != previous {
		t.Fatalf("previous contents = %q, err = %v; want %q", got, err, previous)
	}
	assertOnlyJSONDestination(t, dir, "state.json")
}

func TestWriteJSONFileCleansUpAfterReplacementFailure(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")
	// A nonempty directory makes the final rename fail after staging succeeds.
	sentinel := filepath.Join(path, "keep")
	writeFile(t, sentinel, "previous contents", PrivateFileMode)
	if err := writeJSONFile(path, json.RawMessage(`{}`)); err == nil || isEncodeError(err) {
		t.Fatalf("save error = %v, want filesystem error", err)
	}
	got, err := os.ReadFile(sentinel)
	if err != nil || string(got) != "previous contents" {
		t.Fatalf("existing destination changed: contents = %q, err = %v", got, err)
	}
	assertOnlyJSONDestination(t, dir, "state.json")
}

func assertOnlyJSONDestination(t *testing.T, dir, name string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != name {
		t.Fatalf("unexpected files after save: %v", entries)
	}
}
