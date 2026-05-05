package app

import (
	tea "charm.land/bubbletea/v2"

	"fbrcm/core"
	"fbrcm/tui/components/logs"
	"fbrcm/tui/components/parameters"
	"fbrcm/tui/components/projects"
	"fbrcm/tui/panels"
)

type Model struct {
	svc *core.Core

	projects   projects.Model
	parameters parameters.Model
	logs       logs.Model
	logsHeight int
	active     panels.ID
	prevTop    panels.ID
	capture    panels.ID

	width  int
	height int
}

func New(svc *core.Core) Model {
	m := Model{
		svc:        svc,
		projects:   projects.New(svc),
		parameters: parameters.New(svc),
		logs:       logs.New(svc),
		logsHeight: defaultLogsPanelHeight,
		active:     panels.Projects,
		prevTop:    panels.Projects,
	}

	m.projects = m.projects.SetActive(true)
	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.projects.Init(),
		m.parameters.Init(),
		m.logs.Init(),
	)
}
