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
	tuiconfig "github.com/yumauri/fbrcm/tui/config"
	"github.com/yumauri/fbrcm/tui/styles"
)

func (m Model) ViewWithBorder(active, borderActive bool) string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}
	footer := m.filter.View(max(m.width-1, 1), active, m.conditionCount())
	body := m.bodyLines()
	panel := renderPanel(body, footer, m.width, m.height, active, borderActive)
	return m.filter.OverlayExpressionError(panel, 1)
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
		lines = append(lines, m.renderNode(m.visible[i], i == m.cursor && !m.MoveActive(), width))
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
	conditionStyle := styles.PanelText
	if strings.TrimSpace(condition.TagColor) != "" {
		conditionStyle = lipgloss.NewStyle().Foreground(styles.ConditionLipglossColor(condition.TagColor))
	}
	prefix := fmt.Sprintf(" %3d ", condition.Priority)
	if m.movingCondition(node.projectID, condition.Name) {
		prefix = conditionMovePrefix(tuiconfig.PowerlineGlyphsEnabled())
	}
	line := prefix + conditionStyle.Render("●") + " " + conditionStyle.Render(condition.Name) + styles.PanelMuted.Render(fmt.Sprintf(" · %d uses · %s", len(condition.Usages), condition.Expression))
	line = ansi.Truncate(line, width, "")
	if selected {
		selection := styles.TreeItemSelectionStyle()
		return styles.FillSelectedLine(selection.Render(ansi.Strip(line)), width, selection)
	}
	return viewutil.PadRight(line, width)
}

func conditionMovePrefix(powerline bool) string {
	glyph := "▶︎"
	if powerline {
		glyph = ""
	}
	style := lipgloss.NewStyle().Foreground(styles.TreeItemSelectionStyle().GetBackground())
	if styles.NoColorEnabled() {
		style = styles.PanelText.Bold(true)
	}
	return " " + strings.Repeat(" ", max(3-lipgloss.Width(glyph), 0)) + style.Render(glyph) + " "
}

func (m Model) renderProjectRow(project projectState, selected bool, width int) string {
	badge, rest := m.projectMetaSegments(project, selected)
	meta := joinProjectMeta(badge, rest)
	widthBadge, widthRest := m.projectMetaSegments(project, false)
	metaWidth := lipgloss.Width(joinProjectMeta(widthBadge, widthRest))
	metaStyle := styles.PanelMuted
	if project.err != nil {
		metaStyle = styles.SecondaryTitleError
	}
	leftLimit := max(width-metaWidth-1, 1)
	name := viewutil.TruncatePlain(project.project.Name, leftLimit)
	id := viewutil.TruncatePlain(project.project.ProjectID, max(leftLimit-lipgloss.Width(name)-1, 0))
	nameStyle, idStyle := styles.TreeProjectName, styles.TreeProjectID
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
			separator = idStyle.Render(separator)
		}
		left += separator + idStyle.Render(id)
	}
	gap := max(width-lipgloss.Width(left)-metaWidth, 1)
	metaLineStyle := metaStyle
	if selected {
		selection := styles.TreeProjectSelectionStyle()
		metaLineStyle = selection
	}
	line := left
	if selected && badge != "" {
		line += metaLineStyle.Render(strings.Repeat(" ", gap)) + " " + badge + " "
		if rest != "" {
			line += metaLineStyle.Render(" " + rest)
		}
	} else {
		line += metaLineStyle.Render(strings.Repeat(" ", gap) + meta)
	}
	line = ansi.Truncate(line, width, "")
	if selected {
		return styles.FillSelectedLine(line, width, styles.TreeProjectSelectionStyle())
	}
	return viewutil.PadRight(line, width)
}

func (m Model) projectMetaSegments(project projectState, selected bool) (badge, rest string) {
	if project.hasDraft {
		label := "draft"
		if project.staleDraft {
			label = "staled draft"
			if project.draftVersion != "" {
				label += " v" + project.draftVersion
			}
		}
		badge = styles.RenderDraftBadge(label, selected)
	}

	parts := make([]string, 0, 3)
	if version := project.displayVersion(); version != "" {
		parts = append(parts, "v"+version)
	}
	switch {
	case project.loading:
		parts = append(parts, m.spin.View())
	case project.err != nil:
		parts = append(parts, "error")
	case project.tree != nil:
		source := project.cacheSource
		if source == "" {
			source = project.source
		}
		if source == "draft" || source == "draft-stale" {
			source = "cache"
		}
		if status := core.ParametersStatusLabel(source, project.tree.CachedAt, true, nil); status != "" {
			parts = append(parts, status)
		}
	case project.source != "":
		parts = append(parts, project.source)
	}
	if project.tree != nil && !project.tree.CachedAt.IsZero() {
		parts = append(parts, rcdisplay.FormatLocalDateTime(project.tree.CachedAt))
	}
	return badge, strings.Join(parts, " ")
}

func joinProjectMeta(badge, rest string) string {
	switch {
	case badge != "" && rest != "":
		return badge + " " + rest
	case badge != "":
		return badge
	default:
		return rest
	}
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
