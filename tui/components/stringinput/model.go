package stringinput

import (
	"fmt"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	tuiconfig "github.com/yumauri/fbrcm/tui/config"
	"github.com/yumauri/fbrcm/tui/styles"
)

type Model struct {
	x         int
	y         int
	minWidth  int
	maxWidth  int
	screenW   int
	screenH   int
	fullWidth bool
	expanded  bool
	text      textinput.Model
	area      textarea.Model
	open      bool
}

func New() Model {
	return Model{text: newTextInput(), area: newTextarea()}
}

// Open opens open for Model and returns the resulting state or error.
func (m Model) Open(x, y, minWidth, maxWidth, screenW, screenH int, value string, fullWidth, expanded bool) (Model, tea.Cmd) {
	m.x = x
	m.y = y
	m.minWidth = max(minWidth, 15)
	m.maxWidth = max(maxWidth, 1)
	m.screenW = screenW
	m.screenH = screenH
	m.fullWidth = fullWidth
	m.expanded = expanded
	m.open = true
	m.text = newTextInput()
	m.text.SetValue(value)
	m.text.CursorEnd()
	m.area = newTextarea()
	m.area.SetValue(value)
	m.resize()
	m.resetAreaCursor()
	if m.expanded {
		return m, m.area.Focus()
	}
	return m, m.text.Focus()
}

// Close closes close for Model and returns the resulting state or error.
func (m Model) Close() Model {
	m.open = false
	m.text.Blur()
	m.area.Blur()
	m.text.SetValue("")
	m.area.SetValue("")
	return m
}

// IsOpen reports open for Model and returns the resulting state or error.
func (m Model) IsOpen() bool {
	return m.open
}

// IsExpanded reports expanded for Model and returns the resulting state or error.
func (m Model) IsExpanded() bool {
	return m.expanded
}

func (m Model) Position() (int, int) {
	if m.expanded {
		return 2, 2
	}
	if m.fullWidth {
		return 0, m.y
	}
	return m.x, m.y
}

func (m Model) Value() string {
	if m.expanded {
		return m.area.Value()
	}
	return m.text.Value()
}

// CanCollapse reports whether collapse for Model and returns the resulting state or error.
func (m Model) CanCollapse() bool {
	return !strings.Contains(m.Value(), "\n")
}

// ToggleExpanded toggles expanded for Model and returns the resulting state or error.
func (m Model) ToggleExpanded() (Model, tea.Cmd) {
	if !m.open {
		return m, nil
	}
	if m.expanded {
		if !m.CanCollapse() {
			return m, nil
		}
		value := m.area.Value()
		cursorCol := m.area.Column()
		m.expanded = false
		m.text = newTextInput()
		m.text.SetValue(value)
		m.text.SetCursor(min(cursorCol, len([]rune(value))))
		m.resize()
		return m, m.text.Focus()
	}
	value := m.text.Value()
	cursorCol := m.text.Position()
	m.expanded = true
	m.area = newTextarea()
	m.area.SetValue(value)
	m.resize()
	m.setAreaCursor(0, cursorCol)
	return m, m.area.Focus()
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.open {
		return m, nil
	}
	if m.expanded {
		var cmd tea.Cmd
		m.area, cmd = m.area.Update(msg)
		return m, cmd
	}
	var cmd tea.Cmd
	m.text, cmd = m.text.Update(msg)
	m.resize()
	return m, cmd
}

func (m Model) View() string {
	if !m.open {
		return ""
	}
	if m.expanded {
		return m.renderExpandedBox()
	}
	return singleBorderStyle.Render(" " + m.text.View())
}

func (m *Model) resize() {
	if m.expanded {
		innerWidth := max(m.screenW-6, 4)
		innerHeight := stringContentHeight(m.screenH)
		gutter := lineNumberGutter(m.area.LineCount())
		m.area.SetWidth(max(innerWidth-gutter, 1))
		m.area.SetHeight(innerHeight)
		return
	}
	innerWidth := max(m.minWidth, lipgloss.Width(m.text.Value())+1)
	if m.fullWidth {
		innerWidth = max(m.screenW-4, 1)
	} else {
		innerWidth = min(innerWidth, max(m.maxWidth-4, 1))
	}
	pos := m.text.Position()
	m.text.SetWidth(innerWidth)
	m.text.SetCursor(pos)
}

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

var (
	singleBorderStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(styles.PaletteBlueBright)
)

func textinputStyles() textinput.Styles {
	inputStyles := textinput.DefaultDarkStyles()
	valueStyle := styles.FilterText
	inputStyles.Focused.Text = valueStyle
	inputStyles.Focused.Prompt = valueStyle
	inputStyles.Focused.Placeholder = valueStyle
	inputStyles.Focused.Suggestion = valueStyle
	inputStyles.Blurred.Text = valueStyle
	inputStyles.Blurred.Prompt = valueStyle
	inputStyles.Blurred.Placeholder = valueStyle
	inputStyles.Blurred.Suggestion = valueStyle
	inputStyles.Cursor.Color = styles.PaletteYellow
	return inputStyles
}

func textareaStyles() textarea.Styles {
	s := textarea.DefaultStyles(true)
	textStyle := styles.FilterText
	s.Focused.Text = textStyle
	s.Focused.Prompt = lipgloss.NewStyle()
	s.Focused.Placeholder = styles.PanelMuted
	s.Focused.LineNumber = lipgloss.NewStyle()
	s.Focused.CursorLineNumber = lipgloss.NewStyle()
	s.Focused.CursorLine = lipgloss.NewStyle()
	s.Focused.EndOfBuffer = lipgloss.NewStyle()
	s.Blurred.Text = textStyle
	s.Blurred.Prompt = lipgloss.NewStyle()
	s.Blurred.Placeholder = styles.PanelMuted
	s.Blurred.LineNumber = lipgloss.NewStyle()
	s.Blurred.CursorLineNumber = lipgloss.NewStyle()
	s.Blurred.CursorLine = lipgloss.NewStyle()
	s.Blurred.EndOfBuffer = lipgloss.NewStyle()
	s.Cursor.Color = styles.PaletteYellow
	return s
}

func newTextInput() textinput.Model {
	input := textinput.New()
	input.Prompt = ""
	input.SetStyles(textinputStyles())
	input.Blur()
	return input
}

func newTextarea() textarea.Model {
	input := textarea.New()
	input.Prompt = ""
	input.ShowLineNumbers = false
	input.EndOfBufferCharacter = ' '
	input.SetStyles(textareaStyles())
	input.Blur()
	return input
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

func (m Model) visualLineCount() int {
	lines := strings.Split(m.area.Value(), "\n")
	if len(lines) == 0 {
		return 1
	}
	gutter := lineNumberGutter(len(lines))
	contentWidth := max(max(m.screenW-6, 4)-gutter, 1)
	count := 0
	for _, line := range lines {
		count += len(wrapLine(line, contentWidth))
	}
	return max(count, 1)
}

func stringContentHeight(screenH int) int {
	return max(screenH-7, 3)
}

func stringHelpText(width int) string {
	m := help.New()
	m.ShortSeparator = " • "
	m.Styles.ShortKey = styles.FilterText
	m.Styles.ShortDesc = styles.PanelMuted
	m.Styles.ShortSeparator = styles.PanelMuted
	m.Styles.Ellipsis = styles.PanelMuted
	m.SetWidth(width)
	return m.ShortHelpView([]key.Binding{
		tuiconfig.Binding(tuiconfig.BlockStringInput, tuiconfig.ActionSave, "save"),
		tuiconfig.Binding(tuiconfig.BlockStringInput, tuiconfig.ActionCancel, "cancel"),
		tuiconfig.Binding(tuiconfig.BlockStringInput, tuiconfig.ActionToggleExpanded, "expand/collapse"),
		tuiconfig.Binding(tuiconfig.BlockStringInput, tuiconfig.ActionCopyValue, "copy"),
	})
}

func renderHelpFooter(text string, width int) string {
	if width <= 0 {
		return ""
	}
	return text + strings.Repeat(" ", max(width-lipgloss.Width(text), 0))
}

type expandedScrollbar struct {
	visible    bool
	thumbStart int
	thumbEnd   int
}

func expandedScrollbarState(total, offset, visible int) expandedScrollbar {
	if visible <= 0 {
		return expandedScrollbar{}
	}
	if total <= visible {
		return expandedScrollbar{}
	}
	thumbHeight := max(1, (visible*visible)/total)
	maxThumbStart := visible - thumbHeight
	maxOffset := max(total-visible, 1)
	thumbStart := (min(offset, maxOffset) * maxThumbStart) / maxOffset
	return expandedScrollbar{
		visible:    true,
		thumbStart: thumbStart,
		thumbEnd:   thumbStart + thumbHeight - 1,
	}
}

func (m *Model) resetAreaCursor() {
	for m.area.Line() > 0 {
		m.area.CursorUp()
	}
	m.area.CursorStart()
}

// setAreaCursor sets set area cursor for Model and returns the resulting state or error.
func (m *Model) setAreaCursor(line, col int) {
	m.resetAreaCursor()
	for m.area.Line() < line {
		m.area.CursorDown()
	}
	m.area.SetCursorColumn(col)
}
