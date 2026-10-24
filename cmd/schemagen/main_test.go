package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/yumauri/fbrcm/ops/contract"
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
		contractLock{Version: "1.0.0", SHA256: "same", Released: true},
		contractLock{Version: "1.0.0", SHA256: "same"},
	); err != nil {
		t.Fatalf("unchanged released surface was rejected: %v", err)
	}
	if err := validateContractLock(
		contractLock{Version: "1.0.0", SHA256: "old", Released: true},
		contractLock{Version: "1.0.0", SHA256: "new"},
	); err == nil {
		t.Fatal("changed released surface was accepted at the same contract version")
	}
	if err := validateContractLock(
		contractLock{Version: "1.0.0", SHA256: "old", Released: true},
		contractLock{Version: "1.1.0", SHA256: "new"},
	); err != nil {
		t.Fatalf("versioned generated surface was rejected: %v", err)
	}
}

func TestGenerationPreservesReleasedContractLock(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	stage := filepath.Join(root, "stage")
	schemaDir := filepath.Join(stage, "schemas", "cli", contract.Version)
	write(schemaDir, "example.schema.json", map[string]string{"type": "object"})
	digest, err := generatedContractDigest(stage, schemaDir)
	if err != nil {
		t.Fatal(err)
	}
	previous := contractLock{Version: contract.Version, SHA256: digest, Released: true}
	write(filepath.Join("schemas", "cli"), "contract.lock.json", previous)
	enforceContractLock(stage, schemaDir)
	raw, err := os.ReadFile(filepath.Join(stage, "schemas", "cli", "contract.lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	var generated contractLock
	if err := json.Unmarshal(raw, &generated); err != nil {
		t.Fatal(err)
	}
	if generated != previous {
		t.Fatalf("released lock changed during generation: got %+v, want %+v", generated, previous)
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
