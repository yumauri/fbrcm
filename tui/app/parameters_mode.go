package app

import "github.com/yumauri/fbrcm/tui/panels"

func (m *Model) toggleWorkspaceMaximize() {
	if m.active != panels.Parameters && m.active != panels.Conditions && m.active != panels.History {
		return
	}

	if m.projectsMode == projectsPanelModeCollapsed && m.logsMode == logsPanelModeCollapsed {
		m.setProjectsMode(projectsPanelModeExpanded)
		m.setLogsMode(logsPanelModeExpanded)
		return
	}

	m.setProjectsMode(projectsPanelModeCollapsed)
	m.setLogsMode(logsPanelModeCollapsed)
}
