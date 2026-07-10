package app

import (
	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/tui/components/parameters"
	"github.com/yumauri/fbrcm/tui/panels"
)

func (m *Model) currentBoolValueAnchor() (parameters.BoolValueAnchor, bool) {
	if m.valueEditSource == panels.Details || (m.valueEditSource == panels.None && m.active == panels.Details) {
		return m.details.CurrentBoolValueAnchor()
	}
	return m.parameters.CurrentBoolValueAnchor()
}

func (m *Model) currentNumberValueAnchor() (parameters.NumberValueAnchor, bool) {
	if m.valueEditSource == panels.Details || (m.valueEditSource == panels.None && m.active == panels.Details) {
		return m.details.CurrentNumberValueAnchor()
	}
	return m.parameters.CurrentNumberValueAnchor()
}

func (m *Model) currentStringValueAnchor() (parameters.StringValueAnchor, bool) {
	if m.valueEditSource == panels.Details || (m.valueEditSource == panels.None && m.active == panels.Details) {
		return m.details.CurrentStringValueAnchor(m.width)
	}
	return m.parameters.CurrentStringValueAnchor()
}

func (m *Model) currentJSONValueAnchor() (parameters.JSONValueAnchor, bool) {
	if m.valueEditSource == panels.Details || (m.valueEditSource == panels.None && m.active == panels.Details) {
		return m.details.CurrentJSONValueAnchor()
	}
	return m.parameters.CurrentJSONValueAnchor()
}

func (m *Model) openDetailsValueEditor() tea.Cmd {
	m.valueEditSource = panels.Details
	if _, ok := m.details.CurrentBoolValueAnchor(); ok {
		return m.openBoolPicker()
	}
	if _, ok := m.details.CurrentNumberValueAnchor(); ok {
		return m.openNumberInput()
	}
	if _, ok := m.details.CurrentStringValueAnchor(m.width); ok {
		return m.openStringInput()
	}
	if _, ok := m.details.CurrentJSONValueAnchor(); ok {
		return m.openJSONInput()
	}
	return nil
}
