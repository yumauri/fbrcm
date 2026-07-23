// Package diffview renders generic dictionary diffs as static CLI output.
package diffview

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/rivo/uniseg"

	clistyles "github.com/yumauri/fbrcm/cli/styles"
	"github.com/yumauri/fbrcm/core/dictdiff"
)

const (
	wrappedContextLines = 3
	minimumRenderWidth  = 5
)

type renderedRow struct {
	left, right string
	changed     bool
}

type renderedSideLine struct {
	text    string
	changed bool
}

// Render prints complete, non-interactive dictionary diff blocks. It uses the
// narrowest natural width that fits the content, capped at terminalWidth.
func Render(results []dictdiff.Result, terminalWidth int) string {
	if len(results) == 0 {
		return ""
	}
	width := renderWidth(results, terminalWidth)
	blocks := make([]string, 0, len(results))
	for _, result := range results {
		if len(result.Properties) == 0 {
			continue
		}
		blocks = append(blocks, renderResult(result, width))
	}
	return strings.Join(blocks, "\n\n")
}

func renderResult(result dictdiff.Result, width int) string {
	leftWidth, rightWidth := columnWidths(width)
	lines := []string{strings.TrimRight(renderHeader(result.EntityName, width), " ")}
	for propertyIndex, property := range result.Properties {
		if propertyIndex > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, strings.TrimRight(padRight(headerStyle(property.Name), width), " "))
		for chunkIndex, chunk := range property.Chunks {
			if chunkIndex > 0 {
				lines = append(lines, mutedStyle("⋯"))
			}
			for _, row := range renderChunk(chunk, leftWidth, rightWidth) {
				lines = append(lines, renderSplitLine(row.left, row.right, leftWidth, rightWidth))
			}
		}
	}
	return strings.Join(lines, "\n")
}

func renderChunk(chunk dictdiff.Chunk, leftWidth, rightWidth int) []renderedRow {
	rows := make([]renderedRow, 0)
	for _, row := range chunk.Rows {
		left := renderSide(row.Left, row.Kind, true, leftWidth)
		right := renderSide(row.Right, row.Kind, false, rightWidth)
		height := max(len(left), len(right))
		for line := range height {
			rendered := renderedRow{}
			if line < len(left) {
				rendered.left = left[line].text
				rendered.changed = rendered.changed || left[line].changed
			}
			if line < len(right) {
				rendered.right = right[line].text
				rendered.changed = rendered.changed || right[line].changed
			}
			rows = append(rows, rendered)
		}
	}
	chunks := splitWrappedChunks(rows, wrappedContextLines)
	out := make([]renderedRow, 0, len(rows))
	for index, chunk := range chunks {
		if index > 0 {
			out = append(out, renderedRow{left: mutedStyle("⋯"), right: mutedStyle("⋯")})
		}
		out = append(out, chunk...)
	}
	return out
}

func renderSide(line *dictdiff.Line, kind dictdiff.LineKind, left bool, width int) []renderedSideLine {
	if line == nil {
		return []renderedSideLine{{}}
	}
	segments := line.Segments
	if len(segments) == 0 {
		segmentKind := kind
		if kind == dictdiff.LineChanged {
			segmentKind = dictdiff.LineRemoved
			if !left {
				segmentKind = dictdiff.LineAdded
			}
		}
		segments = []dictdiff.Segment{{Text: line.Text, Kind: segmentKind}}
	}
	return wrapSegments(segments, max(width, 1))
}

func wrapSegments(segments []dictdiff.Segment, width int) []renderedSideLine {
	lines := make([]renderedSideLine, 0, 1)
	var lineBuilder strings.Builder
	lineWidth := 0
	lineChanged := false

	flushLine := func() {
		lines = append(lines, renderedSideLine{text: lineBuilder.String(), changed: lineChanged})
		lineBuilder.Reset()
		lineWidth = 0
		lineChanged = false
	}

	for _, segment := range segments {
		style := segmentStyle(segment.Kind)
		var piece strings.Builder
		flushPiece := func() {
			if piece.Len() == 0 {
				return
			}
			lineBuilder.WriteString(style.Render(piece.String()))
			piece.Reset()
		}
		graphemes := uniseg.NewGraphemes(segment.Text)
		for graphemes.Next() {
			cluster := graphemes.Str()
			clusterWidth := graphemes.Width()
			if lineWidth > 0 && lineWidth+clusterWidth > width {
				flushPiece()
				flushLine()
			}
			piece.WriteString(cluster)
			lineWidth += clusterWidth
			if segment.Kind != dictdiff.LineEqual {
				lineChanged = true
			}
		}
		flushPiece()
	}
	if lineWidth > 0 || len(lines) == 0 {
		flushLine()
	}
	return lines
}

func splitWrappedChunks(rows []renderedRow, budget int) [][]renderedRow {
	chunks := make([][]renderedRow, 0)
	for index := 0; index < len(rows); {
		for index < len(rows) && !rows[index].changed {
			index++
		}
		if index >= len(rows) {
			break
		}
		firstChanged := index
		for index < len(rows) && rows[index].changed {
			index++
		}
		lastChanged := index - 1
		start := max(firstChanged-budget, 0)
		end := min(lastChanged+budget+1, len(rows))
		chunks = append(chunks, rows[start:end])
	}
	if len(chunks) == 0 && len(rows) > 0 {
		return [][]renderedRow{rows}
	}
	return chunks
}

func renderWidth(results []dictdiff.Result, terminalWidth int) int {
	maxSide := 1
	maxFull := 1
	for _, result := range results {
		maxFull = max(maxFull, lipgloss.Width(result.EntityName))
		for _, property := range result.Properties {
			maxFull = max(maxFull, lipgloss.Width(property.Name))
			for _, chunk := range property.Chunks {
				for _, row := range chunk.Rows {
					if row.Left != nil {
						maxSide = max(maxSide, lipgloss.Width(row.Left.Text))
					}
					if row.Right != nil {
						maxSide = max(maxSide, lipgloss.Width(row.Right.Text))
					}
				}
			}
		}
	}
	natural := max(2*maxSide+3, maxFull)
	if terminalWidth <= 0 {
		terminalWidth = natural
	}
	return max(min(natural, terminalWidth), minimumRenderWidth)
}

func columnWidths(width int) (int, int) {
	available := max(width-3, 2)
	left := available / 2
	return left, available - left
}

func renderSplitLine(left, right string, leftWidth, rightWidth int) string {
	line := padRight(left, leftWidth) + " " +
		mutedStyle("│") + " " + padRight(right, rightWidth)
	return strings.TrimRight(line, " ")
}

func renderHeader(text string, width int) string {
	label, value, found := strings.Cut(text, ":")
	if !found {
		return padRight(headerStyle(ansi.Truncate(text, width, "…")), width)
	}
	label += ":"
	labelWidth := lipgloss.Width(label)
	if labelWidth >= width {
		return padRight(mutedStyle(ansi.Truncate(label, width, "…")), width)
	}
	value = ansi.Truncate(value, width-labelWidth, "…")
	return padRight(mutedStyle(label)+headerStyle(value), width)
}

func padRight(text string, width int) string {
	text = ansi.Truncate(text, max(width, 0), "")
	return text + strings.Repeat(" ", max(width-lipgloss.Width(text), 0))
}

func headerStyle(text string) string {
	if clistyles.NoColorEnabled() {
		return text
	}
	return clistyles.PanelText.Bold(true).Render(text)
}

func mutedStyle(text string) string {
	if clistyles.NoColorEnabled() {
		return text
	}
	return clistyles.PanelMuted.Render(text)
}

func segmentStyle(kind dictdiff.LineKind) lipgloss.Style {
	if clistyles.NoColorEnabled() {
		return lipgloss.NewStyle()
	}
	switch kind {
	case dictdiff.LineRemoved:
		return lipgloss.NewStyle().Foreground(clistyles.ColorRemoved)
	case dictdiff.LineAdded:
		return lipgloss.NewStyle().Foreground(clistyles.ColorAdded)
	case dictdiff.LineChanged:
		return lipgloss.NewStyle().Foreground(clistyles.ColorChanged)
	default:
		return clistyles.PanelMuted
	}
}
