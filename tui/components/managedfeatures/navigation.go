package managedfeatures

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

func (m *Model) movePage(deltaRows int) {
	if len(m.visible) == 0 || deltaRows == 0 {
		return
	}
	direction := 1
	if deltaRows < 0 {
		direction = -1
	}
	targetRow := min(max(m.nodeStartRow(m.cursor)+deltaRows, 0), max(m.visibleHeight()-1, 0))
	index, ok := m.nodeIndexAtRow(targetRow)
	if !ok {
		return
	}
	index = m.nearestSelectableIndex(index, direction)
	if index == m.cursor {
		m.moveCursor(direction)
		return
	}
	m.cursor = index
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
	if len(m.visible) == 0 {
		m.offset = 0
		return
	}
	height := m.contentHeight()
	start := m.nodeStartRow(m.cursor)
	nodeHeight := m.nodeHeight(m.visible[m.cursor])
	end := start + nodeHeight
	if start < m.offset {
		m.offset = start
	}
	if nodeHeight >= height {
		m.offset = start
	} else if end > m.offset+height {
		m.offset = end - height
	}
	m.offset = min(max(m.offset, 0), max(m.visibleHeight()-height, 0))
}

func (m Model) nodeAtMouse(x, y int) (int, bool) {
	if !m.contains(x, y) {
		return 0, false
	}
	row := y - m.y - 1
	if row < 0 || row >= m.contentHeight() {
		return 0, false
	}
	index, ok := m.nodeIndexAtRow(m.offset + row)
	return index, ok && m.visible[index].kind != nodeGap
}

func (m Model) contains(x, y int) bool {
	return x >= m.x && x < m.x+m.width && y >= m.y && y < m.y+m.height
}

func (m Model) nodeHeight(node visibleNode) int {
	if node.kind == nodeEntity {
		return 2
	}
	return 1
}

func (m Model) nodeStartRow(index int) int {
	row := 0
	for current := 0; current < min(max(index, 0), len(m.visible)); current++ {
		row += m.nodeHeight(m.visible[current])
	}
	return row
}

func (m Model) visibleHeight() int {
	height := 0
	for _, node := range m.visible {
		height += m.nodeHeight(node)
	}
	return height
}

func (m Model) nodeIndexAtRow(row int) (int, bool) {
	if row < 0 {
		return 0, false
	}
	start := 0
	for index, node := range m.visible {
		end := start + m.nodeHeight(node)
		if row >= start && row < end {
			return index, true
		}
		start = end
	}
	return 0, false
}
