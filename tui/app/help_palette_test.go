package app

import (
	"slices"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/tui/components/minsize"
	"github.com/yumauri/fbrcm/tui/components/setup"
	tuiconfig "github.com/yumauri/fbrcm/tui/config"
	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/panels"
	"github.com/yumauri/fbrcm/tui/styles"
)

func TestHelpPaletteCatalogCoversEveryConfiguredAction(t *testing.T) {
	type actionRef struct {
		block  tuiconfig.Block
		action tuiconfig.Action
	}
	seen := make(map[actionRef]int)
	for _, item := range helpPaletteCatalog() {
		seen[actionRef{block: item.block, action: item.action}]++
	}
	for block, actions := range tuiconfig.DefaultKeyMap() {
		for action := range actions {
			ref := actionRef{block: block, action: action}
			if seen[ref] != 1 {
				t.Errorf("catalog count for %s.%s = %d, want 1", block, action, seen[ref])
			}
		}
	}
}

func TestHelpPaletteCatalogDistinguishesPromotionToggleAndSubmit(t *testing.T) {
	titles := make(map[tuiconfig.Action]string)
	for _, item := range helpPaletteCatalog() {
		if item.block == tuiconfig.BlockPromote {
			titles[item.action] = item.title
		}
	}
	if titles[tuiconfig.ActionToggle] != "Select or clear change" ||
		titles[tuiconfig.ActionSubmit] != "Open selected change diff" {
		t.Fatalf("promotion action titles = %#v", titles)
	}
}

func TestHelpPaletteUsesPromotionWording(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{name: "workspace", got: helpPaletteBlockTitle(tuiconfig.BlockPromote), want: "Promote workspace"},
		{name: "project action", got: helpPaletteActionTitle(tuiconfig.BlockProjects, tuiconfig.ActionPromote), want: "Promote to another project"},
		{name: "focus action", got: helpPaletteActionTitle(tuiconfig.BlockGlobal, tuiconfig.ActionFocusPromote), want: "Focus promote"},
		{name: "close action", got: helpPaletteActionTitle(tuiconfig.BlockPromote, tuiconfig.ActionClose), want: "Close promotion"},
		{
			name: "focus description",
			got: helpPaletteActionDescription(
				tuiconfig.BlockGlobal,
				tuiconfig.ActionFocusPromote,
				helpPaletteActionTitle(tuiconfig.BlockGlobal, tuiconfig.ActionFocusPromote),
			),
			want: "Move keyboard focus to the promote panel.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Fatalf("wording = %q, want %q", tt.got, tt.want)
			}
		})
	}
}

func TestHelpKeyOpensAndClosesPalette(t *testing.T) {
	m := viewTestModel(90, 24, panels.Projects)
	next, _ := m.Update(keyPress('?'))
	m = next.(Model)
	if !m.helpPalette.IsOpen() {
		t.Fatal("help palette did not open for ?")
	}
	if m.mouseMode() != tea.MouseModeNone {
		t.Fatalf("mouse mode = %v, want none while palette is open", m.mouseMode())
	}

	next, _ = m.Update(keyPress('?'))
	m = next.(Model)
	if m.helpPalette.IsOpen() {
		t.Fatal("help palette did not close for second ?")
	}
}

func TestHelpPaletteKeepsLogSubscriptionAlive(t *testing.T) {
	m := viewTestModel(90, 24, panels.Projects)
	m.helpPalette, _ = m.helpPalette.Open()

	next, cmd := m.Update(messages.LogLineMsg{Line: "background log while actions are open"})
	m = next.(Model)
	if !m.helpPalette.IsOpen() {
		t.Fatal("background log closed the Actions palette")
	}
	if cmd == nil {
		t.Fatal("background log did not schedule the next subscription read")
	}
	if view := ansi.Strip(m.logs.View(false)); !strings.Contains(view, "background log while actions are open") {
		t.Fatalf("Logs panel did not receive background line:\n%s", view)
	}
}

func TestHelpKeyRemainsTextInsideEditor(t *testing.T) {
	m := viewTestModel(90, 24, panels.Parameters)
	m.renameInput, _ = m.renameInput.Open(0, 0, 4, 20, "name")

	next, _ := m.Update(keyPress('?'))
	m = next.(Model)
	if m.helpPalette.IsOpen() {
		t.Fatal("help palette captured ? while a text editor was open")
	}
	if got := m.renameInput.Value(); got != "name?" {
		t.Fatalf("rename value = %q, want name?", got)
	}

	m.helpPalette, _ = m.helpPalette.Open()
	groups := helpPaletteGroups(m.helpPaletteActions())
	if len(groups) < 2 || groups[0] != "Rename editor" || groups[1] != "Global" {
		t.Fatalf("editor palette groups = %v, want Rename editor then Global", groups)
	}
}

func TestHelpPaletteOpensOverAccountsAndProfilesWithActiveActions(t *testing.T) {
	svc := newRenameTestService(t)
	if _, err := svc.AddGCloudAuth("main", "main"); err != nil {
		t.Fatalf("AddGCloudAuth = %v", err)
	}
	m := viewTestModel(90, 24, panels.Projects)
	m.svc = svc
	m.setup = setup.New(svc)
	var cmd tea.Cmd
	m.setup, cmd = m.setup.OpenAccounts()
	m = finishSetupInspection(t, m, cmd)

	next, _ := m.Update(keyPress('?'))
	m = next.(Model)
	if !m.helpPalette.IsOpen() || !m.setup.IsOpen() {
		t.Fatalf("Actions over Accounts = actions:%v setup:%v", m.helpPalette.IsOpen(), m.setup.IsOpen())
	}
	groups := helpPaletteGroups(m.helpPaletteActions())
	if len(groups) < 2 || groups[0] != "Accounts panel" || groups[1] != "Global" {
		t.Fatalf("Accounts action groups = %v", groups)
	}
	if enabled, reason := m.globalHelpActionAvailability(tuiconfig.ActionQuit); !enabled || reason != "" {
		t.Fatalf("Accounts quit availability = %v, %q", enabled, reason)
	}
	if enabled, reason := m.globalHelpActionAvailability(tuiconfig.ActionProfiles); !enabled || reason != "" {
		t.Fatalf("Accounts to Profiles availability = %v, %q", enabled, reason)
	}
	if enabled, reason := m.globalHelpActionAvailability(tuiconfig.ActionAccounts); enabled || !strings.Contains(reason, "already active") {
		t.Fatalf("active Accounts availability = %v, %q", enabled, reason)
	}
	view := ansi.Strip(m.View().Content)
	for _, want := range []string{"ˀActions", "Accounts panel", "Delete authentication"} {
		if !strings.Contains(view, want) {
			t.Fatalf("Actions over Accounts missing %q:\n%s", want, view)
		}
	}

	next, _ = m.Update(keyPress('?'))
	m = next.(Model)
	next, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: 'p', Mod: tea.ModCtrl}))
	m = next.(Model)
	next, _ = m.Update(keyPress('?'))
	m = next.(Model)
	groups = helpPaletteGroups(m.helpPaletteActions())
	if len(groups) < 2 || groups[0] != "Profiles panel" || groups[1] != "Global" {
		t.Fatalf("Profiles action groups = %v", groups)
	}
	if enabled, reason := m.globalHelpActionAvailability(tuiconfig.ActionQuit); !enabled || reason != "" {
		t.Fatalf("Profiles quit availability = %v, %q", enabled, reason)
	}
	if enabled, reason := m.globalHelpActionAvailability(tuiconfig.ActionAccounts); !enabled || reason != "" {
		t.Fatalf("Profiles to Accounts availability = %v, %q", enabled, reason)
	}
	if enabled, reason := m.globalHelpActionAvailability(tuiconfig.ActionProfiles); enabled || !strings.Contains(reason, "already active") {
		t.Fatalf("active Profiles availability = %v, %q", enabled, reason)
	}
	if view = ansi.Strip(m.helpPaletteView()); !strings.Contains(view, "Rename profile") || !strings.Contains(view, "Delete profile") {
		t.Fatalf("Profiles actions missing:\n%s", view)
	}
}

func TestActionsTitleUsesConfiguredSuperscriptHintAndSelectedBackground(t *testing.T) {
	m := viewTestModel(90, 24, panels.Projects)
	m.helpPalette, _ = m.helpPalette.Open()
	view := m.helpPaletteView()
	selection := styles.TitleStyle(true)
	key := styles.FilterText.Background(selection.GetBackground()).Render("ˀ")
	if !strings.Contains(view, key) {
		t.Fatalf("Actions title does not render selected configured hint: %q", view)
	}
}

func finishSetupInspection(t *testing.T, m Model, cmd tea.Cmd) Model {
	t.Helper()
	if cmd == nil {
		t.Fatal("setup inspection command is nil")
	}
	msg := cmd()
	batch, ok := msg.(tea.BatchMsg)
	if !ok {
		t.Fatalf("setup command = %T, want tea.BatchMsg", msg)
	}
	for _, item := range batch {
		if item == nil {
			continue
		}
		next, _ := m.Update(item())
		m = next.(Model)
	}
	return m
}

func TestHelpPaletteOrdersActivePanelThenGlobalThenAlphabetically(t *testing.T) {
	m := viewTestModel(90, 24, panels.Parameters)
	groups := helpPaletteGroups(m.helpPaletteActions())
	if len(groups) < 3 || groups[0] != "Parameters panel" || groups[1] != "Global" {
		t.Fatalf("palette groups = %v, want Parameters panel then Global", groups)
	}
	if !slices.IsSorted(groups[2:]) {
		t.Fatalf("remaining palette groups are not alphabetical: %v", groups[2:])
	}
}

func helpPaletteGroups(actions []helpPaletteAction) []string {
	var groups []string
	for _, item := range actions {
		if len(groups) == 0 || groups[len(groups)-1] != item.group {
			groups = append(groups, item.group)
		}
	}
	return groups
}

func TestHelpPaletteSearchShowsDisabledContextReason(t *testing.T) {
	m := viewTestModel(90, 24, panels.Projects)
	m.helpPalette, _ = m.helpPalette.Open()
	m.helpPalette.input.SetValue("collapse all")

	view := ansi.Strip(m.helpPaletteView())
	if !strings.Contains(view, "Parameters panel") || !strings.Contains(view, "Collapse all") {
		t.Fatalf("filtered palette does not show matching grouped action:\n%s", view)
	}
	if !strings.Contains(view, "Unavailable: focus the Parameters panel") {
		t.Fatalf("disabled palette action has no context explanation:\n%s", view)
	}
	if lipgloss.Width(m.helpPaletteView()) > m.width {
		t.Fatalf("palette width = %d, terminal width = %d", lipgloss.Width(m.helpPaletteView()), m.width)
	}
}

func TestHelpPaletteUsesSelectionWithoutMarkerAndStandardPopupPadding(t *testing.T) {
	item := helpPaletteAction{
		group:   "Projects panel",
		title:   "Select",
		keys:    []string{"enter"},
		enabled: true,
	}
	row := ansi.Strip(renderHelpPaletteAction(item, item.group, true, 80))
	if strings.HasPrefix(row, ">") || strings.HasPrefix(row, " ") {
		t.Fatalf("selected row retains marker or left padding: %q", row)
	}

	m := viewTestModel(90, 24, panels.Projects)
	m.authCount = 2
	m.helpPalette, _ = m.helpPalette.Open()
	view := ansi.Strip(m.helpPaletteView())
	if strings.Contains(view, "Search: ") {
		t.Fatalf("palette search retains redundant prefix:\n%s", view)
	}
	if !strings.Contains(view, "│  Bind authentication in the projects panel. Unavailable:") {
		t.Fatalf("palette status does not use the standard popup inset:\n%s", view)
	}
}

func TestHelpPaletteUsesUnderstandableSearchableActionMetadata(t *testing.T) {
	catalog := helpPaletteCatalog()
	for _, item := range catalog {
		if strings.TrimSpace(item.description) == "" {
			t.Errorf("%s.%s has no description", item.block, item.action)
		}
		if slices.Contains([]string{"First", "Last", "Home", "End", "Up", "Down"}, item.title) {
			t.Errorf("%s.%s has ambiguous title %q", item.block, item.action, item.title)
		}
	}

	search := newHelpPaletteModel()
	search.input.SetValue("update")
	updateTitles := make([]string, 0)
	for _, item := range search.filtered(catalog) {
		updateTitles = append(updateTitles, item.title)
	}
	for _, want := range []string{"Update projects", "Update current project", "Update all projects"} {
		if !slices.Contains(updateTitles, want) {
			t.Errorf("update search titles = %v, missing %q", updateTitles, want)
		}
	}

	search.input.SetValue("reload")
	reloadTitles := make([]string, 0)
	for _, item := range search.filtered(catalog) {
		reloadTitles = append(reloadTitles, item.title)
	}
	for _, want := range []string{"Update current project", "Update all projects"} {
		if !slices.Contains(reloadTitles, want) {
			t.Errorf("reload alias search titles = %v, missing %q", reloadTitles, want)
		}
	}
}

func TestSharedFooterActionsAppearInTheActivePaletteGroup(t *testing.T) {
	tests := []struct {
		panel   panels.ID
		query   string
		actions []tuiconfig.Action
		group   string
	}{
		{panel: panels.Conditions, query: "update", actions: []tuiconfig.Action{tuiconfig.ActionReload, tuiconfig.ActionReloadAll}, group: "Conditions panel"},
		{panel: panels.History, query: "maximize", actions: []tuiconfig.Action{tuiconfig.ActionToggleMaximize}, group: "History panel"},
	}
	for _, tt := range tests {
		m := viewTestModel(90, 24, tt.panel)
		m.helpPalette.input.SetValue(tt.query)
		byAction := make(map[tuiconfig.Action]helpPaletteAction)
		for _, item := range m.helpPalette.filtered(m.helpPaletteActions()) {
			byAction[item.action] = item
		}
		for _, action := range tt.actions {
			item, ok := byAction[action]
			if !ok {
				t.Errorf("panel %v search %q has no %s action", tt.panel, tt.query, action)
				continue
			}
			if item.group != tt.group || strings.Contains(item.reason, "focus the Parameters panel") {
				t.Errorf("%s action = group:%q reason:%q, want active group %q", action, item.group, item.reason, tt.group)
			}
		}
	}
}

func TestShortHelpTermsFindTheirPaletteActions(t *testing.T) {
	catalog := helpPaletteCatalog()
	for i := range catalog {
		catalog[i].keys = tuiconfig.Keys(catalog[i].block, catalog[i].action)
	}
	contexts := []helpKeyMap{
		{keyboardCapture: true},
		{active: panels.Projects, canBindAuth: true},
		{active: panels.Projects, projectsMode: projectsPanelModeCollapsed, canBindAuth: true},
		{active: panels.Parameters},
		{active: panels.Conditions},
		{active: panels.History},
		{active: panels.Logs},
		{active: panels.Logs, logsMode: logsPanelModeCollapsed},
		{active: panels.Details},
		{active: panels.Details, conditionDetail: true},
		{active: panels.Details, groupDetail: true},
		{active: panels.Conditions, conditionMove: true},
	}
	for _, context := range contexts {
		for _, binding := range context.ShortHelp() {
			query := binding.Help().Desc
			search := newHelpPaletteModel()
			search.input.SetValue(query)
			found := false
			for _, item := range search.filtered(catalog) {
				if slices.ContainsFunc(binding.Keys(), func(key string) bool { return slices.Contains(item.keys, key) }) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("footer term %q with keys %v finds no matching Actions entry", query, binding.Keys())
			}
		}
	}
}

func TestHelpPaletteAndPanelsRespondToWindowResize(t *testing.T) {
	m := viewTestModel(90, 24, panels.Projects)
	m.helpPalette, _ = m.helpPalette.Open()
	oldPaletteWidth := lipgloss.Width(m.helpPaletteView())
	actions := m.helpPalette.filtered(m.helpPaletteActions())
	m.helpPalette.goTo(30, len(actions), helpPaletteListHeight(m.height))

	next, _ := m.Update(tea.WindowSizeMsg{Width: 110, Height: 30})
	m = next.(Model)
	if m.width != 110 || m.height != 30 {
		t.Fatalf("size = %dx%d, want 110x30", m.width, m.height)
	}
	if !m.helpPalette.IsOpen() {
		t.Fatal("palette closed on a supported resize")
	}
	if got := lipgloss.Width(m.helpPaletteView()); got <= oldPaletteWidth || got > m.width {
		t.Fatalf("resized palette width = %d, previous = %d, terminal = %d", got, oldPaletteWidth, m.width)
	}
	listHeight := helpPaletteListHeight(m.height)
	if m.helpPalette.cursor < m.helpPalette.scroll || m.helpPalette.cursor >= m.helpPalette.scroll+listHeight {
		t.Fatalf("selection %d is outside resized viewport [%d,%d)", m.helpPalette.cursor, m.helpPalette.scroll, m.helpPalette.scroll+listHeight)
	}

	next, _ = m.Update(tea.WindowSizeMsg{Width: minsize.MinWidth, Height: minsize.MinHeight})
	m = next.(Model)
	if !m.helpPalette.IsOpen() {
		t.Fatal("palette closed at the supported minimum terminal size")
	}
	if got := lipgloss.Width(m.helpPaletteView()); got != minsize.MinWidth-8 {
		t.Fatalf("minimum-size palette width = %d, want %d", got, minsize.MinWidth-8)
	}
	if got := lipgloss.Height(m.helpPaletteView()); got > minsize.MinHeight {
		t.Fatalf("minimum-size palette height = %d, terminal height = %d", got, minsize.MinHeight)
	}

	m.helpPalette = m.helpPalette.Close()
	if got := lipgloss.Width(m.baseView()); got != minsize.MinWidth {
		t.Fatalf("base panel width after palette resize = %d, want %d", got, minsize.MinWidth)
	}
}

func TestMinimumSizeViewCoversAndRestoresHelpPalette(t *testing.T) {
	m := viewTestModel(90, 24, panels.Projects)
	m.helpPalette, _ = m.helpPalette.Open()
	m.helpPalette.input.SetValue("focus logs")

	next, _ := m.Update(tea.WindowSizeMsg{Width: minsize.MinWidth - 1, Height: minsize.MinHeight - 1})
	m = next.(Model)
	if !m.helpPalette.IsOpen() {
		t.Fatal("minimum-size view closed the covered palette")
	}
	view := ansi.Strip(m.View().Content)
	if !strings.Contains(view, "Terminal too small") || !strings.Contains(view, "Minimum size 80x20") {
		t.Fatalf("minimum-size message is not visible:\n%s", view)
	}
	if strings.Contains(view, "Actions") || strings.Contains(view, "focus logs") {
		t.Fatalf("covered palette leaked through minimum-size view:\n%s", view)
	}

	next, _ = m.Update(tea.WindowSizeMsg{Width: 90, Height: 24})
	m = next.(Model)
	if !m.helpPalette.IsOpen() || m.helpPalette.input.Value() != "focus logs" {
		t.Fatalf("restored palette lost state: open:%v query:%q", m.helpPalette.IsOpen(), m.helpPalette.input.Value())
	}
	view = ansi.Strip(m.View().Content)
	if !strings.Contains(view, "Actions") || !strings.Contains(view, "Focus logs") {
		t.Fatalf("palette was not restored after widening terminal:\n%s", view)
	}
}

func TestHelpPaletteRunsSelectedAvailableAction(t *testing.T) {
	m := viewTestModel(90, 24, panels.Projects)
	m.helpPalette, _ = m.helpPalette.Open()
	m.helpPalette.input.SetValue("focus logs")
	actions := m.helpPalette.filtered(m.helpPaletteActions())
	if len(actions) != 1 || actions[0].action != tuiconfig.ActionFocusLogs {
		t.Fatalf("filtered actions = %+v, want focus logs", actions)
	}

	next, cmd, handled := m.runHelpPaletteAction(actions, helpPaletteListHeight(m.height))
	if !handled || cmd == nil || next.helpPalette.IsOpen() {
		t.Fatalf("run action = handled:%v cmd:%v open:%v", handled, cmd != nil, next.helpPalette.IsOpen())
	}
	updated, _ := next.Update(cmd())
	if got := updated.(Model).active; got != panels.Logs {
		t.Fatalf("active panel = %v, want logs", got)
	}
}

func TestHelpPaletteDoesNotRunDisabledAction(t *testing.T) {
	m := viewTestModel(90, 24, panels.Projects)
	m.helpPalette, _ = m.helpPalette.Open()
	m.helpPalette.input.SetValue("collapse all")
	actions := m.helpPalette.filtered(m.helpPaletteActions())
	if len(actions) == 0 || actions[0].action != tuiconfig.ActionCollapseAll || actions[0].enabled {
		t.Fatalf("first filtered action = %+v, want disabled Collapse all", actions)
	}

	next, cmd, _ := m.runHelpPaletteAction(actions, helpPaletteListHeight(m.height))
	if cmd != nil || !next.helpPalette.IsOpen() {
		t.Fatalf("disabled action ran: cmd:%v open:%v", cmd != nil, next.helpPalette.IsOpen())
	}
}

func TestHelpPaletteRecognizesGroupDetailsActions(t *testing.T) {
	m := viewTestModel(90, 24, panels.Details)
	m.details = m.details.SetGroupData(&messages.GroupViewData{
		Project: core.Project{Name: "Demo", ProjectID: "demo"},
		Group:   core.ParametersGroup{Key: "checkout", Description: "Checkout flags"},
	})
	m.detailsVisible = true
	m.setActive(panels.Details)

	actions := m.helpPaletteActions()
	availability := make(map[tuiconfig.Action]helpPaletteAction)
	for _, item := range actions {
		if item.block == tuiconfig.BlockDetails {
			availability[item.action] = item
		}
	}
	for _, action := range []tuiconfig.Action{tuiconfig.ActionClose, tuiconfig.ActionSubmit, tuiconfig.ActionRename, tuiconfig.ActionDelete, tuiconfig.ActionCopyName, tuiconfig.ActionCopyPath} {
		if item := availability[action]; !item.enabled {
			t.Errorf("group Details action %s disabled: %s", action, item.reason)
		}
	}
	for _, action := range []tuiconfig.Action{tuiconfig.ActionNew, tuiconfig.ActionEditValue, tuiconfig.ActionMove, tuiconfig.ActionColor, tuiconfig.ActionCopyValue} {
		if item := availability[action]; item.enabled || !strings.Contains(item.reason, "parameter groups") {
			t.Errorf("group-only invalid action %s = enabled:%v reason:%q", action, item.enabled, item.reason)
		}
	}
}

func TestHelpPaletteStylesShortcutColumnLikeApplicationHelp(t *testing.T) {
	item := helpPaletteAction{
		group:   "Projects panel",
		title:   "Select",
		keys:    []string{"enter"},
		enabled: true,
	}
	styledKeyColumn := styles.FilterText.Render(strings.Repeat(" ", 11) + "enter")
	if got := renderHelpPaletteAction(item, item.group, false, 80); !strings.Contains(got, styledKeyColumn) {
		t.Fatalf("action shortcut does not use shared help key style: %q", got)
	}

	item.enabled = false
	item.reason = "no project is selected"
	if got := renderHelpPaletteAction(item, item.group, false, 80); !strings.Contains(got, styledKeyColumn) {
		t.Fatalf("disabled action shortcut does not use shared help key style: %q", got)
	}
}

func TestHelpPaletteFooterUsesApplicationHelpStyles(t *testing.T) {
	m := viewTestModel(90, 24, panels.Projects)
	footer := m.helpPaletteFooter(80)
	for _, want := range []string{
		styles.FilterText.Render("enter"),
		styles.PanelMuted.Render("run"),
		styles.PanelMuted.Render(" • "),
	} {
		if !strings.Contains(footer, want) {
			t.Fatalf("palette footer does not contain shared help style %q: %q", want, footer)
		}
	}
}

func keyPress(r rune) tea.KeyPressMsg {
	return tea.KeyPressMsg(tea.Key{Code: r, Text: string(r)})
}
