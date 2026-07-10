package parameters

import (
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/rootgroup"
)

func (m Model) CurrentMoveAnchor() (MoveAnchor, bool) {
	if m.cursor < 0 || m.cursor >= len(m.visible) {
		return MoveAnchor{}, false
	}
	node := m.visible[m.cursor]
	project := m.projectByID(node.projectID)
	if project == nil || project.tree == nil {
		return MoveAnchor{}, false
	}

	switch node.kind {
	case nodeGroup:
		if node.transient {
			return MoveAnchor{}, false
		}
		screenLine := m.screenLineForOffset(m.cursor, m.offset)
		if screenLine < 0 {
			return MoveAnchor{}, false
		}
		currentNormalized := core.NormalizeRemoteConfigGroupKey(node.groupKey)
		options := make([]MoveOption, 0, len(project.tree.Groups))
		for _, group := range project.tree.Groups {
			groupNormalized := core.NormalizeRemoteConfigGroupKey(group.Key)
			if groupNormalized == currentNormalized {
				continue
			}
			if groupNormalized == "" {
				continue
			}
			options = append(options, MoveOption{Key: group.Key, Label: group.Label})
		}
		if currentNormalized != "" {
			options = append(options, MoveOption{Key: "", Label: rootgroup.Label})
		}
		if len(options) == 0 {
			return MoveAnchor{}, false
		}
		return MoveAnchor{
			Project:  project.project,
			IsGroup:  true,
			GroupKey: node.groupKey,
			Label:    node.label,
			X:        m.x + 1,
			Y:        m.y + screenLine,
			Options:  options,
		}, true
	case nodeParameter, nodeValue:
		_, groupKey, paramKey, ok := m.currentParameterRef()
		if !ok {
			return MoveAnchor{}, false
		}
		paramIndex := m.currentParameterNodeIndex()
		if paramIndex < 0 {
			return MoveAnchor{}, false
		}
		screenLine := m.screenLineForOffset(paramIndex, m.offset)
		if screenLine < 0 {
			return MoveAnchor{}, false
		}
		options := make([]MoveOption, 0, len(project.tree.Groups)+1)
		currentNormalized := core.NormalizeRemoteConfigGroupKey(groupKey)
		for _, group := range project.tree.Groups {
			groupNormalized := core.NormalizeRemoteConfigGroupKey(group.Key)
			if groupNormalized == "" || groupNormalized == currentNormalized {
				continue
			}
			options = append(options, MoveOption{Key: group.Key, Label: group.Label})
		}
		if currentNormalized != "" {
			options = append(options, MoveOption{Key: "", Label: rootgroup.Label})
		}
		layout := m.parameterRenderLayout()
		return MoveAnchor{
			Project:  project.project,
			IsGroup:  false,
			GroupKey: groupKey,
			ParamKey: paramKey,
			Label:    paramKey,
			X:        m.x + layout.paramStart - 1,
			Y:        m.y + screenLine,
			Options:  options,
		}, true
	default:
		return MoveAnchor{}, false
	}
}
