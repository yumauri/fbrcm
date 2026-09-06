package parameters

import (
	"encoding/json"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
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
		Label: "default", Value: "◐ 10% → 20 | (no change)", ValueType: "NUMBER",
		Display: rcdisplay.ValueSummary{
			Kind: rcdisplay.ValueSummaryRollout,
			Text: "◐ 10% → 20 | (no change)",
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

func TestParametersRowsReadRuntimePaletteOnEveryView(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	m := parityTestModel()
	before := m.View(true)
	corestyles.ApplyPalette(corestyles.Palette{corestyles.TokenPrimary: "1"})
	t.Cleanup(corestyles.ResetPalette)
	want := parameterStyle.Render("feature_login")
	if strings.Contains(before, want) {
		t.Fatal("pre-theme parameter row unexpectedly uses preview palette")
	}
	if got := m.View(true); !strings.Contains(got, want) {
		t.Fatalf("parameter row does not read the runtime palette on render:\n%s", got)
	}
}

func TestInAppDefaultUsesEmptyValueStyle(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	const label = "(in-app default)"

	got := parityTestModel().renderParameterValue(core.ParametersValue{
		Value: label, ValueType: "BOOLEAN", UseInAppDefault: true,
	}, false, 40)
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

func TestExpandedShortParameterAlignsConditionalValues(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	for _, otherKey := range []string{"other", "long_parameter_name"} {
		t.Run(otherKey, func(t *testing.T) {
			project := core.Project{Name: "Demo", ProjectID: "demo"}
			tree := &core.ParametersTree{Groups: []core.ParametersGroup{{
				Key:   "__default__",
				Label: "(root)",
				Parameters: []core.ParametersEntry{{
					Key: "test",
					Values: []core.ParametersValue{
						{Label: "test condition", Value: "conditional-result", RawValue: "conditional-result", ValueType: "STRING", Plain: true},
						{Label: "default", Value: "default-result", RawValue: "default-result", ValueType: "STRING", Plain: true},
					},
				}, {
					Key: otherKey,
					Values: []core.ParametersValue{
						{Label: "default", Value: "collapsed-result", RawValue: "collapsed-result", ValueType: "STRING", Plain: true},
					},
				}, {
					Key: "solo",
					Values: []core.ParametersValue{
						{Label: "default", Value: "single-expanded-result", RawValue: "single-expanded-result", ValueType: "STRING", Plain: true},
					},
				}},
			}}}
			m := New(nil).SetBounds(0, 0, 60, 12).SetActive(true)
			m, _ = m.Update(messages.ProjectsSelectionChangedMsg{Projects: []core.Project{project}})
			m, _ = m.Update(messages.ParametersLoadedMsg{Project: project, Tree: tree, Source: "cache"})
			m.setAllGroupsExpanded(true)
			m.paramExpanded[m.paramKey(project.ProjectID, "__default__", "test")] = true
			m.paramExpanded[m.paramKey(project.ProjectID, "__default__", "solo")] = true
			m.syncVisible()

			labelColumns := make([]int, 0, 2)
			valueColumns := make([]int, 0, 3)
			collapsedValueColumn := -1
			for _, node := range m.visible {
				if node.kind == nodeParameter && node.paramKey == otherKey {
					line := ansi.Strip(m.renderParameterNode(node, false))
					collapsedValueColumn = renderedTextColumn(t, line, "collapsed-result")
				}
				if node.kind != nodeValue || (node.paramKey != "test" && node.paramKey != "solo") {
					continue
				}
				line := ansi.Strip(m.renderValueNode(node, false))
				param := m.parameterByKey(node.projectID, node.groupKey, node.paramKey)
				value := param.Values[node.valueIdx]
				label := rcdisplay.FormatConditionLabel(value.Label)
				if node.paramKey == "test" {
					labelColumns = append(labelColumns, renderedTextColumn(t, line, label))
				}
				valueColumn := renderedTextColumn(t, line, value.Value)
				valueColumns = append(valueColumns, valueColumn)
				if got, want := m.valueNodeValueX(node, param), valueColumn; got != want {
					t.Fatalf("value anchor column = %d, rendered value column = %d\n%s", got, want, line)
				}
			}

			if len(labelColumns) != 2 || labelColumns[0] != labelColumns[1] {
				t.Fatalf("condition label columns = %v, want two aligned labels", labelColumns)
			}
			if len(valueColumns) != 3 || valueColumns[0] != valueColumns[1] || valueColumns[0] != valueColumns[2] {
				t.Fatalf("expanded parameter value columns = %v, want three aligned values", valueColumns)
			}
			if collapsedValueColumn < 0 || collapsedValueColumn != valueColumns[0] {
				t.Fatalf("collapsed value column = %d, expanded value columns = %v, want all values aligned", collapsedValueColumn, valueColumns)
			}
		})
	}
}

func renderedTextColumn(t *testing.T, line, text string) int {
	t.Helper()
	before, _, found := strings.Cut(line, text)
	if !found {
		t.Fatalf("rendered line does not contain %q:\n%s", text, line)
	}
	return lipgloss.Width(before)
}

func TestCollapsedParameterIconsAlignNextToValues(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	project := core.Project{Name: "Demo", ProjectID: "demo"}
	tree := &core.ParametersTree{Groups: []core.ParametersGroup{{
		Key:   "__default__",
		Label: "(root)",
		Parameters: []core.ParametersEntry{
			{Key: "a", Values: []core.ParametersValue{{Label: "default", Value: "single-value", ValueType: "STRING", Plain: true}}},
			{Key: "b", Values: []core.ParametersValue{
				{Label: "condition", Value: "conditional-value", ValueType: "STRING", Plain: true},
				{Label: "default", Value: "default-value", ValueType: "STRING", Plain: true},
			}},
			{Key: "long_parameter_name", Values: []core.ParametersValue{{Label: "default", Value: "driver-value", ValueType: "STRING", Plain: true}}},
		},
	}}}
	m := New(nil).SetBounds(0, 0, 80, 12).SetActive(true)
	m, _ = m.Update(messages.ProjectsSelectionChangedMsg{Projects: []core.Project{project}})
	m, _ = m.Update(messages.ParametersLoadedMsg{Project: project, Tree: tree, Source: "cache"})
	m.setAllGroupsExpanded(true)

	type expectedRow struct {
		icon  string
		value string
	}
	want := map[string]expectedRow{
		"a": {icon: "╌", value: "single-value"},
		"b": {icon: "⌥", value: "conditional-value"},
	}
	valueColumn := -1
	iconColumn := -1
	for _, node := range m.visible {
		expected, ok := want[node.paramKey]
		if node.kind != nodeParameter || !ok {
			continue
		}
		line := ansi.Strip(m.renderParameterNode(node, false))
		gotValueColumn := renderedTextColumn(t, line, expected.value)
		gotIconColumn := renderedTextColumn(t, line, expected.icon)
		if gotIconColumn != gotValueColumn-2 {
			t.Fatalf("%s icon column = %d, value column = %d, want icon immediately before value\n%s", node.paramKey, gotIconColumn, gotValueColumn, line)
		}
		if valueColumn >= 0 && gotValueColumn != valueColumn {
			t.Fatalf("%s value column = %d, want shared column %d", node.paramKey, gotValueColumn, valueColumn)
		}
		if iconColumn >= 0 && gotIconColumn != iconColumn {
			t.Fatalf("%s icon column = %d, want shared column %d", node.paramKey, gotIconColumn, iconColumn)
		}
		valueColumn = gotValueColumn
		iconColumn = gotIconColumn
	}
	if valueColumn < 0 || iconColumn < 0 {
		t.Fatal("collapsed parameter rows were not rendered")
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
│Demo Prod demo-prod                              v12 stale│
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
