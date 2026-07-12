package parameters

import (
	"fmt"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/filter"
	"github.com/yumauri/fbrcm/core/rc/diff"
	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
	"github.com/yumauri/fbrcm/core/rootgroup"
)

func (m Model) buildVisible() []visibleNode {
	nodes := make([]visibleNode, 0)
	query := m.filter.Value()
	filtering := query != ""
	for _, project := range m.projects {
		tree := project.tree
		if m.history {
			tree = m.historyTree(project.project.ProjectID, tree)
		}
		nodes = append(nodes, visibleNode{
			kind:      nodeProject,
			projectID: project.project.ProjectID,
			label:     rcdisplay.FormatProject(project.project.Name, project.project.ProjectID),
			expanded:  true,
		})

		if project.loading {
			nodes = append(nodes, visibleNode{
				kind:      nodeValue,
				projectID: project.project.ProjectID,
				label:     "Loading parameters...",
			})
			continue
		}
		if project.err != nil && project.tree == nil {
			nodes = append(nodes, visibleNode{
				kind:      nodeValue,
				projectID: project.project.ProjectID,
				label:     fmt.Sprintf("Load failed: %v", project.err),
			})
			continue
		}
		if tree == nil || len(tree.Groups) == 0 {
			if created := m.transientNew; created != nil && created.projectID == project.project.ProjectID {
				nodes = appendTransientNewRootGroup(nodes, project.project.ProjectID, created)
				continue
			}
			nodes = append(nodes, visibleNode{
				kind:      nodeValue,
				projectID: project.project.ProjectID,
				label:     "No parameters",
			})
			continue
		}

		transientRootShown := false
		for _, group := range tree.Groups {
			if created := m.transientNew; created != nil &&
				created.projectID == project.project.ProjectID &&
				core.NormalizeRemoteConfigGroupKey(created.groupKey) == "" &&
				core.NormalizeRemoteConfigGroupKey(group.Key) == "" {
				transientRootShown = true
			}
			matchedParams := group.Parameters
			if filtering {
				matchedParams = matchedParameters(group.Parameters, query, m.filter.Mode())
				created := m.transientNew
				hasTransientNew := created != nil &&
					created.projectID == project.project.ProjectID &&
					core.NormalizeRemoteConfigGroupKey(created.groupKey) == core.NormalizeRemoteConfigGroupKey(group.Key)
				if len(matchedParams) == 0 && !hasTransientNew {
					continue
				}
			}
			groupExpanded := m.groupExpanded[m.groupKey(project.project.ProjectID, group.Key)]
			nodes = append(nodes, visibleNode{
				kind:      nodeGroup,
				projectID: project.project.ProjectID,
				groupKey:  group.Key,
				label:     group.Label,
				summary:   fmt.Sprintf("%d", len(matchedParams)),
				expanded:  groupExpanded,
			})
			if !groupExpanded {
				continue
			}

			for _, param := range matchedParams {
				paramExpanded := m.paramExpanded[m.paramKey(project.project.ProjectID, group.Key, param.Key)]
				nodes = append(nodes, visibleNode{
					kind:      nodeParameter,
					projectID: project.project.ProjectID,
					groupKey:  group.Key,
					paramKey:  param.Key,
					label:     param.Key,
					summary:   param.Summary,
					expanded:  paramExpanded,
				})
				if dup := m.transientDup; dup != nil &&
					dup.projectID == project.project.ProjectID &&
					dup.groupKey == group.Key &&
					dup.afterParamKey == param.Key &&
					(!filtering || matchedDuplicate(dup.label, query, m.filter.Mode())) {
					nodes = append(nodes, visibleNode{
						kind:      nodeParameter,
						projectID: project.project.ProjectID,
						groupKey:  group.Key,
						paramKey:  param.Key,
						label:     dup.label,
						summary:   param.Summary,
						expanded:  false,
						transient: true,
					})
				}
				if created := m.transientNew; created != nil &&
					created.projectID == project.project.ProjectID &&
					core.NormalizeRemoteConfigGroupKey(created.groupKey) == core.NormalizeRemoteConfigGroupKey(group.Key) &&
					created.afterParamKey == param.Key {
					nodes = append(nodes, visibleNode{
						kind:      nodeParameter,
						projectID: project.project.ProjectID,
						groupKey:  group.Key,
						paramKey:  "",
						label:     created.label,
						summary:   "new parameter",
						expanded:  false,
						transient: true,
					})
				}
				if !paramExpanded {
					continue
				}

				for i, value := range param.Values {
					nodes = append(nodes, visibleNode{
						kind:      nodeValue,
						projectID: project.project.ProjectID,
						groupKey:  group.Key,
						paramKey:  param.Key,
						valueIdx:  i,
						label:     value.Label,
						summary:   value.Value,
					})
				}
			}
			if created := m.transientNew; created != nil &&
				created.projectID == project.project.ProjectID &&
				core.NormalizeRemoteConfigGroupKey(created.groupKey) == core.NormalizeRemoteConfigGroupKey(group.Key) &&
				created.afterParamKey == "" {
				nodes = append(nodes, visibleNode{
					kind:      nodeParameter,
					projectID: project.project.ProjectID,
					groupKey:  group.Key,
					paramKey:  "",
					label:     created.label,
					summary:   "new parameter",
					expanded:  false,
					transient: true,
				})
			}
		}
		if created := m.transientNew; created != nil &&
			created.projectID == project.project.ProjectID &&
			core.NormalizeRemoteConfigGroupKey(created.groupKey) == "" &&
			!transientRootShown {
			nodes = appendTransientNewRootGroup(nodes, project.project.ProjectID, created)
		}
	}

	if m.history && m.historyChangesOnly {
		return m.historyChangesOnlyNodes(nodes)
	}
	return nodes
}

func (m Model) historyChangesOnlyNodes(nodes []visibleNode) []visibleNode {
	changedParams := make(map[string]bool)
	changedGroups := make(map[string]int)
	changedProjects := make(map[string]bool)
	statusProjects := make(map[string]bool)
	for _, project := range m.projects {
		// Project rows remain as navigation anchors even when the active filters
		// remove every group and parameter beneath them. History version actions
		// are scoped through the project at the cursor.
		changedProjects[project.project.ProjectID] = true
		state := m.histories[project.project.ProjectID]
		if project.loading || project.verifying || project.err != nil || state.loading || state.err != nil {
			statusProjects[project.project.ProjectID] = true
		}
	}
	for _, node := range nodes {
		if node.kind != nodeParameter || node.transient {
			continue
		}
		if !isVisibleHistoryChange(m.histories[node.projectID].paramKinds[historyParamKey(node.groupKey, node.paramKey)]) {
			continue
		}
		paramKey := historyVisibleParamKey(node.projectID, node.groupKey, node.paramKey)
		groupKey := historyVisibleGroupKey(node.projectID, node.groupKey)
		changedParams[paramKey] = true
		changedGroups[groupKey]++
		changedProjects[node.projectID] = true
	}

	out := make([]visibleNode, 0, len(nodes))
	for _, node := range nodes {
		switch node.kind {
		case nodeProject:
			if changedProjects[node.projectID] {
				out = append(out, node)
			}
		case nodeGroup:
			if count := changedGroups[historyVisibleGroupKey(node.projectID, node.groupKey)]; count > 0 {
				node.summary = fmt.Sprintf("%d", count)
				out = append(out, node)
			}
		case nodeParameter:
			if changedParams[historyVisibleParamKey(node.projectID, node.groupKey, node.paramKey)] {
				out = append(out, node)
			}
		case nodeValue:
			if node.paramKey == "" {
				if statusProjects[node.projectID] {
					out = append(out, node)
				}
				continue
			}
			if changedParams[historyVisibleParamKey(node.projectID, node.groupKey, node.paramKey)] &&
				isVisibleHistoryChange(m.historyValueKind(node.projectID, node.groupKey, node.paramKey, node.label)) {
				out = append(out, node)
			}
		}
	}
	return out
}

func isVisibleHistoryChange(kind diff.ChangeKind) bool {
	return kind == diff.ChangeAdded || kind == diff.ChangeRemoved || kind == diff.ChangeChanged
}

func historyVisibleGroupKey(projectID, groupKey string) string {
	return projectID + "\x00" + groupKey
}

func historyVisibleParamKey(projectID, groupKey, paramKey string) string {
	return projectID + "\x00" + groupKey + "\x00" + paramKey
}

func appendTransientNewRootGroup(nodes []visibleNode, projectID string, created *transientNewParameter) []visibleNode {
	nodes = append(nodes, visibleNode{
		kind:      nodeGroup,
		projectID: projectID,
		groupKey:  rootgroup.TreeKey,
		label:     rootgroup.Label,
		summary:   "0",
		expanded:  true,
	})
	return append(nodes, visibleNode{
		kind:      nodeParameter,
		projectID: projectID,
		groupKey:  rootgroup.TreeKey,
		paramKey:  "",
		label:     created.label,
		summary:   "new parameter",
		expanded:  false,
		transient: true,
	})
}

func matchedParameters(params []core.ParametersEntry, query string, mode filter.Mode) []core.ParametersEntry {
	if query == "" {
		return params
	}
	out := make([]core.ParametersEntry, 0, len(params))
	for _, param := range params {
		if ok, _ := filter.Match(param.Key, query, mode); ok {
			out = append(out, param)
		}
	}
	return out
}

func matchedDuplicate(label, query string, mode filter.Mode) bool {
	if query == "" {
		return true
	}
	ok, _ := filter.Match(label, query, mode)
	return ok
}
