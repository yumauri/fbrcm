package app

import (
	"context"
	"errors"
	"fmt"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	rcdiffinput "github.com/yumauri/fbrcm/core/rc/diffinput"
	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
	parameterscmp "github.com/yumauri/fbrcm/tui/components/parameters"
)

type parameterBlameLoadedMsg struct {
	project core.Project
	history core.ParameterHistory
	err     error
}

func (m Model) loadParameterBlameCmd(request parameterscmp.BlameRequestedMsg) tea.Cmd {
	return func() tea.Msg {
		if m.svc == nil {
			return parameterBlameLoadedMsg{
				project: request.Project,
				err:     errors.New("firebase service is unavailable"),
			}
		}
		history, err := m.svc.GetRemoteConfigParameterHistory(
			context.Background(),
			request.Project.ProjectID,
			request.Parameter,
			core.ParameterHistoryOptions{At: "current", Limit: 1},
		)
		return parameterBlameLoadedMsg{project: request.Project, history: history, err: err}
	}
}

func (m Model) updateParameterBlameLoaded(msg parameterBlameLoadedMsg) (Model, tea.Cmd, bool) {
	if msg.err != nil {
		m.openErrorDialog("Blame Unavailable", msg.project, msg.err.Error())
		return m, nil, true
	}
	if len(msg.history.Changes) == 0 {
		detail := "No direct change is available in retained Firebase history."
		if boundary := msg.history.Boundary; boundary != nil && boundary.Present {
			detail = fmt.Sprintf("The parameter was already present in the oldest retained version, v%s.", boundary.Version)
		}
		m.openErrorDialog("Blame Unavailable", msg.project, detail)
		return m, nil, true
	}

	change := msg.history.Changes[0]
	input := rcdiffinput.ParameterChange(
		change.Change,
		"Earlier version: v"+change.PreviousVersion,
		"Later version: v"+change.Version.VersionNumber,
	)
	m.openDictionaryDiffWithAttribution(
		input,
		msg.project,
		rcdisplay.FormatRemoteConfigVersionAttribution(change.Version),
	)
	return m, nil, true
}
