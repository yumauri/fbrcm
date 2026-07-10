package parameters

import (
	"testing"

	"github.com/yumauri/fbrcm/core"
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

const parametersViewSnapshot = `╭─ ²Parameters ────────────────────────────────────────────╮
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
