package shared

import (
	"testing"

	"github.com/spf13/cobra"
)

func newValueFlagCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("type", "", "")
	cmd.Flags().String("value", "", "")
	cmd.Flags().Bool("use-in-app-default", false, "")
	return cmd
}

func TestReadValueFlagUseInAppDefault(t *testing.T) {
	cmd := newValueFlagCommand()
	if err := cmd.Flags().Set("use-in-app-default", "true"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("type", "json"); err != nil {
		t.Fatal(err)
	}

	got, err := ReadValueFlag(cmd, true)
	if err != nil {
		t.Fatalf("ReadValueFlag returned error: %v", err)
	}
	if !got.UseInAppDefault || got.Value != "" || got.Type != "JSON" {
		t.Fatalf("ReadValueFlag = %#v, want JSON use-in-app-default", got)
	}
}

func TestReadValueFlagRejectsInAppDefaultWithConcreteValue(t *testing.T) {
	cmd := newValueFlagCommand()
	if err := cmd.Flags().Set("type", "string"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("use-in-app-default", "true"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("value", "remote"); err != nil {
		t.Fatal(err)
	}

	if _, err := ReadValueFlag(cmd, true); err == nil {
		t.Fatal("ReadValueFlag accepted in-app default with a concrete value")
	}
}

func TestParseValueType(t *testing.T) {
	got, err := ParseValueType(" bool ")
	if err != nil || got != "BOOLEAN" {
		t.Fatalf("ParseValueType = %q, %v, want BOOLEAN, nil", got, err)
	}
	if _, err := ParseValueType("object"); err == nil {
		t.Fatal("ParseValueType accepted unsupported type")
	}
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
	if err := cmd.Flags().Set("type", "number"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("value", "1e3"); err != nil {
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
	if err := cmd.Flags().Set("type", "number"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("value", "01"); err != nil {
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

func TestReadValueFlagJSONValue(t *testing.T) {
	cmd := newValueFlagCommand()
	if err := cmd.Flags().Set("type", "json"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("value", `{"enabled":true}`); err != nil {
		t.Fatal(err)
	}

	got, err := ReadValueFlag(cmd, true)
	if err != nil {
		t.Fatalf("ReadValueFlag returned error: %v", err)
	}
	if got.Value != `{"enabled":true}` || got.Type != "JSON" {
		t.Fatalf("ReadValueFlag = %#v, want JSON value", got)
	}
}

func TestReadValueFlagPreservesExplicitEmptyString(t *testing.T) {
	cmd := newValueFlagCommand()
	if err := cmd.Flags().Set("type", "string"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("value", ""); err != nil {
		t.Fatal(err)
	}

	got, err := ReadValueFlag(cmd, true)
	if err != nil {
		t.Fatalf("ReadValueFlag returned error: %v", err)
	}
	if got.Value != "" || got.Type != "STRING" {
		t.Fatalf("ReadValueFlag = %#v, want explicit empty string", got)
	}
}

func TestReadValueFlagRequiresType(t *testing.T) {
	cmd := newValueFlagCommand()
	if err := cmd.Flags().Set("value", "enabled"); err != nil {
		t.Fatal(err)
	}

	if _, err := ReadValueFlag(cmd, true); err == nil || err.Error() != "--type is required with --value" {
		t.Fatalf("ReadValueFlag error = %v, want required type", err)
	}
}

func TestReadValueFlagRejectsTypeWithoutSelection(t *testing.T) {
	cmd := newValueFlagCommand()
	if err := cmd.Flags().Set("type", "string"); err != nil {
		t.Fatal(err)
	}

	if _, err := ReadValueFlag(cmd, false); err == nil || err.Error() != "--type requires --value or --use-in-app-default" {
		t.Fatalf("ReadValueFlag error = %v, want type selection error", err)
	}
}

func TestReadValueFlagValidatesTypedValues(t *testing.T) {
	for _, tc := range []struct {
		name      string
		valueType string
		value     string
	}{
		{name: "boolean", valueType: "boolean", value: "yes"},
		{name: "number", valueType: "number", value: "many"},
		{name: "json", valueType: "json", value: "{"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd := newValueFlagCommand()
			if err := cmd.Flags().Set("type", tc.valueType); err != nil {
				t.Fatal(err)
			}
			if err := cmd.Flags().Set("value", tc.value); err != nil {
				t.Fatal(err)
			}
			if _, err := ReadValueFlag(cmd, true); err == nil {
				t.Fatalf("ReadValueFlag accepted invalid %s value %q", tc.valueType, tc.value)
			}
		})
	}
}
