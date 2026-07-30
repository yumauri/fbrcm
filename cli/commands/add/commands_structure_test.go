package addcmd

import (
	"testing"

	cmdtest "github.com/yumauri/fbrcm/cli/commands/testutil"
)

func TestNewCommandStructure(t *testing.T) {
	cmdtest.AssertCommandStructure(t, New(nil), "add <parameter>",
		"project", "expr", "dry-run", "change-note", "draft", "yes", "description", "group", "type", "value",
		"use-in-app-default", "json")
}
