package projects

import (
	"testing"

	cmdtest "github.com/yumauri/fbrcm/cli/commands/testutil"
)

func TestNewCommandStructure(t *testing.T) {
	cmd := New(nil)
	cmdtest.AssertSubcommands(t, cmd, "list", "update", "forget", "diff", "promote", "path", "reset")
	for _, flag := range []string{"json", "filter", "expr", "url", "update"} {
		cmdtest.AssertFlag(t, cmd, "list", flag)
	}
	for _, flag := range []string{"json", "filter", "expr", "url", "auth"} {
		cmdtest.AssertFlag(t, cmd, "update", flag)
	}
	for _, flag := range []string{"filter", "expr", "yes"} {
		cmdtest.AssertFlag(t, cmd, "forget", flag)
	}
	for _, flag := range []string{"filter", "group", "expr", "search", "parameters", "conditions", "cached", "json", "exit-code"} {
		cmdtest.AssertFlag(t, cmd, "diff", flag)
	}
	for _, flag := range []string{"filter", "group", "expr", "search", "parameters", "conditions", "interactive", "all", "prune", "dry-run", "yes", "json"} {
		cmdtest.AssertFlag(t, cmd, "promote", flag)
	}
	cmdtest.AssertFlag(t, cmd, "path", "json")
	cmdtest.AssertFlag(t, cmd, "reset", "yes")
}
