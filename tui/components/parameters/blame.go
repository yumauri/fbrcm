package parameters

import (
	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
)

// BlameRequestedMsg asks the application to find the newest publication that
// directly changed the selected parameter.
type BlameRequestedMsg struct {
	Project   core.Project
	Parameter string
}

func (m Model) blameRequestedCmd() tea.Cmd {
	project, _, parameter, ok := m.CurrentParameterRef()
	if !ok {
		return nil
	}
	return func() tea.Msg {
		return BlameRequestedMsg{Project: project, Parameter: parameter}
	}
}
