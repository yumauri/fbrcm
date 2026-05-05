package parameters

import (
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"

	"fbrcm/tui/components/filterbox"
	"fbrcm/tui/messages"
)

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.ProjectsSelectionChangedMsg:
		cmd := m.setProjects(msg.Projects)
		if m.anyLoading() {
			return m, tea.Batch(cmd, m.spin.Tick)
		}
		return m, cmd

	case messages.ParametersLoadedMsg:
		cmd := m.updateProject(msg)
		if m.anyLoading() {
			return m, tea.Batch(cmd, m.spin.Tick)
		}
		return m, cmd

	case spinner.TickMsg:
		if !m.anyLoading() {
			return m, nil
		}
		var cmd tea.Cmd
		m.spin, cmd = m.spin.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		if !m.active {
			break
		}

		if mode, ok := filterbox.ModeForKey(msg.String()); ok {
			cmd := m.filter.Activate(mode)
			m.applyFilter()
			return m, tea.Batch(cmd, keyboardCaptureCmd(true))
		}

		if m.filter.Focused() {
			switch msg.String() {
			case "enter":
				m.filter.Blur()
				return m, keyboardCaptureCmd(false)
			case "esc":
				m.filter.ClearAndBlur()
				m.applyFilter()
				return m, keyboardCaptureCmd(false)
			case "up":
				m.filter.Blur()
				m.moveCursor(-1)
				return m, keyboardCaptureCmd(false)
			case "down":
				m.filter.Blur()
				m.moveCursor(1)
				return m, keyboardCaptureCmd(false)
			}

			before := m.filter.Value()
			var cmd tea.Cmd
			m.filter, cmd = m.filter.Update(msg)
			if m.filter.Value() != before {
				m.applyFilter()
			}
			return m, cmd
		}

		switch msg.String() {
		case "up", "k":
			m.moveCursor(-1)
		case "down", "j":
			m.moveCursor(1)
		case "pgdown":
			m.moveToNextGroup()
		case "pgup":
			m.moveToPrevGroup()
		case "left", "h":
			m.collapseCurrent()
		case "right", "l":
			m.expandCurrent()
		case "home":
			m.moveToCurrentProjectHeader()
		case "end":
			m.moveToLastParameterInCurrentProject()
		case ">":
			m.setAllParametersExpanded(true)
		case "<":
			m.setAllParametersExpanded(false)
		case ")":
			m.setAllGroupsExpanded(true)
		case "(":
			m.setAllGroupsExpanded(false)
		case "r":
			return m, m.revalidateCurrentProjectCmd()
		case "R":
			return m, m.revalidateAllProjectsCmd()
		case "y":
			return m, m.copyCurrentParameterNameCmd()
		case "Y":
			return m, m.copyCurrentParameterPathCmd()
		}

	case tea.MouseClickMsg:
		if !m.isMouseInside(msg.Mouse()) {
			if m.filter.Focused() {
				m.filter.Blur()
				return m, keyboardCaptureCmd(false)
			}
			break
		}
		if m.isMouseOnFilter(msg.Mouse()) {
			cmd := m.filter.Activate(m.filter.Mode())
			return m, tea.Batch(cmd, keyboardCaptureCmd(true))
		}
		if m.filter.Focused() {
			m.filter.Blur()
			if index, ok := m.nodeIndexAtMouse(msg.Mouse()); ok {
				m.cursor = index
				m.ensureCursorVisible()
			}
			return m, keyboardCaptureCmd(false)
		}
		if index, ok := m.nodeIndexAtMouse(msg.Mouse()); ok {
			m.cursor = index
			m.ensureCursorVisible()
		}

	case tea.MouseWheelMsg:
		if !m.isMouseInside(msg.Mouse()) {
			break
		}
		switch msg.Mouse().Button {
		case tea.MouseWheelUp:
			m.moveCursor(-1)
		case tea.MouseWheelDown:
			m.moveCursor(1)
		}

	default:
		if m.active && m.filter.Focused() {
			return m.updateFilterInput(msg)
		}
	}

	return m, nil
}

func keyboardCaptureCmd(enabled bool) tea.Cmd {
	return func() tea.Msg {
		return messages.KeyboardCaptureMsg{
			Enabled: enabled,
		}
	}
}

func (m Model) updateFilterInput(msg tea.Msg) (Model, tea.Cmd) {
	before := m.filter.Value()
	var cmd tea.Cmd
	m.filter, cmd = m.filter.Update(msg)
	if m.filter.Value() != before {
		m.applyFilter()
	}
	return m, cmd
}
