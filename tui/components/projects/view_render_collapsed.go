package projects

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/yumauri/fbrcm/tui/styles"
)

var (
	panelTitle = "[1] Projects"

	itemStyle = styles.PanelText
	metaStyle = styles.PanelMuted
)

type scrollbarState struct {
	visible    bool
	thumbStart int
	thumbEnd   int
}

type secondaryTitleState struct {
	text  string
	style lipgloss.Style
}

func renderCollapsedPanel(height int, active bool) string {
	if height <= 0 {
		return ""
	}

	borderStyle := styles.BorderStyle(active)
	titleStyle := collapsedTitleStyle(active)
	railStyle := collapsedRailStyle(active)
	capStyle := collapsedCapStyle(active)

	lines := make([]string, 0, height)
	lines = append(lines, borderStyle.Render("─╮"))

	bodyHeight := max(height-2, 0)
	if bodyHeight == 0 {
		return strings.Join(lines, "\n")
	}

	content := collapsedTitleBody(bodyHeight)
	for _, row := range content {
		switch row {
		case "top":
			if active {
				lines = append(lines, capStyle.Render("▗▄▖"))
			} else {
				lines = append(lines, " "+borderStyle.Render("╵")+" ")
			}
		case "fill":
			lines = append(lines, borderStyle.Render(" │"))
		case "bottom":
			if active {
				lines = append(lines, capStyle.Render("▝▀▘"))
			} else {
				lines = append(lines, " "+borderStyle.Render("╷")+" ")
			}
		default:
			if active {
				lines = append(lines, railStyle.Render("▐")+titleStyle.Render(row)+railStyle.Render("▌"))
			} else {
				lines = append(lines, " "+row+" ")
			}
		}
	}

	if height > 1 {
		lines = append(lines, borderStyle.Render("─╯"))
	}

	return strings.Join(lines, "\n")
}

func collapsedTitleStyle(active bool) lipgloss.Style {
	if !active {
		return lipgloss.NewStyle()
	}
	return styles.TitleStyle(true)
}

func collapsedRailStyle(active bool) lipgloss.Style {
	if !active {
		return lipgloss.NewStyle()
	}
	if styles.NoColorEnabled() {
		return lipgloss.NewStyle().Bold(true)
	}
	return lipgloss.NewStyle().Foreground(styles.PaletteBlueDeep)
}

func collapsedCapStyle(active bool) lipgloss.Style {
	if !active {
		return lipgloss.NewStyle()
	}
	if styles.NoColorEnabled() {
		return lipgloss.NewStyle().Bold(true)
	}
	return lipgloss.NewStyle().Foreground(styles.PaletteBlueDeep)
}

func collapsedTitleBody(bodyHeight int) []string {
	if bodyHeight <= 0 {
		return nil
	}

	letters := []string{"P", "r", "o", "j", "e", "c", "t", "s"}
	rows := []string{"top"}
	if bodyHeight == 1 {
		return rows
	}

	contentSlots := bodyHeight - 2
	if contentSlots <= 0 {
		return append(rows, "bottom")
	}

	if contentSlots >= len(letters) {
		rows = append(rows, letters...)
		rows = append(rows, "bottom")
		for len(rows) < bodyHeight {
			rows = append(rows, "fill")
		}
		return rows
	}

	visibleLetters := max(contentSlots-1, 0)
	rows = append(rows, letters[:visibleLetters]...)
	rows = append(rows, "⋮")
	rows = append(rows, "bottom")
	for len(rows) < bodyHeight {
		rows = append(rows, "fill")
	}
	return rows
}
