package shared

import (
	"os"
	"strconv"
	"strings"

	"golang.org/x/term"
)

// TerminalWidth returns the configured or detected stdout width. It falls
// back to 80 columns when stdout is not an interactive terminal.
func TerminalWidth() int {
	if columns := strings.TrimSpace(os.Getenv("COLUMNS")); columns != "" {
		if width, err := strconv.Atoi(columns); err == nil && width > 0 {
			return width
		}
	}

	info, err := os.Stdout.Stat()
	if err == nil && (info.Mode()&os.ModeCharDevice) != 0 {
		width, _, err := term.GetSize(int(os.Stdout.Fd()))
		if err == nil && width > 0 {
			return width
		}
	}

	return 80
}
