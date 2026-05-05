package parameters

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"fbrcm/core"
	"fbrcm/core/filter"
	"fbrcm/tui/styles"
)

var (
	projectNameStyle = styles.PanelText.Bold(true).Foreground(styles.PaletteError)
	projectIDStyle   = styles.PanelMuted
	groupOpenStyle   = styles.PanelText.Bold(true).Foreground(styles.PaletteYellow)
	groupClosedStyle = styles.PanelMuted
	iconStyle        = styles.PanelMuted
	valueStyle       = styles.PanelText.Foreground(styles.PaletteSlateBright)
	leafLineStyle    = iconStyle

	projectSelectedLineStyle   = lipgloss.NewStyle().Background(styles.PaletteError).Foreground(styles.PaletteSlateBright)
	groupSelectedLineStyle     = lipgloss.NewStyle().Background(styles.PaletteGold).Foreground(styles.PaletteSlateBright)
	parameterSelectedLineStyle = lipgloss.NewStyle().Background(styles.PaletteBlueDeep).Foreground(styles.PaletteSlateBright)
	valueSelectedStyle         = lipgloss.NewStyle().Background(styles.PaletteBlueDeep).Foreground(styles.PaletteSlateBright)
)

func projectSelectionStyle() lipgloss.Style {
	if styles.NoColorEnabled() {
		return lipgloss.NewStyle().Bold(true).Reverse(true)
	}
	return projectSelectedLineStyle
}

func groupSelectionStyle() lipgloss.Style {
	if styles.NoColorEnabled() {
		return lipgloss.NewStyle().Bold(true).Reverse(true)
	}
	return groupSelectedLineStyle
}

func parameterSelectionStyle() lipgloss.Style {
	if styles.NoColorEnabled() {
		return lipgloss.NewStyle().Reverse(true)
	}
	return parameterSelectedLineStyle
}

func valueSelectionStyle() lipgloss.Style {
	if styles.NoColorEnabled() {
		return lipgloss.NewStyle().Reverse(true)
	}
	return valueSelectedStyle
}

func fillSelectedLine(line string, width int, fillStyle lipgloss.Style) string {
	clipped := lipgloss.NewStyle().MaxWidth(width).Render(line)
	padding := max(width-lipgloss.Width(clipped), 0)
	if padding == 0 {
		return clipped
	}
	return clipped + fillStyle.Render(strings.Repeat(" ", padding))
}

func (m Model) View(active bool) string {
	projectLine, groupLine, bodyStart := m.stickyHeaderLines(m.offset)
	bodyLines := m.visibleBodyLines(bodyStart)
	lines := make([]string, 0, len(bodyLines)+2)
	if projectLine != "" {
		lines = append(lines, projectLine)
	}
	if groupLine != "" {
		lines = append(lines, groupLine)
	}
	lines = append(lines, bodyLines...)
	return renderPanel(lines, m.width, m.height, active, m.scrollbar(), m.filter.View(max(m.width-1, 1), active, m.filteredParameterCount()))
}

func (m Model) renderBody() []string {
	if len(m.visible) == 0 {
		return []string{
			"Select project in Projects panel.",
			"",
			"Selected project will appear here immediately.",
		}
	}

	height := m.contentHeight()
	if height <= 0 {
		return nil
	}

	lines := make([]string, 0, len(m.visible)+4)
	for i := 0; i < len(m.visible); i++ {
		lines = append(lines, m.renderNodeBlock(i, false)...)
	}
	return lines
}

func (m Model) visibleBodyLines(startLine int) []string {
	height := m.bodyVisibleLinesForOffset(m.offset)
	if height <= 0 {
		return nil
	}

	if len(m.visible) == 0 {
		width := m.viewportWidth()
		lines := m.renderBody()
		for i := range lines {
			lines[i] = padANSI(lipgloss.NewStyle().MaxWidth(width).Render(lines[i]), width)
		}
		for len(lines) < height {
			lines = append(lines, "")
		}
		return lines[:height]
	}

	width := m.viewportWidth()
	endLine := startLine + height
	lines := make([]string, 0, height)

	for i := 0; i < len(m.visible); i++ {
		rowStart := m.lineIndexByNode[i]
		rowHeight := m.nodeBlockLineCount(i)
		rowEnd := rowStart + rowHeight
		if rowEnd <= startLine {
			continue
		}
		if rowStart >= endLine {
			break
		}

		blockLines := m.renderNodeBlock(i, i == m.cursor)
		sliceStart := max(0, startLine-rowStart)
		sliceEnd := min(len(blockLines), endLine-rowStart)
		for _, line := range blockLines[sliceStart:sliceEnd] {
			lines = append(lines, lipgloss.NewStyle().MaxWidth(width).Render(line))
		}
		if len(lines) >= height {
			break
		}
	}

	for len(lines) < height {
		lines = append(lines, "")
	}
	return lines
}

func (m Model) stickyHeaderLines(offset int) (string, string, int) {
	projectIndex, groupIndex, bodyStart, _ := m.stickyHeaderContext(offset)
	if projectIndex < 0 || projectIndex >= len(m.visible) {
		return "", "", offset
	}

	excludedGroupKey := ""
	if groupIndex >= 0 {
		excludedGroupKey = m.visible[groupIndex].groupKey
	}
	projectUnderlined := m.projectHasHiddenContentAbove(m.visible[projectIndex].projectID, excludedGroupKey, bodyStart)
	projectLine := m.renderProjectNode(m.visible[projectIndex], projectIndex == m.cursor, projectUnderlined)

	if groupIndex < 0 {
		return projectLine, "", bodyStart
	}

	groupUnderlined := m.groupHasHiddenContentAbove(m.visible[groupIndex].projectID, m.visible[groupIndex].groupKey, bodyStart)
	groupLine := m.renderGroupNode(m.visible[groupIndex], groupIndex == m.cursor, groupUnderlined)
	return projectLine, groupLine, bodyStart
}

func (m Model) stickyHeaderContext(offset int) (projectIndex, groupIndex, bodyStart, headerLines int) {
	bodyStart = m.bodyStartForOffset(offset)
	projectIndex = m.projectNodeIndexForLine(bodyStart)
	groupIndex = m.groupNodeIndexForLine(bodyStart)
	if projectIndex >= 0 && groupIndex >= 0 && m.visible[groupIndex].projectID != m.visible[projectIndex].projectID {
		groupIndex = -1
	}
	headerLines = 1
	if groupIndex >= 0 {
		headerLines = 2
	}
	return
}

func (m Model) nodeIndexAtLine(line int) int {
	if len(m.visible) == 0 {
		return -1
	}
	if line <= 0 {
		return 0
	}
	if line >= m.totalLines {
		return len(m.visible) - 1
	}

	for i := 0; i < len(m.visible); i++ {
		start := m.lineIndexByNode[i]
		end := start + m.nodeBlockLineCount(i)
		if line >= start && line < end {
			return i
		}
	}
	return len(m.visible) - 1
}

func (m Model) projectNodeIndexFor(nodeIndex int) int {
	if nodeIndex < 0 || nodeIndex >= len(m.visible) {
		return -1
	}
	for i := min(nodeIndex, len(m.visible)-1); i >= 0; i-- {
		if m.visible[i].kind == nodeProject && m.visible[i].projectID == m.visible[nodeIndex].projectID {
			return i
		}
	}
	return -1
}

func (m Model) groupNodeIndexFor(nodeIndex int) int {
	if nodeIndex < 0 || nodeIndex >= len(m.visible) {
		return -1
	}
	groupKey := m.visible[nodeIndex].groupKey
	if groupKey == "" {
		return -1
	}
	for i := min(nodeIndex, len(m.visible)-1); i >= 0; i-- {
		if m.visible[i].kind == nodeGroup &&
			m.visible[i].projectID == m.visible[nodeIndex].projectID &&
			m.visible[i].groupKey == groupKey {
			return i
		}
	}
	return -1
}

func (m Model) projectNodeIndexForLine(line int) int {
	return m.projectNodeIndexFor(m.nodeIndexAtLine(line))
}

func (m Model) groupNodeIndexForLine(line int) int {
	nodeIndex := m.nodeIndexAtLine(line)
	if nodeIndex < 0 || nodeIndex >= len(m.visible) {
		return -1
	}

	node := m.visible[nodeIndex]
	if node.kind == nodeProject {
		for i := nodeIndex + 1; i < len(m.visible); i++ {
			if m.visible[i].projectID != node.projectID {
				break
			}
			if m.visible[i].kind == nodeGroup {
				return i
			}
		}
		return -1
	}

	return m.groupNodeIndexFor(nodeIndex)
}

func (m Model) stickyHeaderLineCount(offset int) int {
	if len(m.visible) == 0 {
		return 0
	}
	_, _, _, headerLines := m.stickyHeaderContext(offset)
	return headerLines
}

func (m Model) bodyStartForOffset(offset int) int {
	if len(m.visible) == 0 {
		return offset
	}

	bodyStart := max(offset, 0)
	projectIndex := m.projectNodeIndexForLine(offset)
	if projectIndex >= 0 && offset <= m.lineIndexByNode[projectIndex] {
		bodyStart = max(bodyStart, m.lineIndexByNode[projectIndex]+1)
	}
	groupIndex := m.groupNodeIndexForLine(offset + 1)
	if groupIndex < 0 && projectIndex >= 0 {
		groupIndex = m.groupNodeIndexForLine(offset)
	}
	if groupIndex >= 0 && offset <= m.lineIndexByNode[groupIndex] {
		bodyStart = max(bodyStart, m.lineIndexByNode[groupIndex]+1)
	}
	return bodyStart
}

func (m Model) offsetForBodyStart(target int) int {
	if m.totalLines <= 0 {
		return 0
	}

	target = max(target, 0)
	lo := 0
	hi := m.totalLines - 1
	best := hi
	for lo <= hi {
		mid := lo + (hi-lo)/2
		bodyStart := m.bodyStartForOffset(mid)
		if bodyStart >= target {
			best = mid
			hi = mid - 1
		} else {
			lo = mid + 1
		}
	}
	return best
}

func (m Model) bodyVisibleLinesForOffset(offset int) int {
	lines := m.viewportHeight() - m.stickyHeaderLineCount(offset)
	if lines < 1 {
		return 1
	}
	return lines
}

func (m Model) projectHasHiddenContentAbove(projectID, excludedGroupKey string, bodyStart int) bool {
	for i, node := range m.visible {
		if node.projectID != projectID {
			continue
		}
		if node.kind == nodeProject {
			continue
		}
		if node.kind == nodeGroup && node.groupKey == excludedGroupKey {
			continue
		}
		if m.lineIndexByNode[i] < bodyStart {
			return true
		}
	}
	return false
}

func (m Model) groupHasHiddenContentAbove(projectID, groupKey string, bodyStart int) bool {
	for i, node := range m.visible {
		if node.projectID != projectID || node.groupKey != groupKey {
			continue
		}
		if node.kind == nodeGroup {
			continue
		}
		if m.lineIndexByNode[i] < bodyStart {
			return true
		}
	}
	return false
}

func (m Model) screenLineForOffset(cursor, offset int) int {
	if len(m.visible) == 0 || cursor < 0 || cursor >= len(m.visible) {
		return -1
	}

	projectIndex, groupIndex, bodyStart, headerLines := m.stickyHeaderContext(offset)

	if cursor == projectIndex {
		return 0
	}
	if groupIndex >= 0 && cursor == groupIndex {
		return 1
	}
	return headerLines + (m.lineIndexByNode[cursor] - bodyStart)
}

func (m Model) renderNodeBlock(index int, selected bool) []string {
	if index < 0 || index >= len(m.visible) {
		return nil
	}

	node := m.visible[index]
	lines := []string{m.renderNode(node, selected)}
	if node.kind == nodeValue && m.isLastValueNode(index) {
		lines = append(lines, "")
	}
	if node.kind == nodeProject && m.isHeaderOnlyProject(index) {
		lines = append(lines, "")
	}
	if node.kind == nodeParameter && m.isLastInGroup(index) {
		lines = append(lines, "")
	}
	if node.kind == nodeGroup && !node.expanded {
		lines = append(lines, "")
	}
	if node.kind == nodeGroup && node.expanded && m.isEmptyExpandedGroup(index) {
		lines = append(lines, "")
	}

	return lines
}

func (m Model) nodeBlockLineCount(index int) int {
	if index < 0 || index >= len(m.visible) {
		return 0
	}
	count := 1
	node := m.visible[index]
	if node.kind == nodeValue && m.isLastValueNode(index) {
		count++
	}
	if node.kind == nodeProject && m.isHeaderOnlyProject(index) {
		count++
	}
	if node.kind == nodeParameter && m.isLastInGroup(index) {
		count++
	}
	if node.kind == nodeGroup && !node.expanded {
		count++
	}
	if node.kind == nodeGroup && node.expanded && m.isEmptyExpandedGroup(index) {
		count++
	}
	return count
}

func (m Model) isHeaderOnlyProject(index int) bool {
	if index < 0 || index >= len(m.visible) {
		return false
	}
	node := m.visible[index]
	if node.kind != nodeProject {
		return false
	}
	if index == len(m.visible)-1 {
		return true
	}
	next := m.visible[index+1]
	return next.projectID != node.projectID
}

func (m Model) isMouseInside(mouse tea.Mouse) bool {
	if m.width <= 0 || m.height <= 0 {
		return false
	}

	return mouse.X >= m.x && mouse.X < m.x+m.width &&
		mouse.Y >= m.y && mouse.Y < m.y+m.height
}

func (m Model) isMouseOnFilter(mouse tea.Mouse) bool {
	if !m.isMouseInside(mouse) || !m.filter.Visible() {
		return false
	}

	relativeY := mouse.Y - m.y
	filterTop := m.height - 1 - m.filter.Height()
	return relativeY >= filterTop && relativeY < m.height-1 &&
		mouse.X >= m.x && mouse.X < m.x+m.width-1
}

func (m Model) nodeIndexAtMouse(mouse tea.Mouse) (int, bool) {
	if !m.isMouseInside(mouse) {
		return 0, false
	}

	relativeY := mouse.Y - m.y
	if relativeY <= 0 || relativeY >= m.height-1 {
		return 0, false
	}
	if mouse.X >= m.x+m.width-1 {
		return 0, false
	}
	if len(m.visible) == 0 {
		return 0, false
	}

	projectIndex, groupIndex, bodyStart, headerLines := m.stickyHeaderContext(m.offset)
	switch relativeY - 1 {
	case 0:
		if projectIndex >= 0 {
			return projectIndex, true
		}
	case 1:
		if headerLines >= 2 && groupIndex >= 0 {
			return groupIndex, true
		}
	}

	bodyRow := relativeY - 1 - headerLines
	if bodyRow < 0 {
		return 0, false
	}
	contentLine := bodyStart + bodyRow
	if contentLine < 0 || contentLine >= m.totalLines {
		return 0, false
	}

	nodeIndex := m.nodeIndexAtLine(contentLine)
	if nodeIndex < 0 || nodeIndex >= len(m.visible) {
		return 0, false
	}
	return nodeIndex, true
}

func (m Model) renderNode(node visibleNode, selected bool) string {
	switch node.kind {
	case nodeProject:
		return m.renderProjectNode(node, selected, false)
	case nodeGroup:
		return m.renderGroupNode(node, selected, false)
	case nodeParameter:
		return m.renderParameterNode(node, selected)
	case nodeValue:
		return m.renderValueNode(node, selected)
	default:
		return ""
	}
}

func (m Model) renderProjectNode(node visibleNode, selected, underlined bool) string {
	project := m.projectByID(node.projectID)
	if project == nil {
		return ""
	}

	width := max(m.width-2, 0)
	layout := m.layoutForProject(node.projectID)
	meta, metaStyle := m.projectMeta(project), projectMetaStyle
	if project.err != nil && project.tree != nil && !project.loading && !project.verifying {
		metaStyle = styles.SecondaryTitleError
	}
	leftLimit := max(width-layout.metadataWidth-1, 1)

	name := truncatePlain(project.project.Name, leftLimit)
	id := truncatePlain(project.project.ProjectID, max(leftLimit-lipgloss.Width(name)-1, 0))
	nameStyle := projectNameStyle
	idStyle := projectIDStyle
	if underlined {
		nameStyle = nameStyle.Underline(true)
		idStyle = idStyle.Underline(true)
	}
	if selected {
		if styles.NoColorEnabled() {
			nameStyle = lipgloss.NewStyle().Bold(true).Reverse(true)
			idStyle = lipgloss.NewStyle().Bold(true).Reverse(true)
		} else {
			nameStyle = nameStyle.Background(styles.PaletteError).Foreground(styles.PaletteSlateBright)
			idStyle = idStyle.Background(styles.PaletteError).Foreground(styles.PaletteSlateBright)
		}
	}
	left := nameStyle.Render(name)
	if id != "" {
		separator := " "
		if selected {
			separator = idStyle.Render(" ")
		}
		left += separator + idStyle.Render(id)
	}
	leftWidth := lipgloss.Width(left)

	gap := max(width-leftWidth-layout.metadataWidth, 1)
	metaLineStyle := metaStyle
	if selected {
		if styles.NoColorEnabled() {
			metaLineStyle = lipgloss.NewStyle().Bold(true).Reverse(true)
		} else {
			metaLineStyle = metaLineStyle.Background(styles.PaletteError).Foreground(styles.PaletteSlateBright)
		}
	}
	line := left + metaLineStyle.Render(strings.Repeat(" ", gap)+meta)
	if selected {
		return fillSelectedLine(line, width, projectSelectionStyle())
	}
	return padANSI(line, width)
}

func (m Model) renderGroupNode(node visibleNode, selected, underlined bool) string {
	width := max(m.width-2, 0)
	arrow := "▾"
	style := groupOpenStyle
	if !node.expanded {
		arrow = "▸"
		style = groupClosedStyle
	}
	if underlined {
		style = style.Underline(true)
	}
	if selected {
		if styles.NoColorEnabled() {
			style = lipgloss.NewStyle().Bold(true).Reverse(true)
		} else {
			style = style.Background(styles.PaletteGold).Foreground(styles.PaletteSlateBright)
		}
	}

	line := arrow + " " + style.Render(node.label)
	if selected {
		prefixStyle := groupSelectionStyle()
		prefix := prefixStyle.Render(arrow + " ")
		line = prefix + style.Render(node.label)
		return fillSelectedLine(line, width, groupSelectionStyle())
	}
	return padANSI(line, width)
}

func (m Model) renderParameterNode(node visibleNode, selected bool) string {
	width := max(m.width-2, 0)
	layout := m.layoutForProject(node.projectID)
	param := m.parameterByKey(node.projectID, node.groupKey, node.paramKey)
	if param == nil {
		return strings.Repeat(" ", width)
	}

	namePad := strings.Repeat(" ", max(layout.nameWidth-lipgloss.Width(param.Key), 0))
	style := parameterStyle
	if selected {
		if styles.NoColorEnabled() {
			style = lipgloss.NewStyle().Reverse(true)
		} else {
			style = style.Background(styles.PaletteBlueDeep).Foreground(styles.PaletteSlateBright)
		}
	}
	if isDeprecatedDescription(param.Description) {
		style = style.Strikethrough(true).Faint(true)
	}

	if node.expanded {
		left := lipgloss.NewStyle().Render("  ") + m.renderHighlightedParameterKey(param.Key, style, selected)
		if param.Description != "" {
			descStyle := descriptionStyle
			if selected {
				if styles.NoColorEnabled() {
					descStyle = lipgloss.NewStyle().Reverse(true).Italic(true)
				} else {
					descStyle = descStyle.Background(styles.PaletteBlueDeep).Foreground(styles.PaletteSlateBright)
				}
			}
			left += "  "
			left += descStyle.Render(param.Description)
		}
		if selected {
			prefix := parameterSelectionStyle().Render("  ")
			left = prefix + m.renderHighlightedParameterKey(param.Key, style, selected)
			if param.Description != "" {
				var descStyle lipgloss.Style
				if styles.NoColorEnabled() {
					descStyle = lipgloss.NewStyle().Reverse(true).Italic(true)
				} else {
					descStyle = descriptionStyle.Background(styles.PaletteBlueDeep).Foreground(styles.PaletteSlateBright)
				}
				left += parameterSelectionStyle().Render("  ")
				left += descStyle.Render(param.Description)
			}
			return fillSelectedLine(left, width, parameterSelectionStyle())
		}
		return padANSI(left, width)
	}

	icon := "╌"
	if len(param.Values) > 1 {
		icon = "⌥"
	}
	prefixStyle := lipgloss.NewStyle()
	iconLineStyle := iconStyle
	valueLineStyle := valueStyle
	separatorLineStyle := parameterSeparatorStyle
	if selected {
		if styles.NoColorEnabled() {
			prefixStyle = parameterSelectionStyle()
			iconLineStyle = parameterSelectionStyle()
			valueLineStyle = parameterSelectionStyle()
			separatorLineStyle = parameterSelectionStyle()
		} else {
			prefixStyle = prefixStyle.Background(styles.PaletteBlueDeep)
			iconLineStyle = iconLineStyle.Background(styles.PaletteBlueDeep).Foreground(styles.PaletteSlateBright)
			valueLineStyle = valueLineStyle.Background(styles.PaletteBlueDeep).Foreground(styles.PaletteSlateBright)
			separatorLineStyle = separatorLineStyle.Background(styles.PaletteBlueDeep).Foreground(styles.PaletteSlateBright)
		}
	}
	left := prefixStyle.Render("  ") + m.renderHighlightedParameterKey(param.Key, style, selected) + prefixStyle.Render(namePad)
	left += prefixStyle.Render(strings.Repeat(" ", 2)) + iconLineStyle.Render(icon)
	left += prefixStyle.Render(" ")
	line := left + m.renderCollapsedParameterValues(param.Values, valueLineStyle, separatorLineStyle, selected)
	if selected {
		return fillSelectedLine(line, width, parameterSelectionStyle())
	}
	return padANSI(line, width)
}

func isDeprecatedDescription(description string) bool {
	return deprecatedDescriptionPattern.MatchString(description)
}

func (m Model) renderValueNode(node visibleNode, selected bool) string {
	width := max(m.width-2, 0)
	layout := m.layoutForProject(node.projectID)
	param := m.parameterByKey(node.projectID, node.groupKey, node.paramKey)
	if param == nil || node.valueIdx < 0 || node.valueIdx >= len(param.Values) {
		if node.label != "" {
			return padANSI(parameterValueStyle.Render(node.label), width)
		}
		return strings.Repeat(" ", width)
	}

	value := param.Values[node.valueIdx]
	labelStyle := m.conditionStyle(value.Color)
	if value.Label == "default" {
		labelStyle = conditionDefaultStyle
	}

	conditionLabel := displayConditionLabel(value.Label)
	conditionWidth := parameterConditionWidth(param)
	leafOffset := 1
	if len(param.Values) == 1 {
		leafOffset = 2
	}
	leafOffset++
	leafValueStart := layout.valueStart + leafOffset
	labelStart := max(leafValueStart-conditionWidth-4, layout.paramStart+2)
	connector := m.valueConnector(node, param)
	tree := leafLineStyle.Render(branchGlyph(layout.paramStart, labelStart, connector))
	label := conditionLabel
	fillerWidth := max(leafValueStart-labelStart-lipgloss.Width(label)-3, 1)
	filler := strings.Repeat("╌", fillerWidth)
	valueRendered := m.renderParameterValue(value, selected)
	line := tree + " " + labelStyle.Render(label) + leafLineStyle.Render(" "+filler+" ") + valueRendered
	return padANSI(line, width)
}

func (m Model) renderCollapsedParameterValues(values []core.ParametersValue, valueStyle, separatorStyle lipgloss.Style, selected bool) string {
	parts := make([]string, 0, max(0, len(values)*2-1))
	for i, value := range values {
		if i > 0 {
			parts = append(parts, separatorStyle.Render(" / "))
		}
		parts = append(parts, m.renderParameterValueWithBase(value, valueStyle, selected))
	}
	return strings.Join(parts, "")
}

func (m Model) renderParameterValue(value core.ParametersValue, selected bool) string {
	return m.renderParameterValueWithBase(value, valueStyle, selected)
}

func (m Model) renderParameterValueWithBase(value core.ParametersValue, baseStyle lipgloss.Style, selected bool) string {
	if value.Empty {
		style := parameterEmptyValue
		if selected {
			if styles.NoColorEnabled() {
				style = lipgloss.NewStyle().Reverse(true).Italic(true)
			} else {
				style = style.Background(styles.PaletteBlueDeep).Foreground(styles.PaletteSlateBright)
			}
		}
		return style.Render(value.Value)
	}
	if selected {
		return valueSelectionStyle().Render(value.Value)
	}
	return baseStyle.Render(value.Value)
}

func (m Model) renderHighlightedParameterKey(text string, baseStyle lipgloss.Style, selected bool) string {
	query := m.filter.Value()
	if query == "" {
		return baseStyle.Render(text)
	}

	_, indices := filter.Match(text, query, m.filter.Mode())
	if len(indices) == 0 {
		return baseStyle.Render(text)
	}

	highlighted := indicesSet(indices)
	highlightStyle := baseStyle.Foreground(styles.PaletteYellow)
	if selected {
		highlightStyle = baseStyle.Foreground(styles.PaletteYellow)
	}

	var builder strings.Builder
	for i, r := range []rune(text) {
		style := baseStyle
		if highlighted[i] {
			style = highlightStyle
		}
		builder.WriteString(style.Render(string(r)))
	}
	return builder.String()
}

func (m Model) valueConnector(node visibleNode, param *core.ParametersEntry) string {
	if param == nil {
		return "last"
	}
	if len(param.Values) == 1 {
		return "single"
	}
	if node.valueIdx >= len(param.Values)-1 {
		return "last"
	}
	if node.valueIdx == 0 {
		return "first"
	}
	return "mid"
}

func branchGlyph(paramStart, labelStart int, connector string) string {
	totalWidth := max(labelStart-1, 1)
	if totalWidth <= paramStart {
		return strings.Repeat(" ", max(totalWidth-2, 0)) + "╰╌"
	}

	if connector == "first" {
		return strings.Repeat(" ", paramStart) + "╰" + strings.Repeat("╌", max(totalWidth-paramStart-3, 0)) + "┬╌"
	}
	if connector == "single" {
		return strings.Repeat(" ", paramStart) + "╰" + strings.Repeat("╌", max(totalWidth-paramStart-2, 0))
	}

	prefixWidth := max(totalWidth-2, 0)
	switch connector {
	case "mid":
		return strings.Repeat(" ", prefixWidth) + "├╌"
	default:
		return strings.Repeat(" ", prefixWidth) + "╰╌"
	}
}

func parameterConditionWidth(param *core.ParametersEntry) int {
	width := lipgloss.Width("Default value")
	if param == nil {
		return width
	}
	for _, value := range param.Values {
		width = max(width, lipgloss.Width(displayConditionLabel(value.Label)))
	}
	return width
}

func (m Model) isLastValueNode(index int) bool {
	if index < 0 || index >= len(m.visible) {
		return false
	}
	node := m.visible[index]
	if node.kind != nodeValue {
		return false
	}
	if index == len(m.visible)-1 {
		return true
	}
	next := m.visible[index+1]
	return next.kind != nodeValue || next.paramKey != node.paramKey || next.groupKey != node.groupKey || next.projectID != node.projectID
}

func (m Model) isLastInGroup(index int) bool {
	if index < 0 || index >= len(m.visible) {
		return false
	}
	node := m.visible[index]
	if node.kind != nodeParameter {
		return false
	}
	if index == len(m.visible)-1 {
		return true
	}
	next := m.visible[index+1]
	if next.projectID != node.projectID || next.groupKey != node.groupKey {
		return true
	}
	return next.kind == nodeGroup || next.kind == nodeProject
}

func (m Model) isEmptyExpandedGroup(index int) bool {
	if index < 0 || index >= len(m.visible) {
		return false
	}
	node := m.visible[index]
	if node.kind != nodeGroup || !node.expanded {
		return false
	}
	if index == len(m.visible)-1 {
		return true
	}
	next := m.visible[index+1]
	if next.projectID != node.projectID || next.groupKey != node.groupKey {
		return true
	}
	return next.kind == nodeGroup || next.kind == nodeProject
}

func (m Model) conditionStyle(color string) lipgloss.Style {
	return styles.PanelText.Foreground(styles.ConditionLipglossColor(color))
}

func renderPanel(body []string, width, height int, active bool, scrollbar scrollbarState, footer []string) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	borderStyle := styles.BorderStyle(active)
	titleStyle := styles.TitleStyle(active)
	innerWidth := max(width-2, 0)
	contentHeight := max(height-2-len(footer), 0)
	topPrefixWidth := min(2, width)
	titleText := truncatePlain(" "+panelTitle+" ", max(width-topPrefixWidth-1, 0))
	titleWidth := lipgloss.Width(titleText)
	topPrefix := borderStyle.Render("╭" + strings.Repeat("─", max(topPrefixWidth-1, 0)))
	topFillWidth := max(width-topPrefixWidth-titleWidth-1, 0)
	topFill := borderStyle.Render(strings.Repeat("─", topFillWidth))
	topRight := ""
	if width > topPrefixWidth+titleWidth {
		topRight = borderStyle.Render("╮")
	}

	lines := []string{topPrefix + titleStyle.Render(titleText) + topFill + topRight}
	for i := range contentHeight {
		line := ""
		if i < len(body) {
			line = body[i]
		}
		rightEdge := borderStyle.Render("│")
		if scrollbar.visible && i >= scrollbar.thumbStart && i <= scrollbar.thumbEnd {
			rightEdge = styles.ScrollbarThumb.Render("█")
		}
		if line == "" {
			line = strings.Repeat(" ", innerWidth)
		}
		lines = append(lines,
			borderStyle.Render("│")+line+rightEdge,
		)
	}

	for i, footerLine := range footer {
		left := "│"
		if i == 0 {
			left = "├"
		}
		lines = append(lines, borderStyle.Render(left)+footerLine)
	}

	bottom := borderStyle.Render("╰" + strings.Repeat("─", max(width-2, 0)) + "╯")
	lines = append(lines, bottom)

	return strings.Join(lines, "\n")
}

type scrollbarState struct {
	visible    bool
	thumbStart int
	thumbEnd   int
}

func (m Model) scrollbar() scrollbarState {
	contentHeight := m.viewportHeight()
	totalLines := m.totalLines
	if contentHeight <= 0 || totalLines <= contentHeight {
		return scrollbarState{}
	}

	thumbHeight := max(2, (contentHeight*contentHeight)/totalLines)
	thumbHeight = min(thumbHeight, contentHeight)

	maxOffset := max(totalLines-contentHeight, 1)
	maxThumbStart := max(contentHeight-thumbHeight, 0)
	thumbStart := (m.offset * maxThumbStart) / maxOffset

	return scrollbarState{
		visible:    true,
		thumbStart: thumbStart,
		thumbEnd:   min(thumbStart+thumbHeight-1, contentHeight-1),
	}
}

func padANSI(value string, width int) string {
	plainWidth := lipgloss.Width(value)
	return value + strings.Repeat(" ", max(width-plainWidth, 0))
}

func indicesSet(indices []int) map[int]bool {
	set := make(map[int]bool, len(indices))
	for _, index := range indices {
		set[index] = true
	}
	return set
}
