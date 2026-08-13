package shared

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"
)

func TestRequireYesInMachineMode(t *testing.T) {
	SetMachineMode(true)
	t.Cleanup(func() { SetMachineMode(false) })
	err := RequireYesInMachineMode(&cobra.Command{}, false, "deleting the profile", true)
	var interaction *InteractionError
	if !errors.As(err, &interaction) || interaction.RequiredOption != "--yes" || !interaction.Destructive {
		t.Fatalf("error = %#v", err)
	}
	if err := RequireYesInMachineMode(&cobra.Command{}, true, "deleting the profile", true); err != nil {
		t.Fatalf("explicit confirmation failed: %v", err)
	}
}

func TestNonInteractiveInputReturnsStructuredError(t *testing.T) {
	_, err := NonInteractiveInput().Read(make([]byte, 1))
	var interaction *InteractionError
	if !errors.As(err, &interaction) {
		t.Fatalf("error = %T %v", err, err)
	}
}
