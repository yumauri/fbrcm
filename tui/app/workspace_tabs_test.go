package app

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/tui/components/workspaceheader"
	"github.com/yumauri/fbrcm/tui/panels"
)

func TestWorkspaceTitleClickActivatesSamePanelAsFocusKey(t *testing.T) {
	tests := []struct {
		name  string
		panel panels.ID
		key   string
	}{
		{name: "parameters", panel: panels.Parameters, key: "2"},
		{name: "conditions", panel: panels.Conditions, key: "3"},
		{name: "history", panel: panels.History, key: "4"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			m := viewTestModel(120, 30, panels.Projects)
			m.applyLayout()
			layout := newPanelLayout(m.width, m.height, m.projects.PreferredWidth(), m.logsHeight, m.projectsMode)
			localX := workspaceTabColumn(t, layout.rightWidth, workspaceTabIndex(m.selectedParametersTab()), workspaceTabIndex(test.panel))

			clicked, clickCmd, handled := m.updatePanelMouseMessage(tea.MouseClickMsg{
				X: layout.leftWidth + localX, Y: 0, Button: tea.MouseLeft,
			})
			keyed, keyCmd, keyHandled := m.updateGlobalFocusKey(test.key)

			if !handled || !keyHandled {
				t.Fatalf("handled click=%v key=%v; want both true", handled, keyHandled)
			}
			if clicked.active != test.panel || clicked.parametersTab != test.panel {
				t.Fatalf("clicked state active=%v tab=%v, want %v", clicked.active, clicked.parametersTab, test.panel)
			}
			if clicked.active != keyed.active || clicked.parametersTab != keyed.parametersTab {
				t.Fatalf("click state active=%v tab=%v differs from key state active=%v tab=%v", clicked.active, clicked.parametersTab, keyed.active, keyed.parametersTab)
			}
			if (clickCmd == nil) != (keyCmd == nil) {
				t.Fatalf("click command nil=%v differs from key command nil=%v", clickCmd == nil, keyCmd == nil)
			}
		})
	}
}

func workspaceTabColumn(t *testing.T, width, selected, target int) int {
	t.Helper()
	for x := range width {
		if index, ok := workspaceheader.TabAt(width, selected, x); ok && index == target {
			return x
		}
	}
	t.Fatalf("workspace tab %d has no hitbox at width %d", target, width)
	return 0
}
