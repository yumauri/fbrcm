package app

import (
	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/panels"
)

func (m *Model) closeDetailsPanel() {
	m.detailsVisible = false
	m.details = m.details.SetData(nil)
	m.newParameter = nil
	m.parameters.ClearTransientNewParameter()
	if m.active == panels.Details {
		m.setActive(m.selectedParametersTab())
	}
}

func (m *Model) applyConditionSelection(msg messages.ConditionSelectionChangedMsg) {
	if msg.ResetScroll {
		m.details = m.details.ResetScroll()
	}
	if msg.Data == nil {
		return
	}
	if msg.Data != nil && (!m.detailsVisible || !m.details.Dirty()) {
		m.details = m.details.SetConditionData(msg.Data)
	}
	if msg.Activate && msg.Data != nil && !m.details.Dirty() {
		m.detailsVisible = true
		m.setActive(panels.Details)
	}
}

func (m *Model) openNewParameterDetails() tea.Cmd {
	project, groupKey, afterParamKey, ok := m.parameters.CurrentNewParameterTarget()
	if !ok {
		return nil
	}
	m.closeOverlays()
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

func (m *Model) requestCloseDetails() tea.Cmd {
	if m.details.IsCondition() {
		return m.requestCloseConditionDetails()
	}
	if m.details.IsGroup() {
		return m.requestCloseGroupDetails()
	}
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

func (m *Model) requestCloseGroupDetails() tea.Cmd {
	if m.details.FieldActive() {
		m.details = m.details.DeactivateField()
		return nil
	}
	edit, ok := m.details.GroupEdit()
	if !ok {
		m.closeDetailsPanel()
		return nil
	}
	data := m.details.GroupData()
	if data == nil {
		return nil
	}
	if m.details.Invalid() {
		m.openInvalidDetailsDialog(data.Project, m.details.InvalidReasons(), true)
		return nil
	}
	if m.parameters.HasDraft(data.Project.ProjectID) {
		return m.editGroupDetailsCmd(data.Project, edit, false, true)
	}
	m.openEditGroupDetailsDialog(data.Project, edit, true)
	return nil
}

func (m *Model) requestCloseConditionDetails() tea.Cmd {
	if m.details.FieldActive() {
		m.details = m.details.DeactivateField()
		return nil
	}
	edit, ok := m.details.ConditionEdit()
	if !ok {
		m.closeDetailsPanel()
		return nil
	}
	data := m.details.ConditionData()
	if data == nil {
		m.closeDetailsPanel()
		return nil
	}
	if m.details.Invalid() {
		m.openInvalidDetailsDialog(data.Project, m.details.InvalidReasons(), true)
		return nil
	}
	if m.parameters.HasDraft(data.Project.ProjectID) || m.conditions.HasDraft(data.Project.ProjectID) {
		return m.conditionDetailsMutationCmd(data.Project, edit, false, true)
	}
	m.openConditionDetailsDialog(data.Project, edit, true)
	return nil
}

func (m *Model) submitDetailsForm() tea.Cmd {
	if m.details.IsCondition() {
		return m.submitConditionDetailsForm()
	}
	if m.details.IsGroup() {
		return m.submitGroupDetailsForm()
	}
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

func (m *Model) submitGroupDetailsForm() tea.Cmd {
	edit, ok := m.details.GroupEdit()
	if !ok {
		return nil
	}
	data := m.details.GroupData()
	if data == nil {
		return nil
	}
	if m.details.Invalid() {
		m.openInvalidDetailsDialog(data.Project, m.details.InvalidReasons(), false)
		return nil
	}
	if m.parameters.HasDraft(data.Project.ProjectID) {
		return m.editGroupDetailsCmd(data.Project, edit, false, false)
	}
	m.openEditGroupDetailsDialog(data.Project, edit, false)
	return nil
}

func (m *Model) submitConditionDetailsForm() tea.Cmd {
	edit, ok := m.details.ConditionEdit()
	if !ok {
		return nil
	}
	data := m.details.ConditionData()
	if data == nil {
		return nil
	}
	if m.details.Invalid() {
		m.openInvalidDetailsDialog(data.Project, m.details.InvalidReasons(), false)
		return nil
	}
	if m.parameters.HasDraft(data.Project.ProjectID) || m.conditions.HasDraft(data.Project.ProjectID) {
		return m.conditionDetailsMutationCmd(data.Project, edit, false, false)
	}
	m.openConditionDetailsDialog(data.Project, edit, false)
	return nil
}

// requestDeleteDetails opens delete flow for details parameter.
func (m *Model) requestDeleteDetails() tea.Cmd {
	if data := m.details.GroupData(); data != nil {
		if m.parameters.HasDraft(data.Project.ProjectID) {
			return m.deleteGroupCmd(data.Project, data.Group.Key, false, true)
		}
		m.openDeleteGroupDialog(data.Project, data.Group.Key, data.Group.Label, true)
		x, y, width, height := m.details.Bounds()
		m.dialog = m.dialog.CenterWithin(x, y, width, height)
		return nil
	}
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
	if data := m.details.ConditionData(); data != nil {
		return copyToClipboardCmd(data.Condition.Name)
	}
	if data := m.details.GroupData(); data != nil {
		return copyToClipboardCmd(data.Group.Key)
	}
	data := m.details.Data()
	if data == nil {
		return nil
	}
	return copyToClipboardCmd(data.Parameter.Key)
}

// copyDetailsPathCmd copies current details parameter path.
func (m Model) copyDetailsPathCmd() tea.Cmd {
	if data := m.details.ConditionData(); data != nil {
		return copyToClipboardCmd(data.Project.ProjectID + "/conditions/" + data.Condition.Name)
	}
	if data := m.details.GroupData(); data != nil {
		return copyToClipboardCmd(data.Project.ProjectID + "/" + data.Group.Key)
	}
	data := m.details.Data()
	if data == nil {
		return nil
	}
	return copyToClipboardCmd(data.Project.ProjectID + "/" + data.GroupKey + "/" + data.Parameter.Key)
}

// copyDetailsSelectedValueCmd copies selected details value.
func (m Model) copyDetailsSelectedValueCmd() tea.Cmd {
	value, ok := m.details.SelectedRawValue()
	if ok {
		return copyToClipboardCmd(value)
	}
	if data := m.details.ConditionData(); data != nil {
		return copyToClipboardCmd(data.Condition.Expression)
	}
	return nil
}

func (m *Model) openSelectedValueConditionDetails() tea.Cmd {
	anchor, ok := m.details.CurrentConditionalValueAnchor()
	if !ok {
		return nil
	}
	data, ok := m.conditions.Condition(anchor.Project.ProjectID, anchor.ValueLabel)
	if !ok {
		return nil
	}
	return m.handleConditionDetailsSelection(data)
}

func (m *Model) openSelectedUsageParameterDetails() tea.Cmd {
	usage, ok := m.details.SelectedUsage()
	condition := m.details.ConditionData()
	if !ok || condition == nil {
		return nil
	}
	data, ok := m.parameters.ParameterViewData(
		condition.Project.ProjectID, usage.GroupKey, usage.ParameterKey, condition.Condition.Name,
	)
	if !ok {
		return nil
	}
	return m.handleParameterSelection(messages.ParameterSelectionChangedMsg{Data: data, Activate: true})
}

func (m *Model) applyParameterSelection(msg messages.ParameterSelectionChangedMsg) {
	if msg.ResetScroll {
		m.details = m.details.ResetScroll()
	}
	if msg.GroupData != nil {
		m.details = m.details.SetGroupData(msg.GroupData)
	} else if msg.Data != nil {
		m.details = m.details.SetData(msg.Data)
	}
	if msg.Activate && (msg.Data != nil || msg.GroupData != nil) {
		m.detailsVisible = true
		m.setActive(panels.Details)
	}
}

func (m *Model) handleParameterSelection(msg messages.ParameterSelectionChangedMsg) tea.Cmd {
	if msg.Data == nil && msg.GroupData == nil && msg.ResetScroll {
		m.applyParameterSelection(msg)
		return nil
	}
	if !m.detailsVisible || !m.details.Dirty() {
		m.applyParameterSelection(msg)
		return nil
	}
	if m.details.IsCondition() {
		m.pendingDetails = &pendingDetailsSelection{data: msg.Data, groupData: msg.GroupData, activate: msg.Activate}
		return m.saveConditionDetailsForPending()
	}
	if m.details.IsGroup() {
		m.pendingDetails = &pendingDetailsSelection{data: msg.Data, groupData: msg.GroupData, activate: msg.Activate}
		return m.saveGroupDetailsForPending()
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

	m.pendingDetails = &pendingDetailsSelection{data: msg.Data, groupData: msg.GroupData, activate: msg.Activate}
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

func (m *Model) saveGroupDetailsForPending() tea.Cmd {
	edit, ok := m.details.GroupEdit()
	if !ok {
		m.applyPendingDetailsSelection()
		return nil
	}
	data := m.details.GroupData()
	if data == nil {
		m.applyPendingDetailsSelection()
		return nil
	}
	if m.details.Invalid() {
		m.openInvalidDetailsDialog(data.Project, m.details.InvalidReasons(), false)
		return nil
	}
	if m.parameters.HasDraft(data.Project.ProjectID) {
		return m.editGroupDetailsCmd(data.Project, edit, false, false)
	}
	m.openEditGroupDetailsDialog(data.Project, edit, false)
	return nil
}

func (m *Model) handleConditionDetailsSelection(data *messages.ConditionViewData) tea.Cmd {
	if data == nil {
		return nil
	}
	if !m.detailsVisible || !m.details.Dirty() {
		m.details = m.details.SetConditionData(data)
		m.detailsVisible = true
		m.setActive(panels.Details)
		return nil
	}
	m.pendingDetails = &pendingDetailsSelection{conditionData: data, activate: true}
	if m.details.IsCondition() {
		return m.saveConditionDetailsForPending()
	}
	if m.details.IsGroup() {
		return m.saveGroupDetailsForPending()
	}
	edit, ok := m.details.Edit()
	if !ok {
		m.applyPendingDetailsSelection()
		return nil
	}
	current := m.details.Data()
	if current == nil {
		m.applyPendingDetailsSelection()
		return nil
	}
	if m.details.Invalid() {
		m.openInvalidDetailsDialog(current.Project, m.details.InvalidReasons(), false)
		return nil
	}
	if m.parameters.HasDraft(current.Project.ProjectID) {
		return m.editParameterDetailsCmd(current.Project, edit, false, false, false)
	}
	m.openEditDetailsDialog(current.Project, edit, false, false)
	return nil
}

func (m *Model) saveConditionDetailsForPending() tea.Cmd {
	edit, ok := m.details.ConditionEdit()
	if !ok {
		m.applyPendingDetailsSelection()
		return nil
	}
	data := m.details.ConditionData()
	if data == nil {
		m.applyPendingDetailsSelection()
		return nil
	}
	if m.details.Invalid() {
		m.openInvalidDetailsDialog(data.Project, m.details.InvalidReasons(), false)
		return nil
	}
	if m.parameters.HasDraft(data.Project.ProjectID) || m.conditions.HasDraft(data.Project.ProjectID) {
		return m.conditionDetailsMutationCmd(data.Project, edit, false, false)
	}
	m.openConditionDetailsDialog(data.Project, edit, false)
	return nil
}

func (m *Model) applyPendingDetailsSelection() {
	if m.pendingDetails == nil {
		return
	}
	pending := m.pendingDetails
	m.pendingDetails = nil
	if pending.conditionData != nil {
		data := pending.conditionData
		if refreshed, ok := m.conditions.Condition(data.Project.ProjectID, data.Condition.Name); ok {
			data = refreshed
		}
		m.details = m.details.SetConditionData(data)
		m.detailsVisible = true
		if pending.activate {
			m.setActive(panels.Details)
		}
		return
	}
	if pending.groupData != nil {
		m.applyParameterSelection(messages.ParameterSelectionChangedMsg{GroupData: pending.groupData, Activate: pending.activate})
		return
	}
	data := pending.data
	if data != nil {
		valueLabel := ""
		if data.SelectedValueIdx >= 0 && data.SelectedValueIdx < len(data.Parameter.Values) {
			valueLabel = data.Parameter.Values[data.SelectedValueIdx].Label
		}
		if refreshed, ok := m.parameters.ParameterViewData(data.Project.ProjectID, data.GroupKey, data.Parameter.Key, valueLabel); ok {
			data = refreshed
		}
	}
	m.applyParameterSelection(messages.ParameterSelectionChangedMsg{Data: data, Activate: pending.activate})
}
