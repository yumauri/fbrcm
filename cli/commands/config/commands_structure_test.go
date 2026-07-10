package config

import (
	"testing"

	cmdtest "github.com/yumauri/fbrcm/cli/commands/testutil"
)

func TestNewCommandStructure(t *testing.T) {
	cmdtest.AssertNestedFlag(t, New(), []string{"path"}, "json")
}
