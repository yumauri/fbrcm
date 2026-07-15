package conditions

import (
	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/tui/messages"
)

func (m Model) selectionChangedCmd(activate bool) tea.Cmd {
	data, ok := m.currentData()
	if !ok {
		return func() tea.Msg { return messages.ConditionSelectionChangedMsg{ResetScroll: true} }
	}
	return func() tea.Msg {
		return messages.ConditionSelectionChangedMsg{Data: data, Activate: activate}
	}
}

func (m Model) copyCurrentNameCmd() tea.Cmd {
	data, ok := m.currentData()
	if !ok {
		return nil
	}
	return copyCmd(data.Condition.Name)
}

func (m Model) copyCurrentPathCmd() tea.Cmd {
	data, ok := m.currentData()
	if !ok {
		return nil
	}
	return copyCmd(data.Project.ProjectID + "/conditions/" + data.Condition.Name)
}
