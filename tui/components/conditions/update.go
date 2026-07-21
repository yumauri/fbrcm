package conditions

import (
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"

	tuiconfig "github.com/yumauri/fbrcm/tui/config"
	"github.com/yumauri/fbrcm/tui/messages"
)

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.ProjectsSelectionChangedMsg:
		m.setProjects(msg.Projects)
		return m, m.selectionChangedCmd(false)
	case messages.ParametersLoadedMsg:
		idx, ok := m.projectIndex[msg.Project.ProjectID]
		if !ok {
			return m, nil
		}
		state := m.projects[idx]
		if msg.Err != nil {
			state.loading = false
			state.err = msg.Err
			m.projects[idx] = state
			m.syncVisible()
			return m, nil
		}
		state.loading = true
		state.err = nil
		state.hasDraft = msg.HasDraft
		state.staleDraft = msg.StaleDraft
		if msg.CacheSource != "" {
			state.cacheSource = msg.CacheSource
		} else if msg.Source != "draft" && msg.Source != "draft-stale" {
			state.cacheSource = msg.Source
		}
		if msg.CacheVersion != "" {
			state.cacheVersion = msg.CacheVersion
		} else if msg.Tree != nil && !msg.HasDraft {
			state.cacheVersion = msg.Tree.Version
		}
		if msg.DraftVersion != "" {
			state.draftVersion = msg.DraftVersion
		} else if msg.HasDraft && msg.Tree != nil {
			state.draftVersion = msg.Tree.Version
		} else if !msg.HasDraft {
			state.draftVersion = ""
		}
		m.projects[idx] = state
		return m, tea.Batch(m.loadConditionsCmd(msg.Project, msg.Source, msg.SelectConditionName), m.spin.Tick)
	case messages.ConditionsLoadedMsg:
		m.updateLoaded(msg)
		return m, m.selectionChangedCmd(false)
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
		if !m.filter.ExpressionFocused() && tuiconfig.Matches(tuiconfig.BlockFilter, tuiconfig.ActionFilterExpression, k) {
			cmd := m.filter.ActivateExpression()
			m.syncVisible()
			return m, tea.Batch(cmd, messages.KeyboardCaptureCmd(true), m.selectionChangedCmd(false))
		}
		if !m.filter.ExpressionFocused() {
			if mode, ok := tuiconfig.FilterModeForKey(k); ok {
				cmd := m.filter.Activate(mode)
				m.syncVisible()
				return m, tea.Batch(cmd, messages.KeyboardCaptureCmd(true), m.selectionChangedCmd(false))
			}
		}
		if m.filter.Focused() {
			switch {
			case tuiconfig.Matches(tuiconfig.BlockFilter, tuiconfig.ActionFilterApply, k):
				m.filter.Blur()
				return m, messages.KeyboardCaptureCmd(false)
			case tuiconfig.Matches(tuiconfig.BlockFilter, tuiconfig.ActionFilterCancel, k):
				m.filter.ClearAndBlur()
				m.syncVisible()
				return m, tea.Batch(messages.KeyboardCaptureCmd(false), m.selectionChangedCmd(false))
			case tuiconfig.Matches(tuiconfig.BlockFilter, tuiconfig.ActionFilterUp, k):
				m.filter.Blur()
				m.moveCursor(-1)
				return m, tea.Batch(messages.KeyboardCaptureCmd(false), m.selectionChangedCmd(false))
			case tuiconfig.Matches(tuiconfig.BlockFilter, tuiconfig.ActionFilterDown, k):
				m.filter.Blur()
				m.moveCursor(1)
				return m, tea.Batch(messages.KeyboardCaptureCmd(false), m.selectionChangedCmd(false))
			}
			before := m.filter.Value()
			var cmd tea.Cmd
			m.filter, cmd = m.filter.Update(msg)
			if m.filter.Value() != before {
				m.syncVisible()
			}
			return m, tea.Batch(cmd, m.selectionChangedCmd(false))
		}
		switch {
		case tuiconfig.Matches(tuiconfig.BlockConditions, tuiconfig.ActionUp, k):
			m.moveCursor(-1)
			return m, m.selectionChangedCmd(false)
		case tuiconfig.Matches(tuiconfig.BlockConditions, tuiconfig.ActionDown, k):
			m.moveCursor(1)
			return m, m.selectionChangedCmd(false)
		case tuiconfig.Matches(tuiconfig.BlockConditions, tuiconfig.ActionPageUp, k):
			m.moveCursor(-m.contentHeight())
			return m, m.selectionChangedCmd(false)
		case tuiconfig.Matches(tuiconfig.BlockConditions, tuiconfig.ActionPageDown, k):
			m.moveCursor(m.contentHeight())
			return m, m.selectionChangedCmd(false)
		case tuiconfig.Matches(tuiconfig.BlockConditions, tuiconfig.ActionFirst, k):
			m.moveFirst()
			return m, m.selectionChangedCmd(false)
		case tuiconfig.Matches(tuiconfig.BlockConditions, tuiconfig.ActionLast, k):
			m.moveLast()
			return m, m.selectionChangedCmd(false)
		case tuiconfig.Matches(tuiconfig.BlockConditions, tuiconfig.ActionOpenDetails, k):
			return m, m.selectionChangedCmd(true)
		case tuiconfig.Matches(tuiconfig.BlockConditions, tuiconfig.ActionCopyName, k):
			return m, m.copyCurrentNameCmd()
		case tuiconfig.Matches(tuiconfig.BlockConditions, tuiconfig.ActionCopyPath, k):
			return m, m.copyCurrentPathCmd()
		}
	case tea.MouseClickMsg:
		if index, ok := m.nodeAtMouse(msg.Mouse().X, msg.Mouse().Y); ok {
			m.cursor = index
			m.ensureCursorVisible()
			return m, m.selectionChangedCmd(false)
		}
	case tea.MouseWheelMsg:
		if !m.isMouseInside(msg.Mouse().X, msg.Mouse().Y) {
			break
		}
		if msg.Mouse().Button == tea.MouseWheelUp {
			m.moveCursor(-1)
		} else if msg.Mouse().Button == tea.MouseWheelDown {
			m.moveCursor(1)
		}
		return m, m.selectionChangedCmd(false)
	default:
		if m.active && m.filter.Focused() {
			before := m.filter.Value()
			var cmd tea.Cmd
			m.filter, cmd = m.filter.Update(msg)
			if m.filter.Value() != before {
				m.syncVisible()
			}
			return m, tea.Batch(cmd, m.selectionChangedCmd(false))
		}
	}
	return m, nil
}

func (m Model) anyLoading() bool {
	for _, project := range m.projects {
		if project.loading {
			return true
		}
	}
	return false
}
