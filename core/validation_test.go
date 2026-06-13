package core

import "testing"

func TestIsJSONNumber(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{name: "integer", in: "42", want: true},
		{name: "negative decimal", in: "-3.14", want: true},
		{name: "exponent", in: "1e-9", want: true},
		{name: "trimmed", in: " 0.5 ", want: true},
		{name: "empty", in: "", want: false},
		{name: "leading zero", in: "01", want: false},
		{name: "nan", in: "NaN", want: false},
		{name: "trailing text", in: "1px", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsJSONNumber(tt.in); got != tt.want {
				t.Fatalf("IsJSONNumber(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}
