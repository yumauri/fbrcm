package parameters

import (
	"context"
	"fmt"
	"math/big"
	"regexp"
	"slices"
	"sort"
	"strings"
	"time"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/atotto/clipboard"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/filter"
	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/tui/components/filterbox"
	"github.com/yumauri/fbrcm/tui/messages"
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

var deprecatedDescriptionPattern = regexp.MustCompile(`(?i)deprecat|obsolete|sunset|retired?|no longer|use .+ instead|replaced?|superseded?|removed?`)

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

type RenameAnchor struct {
	Project  core.Project
	IsGroup  bool
	GroupKey string
	ParamKey string
	Label    string
	X        int
	Y        int
	Width    int
	MaxWidth int
}

type MoveAnchor struct {
	Project  core.Project
	IsGroup  bool
	GroupKey string
	ParamKey string
	Label    string
	X        int
	Y        int
	Options  []MoveOption
}

type MoveOption struct {
	Key   string
	Label string
}

// ConditionalValueAnchor holds selected conditional value deletion target.
type ConditionalValueAnchor struct {
	Project    core.Project
	GroupKey   string
	ParamKey   string
	ValueLabel string
}

type BoolValueAnchor struct {
	Project      core.Project
	GroupKey     string
	ParamKey     string
	ValueLabel   string
	Value        bool
	CurrentValue string
	X            int
	Y            int
}

type NumberValueAnchor struct {
	Project      core.Project
	GroupKey     string
	ParamKey     string
	ValueLabel   string
	CurrentValue string
	X            int
	Y            int
	Width        int
	MaxWidth     int
}

type StringValueAnchor struct {
	Project      core.Project
	GroupKey     string
	ParamKey     string
	ValueLabel   string
	CurrentValue string
	X            int
	Y            int
	Width        int
	MaxWidth     int
	FullWidth    bool
	Expanded     bool
}

type JSONValueAnchor struct {
	Project      core.Project
	GroupKey     string
	ParamKey     string
	ValueLabel   string
	CurrentValue string
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

// Init initializes init for Model and returns the resulting state or error.
func (m Model) Init() tea.Cmd {
	return m.spin.Tick
}

// SetSize sets size for Model and returns the resulting state or error.
func (m Model) SetSize(width, height int) Model {
	if m.width == width && m.height == height {
		return m
	}
	m.width = width
	m.height = height
	m.syncVisible()
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
	m.syncVisible()
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

func (m *Model) syncVisible() {
	m.visible = m.buildVisible()
	if len(m.visible) == 0 {
		m.lineIndexByNode = nil
		m.cursor = 0
		m.offset = 0
		m.totalLines = 0
		return
	}

	m.cursor = max(0, min(m.cursor, len(m.visible)-1))
	m.recomputeLineLayout()
	m.ensureCursorVisible()
}

func (m *Model) recomputeLineLayout() {
	m.lineIndexByNode = make([]int, len(m.visible))
	line := 0
	for i := range m.visible {
		m.lineIndexByNode[i] = line
		line += m.nodeBlockLineCount(i)
	}
	m.totalLines = line
}

func (m Model) buildVisible() []visibleNode {
	nodes := make([]visibleNode, 0)
	query := m.filter.Value()
	filtering := query != ""
	for _, project := range m.projects {
		nodes = append(nodes, visibleNode{
			kind:      nodeProject,
			projectID: project.project.ProjectID,
			label:     displayProject(project),
			expanded:  true,
		})

		if project.loading {
			nodes = append(nodes, visibleNode{
				kind:      nodeValue,
				projectID: project.project.ProjectID,
				label:     "Loading parameters...",
			})
			continue
		}
		if project.err != nil && project.tree == nil {
			nodes = append(nodes, visibleNode{
				kind:      nodeValue,
				projectID: project.project.ProjectID,
				label:     fmt.Sprintf("Load failed: %v", project.err),
			})
			continue
		}
		if project.tree == nil || len(project.tree.Groups) == 0 {
			if created := m.transientNew; created != nil && created.projectID == project.project.ProjectID {
				nodes = appendTransientNewRootGroup(nodes, project.project.ProjectID, created)
				continue
			}
			nodes = append(nodes, visibleNode{
				kind:      nodeValue,
				projectID: project.project.ProjectID,
				label:     "No parameters",
			})
			continue
		}

		transientRootShown := false
		for _, group := range project.tree.Groups {
			if created := m.transientNew; created != nil &&
				created.projectID == project.project.ProjectID &&
				core.NormalizeRemoteConfigGroupKey(created.groupKey) == "" &&
				core.NormalizeRemoteConfigGroupKey(group.Key) == "" {
				transientRootShown = true
			}
			matchedParams := group.Parameters
			if filtering {
				matchedParams = matchedParameters(group.Parameters, query, m.filter.Mode())
				created := m.transientNew
				hasTransientNew := created != nil &&
					created.projectID == project.project.ProjectID &&
					core.NormalizeRemoteConfigGroupKey(created.groupKey) == core.NormalizeRemoteConfigGroupKey(group.Key)
				if len(matchedParams) == 0 && !hasTransientNew {
					continue
				}
			}
			groupExpanded := m.groupExpanded[m.groupKey(project.project.ProjectID, group.Key)]
			nodes = append(nodes, visibleNode{
				kind:      nodeGroup,
				projectID: project.project.ProjectID,
				groupKey:  group.Key,
				label:     group.Label,
				summary:   fmt.Sprintf("%d", len(matchedParams)),
				expanded:  groupExpanded,
			})
			if !groupExpanded {
				continue
			}

			for _, param := range matchedParams {
				paramExpanded := m.paramExpanded[m.paramKey(project.project.ProjectID, group.Key, param.Key)]
				nodes = append(nodes, visibleNode{
					kind:      nodeParameter,
					projectID: project.project.ProjectID,
					groupKey:  group.Key,
					paramKey:  param.Key,
					label:     param.Key,
					summary:   param.Summary,
					expanded:  paramExpanded,
				})
				if dup := m.transientDup; dup != nil &&
					dup.projectID == project.project.ProjectID &&
					dup.groupKey == group.Key &&
					dup.afterParamKey == param.Key &&
					(!filtering || matchedDuplicate(dup.label, query, m.filter.Mode())) {
					nodes = append(nodes, visibleNode{
						kind:      nodeParameter,
						projectID: project.project.ProjectID,
						groupKey:  group.Key,
						paramKey:  param.Key,
						label:     dup.label,
						summary:   param.Summary,
						expanded:  false,
						transient: true,
					})
				}
				if created := m.transientNew; created != nil &&
					created.projectID == project.project.ProjectID &&
					core.NormalizeRemoteConfigGroupKey(created.groupKey) == core.NormalizeRemoteConfigGroupKey(group.Key) &&
					created.afterParamKey == param.Key {
					nodes = append(nodes, visibleNode{
						kind:      nodeParameter,
						projectID: project.project.ProjectID,
						groupKey:  group.Key,
						paramKey:  "",
						label:     created.label,
						summary:   "new parameter",
						expanded:  false,
						transient: true,
					})
				}
				if !paramExpanded {
					continue
				}

				for i, value := range param.Values {
					nodes = append(nodes, visibleNode{
						kind:      nodeValue,
						projectID: project.project.ProjectID,
						groupKey:  group.Key,
						paramKey:  param.Key,
						valueIdx:  i,
						label:     value.Label,
						summary:   value.Value,
					})
				}
			}
			if created := m.transientNew; created != nil &&
				created.projectID == project.project.ProjectID &&
				core.NormalizeRemoteConfigGroupKey(created.groupKey) == core.NormalizeRemoteConfigGroupKey(group.Key) &&
				created.afterParamKey == "" {
				nodes = append(nodes, visibleNode{
					kind:      nodeParameter,
					projectID: project.project.ProjectID,
					groupKey:  group.Key,
					paramKey:  "",
					label:     created.label,
					summary:   "new parameter",
					expanded:  false,
					transient: true,
				})
			}
		}
		if created := m.transientNew; created != nil &&
			created.projectID == project.project.ProjectID &&
			core.NormalizeRemoteConfigGroupKey(created.groupKey) == "" &&
			!transientRootShown {
			nodes = appendTransientNewRootGroup(nodes, project.project.ProjectID, created)
		}
	}

	return nodes
}

func appendTransientNewRootGroup(nodes []visibleNode, projectID string, created *transientNewParameter) []visibleNode {
	nodes = append(nodes, visibleNode{
		kind:      nodeGroup,
		projectID: projectID,
		groupKey:  "__default__",
		label:     "(root)",
		summary:   "0",
		expanded:  true,
	})
	return append(nodes, visibleNode{
		kind:      nodeParameter,
		projectID: projectID,
		groupKey:  "__default__",
		paramKey:  "",
		label:     created.label,
		summary:   "new parameter",
		expanded:  false,
		transient: true,
	})
}

// matchedParameters matches matched parameters and returns the resulting value or error.
func matchedParameters(params []core.ParametersEntry, query string, mode filter.Mode) []core.ParametersEntry {
	if query == "" {
		return params
	}
	out := make([]core.ParametersEntry, 0, len(params))
	for _, param := range params {
		if ok, _ := filter.Match(param.Key, query, mode); ok {
			out = append(out, param)
		}
	}
	return out
}

// matchedDuplicate matches matched duplicate and returns the resulting value or error.
func matchedDuplicate(label, query string, mode filter.Mode) bool {
	if query == "" {
		return true
	}
	ok, _ := filter.Match(label, query, mode)
	return ok
}

func (m Model) loadParametersCmd(project core.Project) tea.Cmd {
	return func() tea.Msg {
		cache, state, err := m.svc.InspectParametersCache(project.ProjectID)
		if err != nil {
			return messages.ParametersLoadedMsg{Project: project, Err: err}
		}

		switch state {
		case core.ParametersCacheFresh:
			tree, hasDraft, err := m.svc.BuildDraftAwareParametersTree(project.ProjectID, cache)
			source := "cache"
			if hasDraft {
				source = "draft"
			}
			cacheVersion := remoteConfigVersion(cache.RemoteConfig)
			draftVersion := ""
			staleDraft := false
			if hasDraft && tree != nil {
				draftVersion = tree.Version
				staleDraft = versionLess(draftVersion, cacheVersion)
			}
			return messages.ParametersLoadedMsg{Project: project, Tree: tree, Source: source, CacheSource: "cache", CacheVersion: cacheVersion, DraftVersion: draftVersion, Err: err, HasDraft: hasDraft, StaleDraft: staleDraft}
		case core.ParametersCacheStale:
			tree, hasDraft, err := m.svc.BuildDraftAwareParametersTree(project.ProjectID, cache)
			source := "cache-stale"
			if hasDraft {
				source = "draft"
			}
			cacheVersion := remoteConfigVersion(cache.RemoteConfig)
			draftVersion := ""
			staleDraft := false
			if hasDraft && tree != nil {
				draftVersion = tree.Version
				staleDraft = versionLess(draftVersion, cacheVersion)
			}
			return messages.ParametersLoadedMsg{Project: project, Tree: tree, Source: source, CacheSource: "cache-stale", CacheVersion: cacheVersion, DraftVersion: draftVersion, Err: err, Revalidate: true, HasDraft: hasDraft, StaleDraft: staleDraft}
		default:
			cache, source, err := m.svc.GetParameters(context.Background(), project.ProjectID, false)
			if err != nil {
				return messages.ParametersLoadedMsg{Project: project, Err: err}
			}
			tree, hasDraft, err := m.svc.BuildDraftAwareParametersTree(project.ProjectID, cache)
			cacheSource := source
			if hasDraft {
				source = "draft"
			}
			cacheVersion := remoteConfigVersion(cache.RemoteConfig)
			draftVersion := ""
			staleDraft := false
			if hasDraft && tree != nil {
				draftVersion = tree.Version
				staleDraft = versionLess(draftVersion, cacheVersion)
			}
			return messages.ParametersLoadedMsg{Project: project, Tree: tree, Source: source, CacheSource: cacheSource, CacheVersion: cacheVersion, DraftVersion: draftVersion, Err: err, HasDraft: hasDraft, StaleDraft: staleDraft}
		}
	}
}

func (m Model) revalidateParametersCmd(project core.Project) tea.Cmd {
	return func() tea.Msg {
		cache, tree, source, hasDraft, staleDraft, err := m.svc.RefreshDraftAwareParameters(context.Background(), project.ProjectID)
		if err != nil {
			return messages.ParametersLoadedMsg{Project: project, Err: err}
		}
		_ = cache
		cacheSource := source
		if source == "draft" || source == "draft-stale" {
			cacheSource = "firebase"
		}
		cacheVersion := remoteConfigVersion(cache.RemoteConfig)
		draftVersion := ""
		if hasDraft && tree != nil {
			draftVersion = tree.Version
			staleDraft = staleDraft || versionLess(draftVersion, cacheVersion)
		}
		return messages.ParametersLoadedMsg{Project: project, Tree: tree, Source: source, CacheSource: cacheSource, CacheVersion: cacheVersion, DraftVersion: draftVersion, Err: err, HasDraft: hasDraft, StaleDraft: staleDraft}
	}
}

func (m Model) forceParametersCmd(project core.Project) tea.Cmd {
	return func() tea.Msg {
		cache, tree, source, hasDraft, staleDraft, err := m.svc.RefreshDraftAwareParameters(context.Background(), project.ProjectID)
		if err != nil {
			return messages.ParametersLoadedMsg{Project: project, Err: err}
		}
		_ = cache
		cacheSource := source
		if source == "draft" || source == "draft-stale" {
			cacheSource = "firebase"
		}
		cacheVersion := remoteConfigVersion(cache.RemoteConfig)
		draftVersion := ""
		if hasDraft && tree != nil {
			draftVersion = tree.Version
			staleDraft = staleDraft || versionLess(draftVersion, cacheVersion)
		}
		return messages.ParametersLoadedMsg{Project: project, Tree: tree, Source: source, CacheSource: cacheSource, CacheVersion: cacheVersion, DraftVersion: draftVersion, Err: err, HasDraft: hasDraft, StaleDraft: staleDraft}
	}
}

// setProjects sets set projects for Model and returns the resulting state or error.
func (m *Model) setProjects(projects []core.Project) tea.Cmd {
	sort.Slice(projects, func(i, j int) bool {
		leftName := strings.ToLower(strings.TrimSpace(projects[i].Name))
		rightName := strings.ToLower(strings.TrimSpace(projects[j].Name))
		if leftName == "" {
			leftName = strings.ToLower(projects[i].ProjectID)
		}
		if rightName == "" {
			rightName = strings.ToLower(projects[j].ProjectID)
		}
		if leftName == rightName {
			return strings.ToLower(projects[i].ProjectID) < strings.ToLower(projects[j].ProjectID)
		}
		return leftName < rightName
	})

	nextProjects := make([]projectState, 0, len(projects))
	nextIndex := make(map[string]int, len(projects))
	cmds := make([]tea.Cmd, 0)

	for _, project := range projects {
		if idx, ok := m.projectIndex[project.ProjectID]; ok {
			state := m.projects[idx]
			state.project = project
			nextIndex[project.ProjectID] = len(nextProjects)
			nextProjects = append(nextProjects, state)
			continue
		}

		nextIndex[project.ProjectID] = len(nextProjects)
		nextProjects = append(nextProjects, projectState{
			project: project,
			loading: true,
		})
		cmds = append(cmds, m.loadParametersCmd(project))
	}

	m.projects = nextProjects
	m.projectIndex = nextIndex
	m.syncVisible()

	return tea.Batch(cmds...)
}

func (m *Model) updateProject(msg messages.ParametersLoadedMsg) tea.Cmd {
	idx, ok := m.projectIndex[msg.Project.ProjectID]
	if !ok {
		return nil
	}

	state := m.projects[idx]
	if msg.Err != nil {
		if state.tree == nil {
			state.tree = nil
			state.source = msg.Source
		}
		state.err = msg.Err
	} else {
		state.tree = msg.Tree
		state.source = msg.Source
		if msg.CacheSource != "" {
			state.cacheSource = msg.CacheSource
		} else {
			state.cacheSource = msg.Source
		}
		state.err = nil
	}
	state.loading = false
	state.verifying = false
	state.hasDraft = msg.HasDraft
	state.staleDraft = msg.StaleDraft
	if msg.CacheVersion != "" {
		state.cacheVersion = msg.CacheVersion
	} else if msg.Tree != nil && !msg.HasDraft {
		state.cacheVersion = msg.Tree.Version
	}
	if msg.DraftVersion != "" {
		state.draftVersion = msg.DraftVersion
	} else if msg.HasDraft && msg.Tree != nil {
		state.draftVersion = msg.Tree.Version
	} else if !msg.HasDraft {
		state.draftVersion = ""
	}
	m.projects[idx] = state

	cmds := make([]tea.Cmd, 0, 1)
	if msg.Tree != nil {
		for _, group := range msg.Tree.Groups {
			groupKey := m.groupKey(msg.Project.ProjectID, group.Key)
			if _, ok := m.groupExpanded[groupKey]; !ok {
				m.groupExpanded[groupKey] = true
			}
		}
	}
	if msg.Revalidate && msg.Err == nil {
		state.verifying = true
		m.projects[idx] = state
		cmds = append(cmds, m.revalidateParametersCmd(msg.Project))
	}

	m.syncVisible()
	if msg.SelectParamKey != "" {
		m.selectParameter(msg.Project.ProjectID, msg.SelectGroupKey, msg.SelectParamKey)
	}
	return tea.Batch(cmds...)
}

func (m *Model) ensureCursorVisible() {
	if len(m.visible) == 0 {
		m.cursor = 0
		m.offset = 0
		return
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor >= len(m.visible) {
		m.cursor = len(m.visible) - 1
	}

	blockStart, blockEnd := m.selectionBlockRange(m.cursor)
	rowHeight := blockEnd - blockStart
	visibleLines := m.bodyVisibleLinesForOffset(m.offset)
	bodyStart := m.bodyStartForOffset(m.offset)

	desiredBodyStart := bodyStart
	if rowHeight >= visibleLines {
		desiredBodyStart = blockStart
	} else {
		if blockStart < bodyStart {
			desiredBodyStart = blockStart
		}
		if blockEnd > bodyStart+visibleLines {
			desiredBodyStart = blockEnd - visibleLines
		}
	}

	m.offset = m.offsetForBodyStart(desiredBodyStart)

	maxOffset := max(m.totalLines-visibleLines, 0)
	if m.offset > maxOffset {
		m.offset = maxOffset
	}
}

// selectionBlockRange selects selection block range for Model and returns the resulting state or error.
func (m Model) selectionBlockRange(index int) (int, int) {
	if index < 0 || index >= len(m.visible) {
		return 0, 0
	}

	start := m.lineIndexByNode[index]
	end := start + m.nodeBlockLineCount(index)
	node := m.visible[index]
	if node.kind != nodeParameter || !node.expanded {
		return start, end
	}

	for i := index + 1; i < len(m.visible); i++ {
		next := m.visible[i]
		if next.kind != nodeValue ||
			next.projectID != node.projectID ||
			next.groupKey != node.groupKey ||
			next.paramKey != node.paramKey {
			break
		}
		end = m.lineIndexByNode[i] + m.nodeBlockLineCount(i)
	}

	return start, end
}

// moveCursor moves move cursor for Model and returns the resulting state or error.
func (m *Model) moveCursor(delta int) {
	if len(m.visible) == 0 {
		return
	}
	m.cursor = max(0, min(m.cursor+delta, len(m.visible)-1))
	m.ensureCursorVisible()
}

// moveToNextGroup moves move to next group for Model and returns the resulting state or error.
func (m *Model) moveToNextGroup() {
	if len(m.visible) == 0 {
		return
	}
	current := m.cursor
	for i := current + 1; i < len(m.visible); i++ {
		if m.visible[i].kind == nodeGroup {
			m.cursor = i
			m.offset = m.lineIndexByNode[m.cursor]
			maxOffset := max(m.totalLines-m.bodyVisibleLinesForOffset(m.offset), 0)
			if m.offset > maxOffset {
				m.offset = maxOffset
			}
			return
		}
	}
}

// moveToPrevGroup moves move to prev group for Model and returns the resulting state or error.
func (m *Model) moveToPrevGroup() {
	if len(m.visible) == 0 {
		return
	}
	current := m.cursor
	for i := current - 1; i >= 0; i-- {
		if m.visible[i].kind == nodeGroup {
			m.cursor = i
			m.ensureCursorVisible()
			return
		}
	}
}

func (m *Model) collapseCurrent() {
	if len(m.visible) == 0 {
		return
	}

	node := m.visible[m.cursor]
	switch node.kind {
	case nodeParameter:
		key := m.paramKey(node.projectID, node.groupKey, node.paramKey)
		if m.paramExpanded[key] {
			m.paramExpanded[key] = false
		} else {
			m.focusParentGroup(node)
			return
		}
	case nodeGroup:
		key := m.groupKey(node.projectID, node.groupKey)
		if m.groupExpanded[key] {
			m.groupExpanded[key] = false
		}
	case nodeValue:
		m.focusParentParameter(node)
		return
	default:
		return
	}

	m.syncVisible()
}

func (m *Model) expandCurrent() {
	if len(m.visible) == 0 {
		return
	}

	node := m.visible[m.cursor]
	switch node.kind {
	case nodeGroup:
		m.groupExpanded[m.groupKey(node.projectID, node.groupKey)] = true
	case nodeParameter:
		m.paramExpanded[m.paramKey(node.projectID, node.groupKey, node.paramKey)] = true
	default:
		return
	}

	m.syncVisible()
}

// toggleCurrentParameter toggles toggle current parameter for Model and returns the resulting state or error.
func (m *Model) toggleCurrentParameter() {
	if len(m.visible) == 0 {
		return
	}

	node := m.visible[m.cursor]
	if node.kind != nodeParameter {
		return
	}

	key := m.paramKey(node.projectID, node.groupKey, node.paramKey)
	m.paramExpanded[key] = !m.paramExpanded[key]
	m.syncVisible()
}

// focusCurrentParameterDefaultValue focuses focus current parameter default value for Model and returns the resulting state or error.
func (m *Model) focusCurrentParameterDefaultValue() bool {
	if len(m.visible) == 0 || m.cursor < 0 || m.cursor >= len(m.visible) {
		return false
	}

	node := m.visible[m.cursor]
	if node.kind != nodeParameter || node.transient {
		return false
	}

	key := m.paramKey(node.projectID, node.groupKey, node.paramKey)
	if !m.paramExpanded[key] {
		m.paramExpanded[key] = true
		m.syncVisible()
	}

	firstValueIdx := -1
	for i, candidate := range m.visible {
		if candidate.projectID != node.projectID || candidate.groupKey != node.groupKey || candidate.paramKey != node.paramKey {
			continue
		}
		if candidate.kind != nodeValue {
			continue
		}
		if firstValueIdx < 0 {
			firstValueIdx = i
		}
		if candidate.label == "default" {
			m.cursor = i
			m.ensureCursorVisible()
			return true
		}
	}

	if firstValueIdx >= 0 {
		m.cursor = firstValueIdx
		m.ensureCursorVisible()
		return true
	}

	return false
}

// focusParentGroup focuses focus parent group for Model and returns the resulting state or error.
func (m *Model) focusParentGroup(node visibleNode) {
	for i, candidate := range m.visible {
		if candidate.kind == nodeGroup && candidate.projectID == node.projectID && candidate.groupKey == node.groupKey {
			m.cursor = i
			m.ensureCursorVisible()
			return
		}
	}
}

// focusParentParameter focuses focus parent parameter for Model and returns the resulting state or error.
func (m *Model) focusParentParameter(node visibleNode) {
	for i, candidate := range m.visible {
		if candidate.kind == nodeParameter && candidate.projectID == node.projectID && candidate.groupKey == node.groupKey && candidate.paramKey == node.paramKey {
			m.cursor = i
			m.ensureCursorVisible()
			return
		}
	}
}

func (m Model) currentProjectID() string {
	if m.cursor < 0 || m.cursor >= len(m.visible) {
		return ""
	}
	return m.visible[m.cursor].projectID
}

func (m Model) currentProject() (core.Project, bool) {
	project := m.projectByID(m.currentProjectID())
	if project == nil {
		return core.Project{}, false
	}
	return project.project, true
}

// moveToCurrentProjectHeader moves move to current project header for Model and returns the resulting state or error.
func (m *Model) moveToCurrentProjectHeader() {
	if len(m.visible) == 0 || m.cursor < 0 || m.cursor >= len(m.visible) {
		return
	}
	for i := m.cursor; i >= 0; i-- {
		if m.visible[i].kind == nodeProject &&
			m.visible[i].projectID == m.visible[m.cursor].projectID {
			m.cursor = i
			m.ensureCursorVisible()
			return
		}
	}
}

// moveToLastParameterInCurrentProject moves move to last parameter in current project for Model and returns the resulting state or error.
func (m *Model) moveToLastParameterInCurrentProject() {
	projectID := m.currentProjectID()
	if projectID == "" {
		return
	}
	for i, node := range slices.Backward(m.visible) {
		if node.projectID != projectID {
			continue
		}
		if node.kind == nodeParameter {
			m.cursor = i
			m.ensureCursorVisible()
			return
		}
	}
}

// setAllParametersExpanded sets set all parameters expanded for Model and returns the resulting state or error.
func (m *Model) setAllParametersExpanded(expanded bool) {
	snapshot := m.captureSelectionSnapshot(expanded, false)
	for _, project := range m.projects {
		if project.tree == nil {
			continue
		}
		for _, group := range project.tree.Groups {
			for _, param := range group.Parameters {
				m.paramExpanded[m.paramKey(project.project.ProjectID, group.Key, param.Key)] = expanded
			}
		}
	}
	m.syncVisible()
	m.restoreSelectionSnapshot(snapshot)
}

// setAllGroupsExpanded sets set all groups expanded for Model and returns the resulting state or error.
func (m *Model) setAllGroupsExpanded(expanded bool) {
	snapshot := m.captureSelectionSnapshot(expanded, true)
	for _, project := range m.projects {
		if project.tree == nil {
			continue
		}
		for _, group := range project.tree.Groups {
			m.groupExpanded[m.groupKey(project.project.ProjectID, group.Key)] = expanded
		}
	}
	m.syncVisible()
	m.restoreSelectionSnapshot(snapshot)
}

func (m *Model) markProjectRefreshing(projectID string) {
	idx, ok := m.projectIndex[projectID]
	if !ok {
		return
	}
	state := m.projects[idx]
	state.verifying = true
	state.err = nil
	m.projects[idx] = state
}

func (m *Model) revalidateCurrentProjectCmd() tea.Cmd {
	project, ok := m.currentProject()
	if !ok {
		return nil
	}
	m.markProjectRefreshing(project.ProjectID)
	m.syncVisible()
	return tea.Batch(m.forceParametersCmd(project), m.spin.Tick)
}

func (m *Model) revalidateAllProjectsCmd() tea.Cmd {
	if len(m.projects) == 0 {
		return nil
	}
	cmds := make([]tea.Cmd, 0, len(m.projects)+1)
	for _, project := range m.projects {
		m.markProjectRefreshing(project.project.ProjectID)
		cmds = append(cmds, m.forceParametersCmd(project.project))
	}
	m.syncVisible()
	cmds = append(cmds, m.spin.Tick)
	return tea.Batch(cmds...)
}

func (m Model) copyCurrentParameterNameCmd() tea.Cmd {
	_, _, paramKey, ok := m.currentParameterRef()
	if !ok {
		return nil
	}
	return copyToClipboardCmd(paramKey)
}

func (m Model) copyCurrentParameterPathCmd() tea.Cmd {
	projectID, groupKey, paramKey, ok := m.currentParameterRef()
	if !ok {
		return nil
	}
	return copyToClipboardCmd(projectID + "/" + groupKey + "/" + paramKey)
}

func (m Model) currentParameterRef() (projectID, groupKey, paramKey string, ok bool) {
	if m.cursor < 0 || m.cursor >= len(m.visible) {
		return "", "", "", false
	}

	node := m.visible[m.cursor]
	if node.transient {
		return "", "", "", false
	}
	switch node.kind {
	case nodeParameter, nodeValue:
		return node.projectID, node.groupKey, node.paramKey, true
	default:
		return "", "", "", false
	}
}

func (m Model) currentParameterViewData() (*messages.ParameterViewData, bool) {
	if m.cursor < 0 || m.cursor >= len(m.visible) {
		return nil, false
	}

	node := m.visible[m.cursor]
	if node.kind != nodeParameter && node.kind != nodeValue {
		return nil, false
	}

	project := m.projectByID(node.projectID)
	group := m.groupByKey(node.projectID, node.groupKey)
	if project == nil {
		return nil, false
	}
	groupKey := node.groupKey
	groupLabel := "(root)"
	if group != nil {
		groupKey = group.Key
		groupLabel = group.Label
	}
	groups, paramKeys := parameterViewOptions(project)
	if len(groups) == 0 {
		groups = []messages.ParameterGroupOption{{Key: "__default__", Label: "(root)"}}
	}
	if node.transient && m.transientNew != nil &&
		m.transientNew.projectID == node.projectID &&
		core.NormalizeRemoteConfigGroupKey(m.transientNew.groupKey) == core.NormalizeRemoteConfigGroupKey(node.groupKey) {
		return &messages.ParameterViewData{
			Project:       project.project,
			GroupKey:      groupKey,
			GroupLabel:    groupLabel,
			Groups:        groups,
			ParameterKeys: paramKeys,
			Parameter: core.ParametersEntry{
				Key:     "",
				Summary: "new parameter",
				Values: []core.ParametersValue{{
					Label:     "default",
					Value:     "(empty string)",
					RawValue:  "",
					ValueType: "STRING",
					Empty:     true,
					EmptyType: "STRING",
					Plain:     true,
				}},
			},
			SelectedValueIdx: -1,
		}, true
	}
	if node.transient {
		return nil, false
	}
	param := m.parameterByKey(node.projectID, node.groupKey, node.paramKey)
	if param == nil {
		return nil, false
	}

	valueIdx := -1
	if node.kind == nodeValue {
		valueIdx = node.valueIdx
	}

	data := &messages.ParameterViewData{
		Project:          project.project,
		GroupKey:         group.Key,
		GroupLabel:       group.Label,
		Groups:           groups,
		ParameterKeys:    paramKeys,
		Parameter:        *param,
		SelectedValueIdx: valueIdx,
	}
	return data, true
}

func parameterViewOptions(project *projectState) ([]messages.ParameterGroupOption, []string) {
	if project == nil || project.tree == nil {
		return nil, nil
	}
	groups := make([]messages.ParameterGroupOption, 0, len(project.tree.Groups)+1)
	seenRoot := false
	for _, group := range project.tree.Groups {
		if core.NormalizeRemoteConfigGroupKey(group.Key) == "" {
			seenRoot = true
		}
		groups = append(groups, messages.ParameterGroupOption{Key: group.Key, Label: group.Label})
	}
	if !seenRoot {
		groups = append([]messages.ParameterGroupOption{{Key: "__default__", Label: "(root)"}}, groups...)
	}
	paramKeys := make([]string, 0)
	for _, group := range project.tree.Groups {
		for _, param := range group.Parameters {
			paramKeys = append(paramKeys, param.Key)
		}
	}
	return groups, paramKeys
}

func (m Model) CurrentParameterViewData() (*messages.ParameterViewData, bool) {
	return m.currentParameterViewData()
}

// selectionChangedCmd selects selection changed cmd for Model and returns the resulting state or error.
func (m Model) selectionChangedCmd(activate bool) tea.Cmd {
	data, ok := m.currentParameterViewData()
	if !ok {
		return func() tea.Msg {
			return messages.ParameterSelectionChangedMsg{
				ResetScroll: true,
			}
		}
	}

	return func() tea.Msg {
		return messages.ParameterSelectionChangedMsg{
			Data:     data,
			Activate: activate,
		}
	}
}

func copyToClipboardCmd(text string) tea.Cmd {
	if text == "" {
		return nil
	}
	return func() tea.Msg {
		_ = clipboard.WriteAll(text)
		return nil
	}
}

func (m Model) contentHeight() int {
	return max(m.height-2-m.filter.Height(), 0)
}

func (m Model) viewportWidth() int {
	return max(m.width-2, 1)
}

func (m Model) viewportHeight() int {
	return max(m.height-2-m.filter.Height(), 1)
}

func (m Model) groupKey(projectID, groupKey string) string {
	return projectID + "::group::" + groupKey
}

func (m Model) paramKey(projectID, groupKey, paramKey string) string {
	return projectID + "::param::" + groupKey + "::" + paramKey
}

func (p projectState) cacheStateLabel() string {
	if p.tree == nil {
		return core.ParametersStatusLabel(p.cacheSource, time.Time{}, false, p.err)
	}
	return core.ParametersStatusLabel(p.cacheSource, p.tree.CachedAt, true, p.err)
}

func (m Model) anyLoading() bool {
	for _, project := range m.projects {
		if project.loading || project.verifying {
			return true
		}
	}
	return false
}

func (m Model) projectByID(projectID string) *projectState {
	idx, ok := m.projectIndex[projectID]
	if !ok || idx < 0 || idx >= len(m.projects) {
		return nil
	}
	return &m.projects[idx]
}

func (m Model) groupByKey(projectID, groupKey string) *core.ParametersGroup {
	project := m.projectByID(projectID)
	if project == nil || project.tree == nil {
		return nil
	}
	for i := range project.tree.Groups {
		if project.tree.Groups[i].Key == groupKey {
			return &project.tree.Groups[i]
		}
	}
	return nil
}

func (m Model) parameterByKey(projectID, groupKey, paramKey string) *core.ParametersEntry {
	group := m.groupByKey(projectID, groupKey)
	if group == nil {
		return nil
	}
	for i := range group.Parameters {
		if group.Parameters[i].Key == paramKey {
			return &group.Parameters[i]
		}
	}
	return nil
}

func (m Model) layoutForProject(projectID string) projectLayout {
	layout := projectLayout{}

	project := m.projectByID(projectID)
	if project == nil {
		return layout
	}

	metadata := m.projectMeta(project, false)
	layout.metadataWidth = lipgloss.Width(metadata)
	return layout
}

func (m Model) parameterRenderLayout() parameterRenderLayout {
	layout := parameterRenderLayout{
		mode:       parameterRenderModeRegular,
		paramStart: 2,
		nameWidth:  m.maxParameterNameWidth(),
	}
	layout.valueStart = layout.paramStart + layout.nameWidth + 3
	layout.valueWidth = max(m.viewportWidth()-layout.valueStart, 0)
	if layout.valueWidth < 10 {
		layout.mode = parameterRenderModeNarrow
	}
	return layout
}

func (m Model) maxParameterNameWidth() int {
	width := 0
	for _, project := range m.projects {
		if project.tree == nil {
			continue
		}
		for _, group := range project.tree.Groups {
			for _, param := range group.Parameters {
				width = max(width, lipgloss.Width(param.Key))
			}
		}
	}
	return width
}

func (m Model) LongestParameterNameWidth() int {
	return m.maxParameterNameWidth()
}

func (m Model) projectMeta(project *projectState, selected bool) string {
	if project == nil {
		return ""
	}

	badge, rest := m.projectMetaSegments(project, selected)
	switch {
	case badge != "" && rest != "":
		return badge + " " + rest
	case badge != "":
		return badge
	default:
		return rest
	}
}

func (m Model) projectMetaSegments(project *projectState, selected bool) (badge string, rest string) {
	if project == nil {
		return "", ""
	}

	if project.hasDraft {
		label := "draft"
		if project.staleDraft {
			label = "staled draft"
			if project.draftVersion != "" {
				label += " v" + project.draftVersion
			}
		}
		if selected {
			badge = lipgloss.NewStyle().Foreground(styles.PaletteError).Render(label)
		} else {
			badge = draftBadgeStyle.Render(label)
		}
	}

	parts := make([]string, 0, 3)
	version := project.displayVersion()
	if version != "" {
		parts = append(parts, "v"+version)
	}
	if project.loading || project.verifying {
		parts = append(parts, m.spin.View())
	} else if project.err != nil && project.tree != nil {
		parts = append(parts, "error")
	} else if state := project.cacheStateLabel(); state != "" {
		parts = append(parts, state)
	}
	if project.tree != nil && !project.tree.CachedAt.IsZero() {
		parts = append(parts, project.tree.CachedAt.Local().Format("2006-01-02 15:04:05"))
	}
	return badge, strings.Join(parts, " ")
}

func (p projectState) displayVersion() string {
	if p.staleDraft && p.cacheVersion != "" {
		return p.cacheVersion
	}
	if p.tree != nil && p.tree.Version != "" {
		return p.tree.Version
	}
	return p.cacheVersion
}

// filteredParameterCount filters filtered parameter count for Model and returns the resulting state or error.
func (m Model) filteredParameterCount() int {
	count := 0
	for _, node := range m.visible {
		if node.kind == nodeParameter {
			count++
		}
	}
	return count
}

func displayProject(project projectState) string {
	if strings.TrimSpace(project.project.Name) == "" {
		return project.project.ProjectID
	}
	return fmt.Sprintf("%s (%s)", project.project.Name, project.project.ProjectID)
}

func (m Model) CurrentParameterRef() (core.Project, string, string, bool) {
	projectID, groupKey, paramKey, ok := m.currentParameterRef()
	if !ok {
		return core.Project{}, "", "", false
	}
	project := m.projectByID(projectID)
	if project == nil {
		return core.Project{}, "", "", false
	}
	return project.project, groupKey, paramKey, true
}

// FocusCurrentParameterDefaultValue focuses current parameter default value for Model and returns the resulting state or error.
func (m *Model) FocusCurrentParameterDefaultValue() bool {
	return m.focusCurrentParameterDefaultValue()
}

func (m Model) CurrentProject() (core.Project, bool) {
	return m.currentProject()
}

func (m Model) CurrentGroupRef() (core.Project, string, string, bool) {
	if m.cursor < 0 || m.cursor >= len(m.visible) {
		return core.Project{}, "", "", false
	}
	node := m.visible[m.cursor]
	if node.kind != nodeGroup || node.transient || core.NormalizeRemoteConfigGroupKey(node.groupKey) == "" {
		return core.Project{}, "", "", false
	}
	project := m.projectByID(node.projectID)
	if project == nil {
		return core.Project{}, "", "", false
	}
	return project.project, node.groupKey, node.label, true
}

func (m Model) CurrentNewParameterTarget() (core.Project, string, string, bool) {
	if m.cursor < 0 || m.cursor >= len(m.visible) {
		return core.Project{}, "", "", false
	}
	node := m.visible[m.cursor]
	project := m.projectByID(node.projectID)
	if project == nil {
		return core.Project{}, "", "", false
	}
	groupKey := "__default__"
	afterParamKey := ""
	switch node.kind {
	case nodeGroup:
		groupKey = node.groupKey
	case nodeParameter:
		groupKey = node.groupKey
		if !node.transient {
			afterParamKey = node.paramKey
		}
	case nodeValue:
		groupKey = node.groupKey
		afterParamKey = node.paramKey
	case nodeProject:
		groupKey = "__default__"
	default:
		if node.groupKey != "" {
			groupKey = node.groupKey
		}
	}
	return project.project, groupKey, afterParamKey, true
}

func (m Model) DraftProjects() []core.Project {
	out := make([]core.Project, 0)
	for _, project := range m.projects {
		if project.hasDraft {
			out = append(out, project.project)
		}
	}
	return out
}

// HasDraft reports draft for Model and returns the resulting state or error.
func (m Model) HasDraft(projectID string) bool {
	project := m.projectByID(projectID)
	return project != nil && project.hasDraft
}

// HasProject reports project for Model and returns the resulting state or error.
func (m Model) HasProject(projectID string) bool {
	return m.projectByID(projectID) != nil
}

func (m Model) CurrentRenameAnchor() (RenameAnchor, bool) {
	if m.cursor < 0 || m.cursor >= len(m.visible) {
		return RenameAnchor{}, false
	}
	node := m.visible[m.cursor]
	project := m.projectByID(node.projectID)
	if project == nil {
		return RenameAnchor{}, false
	}
	switch node.kind {
	case nodeGroup:
		if core.NormalizeRemoteConfigGroupKey(node.groupKey) == "" {
			return RenameAnchor{}, false
		}
		screenLine := m.screenLineForOffset(m.cursor, m.offset)
		if screenLine < 0 {
			return RenameAnchor{}, false
		}
		return RenameAnchor{
			Project:  project.project,
			IsGroup:  true,
			GroupKey: node.groupKey,
			Label:    node.label,
			X:        m.x + 1,
			Y:        m.y + screenLine,
			Width:    max(lipgloss.Width(node.label), 1),
			MaxWidth: max(m.viewportWidth()-3, 1),
		}, true
	case nodeParameter, nodeValue:
		if node.transient {
			screenLine := m.screenLineForOffset(m.cursor, m.offset)
			if screenLine < 0 {
				return RenameAnchor{}, false
			}
			layout := m.parameterRenderLayout()
			return RenameAnchor{
				Project:  project.project,
				GroupKey: node.groupKey,
				ParamKey: node.paramKey,
				Label:    node.label,
				X:        m.x + layout.paramStart - 1,
				Y:        m.y + screenLine,
				Width:    max(lipgloss.Width(node.label), 1),
				MaxWidth: max(m.viewportWidth()-layout.paramStart-1, 1),
			}, true
		}
		_, groupKey, paramKey, ok := m.currentParameterRef()
		if !ok {
			return RenameAnchor{}, false
		}
		paramIndex := m.currentParameterNodeIndex()
		if paramIndex < 0 {
			return RenameAnchor{}, false
		}
		screenLine := m.screenLineForOffset(paramIndex, m.offset)
		if screenLine < 0 {
			return RenameAnchor{}, false
		}
		layout := m.parameterRenderLayout()
		return RenameAnchor{
			Project:  project.project,
			GroupKey: groupKey,
			ParamKey: paramKey,
			Label:    node.label,
			X:        m.x + layout.paramStart - 1,
			Y:        m.y + screenLine,
			Width:    max(lipgloss.Width(node.label), 1),
			MaxWidth: max(m.viewportWidth()-layout.paramStart-1, 1),
		}, true
	default:
		return RenameAnchor{}, false
	}
}

func (m Model) CurrentTransientDuplicate() (project core.Project, groupKey, sourceParamKey, label string, ok bool) {
	if m.cursor < 0 || m.cursor >= len(m.visible) {
		return core.Project{}, "", "", "", false
	}
	node := m.visible[m.cursor]
	if !node.transient {
		return core.Project{}, "", "", "", false
	}
	projectState := m.projectByID(node.projectID)
	if projectState == nil {
		return core.Project{}, "", "", "", false
	}
	return projectState.project, node.groupKey, node.paramKey, node.label, true
}

func (m Model) valueNodeValueX(node visibleNode, param *core.ParametersEntry) int {
	layout := m.parameterRenderLayout()
	label := displayConditionLabel(param.Values[node.valueIdx].Label)
	conditionWidth := parameterConditionWidth(param)
	if layout.mode == parameterRenderModeNarrow {
		fillerWidth := max(conditionWidth-lipgloss.Width(label)+1, 1)
		return lipgloss.Width(compactBranchGlyph(layout.paramStart, m.valueConnector(node, param))) + 1 + lipgloss.Width(label) + 1 + fillerWidth + 1
	}
	leafOffset := 1
	if len(param.Values) == 1 {
		leafOffset = 2
	}
	leafOffset++
	leafValueStart := layout.valueStart + leafOffset
	labelStart := max(leafValueStart-conditionWidth-4, layout.paramStart+2)
	fillerWidth := max(leafValueStart-labelStart-lipgloss.Width(label)-3, 1)
	return lipgloss.Width(branchGlyph(layout.paramStart, labelStart, m.valueConnector(node, param))) + 1 + lipgloss.Width(label) + 1 + fillerWidth + 1
}

// OpenTransientDuplicate opens transient duplicate for Model and returns the resulting state or error.
func (m *Model) OpenTransientDuplicate(projectID, groupKey, sourceParamKey, label string) {
	m.transientDup = &transientDuplicate{
		projectID:     projectID,
		groupKey:      groupKey,
		afterParamKey: sourceParamKey,
		label:         label,
	}
	m.groupExpanded[m.groupKey(projectID, groupKey)] = true
	m.syncVisible()
	for i, node := range m.visible {
		if node.transient && node.projectID == projectID && node.groupKey == groupKey && node.paramKey == sourceParamKey && node.label == label {
			m.cursor = i
			m.ensureCursorVisible()
			return
		}
	}
}

func (m *Model) ClearTransientDuplicate() {
	if m.transientDup == nil {
		return
	}
	m.transientDup = nil
	m.syncVisible()
}

func (m *Model) ClearTransientDuplicateAndFocusSource() {
	if m.transientDup == nil {
		return
	}
	projectID := m.transientDup.projectID
	groupKey := m.transientDup.groupKey
	paramKey := m.transientDup.afterParamKey
	m.transientDup = nil
	m.syncVisible()
	m.selectParameter(projectID, groupKey, paramKey)
}

func (m *Model) ClearTransientDuplicateAndFocus(projectID, groupKey, paramKey string) {
	if m.transientDup == nil {
		return
	}
	m.transientDup = nil
	m.syncVisible()
	m.selectParameter(projectID, groupKey, paramKey)
}

// OpenTransientNewParameter opens transient new parameter for Model and returns the resulting state or error.
func (m *Model) OpenTransientNewParameter(projectID, groupKey, afterParamKey string) {
	m.transientNew = &transientNewParameter{
		projectID:     projectID,
		groupKey:      groupKey,
		afterParamKey: afterParamKey,
		label:         "",
	}
	m.groupExpanded[m.groupKey(projectID, groupKey)] = true
	m.syncVisible()
	for i, node := range m.visible {
		if node.transient && node.projectID == projectID && node.groupKey == groupKey && node.paramKey == "" {
			m.cursor = i
			m.ensureCursorVisible()
			return
		}
	}
}

func (m *Model) ClearTransientNewParameter() {
	if m.transientNew == nil {
		return
	}
	m.transientNew = nil
	m.syncVisible()
}

func (m *Model) ClearTransientNewParameterAndFocus(projectID, groupKey, paramKey string) {
	if m.transientNew == nil {
		return
	}
	m.transientNew = nil
	m.syncVisible()
	m.selectParameter(projectID, groupKey, paramKey)
}

// FocusParameter focuses parameter for Model and returns the resulting state or error.
func (m *Model) FocusParameter(projectID, groupKey, paramKey string) bool {
	return m.selectParameter(projectID, groupKey, paramKey)
}

// selectParameter selects select parameter for Model and returns the resulting state or error.
func (m *Model) selectParameter(projectID, groupKey, paramKey string) bool {
	if projectID == "" || paramKey == "" {
		return false
	}
	m.groupExpanded[m.groupKey(projectID, groupKey)] = true
	m.syncVisible()
	for i, node := range m.visible {
		if node.kind == nodeParameter && node.projectID == projectID && node.groupKey == groupKey && node.paramKey == paramKey {
			m.cursor = i
			m.ensureCursorVisible()
			return true
		}
	}
	return false
}

func (m Model) CurrentMoveAnchor() (MoveAnchor, bool) {
	if m.cursor < 0 || m.cursor >= len(m.visible) {
		return MoveAnchor{}, false
	}
	node := m.visible[m.cursor]
	project := m.projectByID(node.projectID)
	if project == nil || project.tree == nil {
		return MoveAnchor{}, false
	}

	switch node.kind {
	case nodeGroup:
		if node.transient {
			return MoveAnchor{}, false
		}
		screenLine := m.screenLineForOffset(m.cursor, m.offset)
		if screenLine < 0 {
			return MoveAnchor{}, false
		}
		currentNormalized := core.NormalizeRemoteConfigGroupKey(node.groupKey)
		options := make([]MoveOption, 0, len(project.tree.Groups))
		for _, group := range project.tree.Groups {
			groupNormalized := core.NormalizeRemoteConfigGroupKey(group.Key)
			if groupNormalized == currentNormalized {
				continue
			}
			if groupNormalized == "" {
				continue
			}
			options = append(options, MoveOption{Key: group.Key, Label: group.Label})
		}
		if currentNormalized != "" {
			options = append(options, MoveOption{Key: "", Label: "(root)"})
		}
		if len(options) == 0 {
			return MoveAnchor{}, false
		}
		return MoveAnchor{
			Project:  project.project,
			IsGroup:  true,
			GroupKey: node.groupKey,
			Label:    node.label,
			X:        m.x + 1,
			Y:        m.y + screenLine,
			Options:  options,
		}, true
	case nodeParameter, nodeValue:
		_, groupKey, paramKey, ok := m.currentParameterRef()
		if !ok {
			return MoveAnchor{}, false
		}
		paramIndex := m.currentParameterNodeIndex()
		if paramIndex < 0 {
			return MoveAnchor{}, false
		}
		screenLine := m.screenLineForOffset(paramIndex, m.offset)
		if screenLine < 0 {
			return MoveAnchor{}, false
		}
		options := make([]MoveOption, 0, len(project.tree.Groups)+1)
		currentNormalized := core.NormalizeRemoteConfigGroupKey(groupKey)
		for _, group := range project.tree.Groups {
			groupNormalized := core.NormalizeRemoteConfigGroupKey(group.Key)
			if groupNormalized == "" || groupNormalized == currentNormalized {
				continue
			}
			options = append(options, MoveOption{Key: group.Key, Label: group.Label})
		}
		if currentNormalized != "" {
			options = append(options, MoveOption{Key: "", Label: "(root)"})
		}
		layout := m.parameterRenderLayout()
		return MoveAnchor{
			Project:  project.project,
			IsGroup:  false,
			GroupKey: groupKey,
			ParamKey: paramKey,
			Label:    paramKey,
			X:        m.x + layout.paramStart - 1,
			Y:        m.y + screenLine,
			Options:  options,
		}, true
	default:
		return MoveAnchor{}, false
	}
}

func (m Model) CurrentConditionalValueAnchor() (ConditionalValueAnchor, bool) {
	if m.cursor < 0 || m.cursor >= len(m.visible) {
		return ConditionalValueAnchor{}, false
	}
	node := m.visible[m.cursor]
	if node.kind != nodeValue || node.transient {
		return ConditionalValueAnchor{}, false
	}
	project := m.projectByID(node.projectID)
	param := m.parameterByKey(node.projectID, node.groupKey, node.paramKey)
	if project == nil || param == nil || node.valueIdx <= 0 || node.valueIdx >= len(param.Values) {
		return ConditionalValueAnchor{}, false
	}
	value := param.Values[node.valueIdx]
	if value.Label == "" || value.Label == "default" {
		return ConditionalValueAnchor{}, false
	}
	return ConditionalValueAnchor{
		Project:    project.project,
		GroupKey:   node.groupKey,
		ParamKey:   node.paramKey,
		ValueLabel: value.Label,
	}, true
}

func (m Model) CurrentBoolValueAnchor() (BoolValueAnchor, bool) {
	if m.cursor < 0 || m.cursor >= len(m.visible) {
		return BoolValueAnchor{}, false
	}
	node := m.visible[m.cursor]
	if node.kind != nodeValue || node.transient {
		return BoolValueAnchor{}, false
	}
	project := m.projectByID(node.projectID)
	param := m.parameterByKey(node.projectID, node.groupKey, node.paramKey)
	if project == nil || param == nil || node.valueIdx < 0 || node.valueIdx >= len(param.Values) {
		return BoolValueAnchor{}, false
	}
	value := param.Values[node.valueIdx]
	if strings.TrimSpace(strings.ToLower(value.ValueType)) != "boolean" {
		return BoolValueAnchor{}, false
	}
	switch strings.TrimSpace(strings.ToLower(value.Value)) {
	case "true", "false":
	default:
		return BoolValueAnchor{}, false
	}
	screenLine := m.screenLineForOffset(m.cursor, m.offset)
	if screenLine < 0 {
		return BoolValueAnchor{}, false
	}
	valueX := m.valueNodeValueX(node, param)
	return BoolValueAnchor{
		Project:      project.project,
		GroupKey:     node.groupKey,
		ParamKey:     node.paramKey,
		ValueLabel:   value.Label,
		Value:        strings.EqualFold(value.Value, "true"),
		CurrentValue: value.RawValue,
		X:            m.x + valueX - 1,
		Y:            m.y + screenLine + 1,
	}, true
}

func (m Model) CurrentNumberValueAnchor() (NumberValueAnchor, bool) {
	if m.cursor < 0 || m.cursor >= len(m.visible) {
		return NumberValueAnchor{}, false
	}
	node := m.visible[m.cursor]
	if node.kind != nodeValue || node.transient {
		return NumberValueAnchor{}, false
	}
	project := m.projectByID(node.projectID)
	param := m.parameterByKey(node.projectID, node.groupKey, node.paramKey)
	if project == nil || param == nil || node.valueIdx < 0 || node.valueIdx >= len(param.Values) {
		return NumberValueAnchor{}, false
	}
	value := param.Values[node.valueIdx]
	if strings.TrimSpace(strings.ToLower(value.ValueType)) != "number" {
		return NumberValueAnchor{}, false
	}
	currentValue := strings.TrimSpace(value.Value)
	screenLine := m.screenLineForOffset(m.cursor, m.offset)
	if screenLine < 0 {
		return NumberValueAnchor{}, false
	}
	valueX := m.valueNodeValueX(node, param)
	return NumberValueAnchor{
		Project:      project.project,
		GroupKey:     node.groupKey,
		ParamKey:     node.paramKey,
		ValueLabel:   value.Label,
		CurrentValue: currentValue,
		X:            m.x + valueX - 1,
		Y:            m.y + screenLine,
		Width:        max(lipgloss.Width(currentValue), 3),
		MaxWidth:     max(m.viewportWidth()-valueX-1, 3),
	}, true
}

func (m Model) CurrentStringValueAnchor() (StringValueAnchor, bool) {
	if m.cursor < 0 || m.cursor >= len(m.visible) {
		return StringValueAnchor{}, false
	}
	node := m.visible[m.cursor]
	if node.kind != nodeValue || node.transient {
		return StringValueAnchor{}, false
	}
	project := m.projectByID(node.projectID)
	param := m.parameterByKey(node.projectID, node.groupKey, node.paramKey)
	if project == nil || param == nil || node.valueIdx < 0 || node.valueIdx >= len(param.Values) {
		return StringValueAnchor{}, false
	}
	value := param.Values[node.valueIdx]
	valueType := strings.TrimSpace(strings.ToLower(value.ValueType))
	if valueType != "string" && valueType != "" {
		return StringValueAnchor{}, false
	}
	if !value.Plain {
		return StringValueAnchor{}, false
	}
	screenLine := m.screenLineForOffset(m.cursor, m.offset)
	if screenLine < 0 {
		return StringValueAnchor{}, false
	}
	valueX := m.valueNodeValueX(node, param)
	currentValue := value.RawValue
	minWidth := max(lipgloss.Width(currentValue), 15)
	maxWidth := max(m.width-(valueX-1), 1)
	fullWidth := max(maxWidth-4, 1) < minWidth
	return StringValueAnchor{
		Project:      project.project,
		GroupKey:     node.groupKey,
		ParamKey:     node.paramKey,
		ValueLabel:   value.Label,
		CurrentValue: currentValue,
		X:            m.x + valueX - 1,
		Y:            m.y + screenLine,
		Width:        minWidth,
		MaxWidth:     maxWidth,
		FullWidth:    fullWidth,
		Expanded:     strings.Contains(currentValue, "\n"),
	}, true
}

func (m Model) CurrentJSONValueAnchor() (JSONValueAnchor, bool) {
	if m.cursor < 0 || m.cursor >= len(m.visible) {
		return JSONValueAnchor{}, false
	}
	node := m.visible[m.cursor]
	if node.kind != nodeValue || node.transient {
		return JSONValueAnchor{}, false
	}
	project := m.projectByID(node.projectID)
	param := m.parameterByKey(node.projectID, node.groupKey, node.paramKey)
	if project == nil || param == nil || node.valueIdx < 0 || node.valueIdx >= len(param.Values) {
		return JSONValueAnchor{}, false
	}
	value := param.Values[node.valueIdx]
	if strings.TrimSpace(strings.ToLower(value.ValueType)) != "json" {
		return JSONValueAnchor{}, false
	}
	if !value.Plain {
		return JSONValueAnchor{}, false
	}
	return JSONValueAnchor{
		Project:      project.project,
		GroupKey:     node.groupKey,
		ParamKey:     node.paramKey,
		ValueLabel:   value.Label,
		CurrentValue: value.RawValue,
	}, true
}

func (m Model) ProjectDraftState(projectID string) (bool, bool) {
	project := m.projectByID(projectID)
	if project == nil {
		return false, false
	}
	return project.hasDraft, project.staleDraft
}

func (m Model) currentParameterNodeIndex() int {
	if m.cursor < 0 || m.cursor >= len(m.visible) {
		return -1
	}
	node := m.visible[m.cursor]
	switch node.kind {
	case nodeParameter:
		return m.cursor
	case nodeValue:
		for i := m.cursor - 1; i >= 0; i-- {
			prev := m.visible[i]
			if prev.projectID != node.projectID || prev.groupKey != node.groupKey || prev.paramKey != node.paramKey {
				break
			}
			if prev.kind == nodeParameter {
				return i
			}
		}
	}
	return -1
}

func remoteConfigVersion(raw []byte) string {
	cfg, err := firebase.ParseRemoteConfig(raw)
	if err != nil {
		return ""
	}
	return cfg.Version.VersionNumber
}

func versionLess(left, right string) bool {
	left = strings.TrimSpace(left)
	right = strings.TrimSpace(right)
	if left == "" || right == "" {
		return false
	}
	var leftNum, rightNum big.Int
	if _, ok := leftNum.SetString(left, 10); !ok {
		return false
	}
	if _, ok := rightNum.SetString(right, 10); !ok {
		return false
	}
	return leftNum.Cmp(&rightNum) < 0
}

func displayConditionLabel(label string) string {
	if label == "default" {
		return "Default value"
	}
	return label
}

func truncatePlain(value string, width int) string {
	if width <= 0 {
		return ""
	}

	runes := []rune(value)
	if len(runes) <= width {
		return value
	}

	return string(runes[:width])
}

func (m Model) captureSelectionSnapshot(expanding, groups bool) selectionSnapshot {
	snapshot := selectionSnapshot{valueIdx: -1}
	if len(m.visible) == 0 || m.cursor < 0 || m.cursor >= len(m.visible) {
		return snapshot
	}

	node := m.visible[m.cursor]
	snapshot.projectID = node.projectID
	snapshot.groupKey = node.groupKey
	snapshot.paramKey = node.paramKey
	snapshot.valueIdx = node.valueIdx
	snapshot.kind = node.kind
	snapshot.screenLine = m.screenLineForOffset(m.cursor, m.offset)

	if expanding {
		return snapshot
	}

	if groups {
		if node.kind == nodeParameter || node.kind == nodeValue {
			snapshot.kind = nodeGroup
			snapshot.paramKey = ""
			snapshot.valueIdx = -1
		}
		return snapshot
	}

	if node.kind == nodeValue {
		snapshot.kind = nodeParameter
		snapshot.valueIdx = -1
	}
	return snapshot
}

func (m *Model) applyFilter() {
	snapshot := selectionSnapshot{valueIdx: -1}
	if len(m.visible) > 0 && m.cursor >= 0 && m.cursor < len(m.visible) {
		snapshot = m.captureSelectionSnapshot(true, false)
	}
	m.syncVisible()
	if len(m.visible) > 0 {
		m.restoreSelectionSnapshot(snapshot)
	}
}

func (m *Model) restoreSelectionSnapshot(snapshot selectionSnapshot) {
	if len(m.visible) == 0 {
		m.cursor = 0
		m.offset = 0
		return
	}

	cursor := m.findSelectionSnapshotNode(snapshot)
	if cursor < 0 {
		cursor = min(max(m.cursor, 0), len(m.visible)-1)
	}
	m.cursor = cursor

	bestOffset := m.offset
	bestScore := int(^uint(0) >> 1)
	for offset := 0; offset < max(m.totalLines, 1); offset++ {
		screenLine := m.screenLineForOffset(m.cursor, offset)
		if screenLine < 0 {
			continue
		}
		score := abs(screenLine - snapshot.screenLine)
		if score < bestScore {
			bestScore = score
			bestOffset = offset
			if score == 0 {
				break
			}
		}
	}
	m.offset = bestOffset
	m.ensureCursorVisible()
}

func (m Model) findSelectionSnapshotNode(snapshot selectionSnapshot) int {
	fallbackProject := -1
	fallbackGroup := -1
	fallbackParam := -1

	for i, node := range m.visible {
		if node.projectID != snapshot.projectID {
			continue
		}
		if fallbackProject < 0 && node.kind == nodeProject {
			fallbackProject = i
		}
		if node.groupKey == snapshot.groupKey && fallbackGroup < 0 && node.kind == nodeGroup {
			fallbackGroup = i
		}
		if node.groupKey == snapshot.groupKey && node.paramKey == snapshot.paramKey && fallbackParam < 0 && node.kind == nodeParameter {
			fallbackParam = i
		}

		switch snapshot.kind {
		case nodeProject:
			if node.kind == nodeProject {
				return i
			}
		case nodeGroup:
			if node.kind == nodeGroup && node.groupKey == snapshot.groupKey {
				return i
			}
		case nodeParameter:
			if node.kind == nodeParameter && node.groupKey == snapshot.groupKey && node.paramKey == snapshot.paramKey {
				return i
			}
		case nodeValue:
			if node.kind == nodeValue && node.groupKey == snapshot.groupKey && node.paramKey == snapshot.paramKey && node.valueIdx == snapshot.valueIdx {
				return i
			}
		}
	}

	if snapshot.kind == nodeValue || snapshot.kind == nodeParameter {
		if fallbackParam >= 0 {
			return fallbackParam
		}
	}
	if snapshot.kind == nodeValue || snapshot.kind == nodeParameter || snapshot.kind == nodeGroup {
		if fallbackGroup >= 0 {
			return fallbackGroup
		}
	}
	return fallbackProject
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
