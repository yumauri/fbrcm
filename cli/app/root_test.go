package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
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
	if got, want := commandNames(first), []string{"add", "auth", "cache", "conditions", "config", "delete", "doctor", "draft", "duplicate", "get", "groups", "profile", "project", "projects", "update", "versions"}; !reflect.DeepEqual(got, want) {
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

	if !reflect.DeepEqual(counts, []int{16, 16, 16}) {
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
}

func TestRootCommandSkipsConnectivityProbeForHelpAndVersion(t *testing.T) {
	for _, args := range [][]string{{"--help"}, {"help"}, {"--version"}} {
		t.Run(strings.Join(args, "_"), func(t *testing.T) {
			calls := 0
			cmd := newRootCommandWithOfflineInit(nil, "1.2.3", "abc123", "2026-06-14", func() { calls++ })
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
	cmd := newRootCommandWithOfflineInit(nil, "1.2.3", "abc123", "2026-06-14", func() { calls++ })
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

func TestRootCommandProbesBeforeExecution(t *testing.T) {
	root := t.TempDir()
	t.Setenv(env.ConfigDir, filepath.Join(root, "config"))
	t.Setenv(env.CacheDir, filepath.Join(root, "cache"))
	t.Setenv(env.Profile, "")
	if err := config.SetProfileOverride(""); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = config.SetProfileOverride("") })

	calls := 0
	cmd := newRootCommandWithOfflineInit(nil, "1.2.3", "abc123", "2026-06-14", func() { calls++ })
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"profile"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute profile: %v", err)
	}
	if calls != 1 {
		t.Fatalf("connectivity probe calls = %d, want 1", calls)
	}
}

func TestCommandExitCodeHonorsDiffContract(t *testing.T) {
	cmd := &cobra.Command{Use: "diff"}
	shared.AddDiffExitCodeFlag(cmd)
	original := fmt.Errorf("failed")
	if got := commandExitCode(cmd, original); got != 1 {
		t.Fatalf("default error exit code = %d, want 1", got)
	}
	if err := cmd.Flags().Set("exit-code", "true"); err != nil {
		t.Fatal(err)
	}
	if got := commandExitCode(cmd, original); got != 2 {
		t.Fatalf("diff operational error exit code = %d, want 2", got)
	}
	explicit := shared.WithExitCode(nil, 1)
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
	shared.AddDiffExitCodeFlag(diff)
	root.AddCommand(diff)
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"diff", "--exit-code"})
	executed, err := root.ExecuteC()
	if err == nil {
		t.Fatal("argument error is nil")
	}
	if got := commandExitCode(executed, err); got != 2 {
		t.Fatalf("argument error exit code = %d, want 2", got)
	}
}

func TestCommandExitCodeFindsExitFlagAfterUnknownFlag(t *testing.T) {
	cmd := &cobra.Command{Use: "diff"}
	shared.AddDiffExitCodeFlag(cmd)
	if got := commandExitCode(cmd, fmt.Errorf("unknown flag"), "diff", "--bad-flag", "--exit-code"); got != 2 {
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

	cmd := newRootCommandWithOfflineInit(nil, "1.2.3", "abc123", "2026-06-14", func() {})
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

	cmd := newRootCommandWithOfflineInit(svc, "1.2.3", "abc123", "2026-06-14", func() {})
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

func commandNames(cmd *cobra.Command) []string {
	names := make([]string, 0, len(cmd.Commands()))
	for _, child := range cmd.Commands() {
		names = append(names, child.Name())
	}
	sort.Strings(names)
	return names
}
