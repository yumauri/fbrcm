package setup

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/browser"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/firebase"
)

const oauthClientsURL = "https://console.cloud.google.com/auth/clients"

type inspectedMsg struct {
	state core.StartupState
	err   error
}

type authAddedMsg struct {
	entry config.AuthEntry
	err   error
}

type authReadyMsg struct {
	loginID uint64
	err     error
}

type projectsSyncedMsg struct {
	projects []core.Project
	source   string
	syncID   uint64
	err      error
}

type externalOpenedMsg struct{ err error }

type profileSwitchedMsg struct{ err error }

func (m Model) inspectCmd() tea.Cmd {
	return func() tea.Msg {
		state, err := m.svc.InspectStartupState()
		return inspectedMsg{state: state, err: err}
	}
}

func (m Model) addAuthCmd() tea.Cmd {
	method := m.method
	authID := m.authID
	path := m.filePath
	return func() tea.Msg {
		var (
			entry config.AuthEntry
			err   error
		)
		switch method {
		case methodOAuth, methodServiceAccount:
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				return authAddedMsg{err: fmt.Errorf("read selected file: %w", readErr)}
			}
			if !json.Valid(data) {
				return authAddedMsg{err: fmt.Errorf("selected file is not valid JSON")}
			}
			if method == methodOAuth {
				entry, err = m.svc.AddOAuthAuth(authID, "", data)
			} else {
				entry, err = m.svc.AddServiceAccountAuth(authID, "", data)
			}
		case methodGCloud:
			entry, err = m.svc.AddGCloudAuth(authID, "")
		default:
			err = fmt.Errorf("unsupported authentication method")
		}
		return authAddedMsg{entry: entry, err: err}
	}
}

func (m *Model) startLogin(back mode) tea.Cmd {
	m.cancelLogin()
	ctx, cancel := context.WithCancel(context.Background())
	ctx = firebase.WithOAuthTerminalOutput(ctx, false)
	m.loginStop = cancel
	m.loginID++
	loginID := m.loginID
	authID := m.authID
	m.loginBack = back
	return func() tea.Msg {
		err := m.svc.EnsureAuthLogin(ctx, authID, false)
		return authReadyMsg{loginID: loginID, err: err}
	}
}

func (m *Model) cancelLogin() {
	if m.loginStop != nil {
		m.loginStop()
		m.loginStop = nil
		m.loginID++
	}
}

func (m *Model) startSync(back mode) tea.Cmd {
	m.cancelSync()
	ctx, cancel := context.WithCancel(context.Background())
	ctx = firebase.WithOAuthTerminalOutput(ctx, false)
	m.syncStop = cancel
	m.syncID++
	syncID := m.syncID
	authID := m.syncAuthID
	m.syncBack = back
	return func() tea.Msg {
		if authID == "" {
			projects, source, err := m.svc.SyncProjects(ctx)
			return projectsSyncedMsg{projects: projects, source: source, syncID: syncID, err: err}
		}
		projects, source, err := m.svc.SyncProjectsForAuth(ctx, authID)
		return projectsSyncedMsg{projects: projects, source: source, syncID: syncID, err: err}
	}
}

func (m *Model) cancelSync() {
	if m.syncStop != nil {
		m.syncStop()
		m.syncStop = nil
		m.syncID++
	}
}

func openOAuthClientsCmd() tea.Cmd {
	return func() tea.Msg {
		return externalOpenedMsg{err: browser.OpenURL(oauthClientsURL)}
	}
}

func (m Model) switchProfileCmd() tea.Cmd {
	profile := m.profileTo
	return func() tea.Msg {
		return profileSwitchedMsg{err: m.svc.SwitchProfile(profile)}
	}
}

func clearScreenCmd() tea.Cmd {
	return func() tea.Msg { return tea.ClearScreen() }
}
