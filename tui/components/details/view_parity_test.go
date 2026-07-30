package details

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
	corestyles "github.com/yumauri/fbrcm/core/styles"
	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/testutil"
)

// parityViewData builds a representative parameter view used to lock in the
// rendered output of the details panel before the model is split.
func parityViewData() *messages.ParameterViewData {
	return &messages.ParameterViewData{
		Project:    core.Project{Name: "Demo Prod", ProjectID: "demo-prod"},
		GroupKey:   "",
		GroupLabel: "(root)",
		Groups: []messages.ParameterGroupOption{
			{Key: "", Label: "(root)"},
			{Key: "checkout", Label: "checkout"},
		},
		ParameterKeys: []string{"feature_login"},
		Parameter: core.ParametersEntry{
			Key:         "feature_login",
			Description: "Login feature toggle",
			Summary:     "on",
			Values: []core.ParametersValue{
				{Label: "Default", Value: "on", RawValue: "on", ValueType: "STRING"},
				{Label: "ios", Value: "off", RawValue: "off", ValueType: "STRING"},
			},
		},
		SelectedValueIdx: -1,
	}
}

func parityViewDataWithConditionals() *messages.ParameterViewData {
	data := parityViewData()
	data.Parameter = core.ParametersEntry{
		Key:         "feature_login",
		Description: "Login feature toggle",
		Summary:     "3 values",
		Values: []core.ParametersValue{
			{Label: "android", Value: "a", RawValue: "a", ValueType: "STRING", Plain: true},
			{Label: "ios", Value: "b", RawValue: "b", ValueType: "STRING", Plain: true},
			{Label: "default", Value: "c", RawValue: "c", ValueType: "STRING", Plain: true},
		},
	}
	return data
}

func parityTestModel() Model {
	return New().SetBounds(0, 0, 60, 24).SetActive(true).SetData(parityViewData())
}

func TestDetailsViewSnapshot(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	got := testutil.NormalizeViewSnapshot(parityTestModel().View())
	if got != detailsViewSnapshot {
		t.Fatalf("snapshot mismatch\n--- got ---\n%s\n--- want ---\n%s", got, detailsViewSnapshot)
	}
}

func TestInAppDefaultUsesEmptyValueStyle(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	const label = "(in-app default)"

	got := parityTestModel().renderValueLines(core.ParametersValue{
		Value: label, ValueType: "BOOLEAN", UseInAppDefault: true,
	}, 40)
	want := corestyles.EmptyValueStyle().Render(label)
	if len(got) != 1 || got[0] != want {
		t.Fatalf("renderValueLines = %q, want shared empty-value style %q", got, want)
	}
}

func TestDetailsViewEmptyWithoutBounds(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	m := New().SetData(parityViewData())
	if out := m.View(); out != "" {
		t.Fatalf("details view without bounds = %q, want empty", out)
	}
}

func TestValueRowsSelectThenSubmitOnDoubleClick(t *testing.T) {
	m := New().SetBounds(0, 0, 60, 24).SetActive(true).SetData(parityViewDataWithConditionals())
	click := tea.MouseClickMsg{
		X:      2,
		Y:      m.y + 1 + m.valueConditionLine(1) - m.viewport.YOffset(),
		Button: tea.MouseLeft,
	}
	var cmd tea.Cmd
	m, cmd = m.Update(click)
	if cmd != nil || !m.ValueSelected() || m.selectedValue != 1 {
		t.Fatalf("single click = selected:%v index:%d cmd:%v", m.ValueSelected(), m.selectedValue, cmd)
	}
	m, cmd = m.Update(click)
	if cmd == nil {
		t.Fatal("double click did not submit the selected value")
	}
	if _, ok := cmd().(messages.DetailsSelectionSubmitRequestedMsg); !ok {
		t.Fatalf("double-click message = %T, want DetailsSelectionSubmitRequestedMsg", cmd())
	}
}

// TestCurrentConditionalValueAnchorFirstConditional guards against treating the
// first conditional value (index 0 in Values) as a whole-parameter delete target.
func TestCurrentConditionalValueAnchorFirstConditional(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	m := New().SetBounds(0, 0, 60, 24).SetActive(true).SetData(parityViewDataWithConditionals())

	tests := []struct {
		valueIdx  int
		wantOK    bool
		wantLabel string
	}{
		{valueIdx: 0, wantOK: true, wantLabel: "android"},
		{valueIdx: 1, wantOK: true, wantLabel: "ios"},
		{valueIdx: 2, wantOK: false},
	}

	for _, tt := range tests {
		m.selectedValue = tt.valueIdx
		m.activeField = fieldNone
		anchor, ok := m.CurrentConditionalValueAnchor()
		if ok != tt.wantOK {
			t.Fatalf("valueIdx %d: ok = %v, want %v", tt.valueIdx, ok, tt.wantOK)
		}
		if tt.wantOK && anchor.ValueLabel != tt.wantLabel {
			t.Fatalf("valueIdx %d: label = %q, want %q", tt.valueIdx, anchor.ValueLabel, tt.wantLabel)
		}
	}
}

func TestAddConditionalValuePreservesPriorityAndProducesEdit(t *testing.T) {
	data := parityViewDataWithConditionals()
	data.Conditions = []core.ParametersCondition{
		{Name: "staff", Color: "GREEN"},
		{Name: "android", Color: "BLUE"},
		{Name: "ios", Color: "INDIGO"},
	}
	m := New().SetBounds(0, 0, 60, 24).SetActive(true).SetData(data)

	next, added := m.AddConditionalValue("staff")
	if !added {
		t.Fatal("AddConditionalValue returned false")
	}
	got := next.data.Parameter.Values
	if len(got) != 4 || got[0].Label != "staff" || got[1].Label != "android" || got[2].Label != "ios" || got[3].Label != "default" {
		t.Fatalf("value order = %+v, want staff, android, ios, default", got)
	}
	if got[0].RawValue != "" || got[0].ValueType != "STRING" || got[0].Color != "GREEN" || !got[0].Plain {
		t.Fatalf("new value = %+v, want empty plain STRING/GREEN", got[0])
	}
	edit, ok := next.Edit()
	if !ok || len(edit.ValueEdits) != 1 || edit.ValueEdits[0].Label != "staff" || edit.ValueEdits[0].NextValue != "" {
		t.Fatalf("edit = %+v, ok=%v; want empty staff assignment", edit, ok)
	}

	next = next.RemoveAddedConditionalValue("staff")
	if len(next.data.Parameter.Values) != 3 || next.Dirty() {
		t.Fatalf("cancelled values = %+v, dirty=%v; want original clean form", next.data.Parameter.Values, next.Dirty())
	}
}

func TestAddConditionalValueParticipatesInArrowNavigation(t *testing.T) {
	data := parityViewData()
	data.Conditions = []core.ParametersCondition{{Name: "staff", Color: "GREEN"}}
	m := New().SetBounds(0, 0, 60, 24).SetActive(true).SetData(data)

	m, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyUp}))
	if !m.AddConditionalValueSelected() {
		t.Fatal("Up from no selection did not select Add conditional value")
	}
	m, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyDown}))
	if m.AddConditionalValueSelected() || m.activeField != fieldGroup {
		t.Fatalf("Down after Add selection = add:%v field:%v, want Group", m.AddConditionalValueSelected(), m.activeField)
	}
}

func TestToggleSelectedValueSourceStagesPerValueEdit(t *testing.T) {
	data := parityViewDataWithConditionals()
	data.SelectedValueIdx = 0
	m := New().SetBounds(0, 0, 60, 24).SetActive(true).SetData(data)

	next, toggled := m.ToggleSelectedValueSource()
	value, selected := next.SelectedParameterValue()
	if !toggled || !selected || !value.UseInAppDefault || value.Plain {
		t.Fatalf("in-app-default toggle = toggled:%v selected:%v value:%+v", toggled, selected, value)
	}
	edit, changed := next.Edit()
	if !changed || len(edit.ValueEdits) != 1 || edit.ValueEdits[0].Label != "android" || !edit.ValueEdits[0].NextUseInAppDefault {
		t.Fatalf("in-app-default edit = changed:%v edit:%+v", changed, edit)
	}

	next, toggled = next.ToggleSelectedValueSource()
	value, _ = next.SelectedParameterValue()
	if !toggled || !value.Plain || value.UseInAppDefault || value.RawValue != "" {
		t.Fatalf("remote toggle = toggled:%v value:%+v", toggled, value)
	}
	edit, changed = next.Edit()
	if !changed || len(edit.ValueEdits) != 1 || edit.ValueEdits[0].NextUseInAppDefault || edit.ValueEdits[0].NextValue != "" {
		t.Fatalf("restored remote edit = changed:%v edit:%+v", changed, edit)
	}
}

func TestToggleSelectedInAppDefaultRestoresTypedNeutralValue(t *testing.T) {
	data := parityViewDataWithConditionals()
	data.Parameter.Values[0] = core.ParametersValue{
		Label: "android", Value: "(in-app default)", ValueType: "BOOLEAN", UseInAppDefault: true,
	}
	data.SelectedValueIdx = 0
	m := New().SetBounds(0, 0, 60, 24).SetActive(true).SetData(data)
	m.typeValue = "BOOLEAN"

	next, toggled := m.ToggleSelectedValueSource()
	value, _ := next.SelectedParameterValue()
	if !toggled || !value.Plain || value.UseInAppDefault || value.RawValue != "false" {
		t.Fatalf("restored value = toggled:%v value:%+v, want plain false", toggled, value)
	}
}

func TestManagedValueCannotToggleToInAppDefault(t *testing.T) {
	data := parityViewData()
	data.Parameter.Values[0] = core.ParametersValue{
		Label: "default", Value: "⚗ (a/b test)", ValueType: "BOOLEAN",
		Display: rcdisplay.ValueSummary{Kind: rcdisplay.ValueSummaryExperiment, Text: "⚗ (a/b test)"},
	}
	data.SelectedValueIdx = 0
	m := New().SetBounds(0, 0, 60, 24).SetActive(true).SetData(data)

	next, toggled := m.ToggleSelectedValueSource()
	if toggled || next.Dirty() {
		t.Fatalf("managed toggle = toggled:%v dirty:%v, want false/false", toggled, next.Dirty())
	}
	if _, ok := next.CurrentBoolValueAnchor(); ok {
		t.Fatal("managed BOOLEAN value exposed an editor anchor")
	}
}

const detailsViewSnapshot = ` ╭─ ⁵Details ───────────────────────────────────────────────
 │ Project
 │ Demo Prod (demo-prod)
 │
 │ Group
 │ (root)
 │
 │ Name
 │ feature_login
 │
 │ Type
 │ STRING
 │
 │ Description
 │ Login feature toggle
 │
 │ Values
 │   Default
 │     on
 │
 │   ios
 │     off
 │
 ╰──────────────────────────────────────────────────────────`
