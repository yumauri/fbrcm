package theme

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/cli/shared"
	coreconfig "github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
	corestyles "github.com/yumauri/fbrcm/core/styles"
)

func setupThemeCommandTest(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	t.Setenv(env.ConfigDir, filepath.Join(root, "config"))
	t.Setenv(env.CacheDir, filepath.Join(root, "cache"))
	t.Setenv(env.NoLocalConfig, "")
	coreconfig.SetLocalConfigDisabled(false)
	t.Cleanup(func() { coreconfig.SetLocalConfigDisabled(false) })
	return root
}

func TestImportCommandImportsLocalFile(t *testing.T) {
	root := setupThemeCommandTest(t)
	source := filepath.Join(root, "firebase.toml")
	if err := os.WriteFile(source, []byte("[colors]\nprimary = \"#FFC400\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cmd := newImportCommand(http.DefaultClient)
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetArgs([]string{source})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute = %v", err)
	}
	if !strings.Contains(output.String(), "imported theme: firebase") {
		t.Fatalf("output = %q", output.String())
	}
	if _, err := os.Stat(filepath.Join(coreconfig.GetThemesDirPath(), "firebase.toml")); err != nil {
		t.Fatalf("imported file: %v", err)
	}
}

func TestImportCommandImportsDirectoryAndWarnsForExistingThemes(t *testing.T) {
	root := setupThemeCommandTest(t)
	directory := filepath.Join(root, "theme-pack")
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := coreconfig.ImportTheme("existing", []byte("[colors]\nprimary = \"1\"\n")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "existing.toml"), []byte(strings.Repeat("x", maximumThemeImportBytes+1)), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "fresh.toml"), []byte("inherits = \"existing\"\n[colors]\nerror = \"2\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	cmd := newImportCommand(http.DefaultClient)
	cmd.SetContext(shared.WithMachineState(context.Background()))
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetArgs([]string{directory})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute = %v", err)
	}
	if strings.Contains(output.String(), "imported theme: existing") || !strings.Contains(output.String(), "imported theme: fresh") {
		t.Fatalf("output = %q", output.String())
	}
	warnings := shared.MachineWarnings(cmd)
	if len(warnings) != 1 || warnings[0].Code != "theme.already_exists" || warnings[0].Target != "existing" {
		t.Fatalf("warnings = %#v", warnings)
	}
	resolved, err := coreconfig.LoadTheme("fresh")
	if err != nil {
		t.Fatalf("LoadTheme(fresh) = %v", err)
	}
	if resolved.Palette[corestyles.TokenPrimary] != "1" || resolved.Palette[corestyles.TokenError] != "2" {
		t.Fatalf("fresh palette = %#v", resolved.Palette)
	}
}

func TestImportCommandDirectoryRejectsNameOverride(t *testing.T) {
	root := setupThemeCommandTest(t)
	directory := filepath.Join(root, "theme-pack")
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	cmd := newImportCommand(http.DefaultClient)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"--name", "renamed", directory})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "--name cannot be used") {
		t.Fatalf("execute error = %v", err)
	}
}

func TestImportCommandSingleThemeStillRejectsExistingDestination(t *testing.T) {
	root := setupThemeCommandTest(t)
	source := filepath.Join(root, "firebase.toml")
	if err := os.WriteFile(source, []byte("[colors]\nprimary = \"1\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := coreconfig.ImportTheme("firebase", []byte("[colors]\nprimary = \"2\"\n")); err != nil {
		t.Fatal(err)
	}
	cmd := newImportCommand(http.DefaultClient)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{source})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("execute error = %v", err)
	}
}

func TestReadThemeImportDownloadsHTTPSource(t *testing.T) {
	setupThemeCommandTest(t)
	client := &http.Client{Transport: themeRoundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Status: "200 OK", Body: io.NopCloser(strings.NewReader("[colors]\nprimary = \"1\"\n"))}, nil
	})}
	cmd := newImportCommand(client)
	sourceURL := "https://example.com/remote.toml"
	raw, name, source, err := readThemeImport(cmd, client, sourceURL, "")
	if err != nil {
		t.Fatalf("readThemeImport = %v", err)
	}
	if name != "remote" || source != sourceURL || !bytes.Contains(raw, []byte("primary")) {
		t.Fatalf("result = %q, %q, %q", raw, name, source)
	}
}

func TestImportCommandNameOverridesSourceFilename(t *testing.T) {
	root := setupThemeCommandTest(t)
	source := filepath.Join(root, "firebase.toml")
	if err := os.WriteFile(source, []byte("[colors]\nprimary = \"#FFC400\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cmd := newImportCommand(http.DefaultClient)
	cmd.SetOut(io.Discard)
	cmd.SetArgs([]string{"--name", "custom", source})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute = %v", err)
	}
	if _, err := os.Stat(filepath.Join(coreconfig.GetThemesDirPath(), "custom.toml")); err != nil {
		t.Fatalf("renamed import: %v", err)
	}
}

func TestImportCommandReadsStdinWithRequiredName(t *testing.T) {
	root := setupThemeCommandTest(t)
	stdinPath := filepath.Join(root, "stdin-theme")
	if err := os.WriteFile(stdinPath, []byte("[colors]\nprimary = \"#FFC400\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	stdin, err := os.Open(stdinPath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = stdin.Close() }()

	cmd := newImportCommand(http.DefaultClient)
	cmd.SetOut(io.Discard)
	cmd.SetIn(stdin)
	cmd.SetArgs([]string{"--name", "piped"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute = %v", err)
	}
	if _, err := os.Stat(filepath.Join(coreconfig.GetThemesDirPath(), "piped.toml")); err != nil {
		t.Fatalf("stdin import: %v", err)
	}
}

func TestImportCommandStdinRequiresName(t *testing.T) {
	root := setupThemeCommandTest(t)
	stdinPath := filepath.Join(root, "stdin-theme")
	if err := os.WriteFile(stdinPath, []byte("[colors]\nprimary = \"1\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	stdin, err := os.Open(stdinPath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = stdin.Close() }()

	cmd := newImportCommand(http.DefaultClient)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	cmd.SetIn(stdin)
	err = cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "--name is required") {
		t.Fatalf("execute error = %v", err)
	}
}

func TestImportCommandReadsThemeDirectoryFromStdin(t *testing.T) {
	root := setupThemeCommandTest(t)
	if _, err := coreconfig.ImportTheme("z-base", []byte("[colors]\nprimary = \"#FFC400\"\n")); err != nil {
		t.Fatal(err)
	}
	directory := filepath.Join(root, "theme-pack")
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "a-child.toml"), []byte("inherits = \"z-base\"\n[colors]\nerror = \"1\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "z-base.toml"), []byte("invalid source that must be skipped"), 0o600); err != nil {
		t.Fatal(err)
	}
	stdin, err := os.Open(directory)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = stdin.Close() }()

	cmd := newImportCommand(http.DefaultClient)
	cmd.SetContext(shared.WithMachineState(context.Background()))
	cmd.SetOut(io.Discard)
	cmd.SetIn(stdin)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute = %v", err)
	}
	for _, name := range []string{"a-child.toml", "z-base.toml"} {
		if _, err := os.Stat(filepath.Join(coreconfig.GetThemesDirPath(), name)); err != nil {
			t.Fatalf("imported %s: %v", name, err)
		}
	}
	warnings := shared.MachineWarnings(cmd)
	if len(warnings) != 1 || warnings[0].Target != "z-base" {
		t.Fatalf("warnings = %#v", warnings)
	}
}

func TestImportCommandSourceTakesPrecedenceWithoutConsumingStdin(t *testing.T) {
	root := setupThemeCommandTest(t)
	source := filepath.Join(root, "source.toml")
	stdinPath := filepath.Join(root, "stdin-theme")
	if err := os.WriteFile(source, []byte("[colors]\nprimary = \"1\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(stdinPath, []byte("[colors]\nprimary = \"2\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	stdin, err := os.Open(stdinPath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = stdin.Close() }()

	cmd := newImportCommand(http.DefaultClient)
	cmd.SetOut(io.Discard)
	cmd.SetIn(stdin)
	cmd.SetArgs([]string{source})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute = %v", err)
	}
	offset, err := stdin.Seek(0, io.SeekCurrent)
	if err != nil {
		t.Fatal(err)
	}
	if offset != 0 {
		t.Fatalf("stdin offset = %d, want 0", offset)
	}
}

type themeRoundTripFunc func(*http.Request) (*http.Response, error)

func (fn themeRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return fn(req) }

func TestImportCommandRequiresSourceInMachineMode(t *testing.T) {
	setupThemeCommandTest(t)
	cmd := newImportCommand(http.DefaultClient)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	cmd.Flags().Bool("json", true, "")
	if err := cmd.Flags().Set("json", "true"); err != nil {
		t.Fatal(err)
	}
	err := cmd.Execute()
	var interaction *shared.InteractionError
	if !errors.As(err, &interaction) || interaction.Type != "external_input" {
		t.Fatalf("error = %#v, %v", interaction, err)
	}
}

func TestImportCommandWritesMachineResultForNamedStdin(t *testing.T) {
	root := setupThemeCommandTest(t)
	stdinPath := filepath.Join(root, "stdin-theme")
	if err := os.WriteFile(stdinPath, []byte("[colors]\nprimary = \"#FFC400\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	stdin, err := os.Open(stdinPath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = stdin.Close() }()

	cmd := newImportCommand(http.DefaultClient)
	cmd.Flags().Bool("json", true, "")
	cmd.SetIn(stdin)
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetArgs([]string{"--json", "--name", "piped"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute = %v", err)
	}
	var result themeImportResult
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatalf("decode output %q: %v", output.String(), err)
	}
	if result.Theme != "piped" || result.Source != "<stdin>" || result.Status != "imported" {
		t.Fatalf("result = %#v", result)
	}
}

func TestImportCommandWritesMachineResultForDirectory(t *testing.T) {
	root := setupThemeCommandTest(t)
	directory := filepath.Join(root, "theme-pack")
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := coreconfig.ImportTheme("existing", []byte("[colors]\nprimary = \"1\"\n")); err != nil {
		t.Fatal(err)
	}
	for name, data := range map[string]string{
		"existing.toml": "[colors]\nprimary = \"2\"\n",
		"fresh.toml":    "[colors]\nprimary = \"3\"\n",
	} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(data), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	cmd := newImportCommand(http.DefaultClient)
	cmd.SetContext(shared.WithMachineState(context.Background()))
	cmd.Flags().Bool("json", true, "")
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetArgs([]string{"--json", directory})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute = %v", err)
	}
	var result themeBatchImportResult
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatalf("decode output %q: %v", output.String(), err)
	}
	if result.Count != 2 || result.ImportedCount != 1 || result.SkippedCount != 1 || len(result.Items) != 2 {
		t.Fatalf("result = %#v", result)
	}
	if len(shared.MachineWarnings(cmd)) != 1 {
		t.Fatalf("warnings = %#v", shared.MachineWarnings(cmd))
	}
}

func TestImportCommandRejectsDirectoryStdinInMachineMode(t *testing.T) {
	root := setupThemeCommandTest(t)
	directory := filepath.Join(root, "theme-pack")
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	stdin, err := os.Open(directory)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = stdin.Close() }()

	cmd := newImportCommand(http.DefaultClient)
	cmd.Flags().Bool("json", true, "")
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	cmd.SetIn(stdin)
	cmd.SetArgs([]string{"--json"})
	err = cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "experimental human-mode") {
		t.Fatalf("execute error = %v", err)
	}
}
