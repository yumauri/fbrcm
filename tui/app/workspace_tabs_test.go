package app

import (
	"slices"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/yumauri/fbrcm/tui/components/workspaceheader"
	"github.com/yumauri/fbrcm/tui/panels"
	"github.com/yumauri/fbrcm/tui/styles"
	"github.com/yumauri/fbrcm/tui/testutil"
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
		{name: "A/B tests", panel: panels.ABTests, key: "6"},
		{name: "personalizations", panel: panels.Personalizations, key: "7"},
		{name: "rollouts", panel: panels.Rollouts, key: "8"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			m := viewTestModel(200, 30, panels.Projects)
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

func TestWorkspaceOverflowKeyOpensOnlyHiddenTitles(t *testing.T) {
	m := viewTestModel(90, 24, panels.Projects)
	m.applyLayout()

	next, cmd, handled := m.updateGlobalKeyMessage("\\")
	if !handled || cmd != nil || !next.workspaceMenu {
		t.Fatalf("workspace menu open = handled:%v cmd:%v open:%v, want true nil true", handled, cmd, next.workspaceMenu)
	}

	header, _, ok := next.workspaceHeaderLayout()
	if !ok {
		t.Fatal("workspace header layout is unavailable")
	}
	popup := header.PopupView(next.workspaceCursor, styles.BorderStyle(true))
	menuTabs := header.MenuTabs()
	for _, tab := range menuTabs[next.workspaceCursor:] {
		if !strings.Contains(popup, workspaceTabTitle(tab)) {
			t.Errorf("popup does not contain visible title %q: %q", workspaceTabTitle(tab), popup)
		}
	}
	for _, tab := range menuTabs[:next.workspaceCursor] {
		if strings.Contains(popup, workspaceTabTitle(tab)) {
			t.Errorf("popup contains title %q above the visible stack: %q", workspaceTabTitle(tab), popup)
		}
	}

	view := testutil.NormalizeViewSnapshot(next.View().Content)
	if !strings.Contains(view, "│  ▸    ³Conditions") || !strings.Contains(view, "╰─") || !strings.Contains(view, "Personalizations") {
		t.Fatalf("workspace popup does not show its shifted side and bottom borders:\n%s", view)
	}
}

func TestWorkspaceOverflowArrowsShiftOrderedStackThroughBorderRow(t *testing.T) {
	m := viewTestModel(90, 24, panels.Projects).openWorkspaceMenu()
	originalActive := m.active
	originalTab := m.parametersTab
	header, _, _ := m.workspaceHeaderLayout()
	menuTabs := header.MenuTabs()
	for m.workspaceCursor < len(menuTabs)-1 {
		wantCursor := m.workspaceCursor + 1
		next, _, handled := m.updateWorkspaceMenu(tea.KeyPressMsg(tea.Key{Code: tea.KeyDown}))
		if !handled || !next.workspaceMenu || next.workspaceCursor != wantCursor {
			t.Fatalf("down cursor = handled:%v open:%v cursor:%d, want true true %d", handled, next.workspaceMenu, next.workspaceCursor, wantCursor)
		}
		if next.active != originalActive || next.parametersTab != originalTab {
			t.Fatalf("down activated panel: active=%v tab=%v, want unchanged %v/%v", next.active, next.parametersTab, originalActive, originalTab)
		}
		nextHeader, _, _ := next.workspaceHeaderLayout()
		if nextHeader.MenuCursor() != wantCursor {
			t.Fatalf("down border-row cursor = %d, want %d", nextHeader.MenuCursor(), wantCursor)
		}
		m = next
	}
	header, _, _ = m.workspaceHeaderLayout()
	if got := header.MenuTabs(); !slices.IsSorted(got) {
		t.Fatalf("menu tabs are not in workspace order: %v", got)
	}
	if popup := header.PopupView(m.workspaceCursor, styles.BorderStyle(true)); lipgloss.Height(popup) != 2 {
		t.Fatalf("Rollouts should shift the menu fully upward, got:\n%s", popup)
	}

	next, _, handled := m.updateWorkspaceMenu(tea.KeyPressMsg(tea.Key{Code: tea.KeyUp}))
	if !handled || !next.workspaceMenu || next.workspaceCursor != len(menuTabs)-2 {
		t.Fatalf("up = handled:%v open:%v cursor:%d, want true true %d", handled, next.workspaceMenu, next.workspaceCursor, len(menuTabs)-2)
	}
	if next.active != originalActive || next.parametersTab != originalTab {
		t.Fatalf("up activated panel: active=%v tab=%v, want unchanged %v/%v", next.active, next.parametersTab, originalActive, originalTab)
	}
	header, _, _ = next.workspaceHeaderLayout()
	if popup := header.PopupView(next.workspaceCursor, styles.BorderStyle(true)); !strings.Contains(popup, "Rollouts") {
		t.Fatalf("up did not shift Rollouts below the border row:\n%s", popup)
	}

	wantPanel, _ := workspacePanel(menuTabs[next.workspaceCursor])
	committed, _, handled := next.updateWorkspaceMenu(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	if !handled || committed.workspaceMenu || committed.active != wantPanel || committed.parametersTab != wantPanel {
		t.Fatalf("enter = handled:%v open:%v active:%v tab:%v, want true false %v/%v", handled, committed.workspaceMenu, committed.active, committed.parametersTab, wantPanel, wantPanel)
	}
}

func TestWorkspaceOverflowPreviewKeepsCollapseGeometryStable(t *testing.T) {
	m := viewTestModel(90, 24, panels.Projects).openWorkspaceMenu()
	header, _, ok := m.workspaceHeaderLayout()
	if !ok {
		t.Fatal("workspace header layout is unavailable")
	}
	initialTabs := header.MenuTabs()
	initialX, _, ok := header.OverflowButtonBounds()
	if !ok {
		t.Fatal("compact header has no overflow button")
	}

	for m.workspaceCursor < len(initialTabs)-1 && initialTabs[m.workspaceCursor] != 2 {
		m, _, _ = m.updateWorkspaceMenu(tea.KeyPressMsg(tea.Key{Code: tea.KeyDown}))
	}
	if initialTabs[m.workspaceCursor] != 2 {
		t.Fatalf("History is not available in menu tabs %v", initialTabs)
	}
	previewHeader, _, _ := m.workspaceHeaderLayout()
	previewX, _, _ := previewHeader.OverflowButtonBounds()
	if previewX != initialX {
		t.Fatalf("History preview moved overflow button from %d to %d", initialX, previewX)
	}
	if got := previewHeader.MenuTabs(); !slices.Equal(got, initialTabs) {
		t.Fatalf("History preview changed collapsed sequence from %v to %v", initialTabs, got)
	}
	if m.active != panels.Projects {
		t.Fatalf("History preview activated %v, want Projects", m.active)
	}
}

func TestWorkspaceOverflowActiveTitleStaysSelectedBelowCursor(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	m := viewTestModel(90, 24, panels.Personalizations).openWorkspaceMenu()
	if !m.workspaceMenu {
		t.Fatal("workspace menu did not open")
	}

	next, _, handled := m.updateWorkspaceMenu(tea.KeyPressMsg(tea.Key{Code: tea.KeyUp}))
	if !handled || next.active != panels.Personalizations || next.parametersTab != panels.Personalizations {
		t.Fatalf("up = handled:%v active:%v tab:%v, want Personalizations to remain active", handled, next.active, next.parametersTab)
	}
	header, _, _ := next.workspaceHeaderLayout()
	popup := header.PopupViewFocused(next.workspaceCursor, true, styles.BorderStyle(true))
	if activeLabel := styles.TitleStyle(true).Render("Personalizations "); !strings.Contains(popup, activeLabel) {
		t.Fatalf("active Personalizations title lost its selected background below the cursor: %q", popup)
	}
}

func TestWorkspaceOverflowMouseOpensAndSelectsImmediately(t *testing.T) {
	m := viewTestModel(90, 24, panels.Projects)
	m.applyLayout()
	header, panelX, ok := m.workspaceHeaderLayout()
	if !ok {
		t.Fatal("workspace header layout is unavailable")
	}
	buttonX, _, ok := header.OverflowButtonBounds()
	if !ok {
		t.Fatal("compact header has no overflow button")
	}

	opened, _, handled := m.updatePanelMouseMessage(tea.MouseClickMsg{
		X: panelX + buttonX, Y: 0, Button: tea.MouseLeft,
	})
	if !handled || !opened.workspaceMenu {
		t.Fatalf("button click = handled:%v open:%v, want true true", handled, opened.workspaceMenu)
	}
	header, panelX, _ = opened.workspaceHeaderLayout()
	geometry, ok := header.PopupGeometry()
	if !ok || len(geometry.Tabs)-opened.workspaceCursor-1 < 2 {
		t.Fatalf("popup geometry = %+v, %v", geometry, ok)
	}
	selectedTab := geometry.Tabs[opened.workspaceCursor+2]
	selected, _, handled := opened.updateWorkspaceMenu(tea.MouseClickMsg{
		X: panelX + geometry.X + geometry.LeftPadding + 1, Y: 2, Button: tea.MouseLeft,
	})
	wantPanel, _ := workspacePanel(selectedTab)
	if !handled || selected.workspaceMenu || selected.active != wantPanel {
		t.Fatalf("menu click = handled:%v open:%v active:%v, want true false %v", handled, selected.workspaceMenu, selected.active, wantPanel)
	}
}

func TestWorkspaceOverflowClosesWhenAllTitlesFitAfterResize(t *testing.T) {
	m := viewTestModel(90, 24, panels.Projects).openWorkspaceMenu()
	if !m.workspaceMenu {
		t.Fatal("workspace menu did not open at narrow width")
	}
	next, _, handled := m.updateWorkspaceMenu(tea.WindowSizeMsg{Width: 220, Height: 24})
	if !handled || next.workspaceMenu {
		t.Fatalf("wide resize = handled:%v open:%v, want true false", handled, next.workspaceMenu)
	}
}

func TestWorkspaceOverflowKeepsActivePanelTitleAndBorder(t *testing.T) {
	m := viewTestModel(90, 24, panels.Parameters).openWorkspaceMenu()
	if !m.workspaceMenu {
		t.Fatal("workspace menu did not open")
	}
	if m.popupWindowOpen() {
		t.Fatal("workspace menu dims the active panel")
	}
	if !m.workspaceMenuBorderActive() {
		t.Fatal("workspace menu border is inactive while the workspace panel is active")
	}

	wantTop := lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.projects.ViewWithBorder(false, false),
		m.parameters.ViewWithWorkspacePreview(true, true, headerMenuTab(t, m)),
	)
	want := lipgloss.JoinVertical(
		lipgloss.Left,
		wantTop,
		m.logs.ViewWithBorder(false, false),
		m.helpView(),
	)
	if got := m.baseView(); got != want {
		t.Fatal("workspace menu changed the active title or panel border styling")
	}

	inactive := viewTestModel(90, 24, panels.Projects).openWorkspaceMenu()
	if inactive.workspaceMenuBorderActive() {
		t.Fatal("workspace menu border is active while the Projects panel is active")
	}
}

func TestWorkspaceOverflowQuitUsesNormalQuitFlow(t *testing.T) {
	m := viewTestModel(90, 24, panels.Projects).openWorkspaceMenu()
	next, cmd, handled := m.updateWorkspaceMenu(tea.KeyPressMsg(tea.Key{Code: 'q', Text: "q"}))
	if !handled || cmd == nil || next.workspaceMenu {
		t.Fatalf("q = handled:%v cmd:%v open:%v, want true non-nil false", handled, cmd != nil, next.workspaceMenu)
	}
}

func TestWorkspaceHeaderReserveTracksProfileBadgeWidth(t *testing.T) {
	m := viewTestModel(100, 24, panels.Parameters)
	m.profileName = "x"
	m.applyLayout()
	shortLayout, _, _ := m.workspaceHeaderLayout()

	m.profileName = strings.Repeat("profile", 4)
	m.applyLayout()
	longLayout, _, _ := m.workspaceHeaderLayout()

	if len(longLayout.HiddenTabs()) <= len(shortLayout.HiddenTabs()) {
		t.Fatalf("hidden tabs short=%v long=%v, want longer profile to reserve more header width", shortLayout.HiddenTabs(), longLayout.HiddenTabs())
	}
}

func TestWorkspaceTitleHitboxesUseProfileReservedLayout(t *testing.T) {
	m := viewTestModel(100, 24, panels.Parameters)
	m.profileName = strings.Repeat("profile", 4)
	m.applyLayout()
	header, panelX, ok := m.workspaceHeaderLayout()
	if !ok {
		t.Fatal("workspace header layout is unavailable")
	}
	panelLayout := newPanelLayout(m.width, m.height, m.projects.PreferredWidth(), m.logsHeight, m.projectsMode)
	selected := workspaceTabIndex(m.selectedParametersTab())

	for localX := range panelLayout.rightWidth {
		wantTab, reservedHit := header.TabAt(localX)
		unreservedTab, unreservedHit := workspaceheader.TabAt(panelLayout.rightWidth, selected, localX)
		if !reservedHit || (unreservedHit && unreservedTab == wantTab) {
			continue
		}
		wantPanel, _ := workspacePanel(wantTab)
		gotPanel, hit := m.workspaceTabAt(panelX+localX, 0)
		if !hit || gotPanel != wantPanel {
			t.Fatalf("profile-reserved title at x=%d hit (%v, %v), want (%v, true)", localX, gotPanel, hit, wantPanel)
		}
		return
	}
	t.Fatal("test width did not produce a title hitbox difference between reserved and unreserved layouts")
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

func headerMenuTab(t *testing.T, m Model) int {
	t.Helper()
	header, _, ok := m.workspaceHeaderLayout()
	if !ok {
		t.Fatal("workspace header layout is unavailable")
	}
	menuTabs := header.MenuTabs()
	if len(menuTabs) == 0 {
		t.Fatal("workspace header has no menu tabs")
	}
	return menuTabs[m.workspaceCursor]
}

func workspaceTabTitle(tab int) string {
	switch tab {
	case 0:
		return "Parameters"
	case 1:
		return "Conditions"
	case 2:
		return "History"
	case 3:
		return "A/B Tests"
	case 4:
		return "Personalizations"
	case 5:
		return "Rollouts"
	default:
		return ""
	}
}
