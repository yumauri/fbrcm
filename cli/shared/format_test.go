package shared

import "testing"

func TestFormatParameterHeader(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		group string
		want  string
	}{
		{name: "root", key: "flag", want: "flag"},
		{name: "grouped", key: "flag", group: "mobile", want: "flag [mobile]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatParameterHeader(tt.key, tt.group); got != tt.want {
				t.Fatalf("FormatParameterHeader() = %q, want %q", got, tt.want)
			}
		})
	}
}
