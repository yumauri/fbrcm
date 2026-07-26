package mouseutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"charm.land/bubbles/v2/filepicker"
	"github.com/charmbracelet/x/ansi"
)

func TestSelectFilePickerRowMovesExistingSelection(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"a.json", "b.json"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("{}"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	picker := filepicker.New()
	picker.CurrentDirectory = dir
	picker.AllowedTypes = []string{".json"}
	picker.FileAllowed = true
	picker.SetHeight(4)
	picker, _ = picker.Update(picker.Init()())

	var hit bool
	picker, hit = SelectFilePickerRow(picker, 1)
	if !hit {
		t.Fatal("second visible file was not selectable")
	}
	lines := strings.Split(strings.TrimRight(ansi.Strip(picker.View()), "\n"), "\n")
	if len(lines) < 2 || !strings.HasPrefix(lines[1], picker.Cursor) || strings.HasPrefix(lines[0], picker.Cursor) {
		t.Fatalf("file picker cursor did not move to row 1:\n%s", ansi.Strip(picker.View()))
	}
}
