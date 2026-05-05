package parameters

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strings"
	"time"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/atotto/clipboard"

	"fbrcm/core"
	"fbrcm/core/filter"
	"fbrcm/tui/components/filterbox"
	"fbrcm/tui/messages"
	"fbrcm/tui/styles"
)

const panelTitle = "[2] Parameters"

var (
	projectMetaStyle        = styles.PanelMuted
	parameterStyle          = styles.PanelBody.Foreground(styles.PaletteBlueBright)
	parameterValueStyle     = styles.PanelMuted
	parameterEmptyValue     = styles.PanelMuted.Italic(true)
	parameterSeparatorStyle = styles.PanelMuted
	descriptionStyle        = styles.PanelMuted.Italic(true)
	conditionDefaultStyle   = styles.PanelMuted.Italic(true)
)

var deprecatedDescriptionPattern = regexp.MustCompile(`(?i)deprecat|obsolete|sunset|retired?|no longer|use .+ instead|replaced?|superseded?|removed?`)

type projectState struct {
	project   core.Project
	tree      *core.ParametersTree
	source    string
	err       error
	loading   bool
	verifying bool
}

type projectLayout struct {
	metadataWidth int
	nameWidth     int
	iconWidth     int
	paramStart    int
	valueStart    int
	valueWidth    int
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

func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height
	m.syncVisible()
	return m
}

func (m Model) SetBounds(x, y, width, height int) Model {
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
			nodes = append(nodes, visibleNode{
				kind:      nodeValue,
				projectID: project.project.ProjectID,
				label:     "No parameters",
			})
			continue
		}

		for _, group := range project.tree.Groups {
			matchedParams := group.Parameters
			if filtering {
				matchedParams = matchedParameters(group.Parameters, query, m.filter.Mode())
				if len(matchedParams) == 0 {
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
		}
	}

	return nodes
}

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

func (m Model) loadParametersCmd(project core.Project) tea.Cmd {
	return func() tea.Msg {
		cache, state, err := m.svc.InspectParametersCache(project.ProjectID)
		if err != nil {
			return messages.ParametersLoadedMsg{Project: project, Err: err}
		}

		switch state {
		case core.ParametersCacheFresh:
			tree, err := m.svc.BuildParametersTree(cache)
			return messages.ParametersLoadedMsg{Project: project, Tree: tree, Source: "cache", Err: err}
		case core.ParametersCacheStale:
			tree, err := m.svc.BuildParametersTree(cache)
			return messages.ParametersLoadedMsg{Project: project, Tree: tree, Source: "cache-stale", Err: err, Revalidate: true}
		default:
			cache, source, err := m.svc.GetParameters(context.Background(), project.ProjectID, false)
			if err != nil {
				return messages.ParametersLoadedMsg{Project: project, Err: err}
			}
			tree, err := m.svc.BuildParametersTree(cache)
			return messages.ParametersLoadedMsg{Project: project, Tree: tree, Source: source, Err: err}
		}
	}
}

func (m Model) revalidateParametersCmd(project core.Project) tea.Cmd {
	return func() tea.Msg {
		cache, source, err := m.svc.GetParameters(context.Background(), project.ProjectID, false)
		if err != nil {
			return messages.ParametersLoadedMsg{Project: project, Err: err}
		}
		tree, err := m.svc.BuildParametersTree(cache)
		return messages.ParametersLoadedMsg{Project: project, Tree: tree, Source: source, Err: err}
	}
}

func (m Model) forceParametersCmd(project core.Project) tea.Cmd {
	return func() tea.Msg {
		cache, source, err := m.svc.GetParameters(context.Background(), project.ProjectID, true)
		if err != nil {
			return messages.ParametersLoadedMsg{Project: project, Err: err}
		}
		tree, err := m.svc.BuildParametersTree(cache)
		return messages.ParametersLoadedMsg{Project: project, Tree: tree, Source: source, Err: err}
	}
}

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
		state.err = nil
	}
	state.loading = false
	state.verifying = false
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

func (m *Model) moveCursor(delta int) {
	if len(m.visible) == 0 {
		return
	}
	m.cursor = max(0, min(m.cursor+delta, len(m.visible)-1))
	m.ensureCursorVisible()
}

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

func (m *Model) focusParentGroup(node visibleNode) {
	for i, candidate := range m.visible {
		if candidate.kind == nodeGroup && candidate.projectID == node.projectID && candidate.groupKey == node.groupKey {
			m.cursor = i
			m.ensureCursorVisible()
			return
		}
	}
}

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
	switch node.kind {
	case nodeParameter, nodeValue:
		return node.projectID, node.groupKey, node.paramKey, true
	default:
		return "", "", "", false
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
		return core.ParametersStatusLabel(p.source, time.Time{}, false, p.err)
	}
	return core.ParametersStatusLabel(p.source, p.tree.CachedAt, true, p.err)
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
	layout := projectLayout{
		iconWidth:  1,
		paramStart: 2,
	}

	project := m.projectByID(projectID)
	if project == nil {
		return layout
	}

	metadata := m.projectMeta(project)
	layout.metadataWidth = lipgloss.Width(metadata)
	if strings.TrimSpace(project.project.Name) == "" {
		layout.nameWidth = lipgloss.Width(project.project.ProjectID)
	} else {
		layout.nameWidth = max(
			lipgloss.Width(project.project.Name),
			lipgloss.Width(project.project.ProjectID),
		)
	}

	if project.tree != nil {
		for _, group := range project.tree.Groups {
			for _, param := range group.Parameters {
				layout.nameWidth = max(layout.nameWidth, lipgloss.Width(param.Key))
			}
		}
	}

	layout.valueStart = layout.paramStart + layout.nameWidth + 3
	layout.valueWidth = max(m.width-2-layout.valueStart, 1)
	return layout
}

func (m Model) projectMeta(project *projectState) string {
	if project == nil {
		return ""
	}

	parts := make([]string, 0, 3)
	if project.tree != nil && project.tree.Version != "" {
		parts = append(parts, "v"+project.tree.Version)
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
	return strings.Join(parts, " ")
}

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
