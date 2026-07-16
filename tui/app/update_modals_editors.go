package app

import (
	tea "charm.land/bubbletea/v2"

	tuiconfig "github.com/yumauri/fbrcm/tui/config"
	"github.com/yumauri/fbrcm/tui/panels"
)

func (m Model) updateBoolPicker(msg tea.Msg) (Model, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		k := msg.String()
		if next, cmd, ok := modalCancel(m, tuiconfig.BlockBoolInput, k, func(model Model) Model {
			model.cancelConditionalValueAdd()
			model.closeBoolPicker()
			return model
		}); ok {
			return next, cmd, true
		}
		if tuiconfig.Matches(tuiconfig.BlockBoolInput, tuiconfig.ActionCopyValue, k) {
			if value, ok := m.boolPicker.CurrentString(); ok {
				return m, copyToClipboardCmd(value), true
			}
			return m, nil, true
		}
		if next, cmd, ok := modalSubmit(m, tuiconfig.BlockBoolInput, k, tuiconfig.ActionSubmit, true, (*Model).submitBoolPicker); ok {
			return next, cmd, true
		}
		switch {
		case tuiconfig.Matches(tuiconfig.BlockBoolInput, tuiconfig.ActionUp, k):
			m.boolPicker.Move(-1)
			return m, nil, true
		case tuiconfig.Matches(tuiconfig.BlockBoolInput, tuiconfig.ActionDown, k):
			m.boolPicker.Move(1)
			return m, nil, true
		}
	case tea.MouseClickMsg, tea.MouseMotionMsg, tea.MouseWheelMsg, tea.MouseReleaseMsg:
		return m, nil, true
	}
	return m, nil, false
}

func (m Model) updateJSONInput(msg tea.Msg) (Model, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		k := msg.String()
		if next, cmd, ok := modalCancel(m, tuiconfig.BlockJSONInput, k, func(model Model) Model {
			model.cancelConditionalValueAdd()
			model.closeJSONInput()
			return model
		}); ok {
			return next, cmd, true
		}
		if next, cmd, ok := modalCopy(m, tuiconfig.BlockJSONInput, k, m.jsonInput.PrettyValue()); ok {
			return next, cmd, true
		}
		if tuiconfig.Matches(tuiconfig.BlockJSONInput, tuiconfig.ActionFormat, k) {
			if m.jsonInput.Valid() {
				m.jsonInput = m.jsonInput.Reformat()
			}
			return m, nil, true
		}
		if next, cmd, ok := modalSubmit(m, tuiconfig.BlockJSONInput, k, tuiconfig.ActionSave, m.jsonInput.Valid(), (*Model).submitJSONInput); ok {
			return next, cmd, true
		}
		var cmd tea.Cmd
		m.jsonInput, cmd = m.jsonInput.Update(msg)
		return m, cmd, true
	case tea.PasteMsg, tea.ClipboardMsg:
		var cmd tea.Cmd
		m.jsonInput, cmd = m.jsonInput.Update(msg)
		return m, cmd, true
	case tea.MouseClickMsg, tea.MouseMotionMsg, tea.MouseWheelMsg, tea.MouseReleaseMsg:
		return m, nil, true
	}
	return m, nil, false
}

func (m Model) updateNumberInput(msg tea.Msg) (Model, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		k := msg.String()
		if next, cmd, ok := modalCancel(m, tuiconfig.BlockNumberInput, k, func(model Model) Model {
			model.cancelConditionalValueAdd()
			model.closeNumberInput()
			return model
		}); ok {
			return next, cmd, true
		}
		if next, cmd, ok := modalCopy(m, tuiconfig.BlockNumberInput, k, m.numberInput.Value()); ok {
			return next, cmd, true
		}
		if next, cmd, ok := modalSubmit(m, tuiconfig.BlockNumberInput, k, tuiconfig.ActionSubmit, m.numberInput.Valid(), (*Model).submitNumberInput); ok {
			return next, cmd, true
		}
		return m.updateNumberInputValue(msg)
	case tea.PasteMsg, tea.ClipboardMsg:
		return m.updateNumberInputValue(msg)
	case tea.MouseClickMsg, tea.MouseMotionMsg, tea.MouseWheelMsg, tea.MouseReleaseMsg:
		return m, nil, true
	}
	return m, nil, false
}

func (m Model) updateNumberInputValue(msg tea.Msg) (Model, tea.Cmd, bool) {
	var cmd tea.Cmd
	m.numberInput, cmd = m.numberInput.Update(msg)
	if m.valueEditSource == panels.Details {
		m.details = m.details.SetValuesInvalid(!m.numberInput.Valid())
	}
	return m, cmd, true
}

func (m Model) updateStringInput(msg tea.Msg) (Model, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		k := msg.String()
		if next, cmd, ok := modalCancel(m, tuiconfig.BlockStringInput, k, func(model Model) Model {
			model.cancelConditionalValueAdd()
			model.closeStringInput()
			return model
		}); ok {
			return next, cmd, true
		}
		if next, cmd, ok := modalCopy(m, tuiconfig.BlockStringInput, k, m.stringInput.Value()); ok {
			return next, cmd, true
		}
		if tuiconfig.Matches(tuiconfig.BlockStringInput, tuiconfig.ActionToggleExpanded, k) {
			return m, m.toggleStringInputMode(), true
		}
		if next, cmd, ok := modalSubmit(m, tuiconfig.BlockStringInput, k, tuiconfig.ActionSave, true, (*Model).submitStringInput); ok {
			return next, cmd, true
		}
		if tuiconfig.Matches(tuiconfig.BlockStringInput, tuiconfig.ActionSubmit, k) && !m.stringInput.IsExpanded() {
			return m, m.submitStringInput(), true
		}
		var cmd tea.Cmd
		m.stringInput, cmd = m.stringInput.Update(msg)
		return m, cmd, true
	case tea.PasteMsg, tea.ClipboardMsg:
		var cmd tea.Cmd
		m.stringInput, cmd = m.stringInput.Update(msg)
		return m, cmd, true
	case tea.MouseClickMsg, tea.MouseMotionMsg, tea.MouseWheelMsg, tea.MouseReleaseMsg:
		return m, nil, true
	}
	return m, nil, false
}
