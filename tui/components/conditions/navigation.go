package conditions

func (m *Model) moveCursor(delta int) {
	if len(m.visible) == 0 {
		return
	}
	direction := 1
	if delta < 0 {
		direction = -1
	}
	m.cursor = m.nearestSelectableIndex(min(max(m.cursor+delta, 0), len(m.visible)-1), direction)
	m.ensureCursorVisible()
}

func (m Model) nearestSelectableIndex(index, direction int) int {
	if len(m.visible) == 0 {
		return 0
	}
	index = min(max(index, 0), len(m.visible)-1)
	if m.visible[index].kind != nodeGap {
		return index
	}
	for next := index + direction; next >= 0 && next < len(m.visible); next += direction {
		if m.visible[next].kind != nodeGap {
			return next
		}
	}
	for next := index - direction; next >= 0 && next < len(m.visible); next -= direction {
		if m.visible[next].kind != nodeGap {
			return next
		}
	}
	return index
}

func (m *Model) moveFirst() {
	m.cursor = m.nearestSelectableIndex(0, 1)
	m.ensureCursorVisible()
}

func (m *Model) moveLast() {
	m.cursor = m.nearestSelectableIndex(len(m.visible)-1, -1)
	m.ensureCursorVisible()
}

func (m Model) contentHeight() int {
	return max(m.height-2-m.filter.Height(), 1)
}

func (m *Model) ensureCursorVisible() {
	height := m.contentHeight()
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+height {
		m.offset = m.cursor - height + 1
	}
	maxOffset := max(len(m.visible)-height, 0)
	m.offset = min(max(m.offset, 0), maxOffset)
}

func (m Model) isMouseInside(x, y int) bool {
	return x >= m.x && x < m.x+m.width && y >= m.y && y < m.y+m.height
}

func (m Model) nodeAtMouse(x, y int) (int, bool) {
	if !m.isMouseInside(x, y) {
		return 0, false
	}
	row := y - m.y - 1
	if row < 0 || row >= m.contentHeight() {
		return 0, false
	}
	index := m.offset + row
	return index, index >= 0 && index < len(m.visible) && m.visible[index].kind != nodeGap
}
