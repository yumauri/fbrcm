package conditions

import (
	"testing"

	cmdtest "github.com/yumauri/fbrcm/cli/commands/testutil"
)

func TestNewCommandStructure(t *testing.T) {
	cmd := New(nil)
	cmdtest.AssertSubcommands(t, cmd, "list", "show")
	for _, flag := range []string{"update", "json", "filter", "search"} {
		cmdtest.AssertNestedFlag(t, cmd, []string{"list"}, flag)
	}
	for _, flag := range []string{"update", "json"} {
		cmdtest.AssertNestedFlag(t, cmd, []string{"show"}, flag)
	}
}
