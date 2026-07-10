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
