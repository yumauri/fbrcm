package parameters

func (m Model) selectionBlockRange(index int) (int, int) {
	if index < 0 || index >= len(m.visible) {
		return 0, 0
	}

	start := m.lineIndexByNode[index]
	end := start + m.nodeBlockLineCount(index)
	node := m.visible[index]
	if node.kind != nodeParameter || !node.expanded {
		return start, end
	}

	for i := index + 1; i < len(m.visible); i++ {
		next := m.visible[i]
		if next.kind != nodeValue ||
			next.projectID != node.projectID ||
			next.groupKey != node.groupKey ||
			next.paramKey != node.paramKey {
			break
		}
		end = m.lineIndexByNode[i] + m.nodeBlockLineCount(i)
	}

	return start, end
}
