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
	nameStyle, idStyle := projectIdentityStyles(selected, underlined)
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
		return styles.FillSelectedLine(line, width, styles.TreeProjectSelectionStyle())
	}
	return viewutil.PadRight(line, width)
}

func (m Model) renderHistoryProjectNode(node visibleNode, project *projectState, selected, underlined bool) string {
	width := m.viewportWidth()
	history := m.histories[node.projectID]
	stacked := m.historyStacked()
	meta := ""
	switch {
	case history.loading:
		meta = "loading history…"
	case history.unavailable:
		meta = "history unavailable"
	case history.err != nil:
		meta = "history error"
	case stacked && history.previousVersion != "":
		meta = "v" + history.previousVersion + " → v" + history.currentVersion
	}
	if stacked || history.previousVersion == "" {
		leftLimit := max(width-lipgloss.Width(meta)-1, 1)
		name := viewutil.TruncatePlain(project.project.Name, leftLimit)
		id := viewutil.TruncatePlain(project.project.ProjectID, max(leftLimit-lipgloss.Width(name)-1, 0))
		nameStyle, idStyle := projectIdentityStyles(selected, underlined)
		left := nameStyle.Render(name)
		if id != "" {
			separator := " "
			if selected {
				separator = idStyle.Render(separator)
			}
			left += separator + idStyle.Render(id)
		}
		gap := max(width-lipgloss.Width(left)-lipgloss.Width(meta), 1)
		if selected {
			selection := styles.TreeProjectSelectionStyle()
			line := left + selection.Render(strings.Repeat(" ", gap)+meta)
			return styles.FillSelectedLine(line, width, selection)
		}
		line := left + projectMetaStyle.Render(strings.Repeat(" ", gap)+meta)
		line = viewutil.PadRight(line, width)
		return line
	}
	columns := m.historyColumnLayout()
	name := viewutil.TruncatePlain(project.project.Name, columns.leftBorder)
	id := viewutil.TruncatePlain(project.project.ProjectID, max(columns.leftBorder-lipgloss.Width(name)-1, 0))
	nameStyle, idStyle := projectIdentityStyles(selected, underlined)
	left := nameStyle.Render(name)
	if id != "" {
		separator := " "
		if selected {
			separator = idStyle.Render(separator)
		}
		left += separator + idStyle.Render(id)
	}
	leftPadding := strings.Repeat(" ", max(columns.leftBorder-lipgloss.Width(left), 0))
	if selected {
		left += styles.TreeProjectSelectionStyle().Render(leftPadding)
	} else {
		left += leftPadding
	}

	previous := "  v" + history.previousVersion + " " + history.previousPublished
	current := "  v" + history.currentVersion + " " + history.currentPublished
	previous = viewutil.TruncatePlain(previous, columns.leftWidth)
	current = viewutil.TruncatePlain(current, columns.rightWidth)
	previousCell := previous + strings.Repeat(" ", max(columns.leftWidth-lipgloss.Width(previous), 0))
	currentCell := current + strings.Repeat(" ", max(columns.rightWidth-lipgloss.Width(current), 0))
	if selected {
		selection := styles.TreeProjectSelectionStyle()
		line := left + selection.Render(" "+previousCell+" "+currentCell)
		return styles.FillSelectedLine(line, width, selection)
	}
	line := left + " " + projectMetaStyle.Render(previousCell) + " " + projectMetaStyle.Render(currentCell)
	line = viewutil.PadRight(line, width)
	return line
}

func projectIdentityStyles(selected, underlined bool) (lipgloss.Style, lipgloss.Style) {
	nameStyle := styles.TreeProjectName
	idStyle := styles.TreeProjectID
	if underlined {
		nameStyle = nameStyle.Underline(true)
		idStyle = idStyle.Underline(true)
	}
	if !selected {
		return nameStyle, idStyle
	}
	if styles.NoColorEnabled() {
		return lipgloss.NewStyle().Bold(true).Reverse(true), lipgloss.NewStyle().Bold(true).Reverse(true)
	}
	return nameStyle.Background(styles.PaletteError).Foreground(styles.PaletteSlateBright),
		idStyle.Background(styles.PaletteError).Foreground(styles.PaletteSlateBright)
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
		return styles.FillSelectedLine(line, width, groupSelectionStyle())
	}
	return viewutil.PadRight(line, width)
}
