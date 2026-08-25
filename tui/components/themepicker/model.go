package themepicker

import (
	"maps"
	"time"

	"charm.land/lipgloss/v2"

	corestyles "github.com/yumauri/fbrcm/core/styles"
	"github.com/yumauri/fbrcm/tui/components/mouseutil"
	"github.com/yumauri/fbrcm/tui/styles"
)

const (
	paletteSwatchWidth = 2
	paletteSwatchGlyph = "▇"
)

type Option struct {
	Name    string
	Palette corestyles.Palette
	Err     error
}

type Model struct {
	x, y          int
	width, height int
	options       []Option
	cursor        int
	scroll        int
	open          bool
	saving        bool
	errorText     string
	lastClick     mouseutil.ClickTracker
}

func New() Model { return Model{} }

func (m Model) SetBounds(x, y, width, height int) Model {
	m.x, m.y, m.width, m.height = x, y, width, height
	m.ensureVisible()
	return m
}

func (m Model) Open(options []Option, selected int) Model {
	m.options = append([]Option(nil), options...)
	m.cursor = min(max(selected, 0), max(len(options)-1, 0))
	m.scroll = 0
	m.open = true
	m.saving = false
	m.errorText = ""
	m.lastClick.Reset()
	m.ensureVisible()
	return m
}

func (m Model) Close() Model {
	m.options = nil
	m.cursor = 0
	m.scroll = 0
	m.open = false
	m.saving = false
	m.errorText = ""
	m.lastClick.Reset()
	return m
}

func (m Model) IsOpen() bool { return m.open }

func (m Model) Saving() bool { return m.saving }

func (m Model) SetSaving(saving bool) Model {
	m.saving = saving
	if saving {
		m.errorText = ""
	}
	return m
}

func (m Model) SetError(err error) Model {
	m.saving = false
	if err == nil {
		m.errorText = ""
	} else {
		m.errorText = err.Error()
	}
	return m
}

func (m Model) Current() (Option, bool) {
	if !m.open || m.cursor < 0 || m.cursor >= len(m.options) {
		return Option{}, false
	}
	return m.options[m.cursor], true
}

// SetCurrentPalette replaces the highlighted option after a successful live
// reload. Keeping this update in the picker also refreshes its palette preview
// and allows a theme that was initially invalid to recover while it is open.
func (m Model) SetCurrentPalette(palette corestyles.Palette) Model {
	if !m.open || m.cursor < 0 || m.cursor >= len(m.options) {
		return m
	}
	m.options[m.cursor].Palette = maps.Clone(palette)
	m.options[m.cursor].Err = nil
	m.errorText = ""
	return m
}

func (m *Model) Move(delta int) bool {
	if len(m.options) == 0 || m.saving {
		return false
	}
	previous := m.cursor
	m.cursor = (m.cursor + delta + len(m.options)) % len(m.options)
	m.errorText = ""
	m.ensureVisible()
	return m.cursor != previous
}

func (m *Model) MoveTo(index int) bool {
	if index < 0 || index >= len(m.options) || m.saving {
		return false
	}
	changed := m.cursor != index
	m.cursor = index
	m.errorText = ""
	m.ensureVisible()
	return changed
}

func (m *Model) MoveFirst() bool { return m.MoveTo(0) }

func (m *Model) MoveLast() bool { return m.MoveTo(len(m.options) - 1) }

// SelectOptionAt selects a row and reports double-click activation separately.
func (m *Model) SelectOptionAt(x, y int, at time.Time) (changed, double, hit bool) {
	boxX, boxY := m.Position()
	boxWidth, _ := m.boxSize()
	if !m.open || m.saving || x < boxX+1 || x >= boxX+boxWidth-1 {
		return false, false, false
	}
	visibleIndex := y - (boxY + firstOptionRow)
	index := m.scroll + visibleIndex
	if visibleIndex < 0 || visibleIndex >= m.visibleRows() || index < 0 || index >= len(m.options) {
		return false, false, false
	}
	changed = m.MoveTo(index)
	return changed, m.lastClick.Register(0, index, at), true
}

func (m Model) Position() (int, int) {
	w, h := m.boxSize()
	return max(m.x+(m.width-w)/2, m.x), max(m.y+(m.height-h)/2, m.y)
}

func (m Model) visibleRows() int {
	return min(len(m.options), max(m.height-8, 1))
}

func (m *Model) ensureVisible() {
	rows := m.visibleRows()
	if m.cursor < m.scroll {
		m.scroll = m.cursor
	}
	if m.cursor >= m.scroll+rows {
		m.scroll = m.cursor - rows + 1
	}
	m.scroll = max(0, min(m.scroll, max(len(m.options)-rows, 0)))
}

func (m Model) contentWidth() int {
	width := lipgloss.Width(pickerHelpLine(1_000))
	for _, option := range m.options {
		width = max(width, lipgloss.Width(option.Name)+2+paletteWidth(option))
	}
	if m.errorText != "" {
		width = max(width, min(lipgloss.Width(m.errorText), 72))
	}
	return min(max(width, 24), max(m.width-7, 1))
}

func (m Model) boxSize() (int, int) {
	height := max(m.visibleRows(), 1) + 5
	if m.saving || m.errorText != "" {
		height++
	}
	return m.contentWidth() + 8, height
}

func paletteWidth(option Option) int {
	if styles.NoColorEnabled() {
		return 0
	}
	if option.Err != nil || option.Palette == nil {
		return lipgloss.Width("unavailable")
	}
	return len(corestyles.PreviewTokens()) * paletteSwatchWidth
}
