package shared

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestConfirmFileOverwrite(t *testing.T) {
	cmd := &cobra.Command{}
	newPath := filepath.Join(t.TempDir(), "new.json")
	overwrite, proceed, err := ConfirmFileOverwrite(cmd, newPath, false)
	if err != nil || overwrite || !proceed {
		t.Fatalf("new destination = overwrite %v, proceed %v, err %v", overwrite, proceed, err)
	}

	existingPath := filepath.Join(t.TempDir(), "existing.json")
	if err := os.WriteFile(existingPath, []byte("existing"), 0o600); err != nil {
		t.Fatal(err)
	}
	overwrite, proceed, err = ConfirmFileOverwrite(cmd, existingPath, true)
	if err != nil || !overwrite || !proceed {
		t.Fatalf("confirmed destination = overwrite %v, proceed %v, err %v", overwrite, proceed, err)
	}
}

func TestConfirmFileOverwriteRejectsDirectory(t *testing.T) {
	path := t.TempDir()
	_, _, err := ConfirmFileOverwrite(&cobra.Command{}, path, true)
	if err == nil || !strings.Contains(err.Error(), "is a directory") {
		t.Fatalf("directory error = %v", err)
	}
}
