package parameters

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/yumauri/fbrcm/tui/components/viewutil"
	"github.com/yumauri/fbrcm/tui/styles"
)

func (m Model) renderProjectNode(node visibleNode, selected, underlined bool) string {
	project := m.projectByID(node.projectID)
	if project == nil {
		return ""
	}
	if m.history {
		return m.renderHistoryProjectNode(node, project, selected, underlined)
	}

	width := max(m.width-2, 0)
	layout := m.layoutForProject(node.projectID)
	meta, metaStyle := m.projectMeta(project, selected), projectMetaStyle
	if project.err != nil && project.tree != nil && !project.loading && !project.verifying {
		metaStyle = styles.SecondaryTitleError
	}
	leftLimit := max(width-layout.metadataWidth-1, 1)

	name := viewutil.TruncatePlain(project.project.Name, leftLimit)
	id := viewutil.TruncatePlain(project.project.ProjectID, max(leftLimit-lipgloss.Width(name)-1, 0))
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
	line := left
	if selected && project.hasDraft {
		badge, rest := m.projectMetaSegments(project, true)
		line += metaLineStyle.Render(strings.Repeat(" ", gap))
		line += " "
		line += badge
		line += " "
		if rest != "" {
			line += metaLineStyle.Render(" " + rest)
		}
	} else {
		line += metaLineStyle.Render(strings.Repeat(" ", gap) + meta)
	}
	if selected {
		return fillSelectedLine(line, width, projectSelectionStyle())
	}
	return viewutil.PadRight(line, width)
}

func (m Model) renderHistoryProjectNode(node visibleNode, project *projectState, selected, underlined bool) string {
	width := m.viewportWidth()
	if m.historyStacked() {
		history := m.histories[node.projectID]
		meta := ""
		if history.loading {
			meta = "loading history…"
		} else if history.err != nil {
			meta = "history error"
		} else if history.previousVersion != "" {
			meta = "v" + history.previousVersion + " → v" + history.currentVersion
		}
		name := project.project.Name + " " + project.project.ProjectID
		name = viewutil.TruncatePlain(name, max(width-lipgloss.Width(meta)-1, 1))
		nameStyle := projectNameStyle
		if underlined {
			nameStyle = nameStyle.Underline(true)
		}
		line := nameStyle.Render(name) + strings.Repeat(" ", max(width-lipgloss.Width(name)-lipgloss.Width(meta), 1)) + projectMetaStyle.Render(meta)
		line = viewutil.PadRight(line, width)
		if selected {
			return projectSelectionStyle().Render(line)
		}
		return line
	}
	columns := m.historyColumnLayout()
	name := viewutil.TruncatePlain(project.project.Name, columns.leftBorder)
	id := viewutil.TruncatePlain(project.project.ProjectID, max(columns.leftBorder-lipgloss.Width(name)-1, 0))
	nameStyle := projectNameStyle
	idStyle := projectIDStyle
	if underlined {
		nameStyle = nameStyle.Underline(true)
		idStyle = idStyle.Underline(true)
	}
	left := nameStyle.Render(name)
	if id != "" {
		left += " " + idStyle.Render(id)
	}
	left = viewutil.PadRight(left, columns.leftBorder)

	history := m.histories[node.projectID]
	previous, current := "", ""
	switch {
	case history.loading:
		previous = "loading history…"
	case history.err != nil:
		previous = "history error"
	case history.previousVersion != "":
		previous = "  v" + history.previousVersion + " " + history.previousPublished
		current = "  v" + history.currentVersion + " " + history.currentPublished
	}
	previous = viewutil.TruncatePlain(previous, columns.leftWidth)
	current = viewutil.TruncatePlain(current, columns.rightWidth)
	previousCell := previous + strings.Repeat(" ", max(columns.leftWidth-lipgloss.Width(previous), 0))
	currentCell := current + strings.Repeat(" ", max(columns.rightWidth-lipgloss.Width(current), 0))
	line := left + " " + projectMetaStyle.Render(previousCell) + " " + projectMetaStyle.Render(currentCell)
	line = viewutil.PadRight(line, width)
	if selected {
		return projectSelectionStyle().Render(line)
	}
	return line
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
	return viewutil.PadRight(line, width)
}
