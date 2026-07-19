package dialog

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/tui/components/buttonbar"
	"github.com/yumauri/fbrcm/tui/components/viewutil"
	"github.com/yumauri/fbrcm/tui/styles"
)

var (
	defaultBorderStyle = lipgloss.NewStyle().Foreground(styles.PaletteError)
	successBorderStyle = lipgloss.NewStyle().Foreground(styles.PaletteSuccess)
	titleStyle         = lipgloss.NewStyle().Bold(true).Foreground(styles.PaletteSlateBright)
	bodyStyle          = lipgloss.NewStyle().Foreground(styles.PaletteSlateBright)
)

func (m Model) View() string {
	if !m.open || m.width <= 0 || m.height <= 0 {
		return ""
	}

	contentWidth := m.contentWidth()
	bodyHeight := m.bodyHeight()
	body := m.bodyLines()
	lines := make([]string, 0, bodyHeight+4)

	start := min(m.scroll, len(body))
	end := min(start+bodyHeight, len(body))
	for _, line := range body[start:end] {
		lines = append(lines, fitBodyLine(line, contentWidth))
	}
	for len(lines) < bodyHeight {
		lines = append(lines, strings.Repeat(" ", contentWidth))
	}
	lines = append(lines, strings.Repeat(" ", contentWidth))
	lines = append(lines, renderBlockAlignedRight(m.renderButtons(), contentWidth)...)

	return renderFrame(m.title, lines, contentWidth, m.scrollbar(), bodyHeight, m.borderStyle())
}

func (m Model) borderStyle() lipgloss.Style {
	if m.tone == ToneSuccess {
		return successBorderStyle
	}
	return defaultBorderStyle
}

func (m Model) Position() (int, int) {
	x, y, _, _ := m.boxGeometry()
	return x, y
}

func (m Model) renderButtons() string {
	return m.buttonBar().View()
}

func (m Model) buttonBar() buttonbar.Model {
	buttons := make([]buttonbar.Button, 0, len(m.buttons))
	for _, button := range m.buttons {
		buttons = append(buttons, buttonbar.Button{Label: button.Label, Variant: button.Variant})
	}
	return buttonbar.New(buttons).SetSelected(m.selected).SetFocused(true)
}

type scrollbarState struct {
	visible    bool
	thumbStart int
	thumbEnd   int
}

func renderFrame(title string, body []string, contentWidth int, scrollbar scrollbarState, bodyHeight int, borderStyle lipgloss.Style) string {
	frameWidth := contentWidth + 5
	top := renderTopBorder(title, frameWidth, borderStyle)
	lines := []string{" " + top + " ", " " + borderStyle.Render("│  ") + strings.Repeat(" ", contentWidth) + borderStyle.Render(" │") + " "}
	for i, line := range body {
		rightEdge := borderStyle.Render("│")
		if scrollbar.visible && i < bodyHeight && i >= scrollbar.thumbStart && i <= scrollbar.thumbEnd {
			rightEdge = styles.ScrollbarThumb.Render("█")
		}
		lines = append(lines, " "+borderStyle.Render("│  ")+padToWidth(line, contentWidth)+borderStyle.Render(" ")+rightEdge+" ")
	}
	lines = append(lines, " "+borderStyle.Render("╰"+strings.Repeat("─", contentWidth+3)+"╯")+" ")
	return strings.Join(lines, "\n")
}

func renderTopBorder(title string, frameWidth int, borderStyle lipgloss.Style) string {
	titleText := viewutil.TruncatePlain(" "+title+" ", max(frameWidth-6, 0))
	left := borderStyle.Render("╭─")
	right := borderStyle.Render("─╮")
	fillWidth := max(frameWidth-printableWidth(left)-printableWidth(right)-printableWidth(titleText), 0)
	return left + titleStyle.Render(titleText) + borderStyle.Render(strings.Repeat("─", fillWidth)) + right
}

func fitBodyLine(line string, width int) string {
	line = ansi.Truncate(line, width, "")
	return padToWidth(bodyStyle.Render(line), width)
}

func (m Model) bodyLines() []string {
	width := m.contentWidth()
	lines := make([]string, 0, len(m.body))
	for _, raw := range m.body {
		raw = strings.ReplaceAll(raw, "\r\n", "\n")
		for logical := range strings.SplitSeq(raw, "\n") {
			if logical == "" {
				lines = append(lines, "")
				continue
			}
			wrapped := ansi.Hardwrap(logical, width, true)
			lines = append(lines, strings.Split(wrapped, "\n")...)
		}
	}
	return lines
}

func alignRight(line string, width int) string {
	if printableWidth(line) >= width {
		return padToWidth(line, width)
	}
	return strings.Repeat(" ", width-printableWidth(line)) + line
}

func renderBlockAlignedRight(block string, width int) []string {
	rawLines := strings.Split(block, "\n")
	out := make([]string, 0, len(rawLines))
	for _, line := range rawLines {
		out = append(out, alignRight(line, width))
	}
	return out
}

func padToWidth(value string, width int) string {
	return value + strings.Repeat(" ", max(width-printableWidth(value), 0))
}

func printableWidth(value string) int {
	return lipgloss.Width(value)
}
