package jsoninput

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	rw "github.com/mattn/go-runewidth"
	"github.com/rivo/uniseg"

	tuiconfig "github.com/yumauri/fbrcm/tui/config"
	"github.com/yumauri/fbrcm/tui/styles"
)

type jsonContainer struct {
	kind           rune
	expectingKey   bool
	expectingValue bool
}

// JSONRange holds JSON highlight range state used by the jsoninput package.
type JSONRange struct {
	// Start stores start for JSONRange.
	Start int
	// End stores end for JSONRange.
	End int
	// CursorCol stores cursor col for JSONRange.
	CursorCol int
}

// Model holds model state used by the jsoninput package.
type Model struct {
	// screenW stores screen w for Model.
	screenW int
	// screenH stores screen h for Model.
	screenH int
	// area stores area for Model.
	area textarea.Model
	// open stores open for Model.
	open bool
}

// New constructs new and returns the resulting value or error.
func New() Model {
	return Model{area: newTextarea()}
}

// Open opens open for Model and returns the resulting state or error.
func (m Model) Open(screenW, screenH int, value string) (Model, tea.Cmd) {
	m.screenW = screenW
	m.screenH = screenH
	m.area = newTextarea()
	m.area.SetValue(prettyJSON(value))
	m.resize()
	m.resetAreaCursor()
	m.open = true
	return m, m.area.Focus()
}

// Close closes close for Model and returns the resulting state or error.
func (m Model) Close() Model {
	m.open = false
	m.area.Blur()
	m.area.SetValue("")
	return m
}

// IsOpen reports open for Model and returns the resulting state or error.
func (m Model) IsOpen() bool {
	return m.open
}

// Position handles position for Model and returns the resulting state or error.
func (m Model) Position() (int, int) {
	return 2, 2
}

// Value handles value for Model and returns the resulting state or error.
func (m Model) Value() string {
	return m.area.Value()
}

// Valid handles valid for Model and returns the resulting state or error.
func (m Model) Valid() bool {
	return json.Valid([]byte(m.area.Value()))
}

// CompactedValue handles compacted value for Model and returns the resulting state or error.
func (m Model) CompactedValue() (string, bool) {
	if !m.Valid() {
		return "", false
	}
	var buf bytes.Buffer
	if err := json.Compact(&buf, []byte(m.area.Value())); err != nil {
		return "", false
	}
	return buf.String(), true
}

// PrettyValue handles pretty value for Model and returns the resulting state or error.
func (m Model) PrettyValue() string {
	return prettyJSON(m.area.Value())
}

// Reformat handles reformat for Model and returns the resulting state or error.
func (m Model) Reformat() Model {
	if !m.Valid() {
		return m
	}
	line := m.area.Line()
	col := m.area.Column()
	m.area.SetValue(prettyJSON(m.area.Value()))
	m.resize()
	m.setAreaCursor(line, col)
	return m
}

// Update updates update for Model and returns the resulting state or error.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.open {
		return m, nil
	}
	var cmd tea.Cmd
	m.area, cmd = m.area.Update(msg)
	return m, cmd
}

// View handles view for Model and returns the resulting state or error.
func (m Model) View() string {
	if !m.open {
		return ""
	}
	return m.renderBox()
}

// resize handles resize for Model and returns the resulting state or error.
func (m *Model) resize() {
	innerWidth := max(m.screenW-6, 4)
	innerHeight := jsonContentHeight(m.screenH)
	gutter := lineNumberGutter(m.area.LineCount())
	m.area.SetWidth(max(innerWidth-gutter, 1))
	m.area.SetHeight(innerHeight)
}

// borderStyle handles border style and returns the resulting value or error.
func borderStyle(valid bool) lipgloss.Style {
	color := styles.PaletteBlueBright
	if !valid {
		color = styles.PaletteError
	}
	return lipgloss.NewStyle().Foreground(color)
}

// textareaStyles handles textarea styles and returns the resulting value or error.
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

// newTextarea constructs new textarea and returns the resulting value or error.
func newTextarea() textarea.Model {
	input := textarea.New()
	input.Prompt = ""
	input.ShowLineNumbers = false
	input.EndOfBufferCharacter = ' '
	input.SetStyles(textareaStyles())
	input.Blur()
	return input
}

// prettyJSON handles pretty json and returns the resulting value or error.
func prettyJSON(value string) string {
	if !json.Valid([]byte(value)) {
		return value
	}
	var out bytes.Buffer
	if err := json.Indent(&out, []byte(value), "", "  "); err != nil {
		return value
	}
	return out.String()
}

// renderArea renders render area for Model and returns the resulting state or error.
func (m Model) renderArea() string {
	height := max(m.area.Height(), 1)
	cursorLine := m.area.Line()
	lineInfo := m.area.LineInfo()
	scrollY := m.area.ScrollYOffset()

	// visualLine holds visual line state used by the jsoninput package.
	type visualLine struct {
		// lineIndex stores line index for visualLine.
		lineIndex int
		// start stores start for visualLine.
		start int
		// end stores end for visualLine.
		end int
		// cursorCol stores cursor col for visualLine.
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
	contentWidth := max(max(m.screenW-6, 4)-gutter, 1)

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

// plainWrappedLine holds plain wrapped line state used by the jsoninput package.
type plainWrappedLine struct {
	// text stores text for plainWrappedLine.
	text string
	// start stores start for plainWrappedLine.
	start int
}

// wrapPlainLine handles wrap plain line and returns the resulting value or error.
func wrapPlainLine(value string, width int) []plainWrappedLine {
	if width <= 0 {
		return []plainWrappedLine{{}}
	}
	wrapped := textareaWrap([]rune(value), width)
	out := make([]plainWrappedLine, 0, len(wrapped))
	start := 0
	for _, part := range wrapped {
		text := string(part)
		if uniseg.StringWidth(text) > width {
			text = strings.TrimSuffix(text, " ")
		}
		out = append(out, plainWrappedLine{text: text, start: start})
		start += len(part)
	}
	if len(out) == 0 {
		return []plainWrappedLine{{}}
	}
	return out
}

// textareaWrap handles textarea wrap and returns the resulting value or error.
func textareaWrap(runes []rune, width int) [][]rune {
	lines := [][]rune{{}}
	word := []rune{}
	row := 0
	spaces := 0

	for _, r := range runes {
		if unicode.IsSpace(r) {
			spaces++
		} else {
			word = append(word, r)
		}

		if spaces > 0 {
			if uniseg.StringWidth(string(lines[row]))+uniseg.StringWidth(string(word))+spaces > width {
				row++
				lines = append(lines, []rune{})
			}
			lines[row] = append(lines[row], word...)
			lines[row] = append(lines[row], repeatSpaces(spaces)...)
			spaces = 0
			word = nil
			continue
		}

		lastCharLen := rw.RuneWidth(word[len(word)-1])
		if uniseg.StringWidth(string(word))+lastCharLen > width {
			if len(lines[row]) > 0 {
				row++
				lines = append(lines, []rune{})
			}
			lines[row] = append(lines[row], word...)
			word = nil
		}
	}

	if uniseg.StringWidth(string(lines[row]))+uniseg.StringWidth(string(word))+spaces >= width {
		lines = append(lines, []rune{})
		lines[row+1] = append(lines[row+1], word...)
		spaces++
		lines[row+1] = append(lines[row+1], repeatSpaces(spaces)...)
	} else {
		lines[row] = append(lines[row], word...)
		spaces++
		lines[row] = append(lines[row], repeatSpaces(spaces)...)
	}

	return lines
}

// repeatSpaces handles repeat spaces and returns the resulting value or error.
func repeatSpaces(n int) []rune {
	return []rune(strings.Repeat(" ", n))
}

// highlightJSONRange handles highlight jsonrange and returns the resulting value or error.
func highlightJSONRange(value string, start, end, cursorCol int) string {
	ranges := highlightJSONRanges(value, []JSONRange{{Start: start, End: end, CursorCol: cursorCol}})
	if len(ranges) == 0 {
		return ""
	}
	return ranges[0]
}

// highlightJSONRanges handles highlight jsonranges and returns the resulting value or error.
func highlightJSONRanges(value string, ranges []JSONRange) []string {
	if len(ranges) == 0 {
		return nil
	}
	runes := []rune(value)
	normalized := make([]JSONRange, len(ranges))
	for i, r := range ranges {
		if r.Start < 0 {
			r.Start = 0
		}
		if r.End < r.Start {
			r.End = r.Start
		}
		if r.Start > len(runes) {
			r.Start = len(runes)
		}
		if r.End > len(runes) {
			r.End = len(runes)
		}
		normalized[i] = r
	}

	builders := make([]strings.Builder, len(normalized))
	rangeIndex := 0
	advanceRange := func(pos int) {
		for rangeIndex < len(normalized) && pos >= normalized[rangeIndex].End {
			rangeIndex++
		}
	}
	writeVisible := func(pos int, rendered string) {
		advanceRange(pos)
		if rangeIndex >= len(normalized) {
			return
		}
		r := normalized[rangeIndex]
		if pos >= r.Start && pos < r.End {
			builders[rangeIndex].WriteString(rendered)
		}
	}

	inString := false
	escaped := false
	stringIsKey := false
	stack := make([]jsonContainer, 0, 8)
	lit := strings.Builder{}
	litStart := 0
	flushLit := func() {
		if lit.Len() == 0 {
			return
		}
		token := lit.String()
		tokenRunes := []rune(token)
		style := jsonTokenStyle(token)
		for offset, r := range tokenRunes {
			pos := litStart + offset
			advanceRange(pos)
			rendered := style.Render(string(r))
			if rangeIndex < len(normalized) && pos == normalized[rangeIndex].CursorCol {
				rendered = cursorStyle().Render(rendered)
			}
			writeVisible(pos, rendered)
		}
		lit.Reset()
	}
	writeStyledRune := func(pos int, r rune, style lipgloss.Style) {
		advanceRange(pos)
		rendered := style.Render(string(r))
		if rangeIndex < len(normalized) && pos == normalized[rangeIndex].CursorCol {
			rendered = cursorStyle().Render(rendered)
		}
		writeVisible(pos, rendered)
	}
	currentStringIsKey := func() bool {
		if len(stack) == 0 {
			return false
		}
		top := stack[len(stack)-1]
		return top.kind == '{' && top.expectingKey
	}
	markValueStarted := func() {
		if len(stack) == 0 {
			return
		}
		top := &stack[len(stack)-1]
		if top.kind == '{' && top.expectingValue {
			top.expectingValue = false
		}
	}

	for i, r := range runes {
		if rangeIndex >= len(normalized) {
			break
		}
		if inString {
			if escaped {
				writeStyledRune(i, r, jsonStringContextStyle(stringIsKey))
				escaped = false
				continue
			}
			if r == '\\' {
				writeStyledRune(i, r, jsonStringContextStyle(stringIsKey))
				escaped = true
				continue
			}
			writeStyledRune(i, r, jsonStringContextStyle(stringIsKey))
			if r == '"' {
				inString = false
			}
			continue
		}
		switch r {
		case '"':
			flushLit()
			inString = true
			stringIsKey = currentStringIsKey()
			if stringIsKey {
				stack[len(stack)-1].expectingKey = false
			} else {
				markValueStarted()
			}
			writeStyledRune(i, r, jsonStringContextStyle(stringIsKey))
		case '{', '}', '[', ']', ':', ',':
			flushLit()
			switch r {
			case '{':
				markValueStarted()
				stack = append(stack, jsonContainer{kind: '{', expectingKey: true})
			case '[':
				markValueStarted()
				stack = append(stack, jsonContainer{kind: '['})
			case '}', ']':
				if len(stack) > 0 {
					stack = stack[:len(stack)-1]
				}
			case ':':
				if len(stack) > 0 && stack[len(stack)-1].kind == '{' {
					stack[len(stack)-1].expectingValue = true
				}
			case ',':
				if len(stack) > 0 && stack[len(stack)-1].kind == '{' {
					stack[len(stack)-1].expectingKey = true
					stack[len(stack)-1].expectingValue = false
				}
			}
			writeStyledRune(i, r, jsonPunctuationStyle())
		case ' ', '\n', '\t', '\r':
			flushLit()
			writeStyledRune(i, r, jsonDefaultStyle())
		default:
			if lit.Len() == 0 {
				litStart = i
				markValueStarted()
			}
			lit.WriteRune(r)
		}
	}
	flushLit()
	out := make([]string, len(builders))
	for i := range builders {
		r := normalized[i]
		if r.CursorCol == r.End || (r.CursorCol == len(runes) && r.CursorCol >= r.Start && r.CursorCol <= r.End) {
			builders[i].WriteString(cursorStyle().Render(" "))
		}
		out[i] = builders[i].String()
	}
	return out
}

// renderLineNumber renders render line number and returns the resulting value or error.
func renderLineNumber(n, total int, active bool) string {
	digits := max(len(strconv.Itoa(max(total, 1))), 1)
	style := styles.PanelMuted
	if active {
		style = styles.PanelText.Bold(true)
	}
	return style.Render(fmt.Sprintf("%*d ", digits, n))
}

// lineNumberGutter handles line number gutter and returns the resulting value or error.
func lineNumberGutter(total int) int {
	return max(len(strconv.Itoa(max(total, 1))), 1) + 1
}

// padHighlighted handles pad highlighted and returns the resulting value or error.
func padHighlighted(value string, width int) string {
	return value + strings.Repeat(" ", max(width-lipgloss.Width(value), 0))
}

// cursorStyle handles cursor style and returns the resulting value or error.
func cursorStyle() lipgloss.Style {
	if styles.NoColorEnabled() {
		return lipgloss.NewStyle().Reverse(true).Bold(true)
	}
	return lipgloss.NewStyle().Background(styles.PaletteYellow).Foreground(styles.PaletteBlueDeep).Bold(true)
}

// highlightJSONVisible handles highlight jsonvisible and returns the resulting value or error.
func highlightJSONVisible(value string) string {
	var out strings.Builder
	inString := false
	escaped := false
	stringIsKey := false
	stack := make([]jsonContainer, 0, 8)
	lit := strings.Builder{}
	flushLit := func() {
		if lit.Len() == 0 {
			return
		}
		token := lit.String()
		style := jsonTokenStyle(token)
		out.WriteString(style.Render(token))
		lit.Reset()
	}
	currentStringIsKey := func() bool {
		if len(stack) == 0 {
			return false
		}
		top := stack[len(stack)-1]
		return top.kind == '{' && top.expectingKey
	}
	markValueStarted := func() {
		if len(stack) == 0 {
			return
		}
		top := &stack[len(stack)-1]
		if top.kind == '{' && top.expectingValue {
			top.expectingValue = false
		}
	}
	for _, r := range value {
		if inString {
			if escaped {
				out.WriteString(highlightJSONRune(r, true, stringIsKey))
				escaped = false
				continue
			}
			if r == '\\' {
				out.WriteString(highlightJSONRune(r, true, stringIsKey))
				escaped = true
				continue
			}
			out.WriteString(highlightJSONRune(r, true, stringIsKey))
			if r == '"' {
				inString = false
			}
			continue
		}
		switch r {
		case '"':
			flushLit()
			inString = true
			stringIsKey = currentStringIsKey()
			if stringIsKey {
				stack[len(stack)-1].expectingKey = false
			} else {
				markValueStarted()
			}
			out.WriteString(jsonStringContextStyle(stringIsKey).Render(`"`))
		case '{', '}', '[', ']', ':', ',':
			flushLit()
			switch r {
			case '{':
				markValueStarted()
				stack = append(stack, jsonContainer{kind: '{', expectingKey: true})
			case '[':
				markValueStarted()
				stack = append(stack, jsonContainer{kind: '['})
			case '}', ']':
				if len(stack) > 0 {
					stack = stack[:len(stack)-1]
				}
			case ':':
				if len(stack) > 0 && stack[len(stack)-1].kind == '{' {
					stack[len(stack)-1].expectingValue = true
				}
			case ',':
				if len(stack) > 0 && stack[len(stack)-1].kind == '{' {
					stack[len(stack)-1].expectingKey = true
					stack[len(stack)-1].expectingValue = false
				}
			}
			out.WriteString(highlightJSONRune(r, false, false))
		case ' ', '\n', '\t', '\r':
			flushLit()
			out.WriteRune(r)
		default:
			if lit.Len() == 0 {
				markValueStarted()
			}
			lit.WriteRune(r)
		}
	}
	flushLit()
	return out.String()
}

// HighlightJSONVisible handles highlight jsonvisible and returns the resulting value or error.
func HighlightJSONVisible(value string) string {
	return highlightJSONVisible(value)
}

// HighlightJSONRange handles highlight jsonrange and returns the resulting value or error.
func HighlightJSONRange(value string, start, end, cursorCol int) string {
	return highlightJSONRange(value, start, end, cursorCol)
}

// HighlightJSONRanges handles highlight jsonranges and returns the resulting value or error.
func HighlightJSONRanges(value string, ranges []JSONRange) []string {
	return highlightJSONRanges(value, ranges)
}

// highlightJSONRune handles highlight jsonrune and returns the resulting value or error.
func highlightJSONRune(r rune, inString, stringIsKey bool) string {
	switch {
	case inString && stringIsKey:
		return jsonKeyStyle().Render(string(r))
	case inString:
		return jsonStringStyle().Render(string(r))
	case r == '"':
		return jsonStringContextStyle(stringIsKey).Render(`"`)
	case r == '{' || r == '}' || r == '[' || r == ']' || r == ':' || r == ',':
		return jsonPunctuationStyle().Render(string(r))
	default:
		return jsonDefaultStyle().Render(string(r))
	}
}

// jsonDefaultStyle handles json default style and returns the resulting value or error.
func jsonDefaultStyle() lipgloss.Style {
	return styles.PanelText
}

// jsonPunctuationStyle handles json punctuation style and returns the resulting value or error.
func jsonPunctuationStyle() lipgloss.Style {
	return styles.PanelText
}

// jsonKeyStyle handles json key style and returns the resulting value or error.
func jsonKeyStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(styles.ConditionLipglossColor("CYAN"))
}

// jsonStringStyle handles json string style and returns the resulting value or error.
func jsonStringStyle() lipgloss.Style {
	return styles.PanelText
}

// jsonStringContextStyle handles json string context style and returns the resulting value or error.
func jsonStringContextStyle(key bool) lipgloss.Style {
	if key {
		return jsonKeyStyle()
	}
	return jsonStringStyle()
}

// jsonNumberStyle handles json number style and returns the resulting value or error.
func jsonNumberStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(styles.PaletteBlueBright)
}

// jsonTrueStyle handles json true style and returns the resulting value or error.
func jsonTrueStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(styles.ConditionLipglossColor("GREEN"))
}

// jsonFalseStyle handles json false style and returns the resulting value or error.
func jsonFalseStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(styles.PaletteError)
}

// jsonNullStyle handles json null style and returns the resulting value or error.
func jsonNullStyle() lipgloss.Style {
	return styles.PanelMuted
}

// jsonTokenStyle handles json token style and returns the resulting value or error.
func jsonTokenStyle(token string) lipgloss.Style {
	switch token {
	case "true":
		return jsonTrueStyle()
	case "false":
		return jsonFalseStyle()
	case "null":
		return jsonNullStyle()
	default:
		if token != "" && (token[0] == '-' || (token[0] >= '0' && token[0] <= '9')) {
			return jsonNumberStyle()
		}
		return jsonDefaultStyle()
	}
}

// renderBox renders render box for Model and returns the resulting state or error.
func (m Model) renderBox() string {
	border := borderStyle(m.Valid())
	body := strings.Split(m.renderArea(), "\n")
	innerWidth := max(m.screenW-6, 4)
	contentHeight := jsonContentHeight(m.screenH)
	scrollbar := expandedScrollbarState(m.visualLineCount(), m.area.ScrollYOffset(), contentHeight)

	lines := []string{border.Render("╭" + strings.Repeat("─", innerWidth) + "╮")}
	for i := range contentHeight {
		line := ""
		if i < len(body) {
			line = body[i]
		}
		rightEdge := border.Render("│")
		if scrollbar.visible && i >= scrollbar.thumbStart && i <= scrollbar.thumbEnd {
			rightEdge = styles.ScrollbarThumb.Render("█")
		}
		if line == "" {
			line = strings.Repeat(" ", innerWidth)
		}
		lines = append(lines, border.Render("│")+line+rightEdge)
	}
	lines = append(lines, border.Render("│")+renderHelpFooter(jsonHelpText(innerWidth), innerWidth)+border.Render("│"))
	lines = append(lines, border.Render("╰"+strings.Repeat("─", innerWidth)+"╯"))
	return strings.Join(lines, "\n")
}

// visualLineCount handles visual line count and returns the resulting value or error.
func (m Model) visualLineCount() int {
	lines := strings.Split(m.area.Value(), "\n")
	if len(lines) == 0 {
		return 1
	}
	gutter := lineNumberGutter(len(lines))
	contentWidth := max(max(m.screenW-6, 4)-gutter, 1)
	count := 0
	for _, line := range lines {
		count += len(wrapPlainLine(line, contentWidth))
	}
	return max(count, 1)
}

// jsonContentHeight handles json content height and returns the resulting value or error.
func jsonContentHeight(screenH int) int {
	return max(screenH-7, 3)
}

// jsonHelpText handles json help text and returns the resulting value or error.
func jsonHelpText(width int) string {
	m := help.New()
	m.ShortSeparator = " • "
	m.Styles.ShortKey = styles.FilterText
	m.Styles.ShortDesc = styles.PanelMuted
	m.Styles.ShortSeparator = styles.PanelMuted
	m.Styles.Ellipsis = styles.PanelMuted
	m.SetWidth(width)
	return m.ShortHelpView([]key.Binding{
		tuiconfig.Binding(tuiconfig.BlockJSONInput, tuiconfig.ActionSave, "save"),
		tuiconfig.Binding(tuiconfig.BlockJSONInput, tuiconfig.ActionFormat, "format"),
		tuiconfig.Binding(tuiconfig.BlockJSONInput, tuiconfig.ActionCancel, "cancel"),
		tuiconfig.Binding(tuiconfig.BlockJSONInput, tuiconfig.ActionCopyValue, "copy"),
	})
}

// renderHelpFooter renders help footer and returns the resulting value or error.
func renderHelpFooter(text string, width int) string {
	if width <= 0 {
		return ""
	}
	return text + strings.Repeat(" ", max(width-lipgloss.Width(text), 0))
}

// expandedScrollbar holds expanded scrollbar state used by the jsoninput package.
type expandedScrollbar struct {
	// visible stores visible for expandedScrollbar.
	visible bool
	// thumbStart stores thumb start for expandedScrollbar.
	thumbStart int
	// thumbEnd stores thumb end for expandedScrollbar.
	thumbEnd int
}

// expandedScrollbarState handles expanded scrollbar state and returns the resulting value or error.
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

// resetAreaCursor handles reset area cursor for Model and returns the resulting state or error.
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
