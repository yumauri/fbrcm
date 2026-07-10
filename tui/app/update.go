package app

import (
	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/tui/panels"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if next, cmd, ok := m.updateOpenModal(msg); ok {
		return next, cmd
	}

	next, cmd, ok := m.updateAppMessage(msg)
	m = next
	if ok {
		return next, cmd
	}

	return m.updateChildPanels(msg)
}

func (m *Model) closeDetailsIfOrphaned() {
	if !m.detailsVisible {
		return
	}
	data := m.details.Data()
	if data == nil {
		return
	}
	if m.parameters.HasProject(data.Project.ProjectID) {
		return
	}
	m.closeDetailsPanel()
}

func (m *Model) applyLayout() {
	layout := newPanelLayout(m.width, m.height, m.projects.PreferredWidth(), m.logsHeight, m.projectsMode)

	m.projects = m.projects.SetCollapsed(m.projectsMode == projectsPanelModeCollapsed)
	m.projects = m.projects.SetBounds(0, 0, layout.leftWidth, layout.topHeight)
	m.parameters = m.parameters.SetBounds(layout.leftWidth, 0, layout.rightWidth, layout.topHeight)
	m.dialog = m.dialog.SetBounds(0, 0, m.width, m.height)
	detailsWidth := m.detailsWidthForLayout(layout)
	m.details = m.details.SetBounds(layout.bottomWidth-detailsWidth, 0, detailsWidth, layout.topHeight)
	m.logs = m.logs.SetBounds(0, layout.topHeight, layout.bottomWidth, layout.bottomHeight)
}

func (m Model) nextTabPanel() panels.ID {
	if m.active == panels.Logs {
		if m.detailsVisible {
			if m.prevTop == panels.Details || m.prevTop == panels.Parameters {
				return m.prevTop
			}
			return panels.Parameters
		}
		return m.prevTop
	}

	if m.detailsVisible {
		if m.active == panels.Details {
			return panels.Parameters
		}
		if m.active == panels.Parameters {
			return panels.Details
		}
		return panels.Parameters
	}

	if m.active == panels.Parameters {
		return panels.Projects
	}

	return panels.Parameters
}

func (m *Model) setActive(panel panels.ID) {
	if panel != panels.Logs {
		m.prevTop = panel
	}
	m.active = panel
	if m.capture != panels.None && m.capture != panel {
		m.capture = panels.None
	}
	m.projects = m.projects.SetActive(panel == panels.Projects)
	m.parameters = m.parameters.SetActive(panel == panels.Parameters)
	m.details = m.details.SetActive(panel == panels.Details)
	m.details = m.details.SetBridgeActive(panel == panels.Parameters)
	m.logs = m.logs.SetActive(panel == panels.Logs)
}

func (m Model) keyboardCaptured() bool {
	return m.capture != panels.None
}

func (m Model) panelAt(x, y int) (panels.ID, bool) {
	layout := newPanelLayout(m.width, m.height, m.projects.PreferredWidth(), m.logsHeight, m.projectsMode)
	if x < 0 || y < 0 || x >= layout.bottomWidth || y >= layout.topHeight+layout.bottomHeight {
		return 0, false
	}

	if m.detailsVisible && m.details.Contains(x, y) {
		return panels.Details, true
	}

	if y < layout.topHeight {
		if x < layout.leftWidth {
			return panels.Projects, true
		}
		return panels.Parameters, true
	}

	return panels.None, false
}

func (m *Model) resizeLogsHeight(delta int) {
	layout := newPanelLayout(m.width, m.height, m.projects.PreferredWidth(), m.logsHeight, m.projectsMode)
	m.logsHeight = nextLogsPanelHeight(layout.bottomHeight, delta)
	m.logsHeight = min(m.logsHeight, newPanelLayout(m.width, m.height, m.projects.PreferredWidth(), m.logsHeight, m.projectsMode).bottomHeight)
	if m.width > 0 && m.height > 0 {
		m.applyLayout()
	}
}

func nextLogsPanelHeight(current, delta int) int {
	if delta == 0 {
		return current
	}
	if delta > 0 {
		if current == collapsedLogsPanelHeight {
			return minLogsPanelHeight
		}
		return current + 1
	}
	if current == minLogsPanelHeight {
		return collapsedLogsPanelHeight
	}
	if current == collapsedLogsPanelHeight {
		return collapsedLogsPanelHeight
	}
	return current - 1
}

func initialLogsPanelHeight(terminalHeight int) int {
	if terminalHeight <= 35 {
		return collapsedLogsPanelHeight
	}
	if terminalHeight >= 40 {
		return defaultLogsPanelHeight
	}
	return terminalHeight - 33
}

func (m Model) detailsWidthForLayout(layout panelLayout) int {
	minWidth := max(layout.rightWidth/2, 1)
	maxWidth := max(layout.rightWidth-11, 1)

	// Details content width = panel width - 5:
	// bridge spacer, left border, left padding, right padding, scrollbar lane.
	nameFitWidth := m.parameters.LongestParameterNameWidth() + 5

	desired := max(minWidth, nameFitWidth)
	return min(desired, maxWidth)
}
