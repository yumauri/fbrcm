package env

import (
	"os"
	"testing"
)

func TestNoColorEnabled(t *testing.T) {
	testNonEmptySwitch(t, NoColor, NoColorEnabled)
}

func TestLogPlainEnabled(t *testing.T) {
	testNonEmptySwitch(t, LogPlain, LogPlainEnabled)
}

func testNonEmptySwitch(t *testing.T, name string, enabled func() bool) {
	t.Helper()
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
				previous, existed := os.LookupEnv(name)
				if err := os.Unsetenv(name); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					if existed {
						_ = os.Setenv(name, previous)
					}
				})
			} else {
				t.Setenv(name, *test.value)
			}

			if got := enabled(); got != test.want {
				t.Fatalf("%s enabled = %t, want %t", name, got, test.want)
			}
		})
	}
}

func TestLookupNonEmptyPreservesOpaqueValue(t *testing.T) {
	t.Setenv(GoogleAccessToken, "  opaque token  ")
	value, ok := LookupNonEmpty(GoogleAccessToken)
	if !ok || value != "  opaque token  " {
		t.Fatalf("LookupNonEmpty = %q, %t", value, ok)
	}
	t.Setenv(GoogleAccessToken, "")
	if _, ok := LookupNonEmpty(GoogleAccessToken); ok {
		t.Fatal("LookupNonEmpty accepted an empty value")
	}
}
