package app

import "github.com/yumauri/fbrcm/tui/panels"

func (m *Model) toggleParametersMaximize() {
	if m.active != panels.Parameters {
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
