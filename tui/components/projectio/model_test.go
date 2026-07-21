package projectio

import (
	"os"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/tui/components/viewutil"
)

func TestImportOptionsAndConflictResolution(t *testing.T) {
	m, _ := New().SetBounds(0, 0, 100, 30).OpenImport(core.Project{Name: "Demo", ProjectID: "demo"})
	m.phase = phaseImportOptions
	m.sourceRaw = []byte(`{"parameters":{}}`)
	m.sourcePath = "config.json"
	m.optionCursor = optionStrategy
	m, _ = m.Update(key("space"))
	if !m.OptionSelectorOpen() {
		t.Fatal("strategy selector did not open")
	}
	m, _ = m.Update(key("down"))
	m, _ = m.Update(key("enter"))
	if m.importOptions().Strategy != core.ProjectImportReplace {
		t.Fatalf("strategy = %q", m.importOptions().Strategy)
	}
	m.optionCursor = optionConditions
	m, _ = m.Update(key("space"))
	if !m.OptionSelectorOpen() {
		t.Fatal("condition selector did not open")
	}
	m, _ = m.Update(key("down"))
	m, _ = m.Update(key("enter"))
	if m.importOptions().ConditionPolicy != core.ProjectImportKeepPortableConditions {
		t.Fatalf("condition policy = %q", m.importOptions().ConditionPolicy)
	}
	m = m.OpenConflicts([]core.ProjectImportConflict{{ID: "parameter:flag", Label: "parameter flag"}})
	m, _ = m.Update(key("right"))
	if m.resolutions["parameter:flag"] != core.ProjectImportUseImported {
		t.Fatalf("resolution = %q", m.resolutions["parameter:flag"])
	}
}

func TestExportPathProducesRequest(t *testing.T) {
	m, _ := New().SetBounds(0, 0, 90, 24).OpenExport(core.Project{ProjectID: "demo"}, false)
	m.pathInput.SetValue("out.json")
	m, cmd := m.Update(key("enter"))
	if cmd == nil {
		t.Fatal("export command is nil")
	}
	msg, ok := cmd().(ExportRequestedMsg)
	if !ok || msg.Path != "out.json" || msg.Project.ProjectID != "demo" || msg.Draft {
		t.Fatalf("export request = %#v", msg)
	}
	if !m.IsOpen() {
		t.Fatal("model closed before request was handled")
	}
}

func TestDefaultsFormatAndPathProduceRequest(t *testing.T) {
	project := core.Project{Name: "Demo", ProjectID: "demo"}
	m, _ := New().SetBounds(0, 0, 90, 24).OpenDefaults(project)
	if m.Mode() != ModeDefaults {
		t.Fatalf("mode = %v, want defaults", m.Mode())
	}
	view := ansi.Strip(m.View())
	for _, want := range []string{"Download Application Defaults · Format", "JSON · Web", "XML · Android", "plist · Apple"} {
		if !strings.Contains(view, want) {
			t.Fatalf("format view missing %q:\n%s", want, view)
		}
	}

	m, _ = m.Update(key("down"))
	m, _ = m.Update(key("enter"))
	if m.phase != phaseDefaultsPath {
		t.Fatalf("phase = %v, want defaults path", m.phase)
	}
	if got := m.pathInput.Value(); got != "demo-remote-config-defaults.xml" {
		t.Fatalf("default path = %q", got)
	}
	m.pathInput.SetValue("android.xml")
	m, cmd := m.Update(key("enter"))
	if cmd == nil {
		t.Fatal("defaults command is nil")
	}
	request, ok := cmd().(DefaultsRequestedMsg)
	if !ok || request.Project.ProjectID != "demo" || request.Path != "android.xml" || request.Format != firebase.DefaultsFormatXML {
		t.Fatalf("defaults request = %#v", request)
	}
	if !m.IsOpen() {
		t.Fatal("model closed before request was handled")
	}
}

func TestOptionsViewIdentifiesProjectAndSource(t *testing.T) {
	m, _ := New().SetBounds(0, 0, 100, 30).OpenImport(core.Project{Name: "Demo", ProjectID: "demo"})
	m.phase = phaseImportOptions
	m.sourcePath = "/tmp/config.json"
	m.summary = core.ProjectImportSummary{RootParameters: 2, Groups: 1, Conditions: 5, NonPortableConditions: 3}
	view := ansi.Strip(m.View())
	for _, want := range []string{"Demo (demo)", "/tmp/config.json", "2 parameters", "1 group", "5 conditions", "Keep all conditions (5 kept)", "Review Changes", "Cancel"} {
		if !strings.Contains(view, want) {
			t.Fatalf("view missing %q:\n%s", want, view)
		}
	}
	if strings.Contains(ansi.Strip(strings.Join(m.importOptionBaseLines(72), "\n")), "Review Changes") {
		t.Fatal("Review Changes remains in the option rows")
	}
}

func TestConditionPolicyChoicesShowSourceCounts(t *testing.T) {
	m, _ := New().SetBounds(0, 0, 100, 30).OpenImport(core.Project{Name: "Demo", ProjectID: "demo"})
	m.phase = phaseImportOptions
	m.summary = core.ProjectImportSummary{Conditions: 5, NonPortableConditions: 3}
	m.optionCursor = optionConditions

	tests := []struct {
		policy core.ProjectConditionPolicy
		want   string
	}{
		{core.ProjectImportKeepConditions, "Keep all conditions (5 kept)"},
		{core.ProjectImportKeepPortableConditions, "Keep portable conditions only (2 kept · 3 removed)"},
		{core.ProjectImportRemoveAllConditions, "Remove all conditions (5 removed)"},
	}
	for _, test := range tests {
		m.conditionPolicy = test.policy
		want := test.want
		if got := ansi.Strip(strings.Join(m.importOptionLines(72), "\n")); !strings.Contains(got, want) {
			t.Fatalf("condition choice missing %q:\n%s", want, got)
		}
	}
}

func TestFeatureSelectorsUseProfileSelectorStyle(t *testing.T) {
	project := core.Project{Name: "Demo", ProjectID: "demo"}
	m, _ := New().SetBounds(0, 0, 100, 30).OpenImport(project)
	m.phase = phaseImportOptions
	options := m.importOptionLines(72)
	assertSelectorRows(t, options[6:12])

	m = m.OpenConflicts([]core.ProjectImportConflict{
		{ID: "parameter:first", Label: "parameter first"},
		{ID: "parameter:second", Label: "parameter second"},
	})
	conflicts := m.importConflictLines(72)
	assertSelectorRows(t, conflicts[4:6])

	m, _ = New().SetBounds(0, 0, 100, 30).OpenExport(project, true)
	sources := m.exportSourceLines()
	assertSelectorRows(t, sources[4:6])

	m, _ = New().SetBounds(0, 0, 100, 30).OpenDefaults(project)
	formats := m.defaultsFormatLines()
	assertSelectorRows(t, formats[4:7])
}

func assertSelectorRows(t *testing.T, rows []string) {
	t.Helper()
	for index, row := range rows {
		if !strings.HasPrefix(row, "  ") || strings.HasPrefix(row, "   ") {
			t.Fatalf("selector row %d = %q, want exactly two unstyled leading spaces", index, row)
		}
		if strings.Contains(ansi.Strip(row), "▸") {
			t.Fatalf("selector row %d retains arrow: %q", index, ansi.Strip(row))
		}
	}
}

func TestImportFilePickerMatchesAuthPickerLayout(t *testing.T) {
	m, _ := New().SetBounds(0, 0, 100, 30).OpenImport(core.Project{Name: "Demo", ProjectID: "demo"})
	lines := m.importFileLines(72)

	if got := ansi.Strip(lines[0]); got != "Choose a Remote Config JSON file." {
		t.Fatalf("instruction = %q", got)
	}
	if got := ansi.Strip(lines[1]); got != "" {
		t.Fatalf("line before project = %q, want blank", got)
	}
	if got := ansi.Strip(lines[2]); got != "Project: Demo (demo)" {
		t.Fatalf("project line = %q", got)
	}
	if got := ansi.Strip(lines[3]); got != "" {
		t.Fatalf("line before picker = %q, want blank", got)
	}
	picker := strings.TrimRight(m.picker.View(), "\n")
	wantPicker := " " + strings.ReplaceAll(picker, "\n", "\n ")
	if lines[4] != wantPicker {
		t.Fatalf("picker inset does not match auth picker:\n got: %q\nwant: %q", lines[4], wantPicker)
	}
	if strings.HasSuffix(lines[4], "\n ") {
		t.Fatalf("picker retains a trailing empty line: %q", lines[4])
	}
	if got := ansi.Strip(lines[5]); got != "" {
		t.Fatalf("line after picker = %q, want one blank", got)
	}
	if got := ansi.Strip(lines[6]); !strings.Contains(got, "Choose") || !strings.Contains(got, "Cancel") {
		t.Fatalf("picker buttons = %q, want Choose and Cancel", got)
	}
	for index, line := range strings.Split(lines[4], "\n") {
		if !strings.HasPrefix(ansi.Strip(line), " ") {
			t.Fatalf("picker line %d = %q, want one-cell left inset", index, ansi.Strip(line))
		}
	}
}

func TestOptionsUseBoxSelectorsAndReviewButton(t *testing.T) {
	m, _ := New().SetBounds(0, 0, 100, 30).OpenImport(core.Project{Name: "Demo", ProjectID: "demo"})
	m.phase = phaseImportOptions
	m.sourceRaw = []byte("{\"parameters\":{}}")
	m.sourcePath = "config.json"

	m, _ = m.Update(key("enter"))
	if !m.OptionSelectorOpen() {
		t.Fatal("Enter on Strategy did not open the box selector")
	}
	if view := ansi.Strip(m.OptionSelectorListView()); !strings.Contains(view, "Merge (keep current conflicts)") || !strings.Contains(view, "Replace entire config") {
		t.Fatalf("strategy selector options missing:\n%s", view)
	}
	m, _ = m.Update(key("esc"))
	if m.OptionSelectorOpen() || !m.IsOpen() {
		t.Fatal("Esc should close only the option selector")
	}

	m.buttonsFocused, m.buttonCursor = true, 0
	next, cmd := m.Update(key("enter"))
	if cmd == nil || next.phase != phaseImportWorking {
		t.Fatalf("Review Changes button phase=%v cmd nil=%v", next.phase, cmd == nil)
	}
	if _, ok := cmd().(ImportPlanRequestedMsg); !ok {
		t.Fatalf("Review Changes command returned %T", cmd())
	}
}

func TestAffectedStepsRenderClickableButtons(t *testing.T) {
	project := core.Project{Name: "Demo", ProjectID: "demo"}
	tests := []struct {
		name  string
		model Model
		label string
	}{
		{name: "import file cancel", model: importFileTestModel(project), label: "Cancel"},
		{name: "import options cancel", model: importOptionsTestModel(project), label: "Cancel"},
		{name: "export source cancel", model: exportSourceTestModel(project), label: "Cancel"},
		{name: "export path cancel", model: exportPathTestModel(project), label: "Cancel"},
		{name: "defaults format cancel", model: defaultsFormatTestModel(project), label: "Cancel"},
		{name: "defaults path cancel", model: defaultsPathTestModel(project), label: "Cancel"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			x, y := projectIOTextPoint(t, test.model, test.label)
			next, _ := test.model.Update(tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft})
			if next.IsOpen() {
				t.Fatalf("%s click left popup open", test.label)
			}
		})
	}
}

func TestExportContinueButtonAcceptsMouseClick(t *testing.T) {
	m := exportSourceTestModel(core.Project{Name: "Demo", ProjectID: "demo"})
	x, y := projectIOTextPoint(t, m, "Continue")
	next, _ := m.Update(tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft})
	if next.phase != phaseExportPath {
		t.Fatalf("phase = %v, want export path", next.phase)
	}
}

func TestExportButtonAcceptsMouseClick(t *testing.T) {
	m := exportPathTestModel(core.Project{Name: "Demo", ProjectID: "demo"})
	m.pathInput.SetValue("out.json")
	x, y := projectIOTextPoint(t, m, "Export")
	_, cmd := m.Update(tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft})
	if cmd == nil {
		t.Fatal("Export click returned nil command")
	}
	request, ok := cmd().(ExportRequestedMsg)
	if !ok || request.Path != "out.json" {
		t.Fatalf("Export click request = %#v", request)
	}
}

func TestOptionSelectorsKeepClosedRowAlignmentAndOpenWithRightArrow(t *testing.T) {
	project := core.Project{Name: "Demo", ProjectID: "demo"}
	tests := []struct {
		name  string
		row   int
		label string
		value string
	}{
		{name: "strategy", row: optionStrategy, label: "Strategy", value: "Merge (keep current conflicts)"},
		{name: "conditions", row: optionConditions, label: "Conditions", value: "Keep all conditions (5 kept)"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			m := importOptionsTestModel(project)
			m.summary = core.ProjectImportSummary{Conditions: 5}
			m.optionCursor = test.row
			base := ansi.Strip(m.importOptionBaseLines(m.contentWidth())[6+test.row])
			cardX, _ := m.Position()
			closedLabelX := cardX + 1 + viewutil.PopupPaddingLeft + strings.Index(base, test.label)
			closedValueX := cardX + 1 + viewutil.PopupPaddingLeft + strings.Index(base, test.value)

			m, _ = m.Update(key("right"))
			if !m.OptionSelectorOpen() {
				t.Fatal("Right Arrow did not open selector")
			}
			headerX, _ := m.OptionSelectorPosition()
			header := strings.Split(ansi.Strip(m.OptionSelectorHeaderView()), "\n")[1]
			if got := headerX + textCellIndex(header, test.label); got != closedLabelX {
				t.Fatalf("open label x = %d, closed x = %d", got, closedLabelX)
			}
			listX, _ := m.OptionSelectorListPosition()
			for line := range strings.SplitSeq(ansi.Strip(m.OptionSelectorListView()), "\n") {
				if strings.Contains(line, test.value) {
					if got := listX + textCellIndex(line, test.value); got != closedValueX {
						t.Fatalf("open value x = %d, closed x = %d", got, closedValueX)
					}
					return
				}
			}
			t.Fatalf("selector does not render %q", test.value)
		})
	}
}

func TestImportFilePopupFitsItsContents(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(dir+"/short.json", []byte("{}"), 0o600); err != nil {
		t.Fatalf("WriteFile = %v", err)
	}
	m := importFileTestModel(core.Project{Name: "Demo", ProjectID: "demo"})
	m.picker.CurrentDirectory = dir
	cmd := m.picker.Init()
	if cmd == nil {
		t.Fatal("file picker init command is nil")
	}
	m.picker, _ = m.picker.Update(cmd())

	view := m.View()
	if !strings.Contains(ansi.Strip(view), "short.json") {
		t.Fatalf("file picker does not render test file:\n%s", ansi.Strip(view))
	}
	if got := lipgloss.Width(view); got >= 60 {
		t.Fatalf("content-fit file popup width = %d, want less than 60", got)
	}
	for line := range strings.SplitSeq(ansi.Strip(view), "\n") {
		if strings.Contains(line, "short.json") && !strings.HasSuffix(line, " │") {
			t.Fatalf("file-picker content lacks right padding: %q", line)
		}
	}
}

func TestActionButtonsKeepOneCellRightPadding(t *testing.T) {
	models := []Model{
		importFileTestModel(core.Project{ProjectID: "demo"}),
		importOptionsTestModel(core.Project{ProjectID: "demo"}),
		exportSourceTestModel(core.Project{ProjectID: "demo"}),
		exportPathTestModel(core.Project{ProjectID: "demo"}),
		defaultsFormatTestModel(core.Project{ProjectID: "demo"}),
		defaultsPathTestModel(core.Project{ProjectID: "demo"}),
	}
	for _, m := range models {
		label := m.actionButtons().View()
		label = ansi.Strip(label)
		labelLine := strings.Split(label, "\n")[1]
		labelLine = strings.TrimSpace(strings.Trim(labelLine, "│"))
		for line := range strings.SplitSeq(ansi.Strip(m.View()), "\n") {
			if strings.Contains(line, labelLine) && !strings.HasSuffix(line, " │") {
				t.Fatalf("button line lacks right padding: %q", line)
			}
		}
	}
}

func importFileTestModel(project core.Project) Model {
	m, _ := New().SetBounds(0, 0, 100, 30).OpenImport(project)
	return m
}

func importOptionsTestModel(project core.Project) Model {
	m := importFileTestModel(project)
	m.phase = phaseImportOptions
	return m
}

func exportSourceTestModel(project core.Project) Model {
	m, _ := New().SetBounds(0, 0, 100, 30).OpenExport(project, true)
	return m
}

func exportPathTestModel(project core.Project) Model {
	m, _ := New().SetBounds(0, 0, 100, 30).OpenExport(project, false)
	return m
}

func defaultsFormatTestModel(project core.Project) Model {
	m, _ := New().SetBounds(0, 0, 100, 30).OpenDefaults(project)
	return m
}

func defaultsPathTestModel(project core.Project) Model {
	m, _ := New().SetBounds(0, 0, 100, 30).OpenDefaultsPath(project, "defaults.json", firebase.DefaultsFormatJSON)
	return m
}

func projectIOTextPoint(t *testing.T, m Model, label string) (int, int) {
	t.Helper()
	x, y := m.Position()
	foundX, foundY, found := 0, 0, false
	for row, line := range strings.Split(ansi.Strip(m.View()), "\n") {
		if before, _, match := strings.Cut(line, label); match {
			foundX, foundY = x+lipgloss.Width(before), y+row
			found = true
		}
	}
	if found {
		return foundX, foundY
	}
	t.Fatalf("project I/O popup does not render %q", label)
	return 0, 0
}

func textCellIndex(line, text string) int {
	before, _, found := strings.Cut(line, text)
	if !found {
		return -1
	}
	return lipgloss.Width(before)
}

func key(value string) tea.KeyPressMsg { return tea.KeyPressMsg(tea.Key{Text: value}) }
