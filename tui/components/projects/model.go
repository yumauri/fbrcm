package projects

import (
	"context"
	"time"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/tui/components/filterbox"
	"github.com/yumauri/fbrcm/tui/messages"
)

// Model holds model state used by the projects package.
type Model struct {
	// svc stores svc for Model.
	svc *core.Core

	// allProjects stores all projects for Model.
	allProjects []core.Project
	// projects stores projects for Model.
	projects []core.Project
	// source stores source for Model.
	source string
	// err stores err for Model.
	err error
	// loading stores loading for Model.
	loading bool
	// spinner stores spinner for Model.
	spinner spinner.Model
	// viewport stores viewport for Model.
	viewport viewport.Model
	// filter stores filter for Model.
	filter filterbox.Model
	// active stores active for Model.
	active bool
	// collapsed stores collapsed for Model.
	collapsed bool
	// x stores x for Model.
	x int
	// y stores y for Model.
	y int
	// width stores width for Model.
	width int
	// height stores height for Model.
	height int
	// cursor stores cursor for Model.
	cursor int
	// selected stores selected for Model.
	selected map[string]struct{}
	// lastClick stores last click for Model.
	lastClick struct {
		project int
		at      time.Time
	}

	// lines stores lines for Model.
	lines []string
	// lineKinds stores line kinds for Model.
	lineKinds []lineKind
	// lineProjects stores line projects for Model.
	lineProjects []int
	// lineHighlights stores line highlights for Model.
	lineHighlights [][]int
	// projectStarts stores project starts for Model.
	projectStarts []int
	// projectEnds stores project ends for Model.
	projectEnds []int
}

// New constructs new and returns the resulting value or error.
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

// Init initializes init for Model and returns the resulting state or error.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.listProjectsCmd(),
		m.spinner.Tick,
	)
}

// SetSize sets size for Model and returns the resulting state or error.
func (m Model) SetSize(width, height int) Model {
	if m.width == width && m.height == height {
		return m
	}
	m.width = width
	m.height = height
	m.syncViewport()
	return m
}

// SetBounds sets bounds for Model and returns the resulting state or error.
func (m Model) SetBounds(x, y, width, height int) Model {
	if m.x == x && m.y == y && m.width == width && m.height == height {
		return m
	}
	m.x = x
	m.y = y
	m.width = width
	m.height = height
	m.syncViewport()
	return m
}

// SetActive sets active for Model and returns the resulting state or error.
func (m Model) SetActive(active bool) Model {
	m.active = active
	if !active {
		m.filter.Blur()
	}
	return m
}

// SetCollapsed sets collapsed for Model and returns the resulting state or error.
func (m Model) SetCollapsed(collapsed bool) Model {
	m.collapsed = collapsed
	if collapsed {
		m.filter.Blur()
	}
	return m
}

// viewportWidth handles viewport width for Model and returns the resulting state or error.
func (m Model) viewportWidth() int {
	width := m.width - 1
	return max(width, 1)
}

// viewportHeight handles viewport height for Model and returns the resulting state or error.
func (m Model) viewportHeight() int {
	height := m.height - 2 - m.filter.Height()
	return max(height, 1)
}

// syncViewport handles sync viewport for Model and returns the resulting state or error.
func (m *Model) syncViewport() {
	m.refreshViewport()
	m.ensureCursorVisible()
}

// refreshViewport handles refresh viewport for Model and returns the resulting state or error.
func (m *Model) refreshViewport() {
	m.viewport.SetWidth(m.viewportWidth())
	m.viewport.SetHeight(m.viewportHeight())
	m.contentLines()
	m.viewport.SetContentLines(m.renderContentLines())
}

// PreferredWidth handles preferred width for Model and returns the resulting state or error.
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

// listProjectsCmd lists list projects cmd for Model and returns the resulting state or error.
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

// syncProjectsCmd handles sync projects cmd for Model and returns the resulting state or error.
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
