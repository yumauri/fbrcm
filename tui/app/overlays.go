package app

import (
	"github.com/yumauri/fbrcm/tui/components/textview"
	"github.com/yumauri/fbrcm/tui/panels"
)

func (m Model) currentValueEditSource() panels.ID {
	if m.active == panels.Details {
		return panels.Details
	}
	return panels.Parameters
}

func (m *Model) closeOverlays() {
	m.requestContext.overlay = contextInline
	m.requestContext.pickerBack = contextInline
	m.requestContext.exchanges = nil
	m.requestContext.exchangeSource = ""
	m.requestContext.reading = nil
	m.requestContext.pasting = false
	m.cancelThemePicker()
	m.closeDialog()
	m.diffView = m.diffView.Close()
	m.explainView = textview.Model{}
	m.closeJSONInput()
	m.closeBoolPicker()
	m.closeNumberInput()
	m.closeStringInput()
	m.closeMoveParam()
	m.closeAuthPicker()
	m.closeRenameInput()
}
