package app

import (
	"time"

	tea "charm.land/bubbletea/v2"

	tuiconfig "github.com/yumauri/fbrcm/tui/config"
)

func (m Model) updateMoveParam(msg tea.Msg) (Model, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		k := msg.String()
		if next, cmd, ok := modalCancel(m, tuiconfig.BlockMoveInput, k, func(model Model) Model {
			model.closeMoveParam()
			return model
		}); ok {
			return next, cmd, true
		}
		if tuiconfig.Matches(tuiconfig.BlockMoveInput, tuiconfig.ActionSubmit, k) {
			if _, ok := m.moveParam.Current(); ok {
				return m, m.submitMoveParam(), true
			}
			return m, nil, true
		}
		switch {
		case tuiconfig.Matches(tuiconfig.BlockMoveInput, tuiconfig.ActionUp, k):
			return m, m.moveParam.Move(-1), true
		case tuiconfig.Matches(tuiconfig.BlockMoveInput, tuiconfig.ActionDown, k):
			return m, m.moveParam.Move(1), true
		}
		if m.moveParam.InputSelected() {
			return m, m.moveParam.Update(msg), true
		}
		if m.moveParam.Typeahead(msg.String(), time.Now()) {
			return m, nil, true
		}
	case tea.PasteMsg, tea.ClipboardMsg:
		if m.moveParam.InputSelected() {
			return m, m.moveParam.Update(msg), true
		}
	case tea.MouseClickMsg, tea.MouseMotionMsg, tea.MouseWheelMsg, tea.MouseReleaseMsg:
		return m, nil, true
	}
	return m, nil, false
}

func (m Model) updateRenameInput(msg tea.Msg) (Model, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		k := msg.String()
		if next, cmd, ok := modalCancel(m, tuiconfig.BlockRenameInput, k, func(model Model) Model {
			model.cancelRenameInput()
			return model
		}); ok {
			return next, cmd, true
		}
		if next, cmd, ok := modalSubmit(m, tuiconfig.BlockRenameInput, k, tuiconfig.ActionSubmit, true, m.submitRenameInput); ok {
			return next, cmd, true
		}
		var cmd tea.Cmd
		m.renameInput, cmd = m.renameInput.Update(msg)
		return m, cmd, true
	case tea.PasteMsg, tea.ClipboardMsg:
		var cmd tea.Cmd
		m.renameInput, cmd = m.renameInput.Update(msg)
		return m, cmd, true
	case tea.MouseClickMsg, tea.MouseMotionMsg, tea.MouseWheelMsg, tea.MouseReleaseMsg:
		return m, nil, true
	}
	return m, nil, false
}
