package app

import (
	tea "charm.land/bubbletea/v2"
	"github.com/atotto/clipboard"
)

// copyToClipboardCmd handles copy to clipboard cmd and returns the resulting value or error.
func copyToClipboardCmd(text string) tea.Cmd {
	if text == "" {
		return nil
	}
	return func() tea.Msg {
		_ = clipboard.WriteAll(text)
		return nil
	}
}
