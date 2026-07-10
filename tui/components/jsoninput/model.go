package jsoninput

import (
	"bytes"
	"encoding/json"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/tui/components/inputstyles"
)

type JSONRange struct {
	Start     int
	End       int
	CursorCol int
}

type Model struct {
	screenW int
	screenH int
	area    textarea.Model
	open    bool
}

func New() Model {
	return Model{area: inputstyles.NewTextarea()}
}

func (m Model) Open(screenW, screenH int, value string) (Model, tea.Cmd) {
	m.screenW = screenW
	m.screenH = screenH
	m.area = inputstyles.NewTextarea()
	m.area.SetValue(prettyJSON(value))
	m.resize()
	m.resetAreaCursor()
	m.open = true
	return m, m.area.Focus()
}

func (m Model) Close() Model {
	m.open = false
	m.area.Blur()
	m.area.SetValue("")
	return m
}

func (m Model) IsOpen() bool {
	return m.open
}

func (m Model) Position() (int, int) {
	return 2, 2
}

func (m Model) Value() string {
	return m.area.Value()
}

func (m Model) Valid() bool {
	return json.Valid([]byte(m.area.Value()))
}

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

func (m Model) PrettyValue() string {
	return prettyJSON(m.area.Value())
}

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

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.open {
		return m, nil
	}
	var cmd tea.Cmd
	m.area, cmd = m.area.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if !m.open {
		return ""
	}
	return m.renderBox()
}

func (m *Model) resize() {
	innerWidth := max(m.screenW-6, 4)
	innerHeight := jsonContentHeight(m.screenH)
	gutter := lineNumberGutter(m.area.LineCount())
	m.area.SetWidth(max(innerWidth-gutter, 1))
	m.area.SetHeight(innerHeight)
}

func (m *Model) resetAreaCursor() {
	for m.area.Line() > 0 {
		m.area.CursorUp()
	}
	m.area.CursorStart()
}

func (m *Model) setAreaCursor(line, col int) {
	m.resetAreaCursor()
	for m.area.Line() < line {
		m.area.CursorDown()
	}
	m.area.SetCursorColumn(col)
}
