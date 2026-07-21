package app

import (
	"fmt"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	dialogcmp "github.com/yumauri/fbrcm/tui/components/dialog"
	"github.com/yumauri/fbrcm/tui/components/viewutil"
)

type projectsDeletedMsg struct {
	projects []core.Project
	err      error
}

func (m *Model) requestDeleteProjects() tea.Cmd {
	targets := m.projects.ActionTargets()
	if len(targets) == 0 {
		return nil
	}

	body := make([]string, 0, len(targets)+4)
	if len(targets) == 1 {
		body = append(body, dialogProjectLine(targets[0]))
	} else {
		body = append(body, fmt.Sprintf("Projects (%d):", len(targets)))
		for _, project := range targets {
			body = append(body, "  "+viewutil.ProjectReference(project))
		}
	}
	body = append(body,
		"",
		"This removes the projects, cached configs, versions, and drafts",
		"from this profile. Firebase is not changed.",
	)
	title := "Delete Projects?"
	if len(targets) == 1 {
		title = "Delete Project?"
	}

	m.dialog = m.dialog.Open(dialogcmp.Config{
		Title: title,
		Body:  body,
		Buttons: []dialogcmp.Button{
			{Label: "Delete", Variant: dialogcmp.ButtonVariantDanger, OnPress: m.deleteProjectsCmd(targets)},
			{Label: "Cancel", Variant: dialogcmp.ButtonVariantAccent, OnPress: dialogCanceledCmd()},
		},
	})
	return nil
}

func (m Model) deleteProjectsCmd(projects []core.Project) tea.Cmd {
	projectIDs := make([]string, 0, len(projects))
	for _, project := range projects {
		projectIDs = append(projectIDs, project.ProjectID)
	}
	return func() tea.Msg {
		deleted, err := m.svc.DeleteProjectIDs(projectIDs)
		return projectsDeletedMsg{projects: deleted, err: err}
	}
}

func (m Model) updateProjectsDeleted(msg projectsDeletedMsg) (Model, tea.Cmd, bool) {
	if msg.err != nil {
		m.dialog = m.dialog.Open(dialogcmp.Config{
			Title:   "Delete Projects Failed",
			Body:    []string{msg.err.Error()},
			Buttons: []dialogcmp.Button{{Label: "Close", Variant: dialogcmp.ButtonVariantAccent, OnPress: dialogCanceledCmd()}},
		})
		return m, nil, true
	}
	return m, m.projects.RemoveProjects(msg.projects), true
}
