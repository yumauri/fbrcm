package conditions

import (
	"fmt"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/filter"
	"github.com/yumauri/fbrcm/core/strfold"
	"github.com/yumauri/fbrcm/tui/messages"
)

func (m *Model) setProjects(projects []core.Project) {
	if m.move != nil {
		m.CancelMove()
	}
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
	if m.move != nil {
		m.CancelMove()
	}
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
		if msg.Source == "draft" || msg.Source == "draft-stale" {
			state.hasDraft = true
			state.staleDraft = msg.Source == "draft-stale"
			if state.cacheSource == "" {
				state.cacheSource = "cache"
			}
		}
		if state.cacheVersion == "" && !state.hasDraft && msg.Tree != nil {
			state.cacheVersion = msg.Tree.Version
		}
	}
	m.projects[idx] = state
	m.syncVisible()
	if msg.SelectConditionName != "" {
		m.selectCondition(msg.Project.ProjectID, msg.SelectConditionName)
	}
}

func (m *Model) syncVisible() {
	selected := m.currentIdentity()
	m.visible = m.visible[:0]
	query := m.filter.Value()
	for projectIndex, project := range m.projects {
		if projectIndex > 0 {
			m.visible = append(m.visible, visibleNode{kind: nodeGap, conditionIndex: -1})
		}
		m.visible = append(m.visible, visibleNode{kind: nodeProject, projectID: project.project.ProjectID, conditionIndex: -1})
		if project.tree == nil {
			continue
		}
		for i, condition := range project.tree.Conditions {
			if m.filter.ExpressionMode() {
				matched, err := m.filter.CompiledExpression().MatchCondition(project.project.ProjectID, project.project.Name, condition)
				if err != nil || !matched {
					continue
				}
			} else if !conditionMatches(condition, query, m.filter.Mode()) {
				continue
			}
			m.visible = append(m.visible, visibleNode{kind: nodeCondition, projectID: project.project.ProjectID, conditionIndex: i, conditionName: condition.Name})
		}
	}
	if len(m.visible) == 0 {
		m.cursor, m.offset = 0, 0
		return
	}
	m.cursor = min(max(m.cursor, 0), len(m.visible)-1)
	if selected.projectID != "" {
		for i, node := range m.visible {
			if node.projectID == selected.projectID && node.kind == selected.kind && ((node.kind == nodeCondition && node.conditionName == selected.conditionName) || (node.kind != nodeCondition && node.conditionIndex == selected.conditionIndex)) {
				m.cursor = i
				break
			}
		}
	}
	m.cursor = m.nearestSelectableIndex(m.cursor, 1)
	m.ensureCursorVisible()
}

func (m *Model) selectCondition(projectID, name string) {
	for index, node := range m.visible {
		if node.kind == nodeCondition && node.projectID == projectID && node.conditionName == name {
			m.cursor = index
			m.ensureCursorVisible()
			return
		}
	}
}

func conditionMatches(condition core.ConditionEntry, query string, mode filter.Mode) bool {
	if query == "" {
		return true
	}
	for _, value := range []string{condition.Name, condition.Expression} {
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
	names := make([]string, 0, len(m.projects[idx].tree.Conditions))
	for _, condition := range m.projects[idx].tree.Conditions {
		names = append(names, condition.Name)
	}
	return &messages.ConditionViewData{
		Project:        m.projects[idx].project,
		Condition:      m.projects[idx].tree.Conditions[node.conditionIndex],
		ConditionNames: names,
	}, true
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

func (m Model) CurrentCondition() (*messages.ConditionViewData, bool) {
	return m.currentData()
}

// Condition returns one condition without changing the Conditions panel cursor.
func (m Model) Condition(projectID, name string) (*messages.ConditionViewData, bool) {
	index, ok := m.projectIndex[projectID]
	if !ok || m.projects[index].tree == nil {
		return nil, false
	}
	names := make([]string, 0, len(m.projects[index].tree.Conditions))
	var selected core.ConditionEntry
	found := false
	for _, condition := range m.projects[index].tree.Conditions {
		names = append(names, condition.Name)
		if condition.Name == name {
			selected = condition
			found = true
		}
	}
	if !found {
		return nil, false
	}
	return &messages.ConditionViewData{
		Project: m.projects[index].project, Condition: selected, ConditionNames: names,
	}, true
}

func (m Model) CurrentConditions() ([]core.ConditionEntry, bool) {
	project, ok := m.CurrentProject()
	if !ok {
		return nil, false
	}
	index, ok := m.projectIndex[project.ProjectID]
	if !ok || m.projects[index].tree == nil {
		return nil, false
	}
	return append([]core.ConditionEntry(nil), m.projects[index].tree.Conditions...), true
}

func (m Model) CurrentDeleteImpact() (core.ConditionDeleteImpact, error) {
	data, ok := m.currentData()
	if !ok {
		return core.ConditionDeleteImpact{}, fmt.Errorf("condition not selected")
	}
	index := m.projectIndex[data.Project.ProjectID]
	return m.projects[index].tree.DeleteImpact(data.Condition.Name)
}

func (m Model) CurrentMoveImpact(priority int) (core.ConditionMoveImpact, error) {
	data, ok := m.currentData()
	if !ok {
		return core.ConditionMoveImpact{}, fmt.Errorf("condition not selected")
	}
	index := m.projectIndex[data.Project.ProjectID]
	return m.projects[index].tree.MoveImpact(data.Condition.Name, priority)
}

func (m Model) CurrentEditAnchor() (EditAnchor, bool) {
	data, ok := m.currentData()
	if !ok {
		return EditAnchor{}, false
	}
	row := m.cursor - m.offset
	return EditAnchor{
		Project:   data.Project,
		Condition: data.Condition,
		X:         m.x + 8,
		Y:         m.y + 1 + row,
		Width:     max(len([]rune(data.Condition.Name)), 1),
		MaxWidth:  max(m.width-9, 1),
	}, true
}

func (m Model) CurrentProjectAnchor() (core.Project, int, int, bool) {
	project, ok := m.CurrentProject()
	if !ok {
		return core.Project{}, 0, 0, false
	}
	return project, m.x + 2, m.y + 1 + m.cursor - m.offset, true
}

func (m Model) HasDraft(projectID string) bool {
	index, ok := m.projectIndex[projectID]
	if !ok {
		return false
	}
	return m.projects[index].source == "draft" || m.projects[index].source == "draft-stale"
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
