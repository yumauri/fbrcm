package app

import (
	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/panels"
)

// closeDetailsPanel closes close details panel for Model and returns the resulting state or error.
func (m *Model) closeDetailsPanel() {
	m.detailsVisible = false
	m.details = m.details.SetData(nil)
	m.newParameter = nil
	m.parameters.ClearTransientNewParameter()
	if m.active == panels.Details {
		m.setActive(panels.Parameters)
	}
}

// openNewParameterDetails opens open new parameter details for Model and returns the resulting state or error.
func (m *Model) openNewParameterDetails() tea.Cmd {
	project, groupKey, afterParamKey, ok := m.parameters.CurrentNewParameterTarget()
	if !ok {
		return nil
	}
	m.closeDialog(false)
	m.closeJSONInput()
	m.closeBoolPicker()
	m.closeNumberInput()
	m.closeStringInput()
	m.closeMoveParam()
	m.closeRenameInput()
	m.parameters.OpenTransientNewParameter(project.ProjectID, groupKey, afterParamKey)
	data, ok := m.parameters.CurrentParameterViewData()
	if !ok {
		m.parameters.ClearTransientNewParameter()
		return nil
	}
	m.newParameter = &newParameterSession{projectID: project.ProjectID, groupKey: groupKey}
	if m.detailsVisible && m.details.Dirty() {
		return m.handleParameterSelection(messages.ParameterSelectionChangedMsg{Data: data, Activate: true})
	}
	m.details = m.details.SetData(data)
	m.detailsVisible = true
	m.setActive(panels.Details)
	var cmd tea.Cmd
	m.details, cmd = m.details.ActivateName()
	return cmd
}

// activateDetailsGroup activates details group editor.
func (m *Model) activateDetailsGroup() tea.Cmd {
	if !m.detailsVisible {
		return nil
	}
	var cmd tea.Cmd
	m.details, cmd = m.details.ActivateGroup()
	return cmd
}

// requestCloseDetails handles request close details for Model and returns the resulting state or error.
func (m *Model) requestCloseDetails() tea.Cmd {
	if m.details.FieldActive() {
		m.details = m.details.DeactivateField()
		return nil
	}
	edit, ok := m.details.Edit()
	if !ok {
		m.closeDetailsPanel()
		return nil
	}
	data := m.details.Data()
	if data == nil {
		m.closeDetailsPanel()
		return nil
	}
	if m.details.Invalid() {
		m.openInvalidDetailsDialog(data.Project, m.details.InvalidReasons(), true)
		return nil
	}
	if m.parameters.HasDraft(data.Project.ProjectID) {
		return m.editParameterDetailsCmd(data.Project, edit, false, true, false)
	}
	m.openEditDetailsDialog(data.Project, edit, true, false)
	return nil
}

// submitDetailsForm handles submit details form for Model and returns the resulting state or error.
func (m *Model) submitDetailsForm() tea.Cmd {
	edit, ok := m.details.Edit()
	if !ok {
		return nil
	}
	data := m.details.Data()
	if data == nil {
		return nil
	}
	if m.details.Invalid() {
		m.openInvalidDetailsDialog(data.Project, m.details.InvalidReasons(), false)
		return nil
	}
	if m.parameters.HasDraft(data.Project.ProjectID) {
		return m.editParameterDetailsCmd(data.Project, edit, false, false, true)
	}
	m.openEditDetailsDialog(data.Project, edit, false, true)
	return nil
}

// requestDeleteDetails opens delete flow for details parameter.
func (m *Model) requestDeleteDetails() tea.Cmd {
	if anchor, ok := m.details.CurrentConditionalValueAnchor(); ok {
		if m.parameters.HasDraft(anchor.Project.ProjectID) {
			return m.deleteConditionalValueCmd(anchor.Project, anchor.GroupKey, anchor.ParamKey, anchor.ValueLabel, false)
		}
		m.openDeleteConditionalValueDialog(anchor.Project, anchor.GroupKey, anchor.ParamKey, anchor.ValueLabel)
		x, y, width, height := m.details.Bounds()
		m.dialog = m.dialog.CenterWithin(x, y, width, height)
		return nil
	}
	data := m.details.Data()
	if data == nil {
		return nil
	}
	if m.parameters.HasDraft(data.Project.ProjectID) {
		return m.deleteParameterCmd(data.Project, data.GroupKey, data.Parameter.Key, false, true)
	}
	m.openDeleteDialog(data.Project, data.GroupKey, data.Parameter.Key, true)
	x, y, width, height := m.details.Bounds()
	m.dialog = m.dialog.CenterWithin(x, y, width, height)
	return nil
}

// copyDetailsNameCmd copies current details parameter name.
func (m Model) copyDetailsNameCmd() tea.Cmd {
	data := m.details.Data()
	if data == nil {
		return nil
	}
	return copyToClipboardCmd(data.Parameter.Key)
}

// copyDetailsPathCmd copies current details parameter path.
func (m Model) copyDetailsPathCmd() tea.Cmd {
	data := m.details.Data()
	if data == nil {
		return nil
	}
	return copyToClipboardCmd(data.Project.ProjectID + "/" + data.GroupKey + "/" + data.Parameter.Key)
}

// copyDetailsSelectedValueCmd copies selected details value.
func (m Model) copyDetailsSelectedValueCmd() tea.Cmd {
	value, ok := m.details.SelectedRawValue()
	if !ok {
		return nil
	}
	return copyToClipboardCmd(value)
}

// applyParameterSelection handles apply parameter selection for Model and returns the resulting state or error.
func (m *Model) applyParameterSelection(msg messages.ParameterSelectionChangedMsg) {
	if msg.ResetScroll {
		m.details = m.details.ResetScroll()
	}
	if msg.Data != nil {
		m.details = m.details.SetData(msg.Data)
	}
	if msg.Activate && msg.Data != nil {
		m.detailsVisible = true
		m.setActive(panels.Details)
	}
}

// handleParameterSelection handles handle parameter selection for Model and returns the resulting state or error.
func (m *Model) handleParameterSelection(msg messages.ParameterSelectionChangedMsg) tea.Cmd {
	if msg.Data == nil && msg.ResetScroll {
		m.applyParameterSelection(msg)
		return nil
	}
	if !m.detailsVisible || !m.details.Dirty() {
		m.applyParameterSelection(msg)
		return nil
	}

	edit, ok := m.details.Edit()
	if !ok {
		m.applyParameterSelection(msg)
		return nil
	}
	data := m.details.Data()
	if data == nil {
		m.applyParameterSelection(msg)
		return nil
	}

	m.pendingDetails = &pendingDetailsSelection{data: msg.Data, activate: msg.Activate}
	if m.details.Invalid() {
		m.openInvalidDetailsDialog(data.Project, m.details.InvalidReasons(), false)
		return nil
	}
	if m.parameters.HasDraft(data.Project.ProjectID) {
		return m.editParameterDetailsCmd(data.Project, edit, false, false, false)
	}
	m.openEditDetailsDialog(data.Project, edit, false, false)
	return nil
}

// applyPendingDetailsSelection handles apply pending details selection for Model and returns the resulting state or error.
func (m *Model) applyPendingDetailsSelection() {
	if m.pendingDetails == nil {
		return
	}
	pending := m.pendingDetails
	m.pendingDetails = nil
	m.applyParameterSelection(messages.ParameterSelectionChangedMsg{
		Data:     pending.data,
		Activate: pending.activate,
	})
}
