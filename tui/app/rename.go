package app

import (
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/tui/components/parameters"
)

// openRenameInput opens open rename input for Model and returns the resulting state or error.
func (m *Model) openRenameInput() tea.Cmd {
	anchor, ok := m.parameters.CurrentRenameAnchor()
	if !ok {
		return nil
	}
	m.closeDialog(false)
	m.closeJSONInput()
	m.closeBoolPicker()
	m.closeNumberInput()
	m.closeStringInput()
	var cmd tea.Cmd
	m.renameInput, cmd = m.renameInput.Open(anchor.X, anchor.Y, anchor.Width, anchor.MaxWidth, anchor.Label)
	return cmd
}

// openDuplicateInput opens open duplicate input for Model and returns the resulting state or error.
func (m *Model) openDuplicateInput() tea.Cmd {
	project, groupKey, paramKey, ok := m.parameters.CurrentParameterRef()
	if !ok {
		return nil
	}
	m.closeDialog(false)
	m.closeMoveParam()
	label := paramKey + "_copy"
	m.parameters.OpenTransientDuplicate(project.ProjectID, groupKey, paramKey, label)
	m.duplicate = &duplicateSession{
		project:        project,
		groupKey:       groupKey,
		sourceParamKey: paramKey,
		visibleName:    label,
	}
	return m.openRenameInput()
}

// closeRenameInput closes close rename input for Model and returns the resulting state or error.
func (m *Model) closeRenameInput() {
	if !m.renameInput.IsOpen() {
		return
	}
	m.renameInput = m.renameInput.Close()
}

func (m *Model) submitRenameInput() tea.Cmd {
	anchor, ok := m.parameters.CurrentRenameAnchor()
	if !ok {
		m.closeRenameInput()
		return nil
	}
	nextParamKey := strings.TrimSpace(m.renameInput.Value())
	if nextParamKey == "" {
		if anchor.IsGroup {
			m.openErrorDialog("Rename Group Failed", anchor.Project, "group name is empty")
		} else {
			title := "Rename Failed"
			if m.activeDuplicate(anchor) {
				title = "Duplicate Failed"
			}
			m.openErrorDialog(title, anchor.Project, "invalid name")
		}
		return nil
	}
	if m.activeDuplicate(anchor) {
		session := *m.duplicate
		if _, _, err := m.svc.PreviewDuplicateParameter(session.project.ProjectID, session.groupKey, session.sourceParamKey, nextParamKey); err != nil {
			m.openErrorDialog("Duplicate Failed", anchor.Project, err.Error())
			return nil
		}
		if m.parameters.HasDraft(session.project.ProjectID) {
			m.closeRenameInput()
			return m.duplicateParameterNamedCmd(session.project, session.groupKey, session.sourceParamKey, nextParamKey, false)
		}
		m.openDuplicateDialog(session.project, session.groupKey, session.sourceParamKey, nextParamKey)
		return nil
	}
	if nextParamKey == anchor.Label && unchangedRenameAnchor(anchor) {
		m.closeRenameInput()
		return nil
	}
	if anchor.IsGroup {
		if m.parameters.HasDraft(anchor.Project.ProjectID) {
			if _, _, err := m.svc.PreviewRenameGroup(anchor.Project.ProjectID, anchor.GroupKey, nextParamKey); err != nil {
				m.openErrorDialog("Rename Group Failed", anchor.Project, err.Error())
				return nil
			}
			m.closeRenameInput()
			return m.renameGroupCmd(anchor.Project, anchor.GroupKey, nextParamKey, false)
		}
		if _, _, err := m.svc.PreviewRenameGroup(anchor.Project.ProjectID, anchor.GroupKey, nextParamKey); err != nil {
			m.openErrorDialog("Rename Group Failed", anchor.Project, err.Error())
			return nil
		}
		m.closeRenameInput()
		m.openRenameGroupDialog(anchor.Project, anchor.GroupKey, nextParamKey)
		return nil
	}
	if m.parameters.HasDraft(anchor.Project.ProjectID) {
		if _, _, err := m.svc.PreviewRenameParameter(anchor.Project.ProjectID, anchor.GroupKey, anchor.ParamKey, nextParamKey); err != nil {
			m.openErrorDialog("Rename Failed", anchor.Project, err.Error())
			return nil
		}
		m.closeRenameInput()
		return m.renameParameterCmd(anchor.Project, anchor.GroupKey, anchor.ParamKey, nextParamKey, false)
	}
	if _, _, err := m.svc.PreviewRenameParameter(anchor.Project.ProjectID, anchor.GroupKey, anchor.ParamKey, nextParamKey); err != nil {
		m.openErrorDialog("Rename Failed", anchor.Project, err.Error())
		return nil
	}
	m.closeRenameInput()
	m.openRenameDialog(anchor.Project, anchor.GroupKey, anchor.ParamKey, nextParamKey)
	return nil
}

// closeRenameIfOrphaned closes close rename if orphaned for Model and returns the resulting state or error.
func (m *Model) closeRenameIfOrphaned() {
	if !m.renameInput.IsOpen() {
		return
	}
	if _, ok := m.parameters.CurrentRenameAnchor(); ok {
		return
	}
	m.closeRenameInput()
}

// cancelRenameInput reports whether cancel rename input for Model and returns the resulting state or error.
func (m *Model) cancelRenameInput() tea.Cmd {
	anchor, ok := m.parameters.CurrentRenameAnchor()
	if m.activeDuplicate(anchor) && ok {
		m.duplicate = nil
		m.parameters.ClearTransientDuplicateAndFocusSource()
		m.closeRenameInput()
		return nil
	}
	m.duplicate = nil
	m.closeRenameInput()
	return nil
}

func (m Model) activeDuplicate(anchor parameters.RenameAnchor) bool {
	return m.duplicate != nil &&
		!anchor.IsGroup &&
		m.duplicate.project.ProjectID == anchor.Project.ProjectID &&
		m.duplicate.groupKey == anchor.GroupKey &&
		m.duplicate.visibleName == anchor.Label
}

func unchangedRenameAnchor(anchor parameters.RenameAnchor) bool {
	if anchor.IsGroup {
		return anchor.GroupKey == anchor.Label
	}
	return anchor.ParamKey == anchor.Label
}
