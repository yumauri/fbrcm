package app

import (
	tea "charm.land/bubbletea/v2"
	moveparam "github.com/yumauri/fbrcm/tui/components/moveparam"
)

func (m *Model) openMoveParam() {
	anchor, ok := m.parameters.CurrentMoveAnchor()
	if !ok {
		project, _ := m.parameters.CurrentProject()
		m.openErrorDialog("Move Failed", project, "No target groups available.")
		return
	}
	m.closeOverlays()
	options := make([]moveparam.Option, 0, len(anchor.Options))
	for _, option := range anchor.Options {
		options = append(options, moveparam.Option{Key: option.Key, Label: option.Label})
	}
	m.moveParam = m.moveParam.Open(anchor.X, anchor.Y, anchor.Label, options)
}

func (m *Model) closeMoveParam() {
	if !m.moveParam.IsOpen() {
		return
	}
	m.moveParam = m.moveParam.Close()
}

func (m *Model) submitMoveParam() tea.Cmd {
	anchor, ok := m.parameters.CurrentMoveAnchor()
	if !ok {
		m.closeMoveParam()
		return nil
	}
	target, ok := m.moveParam.Current()
	m.closeMoveParam()
	if !ok {
		return nil
	}
	if anchor.IsGroup {
		if m.parameters.HasDraft(anchor.Project.ProjectID) {
			return m.moveGroupCmd(anchor.Project, anchor.GroupKey, target.Key, false)
		}
		m.openMoveGroupDialog(anchor.Project, anchor.GroupKey, target.Key)
		return nil
	}
	if m.parameters.HasDraft(anchor.Project.ProjectID) {
		return m.moveParameterCmd(anchor.Project, anchor.GroupKey, anchor.ParamKey, target.Key, false)
	}
	m.openMoveDialog(anchor.Project, anchor.GroupKey, anchor.ParamKey, target.Key)
	return nil
}

func (m *Model) closeMoveIfOrphaned() {
	if !m.moveParam.IsOpen() {
		return
	}
	if _, ok := m.parameters.CurrentMoveAnchor(); ok {
		return
	}
	m.closeMoveParam()
}
