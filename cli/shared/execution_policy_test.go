package shared

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/core"
)

func TestRejectStatelessDraft(t *testing.T) {
	stateless := core.WithExecutionPolicy(context.Background(), core.StatelessExecutionPolicy())
	err := RejectStatelessDraft(stateless, true)
	var argument *ArgumentError
	if !errors.As(err, &argument) || !strings.Contains(err.Error(), "--draft cannot be used with --stateless") {
		t.Fatalf("RejectStatelessDraft error = %T %v, want typed stateless draft error", err, err)
	}
	if err := RejectStatelessDraft(stateless, false); err != nil {
		t.Fatalf("stateless non-draft error = %v", err)
	}
	stateful := core.WithExecutionPolicy(context.Background(), core.StatefulExecutionPolicy())
	if err := RejectStatelessDraft(stateful, true); err != nil {
		t.Fatalf("stateful draft error = %v", err)
	}
}
