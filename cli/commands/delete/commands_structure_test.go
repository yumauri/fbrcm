package deletecmd

import (
	"testing"

	cmdtest "github.com/yumauri/fbrcm/cli/commands/testutil"
)

func TestNewCommandStructure(t *testing.T) {
	cmdtest.AssertCommandStructure(t, New(nil), "delete [parameter]",
		"project", "filter", "expr", "search", "dry-run", "draft", "yes")
}
