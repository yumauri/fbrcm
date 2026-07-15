package app

import (
	tea "charm.land/bubbletea/v2"

	tuiconfig "github.com/yumauri/fbrcm/tui/config"
	"github.com/yumauri/fbrcm/tui/panels"
)

func (m Model) updateDetailsKeyMessage(msg tea.KeyMsg, k string) (Model, tea.Cmd, bool) {
	if !m.details.FieldActive() || tuiconfig.Matches(tuiconfig.BlockGlobal, tuiconfig.ActionFocusNext, k) {
		if next, cmd, ok := m.updateGlobalFocusKey(k); ok {
			return next, cmd, true
		}
	}
	if m.details.IsCondition() {
		return m.updateConditionDetailsKeyMessage(msg, k)
	}
	switch {
	case tuiconfig.Matches(tuiconfig.BlockDetails, tuiconfig.ActionClose, k):
		if m.details.FieldActive() || m.details.ValueSelected() {
			var cmd tea.Cmd
			m.details, cmd = m.details.Update(msg)
			return m, cmd, true
		}
		return m, m.requestCloseDetails(), true
	case tuiconfig.Matches(tuiconfig.BlockDetails, tuiconfig.ActionSubmit, k):
		return m, m.submitDetailsForm(), true
	case tuiconfig.Matches(tuiconfig.BlockDetails, tuiconfig.ActionEditValue, k):
		if m.details.ValueSelected() {
			return m, m.openDetailsValueEditor(), true
		}
	}
	if !m.details.TextInputActive() {
		if next, cmd, ok := m.updateInactiveDetailsInputKey(k); ok {
			return next, cmd, true
		}
	}
	if m.details.FieldActive() {
		var cmd tea.Cmd
		m.details, cmd = m.details.Update(msg)
		return m, cmd, true
	}
	return m, nil, false
}

func (m Model) updateConditionDetailsKeyMessage(msg tea.KeyMsg, k string) (Model, tea.Cmd, bool) {
	switch {
	case tuiconfig.Matches(tuiconfig.BlockGlobal, tuiconfig.ActionQuit, k):
		return m, tea.Quit, true
	case tuiconfig.Matches(tuiconfig.BlockDetails, tuiconfig.ActionClose, k):
		m.closeDetailsPanel()
		return m, nil, true
	case tuiconfig.Matches(tuiconfig.BlockDetails, tuiconfig.ActionCopyName, k):
		return m, m.copyDetailsNameCmd(), true
	case tuiconfig.Matches(tuiconfig.BlockDetails, tuiconfig.ActionCopyPath, k):
		return m, m.copyDetailsPathCmd(), true
	case tuiconfig.Matches(tuiconfig.BlockDetails, tuiconfig.ActionCopyValue, k):
		return m, m.copyDetailsSelectedValueCmd(), true
	}
	var cmd tea.Cmd
	m.details, cmd = m.details.Update(msg)
	return m, cmd, true
}

func (m Model) updateInactiveDetailsInputKey(k string) (Model, tea.Cmd, bool) {
	switch {
	case tuiconfig.Matches(tuiconfig.BlockGlobal, tuiconfig.ActionQuit, k):
		return m, tea.Quit, true
	case tuiconfig.Matches(tuiconfig.BlockDetails, tuiconfig.ActionMove, k):
		return m, m.activateDetailsGroup(), true
	case tuiconfig.Matches(tuiconfig.BlockDetails, tuiconfig.ActionRename, k):
		var cmd tea.Cmd
		m.details, cmd = m.details.ActivateName()
		return m, cmd, true
	case tuiconfig.Matches(tuiconfig.BlockDetails, tuiconfig.ActionCopyName, k):
		return m, m.copyDetailsNameCmd(), true
	case tuiconfig.Matches(tuiconfig.BlockDetails, tuiconfig.ActionCopyPath, k):
		return m, m.copyDetailsPathCmd(), true
	case tuiconfig.Matches(tuiconfig.BlockDetails, tuiconfig.ActionCopyValue, k):
		if m.details.ValueSelected() {
			return m, m.copyDetailsSelectedValueCmd(), true
		}
	case tuiconfig.Matches(tuiconfig.BlockDetails, tuiconfig.ActionDelete, k):
		return m, m.requestDeleteDetails(), true
	}
	return m, nil, false
}

func (m Model) updateGlobalKeyMessage(k string) (Model, tea.Cmd, bool) {
	if tuiconfig.Matches(tuiconfig.BlockGlobal, tuiconfig.ActionQuit, k) {
		return m, tea.Quit, true
	}
	if tuiconfig.Matches(tuiconfig.BlockDetails, tuiconfig.ActionClose, k) {
		if m.active == panels.Details && m.detailsVisible {
			m.detailsVisible = false
			m.setActive(panels.Parameters)
		}
		return m, nil, false
	}
	if next, cmd, ok := m.updateGlobalFocusKey(k); ok {
		return next, cmd, true
	}

	switch {
	case tuiconfig.Matches(tuiconfig.BlockProjects, tuiconfig.ActionToggleMode, k), tuiconfig.Matches(tuiconfig.BlockLogs, tuiconfig.ActionToggleMode, k), tuiconfig.Matches(tuiconfig.BlockParameters, tuiconfig.ActionDuplicate, k):
		return m.updateModeOrDuplicateKey()
	case tuiconfig.Matches(tuiconfig.BlockParameters, tuiconfig.ActionToggleMaximize, k):
		if m.active == panels.Parameters || m.active == panels.Conditions || m.active == panels.History {
			m.toggleWorkspaceMaximize()
			return m, nil, true
		}
	case tuiconfig.Matches(tuiconfig.BlockParameters, tuiconfig.ActionReload, k), tuiconfig.Matches(tuiconfig.BlockParameters, tuiconfig.ActionReloadAll, k):
		if m.active == panels.Conditions {
			return m.updateConditionsReloadKey(k)
		}
	case tuiconfig.Matches(tuiconfig.BlockLogs, tuiconfig.ActionResizeGrow, k):
		m.updateLogsResizeKey(1)
	case tuiconfig.Matches(tuiconfig.BlockLogs, tuiconfig.ActionResizeShrink, k):
		m.updateLogsResizeKey(-1)
	default:
		return m.updateGlobalPanelActionKey(k)
	}
	return m, nil, false
}

func (m Model) updateConditionsReloadKey(k string) (Model, tea.Cmd, bool) {
	if tuiconfig.Matches(tuiconfig.BlockParameters, tuiconfig.ActionReloadAll, k) {
		cmd := m.parameters.ReloadAllProjects()
		if cmd != nil {
			var spinCmd tea.Cmd
			m.conditions, spinCmd = m.conditions.MarkAllReloading()
			cmd = tea.Batch(cmd, spinCmd)
		}
		return m, cmd, true
	}
	project, ok := m.conditions.CurrentProject()
	if !ok {
		return m, nil, true
	}
	cmd := m.parameters.ReloadProject(project.ProjectID)
	if cmd != nil {
		var spinCmd tea.Cmd
		m.conditions, spinCmd = m.conditions.MarkProjectReloading(project.ProjectID)
		cmd = tea.Batch(cmd, spinCmd)
	}
	return m, cmd, true
}

func (m Model) updateGlobalFocusKey(k string) (Model, tea.Cmd, bool) {
	switch {
	case tuiconfig.Matches(tuiconfig.BlockGlobal, tuiconfig.ActionFocusProjects, k):
		m.setActive(panels.Projects)
	case tuiconfig.Matches(tuiconfig.BlockGlobal, tuiconfig.ActionFocusParameters, k):
		m.setActive(panels.Parameters)
	case tuiconfig.Matches(tuiconfig.BlockGlobal, tuiconfig.ActionFocusConditions, k):
		m.setActive(panels.Conditions)
	case tuiconfig.Matches(tuiconfig.BlockGlobal, tuiconfig.ActionFocusHistory, k):
		m.setActive(panels.History)
		var cmd tea.Cmd
		m.parameters, cmd = m.parameters.LoadHistory()
		return m, cmd, true
	case tuiconfig.Matches(tuiconfig.BlockGlobal, tuiconfig.ActionFocusDetails, k):
		if !m.detailsVisible {
			return m, nil, false
		}
		m.setActive(panels.Details)
	case tuiconfig.Matches(tuiconfig.BlockGlobal, tuiconfig.ActionFocusLogs, k):
		m.setActive(panels.Logs)
	case tuiconfig.Matches(tuiconfig.BlockGlobal, tuiconfig.ActionFocusNext, k):
		m.setActive(m.nextTabPanel())
	default:
		return m, nil, false
	}
	return m, nil, true
}

func (m Model) updateGlobalPanelActionKey(k string) (Model, tea.Cmd, bool) {
	switch {
	case tuiconfig.Matches(tuiconfig.BlockParameters, tuiconfig.ActionDelete, k), tuiconfig.Matches(tuiconfig.BlockDetails, tuiconfig.ActionDelete, k):
		return m.updateDeleteKey()
	case tuiconfig.Matches(tuiconfig.BlockParameters, tuiconfig.ActionRename, k):
		if m.active == panels.Parameters {
			return m, m.openRenameInput(), true
		}
	case tuiconfig.Matches(tuiconfig.BlockParameters, tuiconfig.ActionNew, k):
		if m.active == panels.Parameters {
			return m, m.openNewParameterDetails(), true
		}
	case tuiconfig.Matches(tuiconfig.BlockParameters, tuiconfig.ActionEdit, k):
		if m.active == panels.Parameters {
			return m.updateParameterEditKey()
		}
	case tuiconfig.Matches(tuiconfig.BlockParameters, tuiconfig.ActionMove, k):
		if m.active == panels.Parameters {
			m.openMoveParam()
			return m, nil, true
		}
	case tuiconfig.Matches(tuiconfig.BlockParameters, tuiconfig.ActionPublish, k):
		if m.active != panels.Parameters {
			return m, nil, false
		}
		return m.openCurrentDraftDialog(dialogModePublishDraft)
	case tuiconfig.Matches(tuiconfig.BlockParameters, tuiconfig.ActionPublishAll, k):
		if m.active != panels.Parameters {
			return m, nil, false
		}
		return m.openDraftDialogs(dialogModePublishDraft)
	case tuiconfig.Matches(tuiconfig.BlockParameters, tuiconfig.ActionDiscard, k):
		if m.active != panels.Parameters {
			return m, nil, false
		}
		return m.openCurrentDraftDialog(dialogModeDiscardDraft)
	case tuiconfig.Matches(tuiconfig.BlockParameters, tuiconfig.ActionDiscardAll, k):
		if m.active != panels.Parameters {
			return m, nil, false
		}
		return m.openDraftDialogs(dialogModeDiscardDraft)
	}
	return m, nil, false
}

func (m Model) updateModeOrDuplicateKey() (Model, tea.Cmd, bool) {
	if m.active == panels.Projects {
		m.toggleProjectsMode()
	}
	if m.active == panels.Logs {
		m.toggleLogsMode()
	}
	if m.active == panels.Parameters {
		return m, m.openDuplicateInput(), true
	}
	return m, nil, false
}

func (m *Model) updateLogsResizeKey(delta int) {
	if m.active != panels.Logs {
		return
	}
	if delta > 0 && m.logsMode == logsPanelModeCollapsed {
		m.growLogsFromCollapsed()
		return
	}
	m.resizeLogsHeight(delta)
}

func (m Model) updateDeleteKey() (Model, tea.Cmd, bool) {
	if m.active == panels.Parameters {
		return m.updateParametersDeleteKey()
	}
	if m.active == panels.Details && m.detailsVisible {
		return m, m.requestDeleteDetails(), true
	}
	return m, nil, false
}

func (m Model) updateParametersDeleteKey() (Model, tea.Cmd, bool) {
	if anchor, ok := m.parameters.CurrentConditionalValueAnchor(); ok {
		if m.parameters.HasDraft(anchor.Project.ProjectID) {
			return m, m.deleteConditionalValueCmd(anchor.Project, anchor.GroupKey, anchor.ParamKey, anchor.ValueLabel, false), true
		}
		m.openDeleteConditionalValueDialog(anchor.Project, anchor.GroupKey, anchor.ParamKey, anchor.ValueLabel)
		return m, nil, true
	}
	project, groupKey, groupLabel, ok := m.parameters.CurrentGroupRef()
	if ok {
		if m.parameters.HasDraft(project.ProjectID) {
			return m, m.deleteGroupCmd(project, groupKey, false), true
		}
		m.openDeleteGroupDialog(project, groupKey, groupLabel)
		return m, nil, true
	}
	project, groupKey, paramKey, ok := m.parameters.CurrentParameterRef()
	if ok {
		if m.parameters.HasDraft(project.ProjectID) {
			return m, m.deleteParameterCmd(project, groupKey, paramKey, false, false), true
		}
		m.openDeleteDialog(project, groupKey, paramKey, false)
		return m, nil, true
	}
	return m, nil, false
}

func (m Model) updateParameterEditKey() (Model, tea.Cmd, bool) {
	if next, cmd, ok := m.openCurrentParameterValueEditor(); ok {
		return next, cmd, true
	}
	if m.parameters.FocusCurrentParameterDefaultValue() {
		if next, cmd, ok := m.openCurrentParameterValueEditor(); ok {
			return next, cmd, true
		}
	}
	return m, nil, true
}

func (m Model) openCurrentParameterValueEditor() (Model, tea.Cmd, bool) {
	if _, ok := m.parameters.CurrentBoolValueAnchor(); ok {
		return m, m.openBoolPicker(), true
	}
	if _, ok := m.parameters.CurrentNumberValueAnchor(); ok {
		return m, m.openNumberInput(), true
	}
	if _, ok := m.parameters.CurrentJSONValueAnchor(); ok {
		return m, m.openJSONInput(), true
	}
	if _, ok := m.parameters.CurrentStringValueAnchor(); ok {
		return m, m.openStringInput(), true
	}
	return m, nil, false
}

func (m Model) openCurrentDraftDialog(mode dialogMode) (Model, tea.Cmd, bool) {
	if m.active != panels.Parameters {
		return m, nil, false
	}
	project, ok := m.parameters.CurrentProject()
	if ok && m.parameters.HasDraft(project.ProjectID) {
		m.openDraftDialog(project, mode, nil)
		return m, nil, true
	}
	return m, nil, false
}

func (m Model) openDraftDialogs(mode dialogMode) (Model, tea.Cmd, bool) {
	if m.active != panels.Parameters {
		return m, nil, false
	}
	projects := m.parameters.DraftProjects()
	if len(projects) == 0 {
		return m, nil, false
	}
	queue := make([]pendingDialog, 0, len(projects)-1)
	for _, project := range projects[1:] {
		queue = append(queue, pendingDialog{project: project, mode: mode})
	}
	m.openDraftDialog(projects[0], mode, queue)
	return m, nil, true
}
