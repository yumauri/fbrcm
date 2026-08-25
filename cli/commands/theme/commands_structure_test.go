package theme

import (
	"testing"

	cmdtest "github.com/yumauri/fbrcm/cli/commands/testutil"
)

func TestNewCommandStructure(t *testing.T) {
	cmd := New()
	cmdtest.AssertSubcommands(t, cmd, "list", "switch", "reset", "delete", "path", "rename", "import")
	cmdtest.AssertFlag(t, cmd, "switch", "scope")
	cmdtest.AssertFlag(t, cmd, "reset", "scope")
	cmdtest.AssertFlag(t, cmd, "delete", "yes")
	if got := cmdtest.FindCommand(t, cmd, "import").Use; got != "import [source]" {
		t.Fatalf("import Use = %q", got)
	}
}
