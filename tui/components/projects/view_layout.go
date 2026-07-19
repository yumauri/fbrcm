package projects

import (
	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/tui/messages"
)

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

// SelectOnly replaces the current project selection and moves the cursor to
// the selected project. It returns the standard selection notification used by
// downstream panels.
func (m *Model) SelectOnly(projectID string) tea.Cmd {
	found := false
	for _, project := range m.allProjects {
		if project.ProjectID == projectID {
			found = true
			break
		}
	}
	if !found {
		return nil
	}

	m.selected = map[string]struct{}{projectID: {}}
	visible := false
	for i, project := range m.projects {
		if project.ProjectID == projectID {
			m.cursor = i
			visible = true
			break
		}
	}
	if !visible {
		m.filter.ClearAndBlur()
		m.applyFilter()
		for i, project := range m.projects {
			if project.ProjectID == projectID {
				m.cursor = i
				break
			}
		}
	}
	m.syncViewport()
	return m.selectionChangedCmd()
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
