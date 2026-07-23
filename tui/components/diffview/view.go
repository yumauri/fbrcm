package diffview

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/rivo/uniseg"

	"github.com/yumauri/fbrcm/core/dictdiff"
	"github.com/yumauri/fbrcm/tui/components/viewutil"
	tuiconfig "github.com/yumauri/fbrcm/tui/config"
	"github.com/yumauri/fbrcm/tui/styles"
)

const wrappedContextLines = 3

const (
	modalPaddingLeft  = 1
	modalPaddingRight = 1
)

type bodyRow struct {
	text     string
	property int
	header   bool
}

type renderedDiffRow struct {
	text string
	kind dictdiff.LineKind
}

type renderedSideLine struct {
	text    string
	changed bool
}

func (m Model) View() string {
	if !m.open {
		return ""
	}
	contentWidth := m.contentWidth()
	innerWidth := modalInnerWidth(contentWidth)
	border := styles.BorderStyle(true)
	rows := m.bodyRows(contentWidth)
	start := min(m.offset, len(rows))
	end := min(start+m.bodyHeight(), len(rows))
	scrollbar := viewutil.ScrollbarState(len(rows), start, m.bodyHeight())

	lines := []string{m.titleLine(innerWidth, border)}
	lines = append(lines,
		border.Render("│")+modalContentLine(m.entityHeader(contentWidth), contentWidth)+border.Render("│"),
		border.Render("│")+modalContentLine(m.dictionaryHeader(contentWidth), contentWidth)+border.Render("│"),
		border.Render("├"+strings.Repeat("─", innerWidth)+"┤"),
	)
	for row := range m.bodyHeight() {
		text := ""
		if start+row < end {
			text = rows[start+row].text
		}
		rightEdge := border.Render("│")
		if scrollbar.Visible && row >= scrollbar.ThumbStart && row <= scrollbar.ThumbEnd {
			rightEdge = styles.ScrollbarThumb.Render("█")
		}
		lines = append(lines, border.Render("│")+modalContentLine(text, contentWidth)+rightEdge)
	}
	lines = append(lines,
		border.Render("│")+modalContentLine(m.footer(contentWidth), contentWidth)+border.Render("│"),
		border.Render("╰"+strings.Repeat("─", innerWidth)+"╯"),
	)
	return strings.Join(lines, "\n")
}

// BodyView renders the complete scrollable body without the modal frame.
func (m Model) BodyView(width int) string {
	rows := m.bodyRows(width)
	lines := make([]string, len(rows))
	for index, row := range rows {
		lines[index] = row.text
	}
	return strings.Join(lines, "\n")
}

func (m Model) titleLine(innerWidth int, border lipgloss.Style) string {
	rendered, titleWidth := styles.PanelHeaderTab("", "Diff", true, true, max(innerWidth-2, 0))
	fill := max(innerWidth-titleWidth-1, 0)
	return border.Render("╭─") + rendered + border.Render(strings.Repeat("─", fill)+"╮")
}

func (m Model) entityHeader(width int) string {
	return styledHeader(m.result.EntityName, width)
}

func styledHeader(text string, width int) string {
	label, value, found := strings.Cut(text, ":")
	if !found {
		text = ansi.Truncate(text, width, "…")
		return viewutil.PadRight(styles.PanelText.Bold(true).Render(text), width)
	}
	label += ":"
	labelWidth := lipgloss.Width(label)
	if labelWidth >= width {
		text := ansi.Truncate(label, width, "…")
		return viewutil.PadRight(styles.PanelMuted.Render(text), width)
	}
	value = ansi.Truncate(value, width-labelWidth, "…")
	rendered := styles.PanelMuted.Render(label) +
		styles.PanelText.Bold(true).Render(value)
	return viewutil.PadRight(rendered, width)
}

func (m Model) dictionaryHeader(width int) string {
	leftWidth, rightWidth := columnWidths(width)
	left := styledHeader(m.result.LeftName, leftWidth)
	right := styledHeader(m.result.RightName, rightWidth)
	return left + styles.PanelBorderInactive.Render("│") + " " + right
}

func (m Model) footer(width int) string {
	return viewutil.ShortHelpView(width,
		viewutil.HelpBinding(
			tuiconfig.Label(tuiconfig.BlockDiffView, tuiconfig.ActionUp)+"/"+
				tuiconfig.Label(tuiconfig.BlockDiffView, tuiconfig.ActionDown),
			"property",
		),
		tuiconfig.Binding(tuiconfig.BlockDiffView, tuiconfig.ActionToggle, "collapse"),
		tuiconfig.Binding(tuiconfig.BlockDiffView, tuiconfig.ActionPageDown, "scroll"),
		tuiconfig.Binding(tuiconfig.BlockDiffView, tuiconfig.ActionClose, "close"),
	)
}

func (m Model) bodyRows(width int) []bodyRow {
	if len(m.result.Properties) == 0 {
		return []bodyRow{{text: styles.DetailsEmptyValue.Render("No differences.")}}
	}
	rows := make([]bodyRow, 0)
	for propertyIndex, property := range m.result.Properties {
		if propertyIndex > 0 {
			rows = append(rows, bodyRow{property: propertyIndex})
		}
		rows = append(rows, bodyRow{
			text:     m.propertyHeader(property, propertyIndex == m.cursor, width),
			property: propertyIndex,
			header:   true,
		})
		if m.collapsed[property.Name] {
			continue
		}
		for chunkIndex, chunk := range property.Chunks {
			if chunkIndex > 0 {
				rows = append(rows, bodyRow{
					text:     styles.DetailsEmptyValue.Render(viewutil.PadRight("⋯", width)),
					property: propertyIndex,
				})
			}
			for _, row := range renderChunk(chunk, width) {
				rows = append(rows, bodyRow{text: row, property: propertyIndex})
			}
		}
	}
	return rows
}

func (m Model) propertyHeader(property dictdiff.Property, selected bool, width int) string {
	marker := "▾"
	if m.collapsed[property.Name] {
		marker = "▸"
	}
	text := marker + " " + property.Name
	text = viewutil.PadRight(ansi.Truncate(text, width, "…"), width)
	style := styles.PanelText.Bold(true)
	if selected {
		selection := styles.TreeItemSelectionStyle()
		style = style.Background(selection.GetBackground())
	}
	return style.Render(text)
}

func renderChunk(chunk dictdiff.Chunk, width int) []string {
	leftWidth, rightWidth := columnWidths(width)
	rows := make([]renderedDiffRow, 0)
	for _, row := range chunk.Rows {
		left := renderSide(row.Left, row.Kind, true, leftWidth)
		right := renderSide(row.Right, row.Kind, false, rightWidth)
		height := max(len(left), len(right))
		for line := range height {
			leftLine, rightLine := "", ""
			changed := false
			if line < len(left) {
				leftLine = left[line].text
				changed = changed || left[line].changed
			}
			if line < len(right) {
				rightLine = right[line].text
				changed = changed || right[line].changed
			}
			kind := dictdiff.LineEqual
			if changed {
				kind = dictdiff.LineChanged
			}
			rows = append(rows, renderedDiffRow{
				text: viewutil.PadRight(leftLine, leftWidth) +
					styles.PanelBorderInactive.Render("│") +
					" " +
					viewutil.PadRight(rightLine, rightWidth),
				kind: kind,
			})
		}
	}
	chunks := splitWrappedChunks(rows, wrappedContextLines)
	out := make([]string, 0, len(rows))
	for chunkIndex, visualChunk := range chunks {
		if chunkIndex > 0 {
			out = append(out, styles.DetailsEmptyValue.Render(viewutil.PadRight("⋯", width)))
		}
		for _, row := range visualChunk {
			out = append(out, row.text)
		}
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

func segmentStyle(kind dictdiff.LineKind) lipgloss.Style {
	switch kind {
	case dictdiff.LineRemoved:
		return lipgloss.NewStyle().Foreground(styles.PaletteRemoved)
	case dictdiff.LineAdded:
		return lipgloss.NewStyle().Foreground(styles.PaletteAdded)
	case dictdiff.LineChanged:
		return lipgloss.NewStyle().Foreground(styles.PaletteChanged)
	default:
		return styles.PanelMuted
	}
}

func splitWrappedChunks(rows []renderedDiffRow, budget int) [][]renderedDiffRow {
	chunks := make([][]renderedDiffRow, 0)
	for index := 0; index < len(rows); {
		for index < len(rows) && rows[index].kind == dictdiff.LineEqual {
			index++
		}
		if index >= len(rows) {
			break
		}
		firstChanged := index
		for index < len(rows) && rows[index].kind != dictdiff.LineEqual {
			index++
		}
		lastChanged := index - 1
		start := max(firstChanged-budget, 0)
		end := min(lastChanged+budget+1, len(rows))
		chunks = append(chunks, rows[start:end])
	}
	if len(chunks) == 0 && len(rows) > 0 {
		return [][]renderedDiffRow{rows}
	}
	return chunks
}

func columnWidths(width int) (int, int) {
	left := max((width-2)/2, 1)
	return left, max(width-left-2, 1)
}

func (m Model) contentWidth() int {
	return max(max(m.screenW-6, 4)-modalPaddingLeft-modalPaddingRight, 1)
}

func (m Model) bodyHeight() int {
	return max(m.screenH-10, 3)
}

func modalInnerWidth(contentWidth int) int {
	return modalPaddingLeft + max(contentWidth, 0) + modalPaddingRight
}

func modalContentLine(content string, contentWidth int) string {
	contentWidth = max(contentWidth, 0)
	content = ansi.Truncate(content, contentWidth, "")
	content = viewutil.PadRight(content, contentWidth)
	return strings.Repeat(" ", modalPaddingLeft) + content + strings.Repeat(" ", modalPaddingRight)
}
