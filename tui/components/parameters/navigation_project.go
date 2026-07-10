package parameters

import (
	"slices"

	"github.com/yumauri/fbrcm/core"
)

func (m Model) currentProjectID() string {
	if m.cursor < 0 || m.cursor >= len(m.visible) {
		return ""
	}
	return m.visible[m.cursor].projectID
}

func (m Model) currentProject() (core.Project, bool) {
	project := m.projectByID(m.currentProjectID())
	if project == nil {
		return core.Project{}, false
	}
	return project.project, true
}

func (m *Model) moveToCurrentProjectHeader() {
	if len(m.visible) == 0 || m.cursor < 0 || m.cursor >= len(m.visible) {
		return
	}
	for i := m.cursor; i >= 0; i-- {
		if m.visible[i].kind == nodeProject &&
			m.visible[i].projectID == m.visible[m.cursor].projectID {
			m.cursor = i
			m.ensureCursorVisible()
			return
		}
	}
}

func (m *Model) moveToLastParameterInCurrentProject() {
	projectID := m.currentProjectID()
	if projectID == "" {
		return
	}
	for i, node := range slices.Backward(m.visible) {
		if node.projectID != projectID {
			continue
		}
		if node.kind == nodeParameter {
			m.cursor = i
			m.ensureCursorVisible()
			return
		}
	}
}

func (m *Model) setAllParametersExpanded(expanded bool) {
	snapshot := m.captureSelectionSnapshot(expanded, false)
	for _, project := range m.projects {
		if project.tree == nil {
			continue
		}
		for _, group := range project.tree.Groups {
			for _, param := range group.Parameters {
				m.paramExpanded[m.paramKey(project.project.ProjectID, group.Key, param.Key)] = expanded
			}
		}
	}
	m.syncVisible()
	m.restoreSelectionSnapshot(snapshot)
}

func (m *Model) setAllGroupsExpanded(expanded bool) {
	snapshot := m.captureSelectionSnapshot(expanded, true)
	for _, project := range m.projects {
		if project.tree == nil {
			continue
		}
		for _, group := range project.tree.Groups {
			m.groupExpanded[m.groupKey(project.project.ProjectID, group.Key)] = expanded
		}
	}
	m.syncVisible()
	m.restoreSelectionSnapshot(snapshot)
}

func (m *Model) markProjectRefreshing(projectID string) {
	idx, ok := m.projectIndex[projectID]
	if !ok {
		return
	}
	state := m.projects[idx]
	state.verifying = true
	state.err = nil
	m.projects[idx] = state
}
