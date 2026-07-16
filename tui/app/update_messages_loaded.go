package app

import (
	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/panels"
)

func (m *Model) updateDetailsAfterParametersLoaded(msg messages.ParametersLoadedMsg) {
	if msg.DetailsSaved {
		m.clearTransientNewParameterAfterLoad(msg)
		if msg.CloseDetails {
			m.pendingDetails = nil
		} else if m.pendingDetails != nil {
			m.applyPendingDetailsSelection()
		} else {
			m.details = m.details.MarkSaved()
		}
		return
	}
	if m.detailsVisible && msg.SelectParamKey != "" {
		if data, ok := m.parameters.CurrentParameterViewData(); ok && data.Project.ProjectID == msg.Project.ProjectID {
			m.details = m.details.SetData(data)
		}
	}
}

func (m *Model) clearTransientNewParameterAfterLoad(msg messages.ParametersLoadedMsg) {
	if m.newParameter == nil || m.newParameter.projectID != msg.Project.ProjectID {
		return
	}
	m.newParameter = nil
	if msg.SelectParamKey != "" {
		m.parameters.ClearTransientNewParameterAndFocus(msg.Project.ProjectID, msg.SelectGroupKey, msg.SelectParamKey)
		return
	}
	m.parameters.ClearTransientNewParameter()
}

func (m *Model) updateDuplicateAfterParametersLoaded(msg messages.ParametersLoadedMsg) {
	if m.duplicate == nil || m.duplicate.project.ProjectID != msg.Project.ProjectID {
		return
	}
	m.duplicate = nil
	if msg.SelectParamKey != "" {
		m.parameters.ClearTransientDuplicateAndFocus(msg.Project.ProjectID, msg.SelectGroupKey, msg.SelectParamKey)
	} else {
		m.parameters.ClearTransientDuplicate()
	}
	m.closeRenameInput()
}

func (m Model) updateDetailsEditCanceled(msg messages.DetailsEditCanceledMsg) (Model, tea.Cmd, bool) {
	if msg.CloseDetails {
		m.pendingDetails = nil
		m.closeDetailsPanel()
		return m, nil, true
	}
	if m.pendingDetails != nil {
		m.newParameter = nil
		m.parameters.ClearTransientNewParameter()
		m.applyPendingDetailsSelection()
	}
	return m, nil, false
}

func (m *Model) updateDetailsInvalidFix() {
	if m.pendingDetails != nil && m.newParameter != nil {
		m.newParameter = nil
		m.parameters.ClearTransientNewParameter()
	}
	m.pendingDetails = nil
	if data := m.details.Data(); data != nil {
		m.parameters.FocusParameter(data.Project.ProjectID, data.GroupKey, data.Parameter.Key)
	}
	if m.detailsVisible {
		m.setActive(panels.Details)
	}
}

func (m Model) updateDetailsInvalidDiscard(msg messages.DetailsInvalidDiscardMsg) (Model, tea.Cmd, bool) {
	if msg.CloseDetails {
		m.pendingDetails = nil
		m.closeDetailsPanel()
		return m, nil, true
	}
	if m.pendingDetails != nil {
		m.newParameter = nil
		m.parameters.ClearTransientNewParameter()
		m.applyPendingDetailsSelection()
		return m, nil, true
	}
	if data := m.details.Data(); data != nil {
		m.details = m.details.SetData(data)
		m.setActive(panels.Details)
	} else if data := m.details.ConditionData(); data != nil {
		m.details = m.details.SetConditionData(data)
		m.setActive(panels.Details)
	}
	return m, nil, false
}
