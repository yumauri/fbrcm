package workspaceheader

import (
	"strings"

	"charm.land/lipgloss/v2"

	tuiconfig "github.com/yumauri/fbrcm/tui/config"
	"github.com/yumauri/fbrcm/tui/styles"
)

const (
	tabCount             = 6
	fullStatusReserve    = 32
	compactStatusReserve = 4
	panelFrameWidth      = 3
	headerSeparatorWidth = 2
)

var tabs = [tabCount]struct {
	action tuiconfig.Action
	label  string
}{
	{tuiconfig.ActionFocusParameters, "Parameters"},
	{tuiconfig.ActionFocusConditions, "Conditions"},
	{tuiconfig.ActionFocusHistory, "History"},
	{tuiconfig.ActionFocusABTests, "A/B Tests"},
	{tuiconfig.ActionFocusPersonalizations, "Personalizations"},
	{tuiconfig.ActionFocusRollouts, "Rollouts"},
}

type part struct {
	tab      int
	overflow bool
	start    int
	width    int
}

// Layout describes the visible workspace titles and any compact overflow menu.
type Layout struct {
	panelWidth int
	selected   int
	keys       [tabCount]string
	parts      []part
	width      int
}

// MenuGeometry describes the dropdown relative to the workspace panel.
type MenuGeometry struct {
	X           int
	Width       int
	LeftPadding int
	Tabs        []int
}

// LayoutFor resolves the responsive workspace header layout.
func LayoutFor(width, selected int) Layout {
	return LayoutForRightReserve(width, selected, 0)
}

// LayoutForRightReserve resolves the header while keeping the rightmost cells
// free for an app-level overlay such as the current-profile badge.
func LayoutForRightReserve(width, selected, rightReserve int) Layout {
	return layoutForRightReserve(width, selected, rightReserve)
}

func layoutForRightReserve(width, selected, rightReserve int) Layout {
	selected = min(max(selected, 0), tabCount-1)
	rightReserve = max(rightReserve, 0)
	keys := keys()
	fullParts := make([]part, 0, tabCount)
	for index := range tabs {
		fullParts = append(fullParts, part{tab: index, width: tabWidth(index, keys)})
	}
	setPartStarts(fullParts)
	fullWidth := partsWidth(fullParts)
	if fullWidth <= availableWidth(width, max(fullStatusReserve, rightReserve)) {
		return Layout{
			panelWidth: width,
			selected:   selected,
			keys:       keys,
			parts:      fullParts,
			width:      fullWidth,
		}
	}

	limit := availableWidth(width, max(compactStatusReserve, rightReserve))
	overflowWidth := overflowButtonWidth()
	for slotIndex := tabCount - 2; slotIndex >= 0; slotIndex-- {
		slotTab := max(slotIndex, selected)
		parts := make([]part, 0, slotIndex+2)
		for index := 0; index < slotIndex; index++ {
			parts = append(parts, part{tab: index, width: tabWidth(index, keys)})
		}
		parts = append(parts, part{overflow: true, width: overflowWidth})
		parts = append(parts, part{tab: slotTab, width: tabWidth(slotTab, keys)})
		setPartStarts(parts)
		partWidth := partsWidth(parts)
		maxSlotWidth := 0
		for index := slotIndex; index < tabCount; index++ {
			maxSlotWidth = max(maxSlotWidth, tabWidth(index, keys))
		}
		fittedWidth := partWidth + maxSlotWidth - tabWidth(slotTab, keys)
		if fittedWidth > limit && slotIndex > 0 {
			continue
		}
		return Layout{
			panelWidth: width,
			selected:   selected,
			keys:       keys,
			parts:      parts,
			width:      partWidth,
		}
	}

	return Layout{panelWidth: width, selected: selected, keys: keys}
}

// Render returns the shared workspace tab strip and width.
func Render(width, selected int, focused bool, borderStyle lipgloss.Style) (string, int) {
	return RenderWithRightReserve(width, selected, focused, borderStyle, 0)
}

// RenderWithRightReserve renders the tab strip while leaving room at the
// panel's right edge for an app-level overlay.
func RenderWithRightReserve(width, selected int, focused bool, borderStyle lipgloss.Style, rightReserve int) (string, int) {
	layout := LayoutForRightReserve(width, selected, rightReserve)
	return renderLayout(layout, focused, borderStyle, -1)
}

// RenderMenuWithRightReserve renders an open menu's preview title without
// changing which workspace title is active.
func RenderMenuWithRightReserve(width, selected, preview int, focused bool, borderStyle lipgloss.Style, rightReserve int) (string, int) {
	layout := LayoutForRightReserve(width, selected, rightReserve).WithPreview(preview)
	return renderLayout(layout, focused, borderStyle, preview)
}

func renderLayout(layout Layout, focused bool, borderStyle lipgloss.Style, preview int) (string, int) {
	rendered := make([]string, 0, len(layout.parts))
	for _, item := range layout.parts {
		if item.overflow {
			rendered = append(rendered, OverflowButton())
			continue
		}
		selected := item.tab == layout.selected
		titleFocused := focused
		if item.tab == preview && !selected {
			selected = true
			titleFocused = false
		}
		title, _ := styles.PanelHeaderTab(layout.keys[item.tab], tabs[item.tab].label, selected, titleFocused, item.width)
		rendered = append(rendered, title)
	}
	return strings.Join(rendered, borderStyle.Render(strings.Repeat("─", headerSeparatorWidth))), layout.width
}

// OverflowButton renders the compact workspace-menu button.
func OverflowButton() string {
	key := tuiconfig.ActionKeyHint(tuiconfig.BlockGlobal, tuiconfig.ActionWorkspaceMenu)
	text := " " + key + "≡ "
	if key == "" || !strings.Contains(text, key) {
		return styles.PanelTitleInactiveTab.Render(text)
	}
	before, after, _ := strings.Cut(text, key)
	return styles.PanelTitleInactiveTab.Render(before) +
		styles.FilterText.Render(key) +
		styles.PanelTitleInactiveTab.Render(after)
}

// TabAt returns the visible tab index at a horizontal panel coordinate.
func TabAt(width, selected, x int) (int, bool) {
	return LayoutFor(width, selected).TabAt(x)
}

// TabAt returns the visible tab index at a horizontal panel coordinate.
func (l Layout) TabAt(x int) (int, bool) {
	for _, item := range l.parts {
		if item.overflow || x < titleStripX(item.start) || x >= titleStripX(item.start+item.width) {
			continue
		}
		return item.tab, true
	}
	return 0, false
}

// OverflowAt reports whether a horizontal panel coordinate hits the button.
func OverflowAt(width, selected, x int) bool {
	return LayoutFor(width, selected).OverflowAt(x)
}

// OverflowAt reports whether a horizontal panel coordinate hits the button.
func (l Layout) OverflowAt(x int) bool {
	buttonX, buttonWidth, ok := l.OverflowButtonBounds()
	return ok && x >= buttonX && x < buttonX+buttonWidth
}

// HasOverflow reports whether some workspace titles are hidden.
func (l Layout) HasOverflow() bool {
	_, _, ok := l.OverflowButtonBounds()
	return ok
}

// HiddenTabs returns hidden tab indexes in their original order.
func (l Layout) HiddenTabs() []int {
	return hiddenTabs(l.parts)
}

// WithPreview returns the same fitted layout with a different ordered menu tab
// in the trailing slot. Fitting is intentionally unchanged because the base
// layout already reserves the widest possible trailing title.
func (l Layout) WithPreview(tab int) Layout {
	menuTabs := l.MenuTabs()
	if len(menuTabs) == 0 || len(l.parts) == 0 {
		return l
	}
	tab = min(max(tab, menuTabs[0]), menuTabs[len(menuTabs)-1])
	parts := append([]part(nil), l.parts...)
	last := len(parts) - 1
	parts[last].tab = tab
	parts[last].width = tabWidth(tab, l.keys)
	setPartStarts(parts)
	l.parts = parts
	l.width = partsWidth(parts)
	return l
}

// MenuTabs returns the complete ordered sequence represented by the overflow
// stack, including the title currently occupying the trailing header slot.
func (l Layout) MenuTabs() []int {
	leading := 0
	hasOverflow := false
	for _, item := range l.parts {
		if item.overflow {
			hasOverflow = true
			break
		}
		leading++
	}
	if !hasOverflow {
		return nil
	}
	out := make([]int, 0, tabCount-leading)
	for index := leading; index < tabCount; index++ {
		out = append(out, index)
	}
	return out
}

// MenuCursor returns the ordered overflow-stack row shown in the panel border.
func (l Layout) MenuCursor() int {
	menuTabs := l.MenuTabs()
	if len(menuTabs) == 0 || len(l.parts) == 0 {
		return 0
	}
	slotTab := l.parts[len(l.parts)-1].tab
	for index, tab := range menuTabs {
		if tab == slotTab {
			return index
		}
	}
	return 0
}

// OverflowButtonBounds returns button coordinates relative to the panel.
func (l Layout) OverflowButtonBounds() (int, int, bool) {
	for _, item := range l.parts {
		if item.overflow {
			return titleStripX(item.start), item.width, true
		}
	}
	return 0, 0, false
}

// PopupGeometry returns the dropdown bounds relative to the workspace panel.
func (l Layout) PopupGeometry() (MenuGeometry, bool) {
	buttonX, _, ok := l.OverflowButtonBounds()
	menuTabs := l.MenuTabs()
	if !ok || len(menuTabs) == 0 {
		return MenuGeometry{}, false
	}

	left := max(buttonX-1, 1)
	slotStart := l.trailingTitleStart()
	leftPadding := max(slotStart-left, 1)
	contentWidth := 0
	for _, tab := range menuTabs {
		suffixWidth := lipgloss.Width(l.keys[tab] + tabs[tab].label + "  ")
		contentWidth = max(contentWidth, leftPadding+suffixWidth)
	}
	width := contentWidth + 2
	stripRight := titleStripX(l.width)
	width = max(width, stripRight-left+1)
	rightEdge := max(l.panelWidth, 1)
	width = min(width, max(rightEdge-1, 1))
	if left+width > rightEdge {
		left = max(rightEdge-width, 1)
		leftPadding = max(slotStart-left, 1)
	}
	return MenuGeometry{X: left, Width: width, LeftPadding: leftPadding, Tabs: menuTabs}, true
}

// PopupTabAt returns the ordered tab at a visible menu-title coordinate.
func (l Layout) PopupTabAt(cursor, x, y int) (int, bool) {
	geometry, ok := l.PopupGeometry()
	if !ok {
		return 0, false
	}
	cursor = min(max(cursor, 0), len(geometry.Tabs)-1)
	tabIndex := cursor + y
	if y < 0 || tabIndex >= len(geometry.Tabs) {
		return 0, false
	}
	tab := geometry.Tabs[tabIndex]
	titleStart := l.trailingTitleStart()
	if x < titleStart || x >= titleStart+tabWidth(tab, l.keys) {
		return 0, false
	}
	return tab, true
}

// PopupView renders the visible part of the ordered dropdown stack.
func (l Layout) PopupView(cursor int, borderStyle lipgloss.Style) string {
	return l.PopupViewFocused(cursor, false, borderStyle)
}

// PopupViewFocused renders the visible part of the ordered dropdown stack,
// preserving the active title's selection style when the workspace is focused.
func (l Layout) PopupViewFocused(cursor int, focused bool, borderStyle lipgloss.Style) string {
	geometry, ok := l.PopupGeometry()
	if !ok {
		return ""
	}
	cursor = min(max(cursor, 0), len(geometry.Tabs)-1)
	innerWidth := max(geometry.Width-2, 0)
	lines := make([]string, 0, len(geometry.Tabs)-cursor+1)
	prefixWidth := max(geometry.LeftPadding-1, 0)
	markerOffset := lipgloss.Width(" " + tuiconfig.ActionKeyHint(tuiconfig.BlockGlobal, tuiconfig.ActionWorkspaceMenu))
	for index, tab := range geometry.Tabs[cursor:] {
		prefix := strings.Repeat(" ", prefixWidth)
		if index == 0 {
			if markerOffset < prefixWidth {
				prefix = strings.Repeat(" ", markerOffset) +
					borderStyle.Render("▸") +
					strings.Repeat(" ", prefixWidth-markerOffset-1)
			}
		}
		title := l.renderMenuTab(tab, focused)
		padding := max(innerWidth-lipgloss.Width(prefix)-lipgloss.Width(title), 0)
		lines = append(lines, borderStyle.Render("│")+prefix+title+strings.Repeat(" ", padding)+borderStyle.Render("│"))
	}
	lines = append(lines, borderStyle.Render("╰"+strings.Repeat("─", innerWidth)+"╯"))
	return strings.Join(lines, "\n")
}

func (l Layout) renderMenuTab(tab int, focused bool) string {
	rendered, _ := styles.PanelHeaderTab(
		l.keys[tab],
		tabs[tab].label,
		true,
		focused && tab == l.selected,
		tabWidth(tab, l.keys),
	)
	return rendered
}

func keys() [tabCount]string {
	var out [tabCount]string
	for index, tab := range tabs {
		out[index] = tuiconfig.ActionKeyHint(tuiconfig.BlockGlobal, tab.action)
	}
	return out
}

func tabWidth(index int, keys [tabCount]string) int {
	return lipgloss.Width(" " + keys[index] + tabs[index].label + " ")
}

func overflowButtonWidth() int {
	return lipgloss.Width(" " + tuiconfig.ActionKeyHint(tuiconfig.BlockGlobal, tuiconfig.ActionWorkspaceMenu) + "≡ ")
}

func availableWidth(panelWidth, reserve int) int {
	return max(panelWidth-panelFrameWidth-reserve, 0)
}

func titleStripX(x int) int {
	return max(x+2, 0)
}

func setPartStarts(parts []part) {
	x := 0
	for index := range parts {
		parts[index].start = x
		x += parts[index].width
		if index < len(parts)-1 {
			x += headerSeparatorWidth
		}
	}
}

func partsWidth(parts []part) int {
	if len(parts) == 0 {
		return 0
	}
	last := parts[len(parts)-1]
	return last.start + last.width
}

func hiddenTabs(parts []part) []int {
	visible := [tabCount]bool{}
	for _, item := range parts {
		if !item.overflow {
			visible[item.tab] = true
		}
	}
	hidden := make([]int, 0, tabCount-len(parts)+1)
	for index := range tabs {
		if !visible[index] {
			hidden = append(hidden, index)
		}
	}
	return hidden
}

func (l Layout) trailingTitleStart() int {
	if len(l.parts) == 0 {
		return 0
	}
	return titleStripX(l.parts[len(l.parts)-1].start)
}
