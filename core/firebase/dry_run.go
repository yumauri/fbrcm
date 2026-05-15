package firebase

import "context"

// dryRunContextKey holds dry run context key state used by the firebase package.
type dryRunContextKey struct{}

// WithDryRun handles with dry run and returns the resulting value or error.
func WithDryRun(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, dryRunContextKey{}, true)
}

// IsDryRun reports dry run and returns the resulting value or error.
func IsDryRun(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	dryRun, _ := ctx.Value(dryRunContextKey{}).(bool)
	return dryRun
}
