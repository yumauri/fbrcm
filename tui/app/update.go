package app

import (
	tea "charm.land/bubbletea/v2"

	"fbrcm/tui/messages"
	"fbrcm/tui/panels"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.KeyboardCaptureMsg:
		if msg.Enabled {
			m.capture = m.active
		} else {
			m.capture = panels.None
		}

	case messages.SetActivePanelMsg:
		m.setActive(msg.Panel)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.applyLayout()

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if !m.keyboardCaptured() {
			switch msg.String() {
			case "q":
				return m, tea.Quit
			case "1":
				m.setActive(panels.Projects)
			case "2":
				m.setActive(panels.Parameters)
			case "0":
				m.setActive(panels.Logs)
			case "=", "+":
				if m.active == panels.Logs {
					m.resizeLogsHeight(1)
				}
			case "-", "_":
				if m.active == panels.Logs {
					m.resizeLogsHeight(-1)
				}
			case "tab":
				m.setActive(m.nextTabPanel())
			}
		}

	case tea.MouseClickMsg:
		if m.active == panels.Logs {
			break
		}
		if panel, ok := m.panelAt(msg.Mouse().X, msg.Mouse().Y); ok {
			m.setActive(panel)
		}

	case tea.MouseWheelMsg:
		if m.active == panels.Logs {
			break
		}
		if panel, ok := m.panelAt(msg.Mouse().X, msg.Mouse().Y); ok {
			m.setActive(panel)
		}
	}

	var cmds []tea.Cmd

	var cmd tea.Cmd
	m.projects, cmd = m.projects.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	if _, ok := msg.(tea.WindowSizeMsg); !ok && m.width > 0 && m.height > 0 {
		m.applyLayout()
	}

	m.parameters, cmd = m.parameters.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	m.logs, cmd = m.logs.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) applyLayout() {
	layout := newPanelLayout(m.width, m.height, m.projects.PreferredWidth(), m.logsHeight)

	m.projects = m.projects.SetBounds(0, 0, layout.leftWidth, layout.topHeight)
	m.parameters = m.parameters.SetBounds(layout.leftWidth, 0, layout.rightWidth, layout.topHeight)
	m.logs = m.logs.SetBounds(0, layout.topHeight, layout.bottomWidth, layout.bottomHeight)
}

func (m Model) nextTabPanel() panels.ID {
	if m.active == panels.Logs {
		return m.prevTop
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
	m.logs = m.logs.SetActive(panel == panels.Logs)
}

func (m Model) keyboardCaptured() bool {
	return m.capture != panels.None
}

func (m Model) panelAt(x, y int) (panels.ID, bool) {
	layout := newPanelLayout(m.width, m.height, m.projects.PreferredWidth(), m.logsHeight)
	if x < 0 || y < 0 || x >= layout.bottomWidth || y >= layout.topHeight+layout.bottomHeight {
		return 0, false
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
	layout := newPanelLayout(m.width, m.height, m.projects.PreferredWidth(), m.logsHeight)
	m.logsHeight = layout.bottomHeight + delta
	if m.height >= minLogsPanelHeight+1 {
		m.logsHeight = max(m.logsHeight, minLogsPanelHeight)
	} else {
		m.logsHeight = max(m.logsHeight, 1)
	}
	m.logsHeight = min(m.logsHeight, max(m.height-1, 1))
	if m.width > 0 && m.height > 0 {
		m.applyLayout()
	}
}
