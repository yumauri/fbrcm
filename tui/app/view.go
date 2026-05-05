package app

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"fbrcm/tui/panels"
)

var rootStyle = lipgloss.NewStyle()

func (m Model) View() tea.View {
	topRow := lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.projects.View(m.active == panels.Projects),
		m.parameters.View(m.active == panels.Parameters),
	)

	body := lipgloss.JoinVertical(
		lipgloss.Left,
		topRow,
		m.logs.View(m.active == panels.Logs),
	)

	v := tea.NewView(rootStyle.Render(body))
	v.AltScreen = true
	if m.active == panels.Logs {
		v.MouseMode = tea.MouseModeNone
	} else {
		v.MouseMode = tea.MouseModeCellMotion
	}
	return v
}
