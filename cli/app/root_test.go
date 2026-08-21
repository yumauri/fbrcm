package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/charmbracelet/x/term"
	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/contract"
	"github.com/yumauri/fbrcm/cli/machine"
	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
	"github.com/yumauri/fbrcm/core/firebase"
	corelog "github.com/yumauri/fbrcm/core/log"
)

func TestNewRootCommandBuildsFreshRoot(t *testing.T) {
	first := newRootCommand(nil, "1.2.3", "abc123", "2026-06-14")
	second := newRootCommand(nil, "1.2.3", "abc123", "2026-06-14")

	if first == second {
		t.Fatalf("newRootCommand returned the same command instance")
	}
	if first.Use != "fbrcm" || first.Short != "Firebase Remote Config manager" {
		t.Fatalf("root metadata = %q/%q, want fbrcm/Firebase Remote Config manager", first.Use, first.Short)
	}
	if first.Version != "1.2.3 (commit abc123, built 2026-06-14)" {
		t.Fatalf("version = %q, want formatted version", first.Version)
	}
	if first.VersionTemplate() != versionTemplate {
		t.Fatalf("version template = %q, want package template", first.VersionTemplate())
	}
	if len(first.Commands()) != len(second.Commands()) {
		t.Fatalf("command counts differ: %d vs %d", len(first.Commands()), len(second.Commands()))
	}
	if _, ok := first.ErrOrStderr().(term.File); !ok {
		t.Fatalf("root stderr type = %T, want terminal-capable progress writer", first.ErrOrStderr())
	}
	if got, want := commandNames(first), []string{"add", "auth", "cache", "capabilities", "completion", "conditions", "config", "delete", "doctor", "draft", "duplicate", "experiments", "get", "groups", "help", "hooks", "personalizations", "profile", "project", "projects", "rollouts", "schema", "update", "versions"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("root commands = %#v, want %#v", got, want)
	}
}

func TestRootCommandKeepsProfileBypassContract(t *testing.T) {
	cmd := newRootCommand(nil, "1.2.3", "abc123", "2026-06-14")

	profile, _, err := cmd.Find([]string{"profile", "list"})
	if err != nil {
		t.Fatalf("find profile list: %v", err)
	}
	if !isProfileCommand(profile) {
		t.Fatalf("profile list command no longer bypasses active profile setup")
	}
}

func TestRootCommandConstructionDoesNotAccumulateSubcommands(t *testing.T) {
	var counts []int
	for range 3 {
		cmd := newRootCommand(nil, "1.2.3", "abc123", "2026-06-14")
		counts = append(counts, len(cmd.Commands()))
	}

	if !reflect.DeepEqual(counts, []int{24, 24, 24}) {
		t.Fatalf("command counts = %#v, want stable counts without accumulation", counts)
	}
}

func TestRootCommandDefinesProfileOverride(t *testing.T) {
	cmd := newRootCommand(nil, "1.2.3", "abc123", "2026-06-14")
	flag := cmd.PersistentFlags().Lookup("profile")
	if flag == nil {
		t.Fatal("root --profile flag is missing")
	}
	if !strings.Contains(flag.Usage, "FBRCM_PROFILE") || !strings.Contains(flag.Usage, "without changing") {
		t.Fatalf("profile usage = %q", flag.Usage)
	}
	localFlag := cmd.PersistentFlags().Lookup("no-local-config")
	if localFlag == nil || !strings.Contains(localFlag.Usage, "FBRCM_NO_LOCAL_CONFIG") {
		t.Fatalf("no-local-config flag = %#v", localFlag)
	}
	statelessFlag := cmd.PersistentFlags().Lookup("stateless")
	if statelessFlag == nil || !strings.Contains(statelessFlag.Usage, "without profiles") || !strings.Contains(statelessFlag.Usage, env.GoogleAccessToken) {
		t.Fatalf("stateless flag = %#v", statelessFlag)
	}
}

func TestRootCommandRejectsWhitespaceOnlyInvocationValues(t *testing.T) {
	tests := [][]string{
		{"--profile", "   ", "cache", "path"},
		{"config", "set", "powerline_glyphs", "   "},
		{"delete", "   "},
	}
	for _, args := range tests {
		t.Run(strings.Join(args, "_"), func(t *testing.T) {
			cmd := NewRootForContract("test")
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			cmd.SetArgs(args)
			_, err := cmd.ExecuteC()
			var argument *shared.ArgumentError
			if !errors.As(err, &argument) {
				t.Fatalf("ExecuteC(%q) error = %T %v, want typed argument error", args, err, err)
			}
		})
	}
}

func TestCommandProgressMessageUsesMeaningfulPhases(t *testing.T) {
	root := &cobra.Command{Use: "fbrcm"}
	projects := &cobra.Command{Use: "projects"}
	update := &cobra.Command{Use: "update"}
	root.AddCommand(projects)
	projects.AddCommand(update)

	if got := commandProgressMessage(update); got != "Syncing projects…" {
		t.Fatalf("projects update progress = %q", got)
	}

	unknown := &cobra.Command{Use: "local"}
	root.AddCommand(unknown)
	if got := commandProgressMessage(unknown); got != "Working…" {
		t.Fatalf("fallback progress = %q", got)
	}
}

func TestRootCommandSkipsConnectivityProbeForHelpAndVersion(t *testing.T) {
	for _, args := range [][]string{{"--help"}, {"help"}, {"--version"}} {
		t.Run(strings.Join(args, "_"), func(t *testing.T) {
			calls := 0
			cmd := newRootCommandWithOfflineInit(nil, "1.2.3", "abc123", "2026-06-14", func(_ context.Context, probe bool) {
				if probe {
					calls++
				}
			})
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})
			cmd.SetArgs(args)
			if err := cmd.Execute(); err != nil {
				t.Fatalf("execute %v: %v", args, err)
			}
			if calls != 0 {
				t.Fatalf("connectivity probe calls for %v = %d, want 0", args, calls)
			}
		})
	}
}

func TestRootCommandTreatsConfigAsLocalRecoverySurface(t *testing.T) {
	root := t.TempDir()
	t.Setenv(env.ConfigDir, filepath.Join(root, "config"))
	t.Setenv(env.CacheDir, filepath.Join(root, "cache"))
	t.Setenv(env.Profile, "../invalid")

	calls := 0
	cmd := newRootCommandWithOfflineInit(nil, "1.2.3", "abc123", "2026-06-14", func(_ context.Context, probe bool) {
		if probe {
			calls++
		}
	})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"config", "show", "powerline_glyphs"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute config: %v", err)
	}
	if calls != 0 {
		t.Fatalf("connectivity probe calls = %d, want 0", calls)
	}
}

func TestRootCommandTreatsHooksAsLocalRecoverySurface(t *testing.T) {
	root := t.TempDir()
	t.Setenv(env.ConfigDir, filepath.Join(root, "config"))
	t.Setenv(env.CacheDir, filepath.Join(root, "cache"))
	t.Setenv(env.Profile, "../invalid")

	calls := 0
	cmd := newRootCommandWithOfflineInit(nil, "1.2.3", "abc123", "2026-06-14", func(_ context.Context, probe bool) {
		if probe {
			calls++
		}
	})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"hooks", "status", "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute hooks: %v", err)
	}
	if calls != 0 {
		t.Fatalf("connectivity probe calls = %d, want 0", calls)
	}
}

func TestRootCommandTreatsProjectAliasesAsLocalRecoverySurface(t *testing.T) {
	root := t.TempDir()
	t.Setenv(env.ConfigDir, filepath.Join(root, "config"))
	t.Setenv(env.CacheDir, filepath.Join(root, "cache"))
	t.Setenv(env.Profile, "../invalid")

	calls := 0
	cmd := newRootCommandWithOfflineInit(nil, "1.2.3", "abc123", "2026-06-14", func(_ context.Context, probe bool) {
		if probe {
			calls++
		}
	})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"projects", "aliases", "list", "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute project aliases: %v", err)
	}
	if calls != 0 {
		t.Fatalf("connectivity probe calls = %d, want 0", calls)
	}
}

func TestRootCommandSkipsProbeForLocalProfileCommand(t *testing.T) {
	root := t.TempDir()
	t.Setenv(env.ConfigDir, filepath.Join(root, "config"))
	t.Setenv(env.CacheDir, filepath.Join(root, "cache"))
	t.Setenv(env.Profile, "")
	if err := config.SetProfileOverride(""); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = config.SetProfileOverride("") })

	calls := 0
	cmd := newRootCommandWithOfflineInit(nil, "1.2.3", "abc123", "2026-06-14", func(_ context.Context, probe bool) {
		if probe {
			calls++
		}
	})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"profile"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute profile: %v", err)
	}
	if calls != 0 {
		t.Fatalf("connectivity probe calls = %d, want 0", calls)
	}
}

func TestCommandExitCodeHonorsDiffContract(t *testing.T) {
	cmd := &cobra.Command{Use: "diff"}
	original := &shared.ExpressionError{Expression: "group ==", Context: "parameter", Err: fmt.Errorf("failed")}
	if got := commandExitCode(cmd, original); got != 8 {
		t.Fatalf("default error exit code = %d, want 8", got)
	}
	explicit := shared.DiffFoundError(cmd)
	if got := commandExitCode(cmd, explicit); got != 1 {
		t.Fatalf("diff found exit code = %d, want 1", got)
	}
	var exitErr *shared.ExitError
	if !errors.As(explicit, &exitErr) {
		t.Fatalf("explicit error = %#v", explicit)
	}
}

func TestCommandExitCodeCoversPreRunDiffErrors(t *testing.T) {
	root := &cobra.Command{Use: "fbrcm"}
	diff := &cobra.Command{Use: "diff <left> <right>", Args: cobra.ExactArgs(2), RunE: func(*cobra.Command, []string) error { return nil }}
	root.AddCommand(diff)
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"diff"})
	executed, err := root.ExecuteC()
	if err == nil {
		t.Fatal("argument error is nil")
	}
	if got := commandExitCode(executed, err); got != 2 {
		t.Fatalf("argument error exit code = %d, want 2", got)
	}
}

func TestCommandExitCodeClassifiesUnknownFlagAsArgumentFailure(t *testing.T) {
	cmd := &cobra.Command{Use: "diff"}
	if got := commandExitCode(cmd, fmt.Errorf("unknown flag")); got != 2 {
		t.Fatalf("unknown flag exit code = %d, want 2", got)
	}
}

func TestRootProfileFlagSelectsWithoutSwitching(t *testing.T) {
	root := t.TempDir()
	t.Setenv(env.ConfigDir, filepath.Join(root, "config"))
	t.Setenv(env.CacheDir, filepath.Join(root, "cache"))
	t.Setenv(env.Profile, "")
	if err := config.SetProfileOverride(""); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = config.SetProfileOverride("") })
	for _, profile := range []string{"active", "automation", "active"} {
		if err := config.SwitchProfile(profile); err != nil {
			t.Fatal(err)
		}
	}

	cmd := newRootCommandWithOfflineInit(nil, "1.2.3", "abc123", "2026-06-14", func(context.Context, bool) {})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--profile", "automation", "cache", "path"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute with profile = %v", err)
	}
	if want := filepath.Join(root, "cache", "automation", "remote-config"); !strings.Contains(out.String(), want) {
		t.Fatalf("cache path = %q, want %q", out.String(), want)
	}
	appConfig, err := config.LoadAppConfig()
	if err != nil {
		t.Fatal(err)
	}
	if appConfig.Profile != "active" {
		t.Fatalf("persisted profile = %q, want active", appConfig.Profile)
	}
}

func TestProfilelessMachineContextSkipsProfileSelection(t *testing.T) {
	root := t.TempDir()
	t.Setenv(env.ConfigDir, filepath.Join(root, "config"))
	t.Setenv(env.CacheDir, filepath.Join(root, "cache"))
	t.Setenv(env.Profile, "../invalid")
	if err := config.SetProfileOverride(""); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = config.SetProfileOverride("") })

	ran := false
	cmd := newRootCommandWithOfflineInit(nil, "1.2.3", "abc123", "2026-06-14", func(context.Context, bool) {})
	cmd.AddCommand(&cobra.Command{
		Use: "profileless-probe",
		RunE: func(*cobra.Command, []string) error {
			ran = true
			return nil
		},
	})
	cmd.SetContext(machine.WithProfileless(shared.WithMachineState(context.Background())))
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"profileless-probe", "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("profileless machine execution = %v", err)
	}
	if !ran {
		t.Fatal("profileless machine command did not run")
	}
	if _, err := os.Stat(filepath.Join(root, "config")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("config root stat = %v, want no profile filesystem access", err)
	}
}

func TestStatelessProjectExportWritesDestinationWithoutProfileFiles(t *testing.T) {
	root := t.TempDir()
	configRoot := filepath.Join(root, "config")
	cacheRoot := filepath.Join(root, "cache")
	destination := filepath.Join(root, "output", "remote-config.json")
	t.Setenv(env.ConfigDir, configRoot)
	t.Setenv(env.CacheDir, cacheRoot)
	t.Setenv(env.Profile, "../invalid")
	t.Setenv(env.GoogleAccessToken, "one-shot-token")
	t.Setenv(env.GoogleCloudQuotaProject, "quota-project")
	if err := config.SetProfileOverride(""); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = config.SetProfileOverride("") })

	requestCount := 0
	originalTransport := http.DefaultTransport
	http.DefaultTransport = appRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		requestCount++
		if got := req.Header.Get("Authorization"); got != "Bearer one-shot-token" {
			t.Fatalf("authorization header = %q", got)
		}
		if got := req.Header.Get("X-Goog-User-Project"); got != "quota-project" {
			t.Fatalf("quota project header = %q", got)
		}
		if !strings.Contains(req.URL.EscapedPath(), "/projects/demo/namespaces/firebase-server/remoteConfig") {
			t.Fatalf("request path = %q", req.URL.EscapedPath())
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     http.Header{"ETag": []string{`"etag-direct"`}},
			Body:       io.NopCloser(strings.NewReader(`{"parameters":{"flag":{"defaultValue":{"value":"\u003cdirect\u003e"}}}}`)),
			Request:    req,
		}, nil
	})
	t.Cleanup(func() { http.DefaultTransport = originalTransport })

	svc, err := core.NewService(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	cmd := newRootCommandWithOfflineInit(svc, "1.2.3", "abc123", "2026-06-14", func(context.Context, bool) {})
	var captured bytes.Buffer
	cmd.SetOut(&captured)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"--stateless", "project", "export", "server@demo", "--to", destination, "--json"})
	executed, err := cmd.ExecuteC()
	if err != nil {
		t.Fatalf("stateless project export = %v", err)
	}
	if policy := core.ExecutionPolicyFromContext(executed.Context()); policy != core.StatelessExecutionPolicy() {
		t.Fatalf("execution policy = %#v, want stateless", policy)
	}
	envelope := contract.BuildEnvelope(executed, "1.2.3", captured.Bytes(), nil)
	if envelope.Context.Profile != nil || envelope.Outcome != "success" {
		t.Fatalf("envelope context/outcome = %#v/%q", envelope.Context, envelope.Outcome)
	}
	if strings.Contains(captured.String(), "one-shot-token") {
		t.Fatal("machine output contains the access token")
	}
	written, err := os.ReadFile(destination)
	if err != nil {
		t.Fatal(err)
	}
	if string(written) != `{"parameters":{"flag":{"defaultValue":{"value":"<direct>"}}}}` {
		t.Fatalf("destination = %q", written)
	}
	if requestCount != 1 {
		t.Fatalf("request count = %d, want 1", requestCount)
	}
	for _, path := range []string{configRoot, cacheRoot} {
		if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("profile path %s stat = %v, want not found", path, err)
		}
	}
}

func TestStatelessProjectOpenRequiresNoTokenOrProfileFiles(t *testing.T) {
	root := t.TempDir()
	configRoot := filepath.Join(root, "config")
	cacheRoot := filepath.Join(root, "cache")
	t.Setenv(env.ConfigDir, configRoot)
	t.Setenv(env.CacheDir, cacheRoot)
	t.Setenv(env.Profile, "../invalid")
	t.Setenv(env.GoogleAccessToken, "")
	if err := config.SetProfileOverride(""); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = config.SetProfileOverride("")
		config.SetLocalConfigDisabled(false)
	})

	cmd := newRootCommandWithOfflineInit(nil, "1.2.3", "abc123", "2026-06-14", func(context.Context, bool) {})
	var captured bytes.Buffer
	cmd.SetOut(&captured)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"--stateless", "project", "open", "demo-project", "--json"})
	executed, err := cmd.ExecuteC()
	if err != nil {
		t.Fatalf("stateless project open = %v", err)
	}
	if policy := core.ExecutionPolicyFromContext(executed.Context()); policy != core.StatelessExecutionPolicy() {
		t.Fatalf("execution policy = %#v, want stateless", policy)
	}
	envelope := contract.BuildEnvelope(executed, "1.2.3", captured.Bytes(), nil)
	compact := compactJSON(t, captured.Bytes())
	for _, marker := range []string{`"project_id":"demo-project"`, `"url":"https://console.firebase.google.com/project/demo-project/config"`, `"opened":false`} {
		if !strings.Contains(compact, marker) {
			t.Fatalf("project open output %s omits %s", captured.String(), marker)
		}
	}
	if envelope.Context.Profile != nil || envelope.Outcome != "success" {
		t.Fatalf("envelope = %#v", envelope)
	}
	assertProfilePathsAbsent(t, configRoot, cacheRoot)
}

func TestStatelessModeLogsActivationWithoutCredentialValues(t *testing.T) {
	const token = "do-not-log-this-access-token"
	t.Setenv(env.GoogleAccessToken, token)
	t.Setenv(env.LogLevel, "info")
	t.Setenv(env.NoColor, "1")

	var logs bytes.Buffer
	corelog.ConfigureCLIOutput(&logs, io.Discard)
	corelog.Init(corelog.ModeCLI)
	t.Cleanup(func() {
		corelog.SetLevel(corelog.SilentLevel)
		corelog.ConfigureCLIOutput(io.Discard, io.Discard)
	})

	cmd := NewRootForContract("test")
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(io.Discard)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"--stateless", "project", "open", "demo-project", "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("stateless project open = %v", err)
	}

	for _, marker := range []string{"stateless mode enabled", "cli.stateless", "project.open"} {
		if !strings.Contains(logs.String(), marker) {
			t.Fatalf("stateless log %q omits %q", logs.String(), marker)
		}
	}
	if strings.Contains(logs.String(), token) {
		t.Fatalf("stateless log exposes access token: %q", logs.String())
	}
	if strings.Contains(output.String(), "stateless mode enabled") {
		t.Fatalf("stateless log was written to JSON stdout: %q", output.String())
	}
}

func TestStatelessProjectDefaultsUsesLiteralTargetWithoutProfileFiles(t *testing.T) {
	requestCount := 0
	cmd, captured, configRoot, cacheRoot := newStatelessHTTPTestCommand(t, func(req *http.Request) (*http.Response, error) {
		requestCount++
		if got := req.Header.Get("Authorization"); got != "Bearer one-shot-token" {
			t.Fatalf("authorization header = %q", got)
		}
		if got := req.Header.Get("X-Goog-User-Project"); got != "quota-project" {
			t.Fatalf("quota project header = %q", got)
		}
		if !strings.Contains(req.URL.EscapedPath(), "/projects/demo/namespaces/firebase-server/remoteConfig:downloadDefaults") {
			t.Fatalf("request path = %q", req.URL.EscapedPath())
		}
		if got := req.URL.Query().Get("format"); got != "JSON" {
			t.Fatalf("format = %q, want JSON", got)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"flag":"on"}`)),
			Request:    req,
		}, nil
	})
	cmd.SetArgs([]string{"--stateless", "project", "defaults", "server@demo", "--json"})
	executed, err := cmd.ExecuteC()
	if err != nil {
		t.Fatalf("stateless project defaults = %v", err)
	}
	envelope := contract.BuildEnvelope(executed, "1.2.3", captured.Bytes(), nil)
	if envelope.Context.Profile != nil || envelope.Outcome != "success" || !strings.Contains(compactJSON(t, captured.Bytes()), `"target":"server@demo"`) {
		t.Fatalf("envelope/output = %#v/%s", envelope, captured.String())
	}
	if requestCount != 1 {
		t.Fatalf("request count = %d, want 1", requestCount)
	}
	assertProfilePathsAbsent(t, configRoot, cacheRoot)
}

func TestStatelessVersionsShowFetchesWithoutSnapshots(t *testing.T) {
	requestCount := 0
	cmd, captured, configRoot, cacheRoot := newStatelessHTTPTestCommand(t, func(req *http.Request) (*http.Response, error) {
		requestCount++
		if !strings.Contains(req.URL.EscapedPath(), "/projects/demo/namespaces/firebase-server/remoteConfig") {
			t.Fatalf("request path = %q", req.URL.EscapedPath())
		}
		if got := req.URL.Query().Get("versionNumber"); got != "7" {
			t.Fatalf("versionNumber = %q, want 7", got)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     http.Header{"ETag": []string{`"etag-7"`}},
			Body: io.NopCloser(strings.NewReader(
				`{"version":{"versionNumber":"7","updateTime":"2026-08-15T00:00:00Z","updateUser":{"email":"dev@example.com"}}}`,
			)),
			Request: req,
		}, nil
	})
	cmd.SetArgs([]string{"--stateless", "versions", "show", "server@demo", "7", "--json"})
	executed, err := cmd.ExecuteC()
	if err != nil {
		t.Fatalf("stateless versions show = %v", err)
	}
	envelope := contract.BuildEnvelope(executed, "1.2.3", captured.Bytes(), nil)
	compact := compactJSON(t, captured.Bytes())
	for _, marker := range []string{`"project_id":"server@demo"`, `"versionNumber":"7"`, `"cached":false`} {
		if !strings.Contains(compact, marker) {
			t.Fatalf("versions show output %s omits %s", captured.String(), marker)
		}
	}
	if envelope.Context.Profile != nil || envelope.Outcome != "success" || requestCount != 1 {
		t.Fatalf("envelope/request count = %#v/%d", envelope, requestCount)
	}
	assertProfilePathsAbsent(t, configRoot, cacheRoot)
}

func TestStatelessVersionsListFetchesWithoutSnapshots(t *testing.T) {
	requestCount := 0
	cmd, captured, configRoot, cacheRoot := newStatelessHTTPTestCommand(t, func(req *http.Request) (*http.Response, error) {
		requestCount++
		if !strings.Contains(req.URL.EscapedPath(), "/projects/demo/namespaces/firebase-server/remoteConfig:listVersions") {
			t.Fatalf("request path = %q", req.URL.EscapedPath())
		}
		if got := req.URL.Query().Get("pageSize"); got != "2" {
			t.Fatalf("pageSize = %q, want 2", got)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     make(http.Header),
			Body: io.NopCloser(strings.NewReader(
				`{"versions":[{"versionNumber":"9","updateTime":"2026-08-20T00:00:00Z"},{"versionNumber":"8","updateTime":"2026-08-19T00:00:00Z"}]}`,
			)),
			Request: req,
		}, nil
	})
	cmd.SetArgs([]string{"--stateless", "versions", "list", "server@demo", "--limit", "2", "--json"})
	executed, err := cmd.ExecuteC()
	if err != nil {
		t.Fatalf("stateless versions list = %v", err)
	}
	compact := compactJSON(t, captured.Bytes())
	for _, marker := range []string{`"versionNumber":"9"`, `"versionNumber":"8"`, `"current":false`, `"cached":false`} {
		if !strings.Contains(compact, marker) {
			t.Fatalf("versions list output %s omits %s", captured.String(), marker)
		}
	}
	envelope := contract.BuildEnvelope(executed, "1.2.3", captured.Bytes(), nil)
	if envelope.Context.Profile != nil || envelope.Outcome != "success" || requestCount != 1 {
		t.Fatalf("envelope/request count = %#v/%d", envelope, requestCount)
	}
	assertProfilePathsAbsent(t, configRoot, cacheRoot)
}

func TestStatelessConditionsListFetchesWithoutCacheOrDrafts(t *testing.T) {
	requestCount := 0
	cmd, captured, configRoot, cacheRoot := newStatelessHTTPTestCommand(t, func(req *http.Request) (*http.Response, error) {
		requestCount++
		var body string
		switch requestCount {
		case 1:
			if !strings.HasSuffix(req.URL.EscapedPath(), "/projects/demo/remoteConfig:listVersions") || req.URL.Query().Get("pageSize") != "1" {
				t.Fatalf("version request = %s", req.URL.String())
			}
			body = `{"versions":[{"versionNumber":"7","updateTime":"2026-08-20T00:00:00Z"}]}`
		case 2:
			if !strings.HasSuffix(req.URL.EscapedPath(), "/projects/demo/remoteConfig") || req.URL.Query().Get("versionNumber") != "7" {
				t.Fatalf("Remote Config request = %s", req.URL.String())
			}
			body = `{"conditions":[{"name":"android","expression":"device.os == 'android'","tagColor":"BLUE"}],"parameters":{"flag":{"defaultValue":{"value":"off"},"conditionalValues":{"android":{"value":"on"}}}},"version":{"versionNumber":"7"}}`
		default:
			t.Fatalf("unexpected request %s", req.URL.String())
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     http.Header{"ETag": []string{`"etag-7"`}},
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    req,
		}, nil
	})
	cmd.SetArgs([]string{"--stateless", "conditions", "list", "demo", "--filter", "=android", "--search", "android", "--expr", "priority == 1", "--json"})
	executed, err := cmd.ExecuteC()
	if err != nil {
		t.Fatalf("stateless conditions list = %v", err)
	}
	compact := compactJSON(t, captured.Bytes())
	for _, marker := range []string{`"name":"android"`, `"priority":1`, `"parameter":"flag"`} {
		if !strings.Contains(compact, marker) {
			t.Fatalf("conditions list output %s omits %s", captured.String(), marker)
		}
	}
	envelope := contract.BuildEnvelope(executed, "1.2.3", captured.Bytes(), nil)
	if envelope.Context.Profile != nil || envelope.Outcome != "success" || requestCount != 2 {
		t.Fatalf("envelope/request count = %#v/%d", envelope, requestCount)
	}
	assertProfilePathsAbsent(t, configRoot, cacheRoot)
}

func TestStatelessGetFetchesLiteralTargetWithoutCacheOrDrafts(t *testing.T) {
	requestCount := 0
	cmd, captured, configRoot, cacheRoot := newStatelessHTTPTestCommand(t, func(req *http.Request) (*http.Response, error) {
		requestCount++
		var body string
		switch requestCount {
		case 1:
			if !strings.HasSuffix(req.URL.EscapedPath(), "/projects/demo/namespaces/firebase-server/remoteConfig:listVersions") || req.URL.Query().Get("pageSize") != "1" {
				t.Fatalf("version request = %s", req.URL.String())
			}
			body = `{"versions":[{"versionNumber":"11","updateTime":"2026-08-20T00:00:00Z"}]}`
		case 2:
			if !strings.HasSuffix(req.URL.EscapedPath(), "/projects/demo/namespaces/firebase-server/remoteConfig") || req.URL.Query().Get("versionNumber") != "11" {
				t.Fatalf("Remote Config request = %s", req.URL.String())
			}
			body = `{"parameterGroups":{"support":{"parameters":{"support_chat_enabled":{"description":"Support chat","defaultValue":{"value":"true"},"valueType":"BOOLEAN"}}}},"version":{"versionNumber":"11"}}`
		default:
			t.Fatalf("unexpected request %s", req.URL.String())
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     http.Header{"ETag": []string{`"etag-11"`}},
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    req,
		}, nil
	})
	cmd.SetArgs([]string{"--stateless", "get", "support_chat_enabled", "--project", "server@=demo", "--search", "support", "--expr", `group == "support"`, "--json"})
	executed, err := cmd.ExecuteC()
	if err != nil {
		t.Fatalf("stateless get = %v", err)
	}
	compact := compactJSON(t, captured.Bytes())
	for _, marker := range []string{`"project_id":"server@demo"`, `"key":"support_chat_enabled"`, `"version":"11"`, `"cached_at":null`, `"status":"fetch"`} {
		if !strings.Contains(compact, marker) {
			t.Fatalf("get output %s omits %s", captured.String(), marker)
		}
	}
	envelope := contract.BuildEnvelope(executed, "1.2.3", captured.Bytes(), nil)
	if envelope.Context.Profile != nil || envelope.Outcome != "success" || requestCount != 2 {
		t.Fatalf("envelope/request count = %#v/%d", envelope, requestCount)
	}
	assertProfilePathsAbsent(t, configRoot, cacheRoot)
}

func TestStatelessGetDiscoversAndFiltersRemoteProjects(t *testing.T) {
	var mu sync.Mutex
	requests := make(map[string]int)
	cmd, captured, configRoot, cacheRoot := newStatelessHTTPTestCommand(t, func(req *http.Request) (*http.Response, error) {
		mu.Lock()
		requests[req.URL.EscapedPath()]++
		mu.Unlock()

		var body string
		switch req.URL.EscapedPath() {
		case "/v1/projects":
			if req.URL.Query().Get("pageSize") != "1000" {
				t.Fatalf("project discovery request = %s", req.URL.String())
			}
			body = `{"projects":[{"projectId":"northstar-wallet","projectNumber":"123","name":"projects/123","lifecycleState":"ACTIVE"}]}`
		case "/v3/projects/northstar-wallet":
			body = `{"projectId":"northstar-wallet","displayName":"Northstar Wallet","state":"ACTIVE"}`
		case "/v1/projects/northstar-wallet/remoteConfig:listVersions":
			body = `{"versions":[{"versionNumber":"12","updateTime":"2026-08-20T00:00:00Z"}]}`
		case "/v1/projects/northstar-wallet/remoteConfig":
			if req.URL.Query().Get("versionNumber") != "12" {
				t.Fatalf("Remote Config request = %s", req.URL.String())
			}
			body = `{"parameters":{"flag":{"defaultValue":{"value":"on"}}},"version":{"versionNumber":"12"}}`
		default:
			t.Fatalf("unexpected request %s", req.URL.String())
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     http.Header{"ETag": []string{`"etag-12"`}},
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    req,
		}, nil
	})
	cmd.SetArgs([]string{"--stateless", "get", "flag", "--project", "/Northstar", "--json"})
	executed, err := cmd.ExecuteC()
	if err != nil {
		t.Fatalf("stateless discovered get = %v", err)
	}
	compact := compactJSON(t, captured.Bytes())
	for _, marker := range []string{`"project":"Northstar Wallet"`, `"project_id":"northstar-wallet"`, `"key":"flag"`} {
		if !strings.Contains(compact, marker) {
			t.Fatalf("get output %s omits %s", captured.String(), marker)
		}
	}
	if envelope := contract.BuildEnvelope(executed, "1.2.3", captured.Bytes(), nil); envelope.Context.Profile != nil || envelope.Outcome != "success" {
		t.Fatalf("envelope = %#v", envelope)
	}
	for path, want := range map[string]int{
		"/v1/projects":                  1,
		"/v3/projects/northstar-wallet": 1,
		"/v1/projects/northstar-wallet/remoteConfig:listVersions": 1,
		"/v1/projects/northstar-wallet/remoteConfig":              1,
	} {
		if requests[path] != want {
			t.Fatalf("requests[%q] = %d, want %d; all requests: %#v", path, requests[path], want, requests)
		}
	}
	assertProfilePathsAbsent(t, configRoot, cacheRoot)
}

func TestStatelessProjectsListFiltersWithDirectRemoteConfigExpression(t *testing.T) {
	var mu sync.Mutex
	requests := make(map[string]int)
	cmd, captured, configRoot, cacheRoot := newStatelessHTTPTestCommand(t, func(req *http.Request) (*http.Response, error) {
		mu.Lock()
		requests[req.URL.EscapedPath()]++
		mu.Unlock()

		var body string
		switch req.URL.EscapedPath() {
		case "/v1/projects":
			body = `{"projects":[{"projectId":"northstar-wallet","projectNumber":"123","name":"projects/123","lifecycleState":"ACTIVE"},{"projectId":"other-project","projectNumber":"456","name":"projects/456","lifecycleState":"ACTIVE"}]}`
		case "/v3/projects/northstar-wallet":
			body = `{"name":"projects/123","projectId":"northstar-wallet","displayName":"Northstar Wallet","state":"ACTIVE"}`
		case "/v3/projects/other-project":
			body = `{"name":"projects/456","projectId":"other-project","displayName":"Other Project","state":"ACTIVE"}`
		case "/v1/projects/northstar-wallet/remoteConfig:listVersions":
			body = `{"versions":[{"versionNumber":"12","updateTime":"2026-08-20T00:00:00Z"}]}`
		case "/v1/projects/northstar-wallet/remoteConfig":
			if req.URL.Query().Get("versionNumber") != "12" {
				t.Fatalf("Remote Config request = %s", req.URL.String())
			}
			body = `{"parameters":{"flag":{"defaultValue":{"value":"on"}}},"version":{"versionNumber":"12"}}`
		default:
			t.Fatalf("unexpected request %s", req.URL.String())
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     http.Header{"ETag": []string{`"etag-12"`}},
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    req,
		}, nil
	})
	cmd.SetArgs([]string{
		"--stateless", "projects", "list", "--filter", "=Northstar Wallet",
		"--expr", `project_id == "northstar-wallet" && parameters["flag"].default == "on"`, "--json",
	})
	executed, err := cmd.ExecuteC()
	if err != nil {
		t.Fatalf("stateless projects list expression = %v", err)
	}
	compact := compactJSON(t, captured.Bytes())
	for _, marker := range []string{`"project":"Northstar Wallet"`, `"project_id":"northstar-wallet"`} {
		if !strings.Contains(compact, marker) {
			t.Fatalf("projects list output %s omits %s", captured.String(), marker)
		}
	}
	if envelope := contract.BuildEnvelope(executed, "1.2.3", captured.Bytes(), nil); envelope.Context.Profile != nil || envelope.Outcome != "success" {
		t.Fatalf("envelope = %#v", envelope)
	}
	for path, want := range map[string]int{
		"/v1/projects":                                            1,
		"/v3/projects/northstar-wallet":                           1,
		"/v3/projects/other-project":                              1,
		"/v1/projects/northstar-wallet/remoteConfig:listVersions": 1,
		"/v1/projects/northstar-wallet/remoteConfig":              1,
	} {
		if requests[path] != want {
			t.Fatalf("requests[%q] = %d, want %d; all requests: %#v", path, requests[path], want, requests)
		}
	}
	assertProfilePathsAbsent(t, configRoot, cacheRoot)
}

func TestStatelessVersionsExportWritesDestinationWithoutSnapshots(t *testing.T) {
	root := t.TempDir()
	destination := filepath.Join(root, "output", "version.json")
	requestCount := 0
	cmd, captured, configRoot, cacheRoot := newStatelessHTTPTestCommand(t, func(req *http.Request) (*http.Response, error) {
		requestCount++
		if !strings.Contains(req.URL.EscapedPath(), "/projects/demo/remoteConfig") {
			t.Fatalf("request path = %q", req.URL.EscapedPath())
		}
		if got := req.URL.Query().Get("versionNumber"); got != "8" {
			t.Fatalf("versionNumber = %q, want 8", got)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     http.Header{"ETag": []string{`"etag-8"`}},
			Body:       io.NopCloser(strings.NewReader(`{"version":{"versionNumber":"8"},"parameters":{"flag":{"defaultValue":{"value":"\u003con\u003e"}}}}`)),
			Request:    req,
		}, nil
	})
	cmd.SetArgs([]string{"--stateless", "versions", "export", "demo", "8", "--to", destination, "--json"})
	executed, err := cmd.ExecuteC()
	if err != nil {
		t.Fatalf("stateless versions export = %v", err)
	}
	written, err := os.ReadFile(destination)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(written), `"value":"<on>"`) {
		t.Fatalf("destination = %s", written)
	}
	envelope := contract.BuildEnvelope(executed, "1.2.3", captured.Bytes(), nil)
	if envelope.Context.Profile != nil || envelope.Outcome != "success" || requestCount != 1 {
		t.Fatalf("envelope/request count = %#v/%d", envelope, requestCount)
	}
	assertProfilePathsAbsent(t, configRoot, cacheRoot)
}

func newStatelessHTTPTestCommand(t *testing.T, roundTrip appRoundTripFunc) (*cobra.Command, *bytes.Buffer, string, string) {
	t.Helper()
	root := t.TempDir()
	configRoot := filepath.Join(root, "config")
	cacheRoot := filepath.Join(root, "cache")
	t.Setenv(env.ConfigDir, configRoot)
	t.Setenv(env.CacheDir, cacheRoot)
	t.Setenv(env.Profile, "../invalid")
	t.Setenv(env.GoogleAccessToken, "one-shot-token")
	t.Setenv(env.GoogleCloudQuotaProject, "quota-project")
	if err := config.SetProfileOverride(""); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = config.SetProfileOverride("")
		config.SetLocalConfigDisabled(false)
	})
	originalTransport := http.DefaultTransport
	http.DefaultTransport = appRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		if got := req.Header.Get("Authorization"); got != "Bearer one-shot-token" {
			t.Fatalf("authorization header = %q", got)
		}
		if got := req.Header.Get("X-Goog-User-Project"); got != "quota-project" {
			t.Fatalf("quota project header = %q", got)
		}
		return roundTrip(req)
	})
	t.Cleanup(func() { http.DefaultTransport = originalTransport })

	svc, err := core.NewService(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	cmd := newRootCommandWithOfflineInit(svc, "1.2.3", "abc123", "2026-06-14", func(context.Context, bool) {})
	captured := &bytes.Buffer{}
	cmd.SetOut(captured)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	return cmd, captured, configRoot, cacheRoot
}

func assertProfilePathsAbsent(t *testing.T, paths ...string) {
	t.Helper()
	for _, path := range paths {
		if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("profile path %s stat = %v, want not found", path, err)
		}
	}
}

func compactJSON(t *testing.T, raw []byte) string {
	t.Helper()
	var compact bytes.Buffer
	if err := json.Compact(&compact, raw); err != nil {
		t.Fatalf("compact JSON: %v", err)
	}
	return compact.String()
}

func TestStatelessProjectExportWithoutJSONBypassesProfileRegistry(t *testing.T) {
	root := t.TempDir()
	configRoot := filepath.Join(root, "config")
	cacheRoot := filepath.Join(root, "cache")
	t.Setenv(env.ConfigDir, configRoot)
	t.Setenv(env.CacheDir, cacheRoot)
	t.Setenv(env.Profile, "../invalid")
	t.Setenv(env.GoogleAccessToken, "one-shot-token")
	t.Setenv(env.Offline, "1")
	if err := config.SetProfileOverride(""); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = config.SetProfileOverride("")
		firebase.SetOfflineMode(false)
	})

	svc, err := core.NewService(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	cmd := newRootCommand(svc, "1.2.3", "abc123", "2026-06-14")
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"--stateless", "project", "export", "my-project"})
	if err := cmd.Execute(); !errors.Is(err, firebase.ErrOffline) {
		t.Fatalf("human stateless export = %v, want direct Firebase offline error", err)
	}
	for _, path := range []string{configRoot, cacheRoot} {
		if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("profile path %s stat = %v, want not found", path, err)
		}
	}
}

func TestStatelessProjectExportProtectsExistingDestinationBeforeNetwork(t *testing.T) {
	root := t.TempDir()
	destination := filepath.Join(root, "remote-config.json")
	if err := os.WriteFile(destination, []byte("original"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv(env.ConfigDir, filepath.Join(root, "config"))
	t.Setenv(env.CacheDir, filepath.Join(root, "cache"))
	t.Setenv(env.GoogleAccessToken, "one-shot-token")
	if err := config.SetProfileOverride(""); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = config.SetProfileOverride("") })

	svc, err := core.NewService(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	cmd := newRootCommandWithOfflineInit(svc, "1.2.3", "abc123", "2026-06-14", func(context.Context, bool) {})
	var captured bytes.Buffer
	cmd.SetOut(&captured)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"--stateless", "project", "export", "demo", "--to", destination, "--json"})
	executed, err := cmd.ExecuteC()
	envelope := contract.BuildEnvelope(executed, "1.2.3", captured.Bytes(), err)
	if envelope.Context.Profile != nil || envelope.ExitCode != 10 || len(envelope.Errors) != 1 || envelope.Errors[0].Code != "interaction.required" {
		t.Fatalf("envelope = %#v", envelope)
	}
	written, readErr := os.ReadFile(destination)
	if readErr != nil || string(written) != "original" {
		t.Fatalf("destination = %q, %v", written, readErr)
	}
}

func TestStatelessProjectExportRejectsMalformedAccessTokenWithoutProfileFiles(t *testing.T) {
	root := t.TempDir()
	configRoot := filepath.Join(root, "config")
	cacheRoot := filepath.Join(root, "cache")
	t.Setenv(env.ConfigDir, configRoot)
	t.Setenv(env.CacheDir, cacheRoot)
	t.Setenv(env.GoogleAccessToken, "secret token")
	if err := config.SetProfileOverride(""); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = config.SetProfileOverride("") })

	svc, err := core.NewService(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	cmd := newRootCommandWithOfflineInit(svc, "1.2.3", "abc123", "2026-06-14", func(context.Context, bool) {})
	var captured bytes.Buffer
	cmd.SetOut(&captured)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"--stateless", "project", "export", "demo", "--json"})
	executed, err := cmd.ExecuteC()
	envelope := contract.BuildEnvelope(executed, "1.2.3", captured.Bytes(), err)
	if envelope.Context.Profile != nil || envelope.ExitCode != 4 || len(envelope.Errors) != 1 || envelope.Errors[0].Code != "auth.credentials_invalid" {
		t.Fatalf("envelope = %#v", envelope)
	}
	if strings.Contains(envelope.Errors[0].Message, "secret token") {
		t.Fatal("authentication problem contains the access token")
	}
	for _, path := range []string{configRoot, cacheRoot} {
		if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("profile path %s stat = %v, want not found", path, err)
		}
	}
}

func TestAccessTokenEnvironmentDoesNotChangeStatefulProjectExport(t *testing.T) {
	root := t.TempDir()
	t.Setenv(env.ConfigDir, filepath.Join(root, "config"))
	t.Setenv(env.CacheDir, filepath.Join(root, "cache"))
	t.Setenv(env.Profile, "../invalid")
	t.Setenv(env.GoogleAccessToken, "one-shot-token")
	if err := config.SetProfileOverride(""); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = config.SetProfileOverride("") })

	svc, err := core.NewService(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	cmd := newRootCommandWithOfflineInit(svc, "1.2.3", "abc123", "2026-06-14", func(context.Context, bool) {})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"project", "export", "my-project"})
	executed, err := cmd.ExecuteC()
	if err == nil || !strings.Contains(err.Error(), "select profile") {
		t.Fatalf("stateful project export error = %v, want profile selection failure", err)
	}
	if policy := core.ExecutionPolicyFromContext(executed.Context()); policy != core.StatefulExecutionPolicy() {
		t.Fatalf("execution policy = %#v, want stateful", policy)
	}
}

func TestStatelessProjectExportRequiresAccessTokenWithoutProfileFiles(t *testing.T) {
	root := t.TempDir()
	configRoot := filepath.Join(root, "config")
	cacheRoot := filepath.Join(root, "cache")
	t.Setenv(env.ConfigDir, configRoot)
	t.Setenv(env.CacheDir, cacheRoot)
	t.Setenv(env.Profile, "../invalid")
	t.Setenv(env.GoogleAccessToken, "")
	if err := config.SetProfileOverride(""); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = config.SetProfileOverride("") })

	cmd := newRootCommandWithOfflineInit(nil, "1.2.3", "abc123", "2026-06-14", func(context.Context, bool) {})
	var captured bytes.Buffer
	cmd.SetOut(&captured)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"--stateless", "project", "export", "my-project", "--json"})
	executed, err := cmd.ExecuteC()
	envelope := contract.BuildEnvelope(executed, "1.2.3", captured.Bytes(), err)
	if envelope.Context.Profile != nil || envelope.ExitCode != 4 || len(envelope.Errors) != 1 || envelope.Errors[0].Code != "auth.configuration_invalid" {
		t.Fatalf("envelope = %#v", envelope)
	}
	if !strings.Contains(envelope.Errors[0].Message, env.GoogleAccessToken) {
		t.Fatalf("problem message = %q", envelope.Errors[0].Message)
	}
	for _, path := range []string{configRoot, cacheRoot} {
		if _, statErr := os.Stat(path); !errors.Is(statErr, os.ErrNotExist) {
			t.Fatalf("profile path %s stat = %v, want not found", path, statErr)
		}
	}
}

func TestStatelessGetStdinDoesNotRequireAccessToken(t *testing.T) {
	root := t.TempDir()
	configRoot := filepath.Join(root, "config")
	cacheRoot := filepath.Join(root, "cache")
	t.Setenv(env.ConfigDir, configRoot)
	t.Setenv(env.CacheDir, cacheRoot)
	t.Setenv(env.Profile, "../invalid")
	t.Setenv(env.GoogleAccessToken, "")
	if err := config.SetProfileOverride(""); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = config.SetProfileOverride("")
		config.SetLocalConfigDisabled(false)
	})
	input, err := os.CreateTemp(root, "remote-config-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = input.Close() }()
	if _, err := input.WriteString(`{"parameters":{"flag":{"defaultValue":{"value":"on"}}}}`); err != nil {
		t.Fatal(err)
	}
	if _, err := input.Seek(0, 0); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCommandWithOfflineInit(nil, "1.2.3", "abc123", "2026-06-14", func(context.Context, bool) {})
	var captured bytes.Buffer
	cmd.SetIn(input)
	cmd.SetOut(&captured)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"--stateless", "get", "flag", "--json"})
	executed, err := cmd.ExecuteC()
	if err != nil {
		t.Fatalf("stateless stdin get = %v", err)
	}
	envelope := contract.BuildEnvelope(executed, "1.2.3", captured.Bytes(), nil)
	if envelope.Context.Profile != nil || envelope.Outcome != "success" || !strings.Contains(compactJSON(t, captured.Bytes()), `"key":"flag"`) {
		t.Fatalf("envelope/output = %#v/%s", envelope, captured.String())
	}
	assertProfilePathsAbsent(t, configRoot, cacheRoot)
}

func TestStatelessModeRejectsUnsupportedCommandAndExplicitProfile(t *testing.T) {
	t.Setenv(env.GoogleAccessToken, "one-shot-token")
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "unsupported command", args: []string{"--stateless", "conditions", "add", "my-project", "example", "--expression", "true", "--json"}, want: "not supported by conditions add"},
		{name: "explicit profile", args: []string{"--stateless", "--profile", "personal", "project", "export", "my-project", "--json"}, want: "--profile cannot be used"},
		{name: "cached version", args: []string{"--stateless", "versions", "show", "my-project", "7", "--cached", "--json"}, want: "--cached cannot be used"},
		{name: "conditions update", args: []string{"--stateless", "conditions", "list", "my-project", "--update", "--json"}, want: "--update cannot be used"},
		{name: "get update", args: []string{"--stateless", "get", "--project", "my-project", "--update", "--json"}, want: "--update cannot be used"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd := newRootCommandWithOfflineInit(nil, "1.2.3", "abc123", "2026-06-14", func(context.Context, bool) {})
			var captured bytes.Buffer
			cmd.SetOut(&captured)
			cmd.SetErr(&bytes.Buffer{})
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true
			cmd.SetArgs(test.args)
			executed, err := cmd.ExecuteC()
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Execute() error = %v, want %q", err, test.want)
			}
			envelope := contract.BuildEnvelope(executed, "1.2.3", captured.Bytes(), err)
			if envelope.Context.Profile != nil || envelope.ExitCode != 2 || len(envelope.Errors) != 1 || envelope.Errors[0].Code != "argument.invalid" {
				t.Fatalf("envelope = %#v", envelope)
			}
		})
	}
}

func TestStatelessProjectExportRequiresLiteralProjectID(t *testing.T) {
	t.Setenv(env.GoogleAccessToken, "one-shot-token")
	cmd := newRootCommandWithOfflineInit(nil, "1.2.3", "abc123", "2026-06-14", func(context.Context, bool) {})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"--stateless", "project", "export", "=my-project", "--json"})
	if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "must not use a filter prefix") {
		t.Fatalf("Execute() error = %v, want literal project ID failure", err)
	}
}

type appRoundTripFunc func(*http.Request) (*http.Response, error)

func (f appRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestRootCommandShowsAuthSetupGuidanceBeforeUsage(t *testing.T) {
	root := t.TempDir()
	t.Setenv(env.ConfigDir, filepath.Join(root, "config"))
	t.Setenv(env.CacheDir, filepath.Join(root, "cache"))
	t.Setenv(env.Profile, "")
	if err := config.SetProfileOverride(""); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = config.SetProfileOverride("") })
	if err := config.SwitchProfile("test"); err != nil {
		t.Fatal(err)
	}
	svc, err := core.NewService(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	cmd := newRootCommandWithOfflineInit(svc, "1.2.3", "abc123", "2026-06-14", func(context.Context, bool) {})
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs([]string{"projects", "list"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("projects list without auth = nil, want error")
	}

	got := output.String()
	errorAt := strings.Index(got, "Error: read auth config:")
	hintAt := strings.Index(got, "Set up authentication by running `fbrcm` for guided setup")
	usageAt := strings.Index(got, "Usage:\n  fbrcm projects list")
	if errorAt < 0 || hintAt < 0 || usageAt < 0 || errorAt >= hintAt || hintAt >= usageAt {
		t.Fatalf("projects list output does not show auth setup guidance between error and usage:\n%s", got)
	}
}

func TestIsProfileCommand(t *testing.T) {
	root := &cobra.Command{Use: "fbrcm"}
	profile := &cobra.Command{Use: "profile"}
	list := &cobra.Command{Use: "list"}
	projects := &cobra.Command{Use: "projects"}
	root.AddCommand(profile, projects)
	profile.AddCommand(list)

	if !isProfileCommand(profile) {
		t.Fatalf("profile command not recognized")
	}
	if !isProfileCommand(list) {
		t.Fatalf("profile subcommand not recognized")
	}
	if isProfileCommand(projects) {
		t.Fatalf("projects command recognized as profile")
	}
}

func TestCompletedCommandErrorHonorsExpiredDeadline(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := completedCommandError(ctx, nil); !errors.Is(err, context.Canceled) {
		t.Fatalf("completedCommandError = %v, want context cancellation", err)
	}
	commandErr := errors.New("command failed")
	if err := completedCommandError(ctx, commandErr); !errors.Is(err, commandErr) {
		t.Fatalf("completedCommandError = %v, want original command error", err)
	}
}

func TestExpiredContextStopsCommandBeforeRun(t *testing.T) {
	root := NewRootForContract("test")
	var output bytes.Buffer
	root.SetOut(&output)
	root.SilenceUsage = true
	root.SetArgs([]string{"capabilities", "--json"})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	root.SetContext(shared.WithMachineState(ctx))
	_, err := root.ExecuteC()
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("ExecuteC error = %v, want context cancellation", err)
	}
	if output.Len() != 0 {
		t.Fatalf("command ran after cancellation: %s", output.String())
	}
}

func commandNames(cmd *cobra.Command) []string {
	names := make([]string, 0, len(cmd.Commands()))
	for _, child := range cmd.Commands() {
		names = append(names, child.Name())
	}
	sort.Strings(names)
	return names
}
