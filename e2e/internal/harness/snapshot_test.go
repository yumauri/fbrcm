package harness

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckSnapshotLifecycle(t *testing.T) {
	path := filepath.Join(t.TempDir(), "stdout.golden")
	if _, err := CheckSnapshot(path, []byte("first\n"), false); err == nil {
		t.Fatal("CheckSnapshot() accepted a missing snapshot in compare mode")
	}
	change, err := CheckSnapshot(path, []byte("first\n"), true)
	if err != nil {
		t.Fatal(err)
	}
	if !change.Created || change.Updated {
		t.Fatalf("create change = %+v", change)
	}
	if _, err := CheckSnapshot(path, []byte("first\n"), false); err != nil {
		t.Fatal(err)
	}
	if _, err := CheckSnapshot(path, []byte("second\n"), false); err == nil || !strings.Contains(err.Error(), "first differing byte") {
		t.Fatalf("mismatch error = %v", err)
	}
	change, err = CheckSnapshot(path, []byte("second\n"), true)
	if err != nil {
		t.Fatal(err)
	}
	if !change.Updated || change.Created {
		t.Fatalf("update change = %+v", change)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != "second\n" {
		t.Fatalf("snapshot = %q", raw)
	}
}

func TestCanonicalizeSnapshotReplacesLongestValuesFirst(t *testing.T) {
	raw := []byte("root=/tmp/run tools=/tmp/run/tools proxy=http://127.0.0.1:3210 address=127.0.0.1:3210\n")
	got := CanonicalizeSnapshot(raw,
		SnapshotReplacement{Old: "/tmp/run", New: "<E2E_RUN_ROOT>"},
		SnapshotReplacement{Old: "/tmp/run/tools", New: "<E2E_TOOLS_ROOT>"},
		SnapshotReplacement{Old: "http://127.0.0.1:3210", New: "<E2E_PROXY_URL>"},
		SnapshotReplacement{Old: "127.0.0.1:3210", New: "<E2E_PROXY_ADDRESS>"},
	)
	want := "root=<E2E_RUN_ROOT> tools=<E2E_TOOLS_ROOT> proxy=<E2E_PROXY_URL> address=<E2E_PROXY_ADDRESS>\n"
	if string(got) != want {
		t.Fatalf("canonical snapshot = %q, want %q", got, want)
	}
	if string(raw) == string(got) {
		t.Fatal("CanonicalizeSnapshot() modified its input")
	}
}
