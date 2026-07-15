package conditions

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/core"
	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
	"github.com/yumauri/fbrcm/tui/components/viewutil"
	"github.com/yumauri/fbrcm/tui/components/workspaceheader"
	"github.com/yumauri/fbrcm/tui/styles"
)

func (m Model) ViewWithBorder(active, borderActive bool) string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}
	footer := m.filter.View(max(m.width-1, 1), active, m.conditionCount())
	body := m.bodyLines()
	return renderPanel(body, footer, m.width, m.height, active, borderActive)
}

func (m Model) bodyLines() []string {
	height := m.contentHeight()
	width := max(m.width-2, 1)
	if len(m.visible) == 0 {
		lines := []string{"Select project in Projects panel.", "", "Selected project conditions will appear here."}
		for i := range lines {
			lines[i] = viewutil.PadRight(ansi.Truncate(lines[i], width, ""), width)
		}
		return lines
	}
	end := min(m.offset+height, len(m.visible))
	lines := make([]string, 0, height)
	for i := m.offset; i < end; i++ {
		lines = append(lines, m.renderNode(m.visible[i], i == m.cursor, width))
	}
	return lines
}

func (m Model) renderNode(node visibleNode, selected bool, width int) string {
	if node.kind == nodeGap {
		return strings.Repeat(" ", width)
	}
	projectIdx, ok := m.projectIndex[node.projectID]
	if !ok {
		return strings.Repeat(" ", width)
	}
	project := m.projects[projectIdx]
	if node.kind == nodeProject {
		return m.renderProjectRow(project, selected, width)
	}
	if project.tree == nil || node.conditionIndex < 0 || node.conditionIndex >= len(project.tree.Conditions) {
		return strings.Repeat(" ", width)
	}
	condition := project.tree.Conditions[node.conditionIndex]
	conditionStyle := lipgloss.NewStyle().Foreground(styles.ConditionLipglossColor(condition.TagColor))
	line := fmt.Sprintf(" %3d ", condition.Priority) + conditionStyle.Render("●") + " " + conditionStyle.Render(condition.Name) + styles.PanelMuted.Render(fmt.Sprintf(" · %d uses · %s", len(condition.Usages), condition.Expression))
	line = ansi.Truncate(line, width, "")
	if selected {
		selection := styles.TreeItemSelectionStyle()
		return styles.FillSelectedLine(selection.Render(ansi.Strip(line)), width, selection)
	}
	return viewutil.PadRight(line, width)
}

func (m Model) renderProjectRow(project projectState, selected bool, width int) string {
	meta, metaStyle := m.projectMeta(project), styles.PanelMuted
	if project.err != nil {
		metaStyle = styles.SecondaryTitleError
	}
	metaWidth := lipgloss.Width(meta)
	leftLimit := max(width-metaWidth-1, 1)
	name := viewutil.TruncatePlain(project.project.Name, leftLimit)
	id := viewutil.TruncatePlain(project.project.ProjectID, max(leftLimit-lipgloss.Width(name)-1, 0))
	left := styles.TreeProjectName.Render(name)
	if id != "" {
		left += " " + styles.TreeProjectID.Render(id)
	}
	gap := max(width-lipgloss.Width(left)-metaWidth, 1)
	line := ansi.Truncate(left+metaStyle.Render(strings.Repeat(" ", gap)+meta), width, "")
	if selected {
		selection := styles.TreeProjectSelectionStyle()
		return styles.FillSelectedLine(selection.Render(ansi.Strip(line)), width, selection)
	}
	return viewutil.PadRight(line, width)
}

func (m Model) projectMeta(project projectState) string {
	parts := make([]string, 0, 2)
	if project.tree != nil && project.tree.Version != "" {
		parts = append(parts, "v"+project.tree.Version)
	}
	switch {
	case project.loading:
		parts = append(parts, m.spin.View())
	case project.err != nil:
		parts = append(parts, "error")
	case project.source == "draft":
		parts = append(parts, "draft")
	case project.source == "draft-stale":
		parts = append(parts, "staled draft")
	case project.tree != nil:
		if status := core.ParametersStatusLabel(project.source, project.tree.CachedAt, true, nil); status != "" {
			parts = append(parts, status)
		}
	case project.source != "":
		parts = append(parts, project.source)
	}
	if project.tree != nil && !project.tree.CachedAt.IsZero() {
		parts = append(parts, rcdisplay.FormatLocalDateTime(project.tree.CachedAt))
	}
	return strings.Join(parts, " ")
}

func (m Model) conditionCount() int {
	count := 0
	for _, node := range m.visible {
		if node.kind == nodeCondition {
			count++
		}
	}
	return count
}

func renderPanel(body, footer []string, width, height int, active, borderActive bool) string {
	borderStyle := styles.BorderStyle(borderActive)
	contentHeight := max(height-2-len(footer), 0)
	topPrefixWidth := min(2, width)
	titles, titleWidth := workspaceheader.Render(width, 1, active, borderStyle)
	top := borderStyle.Render("╭"+strings.Repeat("─", max(topPrefixWidth-1, 0))) + titles + borderStyle.Render(strings.Repeat("─", max(width-topPrefixWidth-titleWidth-1, 0))+"╮")
	lines := []string{top}
	innerWidth := max(width-2, 0)
	for i := range contentHeight {
		line := ""
		if i < len(body) {
			line = body[i]
		}
		line = viewutil.PadRight(ansi.Truncate(line, innerWidth, ""), innerWidth)
		lines = append(lines, borderStyle.Render("│")+line+borderStyle.Render("│"))
	}
	for i, line := range footer {
		left := "│"
		if i == 0 {
			left = "├"
		}
		lines = append(lines, borderStyle.Render(left)+line)
	}
	lines = append(lines, borderStyle.Render("╰"+strings.Repeat("─", max(width-2, 0))+"╯"))
	return strings.Join(lines, "\n")
}
