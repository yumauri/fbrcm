package conditions

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/atotto/clipboard"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/tui/messages"
)

func (m Model) loadConditionsCmd(project core.Project, source, selectConditionName string) tea.Cmd {
	return func() tea.Msg {
		cache, state, err := m.svc.InspectParametersCache(project.ProjectID)
		if err != nil {
			return messages.ConditionsLoadedMsg{Project: project, SelectConditionName: selectConditionName, Err: err}
		}
		if state == core.ParametersCacheMissing || cache == nil {
			return messages.ConditionsLoadedMsg{Project: project, SelectConditionName: selectConditionName, Err: fmt.Errorf("remote config cache is missing")}
		}
		tree, draft, err := m.svc.BuildDraftAwareConditionsTree(project.ProjectID, cache)
		if draft {
			source = "draft"
		}
		return messages.ConditionsLoadedMsg{Project: project, Tree: tree, Source: source, SelectConditionName: selectConditionName, Err: err}
	}
}

func copyCmd(value string) tea.Cmd {
	if value == "" {
		return nil
	}
	return func() tea.Msg {
		_ = clipboard.WriteAll(value)
		return nil
	}
}
