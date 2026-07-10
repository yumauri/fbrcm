package profile

import (
	"testing"

	cmdtest "github.com/yumauri/fbrcm/cli/commands/testutil"
)

func TestNewCommandStructure(t *testing.T) {
	cmd := New()
	cmdtest.AssertSubcommands(t, cmd, "list", "switch", "rename", "path", "purge")
	cmdtest.AssertFlag(t, cmd, "list", "json")
	cmdtest.AssertFlag(t, cmd, "path", "json")
	cmdtest.AssertFlag(t, cmd, "purge", "yes")
}
