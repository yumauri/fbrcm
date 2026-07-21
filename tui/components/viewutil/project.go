package viewutil

import (
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/tui/styles"
)

// ProjectReference renders a project consistently in popup content.
func ProjectReference(project core.Project) string {
	return styles.TreeProjectName.Render(project.Name) + styles.PanelText.Render(" ("+project.ProjectID+")")
}

// ProjectLine renders the standard labeled project line used by popups.
func ProjectLine(project core.Project) string {
	return styles.PanelText.Render("Project: ") + ProjectReference(project)
}
