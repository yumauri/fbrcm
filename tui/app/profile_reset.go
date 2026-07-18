package app

import (
	tea "charm.land/bubbletea/v2"
)

// resetWorkspaceForProfile drops every profile-scoped selection and editor
// while preserving the running log panel and terminal layout.
func (m *Model) resetWorkspaceForProfile() tea.Cmd {
	setupModel := m.setup
	logsModel := m.logs
	helpModel := m.help
	width, height := m.width, m.height
	logsHeight, logsSized := m.logsHeight, m.logsSized
	logsMode, logsSaved := m.logsMode, m.logsSaved

	fresh := New(m.svc)
	fresh.setup = setupModel
	fresh.logs = logsModel
	fresh.help = helpModel
	fresh.width = width
	fresh.height = height
	fresh.logsHeight = logsHeight
	fresh.logsSized = logsSized
	fresh.logsMode = logsMode
	fresh.logsSaved = logsSaved
	*m = fresh
	if width > 0 && height > 0 {
		m.applyLayout()
	}

	return tea.Batch(
		m.parameters.Init(),
		m.conditions.Init(),
		m.details.Init(),
	)
}
