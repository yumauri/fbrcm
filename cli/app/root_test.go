package app

import (
	"bytes"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"

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
	if got, want := commandNames(first), []string{"add", "auth", "cache", "conditions", "config", "delete", "doctor", "draft", "get", "profile", "project", "projects", "update", "versions"}; !reflect.DeepEqual(got, want) {
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

	if !reflect.DeepEqual(counts, []int{14, 14, 14}) {
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

	cmd := newRootCommand(nil, "1.2.3", "abc123", "2026-06-14")
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
