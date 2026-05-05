package firebase

import "context"

type dryRunContextKey struct{}

func WithDryRun(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, dryRunContextKey{}, true)
}

func IsDryRun(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	dryRun, _ := ctx.Value(dryRunContextKey{}).(bool)
	return dryRun
}
