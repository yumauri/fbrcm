package rc

import (
	"errors"
	"fmt"
	"testing"

	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/ops/machine"
)

func TestIsRemoteConfigConflict(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil", err: nil, want: false},
		{name: "http 412", err: &firebase.APIError{StatusCode: 412}, want: true},
		{name: "http 409", err: &firebase.APIError{StatusCode: 409}, want: true},
		{name: "typed", err: &machine.ConflictError{Err: errors.New("changed")}, want: true},
		{name: "wrapped", err: fmt.Errorf("publish: %w", &firebase.APIError{StatusCode: 412}), want: true},
		{name: "untyped wording", err: errors.New("stale etag"), want: false},
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
