package addcmd

import (
	"testing"

	cmdtest "github.com/yumauri/fbrcm/cli/commands/testutil"
)

func TestNewCommandStructure(t *testing.T) {
	cmdtest.AssertCommandStructure(t, New(nil), "add <parameter>",
		"project", "expr", "dry-run", "draft", "description", "group", "boolean", "number", "string", "json")
}
