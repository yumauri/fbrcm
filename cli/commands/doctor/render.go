package doctor

import (
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/cli/shared"
	clistyles "github.com/yumauri/fbrcm/cli/styles"
	"github.com/yumauri/fbrcm/core"
)

func renderDoctorTable(checks []core.DoctorCheck) string {
	return renderDoctorTableAtWidth(checks, shared.TerminalWidth())
}

func renderDoctorTableAtWidth(checks []core.DoctorCheck, terminalWidth int) string {
	const columnCount = 3

	headers := []string{"Status", "Check", "Detail"}
	naturalWidths := []int{
		lipgloss.Width(headers[0]),
		lipgloss.Width(headers[1]),
		lipgloss.Width(headers[2]),
	}
	for _, check := range checks {
		naturalWidths[0] = max(naturalWidths[0], lipgloss.Width(strings.ToUpper(check.Status)))
		naturalWidths[1] = max(naturalWidths[1], lipgloss.Width(check.Check))
		naturalWidths[2] = max(naturalWidths[2], lipgloss.Width(check.Detail))
	}
	widths, tableWidth := doctorTableWidths(naturalWidths, terminalWidth, columnCount)
	for i := range headers {
		headers[i] = ansi.Truncate(headers[i], widths[i], "")
	}

	rows := make([][]string, 0, len(checks))
	for _, check := range checks {
		status := strings.ToUpper(check.Status)
		rows = append(rows, []string{
			status,
			check.Check,
			ansi.Wrap(check.Detail, widths[2], "/:;,."),
		})
	}
	noColor := clistyles.NoColorEnabled()
	styleFunc := func(row, col int) lipgloss.Style {
		style := lipgloss.NewStyle().Padding(0, 1)
		if noColor {
			return style
		}
		if row == table.HeaderRow {
			return style.Bold(true).Foreground(clistyles.PaletteSlateBright)
		}
		if row >= 0 && row%2 == 1 {
			style = style.Background(clistyles.ColorRowStripe)
		}
		if col == 0 && row >= 0 && row < len(checks) {
			switch checks[row].Status {
			case core.DoctorPass:
				return style.Foreground(clistyles.ColorAdded)
			case core.DoctorWarn:
				return style.Foreground(clistyles.PaletteYellow)
			case core.DoctorFail:
				return style.Foreground(clistyles.PaletteError)
			}
		}
		if col == 1 {
			return style.Foreground(clistyles.PaletteSlateBright)
		}
		return style.Foreground(clistyles.PaletteSlateDim)
	}
	tbl := table.New().
		Headers(headers...).
		Rows(rows...).
		Width(tableWidth).
		Wrap(true).
		Border(lipgloss.NormalBorder()).
		BorderHeader(true).
		BorderRow(false).
		StyleFunc(styleFunc)
	if !noColor {
		tbl = tbl.BorderStyle(clistyles.BorderStyle(false))
	}
	return tbl.String()
}

func doctorTableWidths(natural []int, terminalWidth, columnCount int) ([]int, int) {
	const cellHorizontalPadding = 2

	overhead := columnCount + 1 + columnCount*cellHorizontalPadding
	naturalTableWidth := overhead
	for _, width := range natural {
		naturalTableWidth += width
	}
	widths := append([]int(nil), natural...)
	if terminalWidth <= 0 || naturalTableWidth <= terminalWidth {
		return widths, naturalTableWidth
	}

	availableDetailWidth := terminalWidth - overhead - widths[0] - widths[1]
	widths[2] = min(widths[2], max(lipgloss.Width("Detail"), availableDetailWidth))
	tableWidth := overhead + widths[0] + widths[1] + widths[2]
	return widths, tableWidth
}
