package hooks

import "context"

type operationContextKey struct{}

// WithOperation identifies the user operation that caused publication.
func WithOperation(ctx context.Context, operation string) context.Context {
	return context.WithValue(ctx, operationContextKey{}, operation)
}

// OperationFromContext returns the publication operation, defaulting to publish.
func OperationFromContext(ctx context.Context) string {
	if ctx != nil {
		if operation, ok := ctx.Value(operationContextKey{}).(string); ok && operation != "" {
			return operation
		}
	}
	return "publish"
}
