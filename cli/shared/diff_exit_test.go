package shared

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"
)

func TestDiffFoundErrorAlwaysReturnsStatusOne(t *testing.T) {
	root := &cobra.Command{Use: "root"}
	diff := &cobra.Command{Use: "diff"}
	root.AddCommand(diff)
	assertExitCode(t, DiffFoundError(diff), 1)
	if !root.SilenceErrors || !root.SilenceUsage {
		t.Fatal("semantic diff result did not suppress Cobra error and usage output")
	}
	SetMachineMode(true)
	t.Cleanup(func() { SetMachineMode(false) })
	assertExitCode(t, DiffFoundError(diff), 1)
}

func assertExitCode(t *testing.T, err error, want int) {
	t.Helper()
	var exitErr *ExitError
	if !errors.As(err, &exitErr) || exitErr.Code != want {
		t.Fatalf("exit error = %#v, want code %d", err, want)
	}
}
