package stringinput

import (
	"fmt"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/tui/components/viewutil"
	tuiconfig "github.com/yumauri/fbrcm/tui/config"
	"github.com/yumauri/fbrcm/tui/styles"
)

var singleBorderStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(styles.PaletteBlueBright)

func (m Model) renderExpandedArea() string {
	width := max(m.screenW-6, 4)
	height := stringContentHeight(m.screenH)
	cursorLine := m.area.Line()
	lineInfo := m.area.LineInfo()
	cursorSegment := lineInfo.RowOffset
	cursorColumn := lineInfo.CharOffset
	scrollY := m.area.ScrollYOffset()

	type visualLine struct {
		text      string
		lineIndex int
		segment   int
	}

	lines := strings.Split(m.area.Value(), "\n")
	if len(lines) == 0 {
		lines = []string{""}
	}
	gutter := lineNumberGutter(len(lines))
	contentWidth := max(width-gutter, 1)

	visual := make([]visualLine, 0, len(lines))
	for i, line := range lines {
		wrapped := wrapLine(line, contentWidth)
		for seg, part := range wrapped {
			visual = append(visual, visualLine{text: part, lineIndex: i, segment: seg})
		}
	}
	if len(visual) == 0 {
		visual = append(visual, visualLine{text: "", lineIndex: 0, segment: 0})
	}

	rows := make([]string, 0, height)
	for row := range height {
		var lineOut strings.Builder
		idx := scrollY + row
		if idx < len(visual) {
			line := visual[idx]
			lineNumber := renderLineNumber(line.lineIndex+1, len(lines), line.lineIndex == cursorLine)
			text := ""
			if line.lineIndex == cursorLine && line.segment == cursorSegment {
				text = renderPlainWithCursor(line.text, cursorColumn, contentWidth)
			} else {
				text = padRendered(styles.FilterText.Render(line.text), contentWidth)
			}
			lineOut.WriteString(lineNumber)
			lineOut.WriteString(text)
		} else {
			lineOut.WriteString(strings.Repeat(" ", gutter))
			lineOut.WriteString(strings.Repeat(" ", contentWidth))
		}
		rows = append(rows, lineOut.String())
	}
	return strings.Join(rows, "\n")
}

func (m Model) renderExpandedBox() string {
	borderStyle := styles.BorderStyle(true)
	body := strings.Split(m.renderExpandedArea(), "\n")
	innerWidth := max(m.screenW-6, 4)
	contentHeight := stringContentHeight(m.screenH)
	scrollbar := expandedScrollbarState(m.visualLineCount(), m.area.ScrollYOffset(), contentHeight)

	lines := []string{borderStyle.Render("╭" + strings.Repeat("─", innerWidth) + "╮")}
	for i := range contentHeight {
		line := ""
		if i < len(body) {
			line = body[i]
		}
		rightEdge := borderStyle.Render("│")
		if scrollbar.visible && i >= scrollbar.thumbStart && i <= scrollbar.thumbEnd {
			rightEdge = styles.ScrollbarThumb.Render("█")
		}
		if line == "" {
			line = strings.Repeat(" ", innerWidth)
		}
		lines = append(lines, borderStyle.Render("│")+line+rightEdge)
	}
	lines = append(lines, borderStyle.Render("│")+renderHelpFooter(stringHelpText(innerWidth), innerWidth)+borderStyle.Render("│"))
	lines = append(lines, borderStyle.Render("╰"+strings.Repeat("─", innerWidth)+"╯"))
	return strings.Join(lines, "\n")
}

func wrapLine(value string, width int) []string {
	if width <= 0 {
		return []string{""}
	}
	if value == "" {
		return []string{""}
	}
	wrapped := ansi.Hardwrap(value, width, true)
	parts := strings.Split(wrapped, "\n")
	if len(parts) == 0 {
		return []string{""}
	}
	return parts
}

func renderPlainWithCursor(value string, cursorCol, width int) string {
	runes := []rune(value)
	if cursorCol < 0 {
		cursorCol = 0
	}
	if cursorCol > len(runes) {
		cursorCol = len(runes)
	}
	before := styles.FilterText.Render(string(runes[:cursorCol]))
	cursorChar := " "
	after := ""
	if cursorCol < len(runes) {
		cursorChar = string(runes[cursorCol])
		after = string(runes[cursorCol+1:])
	}
	rendered := before + cursorStyle().Render(styles.FilterText.Render(cursorChar)) + styles.FilterText.Render(after)
	return padRendered(rendered, width)
}

func renderLineNumber(n, total int, active bool) string {
	digits := max(len(strconv.Itoa(max(total, 1))), 1)
	style := styles.PanelMuted
	if active {
		style = styles.PanelText.Bold(true)
	}
	return style.Render(fmt.Sprintf("%*d ", digits, n))
}

func lineNumberGutter(total int) int {
	return max(len(strconv.Itoa(max(total, 1))), 1) + 1
}

func padRendered(value string, width int) string {
	return value + strings.Repeat(" ", max(width-lipgloss.Width(value), 0))
}

func cursorStyle() lipgloss.Style {
	if styles.NoColorEnabled() {
		return lipgloss.NewStyle().Reverse(true).Bold(true)
	}
	return lipgloss.NewStyle().Background(styles.PaletteYellow).Foreground(styles.PaletteBlueDeep).Bold(true)
}

func stringHelpText(width int) string {
	return viewutil.ShortHelpView(width,
		tuiconfig.Binding(tuiconfig.BlockStringInput, tuiconfig.ActionSave, "save"),
		tuiconfig.Binding(tuiconfig.BlockStringInput, tuiconfig.ActionCancel, "cancel"),
		tuiconfig.Binding(tuiconfig.BlockStringInput, tuiconfig.ActionToggleExpanded, "expand/collapse"),
		tuiconfig.Binding(tuiconfig.BlockStringInput, tuiconfig.ActionCopyValue, "copy"),
	)
}

func renderHelpFooter(text string, width int) string {
	if width <= 0 {
		return ""
	}
	return text + strings.Repeat(" ", max(width-lipgloss.Width(text), 0))
}
