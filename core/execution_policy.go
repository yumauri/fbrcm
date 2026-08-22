package core

import (
	"context"
	"fmt"
)

// ExecutionPolicy controls access to application-managed local state and
// locally configured hooks for one execution context. It does not govern
// explicitly requested artifact output such as project export --to.
type ExecutionPolicy struct {
	ReadLocalState  bool
	WriteLocalState bool
	RunHooks        bool
}

type executionPolicyKey struct{}

// ExecutionPolicyError reports an operation blocked by the active execution
// policy before the prohibited local access or hook resolution occurs.
type ExecutionPolicyError struct {
	Capability string
	Operation  string
}

func (e *ExecutionPolicyError) Error() string {
	return fmt.Sprintf("execution policy disables %s for %s", e.Capability, e.Operation)
}

func requireLocalStateRead(ctx context.Context, operation string) error {
	if ExecutionPolicyFromContext(ctx).ReadLocalState {
		return nil
	}
	return &ExecutionPolicyError{Capability: "local-state reads", Operation: operation}
}

func requireLocalStateWrite(ctx context.Context, operation string) error {
	if ExecutionPolicyFromContext(ctx).WriteLocalState {
		return nil
	}
	return &ExecutionPolicyError{Capability: "local-state writes", Operation: operation}
}

// StatefulExecutionPolicy preserves the normal fbrcm execution behavior.
func StatefulExecutionPolicy() ExecutionPolicy {
	return ExecutionPolicy{ReadLocalState: true, WriteLocalState: true, RunHooks: true}
}

// StatelessExecutionPolicy disables application-managed local state and hooks.
func StatelessExecutionPolicy() ExecutionPolicy {
	return ExecutionPolicy{}
}

// WithExecutionPolicy binds an explicit policy to the returned context.
func WithExecutionPolicy(ctx context.Context, policy ExecutionPolicy) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, executionPolicyKey{}, policy)
}

// ExecutionPolicyFromContext returns the explicit policy or the fully
// stateful policy when no policy was attached, preserving existing callers.
func ExecutionPolicyFromContext(ctx context.Context) ExecutionPolicy {
	if ctx != nil {
		if policy, ok := ctx.Value(executionPolicyKey{}).(ExecutionPolicy); ok {
			return policy
		}
	}
	return StatefulExecutionPolicy()
}
