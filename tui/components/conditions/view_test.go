package conditions

import (
	"strings"
	"testing"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/tui/styles"
)

func TestViewShowsConditionsInPriorityOrder(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	m := New(nil).SetBounds(0, 0, 72, 10).SetActive(true)
	m.projects = []projectState{{
		project: core.Project{Name: "Demo", ProjectID: "demo"},
		source:  "cache",
		tree: &core.ConditionsTree{Version: "7", Conditions: []core.ConditionEntry{
			{Priority: 1, Name: "zebra", Expression: "app.version > '2'", Usages: []core.ConditionUsage{{}}},
			{Priority: 2, Name: "alpha", Expression: "user in staff"},
		}},
	}}
	m.projectIndex = map[string]int{"demo": 0}
	m.syncVisible()
	got := m.ViewWithBorder(true, true)
	if !strings.Contains(got, "³Conditions") || !strings.Contains(got, "zebra") || !strings.Contains(got, "alpha") {
		t.Fatalf("conditions view missing content:\n%s", got)
	}
	if strings.Index(got, "zebra") > strings.Index(got, "alpha") {
		t.Fatalf("conditions are not in priority order:\n%s", got)
	}
}

func TestProjectRowMatchesWorkspaceStyleAndRightAlignsMetadata(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	m := New(nil)
	project := projectState{
		project: core.Project{Name: "Demo", ProjectID: "demo"},
		source:  "cache",
		tree: &core.ConditionsTree{
			Version:  "7",
			CachedAt: time.Date(2020, time.January, 2, 3, 4, 5, 0, time.Local),
		},
	}
	const width = 48
	got := ansi.Strip(m.renderProjectRow(project, false, width))
	if lipgloss.Width(got) != width {
		t.Fatalf("project row width = %d, want %d: %q", lipgloss.Width(got), width, got)
	}
	if !strings.HasPrefix(got, "Demo demo") || !strings.HasSuffix(got, "v7 staled 2020-01-02 03:04:05") {
		t.Fatalf("project row does not match workspace layout: %q", got)
	}
	selected := m.renderProjectRow(project, true, width)
	if got := lipgloss.Width(selected); got != width {
		t.Fatalf("selected project row width = %d, want %d: %q", got, width, selected)
	}
	selectionPrefix := stylePrefix(styles.TreeProjectSelectionStyle())
	if selectionPrefix == "" || !strings.HasPrefix(selected, selectionPrefix) {
		t.Fatalf("selected project does not style the full row: %q", selected)
	}
}

func TestConditionRowsFillWidthAndColorName(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	m := New(nil)
	project := projectState{
		project: core.Project{Name: "Demo", ProjectID: "demo"},
		tree: &core.ConditionsTree{Conditions: []core.ConditionEntry{{
			Priority: 1, Name: "staff", TagColor: "GREEN", Expression: "true",
		}}},
	}
	m.projects = []projectState{project}
	m.projectIndex = map[string]int{"demo": 0}
	node := visibleNode{kind: nodeCondition, projectID: "demo", conditionIndex: 0}
	const width = 48

	unselected := m.renderNode(node, false, width)
	selected := m.renderNode(node, true, width)
	for name, row := range map[string]string{"unselected": unselected, "selected": selected} {
		if got := lipgloss.Width(row); got != width {
			t.Fatalf("%s condition row width = %d, want %d: %q", name, got, width, row)
		}
	}
	coloredName := lipgloss.NewStyle().Foreground(styles.ConditionLipglossColor("GREEN")).Render("staff")
	if !strings.Contains(unselected, coloredName) {
		t.Fatalf("condition name is not rendered in its tag color: %q", unselected)
	}
	selectionPrefix := stylePrefix(styles.TreeItemSelectionStyle())
	if selectionPrefix == "" || !strings.HasPrefix(selected, selectionPrefix) {
		t.Fatalf("selected condition does not style the full row: %q", selected)
	}
}

func TestConditionRowsReserveThreeCharactersForPriority(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	m := New(nil)
	m.projects = []projectState{{
		project: core.Project{ProjectID: "demo"},
		tree: &core.ConditionsTree{Conditions: []core.ConditionEntry{
			{Priority: 99, Name: "before"},
			{Priority: 100, Name: "after"},
		}},
	}}
	m.projectIndex = map[string]int{"demo": 0}

	before := ansi.Strip(m.renderNode(visibleNode{kind: nodeCondition, projectID: "demo", conditionIndex: 0}, false, 48))
	after := ansi.Strip(m.renderNode(visibleNode{kind: nodeCondition, projectID: "demo", conditionIndex: 1}, false, 48))
	if got, want := strings.Index(before, "before"), strings.Index(after, "after"); got != want {
		t.Fatalf("condition names do not align across priority 99/100: columns %d and %d\n%s\n%s", got, want, before, after)
	}
}

func TestProjectGapIsOneRowAndNonSelectable(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	m := New(nil).SetBounds(0, 0, 60, 12)
	m.projects = []projectState{
		{project: core.Project{Name: "One", ProjectID: "one"}, tree: &core.ConditionsTree{Conditions: []core.ConditionEntry{{Name: "first"}}}},
		{project: core.Project{Name: "Two", ProjectID: "two"}, tree: &core.ConditionsTree{Conditions: []core.ConditionEntry{{Name: "second"}}}},
	}
	m.projectIndex = map[string]int{"one": 0, "two": 1}
	m.syncVisible()

	wantKinds := []nodeKind{nodeProject, nodeCondition, nodeGap, nodeProject, nodeCondition}
	if len(m.visible) != len(wantKinds) {
		t.Fatalf("visible rows = %d, want %d", len(m.visible), len(wantKinds))
	}
	for i, want := range wantKinds {
		if m.visible[i].kind != want {
			t.Fatalf("visible row %d kind = %v, want %v", i, m.visible[i].kind, want)
		}
	}

	m.cursor = 1
	m.moveCursor(1)
	if m.cursor != 3 {
		t.Fatalf("cursor after moving down across gap = %d, want 3", m.cursor)
	}
	m.moveCursor(-1)
	if m.cursor != 1 {
		t.Fatalf("cursor after moving up across gap = %d, want 1", m.cursor)
	}
	if _, ok := m.nodeAtMouse(2, 3); ok {
		t.Fatal("project gap row is mouse-selectable")
	}
}

func TestMarkProjectReloadingStartsConditionsSpinner(t *testing.T) {
	m := New(nil)
	m.projects = []projectState{{project: core.Project{ProjectID: "demo"}}}
	m.projectIndex = map[string]int{"demo": 0}
	m.syncVisible()

	next, tick := m.MarkProjectReloading("demo")
	if tick == nil || !next.projects[0].loading {
		t.Fatalf("MarkProjectReloading loading=%v tick=%v; want true and spinner tick", next.projects[0].loading, tick)
	}
	before := next.spin.View()
	msg := tick()
	next, followup := next.Update(msg)
	if followup == nil {
		t.Fatal("Conditions spinner did not schedule its next animation frame")
	}
	if after := next.spin.View(); after == before {
		t.Fatalf("Conditions spinner frame did not advance: %q", after)
	}
}

func stylePrefix(style lipgloss.Style) string {
	rendered := style.Render("x")
	prefix, _, ok := strings.Cut(rendered, "x")
	if !ok {
		return ""
	}
	return prefix
}
