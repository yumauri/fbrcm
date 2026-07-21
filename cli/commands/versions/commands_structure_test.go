package versions

import (
	"testing"

	cmdtest "github.com/yumauri/fbrcm/cli/commands/testutil"
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
	for _, flag := range []string{"filter", "search", "group", "expr", "parameters", "conditions", "cached", "json", "exit-code"} {
		cmdtest.AssertNestedFlag(t, cmd, []string{"diff"}, flag)
	}
	for _, flag := range []string{"to", "cached", "yes"} {
		cmdtest.AssertNestedFlag(t, cmd, []string{"export"}, flag)
	}
	for _, name := range []string{"rollback", "restore"} {
		for _, flag := range []string{"dry-run", "yes", "json"} {
			cmdtest.AssertNestedFlag(t, cmd, []string{name}, flag)
		}
	}
}
