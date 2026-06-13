package parameters

import (
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"

	tuiconfig "github.com/yumauri/fbrcm/tui/config"
	"github.com/yumauri/fbrcm/tui/messages"
)

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

		k := msg.String()
		if mode, ok := tuiconfig.FilterModeForKey(k); ok {
			cmd := m.filter.Activate(mode)
			m.applyFilter()
			return m, tea.Batch(cmd, keyboardCaptureCmd(true), m.selectionChangedCmd(false))
		}

		if m.filter.Focused() {
			switch {
			case tuiconfig.Matches(tuiconfig.BlockFilter, tuiconfig.ActionFilterApply, k):
				m.filter.Blur()
				return m, keyboardCaptureCmd(false)
			case tuiconfig.Matches(tuiconfig.BlockFilter, tuiconfig.ActionFilterCancel, k):
				m.filter.ClearAndBlur()
				m.applyFilter()
				return m, keyboardCaptureCmd(false)
			case tuiconfig.Matches(tuiconfig.BlockFilter, tuiconfig.ActionFilterUp, k):
				m.filter.Blur()
				m.moveCursor(-1)
				return m, keyboardCaptureCmd(false)
			case tuiconfig.Matches(tuiconfig.BlockFilter, tuiconfig.ActionFilterDown, k):
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

		switch {
		case tuiconfig.Matches(tuiconfig.BlockParameters, tuiconfig.ActionUp, k):
			m.moveCursor(-1)
			return m, m.selectionChangedCmd(false)
		case tuiconfig.Matches(tuiconfig.BlockParameters, tuiconfig.ActionDown, k):
			m.moveCursor(1)
			return m, m.selectionChangedCmd(false)
		case tuiconfig.Matches(tuiconfig.BlockParameters, tuiconfig.ActionNextGroup, k):
			m.moveToNextGroup()
			return m, m.selectionChangedCmd(false)
		case tuiconfig.Matches(tuiconfig.BlockParameters, tuiconfig.ActionPrevGroup, k):
			m.moveToPrevGroup()
			return m, m.selectionChangedCmd(false)
		case tuiconfig.Matches(tuiconfig.BlockParameters, tuiconfig.ActionCollapse, k):
			m.collapseCurrent()
			return m, m.selectionChangedCmd(false)
		case tuiconfig.Matches(tuiconfig.BlockParameters, tuiconfig.ActionExpand, k):
			m.expandCurrent()
			return m, m.selectionChangedCmd(false)
		case tuiconfig.Matches(tuiconfig.BlockParameters, tuiconfig.ActionToggle, k):
			m.toggleCurrentParameter()
			return m, m.selectionChangedCmd(false)
		case tuiconfig.Matches(tuiconfig.BlockParameters, tuiconfig.ActionFirst, k):
			m.moveToCurrentProjectHeader()
			return m, m.selectionChangedCmd(false)
		case tuiconfig.Matches(tuiconfig.BlockParameters, tuiconfig.ActionLast, k):
			m.moveToLastParameterInCurrentProject()
			return m, m.selectionChangedCmd(false)
		case tuiconfig.Matches(tuiconfig.BlockParameters, tuiconfig.ActionExpandAll, k):
			m.setAllParametersExpanded(true)
			return m, m.selectionChangedCmd(false)
		case tuiconfig.Matches(tuiconfig.BlockParameters, tuiconfig.ActionCollapseAll, k):
			m.setAllParametersExpanded(false)
			return m, m.selectionChangedCmd(false)
		case tuiconfig.Matches(tuiconfig.BlockParameters, tuiconfig.ActionExpandGroups, k):
			m.setAllGroupsExpanded(true)
			return m, m.selectionChangedCmd(false)
		case tuiconfig.Matches(tuiconfig.BlockParameters, tuiconfig.ActionCollapseGroups, k):
			m.setAllGroupsExpanded(false)
			return m, m.selectionChangedCmd(false)
		case tuiconfig.Matches(tuiconfig.BlockParameters, tuiconfig.ActionOpenDetails, k):
			return m, m.selectionChangedCmd(true)
		case tuiconfig.Matches(tuiconfig.BlockParameters, tuiconfig.ActionReload, k):
			return m, m.revalidateCurrentProjectCmd()
		case tuiconfig.Matches(tuiconfig.BlockParameters, tuiconfig.ActionReloadAll, k):
			return m, m.revalidateAllProjectsCmd()
		case tuiconfig.Matches(tuiconfig.BlockParameters, tuiconfig.ActionCopyName, k):
			return m, m.copyCurrentParameterNameCmd()
		case tuiconfig.Matches(tuiconfig.BlockParameters, tuiconfig.ActionCopyPath, k):
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
	return m, tea.Batch(cmd, m.selectionChangedCmd(false))
}
