package parameters

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	rcdiff "github.com/yumauri/fbrcm/core/rc/diff"
	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/styles"
)

func TestHistoryRemovedParametersAreSortedAndDoNotLeakIntoParameters(t *testing.T) {
	project := core.Project{ProjectID: "demo", Name: "Demo"}
	current := historyTestTree("alpha", "charlie")
	previous := historyTestTree("alpha", "bravo", "charlie")
	m := New(nil)
	m.projects = []projectState{{project: project, tree: current}}
	m.projectIndex[project.ProjectID] = 0
	m.groupExpanded[m.groupKey(project.ProjectID, "group")] = true
	m.histories[project.ProjectID] = buildHistoryState(historyState{previous: previous, current: current})

	m = m.SetHistory(true)
	if got := visibleParameterKeys(m); !slices.Equal(got, []string{"alpha", "bravo", "charlie"}) {
		t.Fatalf("history parameters = %v", got)
	}
	for i, node := range m.visible {
		if node.kind == nodeParameter && node.paramKey == "bravo" {
			m.cursor = i
			break
		}
	}

	m = m.SetHistory(false)
	if got := visibleParameterKeys(m); !slices.Equal(got, []string{"alpha", "charlie"}) {
		t.Fatalf("parameters after history = %v", got)
	}
	if m.visible[m.cursor].kind != nodeParameter || m.visible[m.cursor].paramKey != "charlie" {
		t.Fatalf("cursor after removed row = %#v, want nearest following parameter", m.visible[m.cursor])
	}
}

func TestHistoryTabSwitchPreservesExistingParameterIdentity(t *testing.T) {
	project := core.Project{ProjectID: "demo", Name: "Demo"}
	current := historyTestTree("bravo", "charlie", "delta", "echo", "foxtrot", "golf", "hotel")
	previous := historyTestTree("alpha", "bravo", "charlie", "delta", "echo", "foxtrot", "golf", "hotel")
	m := New(nil).SetBounds(0, 0, 80, 7)
	m.projects = []projectState{{project: project, tree: current}}
	m.projectIndex[project.ProjectID] = 0
	m.groupExpanded[m.groupKey(project.ProjectID, "group")] = true
	m.histories[project.ProjectID] = buildHistoryState(historyState{previous: previous, current: current})
	m = m.SetHistory(true)
	for i, node := range m.visible {
		if node.kind == nodeParameter && node.paramKey == "echo" {
			m.cursor = i
			break
		}
	}

	m = m.SetHistory(false)
	if node := m.visible[m.cursor]; node.kind != nodeParameter || node.paramKey != "echo" {
		t.Fatalf("cursor after switch = %#v, want echo", node)
	}
}

func TestHistoryExpandAllIncludesRemovedParameters(t *testing.T) {
	project := core.Project{ProjectID: "demo", Name: "Demo"}
	current := historyTestTree("alpha")
	previous := historyTestTree("alpha", "removed")
	m := New(nil)
	m.projects = []projectState{{project: project, tree: current}}
	m.projectIndex[project.ProjectID] = 0
	m.groupExpanded[m.groupKey(project.ProjectID, "group")] = true
	m.histories[project.ProjectID] = buildHistoryState(historyState{previous: previous, current: current})
	m = m.SetHistory(true)
	key := m.paramKey(project.ProjectID, "group", "removed")

	m.setAllParametersExpanded(true)
	if !m.paramExpanded[key] {
		t.Fatal("removed parameter was not expanded")
	}
	m.setAllParametersExpanded(false)
	if m.paramExpanded[key] {
		t.Fatal("removed parameter was not collapsed")
	}
}

func TestHistoryLoadDefersVerificationAndPreservesInflightRequest(t *testing.T) {
	project := core.Project{ProjectID: "demo", Name: "Demo"}
	m := New(nil)
	m.history = true
	m.projects = []projectState{{project: project, tree: &core.ParametersTree{Version: "8"}, cacheVersion: "8", verifying: true}}
	m.projectIndex[project.ProjectID] = 0

	m, cmd := m.LoadHistory()
	if cmd != nil {
		t.Fatal("history load started while parameters were verifying")
	}
	if _, ok := m.histories[project.ProjectID]; ok {
		t.Fatal("history state created while parameters were verifying")
	}

	m.projects[0].verifying = false
	m, cmd = m.LoadHistory()
	if cmd == nil || !m.histories[project.ProjectID].loading {
		t.Fatal("history load did not start after verification")
	}
	m.invalidateHistoryIfVersionChanged(project.ProjectID)
	if !m.histories[project.ProjectID].loading {
		t.Fatal("in-flight history load was invalidated")
	}
}

func TestHistoryLoadMarksNAVersionUnavailableWithoutRequest(t *testing.T) {
	project := core.Project{ProjectID: "demo", Name: "Demo"}
	m := New(nil)
	m.history = true
	m.projects = []projectState{{project: project, tree: &core.ParametersTree{Version: "NA"}, cacheVersion: "NA"}}
	m.projectIndex[project.ProjectID] = 0

	m, cmd := m.LoadHistory()
	if cmd != nil {
		t.Fatal("NA version started a history request")
	}
	state := m.histories[project.ProjectID]
	if !state.unavailable || state.loading || state.err != nil {
		t.Fatalf("NA history state = %#v, want unavailable", state)
	}

	m.projects[0].tree.Version = "1"
	m.projects[0].cacheVersion = "1"
	m.invalidateHistoryIfVersionChanged(project.ProjectID)
	if _, ok := m.histories[project.ProjectID]; ok {
		t.Fatal("NA history state was not invalidated after a version became available")
	}
}

func TestHistoryProjectStatusIsRightAligned(t *testing.T) {
	project := core.Project{ProjectID: "mercato-mobile-9eac5", Name: "Mercato Mobile"}
	tree := &core.ParametersTree{Version: "NA"}
	tests := []struct {
		name  string
		state historyState
		label string
	}{
		{name: "unavailable", state: historyState{unavailable: true}, label: "history unavailable"},
		{name: "error", state: historyState{err: errors.New("failed")}, label: "history error"},
	}
	for _, width := range []int{38, 100} {
		for _, tt := range tests {
			t.Run(tt.name+"/width="+fmt.Sprint(width), func(t *testing.T) {
				m := New(nil).SetBounds(0, 0, width, 10)
				m.history = true
				m.projects = []projectState{{project: project, tree: tree}}
				m.projectIndex[project.ProjectID] = 0
				m.histories[project.ProjectID] = tt.state
				m.syncVisible()

				line := ansi.Strip(m.renderHistoryProjectNode(m.visible[0], &m.projects[0], false, false))
				if got, want := lipgloss.Width(line), m.viewportWidth(); got != want {
					t.Fatalf("line width = %d, want %d: %q", got, want, line)
				}
				if got, want := strings.Index(line, tt.label), m.viewportWidth()-lipgloss.Width(tt.label); got != want {
					t.Fatalf("status starts at %d, want %d: %q", got, want, line)
				}
				if width == 100 {
					identity := project.Name + " " + project.ProjectID
					if !strings.HasPrefix(line, identity) {
						t.Fatalf("project identity cropped from wide status row: %q", line)
					}
				}
			})
		}
	}
}

func TestSelectedHistoryProjectNameIsBold(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	project := core.Project{ProjectID: "demo", Name: "Demo"}
	tree := historyTestTree("parameter_name")
	tests := []struct {
		name  string
		state historyState
	}{
		{name: "unavailable", state: historyState{currentVersion: "NA", unavailable: true}},
		{name: "loaded", state: buildHistoryState(historyState{
			previous: tree, current: tree, previousVersion: "1", currentVersion: "2",
		})},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New(nil).SetBounds(0, 0, 100, 10)
			m.history = true
			m.projects = []projectState{{project: project, tree: tree}}
			m.projectIndex[project.ProjectID] = 0
			m.histories[project.ProjectID] = tt.state
			m.syncVisible()

			nameStyle, _ := projectIdentityStyles(true, false)
			if !nameStyle.GetBold() {
				t.Fatal("selected project-name style is not bold")
			}
			line := m.renderNode(m.visible[0], true)
			if renderedName := nameStyle.Render(project.Name); !strings.Contains(line, renderedName) {
				t.Fatalf("final selected history project row does not preserve bold name style: %q", line)
			}
		})
	}
}

func TestHistoryUnavailableHasNoTreePlaceholder(t *testing.T) {
	project := core.Project{ProjectID: "demo", Name: "Demo"}
	m := New(nil).SetBounds(0, 0, 80, 10)
	m.history = true
	m.projects = []projectState{{project: project, tree: &core.ParametersTree{Version: "NA"}}}
	m.projectIndex[project.ProjectID] = 0
	m.histories[project.ProjectID] = historyState{currentVersion: "NA", unavailable: true}
	m.syncVisible()

	if len(m.visible) != 1 || m.visible[0].kind != nodeProject {
		t.Fatalf("unavailable history nodes = %#v, want only the project row", m.visible)
	}
	view := ansi.Strip(m.View(false))
	for _, unwanted := range []string{"No parameters", "╰╌"} {
		if strings.Contains(view, unwanted) {
			t.Fatalf("unavailable history view contains dangling tree content %q: %q", unwanted, view)
		}
	}
}

func TestHistoryFirstOpenHasNoTreePlaceholderBeforeLoadStarts(t *testing.T) {
	project := core.Project{ProjectID: "demo", Name: "Demo"}
	m := New(nil).SetBounds(0, 0, 80, 10)
	m.projects = []projectState{{project: project, tree: &core.ParametersTree{Version: "NA"}}}
	m.projectIndex[project.ProjectID] = 0
	m = m.SetHistory(true)

	if _, ok := m.histories[project.ProjectID]; ok {
		t.Fatal("test setup unexpectedly created history state before loading")
	}
	if len(m.visible) != 1 || m.visible[0].kind != nodeProject {
		t.Fatalf("first-open history nodes = %#v, want only the project row", m.visible)
	}
	view := ansi.Strip(m.View(false))
	for _, unwanted := range []string{"No parameters", "╰╌"} {
		if strings.Contains(view, unwanted) {
			t.Fatalf("first-open history view contains dangling tree content %q: %q", unwanted, view)
		}
	}
}

func TestHistoryInvalidatesOnlyWhenPublishedVersionChanges(t *testing.T) {
	project := core.Project{ProjectID: "demo", Name: "Demo"}
	m := New(nil)
	m.projects = []projectState{{project: project, tree: &core.ParametersTree{Version: "8"}, cacheVersion: "8"}}
	m.projectIndex[project.ProjectID] = 0
	m.histories[project.ProjectID] = buildHistoryState(historyState{currentVersion: "8", current: &core.ParametersTree{Version: "8"}, versions: []core.RemoteConfigVersionEntry{{RemoteConfigVersion: firebase.RemoteConfigVersion{VersionNumber: "8"}}}})

	m.invalidateHistoryIfVersionChanged(project.ProjectID)
	if _, ok := m.histories[project.ProjectID]; !ok {
		t.Fatal("same-version parameter load invalidated history")
	}
	m.projects[0].cacheVersion = "9"
	m.invalidateHistoryIfVersionChanged(project.ProjectID)
	if _, ok := m.histories[project.ProjectID]; ok {
		t.Fatal("new published version did not invalidate history")
	}
}

func TestHistoryVersionSteppingUsesCachedPairs(t *testing.T) {
	project := core.Project{ProjectID: "demo", Name: "Demo"}
	versions := []core.RemoteConfigVersionEntry{
		{RemoteConfigVersion: firebase.RemoteConfigVersion{VersionNumber: "3"}},
		{RemoteConfigVersion: firebase.RemoteConfigVersion{VersionNumber: "2"}},
		{RemoteConfigVersion: firebase.RemoteConfigVersion{VersionNumber: "1"}},
	}
	pair := func(left, right string) historyPairData {
		return historyPairData{previous: historyTestTree("p"), current: historyTestTree("p"), previousVersion: left, currentVersion: right}
	}
	state := buildHistoryState(historyState{previous: historyTestTree("p"), current: historyTestTree("p"), previousVersion: "2", currentVersion: "3", versions: versions,
		pairs: map[string]historyPairData{historyPairKey("1", "3"): pair("1", "3"), historyPairKey("2", "2"): pair("2", "2"), historyPairKey("1", "2"): pair("1", "2")}})
	m := New(nil)
	m.history = true
	m.projects = []projectState{{project: project, tree: state.current}}
	m.projectIndex[project.ProjectID] = 0
	m.histories[project.ProjectID] = state
	m.syncVisible()

	m, cmd, _ := m.stepHistory("both-older")
	if cmd != nil || m.histories[project.ProjectID].previousVersion != "1" || m.histories[project.ProjectID].currentVersion != "2" {
		t.Fatal("both older did not preserve distance")
	}
}

func TestHistoryVersionKeysTargetCursorProject(t *testing.T) {
	versions := []core.RemoteConfigVersionEntry{
		{RemoteConfigVersion: firebase.RemoteConfigVersion{VersionNumber: "3"}},
		{RemoteConfigVersion: firebase.RemoteConfigVersion{VersionNumber: "2"}},
		{RemoteConfigVersion: firebase.RemoteConfigVersion{VersionNumber: "1"}},
	}
	makeState := func() historyState {
		state := buildHistoryState(historyState{previous: historyTestTree("p"), current: historyTestTree("p"), previousVersion: "2", currentVersion: "3", versions: versions})
		state.pairs = map[string]historyPairData{historyPairKey("1", "2"): {previous: state.previous, current: state.current, previousVersion: "1", currentVersion: "2"}}
		return state
	}
	first := core.Project{ProjectID: "first", Name: "First"}
	second := core.Project{ProjectID: "second", Name: "Second"}
	m := New(nil)
	m.history = true
	m.projects = []projectState{{project: first, tree: historyTestTree("p")}, {project: second, tree: historyTestTree("p")}}
	m.projectIndex[first.ProjectID], m.projectIndex[second.ProjectID] = 0, 1
	m.histories[first.ProjectID], m.histories[second.ProjectID] = makeState(), makeState()
	m.groupExpanded[m.groupKey(first.ProjectID, "group")] = true
	m.groupExpanded[m.groupKey(second.ProjectID, "group")] = true
	m.syncVisible()
	for i, node := range m.visible {
		if node.kind == nodeParameter && node.projectID == second.ProjectID {
			m.cursor = i
			break
		}
	}

	m, cmd, _ := m.stepHistory("both-older")
	if cmd != nil {
		t.Fatal("expected cached second-project pair")
	}
	if got := m.histories[first.ProjectID].previousVersion; got != "2" {
		t.Fatalf("first project changed to %s", got)
	}
	if got := m.histories[second.ProjectID].previousVersion; got != "1" {
		t.Fatalf("second project stayed at %s", got)
	}
	m.openHistoryPicker()
	if m.versionPicker == nil || m.versionPicker.projectID != second.ProjectID {
		t.Fatal("picker did not bind to cursor project")
	}
}

func TestHistoryVersionPickerRendersAuthorsAlignedInsideCompleteBorder(t *testing.T) {
	project := core.Project{ProjectID: "demo", Name: "Demo"}
	versions := []core.RemoteConfigVersionEntry{
		{RemoteConfigVersion: firebase.RemoteConfigVersion{VersionNumber: "3", UpdateTime: "2026-07-13T12:34:56Z", UpdateUser: firebase.RemoteConfigUser{Email: "alice@example.com"}}},
		{RemoteConfigVersion: firebase.RemoteConfigVersion{VersionNumber: "2", UpdateTime: "2026-07-12T11:22:33Z", UpdateUser: firebase.RemoteConfigUser{Name: "Bob Builder"}}},
		{RemoteConfigVersion: firebase.RemoteConfigVersion{VersionNumber: "1", UpdateTime: "2026-07-11T10:20:30Z"}},
	}
	m := New(nil).SetBounds(0, 0, 90, 18)
	m.projects = []projectState{{project: project, tree: historyTestTree("p")}}
	m.projectIndex[project.ProjectID] = 0
	m.histories[project.ProjectID] = buildHistoryState(historyState{
		previous: historyTestTree("p"), current: historyTestTree("p"),
		previousVersion: "1", currentVersion: "2", versions: versions,
	})
	m.versionPicker = &historyVersionPicker{projectID: project.ProjectID, left: true, leftCursor: 2, rightCursor: 1}

	view := m.HistoryPickerView()
	plain := ansi.Strip(view)
	leftArrow, rightArrow := historyPickerArrowGlyphs(true)
	if !strings.Contains(plain, "alice@example.com") || !strings.Contains(plain, "Bob Builder") {
		t.Fatalf("picker does not show version authors:\n%s", plain)
	}
	if strings.Contains(plain, "Status") {
		t.Fatalf("picker still contains status column:\n%s", plain)
	}
	if strings.Contains(plain, "Pick") {
		t.Fatalf("picker still contains Pick header:\n%s", plain)
	}
	if !strings.Contains(plain, leftArrow+" v1") || !strings.Contains(plain, "v2 "+rightArrow) {
		t.Fatalf("picker does not mark chosen versions with directional arrows:\n%s", plain)
	}
	lines := strings.Split(view, "\n")
	w, h := m.historyPickerSize()
	if len(lines) != h {
		t.Fatalf("picker height = %d, want %d", len(lines), h)
	}
	for i, line := range lines {
		if got := lipgloss.Width(line); got != w {
			t.Fatalf("line %d width = %d, want %d: %q", i, got, w, ansi.Strip(line))
		}
	}
	if tabs, bodyTop, last := ansi.Strip(lines[0]), ansi.Strip(lines[2]), ansi.Strip(lines[len(lines)-1]); !strings.Contains(tabs, "╭") || !strings.Contains(tabs, "╮") || (!strings.HasSuffix(bodyTop, "╮") && !strings.HasSuffix(bodyTop, "┤") && !strings.HasSuffix(bodyTop, "│")) || !strings.HasPrefix(last, "╰") || !strings.HasSuffix(last, "╯") {
		t.Fatalf("broken popup border:\n%s", plain)
	}
	geometry := m.historyPickerGeometry()
	columns := m.historyColumnLayout()
	leftLabel := historyPickerVersionLabel(versions, 2)
	rightLabel := historyPickerVersionLabel(versions, 1)
	contentLine := ansi.Strip(lines[1])
	leftAt := strings.Index(contentLine, leftLabel)
	rightAt := strings.Index(contentLine, rightLabel)
	if got, want := geometry.x+lipgloss.Width(contentLine[:leftAt]), m.x+1+columns.leftStart+2; got != want {
		t.Fatalf("left tab text x = %d, want project-row x %d; geometry=%+v content=%q", got, want, geometry, contentLine)
	}
	if got, want := geometry.x+lipgloss.Width(contentLine[:rightAt]), m.x+1+columns.rightStart+2; got != want {
		t.Fatalf("right tab text x = %d, want project-row x %d", got, want)
	}
	if !strings.Contains(view, historyPickerActiveTabStyle().Render(leftLabel)) {
		t.Fatal("left active tab does not use active title style")
	}
	if strings.Contains(view, styles.PanelTitleActive.Render(leftLabel)) {
		t.Fatal("left active tab still uses background highlight")
	}
	if !strings.Contains(view, styles.PanelTitleInactiveTab.Render(rightLabel)) {
		t.Fatal("right inactive tab does not use inactive text style")
	}
	plainTabTops, plainTabContent, plainBodyTop := ansi.Strip(lines[0]), ansi.Strip(lines[1]), ansi.Strip(lines[2])
	if !strings.Contains(plainTabTops, "╭"+strings.Repeat("─", historyPickerTabWidth(leftLabel)-2)+"╮") || !strings.Contains(plainTabContent, "│ "+leftLabel+" │") || !strings.Contains(plainTabContent, "│ "+rightLabel+" │") {
		t.Fatalf("tab content is not fully bordered:\n%s\n%s", plainTabTops, plainTabContent)
	}
	betweenTabs := ansi.Cut(lines[0], geometry.leftTab+historyPickerTabWidth(leftLabel), geometry.rightTab)
	if strings.Trim(ansi.Strip(betweenTabs), "─") != "" {
		t.Fatalf("underlying panel border was not preserved between tabs: %q", ansi.Strip(betweenTabs))
	}
	if strings.Contains(plainBodyTop, "╯╮") {
		t.Fatalf("popup has adjacent, unjoined border corners:\n%s", plainBodyTop)
	}
	rightTabWidth := historyPickerTabWidth(rightLabel)
	if !strings.Contains(view, historyPickerTabBottom(rightTabWidth, false, geometry.rightTab == 0, geometry.rightTab+rightTabWidth == geometry.width)) {
		t.Fatal("right inactive tab does not have a bottom border")
	}

	if got := strings.Trim(ansi.Strip(lines[3]), "│ "); got != "" {
		t.Fatalf("history picker top padding row = %q, want blank", got)
	}
	header := ansi.Strip(lines[4])
	row := ansi.Strip(lines[5])
	for column, value := range map[string]string{"Published": "2026-07-13", "Author": "alice@example.com"} {
		if strings.Index(header, column) != strings.Index(row, value) {
			t.Fatalf("%s column is not aligned:\n%s\n%s", column, header, row)
		}
	}
	if strings.Count(row, "v3") != 2 {
		t.Fatalf("picker row does not show the version on both sides:\n%s", row)
	}
}

func TestHistoryVersionPickerHintUsesApplicationHelpStyles(t *testing.T) {
	hint := historyPickerHintView(100)
	if !strings.Contains(hint, styles.FilterText.Render("left/right")) {
		t.Fatal("picker hint does not use application key style")
	}
	if !strings.Contains(hint, styles.FilterText.Render(",/.")) {
		t.Fatal("picker hint does not show pair movement bindings")
	}
	if !strings.Contains(hint, styles.PanelMuted.Render(" side")) || !strings.Contains(hint, styles.PanelMuted.Render(" • ")) {
		t.Fatal("picker hint does not use application description and separator styles")
	}
}

func TestHistoryVersionPickerSelectionExcludesArrowColumns(t *testing.T) {
	row := pickerVersionTableRow("v7", "2025-04-11 13:34:34", "author@example.com", "v7", 5, 19, 18)
	selected := historyPickerSelectedRow(row, true)
	width := lipgloss.Width(row)
	left := ansi.Cut(row, 0, historyPickerArrowWidth)
	middle := ansi.Cut(row, historyPickerArrowWidth, width-historyPickerArrowWidth)
	right := ansi.Cut(row, width-historyPickerArrowWidth, width)
	want := left + historyPickerSelectionStyle(true).Render(middle) + right
	if selected != want {
		t.Fatal("selection is not limited to the table body between arrow columns")
	}
	leftArrow, rightArrow := historyPickerArrowGlyphs(true)
	if lipgloss.Width(leftArrow) != historyPickerArrowWidth || lipgloss.Width(rightArrow) != historyPickerArrowWidth {
		t.Fatal("Powerline arrow glyph width does not match its reserved column")
	}
	fallbackLeft, fallbackRight := historyPickerArrowGlyphs(false)
	if fallbackLeft != historyPickerFallbackLeftArrow || fallbackRight != historyPickerFallbackRightArrow {
		t.Fatal("standard Unicode arrow fallback is not selected when Powerline glyphs are disabled")
	}
}

func TestHistoryVersionPickerKeepsInactiveSideSelected(t *testing.T) {
	row := pickerVersionTableRow("v7", "2025-04-11 13:34:34", "author@example.com", "v7", 5, 19, 18)
	active := historyPickerSelectedRow(row, true)
	inactive := historyPickerSelectedRow(row, false)
	if active == inactive {
		t.Fatal("active and inactive side selections use the same rendering")
	}

	leftInactive := historyPickerRowArrows(inactive, true, false, false, true)
	leftArrow, rightArrow := historyPickerArrowGlyphs(true)
	if !strings.Contains(leftInactive, historyPickerArrowStyle(false).Render(leftArrow)) {
		t.Fatal("inactive left selection does not use a muted arrow")
	}
	rightActive := historyPickerRowArrows(active, false, true, false, true)
	if !strings.Contains(rightActive, historyPickerArrowStyle(true).Render(rightArrow)) {
		t.Fatal("active right selection does not use an active arrow")
	}
}

func TestHistoryVersionPickerBoundsDisableCrossedVersions(t *testing.T) {
	versions := []core.RemoteConfigVersionEntry{
		{RemoteConfigVersion: firebase.RemoteConfigVersion{VersionNumber: "3"}},
		{RemoteConfigVersion: firebase.RemoteConfigVersion{VersionNumber: "2"}},
		{RemoteConfigVersion: firebase.RemoteConfigVersion{VersionNumber: "1"}},
	}
	if low, high := historyPickerBounds(len(versions), 2, 1, true); low != 1 || high != 2 {
		t.Fatalf("left picker bounds = %d..%d, want 1..2", low, high)
	}
	if low, high := historyPickerBounds(len(versions), 2, 1, false); low != 0 || high != 2 {
		t.Fatalf("right picker bounds = %d..%d, want 0..2", low, high)
	}
}

func TestHistoryVersionPickerStagesBothSidesUntilEnter(t *testing.T) {
	project := core.Project{ProjectID: "demo", Name: "Demo"}
	versions := []core.RemoteConfigVersionEntry{
		{RemoteConfigVersion: firebase.RemoteConfigVersion{VersionNumber: "3"}},
		{RemoteConfigVersion: firebase.RemoteConfigVersion{VersionNumber: "2"}},
		{RemoteConfigVersion: firebase.RemoteConfigVersion{VersionNumber: "1"}},
	}
	state := buildHistoryState(historyState{
		previous: historyTestTree("p"), current: historyTestTree("p"),
		previousVersion: "2", currentVersion: "3", versions: versions,
	})
	state.pairs = map[string]historyPairData{
		historyPairKey("1", "2"): {
			previous: state.previous, current: state.current,
			previousVersion: "1", currentVersion: "2",
		},
	}
	m := New(nil)
	m.history = true
	m.projects = []projectState{{project: project, tree: state.current}}
	m.projectIndex[project.ProjectID] = 0
	m.histories[project.ProjectID] = state
	m.syncVisible()
	m.openHistoryPicker()

	m, _, _ = m.updateHistoryPickerKey(",")
	if got := m.histories[project.ProjectID]; got.previousVersion != "2" || got.currentVersion != "3" {
		t.Fatalf("picker changed pair before apply: %s/%s", got.previousVersion, got.currentVersion)
	}
	if !m.versionPicker.left || m.versionPicker.leftCursor != 2 || m.versionPicker.rightCursor != 1 {
		t.Fatalf("picker did not stage both cursors: %#v", m.versionPicker)
	}
	m, _, _ = m.updateHistoryPickerKey(",")
	if m.versionPicker.leftCursor != 2 || m.versionPicker.rightCursor != 1 {
		t.Fatalf("picker moved staged pair past older bound: %#v", m.versionPicker)
	}
	m, _, _ = m.updateHistoryPickerKey(".")
	if m.versionPicker.leftCursor != 1 || m.versionPicker.rightCursor != 0 {
		t.Fatalf("picker did not stage both cursors newer: %#v", m.versionPicker)
	}
	m, _, _ = m.updateHistoryPickerKey(",")

	m, cmd, _ := m.updateHistoryPickerKey("enter")
	if cmd != nil {
		t.Fatal("expected staged pair to use cache")
	}
	if m.versionPicker != nil {
		t.Fatal("picker remained open after apply")
	}
	if got := m.histories[project.ProjectID]; got.previousVersion != "1" || got.currentVersion != "2" {
		t.Fatalf("applied pair = %s/%s, want 1/2", got.previousVersion, got.currentVersion)
	}
}

func TestHistoryVersionPickerRequestsRollbackForActiveHistoricalVersion(t *testing.T) {
	project := core.Project{ProjectID: "demo", Name: "Demo"}
	versions := []core.RemoteConfigVersionEntry{
		{RemoteConfigVersion: firebase.RemoteConfigVersion{VersionNumber: "3"}, Current: true},
		{RemoteConfigVersion: firebase.RemoteConfigVersion{VersionNumber: "2"}},
		{RemoteConfigVersion: firebase.RemoteConfigVersion{VersionNumber: "1"}},
	}
	state := buildHistoryState(historyState{
		previous: historyTestTree("p"), current: historyTestTree("p"),
		previousVersion: "2", currentVersion: "3", versions: versions,
	})
	m := New(nil)
	m.history = true
	m.projects = []projectState{{project: project, tree: state.current}}
	m.projectIndex[project.ProjectID] = 0
	m.histories[project.ProjectID] = state
	m.syncVisible()
	m.openHistoryPicker()

	next, cmd, handled := m.updateHistoryPickerKey("R")
	if !handled || cmd == nil {
		t.Fatalf("rollback key handled=%v cmd=%v", handled, cmd)
	}
	if next.versionPicker != nil {
		t.Fatal("picker remained open while rollback confirmation is requested")
	}
	request, ok := cmd().(messages.HistoryRollbackRequestedMsg)
	if !ok {
		t.Fatalf("rollback command emitted %T", cmd())
	}
	if request.Project.ProjectID != project.ProjectID || request.Target.VersionNumber != "2" {
		t.Fatalf("rollback request = %#v", request)
	}
	if !request.PickerLeft || request.LeftCursor != 1 || request.RightCursor != 0 {
		t.Fatalf("picker snapshot = left:%v cursors:%d/%d", request.PickerLeft, request.LeftCursor, request.RightCursor)
	}

	next = next.RestoreHistoryPicker(project.ProjectID, request.PickerLeft, request.LeftCursor, request.RightCursor)
	if next.versionPicker == nil || !next.versionPicker.left {
		t.Fatal("cancel did not restore the original chooser side and selection")
	}
	next.versionPicker.left = false
	stillOpen, cmd, handled := next.updateHistoryPickerKey("R")
	if !handled || cmd != nil || stillOpen.versionPicker == nil {
		t.Fatal("current version should be a disabled rollback target")
	}
}

func TestHistoryVersionPickerArrowSidesAndResetAppliesDefaultPair(t *testing.T) {
	project := core.Project{ProjectID: "demo", Name: "Demo"}
	versions := []core.RemoteConfigVersionEntry{
		{RemoteConfigVersion: firebase.RemoteConfigVersion{VersionNumber: "3"}, Current: true},
		{RemoteConfigVersion: firebase.RemoteConfigVersion{VersionNumber: "2"}},
		{RemoteConfigVersion: firebase.RemoteConfigVersion{VersionNumber: "1"}},
	}
	state := buildHistoryState(historyState{
		previous: historyTestTree("p"), current: historyTestTree("p"),
		previousVersion: "1", currentVersion: "2", versions: versions,
	})
	state.pairs = map[string]historyPairData{
		historyPairKey("2", "3"): {
			previous: state.previous, current: state.current,
			previousVersion: "2", currentVersion: "3",
		},
	}
	m := New(nil).SetBounds(0, 0, 90, 18)
	m.history = true
	m.projects = []projectState{{project: project, tree: state.current}}
	m.projectIndex[project.ProjectID] = 0
	m.histories[project.ProjectID] = state
	m.syncVisible()
	m.openHistoryPicker()

	m, _, _ = m.updateHistoryPickerKey("right")
	if m.versionPicker.left {
		t.Fatal("RightArrow did not activate right selection")
	}
	rightView := m.HistoryPickerView()
	leftLabel := historyPickerVersionLabel(versions, m.versionPicker.leftCursor)
	rightLabel := historyPickerVersionLabel(versions, m.versionPicker.rightCursor)
	if !strings.Contains(rightView, historyPickerActiveTabStyle().Render(rightLabel)) {
		t.Fatal("RightArrow did not activate right popup tab")
	}
	if !strings.Contains(rightView, styles.PanelTitleInactiveTab.Render(leftLabel)) {
		t.Fatal("RightArrow did not give left popup tab its inactive style")
	}
	geometry := m.historyPickerGeometry()
	leftTabWidth := historyPickerTabWidth(leftLabel)
	if !strings.Contains(rightView, historyPickerTabBottom(leftTabWidth, false, geometry.leftTab == 0, geometry.leftTab+leftTabWidth == geometry.width)) {
		t.Fatal("RightArrow did not give left popup tab its bottom border")
	}
	m, _, _ = m.updateHistoryPickerKey("left")
	if !m.versionPicker.left {
		t.Fatal("LeftArrow did not activate left selection")
	}
	m, cmd, _ := m.updateHistoryPickerKey("r")
	if cmd != nil {
		t.Fatal("expected reset pair to use cache")
	}
	if m.versionPicker != nil {
		t.Fatal("picker remained open after reset")
	}
	if got := m.histories[project.ProjectID]; got.previousVersion != "2" || got.currentVersion != "3" {
		t.Fatalf("reset pair = %s/%s, want previous/current 2/3", got.previousVersion, got.currentVersion)
	}
}

func TestHistoryVersionPickerScrollsProjectRowToFitMinimumHeight(t *testing.T) {
	first := core.Project{ProjectID: "first", Name: "First"}
	second := core.Project{ProjectID: "second", Name: "Second"}
	versions := []core.RemoteConfigVersionEntry{
		{RemoteConfigVersion: firebase.RemoteConfigVersion{VersionNumber: "3"}, Current: true},
		{RemoteConfigVersion: firebase.RemoteConfigVersion{VersionNumber: "2"}},
	}
	makeState := func(tree *core.ParametersTree) historyState {
		return buildHistoryState(historyState{
			previous: tree, current: tree, previousVersion: "2", currentVersion: "3", versions: versions,
		})
	}
	firstTree := historyTestTree("a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l")
	secondTree := historyTestTree("only")
	m := New(nil).SetBounds(17, 3, 100, 12)
	m.history = true
	m.projects = []projectState{{project: first, tree: firstTree}, {project: second, tree: secondTree}}
	m.projectIndex[first.ProjectID], m.projectIndex[second.ProjectID] = 0, 1
	m.histories[first.ProjectID], m.histories[second.ProjectID] = makeState(firstTree), makeState(secondTree)
	m.groupExpanded[m.groupKey(first.ProjectID, "group")] = true
	m.groupExpanded[m.groupKey(second.ProjectID, "group")] = true
	m.syncVisible()
	projectIndex := -1
	for i, node := range m.visible {
		if node.kind == nodeProject && node.projectID == second.ProjectID {
			projectIndex = i
			m.cursor = i
			break
		}
	}
	if projectIndex < 0 {
		t.Fatal("second project row not found")
	}
	m.ensureCursorVisible()
	before := m.screenLineForOffset(projectIndex, m.offset)
	m.openHistoryPicker()
	after := m.screenLineForOffset(projectIndex, m.offset)
	minimumHeight := min(historyPickerMinHeight, m.viewportHeight())
	if after+minimumHeight > m.viewportHeight() {
		t.Fatalf("project row %d leaves less than %d lines in %d-line viewport", after, minimumHeight, m.viewportHeight())
	}
	if before+minimumHeight > m.viewportHeight() && after >= before {
		t.Fatalf("picker did not scroll project row upward: before=%d after=%d", before, after)
	}
	geometry := m.historyPickerGeometry()
	if want := m.y + after; geometry.y != want {
		t.Fatalf("picker y = %d, want one row above selected project row %d", geometry.y, want)
	}
	if tabContentY, projectRowY := geometry.y+1, m.y+1+after; tabContentY != projectRowY {
		t.Fatalf("tab content y = %d, want project-row y %d", tabContentY, projectRowY)
	}
}

func TestHistoryVersionPickerKeepsBothTabsWhenVersionAnchorsAreClose(t *testing.T) {
	project := core.Project{ProjectID: "demo", Name: "Demo"}
	versions := []core.RemoteConfigVersionEntry{
		{RemoteConfigVersion: firebase.RemoteConfigVersion{VersionNumber: "12", UpdateTime: "2026-07-13T12:34:56Z"}, Current: true},
		{RemoteConfigVersion: firebase.RemoteConfigVersion{VersionNumber: "11", UpdateTime: "2026-07-12T11:22:33Z"}},
	}
	tree := historyTestTree(strings.Repeat("long_parameter_", 5))
	m := New(nil).SetBounds(0, 0, 70, 18)
	m.history = true
	m.projects = []projectState{{project: project, tree: tree}}
	m.projectIndex[project.ProjectID] = 0
	m.histories[project.ProjectID] = buildHistoryState(historyState{
		previous: tree, current: tree, previousVersion: "11", currentVersion: "12", versions: versions,
	})
	m.syncVisible()
	m.openHistoryPicker()
	if !m.historyStacked() {
		t.Fatal("test setup did not produce close stacked version anchors")
	}

	top := ansi.Strip(strings.Split(m.HistoryPickerView(), "\n")[1])
	for _, label := range []string{historyPickerVersionLabel(versions, 1), historyPickerVersionLabel(versions, 0)} {
		if !strings.Contains(top, label) {
			t.Fatalf("close version tabs cropped %q from %q", label, top)
		}
	}
}

func TestHistoryClassifiesCollapsedValuesIndependently(t *testing.T) {
	previous := &core.ParametersTree{Groups: []core.ParametersGroup{{Key: "group", Parameters: []core.ParametersEntry{{Key: "flag", Values: []core.ParametersValue{{Label: "first", Value: "old"}, {Label: "second", Value: "same"}}}}}}}
	current := &core.ParametersTree{Groups: []core.ParametersGroup{{Key: "group", Parameters: []core.ParametersEntry{{Key: "flag", Values: []core.ParametersValue{{Label: "first", Value: "new"}, {Label: "second", Value: "same"}}}}}}}
	state := buildHistoryState(historyState{previous: previous, current: current})
	m := New(nil)
	m.histories["demo"] = state
	if got := m.historyValueKind("demo", "group", "flag", "first"); got != rcdiff.ChangeChanged {
		t.Fatalf("first kind = %s", got)
	}
	if got := m.historyValueKind("demo", "group", "flag", "second"); got != rcdiff.ChangeUnchanged {
		t.Fatalf("second kind = %s", got)
	}
}

func TestHistoryChangesOnlyKeepsChangedParametersAndChangedExpandedValues(t *testing.T) {
	project := core.Project{ProjectID: "demo", Name: "Demo"}
	previous := historyTreeWithParameters(
		core.ParametersEntry{Key: "changed", Values: []core.ParametersValue{{Label: "first", Value: "old"}, {Label: "stable", Value: "same"}}},
		core.ParametersEntry{Key: "description", Summary: "old description", Values: []core.ParametersValue{{Label: "default", Value: "same"}}},
		core.ParametersEntry{Key: "removed", Values: []core.ParametersValue{{Label: "default", Value: "old"}}},
		core.ParametersEntry{Key: "unchanged", Values: []core.ParametersValue{{Label: "default", Value: "same"}}},
	)
	current := historyTreeWithParameters(
		core.ParametersEntry{Key: "added", Values: []core.ParametersValue{{Label: "default", Value: "new"}}},
		core.ParametersEntry{Key: "changed", Values: []core.ParametersValue{{Label: "first", Value: "new"}, {Label: "stable", Value: "same"}}},
		core.ParametersEntry{Key: "description", Summary: "new description", Values: []core.ParametersValue{{Label: "default", Value: "same"}}},
		core.ParametersEntry{Key: "unchanged", Values: []core.ParametersValue{{Label: "default", Value: "same"}}},
	)
	m := New(nil)
	m.history = true
	m.projects = []projectState{{project: project, tree: current}}
	m.projectIndex[project.ProjectID] = 0
	m.histories[project.ProjectID] = buildHistoryState(historyState{previous: previous, current: current})
	m.groupExpanded[m.groupKey(project.ProjectID, "group")] = true
	for _, key := range []string{"added", "changed", "description", "removed", "unchanged"} {
		m.paramExpanded[m.paramKey(project.ProjectID, "group", key)] = true
	}
	m.syncVisible()
	m = m.toggleHistoryChangesOnly()

	if got, want := visibleParameterKeys(m), []string{"added", "changed", "description", "removed"}; !slices.Equal(got, want) {
		t.Fatalf("changes-only parameters = %v, want %v", got, want)
	}
	values := make(map[string][]string)
	for _, node := range m.visible {
		if node.kind == nodeValue {
			values[node.paramKey] = append(values[node.paramKey], node.label)
		}
	}
	if got := values["changed"]; !slices.Equal(got, []string{"first"}) {
		t.Fatalf("changed parameter values = %v, want only changed value", got)
	}
	if got := values["description"]; len(got) != 0 {
		t.Fatalf("description-only change retained unchanged values: %v", got)
	}
	if !slices.Equal(values["added"], []string{"default"}) || !slices.Equal(values["removed"], []string{"default"}) {
		t.Fatalf("added/removed expanded values = added:%v removed:%v", values["added"], values["removed"])
	}
}

func TestHistoryChangesOnlyKeepsCollapsedChangedGroupExpandable(t *testing.T) {
	project := core.Project{ProjectID: "demo", Name: "Demo"}
	previous := historyTreeWithParameters(core.ParametersEntry{
		Key: "changed", Values: []core.ParametersValue{{Label: "default", Value: "old"}},
	})
	current := historyTreeWithParameters(core.ParametersEntry{
		Key: "changed", Values: []core.ParametersValue{{Label: "default", Value: "new"}},
	})
	m := New(nil)
	m.history = true
	m.projects = []projectState{{project: project, tree: current}}
	m.projectIndex[project.ProjectID] = 0
	m.histories[project.ProjectID] = buildHistoryState(historyState{previous: previous, current: current})
	m.groupExpanded[m.groupKey(project.ProjectID, "group")] = true
	m.syncVisible()
	m = m.toggleHistoryChangesOnly()

	for i, node := range m.visible {
		if node.kind == nodeGroup {
			m.cursor = i
			break
		}
	}
	m.collapseCurrent()
	if len(m.visible) != 2 || m.visible[1].kind != nodeGroup || m.visible[1].expanded {
		t.Fatalf("collapsed changed group disappeared or remained expanded: %#v", m.visible)
	}

	m.expandCurrent()
	if got := visibleParameterKeys(m); !slices.Equal(got, []string{"changed"}) {
		t.Fatalf("re-expanded changed group parameters = %v, want [changed]", got)
	}
}

func TestHistoryChangesOnlyPreservesIndependentSelections(t *testing.T) {
	project := core.Project{ProjectID: "demo", Name: "Demo"}
	previous := historyTreeWithParameters(
		core.ParametersEntry{Key: "changed", Values: []core.ParametersValue{{Label: "default", Value: "old"}}},
		core.ParametersEntry{Key: "unchanged", Values: []core.ParametersValue{{Label: "default", Value: "same"}}},
	)
	current := historyTreeWithParameters(
		core.ParametersEntry{Key: "changed", Values: []core.ParametersValue{{Label: "default", Value: "new"}}},
		core.ParametersEntry{Key: "unchanged", Values: []core.ParametersValue{{Label: "default", Value: "same"}}},
	)
	m := New(nil).SetBounds(0, 0, 100, 12)
	m.history = true
	m.projects = []projectState{{project: project, tree: current}}
	m.projectIndex[project.ProjectID] = 0
	m.histories[project.ProjectID] = buildHistoryState(historyState{previous: previous, current: current})
	m.groupExpanded[m.groupKey(project.ProjectID, "group")] = true
	m.syncVisible()
	for i, node := range m.visible {
		if node.kind == nodeParameter && node.paramKey == "unchanged" {
			m.cursor = i
		}
	}

	m = m.toggleHistoryChangesOnly()
	if node := m.visible[m.cursor]; node.kind != nodeParameter || node.paramKey != "changed" {
		t.Fatalf("nearest changes-only selection = %#v", node)
	}
	m = m.toggleHistoryChangesOnly()
	if node := m.visible[m.cursor]; node.kind != nodeParameter || node.paramKey != "unchanged" {
		t.Fatalf("all-mode selection was not restored: %#v", node)
	}
	m = m.toggleHistoryChangesOnly()
	if node := m.visible[m.cursor]; node.kind != nodeParameter || node.paramKey != "changed" {
		t.Fatalf("changes-only selection was not restored: %#v", node)
	}
}

func TestHistoryChangesOnlyRendersResponsiveBorderStatusAndEmptyState(t *testing.T) {
	project := core.Project{ProjectID: "demo", Name: "Demo"}
	previous := historyTreeWithParameters(core.ParametersEntry{Key: "same"})
	current := historyTreeWithParameters(core.ParametersEntry{Key: "same"})
	versions := []core.RemoteConfigVersionEntry{
		{RemoteConfigVersion: firebase.RemoteConfigVersion{VersionNumber: "2"}, Current: true},
		{RemoteConfigVersion: firebase.RemoteConfigVersion{VersionNumber: "1"}},
	}
	m := New(nil).SetBounds(0, 0, 120, 12)
	m.history = true
	m.projects = []projectState{{project: project, tree: current}}
	m.projectIndex[project.ProjectID] = 0
	m.histories[project.ProjectID] = buildHistoryState(historyState{
		previous: previous, current: current, previousVersion: "1", currentVersion: "2", versions: versions,
	})
	m.groupExpanded[m.groupKey(project.ProjectID, "group")] = true
	m.syncVisible()
	m = m.toggleHistoryChangesOnly()

	view := m.ViewWithBorder(true, true)
	plain := ansi.Strip(view)
	if !strings.Contains(strings.Split(plain, "\n")[0], "Δ 0 changed · 0 added · 0 removed") {
		t.Fatalf("wide history border lacks changes mode and counts:\n%s", plain)
	}
	if !strings.Contains(plain, "Demo") || strings.Contains(plain, "No changes between the selected versions.") {
		t.Fatalf("changes-only mode did not retain an empty project row:\n%s", plain)
	}
	if len(m.visible) != 1 || m.visible[0].kind != nodeProject || m.visible[0].projectID != project.ProjectID {
		t.Fatalf("changes-only nodes = %#v, want only project row", m.visible)
	}
	m, cmd, handled := m.updateHistoryKey(tea.KeyPressMsg{}, "v")
	if !handled || cmd != nil || m.versionPicker == nil || m.versionPicker.projectID != project.ProjectID {
		t.Fatal("version chooser did not open from an empty changes-only project row")
	}
	inactive := m.ViewWithBorder(false, false)
	if !strings.Contains(inactive, styles.PanelTitleInactiveTab.Render(" Δ 0 changed · 0 added · 0 removed ")) {
		t.Fatal("history mode did not dim with the inactive panel border")
	}

	m = m.SetBounds(0, 0, 34, 10)
	narrowTop := strings.Split(ansi.Strip(m.ViewWithBorder(true, true)), "\n")[0]
	if !strings.Contains(narrowTop, "Δ") {
		t.Fatalf("narrow history border dropped the compact mode indicator: %q", narrowTop)
	}
}

func historyTreeWithParameters(params ...core.ParametersEntry) *core.ParametersTree {
	return &core.ParametersTree{Groups: []core.ParametersGroup{{Key: "group", Label: "group", Parameters: params}}}
}

func historyTestTree(keys ...string) *core.ParametersTree {
	params := make([]core.ParametersEntry, 0, len(keys))
	for _, key := range keys {
		params = append(params, core.ParametersEntry{Key: key})
	}
	return &core.ParametersTree{Groups: []core.ParametersGroup{{Key: "group", Label: "group", Parameters: params}}}
}

func visibleParameterKeys(m Model) []string {
	var keys []string
	for _, node := range m.visible {
		if node.kind == nodeParameter {
			keys = append(keys, node.paramKey)
		}
	}
	return keys
}
