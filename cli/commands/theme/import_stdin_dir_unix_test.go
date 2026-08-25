//go:build unix

package theme

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/sys/unix"

	coreconfig "github.com/yumauri/fbrcm/core/config"
)

func TestImportCommandReadsDirectoryFromDevStdinHandle(t *testing.T) {
	root := setupThemeCommandTest(t)
	directory := filepath.Join(root, "theme-pack")
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "firebase.toml"), []byte("[colors]\nprimary = \"#FFC400\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	dir, err := os.Open(directory)
	if err != nil {
		t.Fatal(err)
	}
	fd, err := unix.Dup(int(dir.Fd()))
	if closeErr := dir.Close(); err == nil && closeErr != nil {
		err = closeErr
	}
	if err != nil {
		t.Fatal(err)
	}
	stdin := os.NewFile(uintptr(fd), "/dev/stdin")
	if stdin == nil {
		_ = unix.Close(fd)
		t.Fatal("create /dev/stdin directory handle")
	}
	defer func() { _ = stdin.Close() }()

	cmd := newImportCommand(http.DefaultClient)
	cmd.SetIn(stdin)
	cmd.SetOut(io.Discard)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute = %v", err)
	}
	if _, err := os.Stat(filepath.Join(coreconfig.GetThemesDirPath(), "firebase.toml")); err != nil {
		t.Fatalf("imported theme: %v", err)
	}
}
