package parameters

func (m *Model) FocusParameter(projectID, groupKey, paramKey string) bool {
	return m.selectParameter(projectID, groupKey, paramKey)
}

// FocusValue selects a parameter value node by index within the parameter's Values slice.
func (m *Model) FocusValue(projectID, groupKey, paramKey string, valueIdx int) bool {
	if projectID == "" || paramKey == "" || valueIdx < 0 {
		return false
	}
	m.groupExpanded[m.groupKey(projectID, groupKey)] = true
	m.paramExpanded[m.paramKey(projectID, groupKey, paramKey)] = true
	m.syncVisible()
	for i, node := range m.visible {
		if node.kind == nodeValue && node.projectID == projectID && node.groupKey == groupKey && node.paramKey == paramKey && node.valueIdx == valueIdx {
			m.cursor = i
			m.ensureCursorVisible()
			return true
		}
	}
	return false
}

func (m *Model) selectParameter(projectID, groupKey, paramKey string) bool {
	if projectID == "" || paramKey == "" {
		return false
	}
	m.groupExpanded[m.groupKey(projectID, groupKey)] = true
	m.syncVisible()
	for i, node := range m.visible {
		if node.kind == nodeParameter && node.projectID == projectID && node.groupKey == groupKey && node.paramKey == paramKey {
			m.cursor = i
			m.ensureCursorVisible()
			return true
		}
	}
	return false
}
