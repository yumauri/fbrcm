package auth

import (
	"testing"

	cmdtest "github.com/yumauri/fbrcm/cli/commands/testutil"
)

func TestNewCommandStructure(t *testing.T) {
	cmd := New(nil)
	cmdtest.AssertSubcommands(t, cmd, "list", "add", "login", "path", "delete", "bind")
	cmdtest.AssertSubcommands(t, cmdtest.FindCommand(t, cmd, "add"), "oauth", "service-account", "gcloud")
	cmdtest.AssertFlag(t, cmd, "list", "json")
	cmdtest.AssertFlag(t, cmd, "login", "noopen")
	cmdtest.AssertFlag(t, cmd, "path", "json")
	cmdtest.AssertFlag(t, cmd, "delete", "yes")
	cmdtest.AssertFlag(t, cmd, "bind", "auth")
	cmdtest.AssertFlag(t, cmd, "bind", "project")
	cmdtest.AssertNestedFlag(t, cmd, []string{"add", "oauth"}, "from")
	cmdtest.AssertNestedFlag(t, cmd, []string{"add", "oauth"}, "label")
	cmdtest.AssertNestedFlag(t, cmd, []string{"add", "service-account"}, "from")
	cmdtest.AssertNestedFlag(t, cmd, []string{"add", "service-account"}, "label")
	cmdtest.AssertNestedFlag(t, cmd, []string{"add", "gcloud"}, "label")
}
