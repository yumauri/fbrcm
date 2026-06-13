package projects

import (
	"time"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core/browser"
	"github.com/yumauri/fbrcm/core/firebase"
	corelog "github.com/yumauri/fbrcm/core/log"
	tuiconfig "github.com/yumauri/fbrcm/tui/config"
	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/panels"
)

const doubleClickWindow = 400 * time.Millisecond

// Update updates update for Model and returns the resulting state or error.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.ProjectsLoadedMsg:
		m.allProjects = msg.Projects
		m.source = msg.Source
		m.err = msg.Err
		m.loading = false
		m.applyFilter()
		m.syncViewport()

	case spinner.TickMsg:
		if !m.loading {
			break
		}
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		m.refreshViewport()
		return m, cmd

	case tea.KeyMsg:
		if !m.active {
			break
		}
		if m.collapsed {
			return m, nil
		}

		k := msg.String()
		if mode, ok := tuiconfig.FilterModeForKey(k); ok {
			cmd := m.filter.Activate(mode)
			m.applyFilter()
			m.syncViewport()
			return m, tea.Batch(cmd, keyboardCaptureCmd(true))
		}

		if m.filter.Focused() {
			switch {
			case tuiconfig.Matches(tuiconfig.BlockFilter, tuiconfig.ActionFilterApply, k):
				m.filter.Blur()
				m.selectOnlyCurrent()
				m.refreshViewport()
				return m, tea.Batch(m.selectionChangedCmd(), keyboardCaptureCmd(false), setActivePanelCmd(panels.Parameters))
			case tuiconfig.Matches(tuiconfig.BlockFilter, tuiconfig.ActionFilterCancel, k):
				m.filter.ClearAndBlur()
				m.applyFilter()
				m.syncViewport()
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
			m.syncViewport()
			return m, cmd
		}

		switch {
		case tuiconfig.Matches(tuiconfig.BlockProjects, tuiconfig.ActionUp, k):
			m.moveCursor(-1)
		case tuiconfig.Matches(tuiconfig.BlockProjects, tuiconfig.ActionDown, k):
			m.moveCursor(1)
		case tuiconfig.Matches(tuiconfig.BlockProjects, tuiconfig.ActionHome, k):
			m.jumpToFirst()
		case tuiconfig.Matches(tuiconfig.BlockProjects, tuiconfig.ActionEnd, k):
			m.jumpToLast()
		case tuiconfig.Matches(tuiconfig.BlockProjects, tuiconfig.ActionRefresh, k):
			if m.loading {
				break
			}
			m.loading = true
			m.refreshViewport()
			return m, tea.Batch(m.syncProjectsCmd(), m.spinner.Tick)
		case tuiconfig.Matches(tuiconfig.BlockProjects, tuiconfig.ActionPageDown, k):
			m.pageDown()
		case tuiconfig.Matches(tuiconfig.BlockProjects, tuiconfig.ActionPageUp, k):
			m.pageUp()
		case tuiconfig.Matches(tuiconfig.BlockProjects, tuiconfig.ActionSelect, k):
			m.selectOnlyCurrent()
			return m, tea.Batch(m.selectionChangedCmd(), setActivePanelCmd(panels.Parameters))
		case tuiconfig.Matches(tuiconfig.BlockProjects, tuiconfig.ActionOpen, k):
			return m, m.openCurrentProjectCmd()
		case tuiconfig.Matches(tuiconfig.BlockProjects, tuiconfig.ActionMark, k):
			m.toggleCurrentSelection()
			return m, m.selectionChangedCmd()
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
			if index, ok := m.projectIndexAtMouse(msg.Mouse()); ok {
				m.cursor = index
				m.syncViewport()
				if msg.Mouse().Button == tea.MouseLeft && m.isDoubleClick(index) {
					m.selectOnlyCurrent()
					return m, tea.Batch(m.selectionChangedCmd(), keyboardCaptureCmd(false))
				}
				m.rememberClick(index)
			}
			return m, keyboardCaptureCmd(false)
		}

		if index, ok := m.projectIndexAtMouse(msg.Mouse()); ok {
			m.cursor = index
			m.syncViewport()
			if msg.Mouse().Button == tea.MouseLeft && m.isDoubleClick(index) {
				m.selectOnlyCurrent()
				return m, m.selectionChangedCmd()
			}
			m.rememberClick(index)
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

// updateFilterInput updates update filter input for Model and returns the resulting state or error.
func (m Model) updateFilterInput(msg tea.Msg) (Model, tea.Cmd) {
	before := m.filter.Value()
	var cmd tea.Cmd
	m.filter, cmd = m.filter.Update(msg)
	if m.filter.Value() != before {
		m.applyFilter()
	}
	m.syncViewport()
	return m, cmd
}

// keyboardCaptureCmd handles keyboard capture cmd and returns the resulting value or error.
func keyboardCaptureCmd(enabled bool) tea.Cmd {
	return func() tea.Msg {
		return messages.KeyboardCaptureMsg{
			Enabled: enabled,
		}
	}
}

// setActivePanelCmd sets set active panel cmd and returns the resulting value or error.
func setActivePanelCmd(panel panels.ID) tea.Cmd {
	return func() tea.Msg {
		return messages.SetActivePanelMsg{Panel: panel}
	}
}

// isDoubleClick reports is double click for Model and returns the resulting state or error.
func (m Model) isDoubleClick(index int) bool {
	return m.lastClick.project == index && time.Since(m.lastClick.at) <= doubleClickWindow
}

// rememberClick handles remember click for Model and returns the resulting state or error.
func (m *Model) rememberClick(index int) {
	m.lastClick.project = index
	m.lastClick.at = time.Now()
}

// openCurrentProjectCmd opens open current project cmd for Model and returns the resulting state or error.
func (m Model) openCurrentProjectCmd() tea.Cmd {
	if len(m.projects) == 0 || m.cursor < 0 || m.cursor >= len(m.projects) {
		return nil
	}

	project := m.projects[m.cursor]
	url := firebase.RemoteConfigConsoleURL(project.ProjectID)
	return func() tea.Msg {
		logger := corelog.For("tui.projects")
		logger.Info("open project remote config", "project_id", project.ProjectID, "url", url)
		if err := browser.OpenURL(url); err != nil {
			logger.Error("open project remote config failed", "project_id", project.ProjectID, "url", url, "err", err)
		}
		return nil
	}
}
