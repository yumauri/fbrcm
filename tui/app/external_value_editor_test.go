package app

import (
	"os"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	coreconfig "github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/panels"
)

func TestShouldUseExternalValueEditorGuardsTextareaLimit(t *testing.T) {
	if shouldUseExternalValueEditor(strings.Repeat("line\n", textareaHardLineLimit)) != true {
		t.Fatal("value beyond the textarea hard line limit should use the external editor")
	}
	if shouldUseExternalValueEditor("small value") {
		t.Fatal("small value should use the built-in editor")
	}
	if !shouldUseExternalValueEditor(strings.Repeat("x", builtInEditorMaxBytes+1)) {
		t.Fatal("large one-line value should use the external editor")
	}
}

func TestOpenExternalValueEditorStagesPrivatePrettyJSON(t *testing.T) {
	m := New(nil)
	cmd := m.openExternalValueEditor(externalValueEditSession{
		kind:         externalJSONValue,
		project:      core.Project{Name: "Demo", ProjectID: "demo"},
		currentValue: `{"enabled":true}`,
		source:       panels.Parameters,
	})
	if cmd == nil || m.externalEdit == nil {
		t.Fatalf("external editor = cmd:%v session:%#v", cmd != nil, m.externalEdit)
	}
	path := m.externalEdit.path
	t.Cleanup(func() { _ = os.Remove(path) })
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(raw), "{\n  \"enabled\": true\n}"; got != want {
		t.Fatalf("staged JSON = %q, want %q", got, want)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != coreconfig.PrivateFileMode {
		t.Fatalf("staged mode = %o, want %o", info.Mode().Perm(), coreconfig.PrivateFileMode)
	}
}

func TestExternalJSONEditorRejectsInvalidAndPreservesRecoveryFile(t *testing.T) {
	path := writeExternalValueTestFile(t, "{broken")
	m := New(nil)
	m.externalEdit = &externalValueEditSession{
		kind:    externalJSONValue,
		project: core.Project{Name: "Demo", ProjectID: "demo"},
		path:    path,
	}

	next, cmd, handled := m.updateExternalValueEditorFinished(externalValueEditorFinishedMsg{path: path})
	if !handled || cmd != nil || !next.dialog.IsOpen() || next.externalEdit == nil {
		t.Fatalf("invalid completion = handled:%v cmd:%v dialog:%v session:%#v", handled, cmd != nil, next.dialog.IsOpen(), next.externalEdit)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("recovery file was not preserved: %v", err)
	}
	next.dialog = next.dialog.SetBounds(0, 0, 100, 30)
	if view := next.dialog.View(); !strings.Contains(view, "Invalid JSON") || !strings.Contains(view, "The staged value is preserved at:") || !strings.Contains(view, "Reopen") {
		t.Fatalf("invalid JSON dialog = %q", view)
	}
}

func TestExternalStringEditorAppliesDetailsValueAndRemovesStage(t *testing.T) {
	path := writeExternalValueTestFile(t, "after")
	m := viewTestModel(100, 32, panels.Details)
	m.details = m.details.SetData(&messages.ParameterViewData{
		Project: core.Project{Name: "Demo", ProjectID: "demo"},
		Parameter: core.ParametersEntry{Key: "message", Values: []core.ParametersValue{{
			Label: "default", RawValue: "before", Value: "before", ValueType: "STRING", Plain: true,
		}}},
		SelectedValueIdx: 0,
	})
	m.detailsVisible = true
	m.externalEdit = &externalValueEditSession{
		kind:         externalStringValue,
		project:      core.Project{Name: "Demo", ProjectID: "demo"},
		currentValue: "before",
		source:       panels.Details,
		path:         path,
	}

	next, cmd, handled := m.updateExternalValueEditorFinished(externalValueEditorFinishedMsg{path: path})
	edit, changed := next.details.Edit()
	if !handled || cmd != nil || next.externalEdit != nil || !changed || len(edit.ValueEdits) != 1 || edit.ValueEdits[0].NextValue != "after" {
		t.Fatalf("completion = handled:%v cmd:%v session:%#v changed:%v edit:%+v", handled, cmd != nil, next.externalEdit, changed, edit)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("successful staged file still exists: %v", err)
	}
}

func TestLargeJSONEditBypassesBuiltInTextarea(t *testing.T) {
	largeJSON := "[" + strings.Repeat("{\"value\":1},", builtInEditorMaxLines) + "{}]"
	m := viewTestModel(100, 32, panels.Details)
	m.details = m.details.SetData(&messages.ParameterViewData{
		Project: core.Project{Name: "Demo", ProjectID: "demo"},
		Parameter: core.ParametersEntry{Key: "payload", Values: []core.ParametersValue{{
			Label: "default", RawValue: largeJSON, Value: largeJSON, ValueType: "JSON", Plain: true,
		}}},
		SelectedValueIdx: 0,
	})
	m.detailsVisible = true

	cmd := m.openJSONInput()
	if cmd == nil || m.externalEdit == nil || m.jsonInput.IsOpen() {
		t.Fatalf("large JSON editor = cmd:%v external:%#v builtIn:%v", cmd != nil, m.externalEdit, m.jsonInput.IsOpen())
	}
	_ = os.Remove(m.externalEdit.path)
}

func TestDetailsExternalEditKeyStagesTextValue(t *testing.T) {
	m := viewTestModel(100, 32, panels.Details)
	m.details = m.details.SetData(&messages.ParameterViewData{
		Project: core.Project{Name: "Demo", ProjectID: "demo"},
		Parameter: core.ParametersEntry{Key: "message", Values: []core.ParametersValue{{
			Label: "default", RawValue: "hello", Value: "hello", ValueType: "STRING", Plain: true,
		}}},
		SelectedValueIdx: 0,
	})
	m.detailsVisible = true

	next, cmd, handled := m.updateKeyMessage(tea.KeyPressMsg(tea.Key{Code: 'E', Text: "E"}))
	if !handled || cmd == nil || next.externalEdit == nil {
		t.Fatalf("E = handled:%v cmd:%v session:%#v", handled, cmd != nil, next.externalEdit)
	}
	_ = os.Remove(next.externalEdit.path)
}

func writeExternalValueTestFile(t *testing.T, value string) string {
	t.Helper()
	file, err := os.CreateTemp(t.TempDir(), "value-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.WriteString(value); err != nil {
		_ = file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	return file.Name()
}
