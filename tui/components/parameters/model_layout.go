package parameters

import (
	"charm.land/lipgloss/v2"
	"github.com/yumauri/fbrcm/core"
)

func (m *Model) syncVisible() {
	m.parameterNameWidth = m.computeMaxParameterNameWidth()
	m.visible = m.buildVisible()
	m.visibleParamCount = 0
	for _, node := range m.visible {
		if node.kind == nodeParameter {
			m.visibleParamCount++
		}
	}
	if len(m.visible) == 0 {
		m.lineIndexByNode = nil
		m.projectNodeFor, m.groupNodeFor = nil, nil
		m.cursor, m.offset, m.totalLines = 0, 0, 0
		return
	}
	m.cursor = max(0, min(m.cursor, len(m.visible)-1))
	m.recomputeLineLayout()
	m.ensureCursorVisible()
}

func (m *Model) recomputeLineLayout() {
	m.lineIndexByNode = make([]int, len(m.visible))
	m.projectNodeFor = make([]int, len(m.visible))
	m.groupNodeFor = make([]int, len(m.visible))
	line := 0
	projectIndex, groupIndex := -1, -1
	for i := range m.visible {
		if m.visible[i].kind == nodeProject {
			projectIndex, groupIndex = i, -1
		}
		if m.visible[i].kind == nodeGroup {
			groupIndex = i
		}
		m.projectNodeFor[i], m.groupNodeFor[i] = projectIndex, groupIndex
		m.lineIndexByNode[i] = line
		line += m.nodeBlockLineCount(i)
	}
	m.totalLines = line
}

func (m Model) contentHeight() int                         { return max(m.height-2-m.filter.Height(), 0) }
func (m Model) viewportWidth() int                         { return max(m.width-2, 1) }
func (m Model) viewportHeight() int                        { return max(m.height-2-m.filter.Height(), 1) }
func (m Model) groupKey(projectID, groupKey string) string { return projectID + "::group::" + groupKey }
func (m Model) paramKey(projectID, groupKey, paramKey string) string {
	return projectID + "::param::" + groupKey + "::" + paramKey
}

func (m Model) anyLoading() bool {
	for _, project := range m.projects {
		if project.loading || project.verifying {
			return true
		}
	}
	return false
}

func (m Model) parameterRenderLayout() parameterRenderLayout {
	layout := parameterRenderLayout{mode: parameterRenderModeRegular, paramStart: 2, nameWidth: m.maxParameterNameWidth()}
	layout.valueStart = layout.paramStart + layout.nameWidth + 3
	layout.valueWidth = max(m.viewportWidth()-layout.valueStart, 0)
	if layout.valueWidth < 10 {
		layout.mode = parameterRenderModeNarrow
	}
	return layout
}

func (m Model) maxParameterNameWidth() int {
	return m.parameterNameWidth
}

func (m Model) computeMaxParameterNameWidth() int {
	width := 0
	for _, project := range m.projects {
		tree := project.tree
		if tree == nil {
			continue
		}
		for _, group := range tree.Groups {
			for _, param := range group.Parameters {
				width = max(width, lipgloss.Width(param.Key))
			}
		}
	}
	return width
}

func (m Model) LongestParameterNameWidth() int { return m.maxParameterNameWidth() }

func (m Model) filteredParameterCount() int {
	return m.visibleParamCount
}

func (m Model) CurrentProject() (core.Project, bool) { return m.currentProject() }
