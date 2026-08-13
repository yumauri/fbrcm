package versions

import (
	"errors"
	"strings"
	"testing"

	cmdtest "github.com/yumauri/fbrcm/cli/commands/testutil"
	"github.com/yumauri/fbrcm/cli/shared"
)

func TestNewCommandStructure(t *testing.T) {
	cmd := New(nil)
	cmdtest.AssertSubcommands(t, cmd, "diff", "export", "list", "restore", "rollback", "show")
	for _, flag := range []string{"limit", "all", "before", "since", "until", "cached", "json"} {
		cmdtest.AssertNestedFlag(t, cmd, []string{"list"}, flag)
	}
	for _, flag := range []string{"cached", "json"} {
		cmdtest.AssertNestedFlag(t, cmd, []string{"show"}, flag)
	}
	show := cmdtest.FindCommand(t, cmd, "show")
	if show.Flags().Lookup("config") != nil {
		t.Fatal("versions show still exposes removed --config flag")
	}
	for _, flag := range []string{"filter", "search", "group", "expr", "parameters", "conditions", "cached", "json", "side-by-side"} {
		cmdtest.AssertNestedFlag(t, cmd, []string{"diff"}, flag)
	}
	diff := cmdtest.FindCommand(t, cmd, "diff")
	if shorthand := diff.Flags().Lookup("side-by-side").Shorthand; shorthand != "" {
		t.Fatalf("--side-by-side shorthand = %q, want none", shorthand)
	}
	for _, flag := range []string{"to", "cached", "yes"} {
		cmdtest.AssertNestedFlag(t, cmd, []string{"export"}, flag)
	}
	for _, name := range []string{"rollback", "restore"} {
		for _, flag := range []string{"dry-run", "yes", "json"} {
			cmdtest.AssertNestedFlag(t, cmd, []string{name}, flag)
		}
	}
	cmdtest.AssertNestedFlag(t, cmd, []string{"restore"}, "change-note")
	if rollback := cmdtest.FindCommand(t, cmd, "rollback"); rollback.Flags().Lookup("change-note") != nil {
		t.Fatal("versions rollback unexpectedly exposes --change-note")
	}
}

func TestVersionDiffRejectsJSONWithSideBySide(t *testing.T) {
	cmd := New(nil)
	cmd.SetArgs([]string{"diff", "demo", "1", "--json", "--side-by-side"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "if any flags in the group") {
		t.Fatalf("Execute() error = %v, want mutually exclusive flag error", err)
	}
}

func TestValidateBeforeVersionRequiresCanonicalPositiveDecimal(t *testing.T) {
	for _, value := range []string{"1", "10", strings.Repeat("9", 100)} {
		if err := validateBeforeVersion(value); err != nil {
			t.Errorf("validateBeforeVersion(%q) = %v, want nil", value, err)
		}
	}
	for _, value := range []string{"", "0", "01", "+1", "-1", " 1", "1 ", "1.0", "one"} {
		err := validateBeforeVersion(value)
		var argumentErr *shared.ArgumentError
		if !errors.As(err, &argumentErr) {
			t.Errorf("validateBeforeVersion(%q) = %v, want ArgumentError", value, err)
		}
	}
}

func TestVersionsListRejectsExplicitEmptyBefore(t *testing.T) {
	cmd := New(nil)
	cmd.SetArgs([]string{"list", "demo", "--before="})
	err := cmd.Execute()
	var argumentErr *shared.ArgumentError
	if !errors.As(err, &argumentErr) {
		t.Fatalf("Execute() error = %v, want ArgumentError", err)
	}
}
