package projects

import (
	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	corelog "github.com/yumauri/fbrcm/core/log"
	rctarget "github.com/yumauri/fbrcm/core/rc/target"
	"github.com/yumauri/fbrcm/tui/messages"
)

func (m *Model) rebuildTargets() bool {
	targets := make([]core.Project, 0, len(m.baseProjects)*2)
	for _, project := range m.baseProjects {
		for _, kind := range project.TemplateKinds() {
			targets = append(targets, project.TemplateTarget(kind))
		}
	}
	m.allProjects = targets
	selectionChanged := m.dropMissingSelections()
	m.applyFilter()
	return selectionChanged
}

func (m Model) baseProject(projectID string) (core.Project, bool) {
	for _, project := range m.baseProjects {
		if project.ProjectID == projectID {
			return project, true
		}
	}
	return core.Project{}, false
}

func (m *Model) replaceBaseProject(project core.Project) {
	for i := range m.baseProjects {
		if m.baseProjects[i].ProjectID == project.ProjectID {
			m.baseProjects[i] = project
			return
		}
	}
}

func (m *Model) dropMissingSelections() bool {
	available := make(map[string]struct{}, len(m.allProjects))
	for _, project := range m.allProjects {
		available[project.ProjectID] = struct{}{}
	}
	changed := false
	for projectID := range m.selected {
		if _, ok := available[projectID]; !ok {
			delete(m.selected, projectID)
			changed = true
		}
	}
	return changed
}

func (m Model) toggleCurrentTemplatesCmd() tea.Cmd {
	current, ok := m.CurrentProject()
	if !ok {
		return nil
	}
	target, err := rctarget.Parse(current.ProjectID)
	if err != nil {
		return nil
	}
	project, ok := m.baseProject(target.ProjectID)
	if !ok {
		return nil
	}

	var templates []rctarget.Kind
	action := "collapsed"
	if len(project.TemplateKinds()) == 1 {
		templates = []rctarget.Kind{rctarget.Client, rctarget.Server}
		action = "expanded"
	} else {
		templates = []rctarget.Kind{target.Kind}
	}
	return m.setTemplatePreferencesCmd(project, templates, target.Kind, action)
}

func (m Model) makeCurrentPrimaryCmd() tea.Cmd {
	if !m.CanMakeCurrentPrimary() {
		return nil
	}
	current, _ := m.CurrentProject()
	target, _ := rctarget.Parse(current.ProjectID)
	project, _ := m.baseProject(target.ProjectID)
	return m.setTemplatePreferencesCmd(
		project,
		append([]rctarget.Kind(nil), project.Templates...),
		target.Kind,
		"primary",
	)
}

func (m Model) setTemplatePreferencesCmd(
	project core.Project,
	templates []rctarget.Kind,
	primary rctarget.Kind,
	action string,
) tea.Cmd {
	return func() tea.Msg {
		updated, err := m.svc.SetProjectTemplatePreferences(project.ProjectID, templates, primary)
		logger := corelog.For("tui.projects")
		if err != nil {
			logger.Error("update project templates failed", "project_id", project.ProjectID, "action", action, "err", err)
		} else {
			message := "project templates " + action
			if action == "primary" {
				message = "project primary template changed"
			}
			logger.Info(
				message,
				"project_id", project.ProjectID,
				"templates", templates,
				"primary", primary,
			)
		}
		return messages.ProjectTemplatePreferencesUpdatedMsg{
			Project: updated,
			Err:     err,
		}
	}
}
