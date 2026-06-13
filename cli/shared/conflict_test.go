package shared

import (
	"errors"
	"fmt"
	"testing"
)

func TestIsRemoteConfigConflict(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil", err: nil, want: false},
		{name: "http 412", err: errors.New("remote config api returned 412"), want: true},
		{name: "precondition", err: errors.New("Precondition Failed"), want: true},
		{name: "condition not met", err: errors.New("conditionNotMet"), want: true},
		{name: "etag", err: errors.New("stale etag"), want: true},
		{name: "wrapped", err: fmt.Errorf("publish: %w", errors.New("If-Match failed")), want: true},
		{name: "other", err: errors.New("permission denied"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRemoteConfigConflict(tt.err); got != tt.want {
				t.Fatalf("IsRemoteConfigConflict(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
