package parameters

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/tui/components/workspaceheader"
	"github.com/yumauri/fbrcm/tui/styles"
)

var (
	groupOpenStyle   = styles.PanelText.Bold(true).Foreground(styles.PaletteYellow)
	groupClosedStyle = styles.PanelMuted
	iconStyle        = lipgloss.NewStyle().Foreground(styles.PaletteSlateDark)
	leafLineStyle    = iconStyle

	groupSelectedLineStyle = lipgloss.NewStyle().Background(styles.PaletteGold).Foreground(styles.PaletteSlateBright)
	valueSelectedStyle     = lipgloss.NewStyle().Background(styles.PaletteBlueDeep).Foreground(styles.PaletteSlateBright)
	draftBadgeStyle        = lipgloss.NewStyle().Background(styles.PaletteError).Foreground(styles.PaletteSlateBright).Padding(0, 1)
)

func groupSelectionStyle() lipgloss.Style {
	if styles.NoColorEnabled() {
		return lipgloss.NewStyle().Bold(true).Reverse(true)
	}
	return groupSelectedLineStyle
}

func parameterSelectionStyle() lipgloss.Style {
	return styles.TreeItemSelectionStyle()
}

func valueSelectionStyle() lipgloss.Style {
	if styles.NoColorEnabled() {
		return lipgloss.NewStyle().Reverse(true)
	}
	return valueSelectedStyle
}

func (m Model) renderNodeBlock(index int, selected bool) []string {
	if index < 0 || index >= len(m.visible) {
		return nil
	}

	node := m.visible[index]
	lines := []string{m.renderNode(node, selected)}
	if m.history && m.historyStacked() {
		lines = m.renderHistoryStackedNode(node, selected)
	}
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
	if m.history && m.historyStacked() {
		if node.kind == nodeValue {
			count = m.historyStackedValueLineCount(node)
		}
	}
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

func renderPanel(body []string, width, height int, active, borderActive, history bool, historyModes []string, historyBorders []int, scrollbar scrollbarState, footer []string) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	borderStyle := styles.BorderStyle(borderActive)
	innerWidth := max(width-2, 0)
	contentHeight := max(height-2-len(footer), 0)
	topPrefixWidth := min(2, width)
	selectedTab := 0
	if history {
		selectedTab = 2
	}
	titleRendered, titleWidth := workspaceheader.Render(width, selectedTab, active, borderStyle)
	topPrefix := borderStyle.Render("╭" + strings.Repeat("─", max(topPrefixWidth-1, 0)))
	remainingWidth := max(width-topPrefixWidth-titleWidth-1, 0)
	mode := ""
	modeWidth := 0
	for _, candidate := range historyModes {
		candidateWidth := lipgloss.Width(candidate)
		if candidateWidth+1 <= remainingWidth {
			modeStyle := styles.PanelTitleInactiveTab
			if active && borderActive {
				modeStyle = styles.PanelTitle
			}
			mode = modeStyle.Render(candidate) + borderStyle.Render("─")
			modeWidth = candidateWidth + 1
			break
		}
	}
	topFillWidth := max(remainingWidth-modeWidth, 0)
	topFill := borderStyle.Render(strings.Repeat("─", topFillWidth))
	topRight := ""
	if width > topPrefixWidth+titleWidth {
		topRight = borderStyle.Render("╮")
	}

	top := topPrefix + titleRendered + topFill + mode + topRight
	if history {
		top = replacePanelGridGlyphs(top, width, historyBorders, "┬", borderStyle)
	}
	lines := []string{top}
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
			if history {
				line = replaceInnerGridGlyphs(line, innerWidth, historyBorders, borderStyle)
			}
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
		footerRendered := borderStyle.Render(left) + footerLine
		if history {
			glyph := "│"
			if i == 0 {
				glyph = "┼"
			}
			footerRendered = replacePanelGridGlyphs(footerRendered, width, historyBorders, glyph, borderStyle)
		}
		lines = append(lines, footerRendered)
	}

	bottom := borderStyle.Render("╰" + strings.Repeat("─", max(width-2, 0)) + "╯")
	if history {
		bottom = replacePanelGridGlyphs(bottom, width, historyBorders, "┴", borderStyle)
	}
	lines = append(lines, bottom)

	return strings.Join(lines, "\n")
}

func replaceInnerGridGlyphs(line string, width int, positions []int, style lipgloss.Style) string {
	for _, position := range positions {
		if position < 0 || position >= width {
			continue
		}
		line = ansi.Cut(line, 0, position) + style.Render("│") + ansi.Cut(line, position+1, width)
	}
	return line
}

func replacePanelGridGlyphs(line string, width int, positions []int, glyph string, style lipgloss.Style) string {
	shifted := make([]int, 0, len(positions))
	for _, position := range positions {
		shifted = append(shifted, position+1)
	}
	for _, position := range shifted {
		line = ansi.Cut(line, 0, position) + style.Render(glyph) + ansi.Cut(line, position+1, width)
	}
	return line
}
