package project

import (
	"testing"

	cmdtest "github.com/yumauri/fbrcm/cli/commands/testutil"
)

func TestNewCommandStructure(t *testing.T) {
	cmd := New(nil)
	cmdtest.AssertSubcommands(t, cmd, "export", "import")
	cmdtest.AssertFlag(t, cmd, "export", "to")
	for _, flag := range []string{"from", "group", "filter", "expr", "search", "dry-run", "draft", "remove-all-conditions", "remove-project-specific-conditions", "merge", "override", "merge-resolve"} {
		cmdtest.AssertFlag(t, cmd, "import", flag)
	}
}
