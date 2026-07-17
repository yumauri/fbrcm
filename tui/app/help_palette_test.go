package app

import (
	"slices"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/tui/components/minsize"
	tuiconfig "github.com/yumauri/fbrcm/tui/config"
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
	if !strings.Contains(view, "focus the Parameters panel first") {
		t.Fatalf("disabled palette action has no context explanation:\n%s", view)
	}
	if lipgloss.Width(m.helpPaletteView()) > m.width {
		t.Fatalf("palette width = %d, terminal width = %d", lipgloss.Width(m.helpPaletteView()), m.width)
	}
}

func TestHelpPaletteUsesSelectionWithoutMarkerOrBodyPadding(t *testing.T) {
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
	m.helpPalette, _ = m.helpPalette.Open()
	view := ansi.Strip(m.helpPaletteView())
	if strings.Contains(view, "Search: ") {
		t.Fatalf("palette search retains redundant prefix:\n%s", view)
	}
	if !strings.Contains(view, "│Unavailable: no project is selected") || strings.Contains(view, "│ Unavailable:") {
		t.Fatalf("palette status retains left padding:\n%s", view)
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
	if len(actions) != 1 || actions[0].enabled {
		t.Fatalf("filtered actions = %+v, want one disabled action", actions)
	}

	next, cmd, _ := m.runHelpPaletteAction(actions, helpPaletteListHeight(m.height))
	if cmd != nil || !next.helpPalette.IsOpen() {
		t.Fatalf("disabled action ran: cmd:%v open:%v", cmd != nil, next.helpPalette.IsOpen())
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
