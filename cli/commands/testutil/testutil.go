package testutil

import (
	"testing"

	"github.com/spf13/cobra"
)

// AssertCommandStructure checks command Use string and top-level flag names.
func AssertCommandStructure(t *testing.T, cmd *cobra.Command, wantUse string, flags ...string) {
	t.Helper()
	if cmd.Use != wantUse {
		t.Fatalf("Use = %q, want %q", cmd.Use, wantUse)
	}
	for _, flag := range flags {
		if cmd.Flags().Lookup(flag) == nil {
			t.Fatalf("flag %q missing", flag)
		}
	}
}

// AssertSubcommands checks that root has the expected subcommand names.
func AssertSubcommands(t *testing.T, cmd *cobra.Command, want ...string) {
	t.Helper()
	got := make(map[string]bool, len(cmd.Commands()))
	for _, sub := range cmd.Commands() {
		got[sub.Name()] = true
	}
	for _, name := range want {
		if !got[name] {
			t.Fatalf("missing subcommand %q", name)
		}
	}
}

// AssertFlag checks that a subcommand exposes the named flag.
func AssertFlag(t *testing.T, root *cobra.Command, commandName, flagName string) {
	t.Helper()
	AssertNestedFlag(t, root, []string{commandName}, flagName)
}

// AssertNestedFlag checks that a nested command path exposes the named flag.
func AssertNestedFlag(t *testing.T, root *cobra.Command, commandPath []string, flagName string) {
	t.Helper()
	cmd := FindCommand(t, root, commandPath...)
	if cmd.Flags().Lookup(flagName) == nil {
		t.Fatalf("%s flag %q missing", cmd.CommandPath(), flagName)
	}
}

// FindCommand resolves a command path from root.
func FindCommand(t *testing.T, root *cobra.Command, commandPath ...string) *cobra.Command {
	t.Helper()
	cmd, _, err := root.Find(commandPath)
	if err != nil {
		t.Fatalf("find %v: %v", commandPath, err)
	}
	return cmd
}
