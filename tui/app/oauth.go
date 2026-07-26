package app

import (
	"context"
	"errors"

	tea "charm.land/bubbletea/v2"
	"github.com/atotto/clipboard"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/browser"
	dialogcmp "github.com/yumauri/fbrcm/tui/components/dialog"
)

type oauthAuthorizationSession struct {
	flowID uint64
	authID string
	url    string
	cancel func()
}

type oauthLinkAction int

const (
	oauthLinkOpen oauthLinkAction = iota + 1
	oauthLinkCopy
)

type oauthLinkActionCompletedMsg struct {
	flowID uint64
	action oauthLinkAction
	err    error
}

type oauthAuthorizationCanceledMsg struct {
	flowID uint64
}

func (m Model) waitOAuthAuthorizationEventCmd() tea.Cmd {
	if m.oauthEvents == nil {
		return nil
	}
	return func() tea.Msg {
		return <-m.oauthEvents
	}
}

func (m Model) updateOAuthAuthorizationEvent(event core.OAuthAuthorizationEvent) (Model, tea.Cmd) {
	waitCmd := m.waitOAuthAuthorizationEventCmd()
	if event.Done {
		if m.oauthSession == nil || m.oauthSession.flowID != event.FlowID {
			return m, waitCmd
		}
		m.oauthDialog = m.oauthDialog.Close()
		m.oauthSession = nil
		if event.Err != nil && !errors.Is(event.Err, context.Canceled) {
			m.oauthDialog = m.oauthDialog.Open(dialogcmp.Config{
				Title: "Authentication Failed",
				Body:  []string{event.Err.Error()},
				Buttons: []dialogcmp.Button{{
					Label:   "Close",
					Variant: dialogcmp.ButtonVariantAccent,
				}},
			})
		}
		return m, waitCmd
	}
	if event.URL == "" {
		return m, waitCmd
	}
	m.oauthSession = &oauthAuthorizationSession{
		flowID: event.FlowID,
		authID: event.AuthID,
		url:    event.URL,
		cancel: event.Cancel,
	}
	m.openOAuthAuthorizationDialog("")
	return m, waitCmd
}

func (m *Model) openOAuthAuthorizationDialog(actionStatus string) {
	if m.oauthSession == nil {
		return
	}
	session := *m.oauthSession
	body := []string{
		"Browser authorization is required for authentication " + session.authID + ".",
		"",
		"Open this link:",
		session.url,
		"",
		"Waiting for the local OAuth callback…",
	}
	if actionStatus != "" {
		body = append(body, "", actionStatus)
	}
	m.oauthDialog = m.oauthDialog.Open(dialogcmp.Config{
		Title: "Authorize Firebase",
		Body:  body,
		Buttons: []dialogcmp.Button{
			{
				Label:   "Open Browser",
				Variant: dialogcmp.ButtonVariantAccent,
				OnPress: oauthLinkActionCmd(session.flowID, session.url, oauthLinkOpen),
			},
			{
				Label:   "Copy Link",
				Variant: dialogcmp.ButtonVariantNeutral,
				OnPress: oauthLinkActionCmd(session.flowID, session.url, oauthLinkCopy),
			},
			{
				Label:   "Cancel",
				Variant: dialogcmp.ButtonVariantNeutral,
				OnPress: func() tea.Msg { return oauthAuthorizationCanceledMsg{flowID: session.flowID} },
			},
		},
	})
}

func oauthLinkActionCmd(flowID uint64, url string, action oauthLinkAction) tea.Cmd {
	return func() tea.Msg {
		var err error
		if action == oauthLinkOpen {
			err = browser.OpenURL(url)
		} else {
			err = clipboard.WriteAll(url)
		}
		return oauthLinkActionCompletedMsg{flowID: flowID, action: action, err: err}
	}
}

func (m Model) updateOAuthLinkAction(msg oauthLinkActionCompletedMsg) (Model, tea.Cmd) {
	if m.oauthSession == nil || m.oauthSession.flowID != msg.flowID {
		return m, nil
	}
	status := "Link copied to clipboard."
	if msg.action == oauthLinkOpen {
		status = "Browser opened. Complete authorization there."
	}
	if msg.err != nil {
		status = "Action failed: " + msg.err.Error()
	}
	m.openOAuthAuthorizationDialog(status)
	return m, nil
}

func (m Model) cancelOAuthAuthorization(msg oauthAuthorizationCanceledMsg) (Model, tea.Cmd) {
	if m.oauthSession == nil || m.oauthSession.flowID != msg.flowID {
		return m, nil
	}
	cancel := m.oauthSession.cancel
	m.oauthSession = nil
	m.oauthDialog = m.oauthDialog.Close()
	if cancel != nil {
		cancel()
	}
	return m, nil
}
