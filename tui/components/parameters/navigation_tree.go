package parameters

func (m *Model) moveToNextGroup() {
	if len(m.visible) == 0 {
		return
	}
	current := m.cursor
	for i := current + 1; i < len(m.visible); i++ {
		if m.visible[i].kind == nodeGroup {
			m.cursor = i
			m.offset = m.lineIndexByNode[m.cursor]
			maxOffset := max(m.totalLines-m.bodyVisibleLinesForOffset(m.offset), 0)
			if m.offset > maxOffset {
				m.offset = maxOffset
			}
			return
		}
	}
}

func (m *Model) moveToPrevGroup() {
	if len(m.visible) == 0 {
		return
	}
	current := m.cursor
	for i := current - 1; i >= 0; i-- {
		if m.visible[i].kind == nodeGroup {
			m.cursor = i
			m.ensureCursorVisible()
			return
		}
	}
}

func (m *Model) collapseCurrent() {
	if len(m.visible) == 0 {
		return
	}

	node := m.visible[m.cursor]
	switch node.kind {
	case nodeParameter:
		key := m.paramKey(node.projectID, node.groupKey, node.paramKey)
		if m.paramExpanded[key] {
			m.paramExpanded[key] = false
		} else {
			m.focusParentGroup(node)
			return
		}
	case nodeGroup:
		key := m.groupKey(node.projectID, node.groupKey)
		if m.groupExpanded[key] {
			m.groupExpanded[key] = false
		}
	case nodeValue:
		m.focusParentParameter(node)
		return
	default:
		return
	}

	m.syncVisible()
}

func (m *Model) expandCurrent() {
	if len(m.visible) == 0 {
		return
	}

	node := m.visible[m.cursor]
	switch node.kind {
	case nodeGroup:
		m.groupExpanded[m.groupKey(node.projectID, node.groupKey)] = true
	case nodeParameter:
		m.paramExpanded[m.paramKey(node.projectID, node.groupKey, node.paramKey)] = true
	default:
		return
	}

	m.syncVisible()
}

func (m *Model) toggleCurrentParameter() {
	if len(m.visible) == 0 {
		return
	}

	node := m.visible[m.cursor]
	if node.kind != nodeParameter {
		return
	}

	key := m.paramKey(node.projectID, node.groupKey, node.paramKey)
	m.paramExpanded[key] = !m.paramExpanded[key]
	m.syncVisible()
}

func (m *Model) focusCurrentParameterDefaultValue() bool {
	if len(m.visible) == 0 || m.cursor < 0 || m.cursor >= len(m.visible) {
		return false
	}

	node := m.visible[m.cursor]
	if node.kind != nodeParameter || node.transient {
		return false
	}

	key := m.paramKey(node.projectID, node.groupKey, node.paramKey)
	if !m.paramExpanded[key] {
		m.paramExpanded[key] = true
		m.syncVisible()
	}

	firstValueIdx := -1
	for i, candidate := range m.visible {
		if candidate.projectID != node.projectID || candidate.groupKey != node.groupKey || candidate.paramKey != node.paramKey {
			continue
		}
		if candidate.kind != nodeValue {
			continue
		}
		if firstValueIdx < 0 {
			firstValueIdx = i
		}
		if candidate.label == "default" {
			m.cursor = i
			m.ensureCursorVisible()
			return true
		}
	}

	if firstValueIdx >= 0 {
		m.cursor = firstValueIdx
		m.ensureCursorVisible()
		return true
	}

	return false
}
