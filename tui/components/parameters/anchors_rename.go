package parameters

import (
	"charm.land/lipgloss/v2"
	"github.com/yumauri/fbrcm/core"
)

func (m Model) CurrentRenameAnchor() (RenameAnchor, bool) {
	if m.cursor < 0 || m.cursor >= len(m.visible) {
		return RenameAnchor{}, false
	}
	node := m.visible[m.cursor]
	project := m.projectByID(node.projectID)
	if project == nil {
		return RenameAnchor{}, false
	}
	switch node.kind {
	case nodeGroup:
		if core.NormalizeRemoteConfigGroupKey(node.groupKey) == "" {
			return RenameAnchor{}, false
		}
		screenLine := m.screenLineForOffset(m.cursor, m.offset)
		if screenLine < 0 {
			return RenameAnchor{}, false
		}
		return RenameAnchor{
			Project:  project.project,
			IsGroup:  true,
			GroupKey: node.groupKey,
			Label:    node.label,
			X:        m.x + 1,
			Y:        m.y + screenLine,
			Width:    max(lipgloss.Width(node.label), 1),
			MaxWidth: max(m.viewportWidth()-3, 1),
		}, true
	case nodeParameter, nodeValue:
		if node.transient {
			screenLine := m.screenLineForOffset(m.cursor, m.offset)
			if screenLine < 0 {
				return RenameAnchor{}, false
			}
			layout := m.parameterRenderLayout()
			return RenameAnchor{
				Project:  project.project,
				GroupKey: node.groupKey,
				ParamKey: node.paramKey,
				Label:    node.label,
				X:        m.x + layout.paramStart - 1,
				Y:        m.y + screenLine,
				Width:    max(lipgloss.Width(node.label), 1),
				MaxWidth: max(m.viewportWidth()-layout.paramStart-1, 1),
			}, true
		}
		_, groupKey, paramKey, ok := m.currentParameterRef()
		if !ok {
			return RenameAnchor{}, false
		}
		paramIndex := m.currentParameterNodeIndex()
		if paramIndex < 0 {
			return RenameAnchor{}, false
		}
		screenLine := m.screenLineForOffset(paramIndex, m.offset)
		if screenLine < 0 {
			return RenameAnchor{}, false
		}
		layout := m.parameterRenderLayout()
		return RenameAnchor{
			Project:  project.project,
			GroupKey: groupKey,
			ParamKey: paramKey,
			Label:    node.label,
			X:        m.x + layout.paramStart - 1,
			Y:        m.y + screenLine,
			Width:    max(lipgloss.Width(node.label), 1),
			MaxWidth: max(m.viewportWidth()-layout.paramStart-1, 1),
		}, true
	default:
		return RenameAnchor{}, false
	}
}

func (m Model) CurrentTransientDuplicate() (project core.Project, groupKey, sourceParamKey, label string, ok bool) {
	if m.cursor < 0 || m.cursor >= len(m.visible) {
		return core.Project{}, "", "", "", false
	}
	node := m.visible[m.cursor]
	if !node.transient {
		return core.Project{}, "", "", "", false
	}
	projectState := m.projectByID(node.projectID)
	if projectState == nil {
		return core.Project{}, "", "", "", false
	}
	return projectState.project, node.groupKey, node.paramKey, node.label, true
}

func (m *Model) OpenTransientDuplicate(projectID, groupKey, sourceParamKey, label string) {
	m.transientDup = &transientDuplicate{
		projectID:     projectID,
		groupKey:      groupKey,
		afterParamKey: sourceParamKey,
		label:         label,
	}
	m.groupExpanded[m.groupKey(projectID, groupKey)] = true
	m.syncVisible()
	for i, node := range m.visible {
		if node.transient && node.projectID == projectID && node.groupKey == groupKey && node.paramKey == sourceParamKey && node.label == label {
			m.cursor = i
			m.ensureCursorVisible()
			return
		}
	}
}

func (m *Model) ClearTransientDuplicate() {
	if m.transientDup == nil {
		return
	}
	m.transientDup = nil
	m.syncVisible()
}

func (m *Model) ClearTransientDuplicateAndFocusSource() {
	if m.transientDup == nil {
		return
	}
	projectID := m.transientDup.projectID
	groupKey := m.transientDup.groupKey
	paramKey := m.transientDup.afterParamKey
	m.transientDup = nil
	m.syncVisible()
	m.selectParameter(projectID, groupKey, paramKey)
}

func (m *Model) ClearTransientDuplicateAndFocus(projectID, groupKey, paramKey string) {
	if m.transientDup == nil {
		return
	}
	m.transientDup = nil
	m.syncVisible()
	m.selectParameter(projectID, groupKey, paramKey)
}

func (m *Model) OpenTransientNewParameter(projectID, groupKey, afterParamKey string) {
	m.transientNew = &transientNewParameter{
		projectID:     projectID,
		groupKey:      groupKey,
		afterParamKey: afterParamKey,
		label:         "",
	}
	m.groupExpanded[m.groupKey(projectID, groupKey)] = true
	m.syncVisible()
	for i, node := range m.visible {
		if node.transient && node.projectID == projectID && node.groupKey == groupKey && node.paramKey == "" {
			m.cursor = i
			m.ensureCursorVisible()
			return
		}
	}
}

func (m *Model) ClearTransientNewParameter() {
	if m.transientNew == nil {
		return
	}
	m.transientNew = nil
	m.syncVisible()
}

func (m *Model) ClearTransientNewParameterAndFocus(projectID, groupKey, paramKey string) {
	if m.transientNew == nil {
		return
	}
	m.transientNew = nil
	m.syncVisible()
	m.selectParameter(projectID, groupKey, paramKey)
}
