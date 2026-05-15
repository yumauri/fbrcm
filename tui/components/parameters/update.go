package parameters

import (
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"

	"fbrcm/tui/components/filterbox"
	"fbrcm/tui/messages"
)

// Update updates update for Model and returns the resulting state or error.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.ProjectsSelectionChangedMsg:
		cmd := m.setProjects(msg.Projects)
		if m.anyLoading() {
			return m, tea.Batch(cmd, m.spin.Tick, m.selectionChangedCmd(false))
		}
		return m, tea.Batch(cmd, m.selectionChangedCmd(false))

	case messages.ParametersLoadedMsg:
		cmd := m.updateProject(msg)
		if m.anyLoading() {
			return m, tea.Batch(cmd, m.spin.Tick, m.selectionChangedCmd(false))
		}
		return m, tea.Batch(cmd, m.selectionChangedCmd(false))

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
			return m, tea.Batch(cmd, keyboardCaptureCmd(true), m.selectionChangedCmd(false))
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
			return m, tea.Batch(cmd, m.selectionChangedCmd(false))
		}

		switch msg.String() {
		case "up", "k":
			m.moveCursor(-1)
			return m, m.selectionChangedCmd(false)
		case "down", "j":
			m.moveCursor(1)
			return m, m.selectionChangedCmd(false)
		case "pgdown":
			m.moveToNextGroup()
			return m, m.selectionChangedCmd(false)
		case "pgup":
			m.moveToPrevGroup()
			return m, m.selectionChangedCmd(false)
		case "left", "h":
			m.collapseCurrent()
			return m, m.selectionChangedCmd(false)
		case "right", "l":
			m.expandCurrent()
			return m, m.selectionChangedCmd(false)
		case " ", "space":
			m.toggleCurrentParameter()
			return m, m.selectionChangedCmd(false)
		case "home":
			m.moveToCurrentProjectHeader()
			return m, m.selectionChangedCmd(false)
		case "end":
			m.moveToLastParameterInCurrentProject()
			return m, m.selectionChangedCmd(false)
		case ">":
			m.setAllParametersExpanded(true)
			return m, m.selectionChangedCmd(false)
		case "<":
			m.setAllParametersExpanded(false)
			return m, m.selectionChangedCmd(false)
		case ")":
			m.setAllGroupsExpanded(true)
			return m, m.selectionChangedCmd(false)
		case "(":
			m.setAllGroupsExpanded(false)
			return m, m.selectionChangedCmd(false)
		case "enter":
			return m, m.selectionChangedCmd(true)
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
			return m, tea.Batch(keyboardCaptureCmd(false), m.selectionChangedCmd(false))
		}
		if index, ok := m.nodeIndexAtMouse(msg.Mouse()); ok {
			m.cursor = index
			m.ensureCursorVisible()
			return m, m.selectionChangedCmd(false)
		}

	case tea.MouseWheelMsg:
		if !m.isMouseInside(msg.Mouse()) {
			break
		}
		switch msg.Mouse().Button {
		case tea.MouseWheelUp:
			m.moveCursor(-1)
			return m, m.selectionChangedCmd(false)
		case tea.MouseWheelDown:
			m.moveCursor(1)
			return m, m.selectionChangedCmd(false)
		}

	default:
		if m.active && m.filter.Focused() {
			return m.updateFilterInput(msg)
		}
	}

	return m, nil
}

// keyboardCaptureCmd handles keyboard capture cmd and returns the resulting value or error.
func keyboardCaptureCmd(enabled bool) tea.Cmd {
	return func() tea.Msg {
		return messages.KeyboardCaptureMsg{
			Enabled: enabled,
		}
	}
}

// updateFilterInput updates update filter input for Model and returns the resulting state or error.
func (m Model) updateFilterInput(msg tea.Msg) (Model, tea.Cmd) {
	before := m.filter.Value()
	var cmd tea.Cmd
	m.filter, cmd = m.filter.Update(msg)
	if m.filter.Value() != before {
		m.applyFilter()
	}
	return m, tea.Batch(cmd, m.selectionChangedCmd(false))
}
