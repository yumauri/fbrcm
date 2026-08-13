package shared

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestValidateQueryFlagsRejectsEmptyModeQueries(t *testing.T) {
	for _, test := range []struct {
		name  string
		value string
	}{
		{name: "filter", value: "/"},
		{name: "project", value: "server@="},
		{name: "project", value: "client@"},
		{name: "project", value: "   "},
	} {
		t.Run(test.name+test.value, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().StringArray(test.name, nil, "")
			if err := cmd.Flags().Set(test.name, test.value); err != nil {
				t.Fatal(err)
			}
			if err := ValidateQueryFlags(cmd); err == nil {
				t.Fatalf("ValidateQueryFlags accepted %s=%q", test.name, test.value)
			}
		})
	}
}

func TestValidateNonBlankInputsRejectsWhitespaceOnlyValues(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("expr", "", "")
	if err := cmd.Flags().Set("expr", " \t "); err != nil {
		t.Fatal(err)
	}
	if err := ValidateNonBlankInputs(cmd, nil); err == nil {
		t.Fatal("ValidateNonBlankInputs accepted a whitespace-only flag")
	}

	cmd = &cobra.Command{}
	if err := ValidateNonBlankInputs(cmd, []string{" \n "}); err == nil {
		t.Fatal("ValidateNonBlankInputs accepted a whitespace-only argument")
	}
}

func TestValidateNonBlankInputsPreservesIntentionalEmptyContent(t *testing.T) {
	for _, name := range []string{"value", "description", "change-note", "group", "label"} {
		t.Run(name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().String(name, "default", "")
			if err := cmd.Flags().Set(name, ""); err != nil {
				t.Fatal(err)
			}
			if err := ValidateNonBlankInputs(cmd, nil); err != nil {
				t.Fatalf("exact empty content rejected: %v", err)
			}
			if err := cmd.Flags().Set(name, "   "); err != nil {
				t.Fatal(err)
			}
			if err := ValidateNonBlankInputs(cmd, nil); err == nil || !strings.Contains(err.Error(), "non-empty") {
				t.Fatalf("whitespace-only content error = %v", err)
			}
		})
	}
}

func TestValidateQueryFlagsAcceptsCanonicalQueries(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().StringArray("filter", nil, "")
	cmd.Flags().StringArray("project", nil, "")
	for name, value := range map[string]string{"filter": "/checkout", "project": "SERVER@=demo"} {
		if err := cmd.Flags().Set(name, value); err != nil {
			t.Fatal(err)
		}
	}
	if err := ValidateQueryFlags(cmd); err != nil {
		t.Fatalf("ValidateQueryFlags = %v", err)
	}
}

func TestRejectTemplateProjectFilters(t *testing.T) {
	if err := RejectTemplateProjectFilters([]string{"=demo", "/prod"}); err != nil {
		t.Fatalf("ordinary filters rejected: %v", err)
	}
	if err := RejectTemplateProjectFilters([]string{"server@=demo"}); err == nil {
		t.Fatal("template filter accepted by project-scoped command")
	}
}

func TestRejectChangedFlagsRejectsExplicitFalse(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("dry-run", false, "")
	if err := cmd.Flags().Set("dry-run", "false"); err != nil {
		t.Fatal(err)
	}
	if err := RejectChangedFlags(cmd, "stdin mode", "dry-run"); err == nil {
		t.Fatal("explicitly supplied false flag was accepted")
	}
}
