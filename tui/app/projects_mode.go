package app

type projectsPanelMode int

const (
	projectsPanelModeExpanded projectsPanelMode = iota
	projectsPanelModeCollapsed
)

func (m *Model) toggleProjectsMode() {
	if m.projectsMode == projectsPanelModeCollapsed {
		m.projectsMode = projectsPanelModeExpanded
	} else {
		m.projectsMode = projectsPanelModeCollapsed
	}

	if m.width > 0 && m.height > 0 {
		m.applyLayout()
	}
}

func (m *Model) setProjectsMode(mode projectsPanelMode) {
	if m.projectsMode == mode {
		return
	}
	m.projectsMode = mode
	if m.width > 0 && m.height > 0 {
		m.applyLayout()
	}
}
