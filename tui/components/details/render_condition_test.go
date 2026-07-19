package details

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/core"
	corestyles "github.com/yumauri/fbrcm/core/styles"
	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/testutil"
)

func conditionDetailsTestModel() Model {
	return New().SetBounds(0, 0, 60, 40).SetActive(true).SetConditionData(&messages.ConditionViewData{
		Project: core.Project{Name: "Demo", ProjectID: "demo"},
		Condition: core.ConditionEntry{
			Priority:   1,
			Name:       "staff",
			TagColor:   "GREEN",
			Expression: "true",
			Usages: []core.ConditionUsage{
				{GroupLabel: "checkout", ParameterKey: "enabled", ValueType: "BOOLEAN", Value: "true", RawValue: "true", Plain: true},
				{GroupLabel: "(root)", ParameterKey: "payload", ValueType: "JSON", Value: `{"enabled":true,"limit":2}`, RawValue: `{"enabled":true,"limit":2}`, Plain: true},
			},
		},
	})
}

func TestConditionDetailsRendersStyledUsageStructureAndPrettyJSON(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	got := testutil.NormalizeViewSnapshot(ansi.Strip(conditionDetailsTestModel().View()))

	for _, want := range []string{
		"Color\n │ ● GREEN",
		"Used by 2 parameters",
		"checkout / enabled",
		"│   true",
		"(root) / payload",
		"│   {",
		`"enabled": true,`,
		`"limit": 2`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("condition Details missing %q:\n%s", want, got)
		}
	}
	for _, unwanted := range []string{"BOOLEAN ·", "JSON ·"} {
		if strings.Contains(got, unwanted) {
			t.Fatalf("condition Details still renders parameter type %q:\n%s", unwanted, got)
		}
	}
}

func TestConditionDetailsUsesGroupAndTypedValueStyles(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	got := strings.Join(conditionDetailsTestModel().renderContentLines(), "\n")

	for name, style := range map[string]lipgloss.Style{
		"group": groupValueStyle,
		"value": corestyles.ValueTextStyle("true", "BOOLEAN"),
	} {
		prefix := renderedStylePrefix(style)
		if prefix == "" || !strings.Contains(got, prefix) {
			t.Fatalf("condition Details does not use %s style:\n%s", name, got)
		}
	}
}

func TestConditionUsagesParticipateInNavigationAndStageValueEdits(t *testing.T) {
	m := conditionDetailsTestModel()
	m, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyUp}))
	if !m.UsageSelected() {
		t.Fatal("Up from no selection did not select the last condition usage")
	}
	usage, _ := m.SelectedUsage()
	if usage.ParameterKey != "payload" {
		t.Fatalf("selected usage = %q, want payload", usage.ParameterKey)
	}
	if _, ok := m.CurrentJSONValueAnchor(); !ok {
		t.Fatal("selected JSON usage did not expose the JSON editor anchor")
	}
	m = m.SetSelectedValue(`{"enabled":false}`)
	edit, ok := m.ConditionEdit()
	if !ok || len(edit.ValueEdits) != 1 || edit.ValueEdits[0].ParameterKey != "payload" || edit.ValueEdits[0].NextValue != `{"enabled":false}` {
		t.Fatalf("ConditionEdit = %+v, ok=%v; want payload JSON value edit", edit, ok)
	}
}

func TestConditionUsageSelectionStylesOnlyParameterName(t *testing.T) {
	m := conditionDetailsTestModel()
	m.selectedUsage = 0
	lines := strings.Join(m.renderContentLines(), "\n")
	if !strings.Contains(lines, groupValueStyle.Render("checkout")+labelStyle.Render(" / ")+selectedValueStyle.Render("enabled")) {
		t.Fatalf("selected usage does not preserve group/separator styles:\n%s", lines)
	}
	if strings.Contains(lines, selectedValueStyle.Render("checkout")) {
		t.Fatalf("selection style covers group name:\n%s", lines)
	}
}

func TestConditionDetailsFieldsStageOneAtomicEdit(t *testing.T) {
	m := conditionDetailsTestModel()
	m.conditionData.ConditionNames = []string{"staff", "beta"}
	m.priorityInput.SetValue("2")
	m.nameInput.SetValue("employees")
	m.conditionColor = "PURPLE"
	m = m.SetConditionExpression("app.version > '2'")

	edit, ok := m.ConditionEdit()
	if !ok {
		t.Fatal("ConditionEdit reported no changes")
	}
	if edit.Name != "staff" || edit.NextName != "employees" || edit.NextPriority != 2 || edit.NextTagColor != "PURPLE" || edit.NextExpression != "app.version > '2'" {
		t.Fatalf("ConditionEdit = %#v", edit)
	}
	if m.Invalid() {
		t.Fatalf("valid condition form is invalid: %v", m.InvalidReasons())
	}
}

func TestConditionDetailsColorPickerKeepsConditionColors(t *testing.T) {
	m := conditionDetailsTestModel().ActivateConditionColor()
	if !m.DropdownOpen() {
		t.Fatal("condition color dropdown is closed")
	}
	view := m.DropdownListView()
	for _, want := range []string{"● GREEN", "● DEEP ORANGE"} {
		if !strings.Contains(ansi.Strip(view), want) {
			t.Fatalf("condition color dropdown missing %q:\n%s", want, view)
		}
	}
	greenPrefix := renderedStylePrefix(m.conditionStyle("GREEN").Bold(true))
	if greenPrefix == "" || !strings.Contains(view, greenPrefix) {
		t.Fatalf("selected GREEN option is not bold in its own color:\n%s", view)
	}
}

func TestConditionPriorityFieldRejectsNonNumericInput(t *testing.T) {
	m, _ := conditionDetailsTestModel().ActivateConditionPriority()
	before := m.priorityInput.Value()
	m, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: 'x', Text: "x"}))
	if got := m.priorityInput.Value(); got != before {
		t.Fatalf("priority after non-numeric input = %q, want %q", got, before)
	}
}

func TestConditionDetailsFieldsActivateFromMouse(t *testing.T) {
	tests := []struct {
		name         string
		field        fieldID
		wantDropdown bool
	}{
		{name: "priority", field: fieldConditionPriority},
		{name: "name", field: fieldName},
		{name: "color", field: fieldConditionColor, wantDropdown: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			m := conditionDetailsTestModel()
			m, _ = m.Update(tea.MouseClickMsg{X: 3, Y: m.fieldScreenY(test.field), Button: tea.MouseLeft})
			if m.activeField != test.field || m.DropdownOpen() != test.wantDropdown {
				t.Fatalf("mouse field = %v, dropdown = %v; want %v, %v", m.activeField, m.DropdownOpen(), test.field, test.wantDropdown)
			}
		})
	}
}

func TestConditionUsageNavigationScrollsWholeValueIntoView(t *testing.T) {
	usages := make([]core.ConditionUsage, 8)
	for index := range usages {
		usages[index] = core.ConditionUsage{
			GroupLabel: "checkout", ParameterKey: fmt.Sprintf("flag_%d", index),
			ValueType: "BOOLEAN", Value: "true", RawValue: "true", Plain: true,
		}
	}
	m := conditionUsageScrollTestModel(12, usages)
	m, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyUp}))

	last := len(usages) - 1
	if !m.UsageSelected() || m.selectedUsage != last {
		t.Fatalf("selected usage = %d, want %d", m.selectedUsage, last)
	}
	start := m.conditionUsageParameterLine(last)
	end := m.conditionUsageEndLine(last)
	top := m.viewport.YOffset()
	bottom := top + m.viewport.Height() - 1
	if start < top || end > bottom {
		t.Fatalf("selected usage block %d..%d outside viewport %d..%d", start, end, top, bottom)
	}
}

func TestOversizedConditionUsagePinsParameterNameToTop(t *testing.T) {
	usage := core.ConditionUsage{
		GroupLabel: "checkout", ParameterKey: "payload", ValueType: "JSON", Plain: true,
		RawValue: `{"one":1,"two":2,"three":3,"four":4,"five":5,"six":6}`,
		Value:    `{"one":1,"two":2,"three":3,"four":4,"five":5,"six":6}`,
	}
	m := conditionUsageScrollTestModel(8, []core.ConditionUsage{usage})
	m, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyUp}))

	start := m.conditionUsageParameterLine(0)
	end := m.conditionUsageEndLine(0)
	if end-start+1 <= m.viewport.Height() {
		t.Fatalf("test usage block height = %d, viewport = %d; want oversized block", end-start+1, m.viewport.Height())
	}
	if got := m.viewport.YOffset(); got != start {
		t.Fatalf("viewport offset = %d, want parameter line %d", got, start)
	}
}

func conditionUsageScrollTestModel(height int, usages []core.ConditionUsage) Model {
	return New().SetBounds(0, 0, 48, height).SetActive(true).SetConditionData(&messages.ConditionViewData{
		Project: core.Project{Name: "Demo", ProjectID: "demo"},
		Condition: core.ConditionEntry{
			Priority: 1, Name: "staff", Expression: "true", Usages: usages,
		},
	})
}

func renderedStylePrefix(style lipgloss.Style) string {
	rendered := style.Render("x")
	prefix, _, ok := strings.Cut(rendered, "x")
	if !ok {
		return ""
	}
	return prefix
}
