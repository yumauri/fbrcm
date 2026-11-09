package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	charmlog "charm.land/log/v2"

	"github.com/yumauri/fbrcm/cli/app"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
	corelog "github.com/yumauri/fbrcm/core/log"
	"github.com/yumauri/fbrcm/ops/contract"
)

func TestCLIInitializationUsesSemanticExitStatus(t *testing.T) {
	err := &config.InvalidConfigurationError{Path: "config.toml", Stage: "decoding", Err: errors.New("invalid configuration")}
	if got := applicationInitializationExitCode(corelog.ModeCLI, err); got != 3 {
		t.Fatalf("CLI initialization exit code = %d, want 3", got)
	}
	if got := applicationInitializationExitCode(corelog.ModeTUI, err); got != 1 {
		t.Fatalf("TUI initialization exit code = %d, want 1", got)
	}
	if got := applicationInitializationExitCode(corelog.ModeMCP, err); got != 3 {
		t.Fatalf("MCP initialization exit code = %d, want 3", got)
	}
}

func TestApplicationModeSelection(t *testing.T) {
	for _, test := range []struct {
		args []string
		want corelog.Mode
	}{
		{nil, corelog.ModeTUI},
		{[]string{"mcp"}, corelog.ModeMCP},
		{[]string{"--profile", "work", "mcp", "--allow-writes"}, corelog.ModeMCP},
		{[]string{"mcp", "--stateless"}, corelog.ModeMCP},
		{[]string{"get", "mcp"}, corelog.ModeCLI},
		{[]string{"--profile", "mcp", "get"}, corelog.ModeCLI},
		{[]string{"--help"}, corelog.ModeCLI},
	} {
		if got := applicationMode(test.args); got != test.want {
			t.Errorf("%v: mode=%s want %s", test.args, got, test.want)
		}
	}
}

func TestJSONDefaultLogLevelIsSilentUnlessEnvironmentOverridesIt(t *testing.T) {
	t.Setenv(env.LogLevel, "")
	if got := defaultLogLevel(true); got != corelog.SilentLevel {
		t.Fatalf("JSON default log level = %v, want silent", got)
	}
	if got := defaultLogLevel(false); got != charmlog.InfoLevel {
		t.Fatalf("human default log level = %v, want info", got)
	}
	t.Setenv(env.LogLevel, "debug")
	if got := defaultLogLevel(true); got != charmlog.InfoLevel {
		t.Fatalf("environment-overridden default = %v, want initialization default info", got)
	}
}

func TestCLILaunchFlags(t *testing.T) {
	const helper = "FBRCM_TEST_CLI_LAUNCH"
	if os.Getenv(helper) == "1" {
		os.Args = append(os.Args[:1:1], os.Args[3:]...)
		main()
		os.Exit(0)
	}
	tests := []struct {
		name    string
		args    []string
		command string
		problem string
		exit    int
	}{
		{"persistent flag value", []string{"--json", "config", "path", "--profile", "--json=false"}, "config.path", "", 0},
		{"local flag value", []string{"--json", "config", "path", "--scope", "--json=false"}, "config.path", "argument.invalid", 2},
		{"flag-like profile alone", []string{"config", "path", "--profile", "--json"}, "", "", 0},
		{"local shorthand consumes value", []string{"get", "-p", "--json", "--help"}, "", "", 0},
		{"end of options", []string{"help", "--", "--json"}, "", "", 0},
		{"explicit false wins", []string{"--json", "config", "path", "--json=false"}, "", "", 0},
		{"explicit true wins", []string{"--json=false", "config", "path", "--json=true"}, "config.path", "", 0},
		{"unknown flag before JSON", []string{"config", "path", "--unknown", "--json"}, "config.path", "argument.invalid", 2},
		{"malformed flag before JSON", []string{"config", "path", "---bad", "--json"}, "config.path", "argument.invalid", 2},
		{"malformed flag also used as a value", []string{"config", "path", "---bad", "--profile", "---bad", "--json"}, "config.path", "argument.invalid", 2},
		{"invalid JSON value before explicit false", []string{"config", "path", "--json=invalid", "--json=false"}, "config.path", "argument.invalid", 2},
		{"timeout-looking option value", []string{"config", "path", "--json", "--profile", "--timeout=1ns"}, "config.path", "", 0},
		{"MCP flag value", []string{"mcp", "--json", "--confirmation", "--json=false"}, "mcp", "argument.invalid", 2},
		{"explicit capture stdin unavailable", []string{"explain", "flag", "--from", "-", "--json"}, "explain", "capture.invalid", 8},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			args := append([]string{"-test.run=^TestCLILaunchFlags$", "--"}, tc.args...)
			cmd := exec.CommandContext(t.Context(), os.Args[0], args...)
			cmd.Dir = dir
			cmd.Env = []string{"PATH=" + os.Getenv("PATH"), helper + "=1", "HOME=" + dir, env.ConfigDir + "=" + dir,
				env.CacheDir + "=" + filepath.Join(dir, "cache"), env.Offline + "=1", "NO_COLOR=1"}
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			raw, err := cmd.Output()
			exit := 0
			if processError, ok := errors.AsType[*exec.ExitError](err); ok {
				exit = processError.ExitCode()
			} else if err != nil {
				t.Fatal(err)
			}
			if exit != tc.exit {
				t.Fatalf("exit %d, want %d, stdout=%s, stderr=%s", exit, tc.exit, raw, stderr.Bytes())
			}
			if tc.command == "" {
				if json.Valid(raw) {
					t.Fatalf("human invocation returned JSON: %s", raw)
				}
				return
			}
			var envelope contract.Envelope
			if err := json.Unmarshal(raw, &envelope); err != nil || !bytes.HasSuffix(raw, []byte("\n")) || bytes.HasSuffix(raw, []byte("\n\n")) {
				t.Fatalf("expected one newline-terminated JSON envelope: %s (%v)", raw, err)
			}
			if envelope.Command != tc.command || envelope.Schema != contract.SchemaID(tc.command) || envelope.ExitCode != exit {
				t.Fatalf("unexpected envelope: %s", raw)
			}
			if tc.problem != "" && (len(envelope.Errors) != 1 || envelope.Errors[0].Code != tc.problem) {
				t.Fatalf("expected %s: %s", tc.problem, raw)
			}
		})
	}
	launch := app.InspectLaunch([]string{"--timeout=1s", "config", "path", "--timeout", "2s", "--profile", "--timeout=1ns"})
	if launch.Timeout.String() != "2s" {
		t.Fatalf("timeout = %s, want 2s", launch.Timeout)
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
