package app

import tea "charm.land/bubbletea/v2"

func (m Model) updateOpenModal(msg tea.Msg) (Model, tea.Cmd, bool) {
	if m.parameters.HistoryPickerOpen() {
		switch msg.(type) {
		case tea.KeyMsg:
			var cmd tea.Cmd
			m.parameters, cmd = m.parameters.Update(msg)
			return m, cmd, true
		case tea.MouseMsg:
			return m, nil, true
		}
	}
	if m.conditions.MoveActive() {
		return m.updateConditionMove(msg)
	}
	if m.authPicker.IsOpen() {
		return m.updateAuthPicker(msg)
	}
	if m.dialog.IsOpen() {
		return m.updateDialog(msg)
	}
	if m.boolPicker.IsOpen() {
		return m.updateBoolPicker(msg)
	}
	if m.jsonInput.IsOpen() {
		return m.updateJSONInput(msg)
	}
	if m.numberInput.IsOpen() {
		return m.updateNumberInput(msg)
	}
	if m.stringInput.IsOpen() {
		return m.updateStringInput(msg)
	}
	if m.moveParam.IsOpen() {
		return m.updateMoveParam(msg)
	}
	if m.renameInput.IsOpen() {
		return m.updateRenameInput(msg)
	}
	return m, nil, false
}
