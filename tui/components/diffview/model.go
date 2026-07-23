package diffview

import (
	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core/dictdiff"
	tuiconfig "github.com/yumauri/fbrcm/tui/config"
)

type Model struct {
	result    dictdiff.Result
	open      bool
	screenW   int
	screenH   int
	cursor    int
	offset    int
	collapsed map[string]bool
}

func New() Model {
	return Model{collapsed: make(map[string]bool)}
}

func (m Model) Open(screenW, screenH int, result dictdiff.Result) Model {
	m.result = result
	m.open = true
	m.screenW = screenW
	m.screenH = screenH
	m.cursor = 0
	m.offset = 0
	m.collapsed = make(map[string]bool)
	m.ensureSelectedVisible()
	return m
}

func (m Model) Close() Model {
	return New()
}

func (m Model) IsOpen() bool { return m.open }

func (m Model) Position() (int, int) { return 2, 2 }

func (m Model) SetSize(width, height int) Model {
	m.screenW, m.screenH = width, height
	m.ensureSelectedVisible()
	return m
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.open {
		return m, nil
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.SetSize(msg.Width, msg.Height), nil
	case tea.MouseWheelMsg:
		delta := 3
		if msg.Mouse().Button == tea.MouseWheelUp {
			delta = -delta
		}
		m.scroll(delta)
		return m, nil
	case tea.KeyMsg:
		k := msg.String()
		switch {
		case tuiconfig.Matches(tuiconfig.BlockDiffView, tuiconfig.ActionClose, k):
			return m.Close(), nil
		case tuiconfig.Matches(tuiconfig.BlockDiffView, tuiconfig.ActionUp, k):
			m.move(-1)
		case tuiconfig.Matches(tuiconfig.BlockDiffView, tuiconfig.ActionDown, k):
			m.move(1)
		case tuiconfig.Matches(tuiconfig.BlockDiffView, tuiconfig.ActionPageUp, k):
			m.scroll(-m.bodyHeight())
		case tuiconfig.Matches(tuiconfig.BlockDiffView, tuiconfig.ActionPageDown, k):
			m.scroll(m.bodyHeight())
		case tuiconfig.Matches(tuiconfig.BlockDiffView, tuiconfig.ActionHome, k):
			m.cursor = 0
			m.ensureSelectedVisible()
		case tuiconfig.Matches(tuiconfig.BlockDiffView, tuiconfig.ActionEnd, k):
			m.cursor = max(len(m.result.Properties)-1, 0)
			m.ensureSelectedVisible()
		case tuiconfig.Matches(tuiconfig.BlockDiffView, tuiconfig.ActionLeft, k):
			m.setCollapsed(true)
		case tuiconfig.Matches(tuiconfig.BlockDiffView, tuiconfig.ActionRight, k):
			m.setCollapsed(false)
		case tuiconfig.Matches(tuiconfig.BlockDiffView, tuiconfig.ActionToggle, k):
			m.toggleCollapsed()
		}
	}
	return m, nil
}

func (m *Model) move(delta int) {
	if len(m.result.Properties) == 0 {
		m.cursor = 0
		return
	}
	m.cursor = max(0, min(m.cursor+delta, len(m.result.Properties)-1))
	m.ensureSelectedVisible()
}

func (m *Model) toggleCollapsed() {
	if name, ok := m.currentPropertyName(); ok {
		m.collapsed[name] = !m.collapsed[name]
		m.ensureSelectedVisible()
	}
}

func (m *Model) setCollapsed(collapsed bool) {
	if name, ok := m.currentPropertyName(); ok {
		m.collapsed[name] = collapsed
		m.ensureSelectedVisible()
	}
}

func (m Model) currentPropertyName() (string, bool) {
	if m.cursor < 0 || m.cursor >= len(m.result.Properties) {
		return "", false
	}
	return m.result.Properties[m.cursor].Name, true
}

func (m *Model) scroll(delta int) {
	rows := m.bodyRows(m.contentWidth())
	m.offset = min(max(m.offset+delta, 0), max(len(rows)-m.bodyHeight(), 0))
	m.selectNearestVisibleProperty(rows)
}

func (m *Model) selectNearestVisibleProperty(rows []bodyRow) {
	if len(m.result.Properties) == 0 {
		m.cursor = 0
		return
	}
	end := min(m.offset+m.bodyHeight(), len(rows))
	for index := m.offset; index < end; index++ {
		if rows[index].header {
			m.cursor = rows[index].property
			return
		}
	}
	for index := min(m.offset, len(rows)-1); index >= 0; index-- {
		if rows[index].header {
			m.cursor = rows[index].property
			return
		}
	}
}

func (m *Model) ensureSelectedVisible() {
	rows := m.bodyRows(m.contentWidth())
	header := -1
	for index, row := range rows {
		if row.header && row.property == m.cursor {
			header = index
			break
		}
	}
	if header < 0 {
		m.offset = min(m.offset, max(len(rows)-m.bodyHeight(), 0))
		return
	}
	if header < m.offset {
		m.offset = header
	}
	if header >= m.offset+m.bodyHeight() {
		m.offset = header - m.bodyHeight() + 1
	}
	m.offset = min(max(m.offset, 0), max(len(rows)-m.bodyHeight(), 0))
}
