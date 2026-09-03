package projects

import (
	"testing"

	cmdtest "github.com/yumauri/fbrcm/cli/commands/testutil"
)

func TestNewCommandStructure(t *testing.T) {
	cmd := New(nil)
	cmdtest.AssertSubcommands(t, cmd, "list", "update", "forget", "diff", "promote", "aliases", "path", "reset")
	cmdtest.AssertSubcommands(t, cmdtest.FindCommand(t, cmd, "aliases"), "list", "set", "remove", "import")
	cmdtest.AssertNestedFlag(t, cmd, []string{"aliases", "list"}, "json")
	for _, subcommand := range []string{"set", "remove"} {
		cmdtest.AssertNestedFlag(t, cmd, []string{"aliases", subcommand}, "yes")
		cmdtest.AssertNestedFlag(t, cmd, []string{"aliases", subcommand}, "json")
	}
	for _, flag := range []string{"from", "conflict", "dry-run", "yes", "json"} {
		cmdtest.AssertNestedFlag(t, cmd, []string{"aliases", "import"}, flag)
	}
	for _, flag := range []string{"json", "filter", "expr", "url", "update"} {
		cmdtest.AssertFlag(t, cmd, "list", flag)
	}
	for _, flag := range []string{"json", "filter", "expr", "url", "auth"} {
		cmdtest.AssertFlag(t, cmd, "update", flag)
	}
	for _, flag := range []string{"filter", "expr", "yes"} {
		cmdtest.AssertFlag(t, cmd, "forget", flag)
	}
	for _, flag := range []string{"filter", "group", "expr", "search", "parameters", "conditions", "cached", "json"} {
		cmdtest.AssertFlag(t, cmd, "diff", flag)
	}
	for _, flag := range []string{"filter", "group", "expr", "search", "parameters", "conditions", "interactive", "all", "prune", "dry-run", "change-note", "yes", "json"} {
		cmdtest.AssertFlag(t, cmd, "promote", flag)
	}
	cmdtest.AssertFlag(t, cmd, "path", "json")
	cmdtest.AssertFlag(t, cmd, "reset", "yes")
}
