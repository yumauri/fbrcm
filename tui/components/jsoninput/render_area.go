package jsoninput

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/yumauri/fbrcm/tui/styles"
)

func borderStyle(valid bool) lipgloss.Style {
	color := styles.PaletteBlueBright
	if !valid {
		color = styles.PaletteError
	}
	return lipgloss.NewStyle().Foreground(color)
}

func (m Model) renderArea() string {
	height := max(m.area.Height(), 1)
	cursorLine := m.area.Line()
	lineInfo := m.area.LineInfo()
	scrollY := m.area.ScrollYOffset()

	type visualLine struct {
		lineIndex int
		start     int
		end       int
		cursorCol int
	}

	value := m.area.Value()
	lines := strings.Split(value, "\n")
	if len(lines) == 0 {
		lines = []string{""}
	}
	lineStarts := make([]int, len(lines))
	offset := 0
	for i, line := range lines {
		lineStarts[i] = offset
		offset += len([]rune(line)) + 1
	}
	gutter := lineNumberGutter(len(lines))
	contentWidth := max(jsonPopupContentWidth(m.screenW)-gutter, 1)

	visual := make([]visualLine, 0, len(lines))
	for i, line := range lines {
		lineStart := lineStarts[i]
		lineLen := len([]rune(line))
		wrapped := wrapPlainLine(line, contentWidth)
		for seg, part := range wrapped {
			cursorColumn := -1
			if i == cursorLine && seg == lineInfo.RowOffset {
				cursorColumn = lineStart + part.start + min(lineInfo.ColumnOffset, len([]rune(part.text)))
			}
			start := min(part.start, lineLen)
			end := min(part.start+len([]rune(part.text)), lineLen)
			visual = append(visual, visualLine{
				lineIndex: i,
				start:     lineStart + start,
				end:       lineStart + end,
				cursorCol: cursorColumn,
			})
		}
	}
	if len(visual) == 0 {
		visual = append(visual, visualLine{lineIndex: 0, cursorCol: -1})
	}

	visibleRanges := make([]JSONRange, 0, height)
	visibleRangeRows := make([]int, 0, height)
	for row := range height {
		idx := scrollY + row
		if idx < len(visual) {
			line := visual[idx]
			visibleRanges = append(visibleRanges, JSONRange{Start: line.start, End: line.end, CursorCol: line.cursorCol})
			visibleRangeRows = append(visibleRangeRows, row)
		}
	}
	highlighted := highlightJSONRanges(value, visibleRanges)

	rows := make([]string, 0, height)
	highlightedByRow := make(map[int]string, len(highlighted))
	for i, row := range visibleRangeRows {
		highlightedByRow[row] = highlighted[i]
	}
	for row := range height {
		var lineOut strings.Builder
		idx := scrollY + row
		if idx < len(visual) {
			line := visual[idx]
			lineOut.WriteString(renderLineNumber(line.lineIndex+1, len(lines), line.lineIndex == cursorLine))
			lineOut.WriteString(padHighlighted(highlightedByRow[row], contentWidth))
		} else {
			lineOut.WriteString(strings.Repeat(" ", gutter))
			lineOut.WriteString(strings.Repeat(" ", contentWidth))
		}
		rows = append(rows, lineOut.String())
	}
	return strings.Join(rows, "\n")
}
