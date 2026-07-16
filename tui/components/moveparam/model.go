package moveparam

import (
	"image/color"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
)

type Option struct {
	Key                    string
	Label                  string
	Foreground             color.Color
	KeepForegroundOnSelect bool
}

type Model struct {
	x        int
	y        int
	label    string
	options  []Option
	selected int
	input    textinput.Model
	open     bool
	allowNew bool
	search   string
	lastType time.Time
}

const (
	typeaheadTimeout  = 900 * time.Millisecond
	minGroupNameWidth = 9
)

func New() Model {
	return Model{input: newInput()}
}

func (m Model) Open(x, y int, label string, options []Option) Model {
	return m.openWithOptions(x, y, label, options, 0, true)
}

// OpenOptions opens a fixed option list without the free-form "new group" row.
func (m Model) OpenOptions(x, y int, label string, options []Option, selected int) Model {
	return m.openWithOptions(x, y, label, options, selected, false)
}

func (m Model) openWithOptions(x, y int, label string, options []Option, selected int, allowNew bool) Model {
	m.x = x
	m.y = y
	m.label = label
	m.options = append([]Option(nil), options...)
	m.allowNew = allowNew
	m.selected = min(max(selected, 0), max(m.rowsCount()-1, 0))
	m.input = newInput()
	m.open = true
	m.search = ""
	m.lastType = time.Time{}
	if m.selectedInput() {
		_ = m.input.Focus()
	}
	return m
}

func (m Model) Close() Model {
	m.open = false
	m.label = ""
	m.options = nil
	m.selected = 0
	m.allowNew = false
	m.input = newInput()
	m.search = ""
	m.lastType = time.Time{}
	return m
}

func (m Model) IsOpen() bool {
	return m.open
}

func (m Model) Position() (int, int) {
	return m.x, m.y
}

func (m Model) ListPosition() (int, int) {
	connectorWidth, _ := m.layout()
	return m.x + connectorWidth + 2, m.y - m.selected
}

func (m Model) Current() (Option, bool) {
	if !m.open {
		return Option{}, false
	}
	if m.selectedInput() {
		value := strings.TrimSpace(m.input.Value())
		if value == "" {
			return Option{}, false
		}
		return Option{Key: value, Label: value}, true
	}
	optionIndex, ok := m.optionIndexForRow(m.selected)
	if !ok {
		return Option{}, false
	}
	return m.options[optionIndex], true
}

func (m Model) InputSelected() bool {
	return m.selectedInput()
}

func (m Model) rowsCount() int {
	if !m.allowNew {
		return len(m.options)
	}
	if len(m.options) == 0 {
		return 0
	}
	return len(m.options) + 1
}

func (m Model) rootOptionIndex() int {
	if !m.allowNew {
		return -1
	}
	if len(m.options) == 0 {
		return -1
	}
	last := len(m.options) - 1
	if m.options[last].Key == "" {
		return last
	}
	return -1
}

func (m Model) inputRowIndex() int {
	if !m.allowNew {
		return -1
	}
	rootIndex := m.rootOptionIndex()
	if rootIndex >= 0 {
		return rootIndex
	}
	return len(m.options)
}

func (m Model) rowIsInput(row int) bool {
	return m.allowNew && row == m.inputRowIndex()
}

func (m Model) selectedInput() bool {
	return m.rowIsInput(m.selected)
}

func (m Model) optionIndexForRow(row int) (int, bool) {
	if row < 0 || row >= m.rowsCount() || m.rowIsInput(row) {
		return 0, false
	}
	if !m.allowNew {
		return row, true
	}
	inputRow := m.inputRowIndex()
	if row > inputRow {
		return row - 1, true
	}
	return row, true
}
