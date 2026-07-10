package parameters

func (m *Model) ensureCursorVisible() {
	if len(m.visible) == 0 {
		m.cursor = 0
		m.offset = 0
		return
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor >= len(m.visible) {
		m.cursor = len(m.visible) - 1
	}

	blockStart, blockEnd := m.selectionBlockRange(m.cursor)
	rowHeight := blockEnd - blockStart
	visibleLines := m.bodyVisibleLinesForOffset(m.offset)
	bodyStart := m.bodyStartForOffset(m.offset)

	desiredBodyStart := bodyStart
	if rowHeight >= visibleLines {
		desiredBodyStart = blockStart
	} else {
		if blockStart < bodyStart {
			desiredBodyStart = blockStart
		}
		if blockEnd > bodyStart+visibleLines {
			desiredBodyStart = blockEnd - visibleLines
		}
	}

	m.offset = m.offsetForBodyStart(desiredBodyStart)

	maxOffset := max(m.totalLines-visibleLines, 0)
	if m.offset > maxOffset {
		m.offset = maxOffset
	}
}

func (m *Model) moveCursor(delta int) {
	if len(m.visible) == 0 {
		return
	}
	m.cursor = max(0, min(m.cursor+delta, len(m.visible)-1))
	m.ensureCursorVisible()
}

func (m *Model) focusParentGroup(node visibleNode) {
	for i, candidate := range m.visible {
		if candidate.kind == nodeGroup && candidate.projectID == node.projectID && candidate.groupKey == node.groupKey {
			m.cursor = i
			m.ensureCursorVisible()
			return
		}
	}
}

func (m *Model) focusParentParameter(node visibleNode) {
	for i, candidate := range m.visible {
		if candidate.kind == nodeParameter && candidate.projectID == node.projectID && candidate.groupKey == node.groupKey && candidate.paramKey == node.paramKey {
			m.cursor = i
			m.ensureCursorVisible()
			return
		}
	}
}
