package parameters

import (
	tea "charm.land/bubbletea/v2"
	"github.com/atotto/clipboard"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/rootgroup"
	"github.com/yumauri/fbrcm/tui/messages"
)

func (m Model) copyCurrentParameterNameCmd() tea.Cmd {
	_, _, paramKey, ok := m.currentParameterRef()
	if !ok {
		return nil
	}
	return copyToClipboardCmd(paramKey)
}

func (m Model) copyCurrentParameterPathCmd() tea.Cmd {
	projectID, groupKey, paramKey, ok := m.currentParameterRef()
	if !ok {
		return nil
	}
	return copyToClipboardCmd(projectID + "/" + groupKey + "/" + paramKey)
}

func (m Model) currentParameterRef() (projectID, groupKey, paramKey string, ok bool) {
	if m.cursor < 0 || m.cursor >= len(m.visible) {
		return "", "", "", false
	}

	node := m.visible[m.cursor]
	if node.transient {
		return "", "", "", false
	}
	switch node.kind {
	case nodeParameter, nodeValue:
		return node.projectID, node.groupKey, node.paramKey, true
	default:
		return "", "", "", false
	}
}

func (m Model) currentParameterViewData() (*messages.ParameterViewData, bool) {
	if m.cursor < 0 || m.cursor >= len(m.visible) {
		return nil, false
	}

	node := m.visible[m.cursor]
	if node.kind != nodeParameter && node.kind != nodeValue {
		return nil, false
	}

	project := m.projectByID(node.projectID)
	group := m.groupByKey(node.projectID, node.groupKey)
	if project == nil {
		return nil, false
	}
	groupKey := node.groupKey
	groupLabel := rootgroup.Label
	if group != nil {
		groupKey = group.Key
		groupLabel = group.Label
	}
	groups, paramKeys := parameterViewOptions(project)
	conditions := parameterConditionOptions(project)
	if len(groups) == 0 {
		groups = []messages.ParameterGroupOption{{Key: rootgroup.TreeKey, Label: rootgroup.Label}}
	}
	if node.transient && m.transientNew != nil &&
		m.transientNew.projectID == node.projectID &&
		core.NormalizeRemoteConfigGroupKey(m.transientNew.groupKey) == core.NormalizeRemoteConfigGroupKey(node.groupKey) {
		return &messages.ParameterViewData{
			Project:       project.project,
			GroupKey:      groupKey,
			GroupLabel:    groupLabel,
			Groups:        groups,
			ParameterKeys: paramKeys,
			Conditions:    conditions,
			Parameter: core.ParametersEntry{
				Key:     "",
				Summary: "new parameter",
				Values: []core.ParametersValue{{
					Label:     "default",
					Value:     "(empty string)",
					RawValue:  "",
					ValueType: "STRING",
					Empty:     true,
					EmptyType: "STRING",
					Plain:     true,
				}},
			},
			SelectedValueIdx: -1,
		}, true
	}
	if node.transient {
		return nil, false
	}
	valueIdx := -1
	if node.kind == nodeValue {
		valueIdx = node.valueIdx
	}
	return m.parameterViewData(node.projectID, node.groupKey, node.paramKey, valueIdx)
}

func (m Model) currentGroupViewData() (*messages.GroupViewData, bool) {
	if m.cursor < 0 || m.cursor >= len(m.visible) {
		return nil, false
	}
	node := m.visible[m.cursor]
	if node.kind != nodeGroup || node.transient || core.NormalizeRemoteConfigGroupKey(node.groupKey) == "" {
		return nil, false
	}
	project := m.projectByID(node.projectID)
	group := m.groupByKey(node.projectID, node.groupKey)
	if project == nil || group == nil {
		return nil, false
	}
	names := make([]string, 0, len(project.tree.Groups))
	for _, candidate := range project.tree.Groups {
		if core.NormalizeRemoteConfigGroupKey(candidate.Key) != "" {
			names = append(names, candidate.Key)
		}
	}
	return &messages.GroupViewData{Project: project.project, Group: *group, GroupNames: names}, true
}

func (m Model) parameterViewData(projectID, groupKey, paramKey string, valueIdx int) (*messages.ParameterViewData, bool) {
	project := m.projectByID(projectID)
	param := m.parameterByKey(projectID, groupKey, paramKey)
	if project == nil || param == nil {
		return nil, false
	}
	groupLabel := rootgroup.Label
	canonicalGroupKey := groupKey
	if group := m.groupByKey(projectID, groupKey); group != nil {
		canonicalGroupKey = group.Key
		groupLabel = group.Label
	}
	groups, paramKeys := parameterViewOptions(project)
	conditions := parameterConditionOptions(project)
	if len(groups) == 0 {
		groups = []messages.ParameterGroupOption{{Key: rootgroup.TreeKey, Label: rootgroup.Label}}
	}
	return &messages.ParameterViewData{
		Project: project.project, GroupKey: canonicalGroupKey, GroupLabel: groupLabel,
		Groups: groups, ParameterKeys: paramKeys, Conditions: conditions,
		Parameter: *param, SelectedValueIdx: valueIdx,
	}, true
}

// ParameterViewData returns one parameter without moving the Parameters cursor.
func (m Model) ParameterViewData(projectID, groupKey, paramKey, valueLabel string) (*messages.ParameterViewData, bool) {
	data, ok := m.parameterViewData(projectID, groupKey, paramKey, -1)
	if !ok {
		return nil, false
	}
	for index, value := range data.Parameter.Values {
		if value.Label == valueLabel {
			data.SelectedValueIdx = index
			break
		}
	}
	return data, true
}

func parameterConditionOptions(project *projectState) []core.ParametersCondition {
	if project == nil || project.tree == nil {
		return nil
	}
	return append([]core.ParametersCondition(nil), project.tree.Conditions...)
}

func parameterViewOptions(project *projectState) ([]messages.ParameterGroupOption, []string) {
	if project == nil || project.tree == nil {
		return nil, nil
	}
	groups := make([]messages.ParameterGroupOption, 0, len(project.tree.Groups)+1)
	seenRoot := false
	for _, group := range project.tree.Groups {
		if core.NormalizeRemoteConfigGroupKey(group.Key) == "" {
			seenRoot = true
		}
		groups = append(groups, messages.ParameterGroupOption{Key: group.Key, Label: group.Label})
	}
	if !seenRoot {
		groups = append([]messages.ParameterGroupOption{{Key: rootgroup.TreeKey, Label: rootgroup.Label}}, groups...)
	}
	paramKeys := make([]string, 0)
	for _, group := range project.tree.Groups {
		for _, param := range group.Parameters {
			paramKeys = append(paramKeys, param.Key)
		}
	}
	return groups, paramKeys
}

func (m Model) CurrentParameterViewData() (*messages.ParameterViewData, bool) {
	return m.currentParameterViewData()
}

func (m Model) selectionChangedCmd(activate bool) tea.Cmd {
	if m.history {
		return nil
	}
	data, ok := m.currentParameterViewData()
	if !ok {
		if groupData, groupOK := m.currentGroupViewData(); groupOK {
			return func() tea.Msg {
				return messages.ParameterSelectionChangedMsg{GroupData: groupData, Activate: activate}
			}
		}
		return func() tea.Msg {
			return messages.ParameterSelectionChangedMsg{
				ResetScroll: true,
			}
		}
	}

	return func() tea.Msg {
		return messages.ParameterSelectionChangedMsg{
			Data:     data,
			Activate: activate,
		}
	}
}

func copyToClipboardCmd(text string) tea.Cmd {
	if text == "" {
		return nil
	}
	return func() tea.Msg {
		_ = clipboard.WriteAll(text)
		return nil
	}
}

func (m Model) CurrentParameterRef() (core.Project, string, string, bool) {
	projectID, groupKey, paramKey, ok := m.currentParameterRef()
	if !ok {
		return core.Project{}, "", "", false
	}
	project := m.projectByID(projectID)
	if project == nil {
		return core.Project{}, "", "", false
	}
	return project.project, groupKey, paramKey, true
}

func (m *Model) FocusCurrentParameterDefaultValue() bool {
	return m.focusCurrentParameterDefaultValue()
}

func (m Model) CurrentGroupRef() (core.Project, string, string, bool) {
	if m.cursor < 0 || m.cursor >= len(m.visible) {
		return core.Project{}, "", "", false
	}
	node := m.visible[m.cursor]
	if node.kind != nodeGroup || node.transient || core.NormalizeRemoteConfigGroupKey(node.groupKey) == "" {
		return core.Project{}, "", "", false
	}
	project := m.projectByID(node.projectID)
	if project == nil {
		return core.Project{}, "", "", false
	}
	return project.project, node.groupKey, node.label, true
}

func (m Model) HasGroup(projectID, groupKey string) bool {
	return m.groupByKey(projectID, groupKey) != nil
}

func (m Model) CurrentNewParameterTarget() (core.Project, string, string, bool) {
	if m.cursor < 0 || m.cursor >= len(m.visible) {
		return core.Project{}, "", "", false
	}
	node := m.visible[m.cursor]
	project := m.projectByID(node.projectID)
	if project == nil {
		return core.Project{}, "", "", false
	}
	groupKey := rootgroup.TreeKey
	afterParamKey := ""
	switch node.kind {
	case nodeGroup:
		groupKey = node.groupKey
	case nodeParameter:
		groupKey = node.groupKey
		if !node.transient {
			afterParamKey = node.paramKey
		}
	case nodeValue:
		groupKey = node.groupKey
		afterParamKey = node.paramKey
	case nodeProject:
		groupKey = rootgroup.TreeKey
	default:
		if node.groupKey != "" {
			groupKey = node.groupKey
		}
	}
	return project.project, groupKey, afterParamKey, true
}

func (m Model) currentParameterNodeIndex() int {
	if m.cursor < 0 || m.cursor >= len(m.visible) {
		return -1
	}
	node := m.visible[m.cursor]
	switch node.kind {
	case nodeParameter:
		return m.cursor
	case nodeValue:
		for i := m.cursor - 1; i >= 0; i-- {
			prev := m.visible[i]
			if prev.projectID != node.projectID || prev.groupKey != node.groupKey || prev.paramKey != node.paramKey {
				break
			}
			if prev.kind == nodeParameter {
				return i
			}
		}
	}
	return -1
}
