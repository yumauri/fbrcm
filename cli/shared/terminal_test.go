package shared

import "testing"

func TestTerminalWidthUsesColumnsEnvironment(t *testing.T) {
	t.Setenv("COLUMNS", "123")
	if got := TerminalWidth(); got != 123 {
		t.Fatalf("TerminalWidth() = %d, want 123", got)
	}
}

func TestTerminalWidthIgnoresInvalidColumnsEnvironment(t *testing.T) {
	t.Setenv("COLUMNS", "invalid")
	if got := TerminalWidth(); got <= 0 {
		t.Fatalf("TerminalWidth() = %d, want positive fallback", got)
	}
}
