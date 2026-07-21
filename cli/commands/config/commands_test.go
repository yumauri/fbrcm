package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/shared"
	coreconfig "github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
	tuiconfig "github.com/yumauri/fbrcm/tui/config"
)

func setupConfigCommandTest(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	t.Setenv(env.ConfigDir, filepath.Join(root, "config"))
	t.Setenv(env.CacheDir, filepath.Join(root, "cache"))
	t.Setenv(env.Profile, "")
	return root
}

func executeConfigCommand(t *testing.T, cmd *cobra.Command, args ...string) (string, string, error) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetIn(strings.NewReader(""))
	cmd.SetArgs(args)
	err := cmd.Execute()
	return stdout.String(), stderr.String(), err
}

func TestConfigShowMissingUsesDefaultsWithoutCreatingFile(t *testing.T) {
	setupConfigCommandTest(t)
	stdout, _, err := executeConfigCommand(t, New(), "show", "powerline_glyphs", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var result configValueResult
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatal(err)
	}
	if result.Value != true || result.Source != "default" {
		t.Fatalf("result = %+v", result)
	}
	if _, err := os.Stat(coreconfig.GetGlobalConfigFilePath()); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("show created config file: %v", err)
	}
}

func TestConfigShowSelectedCompoundValuesAsScopedTOML(t *testing.T) {
	setupConfigCommandTest(t)
	for _, key := range []string{"keys.global", "keys.global.help"} {
		stdout, _, err := executeConfigCommand(t, New(), "show", key)
		if err != nil {
			t.Fatalf("show %s: %v", key, err)
		}
		if !strings.Contains(stdout, "[keys.global]") || !strings.Contains(stdout, `help = ['?']`) {
			t.Fatalf("show %s output = %q, want scoped TOML", key, stdout)
		}
		cfg, err := coreconfig.DecodeAppConfig([]byte(stdout), true)
		if err != nil {
			t.Fatalf("decode show %s output: %v\n%s", key, err, stdout)
		}
		if got := cfg.Keys["global"]["help"]; !reflect.DeepEqual(got, []string{"?"}) {
			t.Fatalf("show %s decoded help = %v", key, got)
		}
	}
}

func TestConfigSetTypedValuesAndRejectsConflict(t *testing.T) {
	setupConfigCommandTest(t)
	if _, _, err := executeConfigCommand(t, New(), "set", "powerline_glyphs", "false"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := executeConfigCommand(t, New(), "set", "keys.projects.refresh", "u", "ctrl+r"); err != nil {
		t.Fatal(err)
	}
	cfg, err := coreconfig.LoadAppConfigStrict()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.PowerlineGlyphs == nil || *cfg.PowerlineGlyphs {
		t.Fatalf("powerline_glyphs = %v", cfg.PowerlineGlyphs)
	}
	if got := cfg.Keys["projects"]["refresh"]; !reflect.DeepEqual(got, []string{"u", "ctrl+r"}) {
		t.Fatalf("refresh = %v", got)
	}

	_, _, err = executeConfigCommand(t, New(), "set", "keys.projects.refresh", "enter")
	if err == nil || !strings.Contains(err.Error(), "conflicts") {
		t.Fatalf("conflict error = %v", err)
	}
	cfg, err = coreconfig.LoadAppConfigStrict()
	if err != nil {
		t.Fatal(err)
	}
	if got := cfg.Keys["projects"]["refresh"]; !reflect.DeepEqual(got, []string{"u", "ctrl+r"}) {
		t.Fatalf("failed set changed refresh = %v", got)
	}
}

func TestConfigResetPreservesProfile(t *testing.T) {
	setupConfigCommandTest(t)
	if err := coreconfig.SwitchProfile("work"); err != nil {
		t.Fatal(err)
	}
	disabled := false
	if err := coreconfig.SaveAppConfig(&coreconfig.AppConfig{
		Profile: "work", PowerlineGlyphs: &disabled,
		Keys: map[string]map[string][]string{"projects": {"refresh": {"ctrl+r"}}},
	}); err != nil {
		t.Fatal(err)
	}
	stdout, _, err := executeConfigCommand(t, New(), "reset", "--yes", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var result configResetResult
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatal(err)
	}
	if result.Status != "reset" || !result.Changed {
		t.Fatalf("result = %+v", result)
	}
	cfg, err := coreconfig.LoadAppConfigStrict()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Profile != "work" || cfg.PowerlineGlyphs == nil || !*cfg.PowerlineGlyphs {
		t.Fatalf("config after reset = %+v", cfg)
	}
	if !reflect.DeepEqual(cfg.Keys, tuiconfig.ToConfigMap(tuiconfig.DefaultKeyMap())) {
		t.Fatal("reset did not restore default key map")
	}
}

func TestConfigValidateReportsAllKeyErrorsAsJSON(t *testing.T) {
	setupConfigCommandTest(t)
	if err := coreconfig.SaveAppConfigRaw([]byte(`powerline_glyphs = true

[keys.projects]
refresh = ["enter"]
unknown = ["x"]
`)); err != nil {
		t.Fatal(err)
	}
	stdout, _, err := executeConfigCommand(t, New(), "validate", "--json")
	var exitErr *shared.ExitError
	if !errors.As(err, &exitErr) || exitErr.Code != 1 {
		t.Fatalf("validate error = %#v", err)
	}
	var report configValidationResult
	if err := json.Unmarshal([]byte(stdout), &report); err != nil {
		t.Fatal(err)
	}
	if report.Valid || len(report.Errors) < 2 {
		t.Fatalf("report = %+v", report)
	}
}

func TestConfigEditStagesValidChanges(t *testing.T) {
	setupConfigCommandTest(t)
	var gotEditor string
	edit := newEditCommand(func(cmd *cobra.Command, editor, path string) error {
		gotEditor = editor
		return os.WriteFile(path, []byte("powerline_glyphs = false\n"), 0o600)
	})
	root := &cobra.Command{Use: "config"}
	root.AddCommand(edit)
	if _, _, err := executeConfigCommand(t, root, "edit", "--editor", "code --wait"); err != nil {
		t.Fatal(err)
	}
	if gotEditor != "code --wait" {
		t.Fatalf("editor = %q", gotEditor)
	}
	cfg, err := coreconfig.LoadAppConfigStrict()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.PowerlineGlyphs == nil || *cfg.PowerlineGlyphs {
		t.Fatalf("edited config = %+v", cfg)
	}
}

func TestConfigEditInvalidKeepsOriginal(t *testing.T) {
	setupConfigCommandTest(t)
	enabled := true
	if err := coreconfig.SaveAppConfig(&coreconfig.AppConfig{PowerlineGlyphs: &enabled}); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(coreconfig.GetGlobalConfigFilePath())
	if err != nil {
		t.Fatal(err)
	}
	edit := newEditCommand(func(cmd *cobra.Command, editor, path string) error {
		return os.WriteFile(path, []byte("unknown = true\n"), 0o600)
	})
	root := &cobra.Command{Use: "config"}
	root.AddCommand(edit)
	_, _, err = executeConfigCommand(t, root, "edit")
	if err == nil || !strings.Contains(err.Error(), "original was not changed") || !strings.Contains(err.Error(), ".config.toml.edit-") {
		t.Fatalf("edit error = %v", err)
	}
	after, readErr := os.ReadFile(coreconfig.GetGlobalConfigFilePath())
	if readErr != nil {
		t.Fatal(readErr)
	}
	if !bytes.Equal(before, after) {
		t.Fatalf("invalid edit changed original:\nbefore=%q\nafter=%q", before, after)
	}
}

func TestResolveEditorPrecedence(t *testing.T) {
	t.Setenv(env.Editor, "fbrcm-editor")
	t.Setenv("VISUAL", "visual-editor")
	t.Setenv("EDITOR", "editor")
	if got := resolveEditor(""); got != "fbrcm-editor" {
		t.Fatalf("resolved editor = %q", got)
	}
	if got := resolveEditor("explicit"); got != "explicit" {
		t.Fatalf("explicit editor = %q", got)
	}
}
