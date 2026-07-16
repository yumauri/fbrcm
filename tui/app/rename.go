package app

import (
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/tui/components/parameters"
)

func (m *Model) openRenameInput() tea.Cmd {
	anchor, ok := m.parameters.CurrentRenameAnchor()
	if !ok {
		return nil
	}
	m.closeOverlays()
	var cmd tea.Cmd
	m.renameInput, cmd = m.renameInput.Open(anchor.X, anchor.Y, anchor.Width, anchor.MaxWidth, anchor.Label)
	return cmd
}

func (m *Model) openDuplicateInput() tea.Cmd {
	project, groupKey, paramKey, ok := m.parameters.CurrentParameterRef()
	if !ok {
		return nil
	}
	m.closeOverlays()
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

func (m *Model) closeRenameInput() {
	if !m.renameInput.IsOpen() {
		return
	}
	m.renameInput = m.renameInput.Close()
}

func (m *Model) submitRenameInput() tea.Cmd {
	if m.conditionEdit != nil && (m.conditionEdit.mode == conditionAddName || m.conditionEdit.mode == conditionRename) {
		return m.submitConditionRenameInput()
	}
	anchor, ok := m.parameters.CurrentRenameAnchor()
	if !ok {
		m.closeRenameInput()
		return nil
	}
	nextParamKey := strings.TrimSpace(m.renameInput.Value())
	if nextParamKey == "" {
		m.openEmptyRenameError(anchor)
		return nil
	}
	if m.activeDuplicate(anchor) {
		return m.submitDuplicateRename(anchor, nextParamKey)
	}
	if nextParamKey == anchor.Label && unchangedRenameAnchor(anchor) {
		m.closeRenameInput()
		return nil
	}
	if anchor.IsGroup {
		return m.submitGroupRename(anchor, nextParamKey)
	}
	return m.submitParameterRename(anchor, nextParamKey)
}

func (m *Model) openEmptyRenameError(anchor parameters.RenameAnchor) {
	if anchor.IsGroup {
		m.openErrorDialog("Rename Group Failed", anchor.Project, "group name is empty")
		return
	}
	title := "Rename Failed"
	if m.activeDuplicate(anchor) {
		title = "Duplicate Failed"
	}
	m.openErrorDialog(title, anchor.Project, "invalid name")
}

func (m *Model) submitDuplicateRename(anchor parameters.RenameAnchor, nextParamKey string) tea.Cmd {
	session := *m.duplicate
	if !m.previewRename(anchor.Project, "Duplicate Failed", func() error {
		_, _, err := m.svc.PreviewDuplicateParameter(session.project.ProjectID, session.groupKey, session.sourceParamKey, nextParamKey)
		return err
	}) {
		return nil
	}
	if m.parameters.HasDraft(session.project.ProjectID) {
		m.closeRenameInput()
		return m.duplicateParameterNamedCmd(session.project, session.groupKey, session.sourceParamKey, nextParamKey, false)
	}
	m.openDuplicateDialog(session.project, session.groupKey, session.sourceParamKey, nextParamKey)
	return nil
}

func (m *Model) submitGroupRename(anchor parameters.RenameAnchor, nextGroupKey string) tea.Cmd {
	if !m.previewRename(anchor.Project, "Rename Group Failed", func() error {
		_, _, err := m.svc.PreviewRenameGroup(anchor.Project.ProjectID, anchor.GroupKey, nextGroupKey)
		return err
	}) {
		return nil
	}
	m.closeRenameInput()
	if m.parameters.HasDraft(anchor.Project.ProjectID) {
		return m.renameGroupCmd(anchor.Project, anchor.GroupKey, nextGroupKey, false)
	}
	m.openRenameGroupDialog(anchor.Project, anchor.GroupKey, nextGroupKey)
	return nil
}

func (m *Model) submitParameterRename(anchor parameters.RenameAnchor, nextParamKey string) tea.Cmd {
	if !m.previewRename(anchor.Project, "Rename Failed", func() error {
		_, _, err := m.svc.PreviewRenameParameter(anchor.Project.ProjectID, anchor.GroupKey, anchor.ParamKey, nextParamKey)
		return err
	}) {
		return nil
	}
	m.closeRenameInput()
	if m.parameters.HasDraft(anchor.Project.ProjectID) {
		return m.renameParameterCmd(anchor.Project, anchor.GroupKey, anchor.ParamKey, nextParamKey, false)
	}
	m.openRenameDialog(anchor.Project, anchor.GroupKey, anchor.ParamKey, nextParamKey)
	return nil
}

func (m *Model) previewRename(project core.Project, title string, run func() error) bool {
	if err := run(); err != nil {
		m.openErrorDialog(title, project, err.Error())
		return false
	}
	return true
}

func (m *Model) closeRenameIfOrphaned() {
	if !m.renameInput.IsOpen() {
		return
	}
	if _, ok := m.parameters.CurrentRenameAnchor(); ok {
		return
	}
	m.closeRenameInput()
}

func (m *Model) cancelRenameInput() {
	if m.conditionEdit != nil && (m.conditionEdit.mode == conditionAddName || m.conditionEdit.mode == conditionRename) {
		m.conditionEdit = nil
		m.closeRenameInput()
		return
	}
	anchor, ok := m.parameters.CurrentRenameAnchor()
	if m.activeDuplicate(anchor) && ok {
		m.duplicate = nil
		m.parameters.ClearTransientDuplicateAndFocusSource()
		m.closeRenameInput()
		return
	}
	m.duplicate = nil
	m.closeRenameInput()
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
