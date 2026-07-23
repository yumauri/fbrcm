package parameters

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/tui/components/viewutil"
	tuiconfig "github.com/yumauri/fbrcm/tui/config"
	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/styles"
)

func (m Model) updateHistoryKey(msg tea.KeyMsg, key string) (Model, tea.Cmd, bool) {
	if m.versionPicker != nil {
		return m.updateHistoryPickerKey(key)
	}
	action := ""
	switch {
	case tuiconfig.Matches(tuiconfig.BlockHistoryPicker, tuiconfig.ActionHistoryBothOlder, key):
		action = "both-older"
	case tuiconfig.Matches(tuiconfig.BlockHistoryPicker, tuiconfig.ActionHistoryBothNewer, key):
		action = "both-newer"
	case tuiconfig.Matches(tuiconfig.BlockHistory, tuiconfig.ActionHistoryChoose, key):
		m.openHistoryPicker()
		return m, nil, true
	case !m.filter.Focused() && tuiconfig.Matches(tuiconfig.BlockHistory, tuiconfig.ActionHistoryChanges, key):
		return m.toggleHistoryChangesOnly(), nil, true
	case !m.filter.Focused() && tuiconfig.Matches(tuiconfig.BlockHistory, tuiconfig.ActionSubmit, key):
		return m, m.historyDiffRequestedCmd(), true
	default:
		return m, nil, false
	}
	return m.stepHistory(action)
}

func (m Model) toggleHistoryChangesOnly() Model {
	snapshot := selectionSnapshot{valueIdx: -1}
	fallbackCursor := m.cursor
	if len(m.visible) > 0 && m.cursor >= 0 && m.cursor < len(m.visible) {
		snapshot = m.captureSelectionSnapshot(true, false)
		if m.historyViews == nil {
			m.historyViews = make(map[bool]selectionSnapshot)
		}
		m.historyViews[m.historyChangesOnly] = snapshot
	}
	m.historyChangesOnly = !m.historyChangesOnly
	m.syncVisible()
	if destination, ok := m.historyViews[m.historyChangesOnly]; ok {
		m.restoreSelectionSnapshot(destination)
		return m
	}
	if cursor := m.findExactSelectionSnapshotNode(snapshot); cursor >= 0 {
		m.cursor = cursor
	} else if len(m.visible) > 0 {
		m.cursor = min(max(fallbackCursor, 0), len(m.visible)-1)
	}
	m.restoreSelectionScreenLine(snapshot.screenLine)
	return m
}

func (m Model) stepHistory(action string) (Model, tea.Cmd, bool) {
	project, ok := m.currentProject()
	if !ok {
		return m, nil, true
	}
	state := m.histories[project.ProjectID]
	if state.loading {
		return m, nil, true
	}
	left, right := historyVersionIndex(state.versions, state.previousVersion), historyVersionIndex(state.versions, state.currentVersion)
	if left < 0 || right < 0 {
		return m, nil, true
	}
	nextLeft, nextRight := left, right
	switch action {
	case "both-older":
		nextLeft++
		nextRight++
	case "both-newer":
		nextLeft--
		nextRight--
	}
	if nextLeft < 0 || nextRight < 0 || nextLeft >= len(state.versions) || nextRight >= len(state.versions) || nextLeft < nextRight {
		return m, nil, true
	}
	return m.selectHistoryPair(project, nextLeft, nextRight)
}

func historyVersionIndex(versions []core.RemoteConfigVersionEntry, version string) int {
	for i := range versions {
		if versions[i].VersionNumber == version {
			return i
		}
	}
	return -1
}

func (m Model) selectHistoryPair(project core.Project, leftIndex, rightIndex int) (Model, tea.Cmd, bool) {
	state := m.histories[project.ProjectID]
	if state.loading {
		return m, nil, true
	}
	if leftIndex < 0 || rightIndex < 0 || leftIndex >= len(state.versions) || rightIndex >= len(state.versions) {
		return m, nil, true
	}
	left, right := state.versions[leftIndex].VersionNumber, state.versions[rightIndex].VersionNumber
	if data, ok := state.pairs[historyPairKey(left, right)]; ok {
		state.previous, state.current = data.previous, data.current
		state.previousVersion, state.currentVersion = data.previousVersion, data.currentVersion
		state.previousPublished, state.currentPublished = data.previousPublished, data.currentPublished
		state.err, state.loading = nil, false
		m.histories[project.ProjectID] = buildHistoryState(state)
		m.syncVisible()
		return m, nil, true
	}
	state.loading, state.err = true, nil
	m.histories[project.ProjectID] = state
	return m, m.loadHistoryPairCmd(project, left, right), true
}

func (m *Model) openHistoryPicker() {
	project, ok := m.currentProject()
	if !ok {
		return
	}
	state := m.histories[project.ProjectID]
	if len(state.versions) == 0 {
		return
	}
	leftCursor := historyVersionIndex(state.versions, state.previousVersion)
	rightCursor := historyVersionIndex(state.versions, state.currentVersion)
	if leftCursor < 0 || rightCursor < 0 {
		return
	}
	m.versionPicker = &historyVersionPicker{
		projectID:   project.ProjectID,
		left:        true,
		leftCursor:  leftCursor,
		rightCursor: rightCursor,
	}
	m.ensureHistoryPickerSpace()
}

func (m Model) updateHistoryPickerKey(key string) (Model, tea.Cmd, bool) {
	picker := m.versionPicker
	state := m.histories[picker.projectID]
	low, high := historyPickerBounds(len(state.versions), picker.leftCursor, picker.rightCursor, picker.left)
	cursor := picker.rightCursor
	if picker.left {
		cursor = picker.leftCursor
	}
	switch {
	case tuiconfig.Matches(tuiconfig.BlockHistoryPicker, tuiconfig.ActionCancel, key):
		m.versionPicker = nil
		return m, nil, true
	case tuiconfig.Matches(tuiconfig.BlockHistoryPicker, tuiconfig.ActionToggle, key):
		picker.left = !picker.left
		m.versionPicker = picker
		return m, nil, true
	case tuiconfig.Matches(tuiconfig.BlockHistoryPicker, tuiconfig.ActionLeft, key):
		picker.left = true
		m.versionPicker = picker
		return m, nil, true
	case tuiconfig.Matches(tuiconfig.BlockHistoryPicker, tuiconfig.ActionRight, key):
		picker.left = false
		m.versionPicker = picker
		return m, nil, true
	case tuiconfig.Matches(tuiconfig.BlockHistoryPicker, tuiconfig.ActionReset, key):
		left, right, ok := historyDefaultPairIndices(state.versions)
		m.versionPicker = nil
		if !ok {
			return m, nil, true
		}
		project := m.projects[m.projectIndex[picker.projectID]].project
		return m.selectHistoryPair(project, left, right)
	case tuiconfig.Matches(tuiconfig.BlockHistoryPicker, tuiconfig.ActionHistoryBothOlder, key):
		picker = stageHistoryPickerPair(picker, len(state.versions), 1)
		m.versionPicker = picker
		return m, nil, true
	case tuiconfig.Matches(tuiconfig.BlockHistoryPicker, tuiconfig.ActionHistoryBothNewer, key):
		picker = stageHistoryPickerPair(picker, len(state.versions), -1)
		m.versionPicker = picker
		return m, nil, true
	case tuiconfig.Matches(tuiconfig.BlockHistoryPicker, tuiconfig.ActionHistoryRollback, key):
		index := picker.rightCursor
		if picker.left {
			index = picker.leftCursor
		}
		if index < 0 || index >= len(state.versions) || state.versions[index].Current {
			return m, nil, true
		}
		request := messages.HistoryRollbackRequestedMsg{
			Project: m.projects[m.projectIndex[picker.projectID]].project,
			Target:  state.versions[index], PickerLeft: picker.left,
			LeftCursor: picker.leftCursor, RightCursor: picker.rightCursor,
		}
		m.versionPicker = nil
		return m, func() tea.Msg { return request }, true
	case tuiconfig.Matches(tuiconfig.BlockHistoryPicker, tuiconfig.ActionUp, key):
		cursor = max(low, cursor-1)
	case tuiconfig.Matches(tuiconfig.BlockHistoryPicker, tuiconfig.ActionDown, key):
		cursor = min(high, cursor+1)
	case tuiconfig.Matches(tuiconfig.BlockHistoryPicker, tuiconfig.ActionPageUp, key):
		cursor = max(low, cursor-10)
	case tuiconfig.Matches(tuiconfig.BlockHistoryPicker, tuiconfig.ActionPageDown, key):
		cursor = min(high, cursor+10)
	case tuiconfig.Matches(tuiconfig.BlockHistoryPicker, tuiconfig.ActionHome, key):
		cursor = low
	case tuiconfig.Matches(tuiconfig.BlockHistoryPicker, tuiconfig.ActionEnd, key):
		cursor = high
	case tuiconfig.Matches(tuiconfig.BlockHistoryPicker, tuiconfig.ActionSubmit, key):
		project := m.projects[m.projectIndex[picker.projectID]].project
		m.versionPicker = nil
		return m.selectHistoryPair(project, picker.leftCursor, picker.rightCursor)
	}
	if picker.left {
		picker.leftCursor = cursor
	} else {
		picker.rightCursor = cursor
	}
	m.versionPicker = picker
	return m, nil, true
}

func stageHistoryPickerPair(picker *historyVersionPicker, versionCount, delta int) *historyVersionPicker {
	nextLeft := picker.leftCursor + delta
	nextRight := picker.rightCursor + delta
	if nextLeft < 0 || nextRight < 0 || nextLeft >= versionCount || nextRight >= versionCount {
		return picker
	}
	picker.leftCursor = nextLeft
	picker.rightCursor = nextRight
	return picker
}

func (m Model) RestoreHistoryPicker(projectID string, left bool, leftCursor, rightCursor int) Model {
	state, ok := m.histories[projectID]
	if !ok || leftCursor < 0 || rightCursor < 0 || leftCursor >= len(state.versions) || rightCursor >= len(state.versions) {
		return m
	}
	m.versionPicker = &historyVersionPicker{projectID: projectID, left: left, leftCursor: leftCursor, rightCursor: rightCursor}
	m.ensureHistoryPickerSpace()
	return m
}

func historyDefaultPairIndices(versions []core.RemoteConfigVersionEntry) (int, int, bool) {
	if len(versions) == 0 {
		return 0, 0, false
	}
	right := 0
	for i := range versions {
		if versions[i].Current {
			right = i
			break
		}
	}
	left := min(right+1, len(versions)-1)
	return left, right, true
}

func historyPickerBounds(versionCount, leftCursor, rightCursor int, left bool) (int, int) {
	low, high := 0, versionCount-1
	if left {
		low = max(rightCursor, 0)
	} else if leftCursor >= 0 {
		high = leftCursor
	}
	return low, max(high, low)
}

func (m Model) HistoryPickerOpen() bool { return m.versionPicker != nil }

func (m Model) HistoryPickerPosition() (int, int) {
	geometry := m.historyPickerGeometry()
	return geometry.x, geometry.y
}

func (m Model) historyPickerSize() (int, int) {
	geometry := m.historyPickerGeometry()
	return geometry.width, geometry.height
}

const historyPickerMinHeight = 10

type historyPickerLayout struct {
	x, y, width, height   int
	leftTab, rightTab     int
	leftLabel, rightLabel string
}

func (m Model) historyPickerGeometry() historyPickerLayout {
	if m.versionPicker == nil {
		return historyPickerLayout{width: 44, height: historyPickerMinHeight}
	}
	state := m.histories[m.versionPicker.projectID]
	leftLabel := historyPickerVersionLabel(state.versions, m.versionPicker.leftCursor)
	rightLabel := historyPickerVersionLabel(state.versions, m.versionPicker.rightCursor)
	leftAnchor, rightAnchor := m.historyPickerVersionAnchors(m.versionPicker.projectID, leftLabel, rightLabel)
	leftTabWidth := historyPickerTabWidth(leftLabel)
	rightTabWidth := historyPickerTabWidth(rightLabel)
	rightAnchor = max(rightAnchor, leftAnchor+leftTabWidth+1)
	hasBounds := m.width > 0
	panelRight := m.x + m.width
	if panelRight <= 0 {
		panelRight = 44
	}
	tabRight := rightAnchor + rightTabWidth
	if overflow := max(tabRight-panelRight, 0); overflow > 0 {
		leftAnchor = max(leftAnchor-overflow, 0)
		rightAnchor = max(rightAnchor-overflow, leftAnchor+leftTabWidth+1)
		tabRight = rightAnchor + rightTabWidth
	}

	versionWidth, publishedWidth, authorWidth := historyPickerNaturalColumnWidths(state.versions)
	tableWidth := versionWidth*2 + publishedWidth + authorWidth + 6
	width := max(tableWidth+4, tabRight-leftAnchor)
	if !hasBounds {
		panelRight = max(panelRight, width)
	}
	x := tabRight - width
	x = max(x, 0)
	width = min(width, max(panelRight-x, 1))

	projectLine := m.historyPickerProjectScreenLine(m.versionPicker.projectID)
	y := m.y + max(projectLine, 0)
	availableHeight := max(m.viewportHeight()-max(projectLine, 0), 1)
	desiredHeight := min(max(len(state.versions)+6, historyPickerMinHeight), 24)
	height := min(desiredHeight, availableHeight)

	leftTab := max(leftAnchor-x, 0)
	rightTab := max(rightAnchor-x, 0)
	return historyPickerLayout{
		x: x, y: y, width: width, height: height,
		leftTab: leftTab, rightTab: rightTab,
		leftLabel: leftLabel, rightLabel: rightLabel,
	}
}

func (m Model) historyPickerVersionAnchors(projectID, leftLabel, rightLabel string) (int, int) {
	if !m.historyStacked() {
		columns := m.historyColumnLayout()
		return m.x + 1 + columns.leftStart, m.x + 1 + columns.rightStart
	}
	meta := "v" + strings.TrimPrefix(leftLabel, "v")
	if at := strings.Index(meta, " "); at >= 0 {
		meta = meta[:at]
	}
	rightVersion := "v" + strings.TrimPrefix(rightLabel, "v")
	if at := strings.Index(rightVersion, " "); at >= 0 {
		rightVersion = rightVersion[:at]
	}
	projectMeta := meta + " → " + rightVersion
	left := m.x + 1 + max(m.viewportWidth()-lipgloss.Width(projectMeta), 0)
	return max(left-historyPickerTabLabelOffset, 0), max(left+lipgloss.Width(meta+" → ")-historyPickerTabLabelOffset, 0)
}

func historyPickerVersionLabel(versions []core.RemoteConfigVersionEntry, index int) string {
	if index < 0 || index >= len(versions) {
		return ""
	}
	label := "v" + versions[index].VersionNumber
	if published := formatPublished(versions[index].UpdateTime); published != "" {
		label += " " + published
	}
	return label
}

func (m Model) historyPickerProjectScreenLine(projectID string) int {
	for i, node := range m.visible {
		if node.kind == nodeProject && node.projectID == projectID {
			return m.screenLineForOffset(i, m.offset)
		}
	}
	return 0
}

func (m *Model) ensureHistoryPickerSpace() {
	if m.versionPicker == nil || len(m.visible) == 0 {
		return
	}
	projectIndex := -1
	for i, node := range m.visible {
		if node.kind == nodeProject && node.projectID == m.versionPicker.projectID {
			projectIndex = i
			break
		}
	}
	if projectIndex < 0 {
		return
	}
	minimumHeight := min(historyPickerMinHeight, m.viewportHeight())
	line := m.screenLineForOffset(projectIndex, m.offset)
	if line >= 0 && line+minimumHeight <= m.viewportHeight() {
		return
	}
	targetLine := max(m.viewportHeight()-minimumHeight, 0)
	lastOffset := m.lineIndexByNode[projectIndex]
	for offset := max(m.offset, 0); offset <= lastOffset; offset++ {
		line = m.screenLineForOffset(projectIndex, offset)
		if line >= 0 && line <= targetLine {
			m.offset = offset
			return
		}
	}
	m.offset = max(lastOffset, 0)
}

func (m Model) HistoryPickerView() string {
	if m.versionPicker == nil {
		return ""
	}
	state := m.histories[m.versionPicker.projectID]
	geometry := m.historyPickerGeometry()
	w, h := geometry.width, geometry.height
	inner, rows := max(w-2-viewutil.PopupPaddingLeft-viewutil.PopupPaddingRight, 0), max(h-6-viewutil.PopupPaddingTop, 0)
	cursor := m.versionPicker.rightCursor
	if m.versionPicker.left {
		cursor = m.versionPicker.leftCursor
	}
	start := max(min(cursor-rows/2, len(state.versions)-rows), 0)
	tabTop, tabContent, bodyTop := renderHistoryPickerTabs(geometry, m.versionPicker.left, m.historyPickerUnderlyingLine(geometry))
	lines := []string{tabTop, tabContent, bodyTop}
	for range viewutil.PopupPaddingTop {
		lines = append(lines, styles.PanelBorderActive.Render("│")+viewutil.PopupContentLine("", inner)+styles.PanelBorderActive.Render("│"))
	}
	versionWidth, publishedWidth, authorWidth := historyPickerColumnWidths(state.versions, inner)
	header := pickerVersionTableRow("", "Published", "Author", "", versionWidth, publishedWidth, authorWidth)
	header = centerHistoryPickerLine(styles.PanelMuted.Bold(true).Render(header), inner)
	lines = append(lines, styles.PanelBorderActive.Render("│")+viewutil.PopupContentLine(header, inner)+styles.PanelBorderActive.Render("│"))
	low, high := historyPickerBounds(len(state.versions), m.versionPicker.leftCursor, m.versionPicker.rightCursor, m.versionPicker.left)
	for i := range rows {
		index := start + i
		line := strings.Repeat(" ", inner)
		if index < len(state.versions) {
			v := state.versions[index]
			author := historyVersionAuthor(v)
			leftPicked := index == m.versionPicker.leftCursor
			rightPicked := index == m.versionPicker.rightCursor
			table := pickerVersionTableRow("v"+v.VersionNumber, formatPublished(v.UpdateTime), author, "v"+v.VersionNumber, versionWidth, publishedWidth, authorWidth)
			available := index >= low && index <= high
			if !available {
				table = styles.PanelTitleInactiveTab.Render(table)
			} else if index == cursor {
				table = historyPickerSelectedRow(table, true)
			} else if leftPicked || rightPicked {
				table = historyPickerSelectedRow(table, false)
			}
			table = historyPickerRowArrows(table, leftPicked, rightPicked, m.versionPicker.left, !m.versionPicker.left)
			line = centerHistoryPickerLine(table, inner)
		}
		lines = append(lines, styles.PanelBorderActive.Render("│")+viewutil.PopupContentLine(line, inner)+styles.PanelBorderActive.Render("│"))
	}
	lines = append(lines, styles.PanelBorderActive.Render("│")+viewutil.PopupContentLine(historyPickerHintView(inner), inner)+styles.PanelBorderActive.Render("│"))
	lines = append(lines, styles.PanelBorderActive.Render("╰"+strings.Repeat("─", max(w-2, 0))+"╯"))
	return strings.Join(lines, "\n")
}

func renderHistoryPickerTabs(layout historyPickerLayout, leftActive bool, underlying string) (string, string, string) {
	leftWidth := historyPickerTabWidth(layout.leftLabel)
	rightWidth := historyPickerTabWidth(layout.rightLabel)
	tops := viewutil.PadRight(ansi.Truncate(underlying, layout.width, ""), layout.width)
	gapStart := layout.leftTab + leftWidth
	gapEnd := layout.rightTab
	if gapEnd > gapStart {
		tops = replaceHistoryPickerSegment(tops, gapStart, styles.PanelBorderActive.Render(strings.Repeat("─", gapEnd-gapStart)), layout.width)
	}
	tops = replaceHistoryPickerSegment(tops, layout.rightTab, historyPickerTabTop(rightWidth), layout.width)
	tops = replaceHistoryPickerSegment(tops, layout.leftTab, historyPickerTabTop(leftWidth), layout.width)
	contents := strings.Repeat(" ", layout.width)
	contents = replaceHistoryPickerSegment(contents, layout.rightTab, historyPickerTabContent(layout.rightLabel, !leftActive), layout.width)
	contents = replaceHistoryPickerSegment(contents, layout.leftTab, historyPickerTabContent(layout.leftLabel, leftActive), layout.width)

	bodyTop := styles.PanelBorderActive.Render("╭" + strings.Repeat("─", max(layout.width-2, 0)) + "╮")
	bodyTop = replaceHistoryPickerSegment(bodyTop, layout.rightTab, historyPickerTabBottom(rightWidth, !leftActive, layout.rightTab == 0, layout.rightTab+rightWidth == layout.width), layout.width)
	bodyTop = replaceHistoryPickerSegment(bodyTop, layout.leftTab, historyPickerTabBottom(leftWidth, leftActive, layout.leftTab == 0, layout.leftTab+leftWidth == layout.width), layout.width)
	return tops, contents, bodyTop
}

func (m Model) historyPickerUnderlyingLine(layout historyPickerLayout) string {
	panelLines := strings.Split(m.ViewWithBorder(m.active, false), "\n")
	row := layout.y - m.y
	if row < 0 || row >= len(panelLines) {
		return strings.Repeat(" ", layout.width)
	}
	start := layout.x - m.x
	leftPadding := max(-start, 0)
	cutStart := max(start, 0)
	cutEnd := min(start+layout.width, m.width)
	line := strings.Repeat(" ", leftPadding)
	if cutEnd > cutStart {
		line += ansi.Cut(panelLines[row], cutStart, cutEnd)
	}
	return viewutil.PadRight(ansi.Truncate(line, layout.width, ""), layout.width)
}

func historyPickerTabTop(width int) string {
	return styles.PanelBorderActive.Render("╭" + strings.Repeat("─", max(width-2, 0)) + "╮")
}

func historyPickerTabContent(label string, active bool) string {
	if active {
		return styles.PanelBorderActive.Render("│ ") + historyPickerActiveTabStyle().Render(label) + styles.PanelBorderActive.Render(" │")
	}
	return styles.PanelBorderActive.Render("│ ") + styles.PanelTitleInactiveTab.Render(label) + styles.PanelBorderActive.Render(" │")
}

const historyPickerTabLabelOffset = 2

func historyPickerTabWidth(label string) int { return lipgloss.Width(label) + 4 }

func historyPickerActiveTabStyle() lipgloss.Style { return styles.FilterText.Bold(true) }

func historyPickerTabBottom(width int, active, atLeftEdge, atRightEdge bool) string {
	if width < 2 {
		return strings.Repeat(" ", max(width, 0))
	}
	if !active {
		left := "┴"
		if atLeftEdge {
			left = "├"
		}
		right := "┴"
		if atRightEdge {
			right = "┤"
		}
		return styles.PanelBorderActive.Render(left + strings.Repeat("─", width-2) + right)
	}
	left := "╯"
	if atLeftEdge {
		left = "│"
	}
	right := "╰"
	if atRightEdge {
		right = "│"
	}
	return styles.PanelBorderActive.Render(left + strings.Repeat(" ", width-2) + right)
}

func replaceHistoryPickerSegment(line string, start int, segment string, width int) string {
	segmentWidth := lipgloss.Width(segment)
	if start < 0 || start+segmentWidth > width {
		return line
	}
	return ansi.Cut(line, 0, start) + segment + ansi.Cut(line, start+segmentWidth, width)
}

func centerHistoryPickerLine(line string, width int) string {
	line = ansi.Truncate(line, width, "")
	left := max((width-lipgloss.Width(line))/2, 0)
	return strings.Repeat(" ", left) + viewutil.PadRight(line, width-left)
}

func historyPickerHintView(width int) string {
	items := [][2]string{
		{historyPickerKeyPair(tuiconfig.ActionLeft, tuiconfig.ActionRight), "side"},
		{historyPickerKeyPair(tuiconfig.ActionUp, tuiconfig.ActionDown), "move"},
		{historyPickerKeyPair(tuiconfig.ActionHistoryBothOlder, tuiconfig.ActionHistoryBothNewer), "pair"},
		{tuiconfig.Label(tuiconfig.BlockHistoryPicker, tuiconfig.ActionHistoryRollback), "rollback"},
		{tuiconfig.Label(tuiconfig.BlockHistoryPicker, tuiconfig.ActionReset), "reset"},
		{tuiconfig.Label(tuiconfig.BlockHistoryPicker, tuiconfig.ActionSubmit), "apply"},
		{tuiconfig.Label(tuiconfig.BlockHistoryPicker, tuiconfig.ActionCancel), "cancel"},
	}
	var hint strings.Builder
	for i, item := range items {
		if i > 0 {
			hint.WriteString(styles.PanelMuted.Render(" • "))
		}
		hint.WriteString(styles.FilterText.Render(item[0]))
		hint.WriteString(styles.PanelMuted.Render(" " + item[1]))
	}
	return pickerCell(hint.String(), width)
}

func historyPickerKeyPair(left, right tuiconfig.Action) string {
	return tuiconfig.Label(tuiconfig.BlockHistoryPicker, left) + "/" + tuiconfig.Label(tuiconfig.BlockHistoryPicker, right)
}

func historyPickerNaturalColumnWidths(versions []core.RemoteConfigVersionEntry) (int, int, int) {
	versionWidth := historyPickerArrowWidth*2 + 1
	publishedWidth, authorWidth := lipgloss.Width("Published"), lipgloss.Width("Author")
	for _, version := range versions {
		versionWidth = max(versionWidth, historyPickerArrowWidth+1+lipgloss.Width("v"+version.VersionNumber))
		publishedWidth = max(publishedWidth, lipgloss.Width(formatPublished(version.UpdateTime)))
		authorWidth = max(authorWidth, lipgloss.Width(historyVersionAuthor(version)))
	}
	return versionWidth, publishedWidth, authorWidth
}

func historyPickerColumnWidths(versions []core.RemoteConfigVersionEntry, inner int) (int, int, int) {
	versionWidth, publishedWidth, authorWidth := historyPickerNaturalColumnWidths(versions)
	overflow := versionWidth*2 + publishedWidth + authorWidth + 6 - inner
	if overflow > 0 {
		reduction := min(overflow, max(authorWidth-1, 0))
		authorWidth -= reduction
		overflow -= reduction
	}
	if overflow > 0 {
		reduction := min(overflow, max(publishedWidth-1, 0))
		publishedWidth -= reduction
	}
	return versionWidth, publishedWidth, authorWidth
}

func historyVersionAuthor(version core.RemoteConfigVersionEntry) string {
	author := strings.TrimSpace(version.UpdateUser.Email)
	if author == "" {
		author = strings.TrimSpace(version.UpdateUser.Name)
	}
	if author == "" {
		author = "—"
	}
	return author
}

const (
	historyPickerPowerlineLeftArrow  = ""
	historyPickerPowerlineRightArrow = ""
	historyPickerFallbackLeftArrow   = "◀"
	historyPickerFallbackRightArrow  = "▶"
	historyPickerArrowWidth          = 1
)

var historyPickerInactiveSelectionColor = lipgloss.Color("#343A43")

func pickerVersionTableRow(leftVersion, published, author, rightVersion string, versionWidth, publishedWidth, authorWidth int) string {
	left := strings.Repeat(" ", historyPickerArrowWidth) + " " + leftVersion
	right := rightVersion + " " + strings.Repeat(" ", historyPickerArrowWidth)
	columns := []string{pickerCell(left, versionWidth), pickerCell(published, publishedWidth), pickerCell(author, authorWidth), pickerCell(right, versionWidth)}
	return strings.Join(columns, "  ")
}

func historyPickerRowArrows(line string, left, right, leftActive, rightActive bool) string {
	width := lipgloss.Width(line)
	leftArrow, rightArrow := historyPickerArrowGlyphs(tuiconfig.PowerlineGlyphsEnabled())
	if left {
		line = historyPickerArrowStyle(leftActive).Render(leftArrow) + ansi.Cut(line, historyPickerArrowWidth, width)
	}
	if right {
		line = ansi.Cut(line, 0, width-historyPickerArrowWidth) + historyPickerArrowStyle(rightActive).Render(rightArrow)
	}
	return line
}

func historyPickerSelectedRow(line string, active bool) string {
	width := lipgloss.Width(line)
	if width <= historyPickerArrowWidth*2 {
		return line
	}
	left := ansi.Cut(line, 0, historyPickerArrowWidth)
	middle := ansi.Cut(line, historyPickerArrowWidth, width-historyPickerArrowWidth)
	right := ansi.Cut(line, width-historyPickerArrowWidth, width)
	return left + historyPickerSelectionStyle(active).Render(middle) + right
}

func historyPickerSelectionStyle(active bool) lipgloss.Style {
	if active {
		return parameterSelectionStyle()
	}
	if styles.NoColorEnabled() {
		return lipgloss.NewStyle().Underline(true)
	}
	return lipgloss.NewStyle().Background(historyPickerInactiveSelectionColor).Foreground(styles.PaletteSlateBright)
}

func historyPickerArrowStyle(active bool) lipgloss.Style {
	if styles.NoColorEnabled() {
		return lipgloss.NewStyle().Bold(active).Faint(!active)
	}
	color := historyPickerInactiveSelectionColor
	if active {
		color = styles.PaletteBlueDeep
	}
	return lipgloss.NewStyle().Foreground(color)
}

func historyPickerArrowGlyphs(powerline bool) (string, string) {
	if powerline {
		return historyPickerPowerlineLeftArrow, historyPickerPowerlineRightArrow
	}
	return historyPickerFallbackLeftArrow, historyPickerFallbackRightArrow
}

func pickerCell(value string, width int) string {
	return viewutil.PadRight(ansi.Truncate(value, width, ""), width)
}
