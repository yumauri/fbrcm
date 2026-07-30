package workspaceheader

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/tui/styles"
)

func TestWideLayoutShowsEveryFullTitle(t *testing.T) {
	layout := LayoutFor(140, 0)
	if layout.HasOverflow() {
		t.Fatalf("wide layout hidden tabs = %v, want none", layout.HiddenTabs())
	}
	rendered, _ := Render(140, 0, true, lipgloss.NewStyle())
	for _, tab := range tabs {
		if !strings.Contains(rendered, tab.label) {
			t.Errorf("wide header does not contain %q: %q", tab.label, rendered)
		}
	}
}

func TestCompactLayoutKeepsLeadingTitlesAndNextTitleSlot(t *testing.T) {
	layout := LayoutFor(85, 0)
	if got, want := layout.HiddenTabs(), []int{4, 5}; !equalIndexes(got, want) {
		t.Fatalf("hidden tabs = %v, want %v", got, want)
	}
	for _, index := range []int{0, 1, 2, 3} {
		if !layoutHasTab(layout, index) {
			t.Errorf("compact layout does not show tab %d", index)
		}
	}
	if !layout.HasOverflow() {
		t.Fatal("compact layout has no overflow button")
	}
}

func TestHiddenActiveTitleReplacesTrailingSlot(t *testing.T) {
	layout := LayoutFor(85, 4)
	if !layoutHasTab(layout, 4) {
		t.Fatal("active Personalizations title is not visible")
	}
	if layoutHasTab(layout, 3) {
		t.Fatal("default A/B Tests slot remains visible after active replacement")
	}
	if got, want := layout.HiddenTabs(), []int{3, 5}; !equalIndexes(got, want) {
		t.Fatalf("hidden tabs = %v, want %v", got, want)
	}
}

func TestWideActiveReplacementRemovesLeadingTitleBeforeTruncating(t *testing.T) {
	layout := LayoutFor(60, 4)
	if !layoutHasTab(layout, 4) {
		t.Fatal("active Personalizations title is not visible")
	}
	if layoutHasTab(layout, 1) {
		t.Fatal("Conditions remained visible when the active full title needed its width")
	}
	rendered, _ := Render(60, 4, true, lipgloss.NewStyle())
	if !strings.Contains(rendered, "Personalizations") {
		t.Fatalf("active title was truncated: %q", rendered)
	}
}

func TestTabAndOverflowHitboxesMatchLayout(t *testing.T) {
	for _, test := range []struct {
		width    int
		selected int
	}{{140, 0}, {85, 0}, {85, 4}, {60, 4}, {40, 0}} {
		layout := LayoutFor(test.width, test.selected)
		for _, item := range layout.parts {
			for column := titleStripX(item.start); column < titleStripX(item.start+item.width); column++ {
				if item.overflow {
					if !OverflowAt(test.width, test.selected, column) {
						t.Fatalf("OverflowAt(width=%d, selected=%d, x=%d) = false", test.width, test.selected, column)
					}
					if _, ok := TabAt(test.width, test.selected, column); ok {
						t.Fatalf("overflow column %d is also a tab hitbox", column)
					}
					continue
				}
				if got, ok := TabAt(test.width, test.selected, column); !ok || got != item.tab {
					t.Fatalf("TabAt(width=%d, selected=%d, x=%d) = (%d, %v), want (%d, true)", test.width, test.selected, column, got, ok, item.tab)
				}
			}
		}
	}
}

func TestPopupKeepsCollapsedTitlesOrderedAroundBorderRow(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	layout := LayoutFor(85, 4)
	cursor := layout.MenuCursor()
	if got, want := layout.MenuTabs(), []int{3, 4, 5}; !equalIndexes(got, want) {
		t.Fatalf("menu tabs = %v, want ordered tabs %v", got, want)
	}
	if cursor != 1 {
		t.Fatalf("menu cursor = %d, want Personalizations row 1", cursor)
	}
	view := layout.PopupView(cursor, lipgloss.NewStyle())
	if !strings.Contains(view, "Rollouts") {
		t.Fatalf("popup does not contain the title after the border row: %q", view)
	}
	if !strings.Contains(view, "Personalizations") {
		t.Errorf("popup does not contain the border-row title: %q", view)
	}
	if plain := ansi.Strip(view); !strings.Contains(plain, "Personalizations  │") {
		t.Errorf("popup does not keep two spaces after its longest title: %q", plain)
	}
	if strings.Contains(view, "A/B Tests") {
		t.Errorf("popup contains title above the visible stack: %q", view)
	}
	geometry, ok := layout.PopupGeometry()
	if !ok {
		t.Fatal("compact layout has no popup geometry")
	}
	if geometry.LeftPadding != 7 {
		t.Fatalf("popup left padding = %d, want 7 to align with the trailing title", geometry.LeftPadding)
	}
	for row, want := range geometry.Tabs[cursor:] {
		if got, exists := layout.PopupTabAt(cursor, layout.trailingTitleStart(), row); !exists || got != want {
			t.Fatalf("PopupTabAt(row=%d) = (%d, %v), want (%d, true)", row, got, exists, want)
		}
		if key := styles.FilterText.Render(layout.keys[want]); !strings.Contains(view, key) {
			t.Errorf("popup key hint for tab %d is not highlighted: %q", want, view)
		}
	}
}

func TestRolloutsOpensOrderedStackAtMaximumUpwardShift(t *testing.T) {
	layout := LayoutForRightReserve(56, 5, 9)
	if got, want := layout.MenuTabs(), []int{1, 2, 3, 4, 5}; !equalIndexes(got, want) {
		t.Fatalf("menu tabs = %v, want %v", got, want)
	}
	if cursor := layout.MenuCursor(); cursor != 4 {
		t.Fatalf("Rollouts menu cursor = %d, want last row 4", cursor)
	}
	if popup := layout.PopupView(layout.MenuCursor(), lipgloss.NewStyle()); lipgloss.Height(popup) != 2 {
		t.Fatalf("Rollouts menu should be shifted fully upward, got:\n%s", popup)
	}
}

func TestOpenMenuPreviewKeepsCollapseGeometryStable(t *testing.T) {
	initial := LayoutForRightReserve(85, 0, 0)
	opened := initial.WithPreview(0)
	history := initial.WithPreview(2)
	personalizations := initial.WithPreview(4)

	initialX, _, ok := initial.OverflowButtonBounds()
	if !ok {
		t.Fatal("initial menu layout has no overflow button")
	}
	for name, layout := range map[string]Layout{
		"opened":           opened,
		"History":          history,
		"Personalizations": personalizations,
	} {
		x, _, exists := layout.OverflowButtonBounds()
		if !exists {
			t.Fatalf("%s preview has no overflow button", name)
		}
		if x != initialX {
			t.Errorf("%s preview moved overflow button from %d to %d", name, initialX, x)
		}
		if got, want := layout.MenuTabs(), initial.MenuTabs(); !equalIndexes(got, want) {
			t.Errorf("%s preview menu tabs = %v, want stable %v", name, got, want)
		}
	}
	if history.selected != 0 {
		t.Fatalf("History preview changed selected tab to %d, want Parameters", history.selected)
	}
}

func TestMenuPreviewTitleUsesNormalListColor(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	rendered, _ := RenderMenuWithRightReserve(85, 0, 2, true, lipgloss.NewStyle(), 0)
	if !strings.Contains(rendered, styles.PanelTitle.Render("History ")) {
		t.Fatalf("preview title does not use the normal title color: %q", rendered)
	}
	if strings.Contains(rendered, styles.PanelTitleInactiveTab.Render("History ")) {
		t.Fatalf("preview title uses inactive gray: %q", rendered)
	}
}

func TestActiveTitleKeepsSelectionWhenItMovesBelowCursor(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	activeLayout := LayoutForRightReserve(85, 4, 0)
	activeCursor := activeLayout.MenuCursor()
	if activeCursor == 0 {
		t.Fatal("test layout has no title before active Personalizations")
	}
	preview := activeLayout.MenuTabs()[activeCursor-1]
	layout := activeLayout.WithPreview(preview)
	view := layout.PopupViewFocused(layout.MenuCursor(), true, lipgloss.NewStyle())
	activeLabel := styles.TitleStyle(true).Render("Personalizations ")
	if !strings.Contains(view, activeLabel) {
		t.Fatalf("active Personalizations title lost its selection style below the cursor: %q", view)
	}
	firstLine, _, _ := strings.Cut(view, "\n")
	if !strings.HasPrefix(firstLine, "│") || strings.Contains(firstLine, "─") {
		t.Fatalf("cursor row is not a border-free interior row: %q", firstLine)
	}
	if plain := ansi.Strip(firstLine); !strings.HasPrefix(plain, "│  ▸") {
		t.Fatalf("cursor marker is not in the overflow-glyph column: %q", plain)
	}
	if !strings.HasSuffix(view, "╯") {
		t.Fatalf("shifted menu does not retain its surrounding bottom border: %q", view)
	}
}

func TestCursorMarkerInheritsBorderColor(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	layout := LayoutForRightReserve(85, 0, 0).WithPreview(2)
	for _, active := range []bool{false, true} {
		border := styles.BorderStyle(active)
		view := layout.PopupViewFocused(layout.MenuCursor(), active, border)
		if marker := border.Render("▸"); !strings.Contains(view, marker) {
			t.Errorf("active=%v cursor marker does not inherit border style: %q", active, view)
		}
	}
}

func TestOverflowControlKeepsInactiveTitleColor(t *testing.T) {
	rendered := OverflowButton()
	if !strings.Contains(rendered, styles.PanelTitleInactiveTab.Render("≡ ")) {
		t.Fatalf("overflow glyph does not use inactive-title styling: %q", rendered)
	}
	if strings.Contains(rendered, styles.TitleStyle(true).Render("≡")) {
		t.Fatalf("overflow glyph uses active-title styling: %q", rendered)
	}
}

func TestRightReserveKeepsTitlesBeforeOverlay(t *testing.T) {
	const (
		width        = 100
		rightReserve = 28
	)
	layout := LayoutForRightReserve(width, 0, rightReserve)
	if got, limit := titleStripX(layout.width), width-rightReserve; got >= limit {
		t.Fatalf("title strip right edge = %d, want before reserved edge %d", got, limit)
	}
	if !layout.HasOverflow() {
		t.Fatal("right reserve did not compact the title strip")
	}
}

func layoutHasTab(layout Layout, tab int) bool {
	for _, item := range layout.parts {
		if !item.overflow && item.tab == tab {
			return true
		}
	}
	return false
}

func equalIndexes(left, right []int) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
