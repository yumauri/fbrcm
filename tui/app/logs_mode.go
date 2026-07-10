package app

type logsPanelMode int

const (
	logsPanelModeExpanded logsPanelMode = iota
	logsPanelModeCollapsed
)

func (m *Model) toggleLogsMode() {
	if m.logsMode == logsPanelModeCollapsed {
		m.expandLogsFromCollapsed()
		return
	}

	m.logsSaved = m.logsHeight
	m.logsHeight = collapsedLogsPanelHeight
	m.logsMode = logsPanelModeCollapsed
	if m.width > 0 && m.height > 0 {
		m.applyLayout()
	}
}

func (m *Model) expandLogsFromCollapsed() {
	m.logsMode = logsPanelModeExpanded
	if m.logsSaved > 0 {
		m.logsHeight = m.logsSaved
	}
	m.logsSaved = 0
	layout := newPanelLayout(m.width, m.height, m.projects.PreferredWidth(), m.logsHeight, m.projectsMode)
	m.logsHeight = min(m.logsHeight, layout.bottomHeight)
	if m.width > 0 && m.height > 0 {
		m.applyLayout()
	}
}

func (m *Model) growLogsFromCollapsed() {
	m.logsMode = logsPanelModeExpanded
	m.logsSaved = 0
	m.resizeLogsHeight(1)
}

func (m *Model) setLogsMode(mode logsPanelMode) {
	if m.logsMode == mode {
		return
	}
	if mode == logsPanelModeCollapsed {
		m.toggleLogsMode()
		return
	}
	m.expandLogsFromCollapsed()
}
