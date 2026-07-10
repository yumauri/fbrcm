package updatecmd

import (
	"testing"

	cmdtest "github.com/yumauri/fbrcm/cli/commands/testutil"
)

func TestNewCommandStructure(t *testing.T) {
	cmdtest.AssertCommandStructure(t, New(nil), "update [parameter]",
		"project", "filter", "expr", "search", "dry-run", "yes", "description", "group", "no-group", "name",
		"boolean", "number", "string", "json", "remove-all-conditional-values", "remove-conditional-value")
}
