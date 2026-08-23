package app

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	corestyles "github.com/yumauri/fbrcm/core/styles"
	tuiconfig "github.com/yumauri/fbrcm/tui/config"
)

func (m Model) aboutView() string {
	if !m.aboutOpen {
		return ""
	}

	info := strings.TrimSuffix(m.buildInfo.Text(!corestyles.NoColorEnabled()), "\n")
	infoLines := strings.Split(info, "\n")
	contentWidth := 0
	for _, line := range infoLines {
		contentWidth = max(contentWidth, lipgloss.Width(line))
	}
	contentWidth++

	lines := []string{popupTopBorder("", "About", contentWidth), popupLine("", contentWidth)}
	for _, line := range infoLines {
		lines = append(lines, popupLine(line, contentWidth))
	}
	lines = append(lines, popupLine("", contentWidth), popupBottomBorder(contentWidth))
	return strings.Join(lines, "\n")
}

func (m Model) aboutPosition() (int, int) {
	view := m.aboutView()
	return max((m.width-lipgloss.Width(view))/2, 0), max((m.height-lipgloss.Height(view))/2, 0)
}

func (m Model) updateAbout(msg tea.Msg) (Model, tea.Cmd, bool) {
	if !m.aboutOpen {
		return m, nil, false
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.updateWindowSize(msg)
		next, cmd := m.updateChildPanels(msg)
		return next, cmd, true
	case tea.KeyMsg:
		m.aboutOpen = false
		key := msg.String()
		if tuiconfig.Matches(tuiconfig.BlockGlobal, tuiconfig.ActionForceQuit, key) {
			return m, tea.Quit, true
		}
		if tuiconfig.Matches(tuiconfig.BlockGlobal, tuiconfig.ActionQuit, key) {
			return m, m.requestQuit(), true
		}
		return m, nil, true
	case tea.MouseClickMsg:
		m.aboutOpen = false
		return m, nil, true
	case tea.MouseMsg:
		return m, nil, true
	default:
		return m, nil, false
	}
}
