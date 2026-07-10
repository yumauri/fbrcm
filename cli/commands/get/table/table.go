package table

import (
	"image/color"
	"os"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
	lipglosstable "charm.land/lipgloss/v2/table"
	"golang.org/x/term"

	"github.com/yumauri/fbrcm/cli/shared"
	clistyles "github.com/yumauri/fbrcm/cli/styles"
)

// Render formats parameter rows as a terminal table.
func Render(rows []Row, highlightFilters []shared.QueryFilter, allowHideKey, includeProject bool) string {
	noColor := clistyles.NoColorEnabled()
	projectWidth := lipgloss.Width("Project")
	groupWidth := lipgloss.Width("Group")
	keyWidth := lipgloss.Width("Key")
	typeWidth := lipgloss.Width("Type")
	globalLabelWidth := longestConditionWidth(rows)
	layout := chooseTableLayout(rows, globalLabelWidth, includeProject, allowHideKey)
	valuesWidth := max(lipgloss.Width("Values"), layout.valueWidth)
	tableRows := make([][]string, 0, len(rows))

	for _, row := range rows {
		rowIndex := len(tableRows)
		var rowBG color.Color
		if !noColor && isStripedDataRow(rowIndex) {
			rowBG = clistyles.ColorRowStripe
		}
		projectCell := row.Project
		if strings.TrimSpace(projectCell) == "" {
			projectCell = row.ProjectID
		}
		highlights := shared.HighlightFilters(row.Key, highlightFilters)
		keyCell := renderHighlightedText(row.Key, clistyles.PanelText, highlights, rowBG)

		rowCells := make([]string, 0, 5)
		if layout.includeProject {
			rowCells = append(rowCells, projectCell)
		}
		if layout.includeGroup {
			rowCells = append(rowCells, row.Group)
		}
		if layout.includeKey {
			rowCells = append(rowCells, keyCell)
		}
		if layout.includeType {
			rowCells = append(rowCells, row.Type)
		}
		renderedValues := renderValueTree(row.ValueLines, row.Status, globalLabelWidth, layout.showNames, layout.valueWidth, rowBG)
		rowCells = append(rowCells, renderedValues)
		tableRows = append(tableRows, rowCells)

		if layout.includeProject {
			projectWidth = max(projectWidth, lipgloss.Width(projectCell))
		}
		if layout.includeGroup {
			groupWidth = max(groupWidth, lipgloss.Width(row.Group))
		}
		if layout.includeKey {
			keyWidth = max(keyWidth, lipgloss.Width(row.Key))
		}
		if layout.includeType {
			typeWidth = max(typeWidth, lipgloss.Width(row.Type))
		}
		valuesWidth = max(valuesWidth, lipgloss.Width(renderedValues))
	}

	styleFunc := func(row, col int) lipgloss.Style {
		style := lipgloss.NewStyle().Padding(0, 1)
		if noColor {
			return style
		}
		if row == lipglosstable.HeaderRow {
			return style.Bold(true).Foreground(clistyles.PaletteSlateBright)
		}
		if isStripedDataRow(row) {
			style = style.Background(clistyles.ColorRowStripe)
		}
		keyCol := 0
		if layout.includeProject {
			keyCol = 1
		}
		if layout.includeGroup {
			keyCol++
		}
		switch col {
		case 0:
			if layout.includeProject {
				if isErrorStatus(rowStatus(rows, row)) {
					return style.Foreground(clistyles.PaletteError)
				}
				return style.Foreground(clistyles.PaletteSlateBright)
			}
			return style.Foreground(clistyles.PaletteBlueBright)
		case keyCol:
			return style.Foreground(clistyles.PaletteBlueBright)
		default:
			return style.Foreground(clistyles.PaletteSlateDim)
		}
	}

	headers := make([]string, 0, 5)
	if layout.includeProject {
		headers = append(headers, "Project")
	}
	if layout.includeGroup {
		headers = append(headers, "Group")
	}
	if layout.includeKey {
		headers = append(headers, "Key")
	}
	if layout.includeType {
		headers = append(headers, "Type")
	}
	headers = append(headers, "Values")

	width := valuesWidth + tableOverhead(len(headers))
	if layout.includeKey {
		width += keyWidth
	}
	if layout.includeProject {
		width += projectWidth
	}
	if layout.includeGroup {
		width += groupWidth
	}
	if layout.includeType {
		width += typeWidth
	}

	tbl := lipglosstable.New().
		Headers(headers...).
		Rows(tableRows...).
		Width(width).
		Border(lipgloss.NormalBorder()).
		BorderHeader(true).
		BorderRow(false).
		StyleFunc(styleFunc)
	if !noColor {
		tbl = tbl.BorderStyle(clistyles.BorderStyle(false))
	}
	return tbl.String()
}

func isStripedDataRow(row int) bool {
	return row >= 0 && row%2 == 1
}

func rowStatus(rows []Row, row int) string {
	if row < 0 {
		return ""
	}
	if row >= len(rows) {
		return ""
	}
	return rows[row].Status
}

func chooseTableLayout(rows []Row, labelWidth int, includeProject bool, allowHideKey bool) tableLayout {
	terminalWidth := detectTerminalWidth()
	layout := tableLayout{
		includeProject: includeProject,
		includeGroup:   true,
		includeKey:     true,
		includeType:    true,
		showNames:      true,
		valueWidth:     max(lipgloss.Width("Values"), maxValueWidth(rows, labelWidth, true)),
	}
	if terminalWidth <= 0 {
		return layout
	}

	projectWidth := lipgloss.Width("Project")
	groupWidth := lipgloss.Width("Group")
	keyWidth := lipgloss.Width("Key")
	typeWidth := lipgloss.Width("Type")
	for _, row := range rows {
		projectCell := row.Project
		if strings.TrimSpace(projectCell) == "" {
			projectCell = row.ProjectID
		}
		projectWidth = max(projectWidth, lipgloss.Width(projectCell))
		groupWidth = max(groupWidth, lipgloss.Width(row.Group))
		keyWidth = max(keyWidth, lipgloss.Width(row.Key))
		typeWidth = max(typeWidth, lipgloss.Width(row.Type))
	}

	available := func(includeGroup, includeKey, includeType bool) int {
		cols := 0
		width := 0
		if includeProject {
			cols++
			width += projectWidth
		}
		if includeGroup {
			cols++
			width += groupWidth
		}
		if includeKey {
			cols++
			width += keyWidth
		}
		if includeType {
			cols++
			width += typeWidth
		}
		cols++ // values
		return terminalWidth - width - tableOverhead(cols)
	}

	natural := maxValueWidth(rows, labelWidth, true)
	valueWidth := available(true, true, true)
	clippingNeeded := natural > valueWidth
	valueRoom := minValueRoom(rows, labelWidth, true, valueWidth)
	if clippingNeeded && valueRoom < 10 {
		layout.includeType = false
		valueWidth = available(true, true, false)
		clippingNeeded = natural > valueWidth
		valueRoom = minValueRoom(rows, labelWidth, true, valueWidth)
	}
	if clippingNeeded && valueRoom < 10 {
		layout.includeGroup = false
		valueWidth = available(false, true, false)
		clippingNeeded = natural > valueWidth
		valueRoom = minValueRoom(rows, labelWidth, true, valueWidth)
	}
	if clippingNeeded && valueRoom < 10 {
		layout.showNames = false
		natural = maxValueWidth(rows, labelWidth, false)
		valueWidth = available(layout.includeGroup, true, layout.includeType)
		clippingNeeded = natural > valueWidth
		valueRoom = minValueRoom(rows, labelWidth, false, valueWidth)
	}
	if allowHideKey && clippingNeeded && valueRoom < 10 {
		layout.includeKey = false
	}

	natural = maxValueWidth(rows, labelWidth, layout.showNames)
	valueWidth = available(layout.includeGroup, layout.includeKey, layout.includeType)
	if valueWidth <= 0 {
		valueWidth = 1
	}
	layout.valueWidth = max(1, min(natural, valueWidth))
	return layout
}

func detectTerminalWidth() int {
	if columns := strings.TrimSpace(os.Getenv("COLUMNS")); columns != "" {
		if width, err := strconv.Atoi(columns); err == nil && width > 0 {
			return width
		}
	}

	info, err := os.Stdout.Stat()
	if err == nil && (info.Mode()&os.ModeCharDevice) != 0 {
		width, _, err := term.GetSize(int(os.Stdout.Fd()))
		if err == nil && width > 0 {
			return width
		}
	}

	return 80
}

func tableOverhead(cols int) int {
	return cols*3 + 1
}
