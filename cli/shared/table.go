package shared

import (
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"

	clistyles "github.com/yumauri/fbrcm/cli/styles"
)

// StyledTable renders the shared human-readable CLI table style.
func StyledTable(headers []string, rows [][]string, widths []int, rightAligned map[int]bool, customize func(row, col int, style lipgloss.Style) lipgloss.Style) string {
	noColor := clistyles.NoColorEnabled()
	styleFunc := func(row, col int) lipgloss.Style {
		style := lipgloss.NewStyle().Padding(0, 1)
		if rightAligned[col] {
			style = style.AlignHorizontal(lipgloss.Right)
		}
		if noColor {
			return style
		}
		if row == table.HeaderRow {
			return style.Bold(true).Foreground(clistyles.PaletteSlateBright)
		}
		if row >= 0 && row%2 == 1 {
			style = style.Background(clistyles.ColorRowStripe)
		}
		style = style.Foreground(clistyles.PaletteSlateDim)
		if customize != nil {
			return customize(row, col, style)
		}
		return style
	}

	tbl := table.New().Headers(headers...).Rows(rows...).Width(TableWidth(widths)).Border(lipgloss.NormalBorder()).BorderHeader(true).BorderRow(false).StyleFunc(styleFunc)
	if !noColor {
		tbl = tbl.BorderStyle(clistyles.BorderStyle(false))
	}
	return tbl.String()
}

func TableWidth(widths []int) int {
	width := 3*len(widths) + 1
	for _, cellWidth := range widths {
		width += cellWidth
	}
	return width
}

func HeaderWidths(headers []string) []int {
	widths := make([]int, len(headers))
	for i, header := range headers {
		widths[i] = lipgloss.Width(header)
	}
	return widths
}

func UpdateTableWidths(widths []int, row []string) {
	for i, cell := range row {
		widths[i] = max(widths[i], lipgloss.Width(cell))
	}
}
