package viewutil

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/yumauri/fbrcm/tui/styles"
)

// CollapsedSidebarPanel draws a three-column rail. Optional short titles are
// tried in order. Without them, an oversized title ends in a vertical ellipsis.
func CollapsedSidebarPanel(height int, active, borderActive bool, title string, shortTitles ...string) string {
	if height <= 0 {
		return ""
	}

	border := styles.BorderStyle(borderActive)
	titleStyle := styles.TitleStyle(active)
	railStyle := lipgloss.NewStyle().Foreground(styles.PaletteBlueDeep)
	if styles.NoColorEnabled() {
		railStyle = lipgloss.NewStyle().Bold(true)
	}
	if active && !borderActive {
		railStyle = border
	}

	lines := make([]string, 0, height)
	lines = append(lines, border.Render("─╮"))
	bodyHeight := max(height-2, 0)
	slots := max(bodyHeight-2, 0)
	letters := collapsedSidebarLetters(title, shortTitles, slots)
	if bodyHeight > 0 && (len(letters) > 0 || len(shortTitles) == 0) {
		top, bottom := " "+border.Render("╵")+" ", " "+border.Render("╷")+" "
		if active {
			top, bottom = railStyle.Render("▗▄▖"), railStyle.Render("▝▀▘")
		}
		lines = append(lines, top)
		for _, letter := range letters {
			line := " " + string(letter) + " "
			if active {
				line = railStyle.Render("▐") + titleStyle.Render(string(letter)) + railStyle.Render("▌")
			}
			lines = append(lines, line)
		}
		if bodyHeight > 1 {
			lines = append(lines, bottom)
		}
	}
	for len(lines) <= bodyHeight {
		lines = append(lines, " "+border.Render("│")+" ")
	}
	if height > 1 {
		lines = append(lines, border.Render("─╯"))
	}
	return strings.Join(lines, "\n")
}

func collapsedSidebarLetters(title string, shortTitles []string, slots int) []rune {
	letters := []rune(title)
	if len(letters) <= slots {
		return letters
	}
	for _, short := range shortTitles {
		if letters := []rune(short); len(letters) <= slots {
			return letters
		}
	}
	if len(shortTitles) > 0 || slots == 0 {
		return nil
	}
	return append(letters[:slots-1:slots-1], '⋮')
}
