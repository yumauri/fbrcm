//go:build unix

package config

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/sys/unix"
)

func TestWriteJSONFilePreservesPreviousOnPartialWrite(t *testing.T) {
	const pathEnv = "FBRCM_TEST_PARTIAL_JSON_WRITE_PATH"
	if path := os.Getenv(pathEnv); path != "" {
		// Limit only this subprocess: the OS accepts part of the write, then
		// returns EFBIG. This exercises a real I/O failure without disk filling.
		signal.Ignore(unix.SIGXFSZ)
		var limit unix.Rlimit
		if err := unix.Getrlimit(unix.RLIMIT_FSIZE, &limit); err != nil {
			t.Fatal(err)
		}
		limit.Cur = 64
		if err := unix.Setrlimit(unix.RLIMIT_FSIZE, &limit); err != nil {
			t.Fatal(err)
		}
		err := writeJSONFile(path, strings.Repeat("new contents", 100))
		if !errors.Is(err, unix.EFBIG) {
			t.Fatalf("write error = %v, want EFBIG", err)
		}
		return
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")
	previous := "{\"value\":\"previous contents\"}\n"
	writeFile(t, path, previous, PrivateFileMode)
	cmd := exec.Command(os.Args[0], "-test.run=^TestWriteJSONFilePreservesPreviousOnPartialWrite$")
	cmd.Env = append(os.Environ(), pathEnv+"="+path)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("partial-write subprocess: %v\n%s", err, output)
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != previous || !json.Valid(got) {
		t.Fatalf("previous JSON damaged: contents = %q, err = %v; want %q", got, err, previous)
	}
	assertFileMode(t, path, PrivateFileMode)
	assertOnlyJSONDestination(t, dir, "state.json")
}
