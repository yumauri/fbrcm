package duplicatecmd

import (
	"testing"

	cmdtest "github.com/yumauri/fbrcm/cli/commands/testutil"
)

func TestCommandStructure(t *testing.T) {
	cmdtest.AssertCommandStructure(t, New(nil), "duplicate <source> <target>",
		"project", "expr", "dry-run", "draft", "yes")
}
