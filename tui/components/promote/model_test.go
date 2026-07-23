package promote

import (
	"context"
	"fmt"
	"image/color"
	"reflect"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/dictdiff"
	"github.com/yumauri/fbrcm/core/firebase"
	rcdiff "github.com/yumauri/fbrcm/core/rc/diff"
	rcpromote "github.com/yumauri/fbrcm/core/rc/promote"
	"github.com/yumauri/fbrcm/tui/components/jsoninput"
	"github.com/yumauri/fbrcm/tui/styles"
	"github.com/yumauri/fbrcm/tui/testutil"
)

func TestTargetPickerExcludesSourceAndSelectsTarget(t *testing.T) {
	svc, err := core.NewService(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	source := core.Project{Name: "Development", ProjectID: "dev"}
	target := core.Project{Name: "Production", ProjectID: "prod"}
	m := New(svc).Open(source, []core.Project{source, target}).SetBounds(24, 0, 76, 20).SetTargetRow(10)
	view := testutil.NormalizeViewSnapshot(m.TargetView())
	if strings.Contains(view, "Development (dev) · cached only") || !strings.Contains(view, "Production (prod)") {
		t.Fatalf("target picker view:\n%s", view)
	}
	rawView := m.TargetView()
	if !strings.Contains(rawView, styles.PanelMuted.Render(" (prod)")) {
		t.Fatalf("target project ID does not use the Projects-panel muted style:\n%s", view)
	}
	for _, text := range []string{"Promote to…", "Filter", "Type to filter projects"} {
		if !strings.Contains(view, text) {
			t.Fatalf("target picker missing %q:\n%s", text, view)
		}
	}
	lines := strings.Split(view, "\n")
	if len(lines) < 4 || !strings.Contains(lines[1], "Production (prod)") || !strings.Contains(lines[2], "Filter: Type to filter projects") {
		t.Fatalf("filter is not the final list row:\n%s", view)
	}
	if strings.Contains(lines[2], "…") {
		t.Fatalf("filter row contains a synthetic overflow marker: %q", lines[2])
	}
	if x, _ := m.TargetPosition(); x != 22 {
		t.Fatalf("target picker X = %d, want connected seam at 22", x)
	}
	sourceView := testutil.NormalizeViewSnapshot(m.SourceView())
	sourceLines := strings.Split(sourceView, "\n")
	if len(sourceLines) != 4 || strings.HasPrefix(sourceLines[0], "╭") || !strings.Contains(sourceLines[1], " Development") || !strings.Contains(sourceLines[2], "  dev") {
		t.Fatalf("source project overlay:\n%s", sourceView)
	}
	if x, y := m.SourcePosition(); x != 0 || y != 9 {
		t.Fatalf("source overlay position = %d,%d, want 0,9", x, y)
	}
	rawTargetLines := strings.Split(m.TargetView(), "\n")
	wantWidth := lipgloss.Width(rawTargetLines[0])
	for index, line := range rawTargetLines {
		if width := lipgloss.Width(line); width != wantWidth {
			t.Fatalf("target row %d width = %d, want %d:\n%s", index, width, wantWidth, testutil.NormalizeViewSnapshot(m.TargetView()))
		}
	}
	normalizedTop := testutil.NormalizeViewSnapshot(rawTargetLines[0])
	normalizedSelected := testutil.NormalizeViewSnapshot(rawTargetLines[1])
	normalizedFilter := testutil.NormalizeViewSnapshot(rawTargetLines[2])
	normalizedBottom := testutil.NormalizeViewSnapshot(rawTargetLines[3])
	if !strings.HasPrefix(normalizedTop, "──") || !strings.HasPrefix(normalizedSelected, "▸ ") || !strings.HasPrefix(normalizedFilter, "  Filter:") || !strings.HasPrefix(normalizedBottom, "──") {
		t.Fatalf("single-target connected outline is not seamless:\n%s", testutil.NormalizeViewSnapshot(m.TargetView()))
	}
	next, cmd := m.Update(key(tea.KeyEnter))
	if cmd == nil || next.phase != phaseTarget {
		t.Fatalf("target submit = phase:%v cmd:%v", next.phase, cmd != nil)
	}
	msg, ok := cmd().(TargetSelectedMsg)
	if !ok || msg.Source.ProjectID != "dev" || msg.Target.ProjectID != "prod" {
		t.Fatalf("target message = %#v", msg)
	}
}

func TestTargetPickerFiltersAsUserTypes(t *testing.T) {
	source := core.Project{Name: "SignalOps Console", ProjectID: "signalops-console"}
	projects := []core.Project{
		source,
		{Name: "Mercato Mobile", ProjectID: "mercato-mobile-9eac5"},
		{Name: "Northstar Wallet", ProjectID: "northstar-wallet"},
		{Name: "PulseForge Fitness", ProjectID: "pulseforge-fitness-f60f7"},
	}
	m := New(nil).Open(source, projects).SetBounds(24, 0, 76, 20).SetTargetRow(10)
	for _, r := range "WALLET" {
		m, _ = m.Update(keyRune(r))
	}
	view := testutil.NormalizeViewSnapshot(m.TargetView())
	if !strings.Contains(view, "Northstar Wallet (northstar-wallet)") || strings.Contains(view, "Mercato Mobile") || strings.Contains(view, "PulseForge Fitness") {
		t.Fatalf("filtered target picker:\n%s", view)
	}
}

func TestTargetPickerKeepsConnectedOutlineWithoutMatches(t *testing.T) {
	source := core.Project{Name: "Source", ProjectID: "source"}
	m := New(nil).Open(source, []core.Project{
		source,
		{Name: "Target", ProjectID: "target"},
	}).SetBounds(24, 0, 76, 20).SetTargetRow(10)
	for _, r := range "missing" {
		m, _ = m.Update(keyRune(r))
	}
	lines := strings.Split(testutil.NormalizeViewSnapshot(m.TargetView()), "\n")
	if len(lines) != 4 || !strings.HasPrefix(lines[0], "──") || !strings.HasPrefix(lines[1], "  No matching target projects") || !strings.HasPrefix(lines[2], "  Filter:") || !strings.HasPrefix(lines[3], "──") {
		t.Fatalf("empty target picker outline is not seamless:\n%s", strings.Join(lines, "\n"))
	}
	wantWidth := lipgloss.Width(lines[0])
	for index, line := range lines {
		if width := lipgloss.Width(line); width != wantWidth {
			t.Fatalf("empty target row %d width = %d, want %d", index, width, wantWidth)
		}
	}
}

func TestTargetPickerMovesListAroundAnchoredSelection(t *testing.T) {
	source := core.Project{Name: "Source", ProjectID: "source"}
	projects := []core.Project{
		source,
		{Name: "First", ProjectID: "first"},
		{Name: "Second", ProjectID: "second"},
		{Name: "Third", ProjectID: "third"},
	}
	m := New(nil).Open(source, projects).SetBounds(24, 0, 76, 20).SetTargetRow(12)
	_, beforeY := m.TargetPosition()
	m, _ = m.Update(key(tea.KeyDown))
	_, afterY := m.TargetPosition()
	if afterY != beforeY-1 {
		t.Fatalf("popup Y after moving down = %d, want %d", afterY, beforeY-1)
	}
	if selectedRow := afterY + targetFirstOptionRow + m.pickerCursor; selectedRow != 12 {
		t.Fatalf("selected target row = %d, want source row 12", selectedRow)
	}
	if current := m.candidates[m.pickerCursor].ProjectID; current != "second" {
		t.Fatalf("selected target = %q, want second", current)
	}
	lines := strings.Split(testutil.NormalizeViewSnapshot(m.TargetView()), "\n")
	if len(lines) < 6 || !strings.HasPrefix(lines[1], "╯ First") || !strings.HasPrefix(lines[3], "  Third") || !strings.HasPrefix(lines[4], "╮ Filter:") {
		t.Fatalf("row below selected target did not open the shared seam:\n%s", strings.Join(lines, "\n"))
	}
}

func TestTargetPickerCanMoveAboveScreenToPreserveAlignment(t *testing.T) {
	source := core.Project{Name: "Source", ProjectID: "source"}
	projects := []core.Project{
		source,
		{Name: "First", ProjectID: "first"},
		{Name: "Second", ProjectID: "second"},
		{Name: "Third", ProjectID: "third"},
		{Name: "Fourth", ProjectID: "fourth"},
	}
	m := New(nil).Open(source, projects).SetBounds(24, 0, 76, 20).SetTargetRow(1)
	m, _ = m.Update(key(tea.KeyEnd))
	_, y := m.TargetPosition()
	if y != -3 {
		t.Fatalf("popup Y = %d, want -3 so it can extend above the screen", y)
	}
	if selectedRow := y + targetFirstOptionRow + m.pickerCursor; selectedRow != 1 {
		t.Fatalf("selected target row = %d, want source row 1", selectedRow)
	}
}

func TestTargetPickerOverReviewPreservesWorkspaceWhenCanceled(t *testing.T) {
	m := promoteTestModel(t, dependencyPlan(t))
	beforeItem, ok := m.CurrentItem()
	if !ok {
		t.Fatal("review has no current item")
	}
	beforePreview := m.Preview()
	beforeSource, beforeTarget := m.Source(), m.Target()
	pickerSource := core.Project{Name: "Alternate", ProjectID: "alternate"}

	m = m.OpenTargetPicker(pickerSource, []core.Project{
		pickerSource,
		{Name: "Replacement", ProjectID: "replacement"},
	})
	if !m.TargetPickerOpen() || !m.WorkspaceOpen() {
		t.Fatalf("picker over review = picker:%v workspace:%v, want both open", m.TargetPickerOpen(), m.WorkspaceOpen())
	}
	m, command := m.Update(key(tea.KeyEscape))
	if command != nil {
		t.Fatal("canceling a picker over review requested closing the workspace")
	}
	afterItem, ok := m.CurrentItem()
	if !ok || afterItem.ID != beforeItem.ID {
		t.Fatalf("current review item after cancel = %#v, want %#v", afterItem.ID, beforeItem.ID)
	}
	if m.TargetPickerOpen() || !m.WorkspaceOpen() || m.Preview() != beforePreview {
		t.Fatalf("canceled picker = picker:%v workspace:%v preview-preserved:%v", m.TargetPickerOpen(), m.WorkspaceOpen(), m.Preview() == beforePreview)
	}
	if m.Source().ProjectID != beforeSource.ProjectID || m.Target().ProjectID != beforeTarget.ProjectID {
		t.Fatalf("canceled picker changed pair to %v -> %v", m.Source(), m.Target())
	}
}

func TestSelectingParameterMarksRequiredCondition(t *testing.T) {
	m := promoteTestModel(t, dependencyPlan(t))
	next, _ := m.Update(key(tea.KeySpace))
	if !next.HasSelection() || len(next.preview.Required) != 1 || next.preview.Required[0].ID.Name != "beta" {
		t.Fatalf("preview = %#v, want selected flag and required beta", next.preview)
	}
	rawView := next.ViewWithBorder(true, true)
	view := testutil.NormalizeViewSnapshot(rawView)
	for _, text := range []string{"[✓] + flag", "[•] + ● beta required", "1 required"} {
		if !strings.Contains(view, text) {
			t.Fatalf("promote view missing %q:\n%s", text, view)
		}
	}
	for name, want := range map[string]string{
		"required marker": styles.DetailsValue.Render("[•]"),
		"required hint":   styles.DetailsValue.Render(" required"),
	} {
		if !strings.Contains(rawView, want) {
			t.Fatalf("%s does not use the selection marker color:\n%s", name, view)
		}
	}
}

func TestPromoteConditionRowsSeparateDiffAndConditionColors(t *testing.T) {
	source := &firebase.RemoteConfig{Conditions: []firebase.RemoteConfigCondition{
		{Name: "changed", Expression: "new", TagColor: "GREEN"},
		{Name: "added", Expression: "true", TagColor: "PURPLE"},
	}}
	target := &firebase.RemoteConfig{Conditions: []firebase.RemoteConfigCondition{
		{Name: "changed", Expression: "old", TagColor: "RED"},
		{Name: "removed", Expression: "true", TagColor: "DEEP_ORANGE"},
	}}
	m := promoteTestModel(t, makePlan(t, source, target))
	wantColors := map[string]string{
		"added":   "PURPLE",
		"changed": "GREEN",
		"removed": "DEEP_ORANGE",
	}
	for _, item := range m.visible {
		if item.Kind != rcdiff.ItemCondition {
			continue
		}
		row := m.renderChangeRow(item, false, 60)
		diffStyle := changeStyle(item.Change)
		conditionStyle := styles.DetailsConditionValueStyle(wantColors[item.ID.Name])
		if !m.changeSelectable(item) {
			diffStyle = diffStyle.Faint(true)
			conditionStyle = conditionStyle.Faint(true)
		}
		for name, want := range map[string]string{
			"indicator": diffStyle.Render(changeSymbol(item.Change) + " "),
			"circle":    conditionStyle.Render("●"),
			"name":      diffStyle.Render(" " + item.ID.Name),
		} {
			if !strings.Contains(row, want) {
				t.Fatalf("%s for %s condition has the wrong style:\n%s", name, item.ID.Name, testutil.NormalizeViewSnapshot(row))
			}
		}
	}
}

func TestPromoteSelectedChangeStylesContentAndPadding(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	m := promoteTestModel(t, dependencyPlan(t))
	item := m.visible[0]
	const width = 48
	row := m.renderChangeRow(item, true, width)
	selection := styles.TreeItemSelectionStyle()
	background := selection.GetBackground()
	for name, want := range map[string]string{
		"marker": styles.DetailsValue.Background(background).Render("[ ]"),
		"change": changeStyle(item.Change).Background(background).Render(changeSymbol(item.Change) + " "),
		"name":   changeStyle(item.Change).Background(background).Render(item.ID.Name),
	} {
		if !strings.Contains(row, want) {
			t.Fatalf("selected row %s is missing the selection background:\n%s", name, testutil.NormalizeViewSnapshot(row))
		}
	}
	if got := lipgloss.Width(row); got != width {
		t.Fatalf("selected row width = %d, want %d", got, width)
	}
}

func TestTargetOnlyItemRequiresPrune(t *testing.T) {
	source := &firebase.RemoteConfig{Parameters: map[string]firebase.RemoteConfigParam{}}
	target := &firebase.RemoteConfig{Parameters: map[string]firebase.RemoteConfigParam{"obsolete": stringParam("on")}}
	m := promoteTestModel(t, makePlan(t, source, target))
	rawChanges := m.renderChangeRow(m.visible[0], false, 40)
	for name, want := range map[string]string{
		"disabled marker":    styles.DetailsLabel.Faint(true).Render("[ ]"),
		"disabled indicator": changeStyle(rcdiff.ChangeRemoved).Faint(true).Render("- "),
		"disabled name":      changeStyle(rcdiff.ChangeRemoved).Faint(true).Render("obsolete"),
		"disabled hint":      styles.DetailsLabel.Faint(true).Render(" kept"),
	} {
		if !strings.Contains(rawChanges, want) {
			t.Fatalf("%s is not muted:\n%s", name, testutil.NormalizeViewSnapshot(rawChanges))
		}
	}
	if detail := testutil.NormalizeViewSnapshot(strings.Join(m.detailLines(), "\n")); !strings.Contains(detail, "× TARGET-ONLY") {
		t.Fatalf("non-pruned target-only item does not use TARGET-ONLY action wording:\n%s", detail)
	}
	next, _ := m.Update(key(tea.KeySpace))
	if next.HasSelection() {
		t.Fatal("target-only item selected while pruning was off")
	}
	next, _ = next.Update(keyRune('x'))
	if detail := testutil.NormalizeViewSnapshot(strings.Join(next.detailLines(), "\n")); !strings.Contains(detail, "- TO REMOVE") {
		t.Fatalf("pruned target-only item does not use TO REMOVE action wording:\n%s", detail)
	}
	next, _ = next.Update(key(tea.KeySpace))
	if !next.HasSelection() || !next.prune {
		t.Fatalf("pruned selection = selected:%v prune:%v", next.HasSelection(), next.prune)
	}
}

func TestPromoteChangesColumnUsesLongestVisibleName(t *testing.T) {
	const longName = "ANALYTICS_SERVICE_URL_WITH_A_FULL_DESCRIPTIVE_SUFFIX"
	source := &firebase.RemoteConfig{
		Parameters: map[string]firebase.RemoteConfigParam{
			longName: stringParam("https://listener.norago.tv"),
		},
	}
	m := promoteTestModel(t, makePlan(t, source, &firebase.RemoteConfig{}))
	leftWidth, rightWidth := m.promotionColumnWidths(138)
	wantLeftWidth := lipgloss.Width(" [ ] + " + longName)
	if leftWidth < wantLeftWidth {
		t.Fatalf("left width = %d, want at least natural row width %d", leftWidth, wantLeftWidth)
	}
	if rightWidth < 27 {
		t.Fatalf("right width = %d, want responsive details minimum 27", rightWidth)
	}
	left := testutil.NormalizeViewSnapshot(strings.Join(m.changeLines(leftWidth, 10), "\n"))
	if !strings.Contains(left, longName) || strings.Contains(left, longName[:len(longName)-1]+"…") {
		t.Fatalf("wide changes column cropped a name that fits:\n%s", left)
	}

	narrowLeft, narrowRight := m.promotionColumnWidths(50)
	if narrowLeft+narrowRight+1 != 50 || narrowLeft >= leftWidth {
		t.Fatalf("narrow widths = %d + 1 + %d; wide left was %d", narrowLeft, narrowRight, leftWidth)
	}
}

func TestPromoteViewFitsMinimumWorkspaceWidth(t *testing.T) {
	m := promoteTestModel(t, dependencyPlan(t)).SetBounds(0, 0, 77, 18)
	view := m.ViewWithBorder(true, true)
	if width := maxLineWidth(view); width != 77 {
		t.Fatalf("view width = %d, want 77:\n%s", width, testutil.NormalizeViewSnapshot(view))
	}
}

func TestPromoteDetailsWrapAtResponsiveColumnBorder(t *testing.T) {
	const (
		longName = "ANALYTICS_SERVICE_URL_WITH_A_FULL_DESCRIPTIVE_SUFFIX"
	)
	longValue := strings.Repeat("abcdefghij", 14)
	source := &firebase.RemoteConfig{
		Parameters: map[string]firebase.RemoteConfigParam{
			longName: {
				ValueType:    "STRING",
				DefaultValue: &firebase.RemoteConfigValue{Value: longValue},
			},
		},
	}
	m := promoteTestModel(t, makePlan(t, source, &firebase.RemoteConfig{})).SetBounds(0, 0, 100, 48)
	_, rightWidth := m.promotionColumnWidths(m.width - 2)
	if got := m.detail.Width(); got != rightWidth {
		t.Fatalf("Details viewport width = %d, responsive column width = %d", got, rightWidth)
	}

	detail := testutil.NormalizeViewSnapshot(m.detail.View())
	if strings.Contains(detail, longValue) {
		t.Fatalf("long value remained on one line instead of wrapping at width %d:\n%s", rightWidth, detail)
	}
	compact := strings.NewReplacer("\n", "", " ", "").Replace(detail)
	if !strings.Contains(compact, longValue) {
		t.Fatalf("wrapped Details value lost content:\n%s", detail)
	}
	view := testutil.NormalizeViewSnapshot(m.ViewWithBorder(true, true))
	if strings.Contains(view, "…") {
		t.Fatalf("wrapped Details value produced a synthetic ellipsis:\n%s", view)
	}
}

func TestPromoteParameterValueWrapsWithHangingIndentAndLineBudget(t *testing.T) {
	options := detailRenderOptions{width: 24, valueLineBudget: detailValueMaxLineBudget}
	shortLines := parameterValueLines(
		"default",
		"",
		firebase.RemoteConfigValue{Value: strings.Repeat("x", 35)},
		"STRING",
		options,
	)
	if len(shortLines) != 3 {
		t.Fatalf("short wrapped value uses %d lines, want 3:\n%s", len(shortLines), testutil.NormalizeViewSnapshot(strings.Join(shortLines, "\n")))
	}
	short := strings.Split(testutil.NormalizeViewSnapshot(strings.Join(shortLines, "\n")), "\n")
	for index, line := range short[1:] {
		if !strings.HasPrefix(line, "       x") {
			t.Fatalf("continuation %d does not begin two cells after the property name:\n%s", index+1, strings.Join(short, "\n"))
		}
	}
	if strings.Contains(strings.Join(short, "\n"), "…") {
		t.Fatalf("short wrapped value was cropped:\n%s", strings.Join(short, "\n"))
	}

	longLines := parameterValueLines(
		"default",
		"",
		firebase.RemoteConfigValue{Value: strings.Repeat("x", 120)},
		"STRING",
		options,
	)
	long := testutil.NormalizeViewSnapshot(strings.Join(longLines, "\n"))
	if len(longLines) != detailValueMaxLineBudget || !strings.HasSuffix(long, "…") {
		t.Fatalf("long value was not cropped to %d rendered lines with an ellipsis:\n%s", detailValueMaxLineBudget, long)
	}
}

func TestPromoteCropsHugeJSONBeforeSyntaxHighlighting(t *testing.T) {
	huge := `{"payload":"` + strings.Repeat("abcdefghij", 100_000) + `","tail":"must not be scanned"}`
	fragment, cropped := cropValueBeforeRender(huge, 12, 16, detailValueMaxLineBudget)
	if !cropped {
		t.Fatal("huge JSON fragment was not cropped")
	}
	if len(fragment) > 100 || strings.Contains(fragment, "must not be scanned") {
		t.Fatalf("pre-highlight fragment contains too much input: len=%d, tail=%v", len(fragment), strings.Contains(fragment, "must not be scanned"))
	}

	lines := parameterValueLines(
		"default",
		"",
		firebase.RemoteConfigValue{Value: huge},
		"JSON",
		detailRenderOptions{width: 24, valueLineBudget: detailValueMaxLineBudget},
	)
	rendered := testutil.NormalizeViewSnapshot(strings.Join(lines, "\n"))
	if len(lines) != detailValueMaxLineBudget || !strings.HasSuffix(rendered, "…") {
		t.Fatalf("huge JSON preview was not highlighted as a cropped fragment:\n%s", rendered)
	}
}

func TestPromoteCropsHugeStringBeforeStylingAndWrapping(t *testing.T) {
	huge := strings.Repeat("abcdefghij", 100_000) + "tail must not be rendered"
	fragment, cropped := cropValueBeforeRender(huge, 12, 16, detailValueMaxLineBudget)
	if !cropped {
		t.Fatal("huge string fragment was not cropped")
	}
	if len(fragment) > 100 || strings.Contains(fragment, "tail must not be rendered") {
		t.Fatalf("pre-render fragment contains too much input: len=%d, tail=%v", len(fragment), strings.Contains(fragment, "tail must not be rendered"))
	}

	lines := parameterValueLines(
		"default",
		"",
		firebase.RemoteConfigValue{Value: huge},
		"STRING",
		detailRenderOptions{width: 24, valueLineBudget: detailValueMaxLineBudget},
	)
	rendered := testutil.NormalizeViewSnapshot(strings.Join(lines, "\n"))
	if len(lines) != detailValueMaxLineBudget || !strings.HasSuffix(rendered, "…") {
		t.Fatalf("huge string preview was not styled as a cropped fragment:\n%s", rendered)
	}
	if strings.Contains(rendered, "tail must not be rendered") {
		t.Fatalf("huge string preview contains the cropped tail:\n%s", rendered)
	}
}

func TestPromoteDetailReducesValueBudgetToKeepStatusVisible(t *testing.T) {
	source := &firebase.RemoteConfig{Parameters: map[string]firebase.RemoteConfigParam{
		"payload": stringParam(strings.Repeat("abcdefghij", 60)),
	}}
	plan := makePlan(t, source, &firebase.RemoteConfig{})

	tall := promoteTestModel(t, plan).SetBounds(0, 0, 100, 48)
	tallLines := tall.detailLines()
	if rows := renderedValueRows(t, tallLines, "default:", "🬞🬏"); rows != detailValueMaxLineBudget {
		t.Fatalf("tall detail value uses %d rows, want initial budget %d:\n%s", rows, detailValueMaxLineBudget, testutil.NormalizeViewSnapshot(strings.Join(tallLines, "\n")))
	}

	short := promoteTestModel(t, plan).SetBounds(0, 0, 100, 23)
	shortLines := short.detailLines()
	if len(shortLines) > short.detail.Height() {
		t.Fatalf("adaptive detail uses %d rows in a %d-row viewport:\n%s", len(shortLines), short.detail.Height(), testutil.NormalizeViewSnapshot(strings.Join(shortLines, "\n")))
	}
	if rows := renderedValueRows(t, shortLines, "default:", "🬞🬏"); rows != detailValueMinLineBudget {
		t.Fatalf("short detail value uses %d rows, want reduced budget %d:\n%s", rows, detailValueMinLineBudget, testutil.NormalizeViewSnapshot(strings.Join(shortLines, "\n")))
	}
	if detail := testutil.NormalizeViewSnapshot(strings.Join(shortLines, "\n")); !strings.Contains(detail, "Not selected; target remains unchanged.") {
		t.Fatalf("adaptive detail omitted its final status hint:\n%s", detail)
	}
}

func TestPromoteDetailHidesUnchangedFieldsAfterMinimumValueBudget(t *testing.T) {
	sameValues := map[string]firebase.RemoteConfigValue{}
	for index := range 8 {
		name := fmt.Sprintf("same-%d", index)
		sameValues[name] = firebase.RemoteConfigValue{Value: name}
	}
	sourceParam := firebase.RemoteConfigParam{
		ValueType:         "STRING",
		Description:       "source description",
		ConditionalValues: sameValues,
	}
	targetParam := sourceParam
	targetParam.Description = "target description"
	source := &firebase.RemoteConfig{Parameters: map[string]firebase.RemoteConfigParam{"payload": sourceParam}}
	target := &firebase.RemoteConfig{Parameters: map[string]firebase.RemoteConfigParam{"payload": targetParam}}

	m := promoteTestModel(t, makePlan(t, source, target)).SetBounds(0, 0, 100, 19)
	lines := m.detailLines()
	detail := testutil.NormalizeViewSnapshot(strings.Join(lines, "\n"))
	if len(lines) > m.detail.Height() {
		t.Fatalf("compacted detail uses %d rows in a %d-row viewport:\n%s", len(lines), m.detail.Height(), detail)
	}
	for _, want := range []string{"source description", "target description", "Not selected; target remains unchanged."} {
		if !strings.Contains(detail, want) {
			t.Fatalf("compacted detail omitted %q:\n%s", want, detail)
		}
	}
	for _, hidden := range []string{"type:", "group:", "values:", "same-0"} {
		if strings.Contains(detail, hidden) {
			t.Fatalf("compacted detail retained unchanged field %q:\n%s", hidden, detail)
		}
	}
}

func renderedValueRows(t *testing.T, lines []string, valueLabel, following string) int {
	t.Helper()
	start, end := -1, -1
	for index, line := range lines {
		plain := testutil.NormalizeViewSnapshot(line)
		if start < 0 && strings.Contains(plain, valueLabel) {
			start = index
			continue
		}
		if start >= 0 && strings.Contains(plain, following) {
			end = index
			break
		}
	}
	if start < 0 || end <= start {
		t.Fatalf("cannot find value %q followed by %q:\n%s", valueLabel, following, testutil.NormalizeViewSnapshot(strings.Join(lines, "\n")))
	}
	return end - start
}

func TestPromoteViewRendersCompleteLeftBorder(t *testing.T) {
	m := promoteTestModel(t, dependencyPlan(t)).SetBounds(0, 0, 77, 18)
	lines := strings.Split(testutil.NormalizeViewSnapshot(m.ViewWithBorder(true, true)), "\n")
	if !strings.HasPrefix(lines[0], "╭") || !strings.Contains(lines[0], "⁹Promote Remote Config") {
		t.Fatalf("top border does not include the Promote key and left corner:\n%s", strings.Join(lines, "\n"))
	}
	for index, line := range lines[1 : len(lines)-1] {
		if !strings.HasPrefix(line, "│") && !strings.HasPrefix(line, "├") {
			t.Fatalf("line %d has no left border: %q", index+1, line)
		}
	}
	if !strings.HasPrefix(lines[len(lines)-1], "╰") {
		t.Fatalf("bottom border has no left corner: %q", lines[len(lines)-1])
	}
}

func TestPromoteViewOmitsDuplicateActionsAndSyntheticRightEllipses(t *testing.T) {
	m := promoteTestModel(t, dependencyPlan(t)).SetBounds(0, 0, 100, 24)
	view := testutil.NormalizeViewSnapshot(m.ViewWithBorder(true, true))
	for _, action := range []string{"[d Save Draft]", "[p Publish Now]", "[s Swap]", "[x Prune]", "[esc Close]"} {
		if strings.Contains(view, action) {
			t.Fatalf("promote body still contains duplicate action %q:\n%s", action, view)
		}
	}
	if strings.Contains(view, "…") {
		t.Fatalf("promote review contains a synthetic right-pane ellipsis:\n%s", view)
	}
}

func TestPromoteChangeColorsUseSharedDiffPalette(t *testing.T) {
	tests := []struct {
		kind rcdiff.ChangeKind
		want color.Color
	}{
		{kind: rcdiff.ChangeAdded, want: styles.PaletteAdded},
		{kind: rcdiff.ChangeRemoved, want: styles.PaletteRemoved},
		{kind: rcdiff.ChangeChanged, want: styles.PaletteChanged},
	}
	for _, tt := range tests {
		if got := changeStyle(tt.kind).GetForeground(); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("changeStyle(%s) foreground = %v, want %v", tt.kind, got, tt.want)
		}
	}

	view := promoteTestModel(t, dependencyPlan(t)).ViewWithBorder(true, true)
	if want := changeStyle(rcdiff.ChangeAdded).Render("+2 to add"); !strings.Contains(view, want) {
		t.Fatalf("promote summary does not render additions with the shared diff color:\n%s", testutil.NormalizeViewSnapshot(view))
	}

	pruned := promoteTestModel(t, dependencyPlan(t))
	pruned.prune = true
	if want := "prune " + changeStyle(rcdiff.ChangeRemoved).Render("ON"); !strings.Contains(pruned.summaryLine(), want) {
		t.Fatalf("enabled prune summary does not render ON with the removal color:\n%s", testutil.NormalizeViewSnapshot(pruned.summaryLine()))
	}
}

func TestPromoteUsesPromotionActionWording(t *testing.T) {
	tests := []struct {
		kind  rcdiff.ChangeKind
		prune bool
		want  string
	}{
		{kind: rcdiff.ChangeAdded, want: "+ TO ADD"},
		{kind: rcdiff.ChangeChanged, want: "~ TO UPDATE"},
		{kind: rcdiff.ChangeRemoved, want: "× TARGET-ONLY"},
		{kind: rcdiff.ChangeRemoved, prune: true, want: "- TO REMOVE"},
	}
	for _, tt := range tests {
		if got := changeActionLabel(tt.kind, tt.prune); got != tt.want {
			t.Errorf("changeActionLabel(%s, %v) = %q, want %q", tt.kind, tt.prune, got, tt.want)
		}
	}

	view := testutil.NormalizeViewSnapshot(promoteTestModel(t, dependencyPlan(t)).ViewWithBorder(true, true))
	for _, want := range []string{"+2 to add", "~0 to update", "-0 target-only", "+ TO ADD"} {
		if !strings.Contains(view, want) {
			t.Fatalf("Promote view does not use promotion wording %q:\n%s", want, view)
		}
	}
	for _, obsolete := range []string{"+2 added", "~0 changed", "ADDED"} {
		if strings.Contains(view, obsolete) {
			t.Fatalf("Promote view still uses historical diff wording %q:\n%s", obsolete, view)
		}
	}
}

func TestPromoteHeaderShowsNamedProjectsAndAlignedSnapshotStates(t *testing.T) {
	plan := dependencyPlan(t)
	plan.Source.Project = core.Project{Name: "SignalOps Console", ProjectID: "signalops-console"}
	plan.Source.Source = "cache-verified"
	plan.Source.Version = "2"
	plan.Target.Project = core.Project{Name: "Mercato Mobile", ProjectID: "mercato-mobile-9eac5"}
	plan.Target.Source = "cache-verified"
	plan.Target.Version = "1"
	m := promoteTestModel(t, plan)
	raw := m.ViewWithBorder(true, true)
	if !strings.Contains(raw, styles.PanelText.Render("SignalOps Console")) ||
		!strings.Contains(raw, styles.PanelMuted.Render(" (signalops-console)")) ||
		!strings.Contains(raw, styles.PanelText.Render("Mercato Mobile")) ||
		!strings.Contains(raw, styles.PanelMuted.Render(" (mercato-mobile-9eac5)")) {
		t.Fatalf("project names and IDs do not use Projects-panel styles:\n%s", testutil.NormalizeViewSnapshot(raw))
	}

	lines := strings.Split(testutil.NormalizeViewSnapshot(raw), "\n")
	if len(lines) < 5 ||
		!strings.Contains(lines[1], "SignalOps Console (signalops-console)") ||
		!strings.Contains(lines[1], "🬭🬿 Mercato Mobile (mercato-mobile-9eac5)") ||
		!strings.Contains(lines[2], "cache-verified v2") ||
		!strings.Contains(lines[2], "🬂🭚 cache-verified v1") {
		t.Fatalf("promotion header layout is missing:\n%s", strings.Join(lines, "\n"))
	}
	projectConnector := strings.Index(lines[1], "🬭")
	stateConnector := strings.Index(lines[2], "🬂")
	if projectConnector < 0 || stateConnector < 0 ||
		lipgloss.Width(lines[1][:projectConnector]) != lipgloss.Width(lines[2][:stateConnector]) {
		t.Fatalf("header connectors are not aligned:\n%s\n%s", lines[1], lines[2])
	}
	wideAfter := lines[1][projectConnector+len("🬭🬿"):]
	if !strings.HasSuffix(lines[1][:projectConnector], ") ") || !strings.HasPrefix(wideAfter, " ") || strings.HasPrefix(wideAfter, "  ") {
		t.Fatalf("wide header arrow does not have exactly one surrounding space: %q", lines[1])
	}

	narrowLines := strings.Split(testutil.NormalizeViewSnapshot(m.SetBounds(0, 0, 77, 18).ViewWithBorder(true, true)), "\n")
	narrowConnector := strings.Index(narrowLines[1], "🬭")
	if narrowConnector < 0 {
		t.Fatalf("narrow header has no arrow: %q", narrowLines[1])
	}
	narrowBefore := narrowLines[1][:narrowConnector]
	narrowAfter := narrowLines[1][narrowConnector+len("🬭🬿"):]
	if !strings.HasSuffix(narrowBefore, " ") || strings.HasSuffix(narrowBefore, "  ") ||
		!strings.HasPrefix(narrowAfter, " ") || strings.HasPrefix(narrowAfter, "  ") {
		t.Fatalf("narrow header arrow does not have exactly one surrounding space: %q", narrowLines[1])
	}

	plan.Source.Project = core.Project{Name: "PulseForge Fitness", ProjectID: "pulseforge-fitness-f60f7"}
	pulseForgeHeader := strings.Split(testutil.NormalizeViewSnapshot(promoteTestModel(t, plan).ViewWithBorder(true, true)), "\n")[1]
	if !strings.Contains(pulseForgeHeader, "PulseForge Fitness (pulseforge-fitness-f60f7) 🬭🬿") {
		t.Fatalf("source ID was cropped even though the complete header fits: %q", pulseForgeHeader)
	}
}

func TestPromoteLoadingHeaderPersistsIntoReview(t *testing.T) {
	source := core.Project{Name: "SignalOps Console", ProjectID: "signalops-console"}
	target := core.Project{Name: "Mercato Mobile", ProjectID: "mercato-mobile-9eac5"}
	m := New(nil).SetBounds(0, 0, 100, 24).SetLoading(source, target, core.ProjectPromotionEffective)
	loadingView := m.ViewWithBorder(true, true)
	loadingLines := strings.Split(testutil.NormalizeViewSnapshot(loadingView), "\n")
	if len(loadingLines) < 5 ||
		!strings.Contains(loadingLines[1], "SignalOps Console (signalops-console) 🬭🬿 Mercato Mobile (mercato-mobile-9eac5)") ||
		!strings.Contains(loadingLines[2], "🬂🭚") || strings.Count(loadingLines[2], "…") != 2 ||
		!strings.Contains(loadingLines[3], "Loading Remote Config snapshots…") {
		t.Fatalf("loading header is incomplete:\n%s", strings.Join(loadingLines, "\n"))
	}
	if strings.Count(loadingView, styles.PanelMuted.Render(" …")) != 2 {
		t.Fatalf("loading snapshot placeholders do not use the muted style:\n%s", strings.Join(loadingLines, "\n"))
	}

	plan := dependencyPlan(t)
	plan.Source.Project = source
	plan.Target.Project = target
	plan.Source.Source, plan.Target.Source = "cache-verified", "cache-verified"
	plan.Source.Version, plan.Target.Version = "2", "1"
	reviewLines := strings.Split(testutil.NormalizeViewSnapshot(m.SetPlan(plan, true).ViewWithBorder(true, true)), "\n")
	if loadingLines[1] != reviewLines[1] {
		t.Fatalf("project header moved after loading:\nloading: %s\nreview:  %s", loadingLines[1], reviewLines[1])
	}
	if !strings.Contains(reviewLines[2], "cache-verified v2") ||
		!strings.Contains(reviewLines[2], "cache-verified v1") ||
		strings.Contains(reviewLines[3], "Loading Remote Config snapshots…") ||
		!strings.Contains(reviewLines[3], "to add") {
		t.Fatalf("review header did not replace loading metadata:\n%s", strings.Join(reviewLines, "\n"))
	}
}

func TestPromoteHeaderBlockExpandsForLongerSnapshotStatus(t *testing.T) {
	plan := dependencyPlan(t)
	plan.Source.Project = core.Project{Name: "Adyl TV", ProjectID: "adyl-tv"}
	plan.Source.Source = "cache-verified"
	plan.Source.Version = "12345"
	plan.Target.Project = core.Project{Name: "BlueTV", ProjectID: "blue-4bdb4"}
	plan.Target.Source = "cache-verified"
	plan.Target.Version = "81"

	lines := strings.Split(testutil.NormalizeViewSnapshot(promoteTestModel(t, plan).ViewWithBorder(true, true)), "\n")
	if !strings.Contains(lines[2], "cache-verified v12345") || strings.Contains(lines[2], "cache-verified v…") {
		t.Fatalf("source snapshot status was cropped despite available space: %q", lines[2])
	}
	projectConnector := strings.Index(lines[1], "🬭")
	stateConnector := strings.Index(lines[2], "🬂")
	if projectConnector < 0 || stateConnector < 0 ||
		lipgloss.Width(lines[1][:projectConnector]) != lipgloss.Width(lines[2][:stateConnector]) {
		t.Fatalf("longer status did not establish the source block width:\n%s\n%s", lines[1], lines[2])
	}
	beforeArrow := lines[2][:stateConnector]
	if !strings.HasSuffix(beforeArrow, "v12345 ") || strings.HasSuffix(beforeArrow, "v12345  ") {
		t.Fatalf("status block does not have exactly one space before the arrow: %q", lines[2])
	}
}

func TestPromoteParameterSectionsUseTreeIdentityAndSpacing(t *testing.T) {
	source := &firebase.RemoteConfig{
		Conditions: []firebase.RemoteConfigCondition{{Name: "beta", Expression: "true", TagColor: "GREEN"}},
		ParameterGroups: map[string]firebase.RemoteConfigGroup{
			"ANDROID": {
				Description: "Android endpoints",
				Parameters: map[string]firebase.RemoteConfigParam{
					"ANALYTICS_SERVICE_URL": {
						DefaultValue:      &firebase.RemoteConfigValue{Value: "https://listener-go.norago.tv"},
						ConditionalValues: map[string]firebase.RemoteConfigValue{"beta": {Value: "https://beta.norago.tv"}},
					},
				},
			},
		},
	}
	m := promoteTestModel(t, makePlan(t, source, &firebase.RemoteConfig{}))
	leftRaw := strings.Join(m.changeLines(44, 20), "\n")
	left := testutil.NormalizeViewSnapshot(leftRaw)
	if !strings.Contains(left, "+ ANALYTICS_SERVICE_URL") || strings.Contains(left, "ANDROID/ANALYTICS_SERVICE_URL") {
		t.Fatalf("parameter change row did not omit its group:\n%s", left)
	}
	if !strings.HasPrefix(
		leftRaw,
		styles.DetailsLabel.Render(" Changes")+"\n\n"+styles.DetailsLabel.Render(" Parameters"),
	) {
		t.Fatalf("Changes and section headers do not use the muted Details style with an empty row between them:\n%s", left)
	}
	for _, section := range []string{" Parameters", " Groups", " Conditions"} {
		if !strings.Contains(leftRaw, styles.DetailsLabel.Render(section)) {
			t.Fatalf("left section header %q does not use the muted Details style:\n%s", section, left)
		}
	}
	if !strings.Contains(left, "ANALYTICS_SERVICE_URL\n\n Groups") ||
		!strings.Contains(left, "+ ANDROID\n\n Conditions") ||
		strings.Contains(left, "group ANDROID") {
		t.Fatalf("top-level change sections are not separated by empty rows:\n%s", left)
	}
	for index, item := range m.visible {
		if item.Kind != rcdiff.ItemGroupDescription {
			continue
		}
		groupModel := m
		groupModel.cursor = index
		groupDetail := testutil.NormalizeViewSnapshot(strings.Join(groupModel.detailLines(), "\n"))
		if !strings.Contains(groupDetail, "\n ANDROID\n") || strings.Contains(groupDetail, "group ANDROID") {
			t.Fatalf("selected group change includes a type prefix in its name:\n%s", groupDetail)
		}
		break
	}

	detailRaw := strings.Join(m.detailLines(), "\n")
	wantIdentity := styles.ParameterGroup.Render("ANDROID") +
		styles.ParameterSeparator.Render(" / ") +
		styles.ParameterName.Render("ANALYTICS_SERVICE_URL")
	if !strings.Contains(detailRaw, wantIdentity) {
		t.Fatalf("parameter identity does not use Parameters-tree styles:\n%s", testutil.NormalizeViewSnapshot(detailRaw))
	}
	for name, want := range map[string]string{
		"source project":    detailProjectSection(m.plan.Source.Project, m.detail.Width()),
		"target project":    detailProjectSection(m.plan.Target.Project, m.detail.Width()),
		"group value":       styles.DetailsLabel.Render("group: ") + styles.ParameterGroup.Render("ANDROID"),
		"values label":      styles.DetailsLabel.Render("values:"),
		"field label/value": styles.DetailsEmptyValue.Render("default: ") + styles.DetailsValue.Render("https://listener-go.norago.tv"),
		"condition label":   styles.DetailsConditionValueStyle("GREEN").Render("beta: "),
	} {
		if !strings.Contains(detailRaw, want) {
			t.Fatalf("parameter detail does not use Details %s style:\n%s", name, testutil.NormalizeViewSnapshot(detailRaw))
		}
	}
	detail := testutil.NormalizeViewSnapshot(detailRaw)
	if strings.Contains(detail, "\n Group\n") ||
		!strings.Contains(detail, "   description: —\n   group: ANDROID\n   values:\n     beta:") {
		t.Fatalf("parameter fields are not nested under project sections:\n%s", detail)
	}
	assertTextOrder(t, detail, "Development (dev)", "Production (prod)")
	assertDetailTransition(t, detail, "Development (dev)", "Production (prod)")

	for index, item := range m.visible {
		label := changeItemLabel(item)
		renderedLabel := changeSymbol(item.Change) + " " + label
		if item.Kind == rcdiff.ItemCondition {
			renderedLabel = changeSymbol(item.Change) + " ● " + label
		}
		found := false
		for row, line := range strings.Split(left, "\n") {
			if !strings.Contains(line, renderedLabel) {
				continue
			}
			found = true
			got, ok := m.itemIndexAtBodyRow(row)
			if !ok || got != index {
				t.Fatalf("mouse row %d maps to item %d, %v; want %d", row, got, ok, index)
			}
			break
		}
		if !found {
			t.Fatalf("rendered change row for item %d (%s) was not found:\n%s", index, label, left)
		}
	}
}

func TestPromoteJSONParameterValueUsesDetailsSyntaxHighlighting(t *testing.T) {
	const value = `{"enabled":true,"limit":2,"fallback":null}`
	got := strings.Join(parameterValueLines(
		"default",
		"",
		firebase.RemoteConfigValue{Value: value},
		"JSON",
		detailRenderOptions{width: 100, valueLineBudget: detailValueMaxLineBudget},
	), "\n")
	want := styles.DetailsEmptyValue.Render("default: ") + jsoninput.HighlightJSONVisible(value)
	if !strings.Contains(got, want) {
		t.Fatalf("JSON value does not use Details syntax highlighting:\n%s", testutil.NormalizeViewSnapshot(got))
	}
}

func TestPromoteParameterValuesFollowRemoteConfigEvaluationOrder(t *testing.T) {
	param := &firebase.RemoteConfigParam{
		ValueType:    "STRING",
		DefaultValue: &firebase.RemoteConfigValue{Value: "default"},
		ConditionalValues: map[string]firebase.RemoteConfigValue{
			"alpha":  {Value: "alpha"},
			"beta":   {Value: "beta"},
			"orphan": {Value: "orphan"},
		},
	}
	conditions := []firebase.RemoteConfigCondition{
		{Name: "beta", TagColor: "GREEN"},
		{Name: "alpha", TagColor: "BLUE"},
	}
	detail := testutil.NormalizeViewSnapshot(strings.Join(parameterSide(
		core.Project{Name: "Development", ProjectID: "dev"},
		"",
		param,
		conditions,
		parameterHiddenFields{conditionalValue: map[string]bool{}},
		detailRenderOptions{width: 100, valueLineBudget: detailValueMaxLineBudget},
	), "\n"))
	values := strings.Index(detail, "values:")
	beta := strings.Index(detail, "beta: beta")
	alpha := strings.Index(detail, "alpha: alpha")
	orphan := strings.Index(detail, "orphan: orphan")
	defaultValue := strings.Index(detail, "default: default")
	if values < 0 || beta < values || alpha < beta || orphan < alpha || defaultValue < orphan {
		t.Fatalf("parameter values are not in condition priority order followed by default:\n%s", detail)
	}
}

func TestPromoteConditionDetailUsesDetailsFieldStyles(t *testing.T) {
	sourceProject := core.Project{Name: "Development", ProjectID: "dev"}
	targetProject := core.Project{Name: "Production", ProjectID: "prod"}
	options := detailRenderOptions{width: 100, valueLineBudget: detailValueMaxLineBudget}
	raw := strings.Join(conditionDetail(rcdiff.ConditionChange{
		PreviousPosition: 2,
		FinalPosition:    1,
		Current: &firebase.RemoteConfigCondition{
			Expression: "app.version == '1'",
			TagColor:   "DEEP_ORANGE",
		},
		Final: &firebase.RemoteConfigCondition{
			Expression: "app.version == '2'",
			TagColor:   "GREEN",
		},
	}, sourceProject, targetProject, options), "\n")

	for name, want := range map[string]string{
		"source project": detailProjectSection(sourceProject, options.width),
		"target project": detailProjectSection(targetProject, options.width),
		"position label": styles.DetailsLabel.Render("position: "),
		"priority value": styles.ParameterName.Render("1"),
		"expression": styles.DetailsLabel.Render("expression: ") +
			styles.DetailsValue.Render("app.version == '2'"),
		"target color": styles.DetailsLabel.Render("color: ") +
			styles.DetailsConditionValueStyle("DEEP_ORANGE").Render("● DEEP ORANGE"),
		"source color": styles.DetailsLabel.Render("color: ") +
			styles.DetailsConditionValueStyle("GREEN").Render("● GREEN"),
	} {
		if !strings.Contains(raw, want) {
			t.Fatalf("condition detail does not use Details %s style:\n%s", name, testutil.NormalizeViewSnapshot(raw))
		}
	}
}

func TestPromoteDetailsRenderSourceBeforeTarget(t *testing.T) {
	sourceProject := core.Project{Name: "Source Project", ProjectID: "source"}
	targetProject := core.Project{Name: "Target Project", ProjectID: "target"}
	options := detailRenderOptions{width: 100, valueLineBudget: detailValueMaxLineBudget}
	sourceParam := stringParam("source value")
	targetParam := stringParam("target value")
	parameter := testutil.NormalizeViewSnapshot(strings.Join(parameterDetail(rcdiff.ParameterChange{
		Group:         "source group",
		PreviousGroup: "target group",
		Final:         &sourceParam,
		Current:       &targetParam,
	}, sourceProject, targetProject, nil, nil, options), "\n"))
	assertTextOrder(t, parameter, "Source Project (source)", "Target Project (target)")
	assertTextOrder(t, parameter, "group: source group", "group: target group")
	assertDetailTransition(t, parameter, "Source Project (source)", "Target Project (target)")

	condition := testutil.NormalizeViewSnapshot(strings.Join(conditionDetail(rcdiff.ConditionChange{
		FinalPosition:    1,
		PreviousPosition: 2,
		Final:            &firebase.RemoteConfigCondition{Expression: "source expression"},
		Current:          &firebase.RemoteConfigCondition{Expression: "target expression"},
	}, sourceProject, targetProject, options), "\n"))
	assertTextOrder(t, condition, "Source Project (source)", "Target Project (target)")
	assertTextOrder(t, condition, "source expression", "target expression")
	assertDetailTransition(t, condition, "Source Project (source)", "Target Project (target)")
	if strings.Count(condition, "position:") != 2 || strings.Contains(condition, "Source position") || strings.Contains(condition, "Target position") {
		t.Fatalf("condition positions are not nested inside project sections:\n%s", condition)
	}

	group := testutil.NormalizeViewSnapshot(strings.Join(groupDescriptionDetail(rcdiff.GroupDescriptionChange{
		Final:   "source description",
		Current: "target description",
	}, sourceProject, targetProject, options), "\n"))
	assertTextOrder(t, group, "Source Project (source)", "Target Project (target)")
	assertTextOrder(t, group, "description: source description", "description: target description")
	assertDetailTransition(t, group, "Source Project (source)", "Target Project (target)")
}

func TestPromoteGroupDetailsDistinguishAbsentAndEmptyGroups(t *testing.T) {
	sourceProject := core.Project{Name: "Source Project", ProjectID: "source"}
	targetProject := core.Project{Name: "Target Project", ProjectID: "target"}
	options := detailRenderOptions{width: 100, valueLineBudget: detailValueMaxLineBudget}

	added := testutil.NormalizeViewSnapshot(strings.Join(groupDescriptionDetail(rcdiff.GroupDescriptionChange{
		Kind:  rcdiff.ChangeAdded,
		Final: "",
	}, sourceProject, targetProject, options), "\n"))
	if !strings.Contains(added, "Source Project (source)\n   description: —") {
		t.Fatalf("existing source group with an empty description is not rendered as a field:\n%s", added)
	}
	if !strings.Contains(added, "Target Project (target)\n   —") ||
		strings.Contains(added[strings.Index(added, "Target Project (target)"):], "description:") {
		t.Fatalf("absent target group is not rendered as a bare empty value:\n%s", added)
	}

	removed := testutil.NormalizeViewSnapshot(strings.Join(groupDescriptionDetail(rcdiff.GroupDescriptionChange{
		Kind:    rcdiff.ChangeRemoved,
		Current: "target description",
	}, sourceProject, targetProject, options), "\n"))
	if !strings.Contains(removed, "Source Project (source)\n   —") ||
		!strings.Contains(removed, "Target Project (target)\n   description: target description") {
		t.Fatalf("removed group does not distinguish the absent source from the existing target:\n%s", removed)
	}
}

func TestPromotePreparesSelectedEntityForGenericDiff(t *testing.T) {
	source := &firebase.RemoteConfig{ParameterGroups: map[string]firebase.RemoteConfigGroup{
		"ANDROID": {Description: "Mobile app", Parameters: map[string]firebase.RemoteConfigParam{
			"payload": {
				ValueType:   "JSON",
				Description: "source description",
				DefaultValue: &firebase.RemoteConfigValue{
					Value: `{"enabled":true,"items":[1,2]}`,
				},
			},
		}},
	}}
	target := &firebase.RemoteConfig{ParameterGroups: map[string]firebase.RemoteConfigGroup{
		"ANDROID": {Description: "Mobile app", Parameters: map[string]firebase.RemoteConfigParam{
			"payload": {
				ValueType:   "JSON",
				Description: "target description",
				DefaultValue: &firebase.RemoteConfigValue{
					Value: `{"enabled":false,"items":[1,2]}`,
				},
			},
		}},
	}}
	m := promoteTestModel(t, makePlan(t, source, target))
	var item rcpromote.Item
	for _, candidate := range m.visible {
		if candidate.Kind == rcdiff.ItemParameter {
			item = candidate
			break
		}
	}
	input, ok := m.DiffInput(item)
	if !ok {
		t.Fatal("parameter change was not prepared for generic diff")
	}
	if input.EntityName != "Property: ANDROID / payload" ||
		input.Left.Name != "Current target: Production (prod)" ||
		input.Right.Name != "Promotion source: Development (dev)" {
		t.Fatalf("generic diff identity = %q, %q -> %q", input.EntityName, input.Left.Name, input.Right.Name)
	}
	if got := input.Left.Properties["description"].Raw; got != "target description" {
		t.Fatalf("left diff value = %q, want current target value", got)
	}
	if got := input.Right.Properties["description"].Raw; got != "source description" {
		t.Fatalf("right diff value = %q, want promotion source value", got)
	}
	if got := input.Left.Properties["value · default"].Type; got != dictdiff.ValueJSON {
		t.Fatalf("default value type = %q, want JSON", got)
	}
	result, err := dictdiff.Compare(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Properties) != 2 {
		t.Fatalf("changed properties = %d, want description and default value", len(result.Properties))
	}
}

func TestPromoteDiffEntityNamesIncludeKinds(t *testing.T) {
	source := &firebase.RemoteConfig{
		Conditions: []firebase.RemoteConfigCondition{{
			Name:       "For Store",
			Expression: "store == 'source'",
		}},
		ParameterGroups: map[string]firebase.RemoteConfigGroup{
			"WEB": {
				Description: "Shopping cart",
				Parameters: map[string]firebase.RemoteConfigParam{
					"sc_public_term_of_use_url": stringParam("source"),
				},
			},
			"NPAW_PARAMETERS_ANDROID": {Description: "source description"},
		},
	}
	target := &firebase.RemoteConfig{
		Conditions: []firebase.RemoteConfigCondition{{
			Name:       "For Store",
			Expression: "store == 'target'",
		}},
		ParameterGroups: map[string]firebase.RemoteConfigGroup{
			"WEB": {
				Description: "Shopping cart",
				Parameters: map[string]firebase.RemoteConfigParam{
					"sc_public_term_of_use_url": stringParam("target"),
				},
			},
			"NPAW_PARAMETERS_ANDROID": {Description: "target description"},
		},
	}
	m := promoteTestModel(t, makePlan(t, source, target))
	names := make(map[rcdiff.ItemKind]string)
	for _, item := range m.plan.Plan.Items {
		input, ok := m.DiffInput(item)
		if ok {
			names[item.Kind] = input.EntityName
		}
	}
	want := map[rcdiff.ItemKind]string{
		rcdiff.ItemParameter:        "Property: WEB / sc_public_term_of_use_url",
		rcdiff.ItemGroupDescription: "Group: NPAW_PARAMETERS_ANDROID",
		rcdiff.ItemCondition:        "Condition: For Store",
	}
	for kind, name := range want {
		if names[kind] != name {
			t.Errorf("%s diff entity name = %q, want %q", kind, names[kind], name)
		}
	}
}

func assertTextOrder(t *testing.T, text, first, second string) {
	t.Helper()
	firstIndex := strings.Index(text, first)
	secondIndex := strings.Index(text, second)
	if firstIndex < 0 || secondIndex < 0 || firstIndex >= secondIndex {
		t.Fatalf("%q does not appear before %q:\n%s", first, second, text)
	}
}

func assertDetailTransition(t *testing.T, text, source, target string) {
	t.Helper()
	sourceIndex := strings.Index(text, source)
	targetIndex := strings.Index(text, target)
	if sourceIndex < 0 || targetIndex <= sourceIndex {
		t.Fatalf("cannot find ordered project sections %q and %q:\n%s", source, target, text)
	}
	between := text[sourceIndex+len(source) : targetIndex]
	const transition = "\n   🬞🬏\n   🭣🭘\n "
	if !strings.HasSuffix(between, transition) || strings.Contains(between, "\n\n") {
		t.Fatalf("project sections do not have a compact two-line transition:\n%s", text)
	}
}

func dependencyPlan(t *testing.T) *core.ProjectPromotionPlan {
	t.Helper()
	source := &firebase.RemoteConfig{
		Conditions: []firebase.RemoteConfigCondition{{Name: "beta", Expression: "true"}},
		Parameters: map[string]firebase.RemoteConfigParam{
			"flag": {DefaultValue: &firebase.RemoteConfigValue{Value: "off"}, ConditionalValues: map[string]firebase.RemoteConfigValue{"beta": {Value: "on"}}},
		},
	}
	target := &firebase.RemoteConfig{Parameters: map[string]firebase.RemoteConfigParam{}}
	return makePlan(t, source, target)
}

func makePlan(t *testing.T, source, target *firebase.RemoteConfig) *core.ProjectPromotionPlan {
	t.Helper()
	sourceRaw, err := firebase.MarshalRemoteConfig(source)
	if err != nil {
		t.Fatal(err)
	}
	targetRaw, err := firebase.MarshalRemoteConfig(target)
	if err != nil {
		t.Fatal(err)
	}
	return &core.ProjectPromotionPlan{
		Source: core.ProjectPromotionSnapshot{Project: core.Project{Name: "Development", ProjectID: "dev"}, Raw: sourceRaw, PublishedRaw: sourceRaw, Source: "cached"},
		Target: core.ProjectPromotionSnapshot{Project: core.Project{Name: "Production", ProjectID: "prod"}, Raw: targetRaw, PublishedRaw: targetRaw, Source: "cached"},
		Plan:   rcpromote.BuildPlan(source, target, rcpromote.Options{Prune: true}),
	}
}

func promoteTestModel(t *testing.T, plan *core.ProjectPromotionPlan) Model {
	t.Helper()
	svc, err := core.NewService(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	return New(svc).SetBounds(0, 0, 100, 24).SetPlan(plan, true)
}

func stringParam(value string) firebase.RemoteConfigParam {
	return firebase.RemoteConfigParam{DefaultValue: &firebase.RemoteConfigValue{Value: value}}
}

func key(code rune) tea.KeyPressMsg { return tea.KeyPressMsg(tea.Key{Code: code}) }
func keyRune(code rune) tea.KeyPressMsg {
	return tea.KeyPressMsg(tea.Key{Code: code, Text: string(code)})
}

func maxLineWidth(value string) int {
	maxWidth := 0
	for line := range strings.SplitSeq(value, "\n") {
		maxWidth = max(maxWidth, lipgloss.Width(line))
	}
	return maxWidth
}
