package managedfeatures

import (
	"time"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"

	tuiconfig "github.com/yumauri/fbrcm/tui/config"
	"github.com/yumauri/fbrcm/tui/messages"
)

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.ProjectsSelectionChangedMsg:
		m.setProjects(msg.Projects)
		if m.active {
			return m.Activate()
		}
		return m, nil
	case messages.ParametersLoadedMsg:
		index, ok := m.projectIndex[msg.Project.ProjectID]
		if !ok {
			break
		}
		if msg.Err != nil {
			m.projects[index].templateReady = false
			m.projects[index].loading = false
			if m.projects[index].waitingTemplate {
				m.projects[index].loaded = false
				m.projects[index].waitingTemplate = false
			}
			m.projects[index].err = msg.Err
			m.syncVisible()
			return m, nil
		}
		if msg.Revalidate {
			m.projects[index].templateReady = false
			if m.active && m.projects[index].waitingTemplate {
				m.projects[index].loading = true
				m.projects[index].err = nil
				m.syncVisible()
				return m, m.spin.Tick
			}
			return m, nil
		}
		m.projects[index].templateReady = true
		m.projects[index].err = nil
		if m.active && m.projects[index].waitingTemplate && !m.projects[index].loaded {
			m.projects[index].loading = false
			m.projects[index].waitingTemplate = false
			return m.startProjectLoads([]int{index}, false)
		}
		m.projects[index].loading = false
		return m, nil
	case messages.ManagedFeaturesLoadedMsg:
		m.updateLoaded(msg)
		return m, nil
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
		if m.filterEnabled() && !m.filter.ExpressionFocused() {
			if mode, ok := tuiconfig.FilterModeForKey(k); ok {
				cmd := m.filter.Activate(mode)
				m.syncVisible()
				return m, tea.Batch(cmd, messages.KeyboardCaptureCmd(true), m.selectionChangedCmd(false))
			}
		}
		if m.filterEnabled() && m.filter.Focused() {
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
		case tuiconfig.Matches(tuiconfig.BlockManagedFeatures, tuiconfig.ActionUp, k):
			m.moveCursor(-1)
			return m, m.selectionChangedCmd(false)
		case tuiconfig.Matches(tuiconfig.BlockManagedFeatures, tuiconfig.ActionDown, k):
			m.moveCursor(1)
			return m, m.selectionChangedCmd(false)
		case tuiconfig.Matches(tuiconfig.BlockManagedFeatures, tuiconfig.ActionPageUp, k):
			m.movePage(-m.contentHeight())
			return m, m.selectionChangedCmd(false)
		case tuiconfig.Matches(tuiconfig.BlockManagedFeatures, tuiconfig.ActionPageDown, k):
			m.movePage(m.contentHeight())
			return m, m.selectionChangedCmd(false)
		case tuiconfig.Matches(tuiconfig.BlockManagedFeatures, tuiconfig.ActionFirst, k):
			m.moveFirst()
			return m, m.selectionChangedCmd(false)
		case tuiconfig.Matches(tuiconfig.BlockManagedFeatures, tuiconfig.ActionLast, k):
			m.moveLast()
			return m, m.selectionChangedCmd(false)
		case tuiconfig.Matches(tuiconfig.BlockManagedFeatures, tuiconfig.ActionOpenDetails, k):
			return m, m.selectionChangedCmd(true)
		case tuiconfig.Matches(tuiconfig.BlockParameters, tuiconfig.ActionReload, k):
			return m.refreshCurrent()
		case tuiconfig.Matches(tuiconfig.BlockParameters, tuiconfig.ActionReloadAll, k):
			return m.refreshAll()
		}
	case tea.MouseClickMsg:
		if msg.Mouse().Button != tea.MouseLeft {
			break
		}
		if index, ok := m.nodeAtMouse(msg.Mouse().X, msg.Mouse().Y); ok {
			m.cursor = index
			m.ensureCursorVisible()
			if m.lastClick.Register(0, index, time.Now()) {
				return m, m.selectionChangedCmd(true)
			}
			return m, m.selectionChangedCmd(false)
		}
	case tea.MouseWheelMsg:
		if !m.contains(msg.Mouse().X, msg.Mouse().Y) {
			break
		}
		if msg.Mouse().Button == tea.MouseWheelUp {
			m.moveCursor(-1)
		} else if msg.Mouse().Button == tea.MouseWheelDown {
			m.moveCursor(1)
		}
		return m, m.selectionChangedCmd(false)
	default:
		if m.active && m.filterEnabled() && m.filter.Focused() {
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

func (m Model) refreshCurrent() (Model, tea.Cmd) {
	if m.cursor < 0 || m.cursor >= len(m.visible) {
		return m, nil
	}
	index, ok := m.projectIndex[m.visible[m.cursor].projectID]
	if !ok {
		return m, nil
	}
	return m.startProjectLoads([]int{index}, true)
}

func (m Model) refreshAll() (Model, tea.Cmd) {
	indices := make([]int, len(m.projects))
	for index := range m.projects {
		indices[index] = index
	}
	return m.startProjectLoads(indices, true)
}

func (m Model) anyLoading() bool {
	for _, project := range m.projects {
		if project.loading {
			return true
		}
	}
	return false
}
