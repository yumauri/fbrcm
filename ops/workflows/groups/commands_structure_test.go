package groups

import (
	"testing"

	cmdtest "github.com/yumauri/fbrcm/cli/commands/testutil"
)

func TestMutationCommandStructure(t *testing.T) {
	cmd := New(nil)
	for _, command := range []string{"add", "edit", "rename", "delete"} {
		for _, flag := range []string{"project", "draft", "dry-run", "change-note", "yes", "json"} {
			cmdtest.AssertNestedFlag(t, cmd, []string{command}, flag)
		}
	}
}
