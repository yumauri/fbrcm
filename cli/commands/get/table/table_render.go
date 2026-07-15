package table

import (
	"image/color"
	"slices"
	"strings"

	"charm.land/lipgloss/v2"

	clistyles "github.com/yumauri/fbrcm/cli/styles"
	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/core/strfold"
)

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

func renderValueTree(lines []ValueLine, status string, labelWidth int, showNames bool, maxWidth int, rowBG color.Color) string {
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

func longestConditionWidth(rows []Row) int {
	width := lipgloss.Width("Default value")
	for _, row := range rows {
		for _, line := range row.ValueLines {
			width = max(width, lipgloss.Width(line.Label))
		}
	}
	return width
}

func maxValueWidth(rows []Row, labelWidth int, showNames bool) int {
	width := lipgloss.Width("Values")
	for _, row := range rows {
		width = max(width, lipgloss.Width(renderValueTree(row.ValueLines, row.Status, labelWidth, showNames, 1<<30, nil)))
	}
	return width
}

func minValueRoom(rows []Row, labelWidth int, showNames bool, cellWidth int) int {
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
	return applyBackground(lipgloss.NewStyle().Foreground(clistyles.PaletteSlateDark), rowBG).Render(value)
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

func renderValueText(value, valueType string, rowBG color.Color) string {
	if value == "" || clistyles.NoColorEnabled() {
		return value
	}
	style := valueTextStyle(value, valueType)
	return applyBackground(style, rowBG).Render(value)
}

func valueTextStyle(value, valueType string) lipgloss.Style {
	return clistyles.RemoteConfigValueStyle(value, valueType)
}

// ValueTypeKey normalizes a Remote Config value type for display styling.
func ValueTypeKey(valueType string) string {
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

func valueLineHeadWidth(line ValueLine, index, total, labelWidth int, showNames bool) int {
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

// SortedConditionalKeys returns condition names sorted by RC order then name.
func SortedConditionalKeys(items map[string]firebase.RemoteConfigValue, order map[string]int) []string {
	keys := make([]string, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}

	slices.SortFunc(keys, func(leftKey, rightKey string) int {
		left, leftOK := order[leftKey]
		right, rightOK := order[rightKey]
		switch {
		case leftOK && rightOK && left != right:
			if left < right {
				return -1
			}
			return 1
		case leftOK != rightOK:
			if leftOK {
				return -1
			}
			return 1
		default:
			return strfold.Compare(leftKey, rightKey)
		}
	})

	return keys
}

// ValueForJSON maps a display value to its JSON representation.
func ValueForJSON(value string) *string {
	if strings.HasPrefix(value, "(empty ") && strings.HasSuffix(value, ")") {
		return nil
	}
	v := value
	return &v
}
