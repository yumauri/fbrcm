package get

import (
	"image/color"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"golang.org/x/term"

	"github.com/yumauri/fbrcm/cli/shared"
	clistyles "github.com/yumauri/fbrcm/cli/styles"
	"github.com/yumauri/fbrcm/core/firebase"
	corestyles "github.com/yumauri/fbrcm/core/styles"
)

func renderParametersTable(rows []parameterRow, highlightFilters []shared.QueryFilter, allowHideKey, includeProject bool) string {
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
		if row == table.HeaderRow {
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

	tbl := table.New().
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

func rowStatus(rows []parameterRow, row int) string {
	if row < 0 {
		return ""
	}
	if row >= len(rows) {
		return ""
	}
	return rows[row].Status
}

func chooseTableLayout(rows []parameterRow, labelWidth int, includeProject bool, allowHideKey bool) tableLayout {
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

func renderHighlightedText(value string, base lipgloss.Style, highlights []int, rowBG color.Color) string {
	if clistyles.NoColorEnabled() || len(highlights) == 0 {
		return value
	}

	highlightSet := make(map[int]struct{}, len(highlights))
	for _, idx := range highlights {
		highlightSet[idx] = struct{}{}
	}

	runes := []rune(value)
	parts := make([]string, 0, len(runes))
	for i, r := range runes {
		style := applyBackground(base, rowBG)
		if _, ok := highlightSet[i]; ok {
			style = applyBackground(lipgloss.NewStyle().Foreground(clistyles.PaletteYellow), rowBG)
			parts = append(parts, style.Render(string(r)))
			continue
		}
		parts = append(parts, style.Render(string(r)))
	}
	return strings.Join(parts, "")
}

func renderConditionLabel(label string, conditionColor, rowBG color.Color) string {
	if clistyles.NoColorEnabled() {
		return label
	}
	return applyBackground(lipgloss.NewStyle().Foreground(conditionColor), rowBG).Render(label)
}

func renderValueTree(lines []valueLine, status string, labelWidth int, showNames bool, maxWidth int, rowBG color.Color) string {
	if len(lines) == 0 {
		return ""
	}

	rendered := make([]string, 0, len(lines))
	for i, line := range lines {
		prefix := valueTreePrefix(i, len(lines))
		label := line.Label
		if line.Missing {
			label = renderMissingLabel(status, rowBG)
			rendered = append(rendered, clipStyledLine(renderTreeChrome(prefix, rowBG)+renderTreeChrome(" ", rowBG)+label, maxWidth))
			continue
		} else if line.IsDefault {
			label = renderDefaultLabel(label, rowBG)
		} else {
			label = renderConditionLabel(label, line.Color, rowBG)
		}

		if !showNames {
			head := renderTreeChrome(prefix+" ", rowBG)
			value := renderValueText(clipPlainText(line.Value, max(maxWidth-lipgloss.Width(head), 1)), line.ValueType, rowBG)
			rendered = append(rendered, head+value)
			continue
		}
		fillWidth := max(labelWidth-lipgloss.Width(line.Label)+1, 1)
		filler := renderTreeChrome(strings.Repeat("╌", fillWidth), rowBG)
		head := renderTreeChrome(prefix+" ", rowBG) + label + renderTreeChrome(" ", rowBG) + filler + renderTreeChrome(" ", rowBG)
		value := renderValueText(clipPlainText(line.Value, max(maxWidth-lipgloss.Width(head), 1)), line.ValueType, rowBG)
		rendered = append(rendered, head+value)
	}

	return strings.Join(rendered, "\n")
}

func longestConditionWidth(rows []parameterRow) int {
	width := lipgloss.Width("Default value")
	for _, row := range rows {
		for _, line := range row.ValueLines {
			width = max(width, lipgloss.Width(line.Label))
		}
	}
	return width
}

func maxValueWidth(rows []parameterRow, labelWidth int, showNames bool) int {
	width := lipgloss.Width("Values")
	for _, row := range rows {
		width = max(width, lipgloss.Width(renderValueTree(row.ValueLines, row.Status, labelWidth, showNames, 1<<30, nil)))
	}
	return width
}

func minValueRoom(rows []parameterRow, labelWidth int, showNames bool, cellWidth int) int {
	room := 1 << 30
	found := false
	for _, row := range rows {
		for i, line := range row.ValueLines {
			if line.Missing {
				continue
			}
			headWidth := valueLineHeadWidth(line, i, len(row.ValueLines), labelWidth, showNames)
			valueRoom := cellWidth - headWidth
			if !found || valueRoom < room {
				room = valueRoom
				found = true
			}
		}
	}
	if !found {
		return cellWidth
	}
	return room
}

func valueTreePrefix(index, total int) string {
	if total <= 1 {
		return "╌╌╌"
	}
	switch index {
	case 0:
		return "╌┬╌"
	case total - 1:
		return " ╰╌"
	default:
		return " ├╌"
	}
}

func renderTreeChrome(value string, rowBG color.Color) string {
	if clistyles.NoColorEnabled() {
		return value
	}
	return applyBackground(lipgloss.NewStyle().Foreground(clistyles.PaletteSlateDim), rowBG).Render(value)
}

func renderDefaultLabel(label string, rowBG color.Color) string {
	if clistyles.NoColorEnabled() {
		return label
	}
	return applyBackground(lipgloss.NewStyle().Foreground(clistyles.PaletteSlateDim).Italic(true), rowBG).Render(label)
}

func renderMissingLabel(status string, rowBG color.Color) string {
	if clistyles.NoColorEnabled() {
		return "Missing values"
	}
	style := lipgloss.NewStyle().Italic(true).Strikethrough(true)
	if isErrorStatus(status) {
		style = style.Foreground(clistyles.PaletteError)
	} else {
		style = style.Foreground(clistyles.PaletteSlateDim)
	}
	return applyBackground(style, rowBG).Render("Missing values")
}

func isErrorStatus(status string) bool {
	return status == "staled" || status == "missing"
}

func stringPtrOrNil(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	v := value
	return &v
}

func timePtrOrNil(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	v := value
	return &v
}

func renderValueText(value, valueType string, rowBG color.Color) string {
	if value == "" || clistyles.NoColorEnabled() {
		return value
	}
	if strings.HasPrefix(value, "(empty ") && strings.HasSuffix(value, ")") {
		return applyBackground(corestyles.EmptyValueStyle(), rowBG).Render(value)
	}
	style := valueTextStyle(value, valueType)
	return applyBackground(style, rowBG).Render(value)
}

func valueTextStyle(value, valueType string) lipgloss.Style {
	return corestyles.ValueTextStyle(value, valueType)
}

func valueTypeKey(valueType string) string {
	valueType = strings.TrimSpace(strings.ToLower(valueType))
	if valueType == "" {
		return "string"
	}
	return valueType
}

func clipStyledLine(value string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if lipgloss.Width(value) <= maxWidth {
		return value
	}
	return clipPlainText(value, maxWidth)
}

func clipPlainText(value string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= maxWidth {
		return value
	}
	if maxWidth == 1 {
		return "…"
	}
	return string(runes[:maxWidth-1]) + "…"
}

func valueLineHeadWidth(line valueLine, index, total, labelWidth int, showNames bool) int {
	prefixWidth := lipgloss.Width(valueTreePrefix(index, total)) + 1
	if line.Missing {
		return prefixWidth
	}
	if !showNames {
		return prefixWidth
	}
	return prefixWidth + lipgloss.Width(line.Label) + 1 + max(labelWidth-lipgloss.Width(line.Label)+1, 1) + 1
}

func applyBackground(style lipgloss.Style, bg color.Color) lipgloss.Style {
	if bg == nil {
		return style
	}
	return style.Background(bg)
}

func sortedStringKeys[V any](items map[string]V) []string {
	keys := make([]string, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		left := strings.ToLower(keys[i])
		right := strings.ToLower(keys[j])
		if left == right {
			return keys[i] < keys[j]
		}
		return left < right
	})
	return keys
}

func sortedConditionalKeys(items map[string]firebase.RemoteConfigValue, order map[string]int) []string {
	keys := make([]string, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}

	sort.Slice(keys, func(i, j int) bool {
		left, leftOK := order[keys[i]]
		right, rightOK := order[keys[j]]
		switch {
		case leftOK && rightOK && left != right:
			return left < right
		case leftOK != rightOK:
			return leftOK
		default:
			leftKey := strings.ToLower(keys[i])
			rightKey := strings.ToLower(keys[j])
			if leftKey == rightKey {
				return keys[i] < keys[j]
			}
			return leftKey < rightKey
		}
	})

	return keys
}

func formatRemoteConfigValue(value firebase.RemoteConfigValue, valueType string) string {
	switch {
	case value.UseInAppDefault:
		return "<in-app default>"
	case len(value.PersonalizationValue) > 0:
		return "<personalization>"
	case len(value.RolloutValue) > 0:
		return "<rollout>"
	case value.Value == "":
		return "(empty " + emptyValueType(valueType) + ")"
	default:
		return strings.ReplaceAll(value.Value, "\n", "\\n")
	}
}

func valueForJSON(value string) *string {
	if strings.HasPrefix(value, "(empty ") && strings.HasSuffix(value, ")") {
		return nil
	}
	v := value
	return &v
}

func emptyValueType(valueType string) string {
	valueType = strings.TrimSpace(strings.ToLower(valueType))
	if valueType == "" {
		return "string"
	}
	return valueType
}
