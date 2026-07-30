package firebase

import (
	"context"
	"fmt"
	"strings"
	"unicode"
)

type changeNoteContextKey struct{}

type changeNoteContextValue struct {
	value string
}

// NormalizeChangeNote validates and normalizes a user-authored Remote Config
// change note.
func NormalizeChangeNote(value string) (string, error) {
	value = strings.TrimSpace(value)
	for _, r := range value {
		if r == '\n' || r == '\r' || unicode.IsControl(r) {
			return "", fmt.Errorf("change note must be a single line without control characters")
		}
	}
	return value, nil
}

// WithChangeNote records an explicitly supplied change note in ctx. An empty
// value remains explicit so callers can clear a stored draft change note.
func WithChangeNote(ctx context.Context, value string) (context.Context, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	value, err := NormalizeChangeNote(value)
	if err != nil {
		return nil, err
	}
	return context.WithValue(ctx, changeNoteContextKey{}, changeNoteContextValue{value: value}), nil
}

// ChangeNoteFromContext returns an explicitly supplied change note.
func ChangeNoteFromContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	value, ok := ctx.Value(changeNoteContextKey{}).(changeNoteContextValue)
	return value.value, ok
}
