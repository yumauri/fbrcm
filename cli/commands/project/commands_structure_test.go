package project

import (
	"testing"

	cmdtest "github.com/yumauri/fbrcm/cli/commands/testutil"
)

func TestNewCommandStructure(t *testing.T) {
	cmd := New(nil)
	cmdtest.AssertSubcommands(t, cmd, "export", "import")
	cmdtest.AssertFlag(t, cmd, "export", "to")
	for _, flag := range []string{"from", "group", "filter", "expr", "search", "dry-run", "draft", "remove-all-conditions", "keep-portable-conditions-only", "merge", "override", "merge-resolve"} {
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
