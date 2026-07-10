package app

import (
	"reflect"
	"sort"
	"testing"

	"github.com/spf13/cobra"
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
	if got, want := commandNames(first), []string{"add", "auth", "cache", "config", "delete", "get", "profile", "project", "projects", "update"}; !reflect.DeepEqual(got, want) {
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

	if !reflect.DeepEqual(counts, []int{10, 10, 10}) {
		t.Fatalf("command counts = %#v, want stable counts without accumulation", counts)
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
