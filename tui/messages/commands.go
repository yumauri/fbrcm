package messages

import tea "charm.land/bubbletea/v2"

func KeyboardCaptureCmd(enabled bool) tea.Cmd {
	return func() tea.Msg {
		return KeyboardCaptureMsg{Enabled: enabled}
	}
}
