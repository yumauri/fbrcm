package apps

import (
	"testing"

	cmdtest "github.com/yumauri/fbrcm/cli/commands/testutil"
)

func TestAppsAdapterExposesWorkflowCommands(t *testing.T) {
	cmd := New(nil)
	cmdtest.AssertSubcommands(t, cmd, "config", "list", "show")
}
