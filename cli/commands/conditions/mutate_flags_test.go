package conditions

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestConditionDescriptionMutationFlagsAreNotExposed(t *testing.T) {
	cmd := New(nil)
	add, _, err := cmd.Find([]string{"add"})
	if err != nil {
		t.Fatal(err)
	}
	if add.Flags().Lookup("description") != nil {
		t.Fatal("conditions add still exposes --description")
	}
	edit, _, err := cmd.Find([]string{"edit"})
	if err != nil {
		t.Fatal(err)
	}
	for _, flag := range []string{"description", "no-description"} {
		if edit.Flags().Lookup(flag) != nil {
			t.Errorf("conditions edit still exposes --%s", flag)
		}
	}
}

func TestReadConditionEditTreatsFalseNoColorAsAbsent(t *testing.T) {
	for _, test := range []struct {
		name           string
		argv           []string
		wantExpression *string
		wantColor      *string
		wantErr        bool
	}{
		{name: "no flags", wantErr: true},
		{name: "explicit false only", argv: []string{"--no-color=false"}, wantErr: true},
		{name: "expression with explicit false", argv: []string{"--expression", "true", "--no-color=false"}, wantExpression: new("true")},
		{name: "color with explicit false", argv: []string{"--color", "BLUE", "--no-color=false"}, wantColor: new("BLUE")},
		{name: "remove color", argv: []string{"--no-color"}, wantColor: new("")},
		{name: "conflicting color removal", argv: []string{"--color", "BLUE", "--no-color"}, wantErr: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().String("expression", "", "")
			cmd.Flags().String("color", "", "")
			cmd.Flags().Bool("no-color", false, "")
			if err := cmd.ParseFlags(test.argv); err != nil {
				t.Fatal(err)
			}
			expression, color, err := readConditionEdit(cmd)
			if (err != nil) != test.wantErr || !equalOptionalString(expression, test.wantExpression) || !equalOptionalString(color, test.wantColor) {
				t.Fatalf("readConditionEdit = %v, %v, %v", optionalStringValue(expression), optionalStringValue(color), err)
			}
		})
	}
}

func equalOptionalString(left, right *string) bool {
	return left == nil && right == nil || left != nil && right != nil && *left == *right
}

func optionalStringValue(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}
