package parameters

import (
	"math/big"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/core/strfold"
	"github.com/yumauri/fbrcm/tui/messages"
)

func (m *Model) setProjects(projects []core.Project) tea.Cmd {
	strfold.SortProjects(projects, func(p core.Project) string { return p.Name }, func(p core.Project) string { return p.ProjectID })

	nextProjects := make([]projectState, 0, len(projects))
	nextIndex := make(map[string]int, len(projects))
	cmds := make([]tea.Cmd, 0)

	for _, project := range projects {
		if idx, ok := m.projectIndex[project.ProjectID]; ok {
			state := m.projects[idx]
			state.project = project
			nextIndex[project.ProjectID] = len(nextProjects)
			nextProjects = append(nextProjects, state)
			continue
		}

		nextIndex[project.ProjectID] = len(nextProjects)
		nextProjects = append(nextProjects, projectState{
			project: project,
			loading: true,
		})
		cmds = append(cmds, m.loadParametersCmd(project))
	}

	m.projects = nextProjects
	m.projectIndex = nextIndex
	m.syncVisible()

	return tea.Batch(cmds...)
}

func (m *Model) updateProject(msg messages.ParametersLoadedMsg) tea.Cmd {
	idx, ok := m.projectIndex[msg.Project.ProjectID]
	if !ok {
		return nil
	}

	state := m.projects[idx]
	if msg.Err != nil {
		if state.tree == nil {
			state.tree = nil
			state.source = msg.Source
		}
		state.err = msg.Err
	} else {
		state.tree = msg.Tree
		state.source = msg.Source
		if msg.CacheSource != "" {
			state.cacheSource = msg.CacheSource
		} else {
			state.cacheSource = msg.Source
		}
		state.err = nil
	}
	state.loading = false
	state.verifying = false
	state.hasDraft = msg.HasDraft
	state.staleDraft = msg.StaleDraft
	if msg.CacheVersion != "" {
		state.cacheVersion = msg.CacheVersion
	} else if msg.Tree != nil && !msg.HasDraft {
		state.cacheVersion = msg.Tree.Version
	}
	if msg.DraftVersion != "" {
		state.draftVersion = msg.DraftVersion
	} else if msg.HasDraft && msg.Tree != nil {
		state.draftVersion = msg.Tree.Version
	} else if !msg.HasDraft {
		state.draftVersion = ""
	}
	m.projects[idx] = state

	cmds := make([]tea.Cmd, 0, 1)
	if msg.Tree != nil {
		for _, group := range msg.Tree.Groups {
			groupKey := m.groupKey(msg.Project.ProjectID, group.Key)
			if _, ok := m.groupExpanded[groupKey]; !ok {
				m.groupExpanded[groupKey] = true
			}
		}
	}
	if msg.Revalidate && msg.Err == nil {
		state.verifying = true
		m.projects[idx] = state
		cmds = append(cmds, m.revalidateParametersCmd(msg.Project, msg.RevalidateCache))
	}

	m.syncVisible()
	if msg.SelectParamKey != "" {
		m.selectParameter(msg.Project.ProjectID, msg.SelectGroupKey, msg.SelectParamKey)
	}
	return tea.Batch(cmds...)
}

func (m *Model) revalidateCurrentProjectCmd() tea.Cmd {
	project, ok := m.currentProject()
	if !ok {
		return nil
	}
	m.markProjectRefreshing(project.ProjectID)
	m.syncVisible()
	return tea.Batch(m.forceParametersCmd(project), m.spin.Tick)
}

func (m *Model) revalidateAllProjectsCmd() tea.Cmd {
	if len(m.projects) == 0 {
		return nil
	}
	cmds := make([]tea.Cmd, 0, len(m.projects)+1)
	for _, project := range m.projects {
		m.markProjectRefreshing(project.project.ProjectID)
		cmds = append(cmds, m.forceParametersCmd(project.project))
	}
	m.syncVisible()
	cmds = append(cmds, m.spin.Tick)
	return tea.Batch(cmds...)
}

func (p projectState) cacheStateLabel() string {
	if p.tree == nil {
		return core.ParametersStatusLabel(p.cacheSource, time.Time{}, false, p.err)
	}
	return core.ParametersStatusLabel(p.cacheSource, p.tree.CachedAt, true, p.err)
}

func (p projectState) displayVersion() string {
	if p.staleDraft && p.cacheVersion != "" {
		return p.cacheVersion
	}
	if p.tree != nil && p.tree.Version != "" {
		return p.tree.Version
	}
	return p.cacheVersion
}

func (m Model) projectByID(projectID string) *projectState {
	idx, ok := m.projectIndex[projectID]
	if !ok || idx < 0 || idx >= len(m.projects) {
		return nil
	}
	return &m.projects[idx]
}

func (m Model) groupByKey(projectID, groupKey string) *core.ParametersGroup {
	project := m.projectByID(projectID)
	if project == nil || project.tree == nil {
		return nil
	}
	for i := range project.tree.Groups {
		if project.tree.Groups[i].Key == groupKey {
			return &project.tree.Groups[i]
		}
	}
	return nil
}

func (m Model) parameterByKey(projectID, groupKey, paramKey string) *core.ParametersEntry {
	group := m.groupByKey(projectID, groupKey)
	if group == nil {
		return nil
	}
	for i := range group.Parameters {
		if group.Parameters[i].Key == paramKey {
			return &group.Parameters[i]
		}
	}
	return nil
}

func (m Model) DraftProjects() []core.Project {
	out := make([]core.Project, 0)
	for _, project := range m.projects {
		if project.hasDraft {
			out = append(out, project.project)
		}
	}
	return out
}

func (m Model) HasDraft(projectID string) bool {
	project := m.projectByID(projectID)
	return project != nil && project.hasDraft
}

func (m Model) HasProject(projectID string) bool {
	return m.projectByID(projectID) != nil
}

func (m Model) ProjectDraftState(projectID string) (bool, bool) {
	project := m.projectByID(projectID)
	if project == nil {
		return false, false
	}
	return project.hasDraft, project.staleDraft
}

func remoteConfigVersion(raw []byte) string {
	cfg, err := firebase.ParseRemoteConfig(raw)
	if err != nil {
		return ""
	}
	return cfg.Version.VersionNumber
}

func versionLess(left, right string) bool {
	left = strings.TrimSpace(left)
	right = strings.TrimSpace(right)
	if left == "" || right == "" {
		return false
	}
	var leftNum, rightNum big.Int
	if _, ok := leftNum.SetString(left, 10); !ok {
		return false
	}
	if _, ok := rightNum.SetString(right, 10); !ok {
		return false
	}
	return leftNum.Cmp(&rightNum) < 0
}
