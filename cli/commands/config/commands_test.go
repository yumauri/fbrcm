package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/shared"
	coreconfig "github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
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
	if string(result.Value) != "true" || result.Source != "default" {
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
	if cfg.Profile != "work" || cfg.PowerlineGlyphs != nil {
		t.Fatalf("config after reset = %+v", cfg)
	}
	if len(cfg.Keys) != 0 {
		t.Fatal("reset did not remove stored key overrides")
	}
}

func TestConfigResetKeysRemovesObsoleteBindingsFromInvalidConfig(t *testing.T) {
	setupConfigCommandTest(t)
	if err := coreconfig.SaveAppConfigRaw([]byte(`[keys.compare]
close = ["esc"]

[keys.global]
focus_compare = ["9"]

[keys.projects]
compare = ["v"]
`)); err != nil {
		t.Fatal(err)
	}

	stdout, _, err := executeConfigCommand(t, New(), "reset", "keys", "--yes")
	if err != nil {
		t.Fatal(err)
	}
	if stdout != "reset keys\n" {
		t.Fatalf("reset output = %q", stdout)
	}

	cfg, err := coreconfig.LoadAppConfigStrict()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Keys) != 0 {
		t.Fatalf("keys after reset = %#v, want no stored overrides", cfg.Keys)
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

func TestConfigValidateReportsFirebaseRCAliasErrors(t *testing.T) {
	root := setupConfigCommandTest(t)
	withCommandWorkingDirectory(t, root)
	if err := os.WriteFile(filepath.Join(root, coreconfig.FirebaseConfigFileName), []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, coreconfig.FirebaseRCFileName), []byte(`{"projects":{"Prod":"acme-production-42"}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	stdout, _, err := executeConfigCommand(t, New(), "validate", "--scope", "effective", "--json")
	var exitErr *shared.ExitError
	if !errors.As(err, &exitErr) || exitErr.Code != 1 {
		t.Fatalf("validate error = %#v", err)
	}
	var report configValidationResult
	if err := json.Unmarshal([]byte(stdout), &report); err != nil {
		t.Fatal(err)
	}
	if report.Valid || len(report.Errors) != 1 || report.Errors[0].Code != "project_alias_source" || !strings.Contains(report.Errors[0].Message, coreconfig.FirebaseRCFileName) {
		t.Fatalf("Firebase RC validation report = %+v", report)
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

func TestRunEditorUsesProcessTerminalStreams(t *testing.T) {
	process := &exec.Cmd{}
	attachEditorTerminal(process)
	if process.Stdin != os.Stdin || process.Stdout != os.Stdout || process.Stderr != os.Stderr {
		t.Fatalf("editor streams = stdin:%T stdout:%T stderr:%T", process.Stdin, process.Stdout, process.Stderr)
	}
}

func TestConfigShowMergesNearestLocalConfigAndReportsSource(t *testing.T) {
	root := setupConfigCommandTest(t)
	work := filepath.Join(root, "repo", "nested")
	if err := os.MkdirAll(work, 0o755); err != nil {
		t.Fatal(err)
	}
	withCommandWorkingDirectory(t, work)
	enabled := true
	if err := coreconfig.SaveAppConfig(&coreconfig.AppConfig{
		PowerlineGlyphs: &enabled,
		Keys:            map[string]map[string][]string{"projects": {"refresh": {"r"}}},
	}); err != nil {
		t.Fatal(err)
	}
	localPath := filepath.Join(root, "repo", coreconfig.LocalConfigFileName)
	if err := os.WriteFile(localPath, []byte("powerline_glyphs = false\n\n[keys.projects]\ndelete = ['D']\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	stdout, _, err := executeConfigCommand(t, New(), "show", "powerline_glyphs", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var result configValueResult
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatal(err)
	}
	if string(result.Value) != "false" || result.Source != "local" {
		t.Fatalf("result = %+v", result)
	}

	stdout, _, err = executeConfigCommand(t, New(), "show", "keys.projects.refresh", "--json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatal(err)
	}
	if result.Source != "global" {
		t.Fatalf("refresh source = %+v", result)
	}
}

func TestConfigPathLocalRequiresDiscoveredFile(t *testing.T) {
	root := setupConfigCommandTest(t)
	work := filepath.Join(root, "repo", "nested")
	if err := os.MkdirAll(work, 0o755); err != nil {
		t.Fatal(err)
	}
	withCommandWorkingDirectory(t, work)

	stdout, _, err := executeConfigCommand(t, New(), "path", "--scope", "local")
	if err == nil || !strings.Contains(err.Error(), "no local config found") || !strings.Contains(err.Error(), "config edit --scope local") {
		t.Fatalf("path error = %v", err)
	}
	if stdout != "" {
		t.Fatalf("path wrote nonexistent candidate to stdout: %q", stdout)
	}

	path := filepath.Join(root, "repo", coreconfig.LocalConfigFileName)
	if err := os.WriteFile(path, []byte("powerline_glyphs = false\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	stdout, _, err = executeConfigCommand(t, New(), "path", "--scope", "local")
	if err != nil {
		t.Fatal(err)
	}
	gotInfo, gotErr := os.Stat(strings.TrimSpace(stdout))
	wantInfo, wantErr := os.Stat(path)
	if gotErr != nil || wantErr != nil || !os.SameFile(gotInfo, wantInfo) {
		t.Fatalf("local path = %q, want %q (got stat: %v, want stat: %v)", stdout, path, gotErr, wantErr)
	}
}

func TestConfigLocalMutationDoesNotMaterializeEffectiveValues(t *testing.T) {
	root := setupConfigCommandTest(t)
	work := filepath.Join(root, "repo")
	if err := os.MkdirAll(work, 0o755); err != nil {
		t.Fatal(err)
	}
	withCommandWorkingDirectory(t, work)
	if err := coreconfig.SaveAppConfigRaw([]byte("[keys.projects]\nrefresh = ['r']\n")); err != nil {
		t.Fatal(err)
	}

	if _, _, err := executeConfigCommand(t, New(), "set", "powerline_glyphs", "false", "--scope", "local"); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(work, coreconfig.LocalConfigFileName))
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	if !strings.Contains(text, "powerline_glyphs = false") || strings.Contains(text, "refresh") || strings.Contains(text, "[keys") {
		t.Fatalf("local config materialized inherited values:\n%s", text)
	}
	global, err := coreconfig.LoadGlobalAppConfig()
	if err != nil {
		t.Fatal(err)
	}
	if global.PowerlineGlyphs != nil {
		t.Fatalf("local mutation changed global config: %+v", global)
	}
}

func TestConfigProjectAliasKeysAreRepositoryScoped(t *testing.T) {
	root := setupConfigCommandTest(t)
	work := filepath.Join(root, "repo")
	if err := os.MkdirAll(work, 0o755); err != nil {
		t.Fatal(err)
	}
	withCommandWorkingDirectory(t, work)

	if _, _, err := executeConfigCommand(t, New(), "set", "projects.aliases.prod", "acme-production-42", "--scope", "local"); err != nil {
		t.Fatal(err)
	}
	stdout, _, err := executeConfigCommand(t, New(), "show", "projects.aliases.prod", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var result configValueResult
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatal(err)
	}
	if string(result.Value) != `"acme-production-42"` || result.Source != "local" {
		t.Fatalf("alias result = %+v", result)
	}

	_, _, err = executeConfigCommand(t, New(), "set", "projects.aliases.prod", "other-production-42", "--scope", "global")
	if err == nil || !strings.Contains(err.Error(), "repository-scoped") {
		t.Fatalf("global alias error = %v", err)
	}

	if _, _, err := executeConfigCommand(t, New(), "reset", "projects.aliases.prod", "--scope", "local", "--yes"); err != nil {
		t.Fatal(err)
	}
	aliases, err := coreconfig.LoadProjectAliases()
	if err != nil || len(aliases) != 0 {
		t.Fatalf("aliases after reset = %#v, %v", aliases, err)
	}
}

func TestConfigEditFullProvidesGeneratedKeyReference(t *testing.T) {
	root := setupConfigCommandTest(t)
	withCommandWorkingDirectory(t, root)
	var staged string
	edit := newEditCommand(func(cmd *cobra.Command, editor, path string) error {
		raw, err := os.ReadFile(path)
		staged = string(raw)
		return err
	})
	command := &cobra.Command{Use: "config"}
	command.AddCommand(edit)
	if _, _, err := executeConfigCommand(t, command, "edit", "--full"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(staged, "Complete generated template") || !strings.Contains(staged, "[keys.projects]") || !strings.Contains(staged, "refresh") {
		t.Fatalf("full staged config lacks generated key reference:\n%s", staged)
	}
}

func TestConfigEditMissingStartsSparse(t *testing.T) {
	root := setupConfigCommandTest(t)
	withCommandWorkingDirectory(t, root)
	var staged string
	edit := newEditCommand(func(cmd *cobra.Command, editor, path string) error {
		raw, err := os.ReadFile(path)
		staged = string(raw)
		return err
	})
	command := &cobra.Command{Use: "config"}
	command.AddCommand(edit)
	if _, _, err := executeConfigCommand(t, command, "edit"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(staged, "Add only values") || strings.Contains(staged, "[keys.") {
		t.Fatalf("sparse staged config = %q", staged)
	}
}

func TestConfigEditLocalCanRepairInvalidTOML(t *testing.T) {
	root := setupConfigCommandTest(t)
	withCommandWorkingDirectory(t, root)
	path := filepath.Join(root, coreconfig.LocalConfigFileName)
	if err := os.WriteFile(path, []byte("unknown = true\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	edit := newEditCommand(func(cmd *cobra.Command, editor, stagedPath string) error {
		return os.WriteFile(stagedPath, []byte("powerline_glyphs = false\n"), 0o600)
	})
	command := &cobra.Command{Use: "config"}
	command.AddCommand(edit)
	if _, _, err := executeConfigCommand(t, command, "edit", "--scope", "local"); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != "powerline_glyphs = false\n" {
		t.Fatalf("repaired local config = %q", raw)
	}
}

func withCommandWorkingDirectory(t *testing.T, dir string) {
	t.Helper()
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previous); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})
}
