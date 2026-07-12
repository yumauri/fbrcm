package parameters

import (
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/rc/diff"
	"github.com/yumauri/fbrcm/tui/components/filterbox"
	"github.com/yumauri/fbrcm/tui/styles"
)

const (
	panelTitleKey     = "²"
	panelTitleLabel   = "Parameters"
	historyTitleKey   = "⁹"
	historyTitleLabel = "History"
)

var (
	projectMetaStyle        = styles.PanelMuted
	parameterStyle          = styles.PanelBody.Foreground(styles.PaletteBlueBright)
	parameterValueStyle     = styles.PanelMuted
	parameterSeparatorStyle = styles.PanelMuted
	descriptionStyle        = styles.PanelMuted.Italic(true)
	conditionDefaultStyle   = styles.PanelMuted.Italic(true)
)

type projectState struct {
	project      core.Project
	tree         *core.ParametersTree
	source       string
	cacheSource  string
	err          error
	loading      bool
	verifying    bool
	hasDraft     bool
	staleDraft   bool
	cacheVersion string
	draftVersion string
}

type historyState struct {
	previous, current                   *core.ParametersTree
	merged                              *core.ParametersTree
	previousParams, currentParams       map[string]*core.ParametersEntry
	mergedParams                        map[string]*core.ParametersEntry
	previousValues, currentValues       map[string]map[string]*core.ParametersValue
	mergedValues                        map[string]map[string]*core.ParametersValue
	paramKinds                          map[string]diff.ChangeKind
	valueKinds                          map[string]map[string]diff.ChangeKind
	previousVersion, currentVersion     string
	previousPublished, currentPublished string
	err                                 error
	loading                             bool
	versions                            []core.RemoteConfigVersionEntry
	pairs                               map[string]historyPairData
	counts                              historyChangeCounts
}

type historyChangeCounts struct {
	added, removed, changed int
}

type historyPairData struct {
	previous, current                   *core.ParametersTree
	previousVersion, currentVersion     string
	previousPublished, currentPublished string
}

type historyPairSelection struct {
	previous, current string
}

type historyVersionPicker struct {
	projectID   string
	left        bool
	leftCursor  int
	rightCursor int
}

type projectLayout struct {
	metadataWidth int
}

type parameterRenderMode int

const (
	parameterRenderModeRegular parameterRenderMode = iota
	parameterRenderModeNarrow
)

type parameterRenderLayout struct {
	mode       parameterRenderMode
	paramStart int
	nameWidth  int
	valueStart int
	valueWidth int
}

type visibleNodeKind int

const (
	nodeProject visibleNodeKind = iota
	nodeGroup
	nodeParameter
	nodeValue
)

type visibleNode struct {
	kind      visibleNodeKind
	projectID string
	groupKey  string
	paramKey  string
	valueIdx  int
	label     string
	summary   string
	expanded  bool
	transient bool
}

type Model struct {
	svc *core.Core

	width   int
	height  int
	x       int
	y       int
	active  bool
	history bool
	spin    spinner.Model
	filter  filterbox.Model

	projects           []projectState
	parameterNameWidth int
	histories          map[string]historyState
	projectIndex       map[string]int
	groupExpanded      map[string]bool
	paramExpanded      map[string]bool
	visible            []visibleNode
	visibleParamCount  int
	lineIndexByNode    []int
	projectNodeFor     []int
	groupNodeFor       []int
	totalLines         int
	cursor             int
	offset             int
	transientDup       *transientDuplicate
	transientNew       *transientNewParameter
	versionPicker      *historyVersionPicker
	nextHistoryPairs   map[string]historyPairSelection
	historyChangesOnly bool
	historyViews       map[bool]selectionSnapshot
}

type transientDuplicate struct {
	projectID     string
	groupKey      string
	afterParamKey string
	label         string
}

type transientNewParameter struct {
	projectID     string
	groupKey      string
	afterParamKey string
	label         string
}

type selectionSnapshot struct {
	projectID  string
	groupKey   string
	paramKey   string
	valueIdx   int
	valueLabel string
	kind       visibleNodeKind
	screenLine int
}

func New(svc *core.Core) Model {
	return Model{
		svc:              svc,
		projectIndex:     make(map[string]int),
		histories:        make(map[string]historyState),
		groupExpanded:    make(map[string]bool),
		paramExpanded:    make(map[string]bool),
		nextHistoryPairs: make(map[string]historyPairSelection),
		historyViews:     make(map[bool]selectionSnapshot),
		filter:           filterbox.New(),
		spin: spinner.New(
			spinner.WithSpinner(spinner.Line),
		),
	}
}

func (m Model) SetHistory(history bool) Model {
	if m.history == history {
		return m
	}
	snapshot := m.captureSelectionSnapshot(true, false)
	fallbackCursor := m.cursor
	m.history = history
	if !history {
		m.versionPicker = nil
	}
	m.syncVisible()
	if cursor := m.findExactSelectionSnapshotNode(snapshot); cursor >= 0 {
		m.cursor = cursor
	} else if len(m.visible) > 0 {
		m.cursor = min(max(fallbackCursor, 0), len(m.visible)-1)
	}
	m.restoreSelectionScreenLine(snapshot.screenLine)
	return m
}

func (m Model) LoadHistory() (Model, tea.Cmd) {
	cmds := make([]tea.Cmd, 0, len(m.projects))
	for _, project := range m.projects {
		if project.loading || project.verifying {
			continue
		}
		state, ok := m.histories[project.project.ProjectID]
		if ok && (state.loading || state.current != nil || state.err != nil) {
			continue
		}
		state.loading = true
		m.histories[project.project.ProjectID] = state
		preferred, hasPreferred := m.nextHistoryPairs[project.project.ProjectID]
		delete(m.nextHistoryPairs, project.project.ProjectID)
		cmds = append(cmds, m.loadHistoryCmd(project.project, preferred, hasPreferred))
	}
	return m, tea.Batch(cmds...)
}

func (m Model) PreferNextHistoryPair(projectID, previousVersion, currentVersion string) Model {
	if m.nextHistoryPairs == nil {
		m.nextHistoryPairs = make(map[string]historyPairSelection)
	}
	m.nextHistoryPairs[projectID] = historyPairSelection{previous: previousVersion, current: currentVersion}
	return m
}

func (m Model) Init() tea.Cmd {
	return m.spin.Tick
}

func (m Model) SetBounds(x, y, width, height int) Model {
	if m.x == x && m.y == y && m.width == width && m.height == height {
		return m
	}
	m.x = x
	m.y = y
	m.width = width
	m.height = height
	m.syncVisible()
	return m
}

func (m Model) SetActive(active bool) Model {
	m.active = active
	if !active {
		m.filter.Blur()
	}
	return m
}
