package parameters

import (
	"charm.land/lipgloss/v2"
	"github.com/yumauri/fbrcm/core"
)

func (m *Model) syncVisible() {
	m.visible = m.buildVisible()
	if len(m.visible) == 0 {
		m.lineIndexByNode = nil
		m.cursor, m.offset, m.totalLines = 0, 0, 0
		return
	}
	m.cursor = max(0, min(m.cursor, len(m.visible)-1))
	m.recomputeLineLayout()
	m.ensureCursorVisible()
}

func (m *Model) recomputeLineLayout() {
	m.lineIndexByNode = make([]int, len(m.visible))
	line := 0
	for i := range m.visible {
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
	width := 0
	for _, project := range m.projects {
		if project.tree == nil {
			continue
		}
		for _, group := range project.tree.Groups {
			for _, param := range group.Parameters {
				width = max(width, lipgloss.Width(param.Key))
			}
		}
	}
	return width
}

func (m Model) LongestParameterNameWidth() int { return m.maxParameterNameWidth() }

func (m Model) filteredParameterCount() int {
	count := 0
	for _, node := range m.visible {
		if node.kind == nodeParameter {
			count++
		}
	}
	return count
}

func (m Model) CurrentProject() (core.Project, bool) { return m.currentProject() }
