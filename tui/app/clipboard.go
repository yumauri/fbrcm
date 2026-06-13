package app

import (
	tea "charm.land/bubbletea/v2"
	"github.com/atotto/clipboard"
)

func copyToClipboardCmd(text string) tea.Cmd {
	if text == "" {
		return nil
	}
	return func() tea.Msg {
		_ = clipboard.WriteAll(text)
		return nil
	}
}
