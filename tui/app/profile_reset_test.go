package app

import (
	"testing"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/tui/panels"
)

func TestResetWorkspaceForProfileClearsProfileScopedState(t *testing.T) {
	m := viewTestModel(90, 24, panels.Parameters)
	m.detailsVisible = true
	m.parametersTab = panels.History
	m.projectsMode = projectsPanelModeCollapsed

	cmd := m.resetWorkspaceForProfile()

	if cmd == nil {
		t.Fatal("reset init command is nil")
	}
	if m.active != panels.Projects || m.parametersTab != panels.Parameters {
		t.Fatalf("active=%v parametersTab=%v, want Projects/Parameters", m.active, m.parametersTab)
	}
	if m.detailsVisible || m.projectsMode != projectsPanelModeExpanded {
		t.Fatalf("detailsVisible=%v projectsMode=%v, want reset", m.detailsVisible, m.projectsMode)
	}
	if m.width != 90 || m.height != 24 {
		t.Fatalf("size=%dx%d, want preserved 90x24", m.width, m.height)
	}
}

func TestResetWorkspaceForProfilePreservesOAuthEventWaiter(t *testing.T) {
	m := viewTestModel(90, 24, panels.Projects)
	m.oauthEvents = make(chan core.OAuthAuthorizationEvent, 1)
	waiting := m.waitOAuthAuthorizationEventCmd()

	m.resetWorkspaceForProfile()
	m.oauthEvents <- core.OAuthAuthorizationEvent{
		FlowID: 11,
		AuthID: "work",
		URL:    "https://accounts.example.test/authorize",
	}

	updated, _ := m.Update(waiting())
	next := updated.(Model)
	if !next.oauthDialog.IsOpen() || next.oauthSession == nil {
		t.Fatalf("OAuth modal after profile reset = open:%v session:%#v", next.oauthDialog.IsOpen(), next.oauthSession)
	}
	if next.oauthSession.url != "https://accounts.example.test/authorize" {
		t.Fatalf("OAuth modal URL = %q", next.oauthSession.url)
	}
}
