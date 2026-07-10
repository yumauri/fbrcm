package app

import "github.com/yumauri/fbrcm/tui/panels"

func (m Model) currentValueEditSource() panels.ID {
	if m.active == panels.Details {
		return panels.Details
	}
	return panels.Parameters
}

func (m *Model) closeOverlays() {
	m.closeDialog()
	m.closeJSONInput()
	m.closeBoolPicker()
	m.closeNumberInput()
	m.closeStringInput()
	m.closeMoveParam()
	m.closeRenameInput()
}
