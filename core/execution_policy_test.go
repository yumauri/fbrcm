package core

import (
	"context"
	"errors"
	"testing"
)

func TestExecutionPolicyDefaultsToStateful(t *testing.T) {
	want := StatefulExecutionPolicy()
	for _, ctx := range []context.Context{nil, context.Background()} {
		if got := ExecutionPolicyFromContext(ctx); got != want {
			t.Fatalf("ExecutionPolicyFromContext() = %#v, want %#v", got, want)
		}
	}
}

func TestStatelessExecutionPolicyDisablesManagedPersistenceAndHooks(t *testing.T) {
	parent := context.Background()
	ctx := WithExecutionPolicy(parent, StatelessExecutionPolicy())
	if got := ExecutionPolicyFromContext(ctx); got != (ExecutionPolicy{}) {
		t.Fatalf("stateless policy = %#v", got)
	}
	if got := ExecutionPolicyFromContext(parent); got != StatefulExecutionPolicy() {
		t.Fatalf("parent policy changed = %#v", got)
	}
}

func TestExecutionPolicyPreservesIndependentControls(t *testing.T) {
	want := ExecutionPolicy{ReadLocalState: true, RunHooks: true}
	ctx := WithExecutionPolicy(context.Background(), want)
	if got := ExecutionPolicyFromContext(ctx); got != want {
		t.Fatalf("policy = %#v, want %#v", got, want)
	}
}

func TestDraftEntryPointsRejectDisabledLocalStateBeforeAccess(t *testing.T) {
	svc := setupCoreTestEnv(t)

	ctx := WithExecutionPolicy(context.Background(), StatelessExecutionPolicy())
	_, err := svc.PrepareDraftPublish(ctx, "demo")
	var policyErr *ExecutionPolicyError
	if !errors.As(err, &policyErr) || policyErr.Capability != "local-state reads" {
		t.Fatalf("PrepareDraftPublish error = %v, want local-read ExecutionPolicyError", err)
	}

	ctx = WithExecutionPolicy(context.Background(), ExecutionPolicy{ReadLocalState: true})
	_, _, _, _, err = svc.DuplicateParameter(ctx, "demo", "group", "parameter")
	if !errors.As(err, &policyErr) || policyErr.Capability != "local-state writes" {
		t.Fatalf("DuplicateParameter error = %v, want local-write ExecutionPolicyError", err)
	}
}
