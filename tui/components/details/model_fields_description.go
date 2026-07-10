package details

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/tui/components/viewutil"
	"github.com/yumauri/fbrcm/tui/styles"
)

func (m Model) renderGroupField() string {
	value := m.groupLabel
	if m.activeField == fieldGroup {
		return selectedDropdownFieldStyle().Render(value)
	}
	return groupValueStyle.Render(value)
}

func (m Model) renderNameField() string {
	if m.activeField == fieldName {
		return styles.FilterText.Render(m.nameInput.View())
	}
	return parameterKeyStyle.Render(strings.TrimSpace(m.nameInput.Value()))
}

func (m Model) renderTypeField() string {
	value := m.typeValue
	if m.activeField == fieldType {
		return selectedDropdownFieldStyle().Render(value)
	}
	return styles.PanelText.Render(value)
}

func (m Model) renderDescriptionField() string {
	width := m.descriptionTextWidth()
	if m.activeField == fieldDescription {
		return m.renderActiveDescription(width)
	}
	rawValue := m.descInput.Value()
	value := rawValue
	if value == "" {
		value = "No description."
	}
	segments := descriptionWrapSegments(value, width)
	lines := make([]string, 0, len(segments))
	for _, segment := range segments {
		if rawValue == "" {
			lines = append(lines, styles.PanelMuted.Italic(true).Render(segment.text))
		} else {
			lines = append(lines, styles.PanelText.Render(segment.text))
		}
	}
	return strings.Join(lines, "\n")
}

func (m *Model) resizeDescriptionInput() {
	inputWidth := m.descriptionTextWidth()
	m.descInput.SetWidth(inputWidth)
	m.descInput.SetHeight(m.descriptionVisualHeightForWidth(inputWidth))
}

func (m *Model) normalizeDescriptionInput() {
	value := singleLineValue(m.descInput.Value())
	pos := m.descInput.Column()
	maxPos := len([]rune(value))
	if value != m.descInput.Value() {
		m.descInput.SetValue(value)
	}
	if pos > maxPos {
		m.descInput.SetCursorColumn(maxPos)
	}
}

func (m Model) descriptionVisualHeight() int {
	width := m.descriptionTextWidth()
	return m.descriptionVisualHeightForWidth(width)
}

func (m Model) descriptionTextWidth() int {
	return max(m.width-6, 1)
}

func (m Model) descriptionVisualHeightForWidth(width int) int {
	value := m.descInput.Value()
	if value == "" {
		value = "No description."
	}
	return max(len(descriptionWrapSegments(value, width)), 1)
}

func (m Model) renderActiveDescription(width int) string {
	value := m.descInput.Value()
	segments := descriptionWrapSegments(value, width)
	cursor := m.descInput.Column()
	lines := make([]string, 0, len(segments))
	cursorLine, cursorCol := wrappedCursorPosition(segments, cursor)
	for i, segment := range segments {
		if i == cursorLine {
			lines = append(lines, renderWithCursor(segment.text, cursorCol, width))
		} else {
			lines = append(lines, styles.FilterText.Render(viewutil.PadRight(segment.text, width)))
		}
	}
	return strings.Join(lines, "\n")
}

type descriptionSegment struct {
	text string
	// start, end store start end values for descriptionSegment.
	start, end int
}

func descriptionWrapSegments(value string, width int) []descriptionSegment {
	if width <= 0 {
		return []descriptionSegment{{text: ""}}
	}
	if value == "" {
		return []descriptionSegment{{text: ""}}
	}
	wrapped := ansi.Wordwrap(value, width, " ")
	parts := strings.Split(wrapped, "\n")
	if len(parts) == 0 {
		return []descriptionSegment{{text: ""}}
	}
	valueRunes := []rune(value)
	pos := 0
	segments := make([]descriptionSegment, 0, len(parts))
	for _, part := range parts {
		partRunes := []rune(part)
		for len(partRunes) > 0 && pos < len(valueRunes) && valueRunes[pos] != partRunes[0] {
			pos++
		}
		start := pos
		for _, r := range partRunes {
			for pos < len(valueRunes) && valueRunes[pos] != r {
				pos++
			}
			if pos < len(valueRunes) {
				pos++
			}
		}
		segments = append(segments, descriptionSegment{text: part, start: start, end: pos})
	}
	for pos < len(valueRunes) && valueRunes[pos] == ' ' {
		if len(segments) == 0 || lipgloss.Width(segments[len(segments)-1].text) >= width {
			segments = append(segments, descriptionSegment{text: "", start: pos, end: pos})
		}
		last := &segments[len(segments)-1]
		last.text += " "
		pos++
		last.end = pos
	}
	return segments
}

func wrappedOffsetForClick(segments []descriptionSegment, line, col int) int {
	if len(segments) == 0 {
		return 0
	}
	line = min(max(line, 0), len(segments)-1)
	segment := segments[line]
	return segment.start + min(max(col, 0), len([]rune(segment.text)))
}

func wrappedCursorPosition(segments []descriptionSegment, cursor int) (int, int) {
	if len(segments) == 0 {
		return 0, 0
	}
	for i, segment := range segments {
		if cursor >= segment.start && cursor <= segment.end {
			return i, min(max(cursor-segment.start, 0), len([]rune(segment.text)))
		}
		if cursor < segment.start {
			return i, 0
		}
	}
	last := len(segments) - 1
	return last, len([]rune(segments[last].text))
}

func renderWithCursor(value string, cursorCol, width int) string {
	runes := []rune(value)
	cursorCol = min(max(cursorCol, 0), len(runes))
	before := styles.FilterText.Render(string(runes[:cursorCol]))
	cursorChar := " "
	after := ""
	if cursorCol < len(runes) {
		cursorChar = string(runes[cursorCol])
		after = string(runes[cursorCol+1:])
	}
	rendered := before + descriptionCursorStyle().Render(styles.FilterText.Render(cursorChar)) + styles.FilterText.Render(after)
	return viewutil.PadRight(rendered, width)
}

func descriptionCursorStyle() lipgloss.Style {
	if styles.NoColorEnabled() {
		return lipgloss.NewStyle().Reverse(true).Bold(true)
	}
	return lipgloss.NewStyle().Background(styles.PaletteYellow).Foreground(styles.PaletteBlueDeep).Bold(true)
}

func singleLineValue(value string) string {
	return strings.Join(strings.FieldsFunc(value, func(r rune) bool {
		return r == '\n' || r == '\r'
	}), " ")
}
