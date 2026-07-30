package app

import "github.com/yumauri/fbrcm/tui/panels"

func (m *Model) toggleWorkspaceMaximize() {
	if !isWorkspacePanel(m.active) && m.active != panels.Promote {
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
