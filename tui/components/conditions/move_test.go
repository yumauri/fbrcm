package conditions

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/tui/styles"
)

func TestMoveConditionIsReversibleAndConfinedToProject(t *testing.T) {
	m := moveTestModel()

	if !m.StartMove() {
		t.Fatal("StartMove returned false")
	}
	if m.MoveActiveCondition(-1) {
		t.Fatal("first condition moved above its project row")
	}
	if !m.MoveActiveCondition(1) {
		t.Fatal("condition did not move down")
	}
	if got := conditionNames(m.projects[0].tree.Conditions); strings.Join(got, ",") != "second,first" {
		t.Fatalf("first project order = %v, want [second first]", got)
	}
	if m.projects[0].tree.Conditions[0].Priority != 1 || m.projects[0].tree.Conditions[1].Priority != 2 {
		t.Fatalf("live priorities = %#v, want 1 and 2", m.projects[0].tree.Conditions)
	}
	if m.MoveActiveCondition(1) {
		t.Fatal("condition moved beyond the last condition in its project")
	}
	if got := conditionNames(m.projects[1].tree.Conditions); strings.Join(got, ",") != "third,fourth" {
		t.Fatalf("second project order changed: %v", got)
	}

	m.CancelMove()
	if m.MoveActive() {
		t.Fatal("move mode remained active after cancel")
	}
	if got := conditionNames(m.projects[0].tree.Conditions); strings.Join(got, ",") != "first,second" {
		t.Fatalf("cancelled order = %v, want [first second]", got)
	}
}

func TestFinishMoveReturnsPriorityAndRestoresLoadedTree(t *testing.T) {
	m := moveTestModel()
	m.cursor = 2
	if !m.StartMove() || !m.MoveActiveCondition(-1) {
		t.Fatal("failed to move second condition up")
	}

	priority, changed, ok := m.FinishMove()
	if !ok || !changed || priority != 1 {
		t.Fatalf("FinishMove = (%d, %v, %v), want (1, true, true)", priority, changed, ok)
	}
	if got := conditionNames(m.projects[0].tree.Conditions); strings.Join(got, ",") != "first,second" {
		t.Fatalf("loaded tree after finish = %v, want original order", got)
	}
}

func TestMoveModeRendersArrowWithoutRowSelection(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	m := moveTestModel().SetBounds(0, 0, 60, 10).SetActive(true)
	if !m.StartMove() {
		t.Fatal("StartMove returned false")
	}

	row := m.bodyLines()[1]
	if !strings.Contains(ansi.Strip(row), "") {
		t.Fatalf("moving row does not contain the Powerline arrow: %q", row)
	}
	selectionPrefix := stylePrefix(styles.TreeItemSelectionStyle())
	if selectionPrefix != "" && strings.HasPrefix(row, selectionPrefix) {
		t.Fatalf("moving row still has selection styling: %q", row)
	}
	for _, test := range []struct {
		powerline bool
		glyph     string
	}{{true, ""}, {false, "▶︎"}} {
		prefix := conditionMovePrefix(test.powerline)
		if lipgloss.Width(prefix) != 5 || !strings.Contains(ansi.Strip(prefix), test.glyph) {
			t.Fatalf("conditionMovePrefix(%v) = %q, width %d; want %q at width 5", test.powerline, prefix, lipgloss.Width(prefix), test.glyph)
		}
	}
}

func moveTestModel() Model {
	m := New(nil)
	m.projects = []projectState{
		{project: core.Project{ProjectID: "one"}, tree: &core.ConditionsTree{Conditions: []core.ConditionEntry{{Priority: 1, Name: "first"}, {Priority: 2, Name: "second"}}}},
		{project: core.Project{ProjectID: "two"}, tree: &core.ConditionsTree{Conditions: []core.ConditionEntry{{Priority: 1, Name: "third"}, {Priority: 2, Name: "fourth"}}}},
	}
	m.projectIndex = map[string]int{"one": 0, "two": 1}
	m.syncVisible()
	m.cursor = 1
	return m
}

func conditionNames(conditions []core.ConditionEntry) []string {
	names := make([]string, len(conditions))
	for index := range conditions {
		names[index] = conditions[index].Name
	}
	return names
}
