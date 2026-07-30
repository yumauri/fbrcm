package updatecmd

import (
	"testing"

	cmdtest "github.com/yumauri/fbrcm/cli/commands/testutil"
)

func TestNewCommandStructure(t *testing.T) {
	cmdtest.AssertCommandStructure(t, New(nil), "update [parameter]",
		"project", "filter", "expr", "search", "dry-run", "draft", "yes", "description", "group", "no-group", "name",
		"type", "value", "use-in-app-default", "condition",
		"remove-all-conditional-values", "remove-conditional-value", "json")
}
