package app

import (
	"context"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/tui/panels"
)

func TestOAuthAuthorizationModalShowsLinkAndRestoresWorkspaceOnSuccess(t *testing.T) {
	m := viewTestModel(100, 30, panels.Projects)
	const authorizationURL = "https://accounts.example.test/authorize?client_id=fbrcm&state=secret"

	next, cmd := m.updateOAuthAuthorizationEvent(core.OAuthAuthorizationEvent{
		FlowID: 7,
		AuthID: "work",
		URL:    authorizationURL,
	})
	if cmd != nil {
		t.Fatal("OAuth event without an event channel returned a wait command")
	}
	if !next.oauthDialog.IsOpen() || next.oauthSession == nil {
		t.Fatalf("OAuth modal state = open:%v session:%#v", next.oauthDialog.IsOpen(), next.oauthSession)
	}
	dialogView := ansi.Strip(next.oauthDialog.View())
	for _, want := range []string{
		"Authorize Firebase",
		"authentication work",
		"https://accounts.example.test/authorize",
		"client_id=fbrcm&state=secret",
		"Open Browser",
		"Copy Link",
		"Cancel",
	} {
		if !strings.Contains(dialogView, want) {
			t.Fatalf("OAuth modal does not contain %q:\n%s", want, dialogView)
		}
	}
	if view := ansi.Strip(next.View().Content); !strings.Contains(view, "Projects") || !strings.Contains(view, "Authorize Firebase") {
		t.Fatalf("OAuth modal did not render over the workspace:\n%s", view)
	}

	next, cmd = next.updateOAuthAuthorizationEvent(core.OAuthAuthorizationEvent{FlowID: 7, Done: true})
	if cmd != nil {
		t.Fatal("OAuth completion without an event channel returned a wait command")
	}
	if next.oauthDialog.IsOpen() || next.oauthSession != nil {
		t.Fatalf("OAuth completion did not restore workspace: open=%v session=%#v", next.oauthDialog.IsOpen(), next.oauthSession)
	}
	if view := ansi.Strip(next.View().Content); !strings.Contains(view, "Projects") || strings.Contains(view, "Authorize Firebase") {
		t.Fatalf("workspace was not restored after OAuth:\n%s", view)
	}
}

func TestOAuthAuthorizationModalEscapeCancelsFlow(t *testing.T) {
	m := viewTestModel(100, 30, panels.Projects)
	canceled := false
	m, _ = m.updateOAuthAuthorizationEvent(core.OAuthAuthorizationEvent{
		FlowID: 8,
		AuthID: "default",
		URL:    "https://accounts.example.test/authorize",
		Cancel: func() { canceled = true },
	})

	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	next := updated.(Model)
	if cmd != nil {
		t.Fatal("canceling OAuth returned a command")
	}
	if !canceled {
		t.Fatal("Escape did not cancel the OAuth context")
	}
	if next.oauthDialog.IsOpen() || next.oauthSession != nil {
		t.Fatalf("Escape did not close OAuth modal: open=%v session=%#v", next.oauthDialog.IsOpen(), next.oauthSession)
	}
}

func TestOAuthAuthorizationFailureReplacesWaitingModal(t *testing.T) {
	m := viewTestModel(100, 30, panels.Projects)
	m, _ = m.updateOAuthAuthorizationEvent(core.OAuthAuthorizationEvent{
		FlowID: 9,
		AuthID: "default",
		URL:    "https://accounts.example.test/authorize",
	})
	m, _ = m.updateOAuthAuthorizationEvent(core.OAuthAuthorizationEvent{
		FlowID: 9,
		Done:   true,
		Err:    context.DeadlineExceeded,
	})

	if !m.oauthDialog.IsOpen() || m.oauthSession != nil {
		t.Fatalf("OAuth failure modal state = open:%v session:%#v", m.oauthDialog.IsOpen(), m.oauthSession)
	}
	view := ansi.Strip(m.oauthDialog.View())
	if !strings.Contains(view, "Authentication Failed") || !strings.Contains(view, context.DeadlineExceeded.Error()) {
		t.Fatalf("OAuth failure modal missing error:\n%s", view)
	}
}
