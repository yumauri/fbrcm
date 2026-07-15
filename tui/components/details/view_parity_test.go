package details

import (
	"testing"

	"github.com/yumauri/fbrcm/core"
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

func TestDetailsViewEmptyWithoutBounds(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	m := New().SetData(parityViewData())
	if out := m.View(); out != "" {
		t.Fatalf("details view without bounds = %q, want empty", out)
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
