package project

import (
	"testing"

	cmdtest "github.com/yumauri/fbrcm/cli/commands/testutil"
)

func TestNewCommandStructure(t *testing.T) {
	cmd := New(nil)
	cmdtest.AssertSubcommands(t, cmd, "show", "open", "export", "import", "defaults")
	cmdtest.AssertFlag(t, cmd, "show", "update")
	cmdtest.AssertFlag(t, cmd, "show", "json")
	cmdtest.AssertFlag(t, cmd, "export", "to")
	cmdtest.AssertFlag(t, cmd, "export", "yes")
	for _, flag := range []string{"format", "to", "yes"} {
		cmdtest.AssertFlag(t, cmd, "defaults", flag)
	}
	for _, flag := range []string{"from", "group", "filter", "expr", "search", "dry-run", "draft", "remove-all-conditions", "keep-portable-conditions-only", "merge", "override", "merge-resolve", "yes", "json"} {
		cmdtest.AssertFlag(t, cmd, "import", flag)
	}
	importCmd, _, err := cmd.Find([]string{"import"})
	if err != nil {
		t.Fatal(err)
	}
	if importCmd.Flags().Lookup("remove-project-specific-conditions") != nil {
		t.Fatal("import still exposes --remove-project-specific-conditions")
	}
}
