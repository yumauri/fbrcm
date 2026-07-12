package app

import (
	tea "charm.land/bubbletea/v2"

	tuiconfig "github.com/yumauri/fbrcm/tui/config"
)

func (m Model) updateDialog(msg tea.Msg) (Model, tea.Cmd, bool) {
	if m.historyRollbackModalLocked() {
		switch msg.(type) {
		case tea.KeyMsg, tea.MouseClickMsg, tea.MouseMotionMsg, tea.MouseWheelMsg, tea.MouseReleaseMsg:
			return m, nil, true
		}
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if tuiconfig.Matches(tuiconfig.BlockDialog, tuiconfig.ActionCancel, msg.String()) {
			if m.historyRollback != nil {
				m.cancelHistoryRollback()
				return m, nil, true
			}
			m.closeDialog()
			return m, nil, true
		}
		var cmd tea.Cmd
		m.dialog, cmd = m.dialog.Update(msg)
		return m, cmd, true
	case tea.MouseClickMsg, tea.MouseMotionMsg, tea.MouseWheelMsg, tea.MouseReleaseMsg:
		var cmd tea.Cmd
		m.dialog, cmd = m.dialog.Update(msg)
		return m, cmd, true
	}
	return m, nil, false
}
