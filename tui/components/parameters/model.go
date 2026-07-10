package parameters

import (
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/tui/components/filterbox"
	"github.com/yumauri/fbrcm/tui/styles"
)

const panelTitle = "[2] Parameters"

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

	width  int
	height int
	x      int
	y      int
	active bool
	spin   spinner.Model
	filter filterbox.Model

	projects        []projectState
	projectIndex    map[string]int
	groupExpanded   map[string]bool
	paramExpanded   map[string]bool
	visible         []visibleNode
	lineIndexByNode []int
	totalLines      int
	cursor          int
	offset          int
	transientDup    *transientDuplicate
	transientNew    *transientNewParameter
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
	kind       visibleNodeKind
	screenLine int
}

func New(svc *core.Core) Model {
	return Model{
		svc:           svc,
		projectIndex:  make(map[string]int),
		groupExpanded: make(map[string]bool),
		paramExpanded: make(map[string]bool),
		filter:        filterbox.New(),
		spin: spinner.New(
			spinner.WithSpinner(spinner.Line),
		),
	}
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
