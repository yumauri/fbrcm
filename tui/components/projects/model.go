package projects

import (
	"context"
	"time"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"fbrcm/core"
	"fbrcm/tui/components/filterbox"
	"fbrcm/tui/messages"
)

type Model struct {
	svc *core.Core

	allProjects []core.Project
	projects    []core.Project
	source      string
	err         error
	loading     bool
	spinner     spinner.Model
	viewport    viewport.Model
	filter      filterbox.Model
	active      bool
	x           int
	y           int
	width       int
	height      int
	cursor      int
	selected    map[string]struct{}
	lastClick   struct {
		project int
		at      time.Time
	}

	lines          []string
	lineKinds      []lineKind
	lineProjects   []int
	lineHighlights [][]int
	projectStarts  []int
	projectEnds    []int
}

func New(svc *core.Core) Model {
	vp := viewport.New(
		viewport.WithWidth(1),
		viewport.WithHeight(1),
	)
	vp.SoftWrap = false
	return Model{
		svc:      svc,
		viewport: vp,
		filter:   filterbox.New(),
		loading:  true,
		spinner: spinner.New(
			spinner.WithSpinner(spinner.Line),
		),
		selected: make(map[string]struct{}),
		lastClick: struct {
			project int
			at      time.Time
		}{
			project: -1,
		},
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.listProjectsCmd(),
		m.spinner.Tick,
	)
}

func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height
	m.syncViewport()
	return m
}

func (m Model) SetBounds(x, y, width, height int) Model {
	m.x = x
	m.y = y
	m.width = width
	m.height = height
	m.syncViewport()
	return m
}

func (m Model) SetActive(active bool) Model {
	m.active = active
	if !active {
		m.filter.Blur()
	}
	return m
}

func (m Model) viewportWidth() int {
	width := m.width - 1
	return max(width, 1)
}

func (m Model) viewportHeight() int {
	height := m.height - 2 - m.filter.Height()
	return max(height, 1)
}

func (m *Model) syncViewport() {
	m.refreshViewport()
	m.ensureCursorVisible()
}

func (m *Model) refreshViewport() {
	m.viewport.SetWidth(m.viewportWidth())
	m.viewport.SetHeight(m.viewportHeight())
	m.contentLines()
	m.viewport.SetContentLines(m.renderContentLines())
}

func (m Model) PreferredWidth() int {
	longest := lipgloss.Width(panelTitle)
	for _, project := range m.allProjects {
		longest = max(longest, lipgloss.Width(project.Name))
		longest = max(longest, lipgloss.Width(" "+project.ProjectID))
	}

	mainTitleWidth := lipgloss.Width(" " + panelTitle + " ")
	secondaryWidth := max(lipgloss.Width(" "+m.secondaryTitleText()+" "), 3)
	headerWidth := 3 + mainTitleWidth + 2 + secondaryWidth + 2 + 1

	// left padding + right padding + right border
	return max(max(longest+3, headerWidth), 25)
}

func (m Model) listProjectsCmd() tea.Cmd {
	return func() tea.Msg {
		projects, source, err := m.svc.ListProjects(context.Background())
		return messages.ProjectsLoadedMsg{
			Projects: projects,
			Source:   source,
			Err:      err,
		}
	}
}

func (m Model) syncProjectsCmd() tea.Cmd {
	return func() tea.Msg {
		projects, source, err := m.svc.SyncProjects(context.Background())
		return messages.ProjectsLoadedMsg{
			Projects: projects,
			Source:   source,
			Err:      err,
		}
	}
}
