package app

import (
	tea "charm.land/bubbletea/v2"

	"fbrcm/tui/components/parameters"
	"fbrcm/tui/panels"
)

// currentBoolValueAnchor handles current bool value anchor for Model and returns the resulting state or error.
func (m *Model) currentBoolValueAnchor() (parameters.BoolValueAnchor, bool) {
	if m.valueEditSource == panels.Details || (m.valueEditSource == panels.None && m.active == panels.Details) {
		return m.details.CurrentBoolValueAnchor()
	}
	return m.parameters.CurrentBoolValueAnchor()
}

// currentNumberValueAnchor handles current number value anchor for Model and returns the resulting state or error.
func (m *Model) currentNumberValueAnchor() (parameters.NumberValueAnchor, bool) {
	if m.valueEditSource == panels.Details || (m.valueEditSource == panels.None && m.active == panels.Details) {
		return m.details.CurrentNumberValueAnchor()
	}
	return m.parameters.CurrentNumberValueAnchor()
}

// currentStringValueAnchor handles current string value anchor for Model and returns the resulting state or error.
func (m *Model) currentStringValueAnchor() (parameters.StringValueAnchor, bool) {
	if m.valueEditSource == panels.Details || (m.valueEditSource == panels.None && m.active == panels.Details) {
		return m.details.CurrentStringValueAnchor(m.width)
	}
	return m.parameters.CurrentStringValueAnchor()
}

// currentJSONValueAnchor handles current jsonvalue anchor for Model and returns the resulting state or error.
func (m *Model) currentJSONValueAnchor() (parameters.JSONValueAnchor, bool) {
	if m.valueEditSource == panels.Details || (m.valueEditSource == panels.None && m.active == panels.Details) {
		return m.details.CurrentJSONValueAnchor()
	}
	return m.parameters.CurrentJSONValueAnchor()
}

// openDetailsValueEditor opens open details value editor for Model and returns the resulting state or error.
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
