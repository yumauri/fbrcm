package numberinput

import "testing"

func TestModelValidUsesJSONNumberRules(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{name: "integer", in: "42", want: true},
		{name: "decimal", in: "3.14", want: true},
		{name: "leading zero rejected", in: "01", want: false},
		{name: "blank rejected", in: " ", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, _ := New().Open(0, 0, 3, 10, tt.in)
			if got := m.Valid(); got != tt.want {
				t.Fatalf("Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}
