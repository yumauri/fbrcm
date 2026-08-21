package projects

import (
	"errors"
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/cli/machine"
	"github.com/yumauri/fbrcm/core"
)

func TestListCommandRejectsUpdateInStatelessMode(t *testing.T) {
	cmd := newListCommand(nil)
	cmd.SetContext(core.WithExecutionPolicy(cmd.Context(), core.StatelessExecutionPolicy()))
	cmd.SetArgs([]string{"--update"})

	err := cmd.Execute()
	var argumentErr *machine.ArgumentError
	if !errors.As(err, &argumentErr) || !strings.Contains(err.Error(), "--update cannot be used with --stateless") {
		t.Fatalf("Execute() error = %#v, want typed stateless update error", err)
	}
}
