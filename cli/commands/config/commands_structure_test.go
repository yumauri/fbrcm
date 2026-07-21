package config

import (
	"testing"

	cmdtest "github.com/yumauri/fbrcm/cli/commands/testutil"
)

func TestNewCommandStructure(t *testing.T) {
	cmd := New()
	cmdtest.AssertSubcommands(t, cmd, "edit", "path", "reset", "set", "show", "validate")
	cmdtest.AssertNestedFlag(t, cmd, []string{"path"}, "json")
	cmdtest.AssertNestedFlag(t, cmd, []string{"show"}, "json")
	cmdtest.AssertNestedFlag(t, cmd, []string{"set"}, "json")
	for _, flag := range []string{"yes", "json"} {
		cmdtest.AssertNestedFlag(t, cmd, []string{"reset"}, flag)
	}
	cmdtest.AssertNestedFlag(t, cmd, []string{"validate"}, "json")
	cmdtest.AssertNestedFlag(t, cmd, []string{"edit"}, "editor")
}
