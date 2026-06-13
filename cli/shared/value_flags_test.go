package shared

import (
	"testing"

	"github.com/spf13/cobra"
)

func newValueFlagCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("boolean", "", "")
	cmd.Flags().String("number", "", "")
	cmd.Flags().String("string", "", "")
	cmd.Flags().String("json", "", "")
	return cmd
}

func TestReadValueFlagRequired(t *testing.T) {
	cmd := newValueFlagCommand()

	_, err := ReadValueFlag(cmd, true)
	if err == nil {
		t.Fatalf("ReadValueFlag required without flags returned nil error")
	}
}

func TestReadValueFlagOptional(t *testing.T) {
	cmd := newValueFlagCommand()

	got, err := ReadValueFlag(cmd, false)
	if err != nil {
		t.Fatalf("ReadValueFlag optional returned error: %v", err)
	}
	if got != nil {
		t.Fatalf("ReadValueFlag optional = %#v, want nil", got)
	}
}

func TestReadValueFlagNumber(t *testing.T) {
	cmd := newValueFlagCommand()
	if err := cmd.Flags().Set("number", "1e3"); err != nil {
		t.Fatal(err)
	}

	got, err := ReadValueFlag(cmd, true)
	if err != nil {
		t.Fatalf("ReadValueFlag returned error: %v", err)
	}
	if got.Value != "1e3" || got.Type != "NUMBER" {
		t.Fatalf("ReadValueFlag = %#v, want NUMBER 1e3", got)
	}
}

func TestReadValueFlagPreservesParseFloatNumberBehavior(t *testing.T) {
	cmd := newValueFlagCommand()
	if err := cmd.Flags().Set("number", "01"); err != nil {
		t.Fatal(err)
	}

	got, err := ReadValueFlag(cmd, true)
	if err != nil {
		t.Fatalf("ReadValueFlag returned error: %v", err)
	}
	if got.Value != "01" || got.Type != "NUMBER" {
		t.Fatalf("ReadValueFlag = %#v, want NUMBER 01", got)
	}
}
