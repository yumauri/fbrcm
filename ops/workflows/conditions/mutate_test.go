package conditions

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/ops/shared"
)

func TestConditionMutationRejectsDraftInStatelessMode(t *testing.T) {
	cmd := newAddCommand(nil)
	cmd.SetContext(core.WithExecutionPolicy(context.Background(), core.StatelessExecutionPolicy()))
	cmd.SetArgs([]string{"demo", "example", "--expression", "percent <= 1", "--draft", "--yes"})

	err := cmd.Execute()
	var argument *shared.ArgumentError
	if !errors.As(err, &argument) || !strings.Contains(err.Error(), "--draft cannot be used with --stateless") {
		t.Fatalf("Execute error = %T %v, want typed stateless draft error", err, err)
	}
}

func TestConditionAddArgumentsPreserveScalarAndBulkForms(t *testing.T) {
	project, name := conditionAddArguments([]string{"demo", "Beta users"})
	if project == nil || *project != "demo" || name != "Beta users" {
		t.Fatalf("scalar add arguments = project %#v, name %q", project, name)
	}

	project, name = conditionAddArguments([]string{"Beta users"})
	if project != nil || name != "Beta users" {
		t.Fatalf("bulk add arguments = project %#v, name %q", project, name)
	}
}

func TestConditionDeleteArgumentsPreserveScalarAndBulkForms(t *testing.T) {
	project, condition := conditionDeleteArguments([]string{"demo", "Beta users"})
	if project == nil || *project != "demo" || condition == nil || *condition != "Beta users" {
		t.Fatalf("scalar delete arguments = project %#v, condition %#v", project, condition)
	}

	project, condition = conditionDeleteArguments([]string{"demo"})
	if project == nil || *project != "demo" || condition != nil {
		t.Fatalf("project-filtered delete arguments = project %#v, condition %#v", project, condition)
	}

	project, condition = conditionDeleteArguments(nil)
	if project != nil || condition != nil {
		t.Fatalf("bulk delete arguments = project %#v, condition %#v", project, condition)
	}
}

func TestConditionMutationsRejectPositionalAndFlagProjectSelection(t *testing.T) {
	tests := [][]string{
		{"add", "demo", "Beta", "--project", "=other", "--expression", "true", "--yes"},
		{"delete", "demo", "Beta", "--project", "=other", "--yes"},
	}
	for _, args := range tests {
		cmd := New(nil)
		cmd.SetArgs(args)
		err := cmd.Execute()
		var argument *shared.ArgumentError
		if !errors.As(err, &argument) || !strings.Contains(err.Error(), "<project> and --project cannot be used together") {
			t.Fatalf("%v error = %T %v, want project selection ArgumentError", args, err, err)
		}
	}
}

func TestConditionDeleteRejectsPositionalConditionAndFilter(t *testing.T) {
	cmd := New(nil)
	cmd.SetArgs([]string{"delete", "demo", "Beta", "--filter", "=Beta", "--yes"})
	err := cmd.Execute()
	var argument *shared.ArgumentError
	if !errors.As(err, &argument) || !strings.Contains(err.Error(), "condition argument cannot be used together with --filter") {
		t.Fatalf("Execute error = %T %v, want condition selection ArgumentError", err, err)
	}
}
