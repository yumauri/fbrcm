package conditions

import (
	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/filter"
	"github.com/yumauri/fbrcm/core/strfold"
	"github.com/yumauri/fbrcm/tui/messages"
)

func (m *Model) setProjects(projects []core.Project) {
	projects = append([]core.Project(nil), projects...)
	strfold.SortProjects(projects, func(p core.Project) string { return p.Name }, func(p core.Project) string { return p.ProjectID })
	next := make([]projectState, 0, len(projects))
	nextIndex := make(map[string]int, len(projects))
	for _, project := range projects {
		state := projectState{project: project, loading: true}
		if previous, ok := m.projectIndex[project.ProjectID]; ok {
			state = m.projects[previous]
			state.project = project
		}
		nextIndex[project.ProjectID] = len(next)
		next = append(next, state)
	}
	m.projects = next
	m.projectIndex = nextIndex
	m.syncVisible()
}

func (m *Model) updateLoaded(msg messages.ConditionsLoadedMsg) {
	idx, ok := m.projectIndex[msg.Project.ProjectID]
	if !ok {
		return
	}
	state := m.projects[idx]
	state.loading = false
	state.err = msg.Err
	if msg.Err == nil {
		state.tree = msg.Tree
		state.source = msg.Source
	}
	m.projects[idx] = state
	m.syncVisible()
}

func (m *Model) syncVisible() {
	selected := m.currentIdentity()
	m.visible = m.visible[:0]
	query := m.filter.Value()
	mode := m.filter.Mode()
	for projectIndex, project := range m.projects {
		if projectIndex > 0 {
			m.visible = append(m.visible, visibleNode{kind: nodeGap, conditionIndex: -1})
		}
		m.visible = append(m.visible, visibleNode{kind: nodeProject, projectID: project.project.ProjectID, conditionIndex: -1})
		if project.tree == nil {
			continue
		}
		for i, condition := range project.tree.Conditions {
			if !conditionMatches(condition, query, mode) {
				continue
			}
			m.visible = append(m.visible, visibleNode{kind: nodeCondition, projectID: project.project.ProjectID, conditionIndex: i})
		}
	}
	if len(m.visible) == 0 {
		m.cursor, m.offset = 0, 0
		return
	}
	m.cursor = min(max(m.cursor, 0), len(m.visible)-1)
	if selected.projectID != "" {
		for i, node := range m.visible {
			if node.projectID == selected.projectID && node.kind == selected.kind && node.conditionIndex == selected.conditionIndex {
				m.cursor = i
				break
			}
		}
	}
	m.cursor = m.nearestSelectableIndex(m.cursor, 1)
	m.ensureCursorVisible()
}

func conditionMatches(condition core.ConditionEntry, query string, mode filter.Mode) bool {
	if query == "" {
		return true
	}
	for _, value := range []string{condition.Name, condition.Expression, condition.Description} {
		if matched, _ := filter.Match(value, query, mode); matched {
			return true
		}
	}
	return false
}

func (m Model) currentIdentity() visibleNode {
	if m.cursor < 0 || m.cursor >= len(m.visible) {
		return visibleNode{conditionIndex: -1}
	}
	if m.visible[m.cursor].kind == nodeGap {
		return visibleNode{conditionIndex: -1}
	}
	return m.visible[m.cursor]
}

func (m Model) currentData() (*messages.ConditionViewData, bool) {
	if m.cursor < 0 || m.cursor >= len(m.visible) {
		return nil, false
	}
	node := m.visible[m.cursor]
	if node.kind != nodeCondition {
		return nil, false
	}
	idx, ok := m.projectIndex[node.projectID]
	if !ok || m.projects[idx].tree == nil || node.conditionIndex < 0 || node.conditionIndex >= len(m.projects[idx].tree.Conditions) {
		return nil, false
	}
	return &messages.ConditionViewData{Project: m.projects[idx].project, Condition: m.projects[idx].tree.Conditions[node.conditionIndex]}, true
}

// CurrentProject returns the project represented by the current Conditions row.
func (m Model) CurrentProject() (core.Project, bool) {
	if m.cursor < 0 || m.cursor >= len(m.visible) {
		return core.Project{}, false
	}
	node := m.visible[m.cursor]
	idx, ok := m.projectIndex[node.projectID]
	if !ok {
		return core.Project{}, false
	}
	return m.projects[idx].project, true
}

// MarkProjectReloading updates one Conditions project row and starts its spinner.
func (m Model) MarkProjectReloading(projectID string) (Model, tea.Cmd) {
	idx, ok := m.projectIndex[projectID]
	if !ok {
		return m, nil
	}
	m.projects[idx].loading = true
	m.projects[idx].err = nil
	m.syncVisible()
	return m, m.spin.Tick
}

// MarkAllReloading updates all Conditions project rows and starts their spinner.
func (m Model) MarkAllReloading() (Model, tea.Cmd) {
	for i := range m.projects {
		m.projects[i].loading = true
		m.projects[i].err = nil
	}
	m.syncVisible()
	if len(m.projects) == 0 {
		return m, nil
	}
	return m, m.spin.Tick
}
