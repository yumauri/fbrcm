package parameters

import (
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/dictdiff"
	"github.com/yumauri/fbrcm/core/firebase"
	rcdiffinput "github.com/yumauri/fbrcm/core/rc/diffinput"
)

// HistoryDiffRequestedMsg asks the application to open the generic diff modal
// for one parameter across the selected History versions.
type HistoryDiffRequestedMsg struct {
	Project core.Project
	Input   dictdiff.Input
}

// HistoryDiffAvailable reports whether the History cursor is on a parameter
// backed by a loaded version pair.
func (m Model) HistoryDiffAvailable() bool {
	_, _, ok := m.historyDiffInput()
	return ok
}

func (m Model) historyDiffRequestedCmd() tea.Cmd {
	project, input, ok := m.historyDiffInput()
	if !ok {
		return nil
	}
	return func() tea.Msg {
		return HistoryDiffRequestedMsg{Project: project, Input: input}
	}
}

func (m Model) historyDiffInput() (core.Project, dictdiff.Input, bool) {
	if !m.history || m.cursor < 0 || m.cursor >= len(m.visible) {
		return core.Project{}, dictdiff.Input{}, false
	}
	node := m.visible[m.cursor]
	if node.kind != nodeParameter || node.transient {
		return core.Project{}, dictdiff.Input{}, false
	}
	projectState := m.projectByID(node.projectID)
	if projectState == nil {
		return core.Project{}, dictdiff.Input{}, false
	}
	history, ok := m.histories[node.projectID]
	if !ok || history.loading || history.err != nil || history.unavailable ||
		history.previous == nil || history.current == nil {
		return core.Project{}, dictdiff.Input{}, false
	}
	previous, previousOK := historyRemoteConfigParameter(history.previous, node.groupKey, node.paramKey)
	current, currentOK := historyRemoteConfigParameter(history.current, node.groupKey, node.paramKey)
	if !previousOK && !currentOK {
		return core.Project{}, dictdiff.Input{}, false
	}
	group := core.NormalizeRemoteConfigGroupKey(node.groupKey)
	input := dictdiff.Input{
		EntityName: rcdiffinput.ParameterEntityName(group, node.paramKey),
		Left: dictdiff.NamedDictionary{
			Name:       "Earlier version: " + displayHistoryVersion(history.previousVersion),
			Properties: rcdiffinput.Parameter(group, previous),
		},
		Right: dictdiff.NamedDictionary{
			Name:       "Later version: " + displayHistoryVersion(history.currentVersion),
			Properties: rcdiffinput.Parameter(group, current),
		},
	}
	return projectState.project, input, true
}

func historyRemoteConfigParameter(
	tree *core.ParametersTree,
	groupKey string,
	parameterKey string,
) (*firebase.RemoteConfigParam, bool) {
	if tree == nil || tree.RemoteConfig() == nil {
		return nil, false
	}
	config := tree.RemoteConfig()
	groupKey = core.NormalizeRemoteConfigGroupKey(groupKey)
	if groupKey == "" {
		parameter, ok := config.Parameters[parameterKey]
		if !ok {
			return nil, false
		}
		return &parameter, true
	}
	group, ok := config.ParameterGroups[groupKey]
	if !ok {
		return nil, false
	}
	parameter, ok := group.Parameters[parameterKey]
	if !ok {
		return nil, false
	}
	return &parameter, true
}

func displayHistoryVersion(version string) string {
	version = strings.TrimSpace(version)
	if version == "" {
		return "…"
	}
	if strings.HasPrefix(strings.ToLower(version), "v") {
		return version
	}
	return "v" + version
}
