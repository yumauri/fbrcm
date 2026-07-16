package workspaceheader

import (
	"strings"

	"charm.land/lipgloss/v2"

	tuiconfig "github.com/yumauri/fbrcm/tui/config"
	"github.com/yumauri/fbrcm/tui/styles"
)

const tabCount = 3

var tabs = [tabCount]struct {
	action tuiconfig.Action
	label  string
}{
	{tuiconfig.ActionFocusParameters, "Parameters"},
	{tuiconfig.ActionFocusConditions, "Conditions"},
	{tuiconfig.ActionFocusHistory, "History"},
}

// Render returns the shared Parameters/Conditions/History tab strip and width.
func Render(width, selected int, focused bool, borderStyle lipgloss.Style) (string, int) {
	keys := keys()
	widths := tabWidths(width, selected, keys)
	parts := [tabCount]string{}
	totalWidth := 0
	for i, tab := range tabs {
		parts[i], widths[i] = styles.PanelHeaderTab(keys[i], tab.label, i == selected, focused, widths[i])
		totalWidth += widths[i]
	}
	return strings.Join(parts[:], borderStyle.Render("──")), totalWidth + 2*(tabCount-1)
}

// TabAt returns the tab index at a horizontal coordinate relative to the
// panel. Border prefix and separator cells are not part of any tab hitbox.
func TabAt(width, selected, x int) (int, bool) {
	if x < 0 {
		return 0, false
	}
	widths := tabWidths(width, selected, keys())
	x -= min(2, width)
	for index, tabWidth := range widths {
		if x >= 0 && x < tabWidth {
			return index, true
		}
		x -= tabWidth
		if index < tabCount-1 {
			if x >= 0 && x < 2 {
				return 0, false
			}
			x -= 2
		}
	}
	return 0, false
}

func keys() [tabCount]string {
	var out [tabCount]string
	for i, tab := range tabs {
		out[i] = tuiconfig.ActionKeyHint(tuiconfig.BlockGlobal, tab.action)
	}
	return out
}

func tabWidths(width, selected int, keys [tabCount]string) [tabCount]int {
	var full [tabCount]int
	fullWidth := 0
	for i, tab := range tabs {
		full[i] = lipgloss.Width(" " + keys[i] + tab.label + " ")
		fullWidth += full[i]
	}
	available := max(width-7, 0) // prefix, two separators, and right corner
	if fullWidth <= available {
		return full
	}
	compact := [tabCount]int{3, 3, 3}
	compact[selected] = min(full[selected], max(available-6, 3))
	return compact
}
