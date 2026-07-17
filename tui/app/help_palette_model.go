package app

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	corefilter "github.com/yumauri/fbrcm/core/filter"
	"github.com/yumauri/fbrcm/tui/components/inputstyles"
	tuiconfig "github.com/yumauri/fbrcm/tui/config"
)

type helpPaletteModel struct {
	input  textinput.Model
	open   bool
	cursor int
	scroll int
}

type helpPaletteKeyMsg struct{ value string }

func (m helpPaletteKeyMsg) String() string { return m.value }
func (m helpPaletteKeyMsg) Key() tea.Key   { return tea.Key{} }

func newHelpPaletteModel() helpPaletteModel {
	input := textinput.New()
	input.Prompt = ""
	input.Placeholder = "Search actions, panels, or shortcuts"
	input.SetStyles(inputstyles.TextInput())
	input.Blur()
	return helpPaletteModel{input: input}
}

func (m helpPaletteModel) IsOpen() bool { return m.open }

func (m helpPaletteModel) Open() (helpPaletteModel, tea.Cmd) {
	m.open = true
	m.cursor = 0
	m.scroll = 0
	m.input.SetValue("")
	return m, m.input.Focus()
}

func (m helpPaletteModel) Close() helpPaletteModel {
	m.open = false
	m.cursor = 0
	m.scroll = 0
	m.input.SetValue("")
	m.input.Blur()
	return m
}

func (m helpPaletteModel) filtered(actions []helpPaletteAction) []helpPaletteAction {
	query := strings.TrimSpace(m.input.Value())
	if query == "" {
		return actions
	}
	out := make([]helpPaletteAction, 0, len(actions))
	for _, item := range actions {
		haystack := strings.Join([]string{item.group, item.title, strings.Join(item.keys, " ")}, " ")
		if matched, _ := corefilter.Match(haystack, query, corefilter.ModeFuzzy); matched {
			out = append(out, item)
		}
	}
	return out
}

func (m *helpPaletteModel) move(delta, count, height int) {
	if count == 0 {
		m.cursor, m.scroll = 0, 0
		return
	}
	m.cursor = min(max(m.cursor+delta, 0), count-1)
	m.ensureVisible(count, height)
}

func (m *helpPaletteModel) goTo(index, count, height int) {
	if count == 0 {
		m.cursor, m.scroll = 0, 0
		return
	}
	m.cursor = min(max(index, 0), count-1)
	m.ensureVisible(count, height)
}

func (m *helpPaletteModel) ensureVisible(count, height int) {
	height = max(height, 1)
	m.cursor = min(max(m.cursor, 0), max(count-1, 0))
	if m.cursor < m.scroll {
		m.scroll = m.cursor
	}
	if m.cursor >= m.scroll+height {
		m.scroll = m.cursor - height + 1
	}
	m.scroll = min(max(m.scroll, 0), max(count-height, 0))
}

func (m Model) updateHelpPalette(msg tea.Msg) (Model, tea.Cmd, bool) {
	if !m.helpPalette.IsOpen() {
		return m, nil, false
	}

	actions := m.helpPalette.filtered(m.helpPaletteActions())
	height := helpPaletteListHeight(m.height)
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		k := keyMsg.String()
		switch {
		case tuiconfig.Matches(tuiconfig.BlockHelp, tuiconfig.ActionCancel, k):
			m.helpPalette = m.helpPalette.Close()
			return m, nil, true
		case tuiconfig.Matches(tuiconfig.BlockHelp, tuiconfig.ActionUp, k):
			m.helpPalette.move(-1, len(actions), height)
			return m, nil, true
		case tuiconfig.Matches(tuiconfig.BlockHelp, tuiconfig.ActionDown, k):
			m.helpPalette.move(1, len(actions), height)
			return m, nil, true
		case tuiconfig.Matches(tuiconfig.BlockHelp, tuiconfig.ActionPageUp, k):
			m.helpPalette.move(-height, len(actions), height)
			return m, nil, true
		case tuiconfig.Matches(tuiconfig.BlockHelp, tuiconfig.ActionPageDown, k):
			m.helpPalette.move(height, len(actions), height)
			return m, nil, true
		case tuiconfig.Matches(tuiconfig.BlockHelp, tuiconfig.ActionHome, k):
			m.helpPalette.goTo(0, len(actions), height)
			return m, nil, true
		case tuiconfig.Matches(tuiconfig.BlockHelp, tuiconfig.ActionEnd, k):
			m.helpPalette.goTo(len(actions)-1, len(actions), height)
			return m, nil, true
		case tuiconfig.Matches(tuiconfig.BlockHelp, tuiconfig.ActionSubmit, k):
			return m.runHelpPaletteAction(actions, height)
		}
	}

	before := m.helpPalette.input.Value()
	var cmd tea.Cmd
	m.helpPalette.input, cmd = m.helpPalette.input.Update(msg)
	if m.helpPalette.input.Value() != before {
		m.helpPalette.cursor = 0
		m.helpPalette.scroll = 0
	}
	return m, cmd, true
}

func (m Model) runHelpPaletteAction(actions []helpPaletteAction, height int) (Model, tea.Cmd, bool) {
	if len(actions) == 0 || m.helpPalette.cursor < 0 || m.helpPalette.cursor >= len(actions) {
		return m, nil, true
	}
	item := actions[m.helpPalette.cursor]
	if !item.enabled {
		return m, nil, true
	}
	if item.block == tuiconfig.BlockHelp {
		switch item.action {
		case tuiconfig.ActionCancel:
			m.helpPalette = m.helpPalette.Close()
		case tuiconfig.ActionUp:
			m.helpPalette.move(-1, len(actions), height)
		case tuiconfig.ActionDown:
			m.helpPalette.move(1, len(actions), height)
		case tuiconfig.ActionPageUp:
			m.helpPalette.move(-height, len(actions), height)
		case tuiconfig.ActionPageDown:
			m.helpPalette.move(height, len(actions), height)
		case tuiconfig.ActionHome:
			m.helpPalette.goTo(0, len(actions), height)
		case tuiconfig.ActionEnd:
			m.helpPalette.goTo(len(actions)-1, len(actions), height)
		}
		return m, nil, true
	}
	if item.block == tuiconfig.BlockGlobal && item.action == tuiconfig.ActionHelp {
		m.helpPalette = m.helpPalette.Close()
		return m, nil, true
	}

	key := item.keys[0]
	m.helpPalette = m.helpPalette.Close()
	return m, func() tea.Msg { return helpPaletteKeyMsg{value: key} }, true
}

func helpPaletteListHeight(terminalHeight int) int {
	return max(min(terminalHeight-8, 20), 5)
}
