package app

import (
	"fmt"
	"slices"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/tui/components/authpicker"
	tuiconfig "github.com/yumauri/fbrcm/tui/config"
)

type authBindingSession struct {
	targets []core.Project
}

type projectAuthBoundMsg struct {
	projects []core.Project
	err      error
}

func (m *Model) openProjectAuthPicker() tea.Cmd {
	targets := m.projects.ActionTargets()
	if len(targets) == 0 {
		return nil
	}
	entries, _, err := m.svc.ListAuth()
	if err != nil {
		m.openErrorDialog("Bind Authentication Failed", targets[0], err.Error())
		return nil
	}
	if len(entries) <= 1 {
		return nil
	}

	selected := 0
	currentID := targets[0].AuthID
	sameCurrent := true
	options := make([]authpicker.Option, 0, len(entries))
	for index, entry := range entries {
		verified := 0
		for _, project := range targets {
			if slices.Contains(project.DiscoveredBy, entry.ID) {
				verified++
			}
		}
		detail := authTypeName(entry.Type)
		if verified == len(targets) {
			detail += "  ·  verified"
		} else {
			detail += fmt.Sprintf("  ·  verified %d/%d", verified, len(targets))
		}
		options = append(options, authpicker.Option{Key: entry.ID, Label: entry.ID, Detail: detail})
		if entry.ID == currentID {
			selected = index
		}
	}
	for _, project := range targets[1:] {
		if project.AuthID != currentID {
			sameCurrent = false
			break
		}
	}
	if !sameCurrent {
		selected = 0
	}

	body := make([]string, 0, len(targets)+1)
	if len(targets) == 1 {
		body = append(body, "Project: "+targets[0].ProjectID)
	} else {
		body = append(body, fmt.Sprintf("Projects (%d):", len(targets)))
		for _, project := range targets {
			body = append(body, "  "+project.ProjectID)
		}
	}
	m.closeOverlays()
	m.authBind = &authBindingSession{targets: append([]core.Project(nil), targets...)}
	m.authPicker = m.authPicker.SetBounds(0, 0, m.width, m.height).Open("Bind authentication", body, options, selected)
	return nil
}

func (m Model) updateAuthPicker(msg tea.Msg) (Model, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		k := msg.String()
		switch {
		case tuiconfig.Matches(tuiconfig.BlockAuthPicker, tuiconfig.ActionCancel, k):
			m.closeAuthPicker()
		case tuiconfig.Matches(tuiconfig.BlockAuthPicker, tuiconfig.ActionSubmit, k):
			return m, m.submitAuthBinding(), true
		case tuiconfig.Matches(tuiconfig.BlockAuthPicker, tuiconfig.ActionUp, k):
			m.authPicker.Move(-1)
		case tuiconfig.Matches(tuiconfig.BlockAuthPicker, tuiconfig.ActionDown, k):
			m.authPicker.Move(1)
		}
		return m, nil, true
	case tea.MouseClickMsg, tea.MouseMotionMsg, tea.MouseWheelMsg, tea.MouseReleaseMsg:
		return m, nil, true
	}
	return m, nil, false
}

func (m *Model) submitAuthBinding() tea.Cmd {
	option, ok := m.authPicker.Current()
	session := m.authBind
	m.authPicker = m.authPicker.Close()
	m.authBind = nil
	if !ok || session == nil {
		return nil
	}
	projectIDs := make([]string, 0, len(session.targets))
	for _, project := range session.targets {
		projectIDs = append(projectIDs, project.ProjectID)
	}
	authID := option.Key
	return func() tea.Msg {
		projects, err := m.svc.BindProjectIDsAuth(projectIDs, authID)
		return projectAuthBoundMsg{projects: projects, err: err}
	}
}

func (m Model) updateProjectAuthBound(msg projectAuthBoundMsg) (Model, tea.Cmd, bool) {
	if msg.err != nil {
		m.openErrorDialog("Bind Authentication Failed", core.Project{}, msg.err.Error())
		return m, nil, true
	}
	cmd := m.projects.ApplyProjectUpdates(msg.projects)
	return m, cmd, true
}

func (m *Model) closeAuthPicker() {
	m.authPicker = m.authPicker.Close()
	m.authBind = nil
}

func authTypeName(value string) string {
	switch value {
	case config.AuthTypeOAuth:
		return "OAuth"
	case config.AuthTypeServiceAccount:
		return "Service account"
	case config.AuthTypeGCloud:
		return "gcloud ADC"
	default:
		return value
	}
}
