package projects

import (
	"context"
	"time"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/tui/components/filterbox"
	"github.com/yumauri/fbrcm/tui/messages"
)

type Model struct {
	svc *core.Core

	allProjects            []core.Project
	projects               []core.Project
	source                 string
	notice                 string
	err                    error
	loading                bool
	spinner                spinner.Model
	viewport               viewport.Model
	filter                 filterbox.Model
	expressionConfigs      map[string]*firebase.RemoteConfig
	expressionOverrides    map[string]*firebase.RemoteConfig
	expressionConfigsReady bool
	active                 bool
	collapsed              bool
	x                      int
	y                      int
	width                  int
	height                 int
	cursor                 int
	selected               map[string]struct{}
	lastClick              struct {
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
		svc:                 svc,
		viewport:            vp,
		filter:              filterbox.New(),
		expressionConfigs:   make(map[string]*firebase.RemoteConfig),
		expressionOverrides: make(map[string]*firebase.RemoteConfig),
		loading:             true,
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

func (m Model) SetActive(active bool) Model {
	m.active = active
	if !active {
		m.filter.Blur()
	}
	return m
}

func (m Model) SetCollapsed(collapsed bool) Model {
	m.collapsed = collapsed
	if collapsed {
		m.filter.Blur()
	}
	return m
}

// SetNotice sets a non-project status line shown above the project list.
func (m Model) SetNotice(notice string) Model {
	m.notice = notice
	m.syncViewport()
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
	key := panelTitleKey()
	longest := lipgloss.Width(key + panelTitleLabel)
	for _, project := range m.allProjects {
		longest = max(longest, lipgloss.Width(project.Name))
		projectID := " " + project.ProjectID
		if project.Disabled {
			projectID += " · disabled"
		}
		longest = max(longest, lipgloss.Width(projectID))
	}

	mainTitleWidth := lipgloss.Width(" " + key + panelTitleLabel + " ")
	secondaryWidth := max(lipgloss.Width(" "+m.secondaryTitleText()+" "), 3)
	headerWidth := 3 + mainTitleWidth + 2 + secondaryWidth + 2 + 1

	// left padding + right padding + right border
	return max(max(longest+3, headerWidth), 25)
}

// HasCurrentProject reports whether project actions have a current target.
func (m Model) HasCurrentProject() bool {
	_, ok := m.CurrentProject()
	return ok
}

// CurrentProject returns the project under the Projects panel cursor.
func (m Model) CurrentProject() (core.Project, bool) {
	if len(m.projects) == 0 || m.cursor < 0 || m.cursor >= len(m.projects) {
		return core.Project{}, false
	}
	return m.projects[m.cursor], true
}

// CurrentProjectEnabled reports whether the project under the cursor can be
// used by project actions that require Firebase access.
func (m Model) CurrentProjectEnabled() bool {
	project, ok := m.CurrentProject()
	return ok && !project.Disabled
}

// ActionTargets returns marked projects, or the current project when nothing
// is marked. Project-level batch actions share this targeting convention.
func (m Model) ActionTargets() []core.Project {
	if selected := m.selectedProjects(); len(selected) > 0 {
		return selected
	}
	project, ok := m.CurrentProject()
	if !ok {
		return nil
	}
	return []core.Project{project}
}

// AuthBindingAvailable reports whether every action target is enabled and at
// least two auth identities discovered every target.
func (m Model) AuthBindingAvailable() bool {
	targets := m.ActionTargets()
	if len(targets) == 0 {
		return false
	}
	common := make(map[string]struct{}, len(targets[0].DiscoveredBy))
	for _, authID := range targets[0].DiscoveredBy {
		common[authID] = struct{}{}
	}
	for _, project := range targets {
		if project.Disabled {
			return false
		}
		discovered := make(map[string]struct{}, len(project.DiscoveredBy))
		for _, authID := range project.DiscoveredBy {
			discovered[authID] = struct{}{}
		}
		for authID := range common {
			if _, ok := discovered[authID]; !ok {
				delete(common, authID)
			}
		}
	}
	return len(common) >= 2
}

// ApplyProjectUpdates replaces matching cached project values and notifies
// downstream panels when their selected projects changed.
func (m *Model) ApplyProjectUpdates(updates []core.Project) tea.Cmd {
	byID := make(map[string]core.Project, len(updates))
	for _, project := range updates {
		byID[project.ProjectID] = project
	}
	for i := range m.allProjects {
		if project, ok := byID[m.allProjects[i].ProjectID]; ok {
			m.allProjects[i] = project
		}
	}
	for i := range m.projects {
		if project, ok := byID[m.projects[i].ProjectID]; ok {
			m.projects[i] = project
		}
	}
	selectionChanged := m.dropDisabledSelections()
	m.syncViewport()
	if len(m.selected) == 0 && !selectionChanged {
		return nil
	}
	return m.selectionChangedCmd()
}

// RemoveProjects deletes matching projects from the panel and notifies
// downstream panels when a selected project was removed.
func (m *Model) RemoveProjects(projects []core.Project) tea.Cmd {
	ids := make(map[string]struct{}, len(projects))
	for _, project := range projects {
		ids[project.ProjectID] = struct{}{}
	}
	if len(ids) == 0 {
		return nil
	}

	selectionChanged := false
	for projectID := range ids {
		if _, ok := m.selected[projectID]; ok {
			delete(m.selected, projectID)
			selectionChanged = true
		}
	}
	remaining := make([]core.Project, 0, len(m.allProjects))
	for _, project := range m.allProjects {
		if _, deleted := ids[project.ProjectID]; !deleted {
			remaining = append(remaining, project)
		}
	}
	m.allProjects = remaining
	m.applyFilter()
	m.syncViewport()
	if selectionChanged {
		return m.selectionChangedCmd()
	}
	return nil
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

func (m Model) loadExpressionConfigsCmd() tea.Cmd {
	projects := append([]core.Project(nil), m.allProjects...)
	return func() tea.Msg {
		configs := make(map[string]*firebase.RemoteConfig, len(projects))
		for _, project := range projects {
			cfg, err := m.svc.LoadCachedRemoteConfig(project.ProjectID)
			if err != nil {
				continue
			}
			configs[project.ProjectID] = cfg
		}
		return messages.ProjectExpressionConfigsLoadedMsg{Configs: configs}
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
