package managedfeatures

import (
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/tui/components/filterbox"
	"github.com/yumauri/fbrcm/tui/components/mouseutil"
	"github.com/yumauri/fbrcm/tui/messages"
)

type projectState struct {
	project          core.Project
	template         core.ManagedFeatureTemplate
	experiments      []core.ExperimentEntry
	personalizations []core.PersonalizationEntry
	rollouts         []core.RolloutEntry
	templateReady    bool
	loaded           bool
	waitingTemplate  bool
	loading          bool
	err              error
}

type nodeKind int

const (
	nodeProject nodeKind = iota
	nodeEntity
	nodeGap
)

type visibleNode struct {
	kind      nodeKind
	projectID string
	index     int
	entityID  string
}

type Model struct {
	svc  *core.Core
	kind messages.ManagedFeatureKind

	x, y, width, height int
	active              bool
	spin                spinner.Model
	filter              filterbox.Model
	projects            []projectState
	projectIndex        map[string]int
	visible             []visibleNode
	cursor              int
	offset              int
	lastClick           mouseutil.ClickTracker
	headerRightReserve  int
}

func New(svc *core.Core, kind messages.ManagedFeatureKind) Model {
	return Model{
		svc:          svc,
		kind:         kind,
		spin:         spinner.New(spinner.WithSpinner(spinner.Line)),
		filter:       filterbox.New(),
		projectIndex: make(map[string]int),
	}
}

func (m Model) Init() tea.Cmd { return m.spin.Tick }

func (m Model) SetBounds(x, y, width, height int) Model {
	m.x, m.y, m.width, m.height = x, y, width, height
	m.ensureCursorVisible()
	return m
}

// SetHeaderRightReserve keeps workspace titles clear of app-level overlays.
func (m Model) SetHeaderRightReserve(width int) Model {
	m.headerRightReserve = max(width, 0)
	return m
}

func (m Model) SetActive(active bool) Model {
	m.active = active
	if !active {
		m.filter.Blur()
	}
	return m
}

func (m Model) filterEnabled() bool {
	return m.kind == messages.ManagedFeatureExperiment
}

// Activate focuses the panel and refreshes only the currently selected client
// projects. Inactive managed-feature panels retain selection state without
// issuing Firebase requests.
func (m Model) Activate() (Model, tea.Cmd) {
	m.active = true
	indices := make([]int, len(m.projects))
	for index := range m.projects {
		indices[index] = index
	}
	return m.startProjectLoads(indices, false)
}

func (m Model) HasProject(projectID string) bool {
	_, ok := m.projectIndex[projectID]
	return ok
}

func (m Model) HasEntity(projectID, entityID string) bool {
	index, ok := m.projectIndex[projectID]
	if !ok {
		return false
	}
	for entityIndex := range m.entityCount(m.projects[index]) {
		if m.entityID(m.projects[index], entityIndex) == entityID {
			return true
		}
	}
	return false
}

// CurrentProject returns the project represented by the selected row.
func (m Model) CurrentProject() (core.Project, bool) {
	if m.cursor < 0 || m.cursor >= len(m.visible) {
		return core.Project{}, false
	}
	index, ok := m.projectIndex[m.visible[m.cursor].projectID]
	if !ok {
		return core.Project{}, false
	}
	return m.projects[index].project, true
}

func (m Model) HasProjects() bool {
	return len(m.projects) > 0
}

func (m Model) HasCurrentEntity() bool {
	_, ok := m.currentData()
	return ok
}
