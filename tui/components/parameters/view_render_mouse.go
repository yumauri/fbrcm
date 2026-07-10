package parameters

import (
	tea "charm.land/bubbletea/v2"
)

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

func (m Model) nodeIndexAtMouse(mouse tea.Mouse) (int, bool) {
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
	if len(m.visible) == 0 {
		return 0, false
	}

	projectIndex, groupIndex, bodyStart, headerLines := m.stickyHeaderContext(m.offset)
	switch relativeY - 1 {
	case 0:
		if projectIndex >= 0 {
			return projectIndex, true
		}
	case 1:
		if headerLines >= 2 && groupIndex >= 0 {
			return groupIndex, true
		}
	}

	bodyRow := relativeY - 1 - headerLines
	if bodyRow < 0 {
		return 0, false
	}
	contentLine := bodyStart + bodyRow
	if contentLine < 0 || contentLine >= m.totalLines {
		return 0, false
	}

	nodeIndex := m.nodeIndexAtLine(contentLine)
	if nodeIndex < 0 || nodeIndex >= len(m.visible) {
		return 0, false
	}
	return nodeIndex, true
}
