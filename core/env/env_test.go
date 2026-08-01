package env

import (
	"os"
	"testing"
)

func TestNoColorEnabled(t *testing.T) {
	tests := []struct {
		name  string
		value *string
		want  bool
	}{
		{name: "unset", want: false},
		{name: "empty", value: new(""), want: false},
		{name: "one", value: new("1"), want: true},
		{name: "zero", value: new("0"), want: true},
		{name: "false", value: new("false"), want: true},
		{name: "arbitrary", value: new("please"), want: true},
		{name: "whitespace", value: new(" "), want: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.value == nil {
				previous, existed := os.LookupEnv(NoColor)
				if err := os.Unsetenv(NoColor); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					if existed {
						_ = os.Setenv(NoColor, previous)
					}
				})
			} else {
				t.Setenv(NoColor, *test.value)
			}

			if got := NoColorEnabled(); got != test.want {
				t.Fatalf("NoColorEnabled() = %t, want %t", got, test.want)
			}
		})
	}
}
