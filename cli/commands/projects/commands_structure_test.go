package projects

import (
	"testing"

	cmdtest "github.com/yumauri/fbrcm/cli/commands/testutil"
)

func TestNewCommandStructure(t *testing.T) {
	cmd := New(nil)
	cmdtest.AssertSubcommands(t, cmd, "list", "update", "path", "purge")
	for _, flag := range []string{"json", "filter", "expr", "url", "update"} {
		cmdtest.AssertFlag(t, cmd, "list", flag)
	}
	for _, flag := range []string{"json", "filter", "expr", "url", "auth"} {
		cmdtest.AssertFlag(t, cmd, "update", flag)
	}
	cmdtest.AssertFlag(t, cmd, "path", "json")
	cmdtest.AssertFlag(t, cmd, "purge", "yes")
}
