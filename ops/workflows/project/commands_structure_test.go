package project

import (
	"testing"

	cmdtest "github.com/yumauri/fbrcm/cli/commands/testutil"
)

func TestNewCommandStructure(t *testing.T) {
	cmd := New(nil)
	cmdtest.AssertSubcommands(t, cmd, "show", "templates", "open", "export", "import", "defaults")
	templates := cmdtest.FindCommand(t, cmd, "templates")
	cmdtest.AssertSubcommands(t, templates, "show", "set")
	cmdtest.AssertFlag(t, cmd, "show", "update")
	cmdtest.AssertFlag(t, cmd, "show", "json")
	cmdtest.AssertNestedFlag(t, cmd, []string{"templates", "show"}, "json")
	for _, flag := range []string{"templates", "primary", "json"} {
		cmdtest.AssertNestedFlag(t, cmd, []string{"templates", "set"}, flag)
	}
	cmdtest.AssertFlag(t, cmd, "export", "to")
	cmdtest.AssertFlag(t, cmd, "export", "yes")
	for _, flag := range []string{"format", "to", "yes"} {
		cmdtest.AssertFlag(t, cmd, "defaults", flag)
	}
	for _, flag := range []string{"from", "group", "filter", "expr", "search", "dry-run", "draft", "change-note", "remove-all-conditions", "keep-portable-conditions-only", "merge", "override", "merge-resolve", "yes", "json"} {
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
