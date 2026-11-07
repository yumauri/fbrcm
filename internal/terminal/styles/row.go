package styles

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// RowSpan contains plain text and its foreground/text style.
type RowSpan struct {
	Text  string
	Style lipgloss.Style
}

// RenderBackgroundRow wraps a heading explicitly and paints every physical
// row to width. Styling each span with the same background prevents its ANSI
// resets from leaving gaps, including when a span continues on another row.
// NO_COLOR keeps the text and wrapping without paint or trailing padding.
func RenderBackgroundRow(width int, background color.Color, spans ...RowSpan) string {
	width = max(width, 1)
	var plain, styled strings.Builder
	type segment struct {
		start, end int
		style      lipgloss.Style
	}
	var segments []segment
	for _, span := range spans {
		start := plain.Len()
		plain.WriteString(span.Text)
		segments = append(segments, segment{start, plain.Len(), span.Style.Background(background)})
		styled.WriteString(span.Style.Render(span.Text))
	}
	if NoColorEnabled() {
		return ansi.Hardwrap(plain.String(), width, true)
	}
	if background == nil {
		return ansi.Hardwrap(styled.String(), width, true)
	}
	text := plain.String()
	var rows []string
	cursor := 0
	for line := range strings.SplitSeq(ansi.Hardwrap(text, width, true), "\n") {
		var ranges []lipgloss.Range
		for _, span := range segments {
			begin, end := max(span.start-cursor, 0), min(span.end-cursor, len(line))
			if begin < end {
				ranges = append(ranges, lipgloss.NewRange(lipgloss.Width(line[:begin]), lipgloss.Width(line[:end]), span.style))
			}
		}
		painted := lipgloss.StyleRanges(line, ranges...)
		padding := strings.Repeat(" ", max(width-lipgloss.Width(line), 0))
		rows = append(rows, painted+lipgloss.NewStyle().Background(background).Render(padding))
		cursor += len(line)
		if cursor < len(text) && text[cursor] == '\n' {
			cursor++
		}
	}
	return strings.Join(rows, "\n")
}
