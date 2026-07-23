package app

import (
	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/tui/components/workspaceheader"
	"github.com/yumauri/fbrcm/tui/panels"
)

func (m Model) workspaceTabAt(x, y int) (panels.ID, bool) {
	if m.promote.WorkspaceOpen() {
		return panels.None, false
	}
	if y != 0 || (m.detailsVisible && m.details.Contains(x, y)) {
		return panels.None, false
	}
	layout := newPanelLayout(m.width, m.height, m.projects.PreferredWidth(), m.logsHeight, m.projectsMode)
	if x < layout.leftWidth || x >= layout.leftWidth+layout.rightWidth {
		return panels.None, false
	}
	index, ok := workspaceheader.TabAt(layout.rightWidth, workspaceTabIndex(m.selectedParametersTab()), x-layout.leftWidth)
	if !ok {
		return panels.None, false
	}
	return workspacePanel(index)
}

func (m Model) activateWorkspacePanel(panel panels.ID) (Model, tea.Cmd, bool) {
	if m.promote.WorkspaceOpen() {
		return m, nil, true
	}
	if workspaceTabIndex(panel) < 0 {
		return m, nil, false
	}
	m.setActive(panel)
	if panel == panels.History {
		var cmd tea.Cmd
		m.parameters, cmd = m.parameters.LoadHistory()
		return m, cmd, true
	}
	return m, nil, true
}

func workspaceTabIndex(panel panels.ID) int {
	switch panel {
	case panels.Parameters:
		return 0
	case panels.Conditions:
		return 1
	case panels.History:
		return 2
	default:
		return -1
	}
}

func workspacePanel(index int) (panels.ID, bool) {
	switch index {
	case 0:
		return panels.Parameters, true
	case 1:
		return panels.Conditions, true
	case 2:
		return panels.History, true
	default:
		return panels.None, false
	}
}
