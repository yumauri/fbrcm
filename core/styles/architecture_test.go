package styles

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

var productionHexColor = regexp.MustCompile(`#[0-9A-Fa-f]{3}(?:[0-9A-Fa-f]{3})?`)

func TestProductionColorLiteralsAreCentralized(t *testing.T) {
	_, source, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test source")
	}
	repository := filepath.Clean(filepath.Join(filepath.Dir(source), "..", ".."))
	for _, directory := range []string{"cli", "core", "tui"} {
		err := filepath.WalkDir(filepath.Join(repository, directory), func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") || filepath.Dir(path) == filepath.Join(repository, "core", "styles") {
				return nil
			}
			raw, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			text := string(raw)
			if strings.Contains(text, `lipgloss.Color("`) || productionHexColor.MatchString(text) {
				relative, _ := filepath.Rel(repository, path)
				t.Errorf("production color literal outside core/styles: %s", relative)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestThemingDocumentationListsEveryToken(t *testing.T) {
	_, source, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test source")
	}
	raw, err := os.ReadFile(filepath.Join(filepath.Dir(source), "..", "..", "docs", "theming.md"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	for token, value := range DefaultPalette() {
		if !strings.Contains(text, fmt.Sprintf("%s = %q", token, value)) {
			t.Errorf("theme token %q does not document built-in value %q", token, value)
		}
	}
}
