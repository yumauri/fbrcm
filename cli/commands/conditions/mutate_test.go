package conditions

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/core"
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
