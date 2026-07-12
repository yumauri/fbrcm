package parameters

import "sort"

func (m Model) stickyHeaderLines(offset int) (string, string, int) {
	projectIndex, groupIndex, bodyStart, _ := m.stickyHeaderContext(offset)
	if projectIndex < 0 || projectIndex >= len(m.visible) {
		return "", "", offset
	}

	excludedGroupKey := ""
	if groupIndex >= 0 {
		excludedGroupKey = m.visible[groupIndex].groupKey
	}
	projectUnderlined := m.projectHasHiddenContentAbove(m.visible[projectIndex].projectID, excludedGroupKey, bodyStart)
	projectLine := m.renderProjectNode(m.visible[projectIndex], projectIndex == m.cursor, projectUnderlined)
	if m.history && !m.historyStacked() {
		projectLine = m.renderHistoryGridLine(projectLine, projectIndex == m.cursor, nodeProject)
	}

	if groupIndex < 0 {
		return projectLine, "", bodyStart
	}

	groupUnderlined := m.groupHasHiddenContentAbove(m.visible[groupIndex].projectID, m.visible[groupIndex].groupKey, bodyStart)
	groupLine := m.renderGroupNode(m.visible[groupIndex], groupIndex == m.cursor, groupUnderlined)
	if m.history && !m.historyStacked() {
		groupLine = m.renderHistoryGridLine(groupLine, groupIndex == m.cursor, nodeGroup)
	}
	return projectLine, groupLine, bodyStart
}

func (m Model) stickyHeaderContext(offset int) (projectIndex, groupIndex, bodyStart, headerLines int) {
	bodyStart = m.bodyStartForOffset(offset)
	projectIndex = m.projectNodeIndexForLine(bodyStart)
	groupIndex = m.groupNodeIndexForLine(bodyStart)
	if projectIndex >= 0 && groupIndex >= 0 && m.visible[groupIndex].projectID != m.visible[projectIndex].projectID {
		groupIndex = -1
	}
	headerLines = 1
	if groupIndex >= 0 {
		headerLines = 2
	}
	return
}

func (m Model) nodeIndexAtLine(line int) int {
	if len(m.visible) == 0 {
		return -1
	}
	if line <= 0 {
		return 0
	}
	if line >= m.totalLines {
		return len(m.visible) - 1
	}

	index := sort.Search(len(m.visible), func(i int) bool {
		return m.lineIndexByNode[i]+m.nodeBlockLineCount(i) > line
	})
	return min(index, len(m.visible)-1)
}

func (m Model) projectNodeIndexFor(nodeIndex int) int {
	if nodeIndex < 0 || nodeIndex >= len(m.visible) {
		return -1
	}
	if nodeIndex < len(m.projectNodeFor) {
		return m.projectNodeFor[nodeIndex]
	}
	return -1
}

func (m Model) groupNodeIndexFor(nodeIndex int) int {
	if nodeIndex < 0 || nodeIndex >= len(m.visible) {
		return -1
	}
	groupKey := m.visible[nodeIndex].groupKey
	if groupKey == "" {
		return -1
	}
	if nodeIndex < len(m.groupNodeFor) {
		index := m.groupNodeFor[nodeIndex]
		if index >= 0 && m.visible[index].groupKey == groupKey {
			return index
		}
	}
	return -1
}

func (m Model) projectNodeIndexForLine(line int) int {
	return m.projectNodeIndexFor(m.nodeIndexAtLine(line))
}

func (m Model) groupNodeIndexForLine(line int) int {
	nodeIndex := m.nodeIndexAtLine(line)
	if nodeIndex < 0 || nodeIndex >= len(m.visible) {
		return -1
	}

	node := m.visible[nodeIndex]
	if node.kind == nodeProject {
		for i := nodeIndex + 1; i < len(m.visible); i++ {
			if m.visible[i].projectID != node.projectID {
				break
			}
			if m.visible[i].kind == nodeGroup {
				return i
			}
		}
		return -1
	}

	return m.groupNodeIndexFor(nodeIndex)
}

func (m Model) stickyHeaderLineCount(offset int) int {
	if len(m.visible) == 0 {
		return 0
	}
	_, _, _, headerLines := m.stickyHeaderContext(offset)
	return headerLines
}

func (m Model) bodyStartForOffset(offset int) int {
	if len(m.visible) == 0 {
		return offset
	}

	bodyStart := max(offset, 0)
	projectIndex := m.projectNodeIndexForLine(offset)
	if projectIndex >= 0 && offset <= m.lineIndexByNode[projectIndex] {
		bodyStart = max(bodyStart, m.lineIndexByNode[projectIndex]+1)
	}
	groupIndex := m.groupNodeIndexForLine(offset + 1)
	if groupIndex < 0 && projectIndex >= 0 {
		groupIndex = m.groupNodeIndexForLine(offset)
	}
	if groupIndex >= 0 && offset <= m.lineIndexByNode[groupIndex] {
		bodyStart = max(bodyStart, m.lineIndexByNode[groupIndex]+1)
	}
	return bodyStart
}

func (m Model) offsetForBodyStart(target int) int {
	if m.totalLines <= 0 {
		return 0
	}

	target = max(target, 0)
	lo := 0
	hi := m.totalLines - 1
	best := hi
	for lo <= hi {
		mid := lo + (hi-lo)/2
		bodyStart := m.bodyStartForOffset(mid)
		if bodyStart >= target {
			best = mid
			hi = mid - 1
		} else {
			lo = mid + 1
		}
	}
	return best
}

func (m Model) bodyVisibleLinesForOffset(offset int) int {
	lines := m.viewportHeight() - m.stickyHeaderLineCount(offset)
	if lines < 1 {
		return 1
	}
	return lines
}

func (m Model) projectHasHiddenContentAbove(projectID, excludedGroupKey string, bodyStart int) bool {
	for i := m.nodeIndexAtLine(bodyStart - 1); i >= 0; i-- {
		node := m.visible[i]
		if node.projectID != projectID {
			return false
		}
		if node.kind == nodeProject {
			return false
		}
		if node.kind == nodeGroup && node.groupKey == excludedGroupKey {
			continue
		}
		if m.lineIndexByNode[i] < bodyStart {
			return true
		}
	}
	return false
}

func (m Model) groupHasHiddenContentAbove(projectID, groupKey string, bodyStart int) bool {
	for i := m.nodeIndexAtLine(bodyStart - 1); i >= 0; i-- {
		node := m.visible[i]
		if node.projectID != projectID || node.groupKey != groupKey {
			return false
		}
		if node.kind == nodeGroup {
			return false
		}
		if m.lineIndexByNode[i] < bodyStart {
			return true
		}
	}
	return false
}

func (m Model) screenLineForOffset(cursor, offset int) int {
	if len(m.visible) == 0 || cursor < 0 || cursor >= len(m.visible) {
		return -1
	}

	projectIndex, groupIndex, bodyStart, headerLines := m.stickyHeaderContext(offset)

	if cursor == projectIndex {
		return 0
	}
	if groupIndex >= 0 && cursor == groupIndex {
		return 1
	}
	return headerLines + (m.lineIndexByNode[cursor] - bodyStart)
}
