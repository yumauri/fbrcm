package details

import (
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

func renderedStylePrefix(style lipgloss.Style) string {
	rendered := style.Render("x")
	prefix, _, ok := strings.Cut(rendered, "x")
	if !ok {
		return ""
	}
	return prefix
}
