package parameters

func (m Model) captureSelectionSnapshot(expanding, groups bool) selectionSnapshot {
	snapshot := selectionSnapshot{valueIdx: -1}
	if len(m.visible) == 0 || m.cursor < 0 || m.cursor >= len(m.visible) {
		return snapshot
	}

	node := m.visible[m.cursor]
	snapshot.projectID = node.projectID
	snapshot.groupKey = node.groupKey
	snapshot.paramKey = node.paramKey
	snapshot.valueIdx = node.valueIdx
	if node.kind == nodeValue {
		snapshot.valueLabel = node.label
	}
	snapshot.kind = node.kind
	snapshot.screenLine = m.screenLineForOffset(m.cursor, m.offset)

	if expanding {
		return snapshot
	}

	if groups {
		if node.kind == nodeParameter || node.kind == nodeValue {
			snapshot.kind = nodeGroup
			snapshot.paramKey = ""
			snapshot.valueIdx = -1
		}
		return snapshot
	}

	if node.kind == nodeValue {
		snapshot.kind = nodeParameter
		snapshot.valueIdx = -1
	}
	return snapshot
}

func (m *Model) applyFilter() {
	snapshot := selectionSnapshot{valueIdx: -1}
	if len(m.visible) > 0 && m.cursor >= 0 && m.cursor < len(m.visible) {
		snapshot = m.captureSelectionSnapshot(true, false)
	}
	m.syncVisible()
	if len(m.visible) > 0 {
		m.restoreSelectionSnapshot(snapshot)
	}
}

func (m *Model) restoreSelectionSnapshot(snapshot selectionSnapshot) {
	if len(m.visible) == 0 {
		m.cursor = 0
		m.offset = 0
		return
	}

	cursor := m.findSelectionSnapshotNode(snapshot)
	if cursor < 0 {
		cursor = min(max(m.cursor, 0), len(m.visible)-1)
	}
	m.cursor = cursor
	m.restoreSelectionScreenLine(snapshot.screenLine)
}

func (m *Model) restoreSelectionScreenLine(target int) {
	if len(m.visible) == 0 {
		m.cursor, m.offset = 0, 0
		return
	}
	bestOffset := m.offset
	bestScore := int(^uint(0) >> 1)
	for offset := 0; offset < max(m.totalLines, 1); offset++ {
		current := m.screenLineForOffset(m.cursor, offset)
		if current < 0 {
			continue
		}
		score := abs(current - target)
		if score < bestScore {
			bestScore = score
			bestOffset = offset
			if score == 0 {
				break
			}
		}
	}
	m.offset = bestOffset
	m.ensureCursorVisible()
}

func (m Model) findExactSelectionSnapshotNode(snapshot selectionSnapshot) int {
	for i, node := range m.visible {
		if node.projectID != snapshot.projectID || node.kind != snapshot.kind {
			continue
		}
		switch snapshot.kind {
		case nodeProject:
			return i
		case nodeGroup:
			if node.groupKey == snapshot.groupKey {
				return i
			}
		case nodeParameter:
			if node.groupKey == snapshot.groupKey && node.paramKey == snapshot.paramKey {
				return i
			}
		case nodeValue:
			if node.groupKey == snapshot.groupKey && node.paramKey == snapshot.paramKey && ((snapshot.valueLabel == "" && node.valueIdx == snapshot.valueIdx) || node.label == snapshot.valueLabel) {
				return i
			}
		}
	}
	return -1
}

func (m Model) findSelectionSnapshotNode(snapshot selectionSnapshot) int {
	fallbackProject := -1
	fallbackGroup := -1
	fallbackParam := -1

	for i, node := range m.visible {
		if node.projectID != snapshot.projectID {
			continue
		}
		if fallbackProject < 0 && node.kind == nodeProject {
			fallbackProject = i
		}
		if node.groupKey == snapshot.groupKey && fallbackGroup < 0 && node.kind == nodeGroup {
			fallbackGroup = i
		}
		if node.groupKey == snapshot.groupKey && node.paramKey == snapshot.paramKey && fallbackParam < 0 && node.kind == nodeParameter {
			fallbackParam = i
		}

		switch snapshot.kind {
		case nodeProject:
			if node.kind == nodeProject {
				return i
			}
		case nodeGroup:
			if node.kind == nodeGroup && node.groupKey == snapshot.groupKey {
				return i
			}
		case nodeParameter:
			if node.kind == nodeParameter && node.groupKey == snapshot.groupKey && node.paramKey == snapshot.paramKey {
				return i
			}
		case nodeValue:
			if node.kind == nodeValue && node.groupKey == snapshot.groupKey && node.paramKey == snapshot.paramKey && node.valueIdx == snapshot.valueIdx {
				return i
			}
		}
	}

	if snapshot.kind == nodeValue || snapshot.kind == nodeParameter {
		if fallbackParam >= 0 {
			return fallbackParam
		}
	}
	if snapshot.kind == nodeValue || snapshot.kind == nodeParameter || snapshot.kind == nodeGroup {
		if fallbackGroup >= 0 {
			return fallbackGroup
		}
	}
	return fallbackProject
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
