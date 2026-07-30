package managedfeatures

import (
	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/filter"
	"github.com/yumauri/fbrcm/core/firebase"
	rctarget "github.com/yumauri/fbrcm/core/rc/target"
	"github.com/yumauri/fbrcm/core/strfold"
	"github.com/yumauri/fbrcm/tui/messages"
)

func (m *Model) setProjects(projects []core.Project) {
	clientProjects := make([]core.Project, 0, len(projects))
	seen := make(map[string]struct{}, len(projects))
	for _, project := range projects {
		target, err := rctarget.Parse(project.ProjectID)
		if err != nil || target.Kind != rctarget.Client {
			continue
		}
		if _, ok := seen[target.ProjectID]; ok {
			continue
		}
		seen[target.ProjectID] = struct{}{}
		project.ProjectID = target.ProjectID
		clientProjects = append(clientProjects, project)
	}
	strfold.SortProjects(clientProjects, func(p core.Project) string { return p.Name }, func(p core.Project) string { return p.ProjectID })

	next := make([]projectState, 0, len(clientProjects))
	nextIndex := make(map[string]int, len(clientProjects))
	for _, project := range clientProjects {
		state := projectState{project: project}
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

func (m *Model) updateLoaded(msg messages.ManagedFeaturesLoadedMsg) {
	if msg.Kind != m.kind {
		return
	}
	index, ok := m.projectIndex[msg.Project.ProjectID]
	if !ok {
		return
	}
	state := m.projects[index]
	state.loading = false
	state.waitingTemplate = false
	state.loaded = msg.Err == nil
	state.err = msg.Err
	if msg.Err == nil {
		state.templateReady = true
		state.template = msg.Template
		state.experiments = append([]core.ExperimentEntry(nil), msg.Experiments...)
		state.personalizations = append([]core.PersonalizationEntry(nil), msg.Personalizations...)
		state.rollouts = append([]core.RolloutEntry(nil), msg.Rollouts...)
	}
	m.projects[index] = state
	m.syncVisible()
}

func (m *Model) syncVisible() {
	selected := m.currentIdentity()
	m.visible = m.visible[:0]
	for projectIndex, project := range m.projects {
		if projectIndex > 0 {
			m.visible = append(m.visible, visibleNode{kind: nodeGap, index: -1})
		}
		m.visible = append(m.visible, visibleNode{kind: nodeProject, projectID: project.project.ProjectID, index: -1})
		for index := range m.entityCount(project) {
			if !m.entityMatches(project, index) {
				continue
			}
			m.visible = append(m.visible, visibleNode{
				kind: nodeEntity, projectID: project.project.ProjectID, index: index, entityID: m.entityID(project, index),
			})
		}
	}
	if len(m.visible) == 0 {
		m.cursor, m.offset = 0, 0
		return
	}
	m.cursor = min(max(m.cursor, 0), len(m.visible)-1)
	if selected.projectID != "" {
		for index, node := range m.visible {
			if node.kind == selected.kind && node.projectID == selected.projectID && node.entityID == selected.entityID {
				m.cursor = index
				break
			}
		}
	}
	m.cursor = m.nearestSelectableIndex(m.cursor, 1)
	m.ensureCursorVisible()
}

func (m Model) entityMatches(project projectState, index int) bool {
	if !m.filterEnabled() || m.filter.Value() == "" {
		return true
	}
	if index < 0 || index >= len(project.experiments) {
		return false
	}
	matched, _ := filter.Match(project.experiments[index].Definition.DisplayName, m.filter.Value(), m.filter.Mode())
	return matched
}

func (m Model) visibleEntityCount() int {
	count := 0
	for _, node := range m.visible {
		if node.kind == nodeEntity {
			count++
		}
	}
	return count
}

func (m Model) visibleProjectEntityCount(projectID string) int {
	count := 0
	for _, node := range m.visible {
		if node.kind == nodeEntity && node.projectID == projectID {
			count++
		}
	}
	return count
}

func (m Model) entityCount(project projectState) int {
	switch m.kind {
	case messages.ManagedFeatureExperiment:
		return len(project.experiments)
	case messages.ManagedFeaturePersonalization:
		return len(project.personalizations)
	case messages.ManagedFeatureRollout:
		return len(project.rollouts)
	default:
		return 0
	}
}

func (m Model) entityID(project projectState, index int) string {
	if index < 0 || index >= m.entityCount(project) {
		return ""
	}
	switch m.kind {
	case messages.ManagedFeatureExperiment:
		return firebase.ManagedFeatureID(project.experiments[index].Name)
	case messages.ManagedFeaturePersonalization:
		return project.personalizations[index].ID
	case messages.ManagedFeatureRollout:
		return firebase.ManagedFeatureID(project.rollouts[index].Name)
	default:
		return ""
	}
}

func (m Model) currentIdentity() visibleNode {
	if m.cursor < 0 || m.cursor >= len(m.visible) || m.visible[m.cursor].kind == nodeGap {
		return visibleNode{index: -1}
	}
	return m.visible[m.cursor]
}

func (m Model) currentData() (*messages.ManagedFeatureViewData, bool) {
	if m.cursor < 0 || m.cursor >= len(m.visible) {
		return nil, false
	}
	node := m.visible[m.cursor]
	if node.kind != nodeEntity {
		return nil, false
	}
	projectIndex, ok := m.projectIndex[node.projectID]
	if !ok {
		return nil, false
	}
	project := m.projects[projectIndex]
	if node.index < 0 || node.index >= m.entityCount(project) {
		return nil, false
	}
	data := &messages.ManagedFeatureViewData{Kind: m.kind, Project: project.project, Template: project.template}
	switch m.kind {
	case messages.ManagedFeatureExperiment:
		value := project.experiments[node.index]
		data.Experiment = &value
	case messages.ManagedFeaturePersonalization:
		value := project.personalizations[node.index]
		data.Personalization = &value
	case messages.ManagedFeatureRollout:
		value := project.rollouts[node.index]
		data.Rollout = &value
	default:
		return nil, false
	}
	return data, true
}

func (m Model) selectionChangedCmd(activate bool) tea.Cmd {
	data, ok := m.currentData()
	if !ok {
		return func() tea.Msg { return messages.ManagedFeatureSelectionChangedMsg{ResetScroll: true} }
	}
	return func() tea.Msg {
		return messages.ManagedFeatureSelectionChangedMsg{Data: data, Activate: activate}
	}
}
