package dialog

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/yumauri/fbrcm/tui/styles"
)

var (
	borderStyle       = lipgloss.NewStyle().Foreground(styles.PaletteError)
	titleStyle        = lipgloss.NewStyle().Bold(true).Foreground(styles.PaletteSlateBright)
	bodyStyle         = lipgloss.NewStyle().Foreground(styles.PaletteSlateBright)
	buttonStyle       = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(styles.PaletteSlateDark).Foreground(styles.PaletteSlateBright).Padding(0, 1)
	publishFocusStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(styles.PaletteError).Foreground(styles.PaletteError).Bold(true).Padding(0, 1)
	blueFocusStyle    = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(styles.PaletteBlueBright).Foreground(styles.PaletteBlueBright).Bold(true).Padding(0, 1)
)

// View handles view for Model and returns the resulting state or error.
func (m Model) View() string {
	if !m.open || m.width <= 0 || m.height <= 0 {
		return ""
	}

	contentWidth := m.contentWidth()
	bodyHeight := m.bodyHeight()
	lines := make([]string, 0, bodyHeight+4)

	start := min(m.scroll, len(m.body))
	end := min(start+bodyHeight, len(m.body))
	for _, line := range m.body[start:end] {
		lines = append(lines, fitBodyLine(line, contentWidth))
	}
	for len(lines) < bodyHeight {
		lines = append(lines, strings.Repeat(" ", contentWidth))
	}
	lines = append(lines, strings.Repeat(" ", contentWidth))
	lines = append(lines, renderBlockAlignedRight(m.renderButtons(), contentWidth)...)

	return renderFrame(m.title, lines, contentWidth, m.scrollbar(), bodyHeight)
}

// Position handles position for Model and returns the resulting state or error.
func (m Model) Position() (int, int) {
	x, y, _, _ := m.boxGeometry()
	return x, y
}

// renderButtons renders render buttons for Model and returns the resulting state or error.
func (m Model) renderButtons() string {
	return lipgloss.JoinHorizontal(lipgloss.Top, appendInterleavedSpaces(m.renderedButtons())...)
}

// renderedButtons renders rendered buttons for Model and returns the resulting state or error.
func (m Model) renderedButtons() []string {
	out := make([]string, 0, len(m.buttons))
	for i, button := range m.buttons {
		label := button.Label
		style := buttonStyle
		if i == m.selected {
			if styles.NoColorEnabled() {
				label = lipgloss.NewStyle().Bold(true).Reverse(true).Render(button.Label)
			} else if button.Variant == ButtonVariantDanger {
				style = publishFocusStyle
			} else {
				style = blueFocusStyle
			}
		}
		out = append(out, style.Render(label))
	}
	return out
}

// appendInterleavedSpaces handles append interleaved spaces and returns the resulting value or error.
func appendInterleavedSpaces(items []string) []string {
	if len(items) <= 1 {
		return items
	}
	out := make([]string, 0, len(items)*2-1)
	for i, item := range items {
		if i > 0 {
			out = append(out, " ")
		}
		out = append(out, item)
	}
	return out
}

// scrollbarState holds scrollbar state state used by the dialog package.
type scrollbarState struct {
	// visible stores visible for scrollbarState.
	visible bool
	// thumbStart stores thumb start for scrollbarState.
	thumbStart int
	// thumbEnd stores thumb end for scrollbarState.
	thumbEnd int
}

// renderFrame renders render frame and returns the resulting value or error.
func renderFrame(title string, body []string, contentWidth int, scrollbar scrollbarState, bodyHeight int) string {
	frameWidth := contentWidth + 5
	top := renderTopBorder(title, frameWidth)
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

// renderTopBorder renders render top border and returns the resulting value or error.
func renderTopBorder(title string, frameWidth int) string {
	titleText := truncatePlain(" "+title+" ", max(frameWidth-6, 0))
	left := borderStyle.Render("╭─")
	right := borderStyle.Render("─╮")
	fillWidth := max(frameWidth-printableWidth(left)-printableWidth(right)-printableWidth(titleText), 0)
	return left + titleStyle.Render(titleText) + borderStyle.Render(strings.Repeat("─", fillWidth)) + right
}

// fitBodyLine handles fit body line and returns the resulting value or error.
func fitBodyLine(line string, width int) string {
	line = ansi.Truncate(line, width, "")
	return padToWidth(bodyStyle.Render(line), width)
}

// alignRight handles align right and returns the resulting value or error.
func alignRight(line string, width int) string {
	if printableWidth(line) >= width {
		return padToWidth(line, width)
	}
	return strings.Repeat(" ", width-printableWidth(line)) + line
}

// renderBlockAlignedRight renders render block aligned right and returns the resulting value or error.
func renderBlockAlignedRight(block string, width int) []string {
	rawLines := strings.Split(block, "\n")
	out := make([]string, 0, len(rawLines))
	for _, line := range rawLines {
		out = append(out, alignRight(line, width))
	}
	return out
}

// padToWidth handles pad to width and returns the resulting value or error.
func padToWidth(value string, width int) string {
	return value + strings.Repeat(" ", max(width-printableWidth(value), 0))
}

// truncatePlain handles truncate plain and returns the resulting value or error.
func truncatePlain(value string, width int) string {
	if width <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= width {
		return value
	}
	return string(runes[:width])
}

// printableWidth handles printable width and returns the resulting value or error.
func printableWidth(value string) int {
	return lipgloss.Width(value)
}
