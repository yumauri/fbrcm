package project

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/cli/shared/rc"
)

func TestWriteRemoteConfigFileNormalizesExportJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "remote-config.json")
	raw := []byte("{\"parameters\":{\"flag\":{\"defaultValue\":{\"value\":\"\\u003ctag\\u003e \\u0026 more\"}}}}\n\n")

	if err := rc.WriteRemoteConfigFile(path, raw); err != nil {
		t.Fatalf("WriteRemoteConfigFile returned error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	got := string(data)
	if strings.HasSuffix(got, "\n") || strings.HasSuffix(got, "\r") {
		t.Fatalf("output has trailing line break: %q", got)
	}
	if !strings.Contains(got, `"<tag> & more"`) {
		t.Fatalf("output did not normalize JSON escapes: %s", got)
	}
}

func TestCreateRemoteConfigFileDoesNotOverwriteExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "remote-config.json")
	if err := os.WriteFile(path, []byte("original"), 0o600); err != nil {
		t.Fatal(err)
	}

	err := rc.CreateRemoteConfigFile(path, []byte(`{"parameters":{}}`))
	if err == nil || !errors.Is(err, os.ErrExist) {
		t.Fatalf("CreateRemoteConfigFile error = %v, want file exists", err)
	}
	data, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(data) != "original" {
		t.Fatalf("existing file = %q, want original", data)
	}
}
