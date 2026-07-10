package testutil

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// NormalizeViewSnapshot strips ANSI codes and trailing whitespace from a TUI
// view string so snapshot tests compare stable plain-text output.
func NormalizeViewSnapshot(view string) string {
	plain := ansi.Strip(view)
	lines := strings.Split(plain, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " ")
	}
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return strings.Join(lines, "\n")
}
