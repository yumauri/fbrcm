package cache

import (
	"testing"

	cmdtest "github.com/yumauri/fbrcm/cli/commands/testutil"
)

func TestNewCommandStructure(t *testing.T) {
	cmd := New()
	cmdtest.AssertSubcommands(t, cmd, "path", "clear", "list")
	cmdtest.AssertFlag(t, cmd, "path", "json")
	cmdtest.AssertFlag(t, cmd, "clear", "yes")
	cmdtest.AssertFlag(t, cmd, "list", "json")
}
