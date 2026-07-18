package projects

import (
	"fmt"

	"github.com/yumauri/fbrcm/core/filter"
)

type lineKind int

const (
	lineKindPlain lineKind = iota
	lineKindProjectName
	lineKindProjectID
	lineKindProjectSpacer
	lineKindMeta
)

func (m *Model) contentLines() []string {
	var lines []string
	var lineKinds []lineKind
	var lineProjects []int
	var lineHighlights [][]int
	var projectStarts []int
	var projectEnds []int
	if m.notice != "" {
		lines = append(lines, m.notice, "")
		lineKinds = append(lineKinds, lineKindMeta, lineKindProjectSpacer)
		lineProjects = append(lineProjects, -1, -1)
		lineHighlights = append(lineHighlights, nil, nil)
	}

	if m.err != nil && len(m.projects) == 0 {
		lines = append(lines, fmt.Sprintf("Load failed: %v", m.err))
		lineKinds = append(lineKinds, lineKindPlain)
		lineProjects = append(lineProjects, -1)
		lineHighlights = append(lineHighlights, nil)
	} else if len(m.projects) == 0 {
		if m.loading {
			lines = append(lines, "Loading projects...")
		} else if len(m.allProjects) == 0 && m.filter.Value() == "" {
			lines = append(lines, "No projects configured")
		} else {
			lines = append(lines, "No matching projects")
		}
		lineKinds = append(lineKinds, lineKindPlain)
		lineProjects = append(lineProjects, -1)
		lineHighlights = append(lineHighlights, nil)
	} else {
		for i, project := range m.projects {
			_, nameHighlights := filter.Match(project.Name, m.filter.Value(), m.filter.Mode())
			_, idHighlights := filter.Match(project.ProjectID, m.filter.Value(), m.filter.Mode())
			idHighlights = offsetIndices(idHighlights, 1)

			projectStarts = append(projectStarts, len(lines))
			lines = append(lines, project.Name)
			lineKinds = append(lineKinds, lineKindProjectName)
			lineProjects = append(lineProjects, i)
			lineHighlights = append(lineHighlights, nameHighlights)
			lines = append(lines, " "+project.ProjectID)
			lineKinds = append(lineKinds, lineKindProjectID)
			lineProjects = append(lineProjects, i)
			lineHighlights = append(lineHighlights, idHighlights)

			projectEnds = append(projectEnds, len(lines)-1)
			if i < len(m.projects)-1 {
				lines = append(lines, "")
				lineKinds = append(lineKinds, lineKindProjectSpacer)
				lineProjects = append(lineProjects, -1)
				lineHighlights = append(lineHighlights, nil)
			}
		}
	}

	m.lines = lines
	m.lineKinds = lineKinds
	m.lineProjects = lineProjects
	m.lineHighlights = lineHighlights
	m.projectStarts = projectStarts
	m.projectEnds = projectEnds

	return m.lines
}

func (m Model) View(active bool) string {
	return m.ViewWithBorder(active, active)
}

func (m Model) ViewWithBorder(active, borderActive bool) string {
	if m.collapsed {
		return renderCollapsedPanel(m.height, active, borderActive)
	}
	return renderPanel(
		m.viewport.View(),
		m.width,
		m.height,
		active,
		borderActive,
		m.scrollbar(),
		m.secondaryTitle(),
		m.filter.View(m.width, active, len(m.projects)),
	)
}

func (m *Model) applyFilter() {
	currentID := ""
	if m.cursor >= 0 && m.cursor < len(m.projects) {
		currentID = m.projects[m.cursor].ProjectID
	}

	query := m.filter.Value()
	m.projects = m.projects[:0]
	for _, project := range m.allProjects {
		nameMatch, _ := filter.Match(project.Name, query, m.filter.Mode())
		idMatch, _ := filter.Match(project.ProjectID, query, m.filter.Mode())
		if nameMatch || idMatch {
			m.projects = append(m.projects, project)
		}
	}

	m.cursor = 0
	if currentID != "" {
		for i, project := range m.projects {
			if project.ProjectID == currentID {
				m.cursor = i
				break
			}
		}
	}
	if len(m.projects) == 0 {
		m.cursor = 0
		return
	}
	m.cursor = max(0, min(m.cursor, len(m.projects)-1))
}

func offsetIndices(indices []int, offset int) []int {
	if len(indices) == 0 {
		return nil
	}
	shifted := make([]int, len(indices))
	for i, index := range indices {
		shifted[i] = index + offset
	}
	return shifted
}

func indicesSet(indices []int) map[int]bool {
	set := make(map[int]bool, len(indices))
	for _, index := range indices {
		set[index] = true
	}
	return set
}
