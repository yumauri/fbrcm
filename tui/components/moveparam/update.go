package moveparam

import (
	tea "charm.land/bubbletea/v2"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

func (m *Model) Move(delta int) tea.Cmd {
	if m.rowsCount() == 0 {
		return nil
	}
	wasInput := m.selectedInput()
	m.selected = (m.selected + delta + m.rowsCount()) % m.rowsCount()
	m.search, m.lastType = "", time.Time{}
	if m.selectedInput() && !wasInput {
		return m.input.Focus()
	}
	if wasInput && !m.selectedInput() {
		m.input.Blur()
	}
	return nil
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	if !m.open || !m.selectedInput() {
		return nil
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return cmd
}

func (m *Model) Typeahead(key string, now time.Time) bool {
	if len(m.options) == 0 || utf8.RuneCountInString(key) != 1 || m.selectedInput() {
		return false
	}
	r, _ := utf8.DecodeRuneInString(key)
	if !unicode.IsPrint(r) {
		return false
	}
	folded := strings.ToLower(key)
	needle := folded
	if !m.lastType.IsZero() && now.Sub(m.lastType) <= typeaheadTimeout {
		needle = m.search + needle
	}
	if m.selectPrefix(needle, now) {
		return true
	}
	if needle != folded && m.selectPrefix(folded, now) {
		return true
	}
	m.search, m.lastType = folded, now
	return false
}

func (m *Model) selectPrefix(needle string, now time.Time) bool {
	for row := 0; row < m.rowsCount(); row++ {
		optionIndex, ok := m.optionIndexForRow(row)
		if ok && strings.HasPrefix(strings.ToLower(m.options[optionIndex].Label), needle) {
			m.selected, m.search, m.lastType = row, needle, now
			return true
		}
	}
	return false
}
