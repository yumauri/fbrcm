package conditions

import (
	"testing"

	cmdtest "github.com/yumauri/fbrcm/cli/commands/testutil"
)

func TestNewCommandStructure(t *testing.T) {
	cmd := New(nil)
	cmdtest.AssertSubcommands(t, cmd, "add", "delete", "edit", "list", "move", "rename", "show", "validate")
	for _, flag := range []string{"update", "json", "filter", "search", "expr"} {
		cmdtest.AssertNestedFlag(t, cmd, []string{"list"}, flag)
	}
	for _, flag := range []string{"update", "json"} {
		cmdtest.AssertNestedFlag(t, cmd, []string{"show"}, flag)
	}
	for _, command := range []string{"add", "delete", "edit", "move", "rename"} {
		for _, flag := range []string{"draft", "dry-run", "yes"} {
			cmdtest.AssertNestedFlag(t, cmd, []string{command}, flag)
		}
	}
	for _, flag := range []string{"expression", "color", "priority"} {
		cmdtest.AssertNestedFlag(t, cmd, []string{"add"}, flag)
	}
	for _, flag := range []string{"expression", "color", "no-color"} {
		cmdtest.AssertNestedFlag(t, cmd, []string{"edit"}, flag)
	}
	cmdtest.AssertNestedFlag(t, cmd, []string{"validate"}, "json")
}
