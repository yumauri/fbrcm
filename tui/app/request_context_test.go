package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/conditions"
	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/core/parameters"
	"github.com/yumauri/fbrcm/core/rc/evaluate"
	"github.com/yumauri/fbrcm/core/rootgroup"
	"github.com/yumauri/fbrcm/tui/components/projectio"
	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/panels"
	"github.com/yumauri/fbrcm/tui/styles"
)

func TestContextAddsAndEditsFactUsingBothLines(t *testing.T) {
	m := viewTestModel(80, 24, panels.Projects)
	click := func(x, y int) tea.MouseClickMsg {
		return tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft}
	}
	next, _ := m.Update(keyPress('9'))
	m = next.(Model)
	if m.active != panels.Context || m.requestContext.overlay != contextInline {
		t.Fatal("9 did not focus the inline Context panel")
	}
	next, _ = m.Update(keyPress('a'))
	m = next.(Model)
	if m.requestContext.overlay != contextFields {
		t.Fatal("a did not open the Context field picker")
	}
	next, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	m = next.(Model)
	next, _ = m.Update(tea.PasteMsg{Content: "web"})
	m = next.(Model)
	next, apply := m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	if apply == nil {
		t.Fatal("Enter in the field editor did not submit")
	}
	next, _ = next.Update(apply())
	m = next.(Model)
	if got := m.requestContext.facts["device.os"].Value; got != "web" {
		t.Fatalf("context platform = %q, want web", got)
	}
	m.applyContextFact(contextFactMsg{assignment: "app.customSignal['environment']=production"})
	layout := m.contextLayout(false)
	view := ansi.Strip(m.sidebarView(false, true))
	if !strings.Contains(view, "Signal · environment") || !strings.Contains(view, " ╰ production") {
		t.Fatalf("two-line input missing from sidebar:\n%s", view)
	}
	lines := strings.Split(view, "\n")
	header := lines[layout.Y]
	if !strings.HasPrefix(header, "── ⁹Context ") || !strings.HasSuffix(header, " 2 ─╮") {
		t.Fatalf("Context header does not use the sidebar title and count: %q", header)
	}
	if strings.Contains(view, "[+ Add]") || strings.Contains(view, "[Import…]") {
		t.Fatal("Context sidebar still has an action row")
	}
	for _, line := range lines[layout.Y+1:] {
		if strings.HasPrefix(line, "│") || strings.HasPrefix(line, "╰") {
			t.Fatalf("Context sidebar still has a left border: %q", line)
		}
	}
	for line := range strings.SplitSeq(view, "\n") {
		if lipgloss.Width(line) != layout.Width {
			t.Fatalf("sidebar line has width %d, want %d: %q", lipgloss.Width(line), layout.Width, line)
		}
	}

	next, _ = m.Update(click(layout.X, layout.Y+2))
	m = next.(Model)
	if m.dialog.IsOpen() || m.requestContext.list.Current().Key != "app.customSignal['environment']" {
		t.Fatal("clicking the value line did not select the fact")
	}
	next, _ = m.Update(click(layout.X+3, layout.Y+1))
	m = next.(Model)
	if !m.dialog.IsOpen() || !strings.Contains(m.dialog.InputValue(), "app.customSignal['environment']=") {
		t.Fatal("clicking both lines did not open the editor for the same fact")
	}
}

func TestContextProjectsTemplateWithoutChangingValueIndexes(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	m := viewTestModel(132, 45, panels.Parameters)
	project := core.Project{ProjectID: "demo", ProjectNumber: "123456789"}
	config := &firebase.RemoteConfig{
		Version: firebase.RemoteConfigVersion{VersionNumber: "18"},
		Conditions: []firebase.RemoteConfigCondition{
			{Name: "ios_only", Expression: "device.os == 'ios'"},
			{Name: "web_beta", Expression: "device.os == 'web' && app.customSignal['environment'].exactlyMatches(['beta'])"},
			{Name: "web_all", Expression: "device.os == 'web'"},
		},
		Parameters: map[string]firebase.RemoteConfigParam{
			"flag": {DefaultValue: &firebase.RemoteConfigValue{Value: "baseline"}, ConditionalValues: map[string]firebase.RemoteConfigValue{
				"ios_only": {Value: "ios-value"}, "web_beta": {Value: "beta-value"}, "web_all": {Value: "web-value"},
			}},
		},
	}
	selection := messages.ProjectsSelectionChangedMsg{Projects: []core.Project{project}}
	m.parameters, _ = m.parameters.Update(selection)
	m.conditions, _ = m.conditions.Update(selection)
	m.parameters, _ = m.parameters.Update(messages.ParametersLoadedMsg{
		Project: project, Tree: parameters.BuildTree(config, time.Time{}, ""), HasDraft: true, Source: "draft",
	})
	m.conditions, _ = m.conditions.Update(messages.ConditionsLoadedMsg{
		Project: project, Tree: conditions.BuildTree(config, time.Time{}, ""), Source: "draft",
	})
	if !m.parameters.FocusValue(project.ProjectID, rootgroup.TreeKey, "flag", 2) {
		t.Fatal("could not select web value in the original tree")
	}
	data, _ := m.parameters.CurrentParameterViewData()
	m.details = m.details.SetData(data)
	m.detailsVisible = true
	m.applyContextFact(contextFactMsg{assignment: "device.os=web"})

	for name, view := range map[string]string{
		"parameters": m.parameters.ViewWithBorder(true, true),
		"conditions": m.conditions.ViewWithBorder(true, true),
		"details":    m.details.ViewWithBorder(true),
	} {
		view = ansi.Strip(view)
		if strings.Contains(view, "ios_only") || strings.Contains(view, "ios-value") {
			t.Errorf("%s shows excluded iOS condition:\n%s", name, view)
		}
		if !strings.Contains(view, "web_beta") || !strings.Contains(view, "Needs input") {
			t.Errorf("%s omits unresolved beta condition:\n%s", name, view)
		}
	}
	current, _ := m.parameters.CurrentParameterViewData()
	if current.SelectedValueIdx != 2 || len(current.Parameter.Values) != 4 {
		t.Fatalf("projection changed selection or edit data: %#v", current)
	}
	if selected, ok := m.details.SelectedParameterValue(); !ok || selected.Label != "web_all" {
		t.Fatalf("Details selection changed: %+v, %v", selected, ok)
	}

	m.details = m.details.SetSelectedValue("pending-web-edit")
	m.applyContextFact(contextFactMsg{assignment: "app.customSignal['environment']=beta"})
	for name, view := range map[string]string{
		"parameters": m.parameters.ViewWithBorder(true, true),
		"details":    m.details.ViewWithBorder(true),
	} {
		view = ansi.Strip(view)
		if !strings.Contains(view, "beta-value") || !strings.Contains(view, "selected") ||
			strings.Contains(view, "ios_only") || strings.Contains(view, "web_all") || strings.Contains(view, "baseline") {
			t.Errorf("%s should show only the selected beta value:\n%s", name, view)
		}
	}
	m.useContextCapture(evaluate.Exchange{
		Project: "demo", Entry: 3, Version: "17", HasResponse: true,
		Entries: map[string]string{"flag": "captured-value"}, Facts: m.requestContext.facts,
	}, "capture.har")
	explainKey := tea.KeyPressMsg(tea.Key{Code: 'e', Mod: tea.ModCtrl})
	next, _ := m.Update(explainKey)
	viewer := next.(Model)
	view := viewer.explainView.View()
	plain := ansi.Strip(view)
	if !viewer.explainView.IsOpen() || !strings.HasPrefix(view, styles.BorderStyle(true).Render("╭─")) ||
		lipgloss.Width(view) != m.width-4 || lipgloss.Height(view) != m.height-4 {
		t.Fatalf("explanation should open in the blue expanded viewer, got %dx%d", lipgloss.Width(view), lipgloss.Height(view))
	}
	for _, want := range []string{"Template v18 · draft", "Request: capture.har · entry 3", "does not match", "eval: FALSE"} {
		if !strings.Contains(plain, want) {
			t.Errorf("explanation is missing %q:\n%s", want, plain)
		}
	}
	next, _ = next.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnd}))
	plain = ansi.Strip(next.(Model).explainView.View())
	for _, want := range []string{"Winner: [2] web_beta", "Captured response: captured-value", "Capture template v17 differs from evaluated template v18"} {
		if !strings.Contains(plain, want) {
			t.Errorf("explanation is missing %q:\n%s", want, plain)
		}
	}
	next, _ = next.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	view = next.(Model).explainView.View()
	if lipgloss.Width(view) != 76 || lipgloss.Height(view) != 20 {
		t.Fatalf("resized explanation = %dx%d, want 76x20", lipgloss.Width(view), lipgloss.Height(view))
	}
	next, _ = next.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape}))
	if next.(Model).explainView.IsOpen() {
		t.Fatal("Esc should close the explanation")
	}

	m, _, _ = m.updateContextKey("H", false)
	if !m.details.Dirty() {
		t.Fatal("showing excluded rows discarded the pending Details edit")
	}
	red := lipgloss.NewStyle().Foreground(styles.PaletteError)
	green := lipgloss.NewStyle().Foreground(styles.PaletteSuccess)
	selected := lipgloss.NewStyle().Foreground(styles.PaletteSlateBright).Background(styles.PaletteHistoryAddedBackground)
	for name, view := range map[string]string{
		"parameters": m.parameters.ViewWithBorder(true, true),
		"details":    m.details.ViewWithBorder(true),
	} {
		for _, note := range []string{red.Render("does not match"), selected.Render("selected"), green.Render("lower priority"), red.Render("not used")} {
			if !strings.Contains(view, note) {
				t.Errorf("%s missing the CLI-styled outcome %q:\n%s", name, ansi.Strip(note), ansi.Strip(view))
			}
		}
	}
	if !strings.Contains(ansi.Strip(m.conditions.ViewWithBorder(true, true)), "ios_only · Excluded") {
		t.Fatal("show excluded did not restore the condition with its status")
	}
	m, _, _ = m.updateContextKey("H", false)
	m.applyContextFact(contextFactMsg{assignment: "device.os=android"})
	for name, view := range map[string]string{
		"parameters": m.parameters.ViewWithBorder(true, true),
		"details":    m.details.ViewWithBorder(true),
	} {
		plain := ansi.Strip(view)
		if !strings.Contains(view, selected.Render("selected")) || !strings.Contains(plain, "baseline") ||
			strings.Contains(plain, "web_beta") || strings.Contains(plain, "web_all") {
			t.Errorf("%s should show the default as the only selected value:\n%s", name, plain)
		}
	}
	m, _, _ = m.updateContextKey("space", false)
	if m.requestContext.scenario != nil || !strings.Contains(ansi.Strip(m.conditions.ViewWithBorder(true, true)), "ios_only") {
		t.Fatal("pausing did not restore the complete template")
	}
}

func TestContextCapturePasteAndHARSelection(t *testing.T) {
	m := viewTestModel(90, 24, panels.Parameters)
	m.setProjectsMode(projectsPanelModeCollapsed)
	next, _ := m.Update(tea.KeyPressMsg(tea.Key{Code: 'x', Mod: tea.ModCtrl}))
	m = next.(Model)
	if m.requestContext.overlay != contextOverview {
		t.Fatal("Context should open when Projects is collapsed")
	}
	next, _ = m.Update(keyPress('i'))
	m = next.(Model)
	view := ansi.Strip(m.dialog.View())
	if !strings.Contains(view, "Paste a Firebase fetch cURL command or HAR content here") || strings.Contains(view, "Paste cURL") {
		t.Fatalf("import popup should offer direct paste and no paste button:\n%s", view)
	}
	const url = "https://firebaseremoteconfig.googleapis.com/v1/projects/123456789/namespaces/firebase:fetch?key=example"
	command := fmt.Sprintf("curl '%s' \\\n  --data-raw '{\"app_id\":\"1:123456789:web:abc\"}'", url)
	m = pasteContextCapture(t, m, command)
	client := core.Project{ProjectID: "demo", ProjectNumber: "123456789"}
	server := core.Project{ProjectID: "server@demo", ProjectNumber: "123456789"}
	if !m.requestContext.scenario.AppliesTo(client) || m.requestContext.scenario.AppliesTo(server) {
		t.Fatal("cURL context was not scoped to the captured client template")
	}
	if got := m.requestContext.facts["app.id"].Value; got != "1:123456789:web:abc" {
		t.Fatalf("pasted cURL input = %q", got)
	}

	const harFormat = `{"log":{"entries":[
		{"request":{"method":"GET","url":"https://example.com/"}},
		{"request":{"method":"POST","url":%q,"postData":{"text":"{\"app_id\":\"1:123456789:web:abc\"}"}}},
		{"request":{"method":"POST","url":%q,"postData":{"text":"{}"}}}
	]}}`
	har := fmt.Sprintf(harFormat, url, strings.Replace(url, "firebase:fetch", "firebase-server:fetch", 1))
	for _, source := range []string{"clipboard", "file"} {
		t.Run(source, func(t *testing.T) {
			next, _ := m.Update(keyPress('i'))
			if source == "clipboard" {
				next = pasteContextCapture(t, next, har)
			} else {
				path := filepath.Join(t.TempDir(), "requests.har")
				if err := os.WriteFile(path, []byte(har), 0o600); err != nil {
					t.Fatal(err)
				}
				x, y := dialogLabelPoint(t, next.(Model), "Choose File")
				var choose tea.Cmd
				next, choose = next.Update(tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft})
				if choose == nil {
					t.Fatal("clicking Choose File should open the file picker")
				}
				next, _ = next.Update(choose())
				picker := next.(Model)
				if !picker.projectIO.IsOpen() {
					t.Fatal("file picker did not open")
				}
				picker.projectIO = picker.projectIO.Close()
				var load tea.Cmd
				next, load = picker.Update(projectio.CaptureChosenMsg{Path: path})
				if load == nil {
					t.Fatal("could not load the selected HAR file")
				}
				next, _ = next.Update(load())
			}
			loaded := next.(Model)
			if loaded.requestContext.overlay != contextRequests || !strings.Contains(ansi.Strip(loaded.View().Content), "Entry 3") {
				t.Fatal("multi-request HAR did not show the original entry numbers")
			}
			layout := loaded.contextLayout(true)
			next, _ = next.Update(tea.MouseClickMsg{X: layout.X + 3, Y: layout.Y + 4, Button: tea.MouseLeft})
			next, _ = next.Update(tea.MouseClickMsg{X: layout.X + 3, Y: layout.Y + 3, Button: tea.MouseLeft})
			scenario := next.(Model).requestContext.scenario
			if scenario.AppliesTo(client) || !scenario.AppliesTo(server) {
				t.Fatal("HAR entry selection did not apply the server template scope")
			}
		})
	}
}

func pasteContextCapture(t *testing.T, m tea.Model, content string) Model {
	t.Helper()
	next, pasted := m.Update(tea.PasteMsg{Content: content})
	if pasted == nil || next.(Model).stringInput.IsOpen() {
		t.Fatal("pasting should submit the capture directly from the import popup")
	}
	next, load := next.Update(pasted())
	if load == nil {
		t.Fatal("pasting should start reading the capture")
	}
	next, _ = next.Update(load())
	return next.(Model)
}

func TestContextTemplateScopeCanBeAddedManually(t *testing.T) {
	m := viewTestModel(80, 24, panels.Context)
	next, _ := m.Update(keyPress('a'))
	next, _ = next.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnd}))
	next, _ = next.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyUp}))
	next, _ = next.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	if !next.(Model).dialog.IsOpen() {
		t.Fatal("choosing Template scope should open its editor")
	}
	next, _ = next.Update(tea.PasteMsg{Content: "server@123456789"})
	next, apply := next.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	if apply == nil {
		t.Fatal("Enter should apply the scope")
	}
	next, _ = next.Update(apply())
	m = next.(Model)
	client := core.Project{ProjectID: "demo", ProjectNumber: "123456789"}
	server := core.Project{ProjectID: "server@demo", ProjectNumber: "123456789"}
	scenario := m.requestContext.scenario
	if !scenario.AppliesTo(server) || scenario.AppliesTo(client) || scenario.AppliesTo(core.Project{ProjectID: "server@other"}) {
		t.Fatal("manual scope should match only the server template with the given project number")
	}
	if row := m.requestContext.list.Current(); row.Key != contextScopeKey || row.Value != "server@123456789" {
		t.Fatalf("manual scope row = %+v", row)
	}
}

func TestImportedContextScopeCanBeEditedAndRemovedIndependently(t *testing.T) {
	m := viewTestModel(90, 24, panels.Context)
	m.useContextCapture(evaluate.Exchange{
		Project: "demo", Namespace: "firebase-server",
		Facts: map[string]evaluate.Fact{
			"device.os": {Value: "web", Known: true},
		},
	}, "capture.har")
	next, _ := m.Update(tea.KeyPressMsg(tea.Key{Code: 'x', Mod: tea.ModCtrl}))
	m = next.(Model)
	layout := m.contextLayout(true)
	if layout.Title != "Context" || layout.Count != 2 {
		t.Fatalf("popup title = %q, count = %d, want separate scope row", layout.Title, layout.Count)
	}
	if row := m.requestContext.list.Current(); row.Key != contextScopeKey || row.Value != "server@demo" {
		t.Fatalf("imported scope row = %+v", row)
	}
	click := tea.MouseClickMsg{X: layout.X + 2, Y: layout.Y + 1, Button: tea.MouseLeft}
	next, _ = m.Update(click)
	next, _ = next.Update(click)
	m = next.(Model)
	if !m.dialog.IsOpen() || m.dialog.InputValue() != "server@demo" {
		t.Fatal("double-clicking the scope should edit its canonical target")
	}
	next, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape}))
	next, _ = next.Update(contextFactMsg{original: contextScopeKey, assignment: "client@other"})
	m = next.(Model)
	other := core.Project{ProjectID: "other"}
	demo := core.Project{ProjectID: "server@demo"}
	if !m.requestContext.scenario.AppliesTo(other) || m.requestContext.scenario.AppliesTo(demo) ||
		m.requestContext.list.Current().Value != "other" {
		t.Fatal("editing the imported scope should use the canonical client target")
	}
	next, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyDelete}))
	m = next.(Model)
	scenario := m.requestContext.scenario
	if m.requestContext.scope.ProjectID != "" || m.contextLayout(true).Count != 1 ||
		m.requestContext.facts["device.os"].Value != "web" || !scenario.AppliesTo(other) || !scenario.AppliesTo(demo) {
		t.Fatal("deleting the scope should apply the retained request inputs to both templates")
	}
}

func TestContextPopupCapturesMouseOverLogs(t *testing.T) {
	m := viewTestModel(80, 20, panels.Logs)
	if m.mouseMode() != tea.MouseModeNone {
		t.Fatal("Logs should allow native selection")
	}
	next, _ := m.Update(tea.KeyPressMsg(tea.Key{Code: 'x', Mod: tea.ModCtrl}))
	m = next.(Model)
	if m.mouseMode() != tea.MouseModeAllMotion {
		t.Fatal("the open Context popup should enable mouse reporting")
	}
	next, _ = m.Update(tea.MouseClickMsg{X: 0, Y: 19, Button: tea.MouseLeft})
	m = next.(Model)
	if m.active != panels.Logs || m.mouseMode() != tea.MouseModeNone {
		t.Fatal("closing Context should return native selection to Logs")
	}
}

func TestContextPopupHeaderShowsShortcutAndCount(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	m := viewTestModel(80, 20, panels.Parameters)
	m.applyContextFact(contextFactMsg{assignment: "device.os=web"})
	next, _ := m.Update(tea.KeyPressMsg(tea.Key{Code: 'x', Mod: tea.ModCtrl}))
	m = next.(Model)
	layout := m.contextLayout(true)
	header, body, _ := strings.Cut(m.requestContext.list.View(layout, true), "\n")
	if got := ansi.Strip(header); !strings.HasPrefix(got, "╭─ ˆˣContext ") ||
		!strings.HasSuffix(got, " 1 ─╮") || lipgloss.Width(got) != layout.Width {
		t.Fatalf("Context popup header = %q, want shortcut and count in a %d-column border", got, layout.Width)
	}
	if !strings.Contains(header, styles.SecondaryTitleCount.Render(" 1 ")) {
		t.Fatal("Context popup count does not use the sidebar count style")
	}
	shortcutStyle := styles.FilterText.Background(styles.TitleStyle(true).GetBackground())
	if !strings.Contains(header, shortcutStyle.Render("ˆˣ")) {
		t.Fatal("Context popup header does not highlight the shortcut like Accounts")
	}
	branchStyle := styles.TreeBranch.Background(styles.TreeItemSelectionStyle().GetBackground())
	if !strings.Contains(ansi.Strip(body), "│ ╰ web") || !strings.Contains(body, branchStyle.Render(" ╰ ")) {
		t.Fatal("Context popup value does not use the shared dim branch style")
	}
}

func TestContextPopupGrowsAndClearShrinksIt(t *testing.T) {
	m := viewTestModel(80, 40, panels.Parameters)
	m.useContextCapture(evaluate.Exchange{
		Project: "demo",
		Facts: map[string]evaluate.Fact{
			"device.os": {Value: "web", Known: true},
		},
	}, "capture.har")
	next, _ := m.Update(tea.KeyPressMsg(tea.Key{Code: 'x', Mod: tea.ModCtrl}))
	m = next.(Model)
	checkHeight := func(want int) {
		t.Helper()
		layout := m.contextLayout(true)
		view := m.requestContext.list.View(layout, true)
		if layout.Height != want || lipgloss.Height(view) != want || layout.Y != (m.height-want)/2 {
			t.Fatalf("Context popup height = %d, rendered height = %d, top = %d, want height %d centered", layout.Height, lipgloss.Height(view), layout.Y, want)
		}
		lines := strings.Split(ansi.Strip(view), "\n")
		footer := lines[len(lines)-2]
		for _, label := range []string{"esc close", "a add", "i import", "X clear"} {
			if !strings.Contains(footer, label) {
				t.Fatalf("footer %q is missing %q", footer, label)
			}
		}
		if !strings.Contains(footer, " • ") || strings.Contains(footer, "[") {
			t.Fatalf("footer %q should use the shared help style", footer)
		}
		if !strings.HasSuffix(strings.TrimSpace(strings.TrimSuffix(footer, "│")), "esc close") {
			t.Fatalf("footer %q should end with esc close", footer)
		}
	}

	checkHeight(7)
	m.applyContextFact(contextFactMsg{assignment: "app.version=2"})
	checkHeight(9)
	for i := range 24 {
		m.applyContextFact(contextFactMsg{assignment: fmt.Sprintf("app.customSignal['fact_%d']=value", i)})
	}
	checkHeight(m.height - 4)
	next, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeySpace}))
	m = next.(Model)
	badge, _ := m.contextBadge()
	if !strings.Contains(m.contextLayout(true).Title, "Context ⏸") || badge != styles.PanelTitleInactiveTab.Padding(0, 1).Render("Context ⏸") {
		t.Fatal("paused Context should show the pause symbol and use the inactive tab badge style")
	}
	checkHeight(m.height - 4)
	next, _ = m.Update(keyPress('X'))
	m = next.(Model)
	if len(m.requestContext.facts) != 0 || m.requestContext.capture != nil ||
		m.requestContext.scope.ProjectID != "" || m.requestContext.scenario != nil ||
		m.requestContext.paused || m.requestContext.overlay != contextOverview {
		t.Fatal("X should clear all Context inputs and capture scope while keeping the popup open")
	}
	checkHeight(5)
}

func TestContextHeaderFocusesWithoutResizing(t *testing.T) {
	m := viewTestModel(80, 20, panels.Projects)
	m.logsHeight = 12
	m.applyLayout()
	layout := m.contextLayout(false)
	if layout.Height != 4 {
		t.Fatalf("empty Context at minimum workspace height = %d, want four rows", layout.Height)
	}
	next, _ := m.Update(tea.MouseWheelMsg{X: 3, Y: layout.Y + 1, Button: tea.MouseWheelDown})
	if next.(Model).active != panels.Projects {
		t.Fatal("scrolling over Context stole focus from Projects")
	}
	next, _ = m.Update(tea.MouseClickMsg{X: 3, Y: layout.Y, Button: tea.MouseLeft})
	focused := next.(Model)
	if focused.active != panels.Context || focused.requestContext.overlay != contextInline ||
		focused.contextLayout(false).Height != layout.Height {
		t.Fatal("clicking the Context header should focus it in place")
	}
}

func TestContextResizesWithInputsAndTerminal(t *testing.T) {
	m := viewTestModel(80, 24, panels.Projects)
	checkHeight := func(want int) {
		t.Helper()
		layout := m.contextLayout(false)
		sidebar := m.workspaceLayout()
		if layout.Height != want || layout.Y+layout.Height != sidebar.topHeight {
			t.Fatalf("Context at height %d occupies rows %d through %d, want %d rows at bottom of sidebar", m.height, layout.Y, layout.Y+layout.Height, want)
		}
	}
	checkHeight(4)
	for i := range 12 {
		m.applyContextFact(contextFactMsg{assignment: fmt.Sprintf("app.customSignal['fact_%d']=value", i)})
		if i < 3 {
			checkHeight(4 + i*2)
		}
	}
	checkHeight(10) // Two-thirds of 16 rows, rounded down to whole items.
	m.updateWindowSize(tea.WindowSizeMsg{Width: 80, Height: 31})
	checkHeight(14)
	m.updateWindowSize(tea.WindowSizeMsg{Width: 80, Height: 20})
	checkHeight(8)
	for i := range 12 {
		m.applyContextFact(contextFactMsg{original: fmt.Sprintf("app.customSignal['fact_%d']", i), remove: true})
	}
	checkHeight(4)
}

func TestTabSwitchesBetweenContextAndLogs(t *testing.T) {
	m := viewTestModel(80, 24, panels.Projects)
	m = updateFocus(t, m, keyPress('9'), panels.Context)
	m = updateFocus(t, m, keyPress('0'), panels.Logs)
	m = pressTab(t, m, panels.Context)
	pressTab(t, m, panels.Logs)
}

func TestContextCollapseKeepsSeparateRailsAndFocus(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	m := viewTestModel(80, 24, panels.Context)
	m.applyContextFact(contextFactMsg{assignment: "device.os=web"})
	expanded := m.contextLayout(false)
	next, _ := m.Update(keyPress('c'))
	collapsed := next.(Model)
	layout := collapsed.contextLayout(false)
	if collapsed.active != panels.Context || collapsed.projectsMode != projectsPanelModeCollapsed ||
		layout.Height != expanded.Height || layout.Y != expanded.Y || layout.Width != collapsedProjectsWidth {
		t.Fatal("collapsing should preserve both panel heights and keep Context focused")
	}
	sidebar := strings.Split(collapsed.sidebarView(false, true), "\n")
	if len(sidebar) != collapsed.workspaceLayout().topHeight {
		t.Fatalf("collapsed sidebar has %d rows", len(sidebar))
	}
	if !strings.HasPrefix(ansi.Strip(sidebar[layout.Y-1]), "─╯") {
		t.Fatal("Projects border should end above Context")
	}
	context := strings.Join(sidebar[layout.Y:], "\n")
	if strings.TrimRight(ansi.Strip(sidebar[layout.Y]), " ") != "─╮" ||
		!strings.Contains(context, styles.BorderStyle(true).Render("│")) {
		t.Fatalf("empty-title Context rail does not have its own active border: %q", context)
	}

	next, _ = collapsed.Update(keyPress('1'))
	next, _ = next.Update(keyPress('9'))
	focused := next.(Model)
	if focused.active != panels.Context || focused.requestContext.overlay != contextInline {
		t.Fatal("9 should focus the collapsed Context rail")
	}
	ctrlX := tea.KeyPressMsg(tea.Key{Code: 'x', Mod: tea.ModCtrl})
	popup, _ := focused.Update(ctrlX)
	if popup.(Model).requestContext.overlay != contextOverview {
		t.Fatal("Ctrl+X should open the Context popup")
	}
	next, _ = popup.Update(ctrlX)
	if next.(Model).requestContext.overlay != contextInline || next.(Model).active != panels.Context {
		t.Fatal("closing the popup should return focus to the Context rail")
	}

	next, _ = focused.Update(keyPress('c'))
	restored := next.(Model)
	if restored.active != panels.Context || restored.projectsMode != projectsPanelModeExpanded ||
		restored.contextLayout(false).Height != expanded.Height || restored.requestContext.facts["device.os"].Value != "web" {
		t.Fatal("expanding Context should restore the sidebar and its facts")
	}
}

func TestCollapsedContextBadgeRequiresInputs(t *testing.T) {
	m := viewTestModel(80, 24, panels.Context)
	next, _ := m.Update(keyPress('c'))
	m = next.(Model)
	if m.projectsMode != projectsPanelModeCollapsed {
		t.Fatal("sidebar should be collapsed")
	}
	if badge, _ := m.contextBadge(); badge != "" {
		t.Fatalf("empty Context displayed a badge: %q", badge)
	}

	m.applyContextFact(contextFactMsg{assignment: "device.os=web"})
	badge, x := m.contextBadge()
	if strings.TrimSpace(ansi.Strip(badge)) != "Context on" || !m.contextBadgeAt(x, 0) {
		t.Fatalf("active Context badge = %q, want a clickable Context on badge", badge)
	}
	next, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeySpace}))
	m = next.(Model)
	if badge, _ := m.contextBadge(); strings.TrimSpace(ansi.Strip(badge)) != "Context"+contextPauseMark {
		t.Fatalf("paused Context badge = %q", badge)
	}
	next, _ = m.Update(keyPress('X'))
	m = next.(Model)
	if badge, _ := m.contextBadge(); badge != "" || m.contextBadgeAt(x, 0) {
		t.Fatalf("cleared Context retained a badge or hit target: %q", badge)
	}
}

func TestClickCollapsedContextRailOnlyFocusesIt(t *testing.T) {
	m := viewTestModel(80, 24, panels.Projects)
	m.applyContextFact(contextFactMsg{assignment: "device.os=web"})
	m.setProjectsMode(projectsPanelModeCollapsed)
	click := tea.MouseClickMsg{X: 1, Y: m.contextLayout(false).Y + 1, Button: tea.MouseLeft}
	next, _ := m.Update(click)
	next, _ = next.Update(click)
	focused := next.(Model)
	if focused.active != panels.Context || focused.projectsMode != projectsPanelModeCollapsed ||
		focused.requestContext.overlay != contextInline || focused.dialog.IsOpen() {
		t.Fatal("clicking the Context rail should focus it without editing a hidden input")
	}
}
