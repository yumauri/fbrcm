package projects

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"fbrcm/core"
	"fbrcm/core/filter"
	"fbrcm/tui/messages"
	"fbrcm/tui/styles"
)

var (
	panelTitle = "[1] Projects"

	itemStyle = styles.PanelText
	metaStyle = styles.PanelMuted
)

type lineKind int

const (
	lineKindPlain lineKind = iota
	lineKindProjectName
	lineKindProjectID
	lineKindProjectSpacer
	lineKindMeta
)

func (m *Model) contentLines() []string {
	var lines []string
	var lineKinds []lineKind
	var lineProjects []int
	var lineHighlights [][]int
	var projectStarts []int
	var projectEnds []int

	if m.err != nil && len(m.projects) == 0 {
		lines = append(lines, fmt.Sprintf("Load failed: %v", m.err))
		lineKinds = append(lineKinds, lineKindPlain)
		lineProjects = append(lineProjects, -1)
		lineHighlights = append(lineHighlights, nil)
	} else if len(m.projects) == 0 {
		if m.loading {
			lines = append(lines, "Loading projects...")
		} else {
			lines = append(lines, "No matching projects")
		}
		lineKinds = append(lineKinds, lineKindPlain)
		lineProjects = append(lineProjects, -1)
		lineHighlights = append(lineHighlights, nil)
	} else {
		for i, project := range m.projects {
			_, nameHighlights := filter.Match(project.Name, m.filter.Value(), m.filter.Mode())
			_, idHighlights := filter.Match(project.ProjectID, m.filter.Value(), m.filter.Mode())
			idHighlights = offsetIndices(idHighlights, 1)

			projectStarts = append(projectStarts, len(lines))
			lines = append(lines, project.Name)
			lineKinds = append(lineKinds, lineKindProjectName)
			lineProjects = append(lineProjects, i)
			lineHighlights = append(lineHighlights, nameHighlights)
			lines = append(lines, " "+project.ProjectID)
			lineKinds = append(lineKinds, lineKindProjectID)
			lineProjects = append(lineProjects, i)
			lineHighlights = append(lineHighlights, idHighlights)

			projectEnds = append(projectEnds, len(lines)-1)
			if i < len(m.projects)-1 {
				lines = append(lines, "")
				lineKinds = append(lineKinds, lineKindProjectSpacer)
				lineProjects = append(lineProjects, -1)
				lineHighlights = append(lineHighlights, nil)
			}
		}
	}

	m.lines = lines
	m.lineKinds = lineKinds
	m.lineProjects = lineProjects
	m.lineHighlights = lineHighlights
	m.projectStarts = projectStarts
	m.projectEnds = projectEnds

	return m.lines
}

func (m Model) View(active bool) string {
	return renderPanel(
		m.viewport.View(),
		m.width,
		m.height,
		active,
		m.scrollbar(),
		m.secondaryTitle(),
		m.filter.View(m.width, active, len(m.projects)),
	)
}

type scrollbarState struct {
	visible    bool
	thumbStart int
	thumbEnd   int
}

type secondaryTitleState struct {
	text  string
	style lipgloss.Style
}

func renderPanel(body string, width, height int, active bool, scrollbar scrollbarState, secondary secondaryTitleState, footer []string) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	borderStyle := styles.BorderStyle(active)
	titleStyle := styles.TitleStyle(active)
	innerWidth := max(width-1, 0)
	contentHeight := max(height-2-len(footer), 0)
	topPrefixWidth := min(2, width)
	titleText := truncatePlain(" "+panelTitle+" ", max(width-topPrefixWidth-1, 0))
	titleWidth := lipgloss.Width(titleText)
	topPrefix := borderStyle.Render(strings.Repeat("─", topPrefixWidth))
	mainRendered := titleStyle.Render(titleText)
	rightMarginWidth := 1
	secondaryText := secondary.text
	secondaryRendered := ""
	secondaryWidth := 0
	if secondaryText != "" {
		secondaryText = " " + truncatePlain(secondaryText, max(width-topPrefixWidth-titleWidth-rightMarginWidth-3-1, 0)) + " "
		secondaryRendered = secondary.style.Render(secondaryText)
		secondaryWidth = lipgloss.Width(secondaryText)
	}

	gapWidth := max(width-topPrefixWidth-titleWidth-secondaryWidth-rightMarginWidth-1, 0)
	if secondaryWidth > 0 {
		gapWidth = max(gapWidth, 2)
	}
	topGap := borderStyle.Render(strings.Repeat("─", gapWidth))
	topRightFill := borderStyle.Render(strings.Repeat("─", rightMarginWidth))
	top := topPrefix + mainRendered + topGap + secondaryRendered + topRightFill + borderStyle.Render("╮")

	lines := []string{top}
	bodyLines := strings.Split(body, "\n")
	for i := range contentHeight {
		line := ""
		if i < len(bodyLines) {
			line = bodyLines[i]
		}
		padding := max(innerWidth-lipgloss.Width(line), 0)
		fill := strings.Repeat(" ", padding)
		rightEdge := borderStyle.Render("│")
		if scrollbar.visible && i >= scrollbar.thumbStart && i <= scrollbar.thumbEnd {
			rightEdge = styles.ScrollbarThumb.Render("█")
		}
		lines = append(lines, line+fill+rightEdge)
	}

	lines = append(lines, footer...)

	bottomFillWidth := max(width-1, 0)
	bottom := borderStyle.Render(strings.Repeat("─", bottomFillWidth))
	if width > 0 {
		bottom += borderStyle.Render("╯")
	}
	lines = append(lines, bottom)

	return strings.Join(lines, "\n")
}

func truncatePlain(value string, width int) string {
	if width <= 0 {
		return ""
	}

	runes := []rune(value)
	if len(runes) <= width {
		return value
	}

	return string(runes[:width])
}

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
	line = truncatePlain(line, max(width-1, 0))

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
	highlight := normal.Foreground(styles.PaletteYellow)
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

func (m Model) scrollbar() scrollbarState {
	contentHeight := m.viewportHeight()
	totalLines := len(m.lines)
	if contentHeight <= 0 || totalLines <= contentHeight {
		return scrollbarState{}
	}

	thumbHeight := max(2, (contentHeight*contentHeight)/totalLines)
	thumbHeight = min(thumbHeight, contentHeight)

	maxOffset := max(totalLines-contentHeight, 1)
	maxThumbStart := max(contentHeight-thumbHeight, 0)
	thumbStart := (m.viewport.YOffset() * maxThumbStart) / maxOffset

	return scrollbarState{
		visible:    true,
		thumbStart: thumbStart,
		thumbEnd:   min(thumbStart+thumbHeight-1, contentHeight-1),
	}
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

func (m *Model) applyFilter() {
	currentID := ""
	if m.cursor >= 0 && m.cursor < len(m.projects) {
		currentID = m.projects[m.cursor].ProjectID
	}

	query := m.filter.Value()
	m.projects = m.projects[:0]
	for _, project := range m.allProjects {
		nameMatch, _ := filter.Match(project.Name, query, m.filter.Mode())
		idMatch, _ := filter.Match(project.ProjectID, query, m.filter.Mode())
		if nameMatch || idMatch {
			m.projects = append(m.projects, project)
		}
	}

	m.cursor = 0
	if currentID != "" {
		for i, project := range m.projects {
			if project.ProjectID == currentID {
				m.cursor = i
				break
			}
		}
	}
	if len(m.projects) == 0 {
		m.cursor = 0
		return
	}
	m.cursor = max(0, min(m.cursor, len(m.projects)-1))
}

func offsetIndices(indices []int, offset int) []int {
	if len(indices) == 0 {
		return nil
	}
	shifted := make([]int, len(indices))
	for i, index := range indices {
		shifted[i] = index + offset
	}
	return shifted
}

func indicesSet(indices []int) map[int]bool {
	set := make(map[int]bool, len(indices))
	for _, index := range indices {
		set[index] = true
	}
	return set
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

func (m Model) projectStateStyle(index int) lipgloss.Style {
	_, selected := m.selected[m.projects[index].ProjectID]
	return styles.ProjectStateStyle(index == m.cursor, selected)
}

func (m *Model) moveCursor(delta int) {
	if len(m.projects) == 0 {
		return
	}

	m.cursor = max(0, min(m.cursor+delta, len(m.projects)-1))
	m.syncViewport()
}

func (m *Model) toggleCurrentSelection() {
	if len(m.projects) == 0 || m.cursor < 0 || m.cursor >= len(m.projects) {
		return
	}

	projectID := m.projects[m.cursor].ProjectID
	if _, ok := m.selected[projectID]; ok {
		delete(m.selected, projectID)
	} else {
		m.selected[projectID] = struct{}{}
	}

	// Placeholder: later selection changes will update Parameters panel.
	m.syncViewport()
}

func (m *Model) selectOnlyCurrent() {
	if len(m.projects) == 0 || m.cursor < 0 || m.cursor >= len(m.projects) {
		return
	}

	projectID := m.projects[m.cursor].ProjectID
	m.selected = map[string]struct{}{
		projectID: {},
	}

	m.syncViewport()
}

func (m *Model) selectionChangedCmd() tea.Cmd {
	projects := m.selectedProjects()
	return func() tea.Msg {
		return messages.ProjectsSelectionChangedMsg{
			Projects: projects,
		}
	}
}

func (m Model) selectedProjects() []core.Project {
	if len(m.selected) == 0 {
		return nil
	}

	projects := make([]core.Project, 0, len(m.selected))
	for _, project := range m.allProjects {
		if _, ok := m.selected[project.ProjectID]; ok {
			projects = append(projects, project)
		}
	}

	return projects
}

func (m *Model) ensureCursorVisible() {
	if len(m.projects) == 0 || m.cursor < 0 || m.cursor >= len(m.projectStarts) {
		return
	}

	top := m.projectStarts[m.cursor]
	bottom := m.projectEnds[m.cursor]
	viewportTop := m.viewport.YOffset()
	viewportBottom := viewportTop + m.viewport.Height() - 1

	switch {
	case top < viewportTop:
		m.viewport.SetYOffset(top)
	case bottom > viewportBottom:
		m.viewport.SetYOffset(bottom - m.viewport.Height() + 1)
	}
}

func (m *Model) jumpToFirst() {
	if len(m.projects) == 0 {
		return
	}

	m.cursor = 0
	m.syncViewport()
}

func (m *Model) jumpToLast() {
	if len(m.projects) == 0 {
		return
	}

	m.cursor = len(m.projects) - 1
	m.syncViewport()
}

func (m *Model) pageDown() {
	if len(m.projects) == 0 {
		return
	}

	_, bottom, ok := m.visibleProjectRange()
	if !ok {
		return
	}

	m.cursor = bottom
	m.refreshViewport()
	m.viewport.SetYOffset(m.projectStarts[m.cursor])
	m.refreshViewport()
}

func (m *Model) pageUp() {
	if len(m.projects) == 0 {
		return
	}

	top, _, ok := m.visibleProjectRange()
	if !ok {
		return
	}

	targetTopLine := max(m.projectEnds[top]-m.viewport.Height()+1, 0)
	newTop := 0
	for i := 0; i <= top; i++ {
		if m.projectStarts[i] >= targetTopLine {
			newTop = i
			break
		}
	}

	m.cursor = newTop
	m.refreshViewport()
	m.viewport.SetYOffset(m.projectStarts[newTop])
	m.refreshViewport()
}

func (m Model) visibleProjectRange() (int, int, bool) {
	if len(m.projects) == 0 {
		return 0, 0, false
	}

	topLine := m.viewport.YOffset()
	bottomLine := topLine + m.viewport.Height() - 1

	topProject := -1
	bottomProject := -1
	for i := range m.projects {
		if m.projectEnds[i] < topLine {
			continue
		}
		if m.projectStarts[i] > bottomLine {
			break
		}

		if topProject == -1 {
			topProject = i
		}
		bottomProject = i
	}

	if topProject == -1 || bottomProject == -1 {
		return 0, 0, false
	}

	return topProject, bottomProject, true
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

func (m Model) projectIndexAtMouse(mouse tea.Mouse) (int, bool) {
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

	lineIndex := m.viewport.YOffset() + relativeY - 1
	if lineIndex < 0 || lineIndex >= len(m.lineProjects) {
		return 0, false
	}

	projectIndex := m.lineProjects[lineIndex]
	if projectIndex < 0 || projectIndex >= len(m.projects) {
		return 0, false
	}

	return projectIndex, true
}
