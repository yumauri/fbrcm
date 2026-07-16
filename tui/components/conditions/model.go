package conditions

import (
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/tui/components/filterbox"
)

type projectState struct {
	project      core.Project
	tree         *core.ConditionsTree
	source       string
	cacheSource  string
	cacheVersion string
	draftVersion string
	hasDraft     bool
	staleDraft   bool
	err          error
	loading      bool
}

type nodeKind int

const (
	nodeProject nodeKind = iota
	nodeCondition
	nodeGap
)

type visibleNode struct {
	kind           nodeKind
	projectID      string
	conditionIndex int
	conditionName  string
}

type conditionMoveState struct {
	projectID     string
	conditionName string
	original      []core.ConditionEntry
}

type EditAnchor struct {
	Project   core.Project
	Condition core.ConditionEntry
	X         int
	Y         int
	Width     int
	MaxWidth  int
}

// NameOverlayPosition returns the overlay origin whose bordered content starts
// at the rendered condition name.
func (a EditAnchor) NameOverlayPosition() (int, int) {
	return max(a.X-2, 0), max(a.Y-1, 0)
}

type Model struct {
	svc *core.Core

	x, y, width, height int
	active              bool
	spin                spinner.Model
	filter              filterbox.Model
	projects            []projectState
	projectIndex        map[string]int
	visible             []visibleNode
	cursor              int
	offset              int
	move                *conditionMoveState
}

func New(svc *core.Core) Model {
	return Model{
		svc:          svc,
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

func (m Model) SetActive(active bool) Model {
	m.active = active
	if !active {
		m.filter.Blur()
	}
	return m
}

func (m Model) HasProject(projectID string) bool {
	_, ok := m.projectIndex[projectID]
	return ok
}

func (m Model) LongestConditionNameWidth() int {
	longest := 0
	for _, project := range m.projects {
		if project.tree == nil {
			continue
		}
		for _, condition := range project.tree.Conditions {
			longest = max(longest, len([]rune(condition.Name)))
		}
	}
	return longest
}
