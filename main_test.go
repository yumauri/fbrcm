package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	charmlog "charm.land/log/v2"

	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
	corelog "github.com/yumauri/fbrcm/core/log"
)

func TestCLIInitializationUsesSemanticExitStatus(t *testing.T) {
	err := &config.InvalidConfigurationError{Path: "config.toml", Stage: "decoding", Err: errors.New("invalid configuration")}
	if got := applicationInitializationExitCode(corelog.ModeCLI, err); got != 3 {
		t.Fatalf("CLI initialization exit code = %d, want 3", got)
	}
	if got := applicationInitializationExitCode(corelog.ModeTUI, err); got != 1 {
		t.Fatalf("TUI initialization exit code = %d, want 1", got)
	}
}

func TestJSONDefaultLogLevelIsSilentUnlessEnvironmentOverridesIt(t *testing.T) {
	t.Setenv(env.LogLevel, "")
	if got := defaultLogLevel(corelog.ModeCLI, []string{"capabilities", "--json"}); got != corelog.SilentLevel {
		t.Fatalf("JSON default log level = %v, want silent", got)
	}
	if got := defaultLogLevel(corelog.ModeCLI, []string{"capabilities"}); got != charmlog.InfoLevel {
		t.Fatalf("human default log level = %v, want info", got)
	}
	t.Setenv(env.LogLevel, "debug")
	if got := defaultLogLevel(corelog.ModeCLI, []string{"capabilities", "--json"}); got != charmlog.InfoLevel {
		t.Fatalf("environment-overridden default = %v, want initialization default info", got)
	}
}

func TestBundledThemesAreValid(t *testing.T) {
	entries, err := os.ReadDir("themes")
	if err != nil {
		t.Fatalf("read bundled themes = %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("bundled themes directory is empty")
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".toml" {
			t.Fatalf("unexpected entry in themes directory: %q", entry.Name())
		}
		data, readErr := os.ReadFile(filepath.Join("themes", entry.Name()))
		if readErr != nil {
			t.Errorf("read bundled theme %q = %v", entry.Name(), readErr)
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".toml")
		if validateErr := config.ValidateThemeData(name, data); validateErr != nil {
			t.Errorf("validate bundled theme %q = %v", name, validateErr)
		}
	}
}
