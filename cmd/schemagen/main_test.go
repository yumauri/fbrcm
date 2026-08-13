package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestErrorSchemaGenerationIsDeterministic(t *testing.T) {
	want, err := json.Marshal(errorSchema())
	if err != nil {
		t.Fatal(err)
	}
	for range 20 {
		got, err := json.Marshal(errorSchema())
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != string(want) {
			t.Fatal("error schema condition order changed between generations")
		}
	}
}

func TestContractLockAllowsUnreleasedSurfaceToChangeAtSameVersion(t *testing.T) {
	if err := validateContractLock(
		contractLock{Version: "1.0.0", SHA256: "old"},
		contractLock{Version: "1.0.0", SHA256: "new"},
	); err != nil {
		t.Fatalf("unreleased generated surface was rejected: %v", err)
	}
}

func TestReleasedContractLockRequiresVersionBump(t *testing.T) {
	if err := validateContractLock(
		contractLock{Version: "1.0.0", SHA256: "old", Released: true},
		contractLock{Version: "1.0.0", SHA256: "new"},
	); err == nil {
		t.Fatal("changed released surface was accepted at the same contract version")
	}
	if err := validateContractLock(
		contractLock{Version: "1.0.0", SHA256: "old"},
		contractLock{Version: "1.1.0", SHA256: "new"},
	); err != nil {
		t.Fatalf("versioned generated surface was rejected: %v", err)
	}
}

func TestPublishTransactionRollsBackEveryCommittedAsset(t *testing.T) {
	root := t.TempDir()
	targetOne := filepath.Join(root, "published", "one")
	targetTwo := filepath.Join(root, "published", "two")
	stagedOne := filepath.Join(root, "staged", "one")
	missingTwo := filepath.Join(root, "staged", "missing")
	for path, value := range map[string]string{targetOne: "old-one", targetTwo: "old-two", stagedOne: "new-one"} {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(value), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	err := publishTransaction([]publishItem{{staged: stagedOne, target: targetOne}, {staged: missingTwo, target: targetTwo}})
	if err == nil {
		t.Fatal("publishTransaction accepted a missing staged asset")
	}
	for path, want := range map[string]string{targetOne: "old-one", targetTwo: "old-two"} {
		raw, readErr := os.ReadFile(path)
		if readErr != nil || string(raw) != want {
			t.Fatalf("%s after rollback = %q, %v; want %q", path, raw, readErr, want)
		}
	}
}
