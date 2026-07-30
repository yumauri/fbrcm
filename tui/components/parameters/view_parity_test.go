package parameters

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/core"
	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
	corestyles "github.com/yumauri/fbrcm/core/styles"
	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/testutil"
)

// parityTree builds a representative parameters tree used to lock in the
// rendered output of the parameters panel before the model is split.
func parityTree() *core.ParametersTree {
	return &core.ParametersTree{
		Version: "12",
		Groups: []core.ParametersGroup{
			{
				Key:   "__default__",
				Label: "(root)",
				Parameters: []core.ParametersEntry{
					{
						Key:     "feature_login",
						Summary: "on",
						Values: []core.ParametersValue{
							{Label: "Default", Value: "on", RawValue: "on", ValueType: "STRING"},
						},
					},
				},
			},
		},
	}
}

func TestManagedNumberValueHasNoEditorAnchor(t *testing.T) {
	m := parityTestModel()
	project := &m.projects[0]
	project.tree.Groups[0].Parameters[0].Values = []core.ParametersValue{{
		Label: "default", Value: "◐ 10% → 20 / ◑ (no change)", ValueType: "NUMBER",
		Display: rcdisplay.ValueSummary{
			Kind: rcdisplay.ValueSummaryRollout,
			Text: "◐ 10% → 20 / ◑ (no change)",
			Rollout: &rcdisplay.RolloutSummary{
				Percentage: "10%",
				Value:      "20",
			},
		},
		RawValue: string(json.RawMessage(`{"rolloutId":"rollout-1"}`)),
	}}
	m.syncVisible()
	for index, node := range m.visible {
		if node.kind == nodeValue {
			m.cursor = index
			break
		}
	}

	if _, ok := m.CurrentNumberValueAnchor(); ok {
		t.Fatal("managed NUMBER value exposed an editor anchor")
	}
}

func parityTestModel() Model {
	m := New(nil).SetBounds(0, 0, 60, 24).SetActive(true)
	m, _ = m.Update(messages.ProjectsSelectionChangedMsg{
		Projects: []core.Project{{Name: "Demo Prod", ProjectID: "demo-prod"}},
	})
	m, _ = m.Update(messages.ParametersLoadedMsg{
		Project: core.Project{Name: "Demo Prod", ProjectID: "demo-prod"},
		Tree:    parityTree(),
		Source:  "cache",
	})
	m.setAllParametersExpanded(true)
	return m
}

func TestParametersViewSnapshot(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	got := testutil.NormalizeViewSnapshot(parityTestModel().View(true))
	if got != parametersViewSnapshot {
		t.Fatalf("snapshot mismatch\n--- got ---\n%s\n--- want ---\n%s", got, parametersViewSnapshot)
	}
}

func TestInAppDefaultUsesEmptyValueStyle(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	const label = "(in-app default)"

	got := parityTestModel().renderParameterValue(core.ParametersValue{
		Value: label, ValueType: "BOOLEAN", UseInAppDefault: true,
	}, false)
	want := corestyles.EmptyValueStyle().Render(label)
	if got != want {
		t.Fatalf("renderParameterValue = %q, want shared empty-value style %q", got, want)
	}
}

func TestParametersViewShowsEmptyGroups(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	m := New(nil).SetBounds(0, 0, 60, 12).SetActive(true)
	m, _ = m.Update(messages.ProjectsSelectionChangedMsg{
		Projects: []core.Project{{Name: "Demo", ProjectID: "demo"}},
	})
	m, _ = m.Update(messages.ParametersLoadedMsg{
		Project: core.Project{Name: "Demo", ProjectID: "demo"},
		Tree: &core.ParametersTree{Groups: []core.ParametersGroup{
			{Key: "empty", Label: "empty"},
			{Key: "ROKU", Label: "ROKU"},
		}},
		Source: "cache",
	})

	view := testutil.NormalizeViewSnapshot(m.View(true))
	for _, group := range []string{"empty", "ROKU"} {
		if !strings.Contains(view, group) {
			t.Fatalf("view does not show empty group %q:\n%s", group, view)
		}
	}
}

// TestCurrentConditionalValueAnchorFirstConditional guards against a regression
// where pressing delete on the first conditional value (valueIdx 0) was treated
// as a whole-parameter delete. Conditional values are listed first and the
// default value last, so valueIdx 0 is the first conditional, not the default.
func TestCurrentConditionalValueAnchorFirstConditional(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	tree := &core.ParametersTree{
		Version: "1",
		Groups: []core.ParametersGroup{
			{
				Key:   "__default__",
				Label: "(root)",
				Parameters: []core.ParametersEntry{
					{
						Key:     "feature_login",
						Summary: "3 values",
						Values: []core.ParametersValue{
							{Label: "android", Value: "a", RawValue: "a", ValueType: "STRING", Plain: true},
							{Label: "ios", Value: "b", RawValue: "b", ValueType: "STRING", Plain: true},
							{Label: "default", Value: "c", RawValue: "c", ValueType: "STRING", Plain: true},
						},
					},
				},
			},
		},
	}

	m := New(nil).SetBounds(0, 0, 80, 24).SetActive(true)
	m, _ = m.Update(messages.ProjectsSelectionChangedMsg{
		Projects: []core.Project{{Name: "Demo", ProjectID: "demo"}},
	})
	m, _ = m.Update(messages.ParametersLoadedMsg{
		Project: core.Project{Name: "Demo", ProjectID: "demo"},
		Tree:    tree,
		Source:  "cache",
	})
	m.setAllParametersExpanded(true)

	seen := map[int]bool{}
	for idx, node := range m.visible {
		if node.kind != nodeValue || node.paramKey != "feature_login" {
			continue
		}
		seen[node.valueIdx] = true
		m.cursor = idx
		anchor, ok := m.CurrentConditionalValueAnchor()
		switch node.valueIdx {
		case 0:
			if !ok || anchor.ValueLabel != "android" {
				t.Fatalf("first conditional (valueIdx 0): anchor=%+v ok=%v, want ok with label android", anchor, ok)
			}
		case 1:
			if !ok || anchor.ValueLabel != "ios" {
				t.Fatalf("second conditional (valueIdx 1): anchor=%+v ok=%v, want ok with label ios", anchor, ok)
			}
		case 2:
			if ok {
				t.Fatalf("default value (valueIdx 2): got conditional anchor %+v, want none", anchor)
			}
		}
	}

	for _, idx := range []int{0, 1, 2} {
		if !seen[idx] {
			t.Fatalf("value node with valueIdx %d not found among visible nodes", idx)
		}
	}
}

const parametersViewSnapshot = `╭─ ²Parameters ── \≡ ── ³Conditions ───────────────────────╮
│Demo Prod demo-prod                             v12 staled│
│▾ (root)                                                  │
│  feature_login                                           │
│  ╰ Default ╌╌╌╌╌╌╌ on                                    │
│                                                          │
│                                                          │
│                                                          │
│                                                          │
│                                                          │
│                                                          │
│                                                          │
│                                                          │
│                                                          │
│                                                          │
│                                                          │
│                                                          │
│                                                          │
│                                                          │
│                                                          │
│                                                          │
│                                                          │
│                                                          │
╰──────────────────────────────────────────────────────────╯`
