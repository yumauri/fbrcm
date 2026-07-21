package projects

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/yumauri/fbrcm/tui/components/viewutil"
	"github.com/yumauri/fbrcm/tui/styles"
)

func (m Model) renderContentLines() []string {
	width := m.viewportWidth()
	if width <= 0 {
		return nil
	}

	lines := make([]string, 0, len(m.lines))
	for i, line := range m.lines {
		lines = append(lines, m.renderContentLine(i, line, width))
	}
	return lines
}

func (m Model) renderContentLine(index int, line string, width int) string {
	line = viewutil.TruncatePlain(line, max(width-1, 0))

	base := lipgloss.NewStyle()
	if index >= 0 && index < len(m.lineKinds) {
		switch m.lineKinds[index] {
		case lineKindProjectName:
			base = itemStyle
		case lineKindProjectID, lineKindMeta:
			base = metaStyle
		case lineKindPlain:
			if strings.HasPrefix(strings.TrimSpace(m.lines[index]), "Load failed:") || strings.HasPrefix(strings.TrimSpace(m.lines[index]), "Loading projects...") {
				base = itemStyle
			}
		}
	}

	projectIndex := -1
	if index >= 0 && index < len(m.lineProjects) {
		projectIndex = m.lineProjects[index]
	}
	if projectIndex >= 0 {
		return m.renderProjectLine(line, index, projectIndex, width, base)
	}

	line = " " + line
	padding := max(width-lipgloss.Width(line), 0)
	line += strings.Repeat(" ", padding)
	return base.Render(line)
}

func (m Model) renderProjectLine(line string, lineIndex, projectIndex, width int, base lipgloss.Style) string {
	state := m.projectStateStyle(projectIndex)
	normal := base.Inherit(state)
	if m.projects[projectIndex].Disabled {
		normal = normal.Foreground(styles.PanelTitleInactiveTab.GetForeground())
	}
	highlight := normal
	if !m.projects[projectIndex].Disabled {
		highlight = normal.Foreground(styles.PaletteYellow)
	}
	highlighted := indicesSet(nil)
	if lineIndex >= 0 && lineIndex < len(m.lineHighlights) {
		highlighted = indicesSet(m.lineHighlights[lineIndex])
	}

	var builder strings.Builder
	builder.WriteString(normal.Render(" "))
	for i, r := range []rune(line) {
		style := normal
		if highlighted[i] {
			style = highlight
		}
		builder.WriteString(style.Render(string(r)))
	}

	padding := max(width-lipgloss.Width(" "+line), 0)
	if padding > 0 {
		builder.WriteString(normal.Render(strings.Repeat(" ", padding)))
	}

	return builder.String()
}

func (m Model) projectStateStyle(index int) lipgloss.Style {
	_, selected := m.selected[m.projects[index].ProjectID]
	style := styles.ProjectStateStyle(index == m.cursor, selected)
	if m.projects[index].Disabled {
		style = style.Foreground(styles.PanelTitleInactiveTab.GetForeground())
	}
	return style
}

func (m Model) secondaryTitle() secondaryTitleState {
	switch {
	case m.loading:
		return secondaryTitleState{
			text:  m.spinner.View(),
			style: styles.SecondaryTitleSpinner,
		}
	case m.err != nil:
		return secondaryTitleState{
			text:  "E",
			style: styles.SecondaryTitleError,
		}
	default:
		return secondaryTitleState{
			text:  fmt.Sprintf("%d", len(m.allProjects)),
			style: styles.SecondaryTitleCount,
		}
	}
}

func (m Model) secondaryTitleText() string {
	switch {
	case m.loading:
		if len(m.spinner.Spinner.Frames) == 0 {
			return "|"
		}
		return m.spinner.Spinner.Frames[0]
	case m.err != nil:
		return "E"
	default:
		return fmt.Sprintf("%d", len(m.allProjects))
	}
}
