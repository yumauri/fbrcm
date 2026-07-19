package shared

import (
	"errors"
	"fmt"
	"testing"

	"github.com/spf13/cobra"
)

func TestDiffExitCodeContract(t *testing.T) {
	cmd := &cobra.Command{Use: "diff"}
	AddDiffExitCodeFlag(cmd)
	if DiffExitCodeEnabled(cmd) {
		t.Fatal("diff exit code enabled by default")
	}
	if err := DiffFoundError(cmd); err != nil {
		t.Fatalf("default difference error = %v", err)
	}
	original := fmt.Errorf("comparison failed")
	if err := DiffCommandError(cmd, original); !errors.Is(err, original) {
		t.Fatalf("default command error = %v, want original", err)
	}

	if err := cmd.Flags().Set("exit-code", "true"); err != nil {
		t.Fatal(err)
	}
	if !DiffExitCodeEnabled(cmd) {
		t.Fatal("diff exit code is not enabled")
	}
	if !DiffExitCodeRequested(cmd, nil) {
		t.Fatal("enabled diff exit code was not requested")
	}
	assertExitCode(t, DiffFoundError(cmd), 1)
	assertExitCode(t, DiffCommandError(cmd, original), 2)

	explicit := WithExitCode(nil, 1)
	if got := DiffCommandError(cmd, explicit); got != explicit {
		t.Fatalf("explicit exit error replaced: got %#v, want %#v", got, explicit)
	}
}

func TestDiffExitCodeRequestedFromRawArguments(t *testing.T) {
	cmd := &cobra.Command{Use: "diff"}
	AddDiffExitCodeFlag(cmd)
	for _, args := range [][]string{
		{"diff", "--bad-flag", "--exit-code"},
		{"diff", "--exit-code=true"},
		{"diff", "--exit-code=invalid"},
	} {
		if !DiffExitCodeRequested(cmd, args) {
			t.Fatalf("DiffExitCodeRequested(%q) = false, want true", args)
		}
	}
	for _, args := range [][]string{
		{"diff", "--bad-flag"},
		{"diff", "--exit-code=false"},
		{"diff", "--", "--exit-code"},
	} {
		if DiffExitCodeRequested(cmd, args) {
			t.Fatalf("DiffExitCodeRequested(%q) = true, want false", args)
		}
	}
}

func assertExitCode(t *testing.T, err error, want int) {
	t.Helper()
	var exitErr *ExitError
	if !errors.As(err, &exitErr) || exitErr.Code != want {
		t.Fatalf("exit error = %#v, want code %d", err, want)
	}
}
