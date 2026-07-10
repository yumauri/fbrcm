package draft

import (
	"testing"

	"github.com/yumauri/fbrcm/core/rootgroup"
)

func TestNormalizeGroupKey(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "tree key to wire", input: rootgroup.TreeKey, want: rootgroup.WireKey},
		{name: "named group unchanged", input: "checkout", want: "checkout"},
		{name: "empty string unchanged", input: "", want: ""},
		{name: "root label unchanged", input: rootgroup.Label, want: rootgroup.Label},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeGroupKey(tt.input); got != tt.want {
				t.Fatalf("NormalizeGroupKey(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
