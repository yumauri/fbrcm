package project

import (
	"testing"

	cmdtest "github.com/yumauri/fbrcm/cli/commands/testutil"
)

func TestNewCommandStructure(t *testing.T) {
	cmd := New(nil)
	cmdtest.AssertSubcommands(t, cmd, "export", "import", "versions")
	cmdtest.AssertFlag(t, cmd, "export", "to")
	for _, flag := range []string{"from", "group", "filter", "expr", "search", "dry-run", "draft", "remove-all-conditions", "remove-project-specific-conditions", "merge", "override", "merge-resolve"} {
		cmdtest.AssertFlag(t, cmd, "import", flag)
	}
	versions := cmdtest.FindCommand(t, cmd, "versions")
	cmdtest.AssertSubcommands(t, versions, "diff", "export", "list", "restore", "rollback", "show")
	for _, flag := range []string{"limit", "all", "before", "since", "until", "cached", "json"} {
		cmdtest.AssertNestedFlag(t, cmd, []string{"versions", "list"}, flag)
	}
	for _, flag := range []string{"cached", "json"} {
		cmdtest.AssertNestedFlag(t, cmd, []string{"versions", "show"}, flag)
	}
	show := cmdtest.FindCommand(t, cmd, "versions", "show")
	if show.Flags().Lookup("config") != nil {
		t.Fatal("versions show still exposes removed --config flag")
	}
	for _, flag := range []string{"filter", "search", "group", "expr", "parameters", "conditions", "cached", "json"} {
		cmdtest.AssertNestedFlag(t, cmd, []string{"versions", "diff"}, flag)
	}
	for _, name := range []string{"rollback", "restore"} {
		for _, flag := range []string{"dry-run", "yes", "json"} {
			cmdtest.AssertNestedFlag(t, cmd, []string{"versions", name}, flag)
		}
	}
}
