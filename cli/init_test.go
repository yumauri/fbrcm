package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
	corestyles "github.com/yumauri/fbrcm/core/styles"
)

func TestConfigureThemeSkipsColorlessMachineAndStatelessRuns(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		noColor string
	}{
		{name: "no color", noColor: "1"},
		{name: "JSON", args: []string{"version", "--json"}},
		{name: "stateless", args: []string{"get", "--stateless"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			t.Setenv(env.ConfigDir, filepath.Join(root, "config"))
			t.Setenv(env.CacheDir, filepath.Join(root, "cache"))
			t.Setenv(env.NoColor, test.noColor)
			if err := os.MkdirAll(filepath.Join(root, "config", "themes"), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, "config", "config.toml"), []byte("theme = \"custom\"\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, "config", "themes", "custom.toml"), []byte("[colors]\nprimary = \"1\"\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			corestyles.ResetPalette()
			t.Cleanup(corestyles.ResetPalette)

			configureTheme(test.args)
			if got := corestyles.ColorValue(corestyles.TokenPrimary); got != corestyles.DefaultPalette()[corestyles.TokenPrimary] {
				t.Fatalf("primary = %q, theme file was read", got)
			}
		})
	}
}

func TestConfigureThemeLoadsHumanTheme(t *testing.T) {
	root := t.TempDir()
	t.Setenv(env.ConfigDir, filepath.Join(root, "config"))
	t.Setenv(env.CacheDir, filepath.Join(root, "cache"))
	t.Setenv(env.NoColor, "")
	if err := os.MkdirAll(filepath.Join(root, "config", "themes"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "config", "config.toml"), []byte("theme = \"custom\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "config", "themes", "custom.toml"), []byte("[colors]\nprimary = \"1\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	corestyles.ResetPalette()
	t.Cleanup(corestyles.ResetPalette)
	t.Cleanup(func() { config.SetLocalConfigDisabled(false) })

	configureTheme([]string{"projects", "list", "--no-local-config"})
	if got := corestyles.ColorValue(corestyles.TokenPrimary); got != "1" {
		t.Fatalf("primary = %q", got)
	}
}

func TestBooleanFlagEnabled(t *testing.T) {
	if !booleanFlagEnabled([]string{"get", "--stateless"}, "stateless") {
		t.Fatal("bare flag not enabled")
	}
	if booleanFlagEnabled([]string{"get", "--stateless=false"}, "stateless") {
		t.Fatal("false flag enabled")
	}
	if booleanFlagEnabled([]string{"get", "--", "--stateless"}, "stateless") {
		t.Fatal("flag after argument terminator enabled")
	}
}
