package parameters

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/yumauri/fbrcm/tui/components/viewutil"
	"github.com/yumauri/fbrcm/tui/styles"
)

var (
	projectNameStyle = styles.PanelText.Bold(true).Foreground(styles.PaletteError)
	projectIDStyle   = styles.PanelMuted
	groupOpenStyle   = styles.PanelText.Bold(true).Foreground(styles.PaletteYellow)
	groupClosedStyle = styles.PanelMuted
	iconStyle        = styles.PanelMuted
	leafLineStyle    = iconStyle

	projectSelectedLineStyle   = lipgloss.NewStyle().Background(styles.PaletteError).Foreground(styles.PaletteSlateBright)
	groupSelectedLineStyle     = lipgloss.NewStyle().Background(styles.PaletteGold).Foreground(styles.PaletteSlateBright)
	parameterSelectedLineStyle = lipgloss.NewStyle().Background(styles.PaletteBlueDeep).Foreground(styles.PaletteSlateBright)
	valueSelectedStyle         = lipgloss.NewStyle().Background(styles.PaletteBlueDeep).Foreground(styles.PaletteSlateBright)
	draftBadgeStyle            = lipgloss.NewStyle().Background(styles.PaletteError).Foreground(styles.PaletteSlateBright).Padding(0, 1)
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

func renderPanel(body []string, width, height int, active bool, scrollbar scrollbarState, footer []string) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	borderStyle := styles.BorderStyle(active)
	titleStyle := styles.TitleStyle(active)
	innerWidth := max(width-2, 0)
	contentHeight := max(height-2-len(footer), 0)
	topPrefixWidth := min(2, width)
	titleText := viewutil.TruncatePlain(" "+panelTitle+" ", max(width-topPrefixWidth-1, 0))
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
