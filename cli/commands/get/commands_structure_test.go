package get

import (
	"testing"

	cmdtest "github.com/yumauri/fbrcm/cli/commands/testutil"
)

func TestNewCommandStructure(t *testing.T) {
	cmdtest.AssertCommandStructure(t, New(nil), "get [parameter]",
		"json", "project", "filter", "expr", "search", "all", "update")
}
