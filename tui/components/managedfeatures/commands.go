package managedfeatures

import (
	"context"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/tui/messages"
)

func (m Model) loadProjectCmd(project core.Project, update bool) tea.Cmd {
	return func() tea.Msg {
		msg := messages.ManagedFeaturesLoadedMsg{Kind: m.kind, Project: project}
		switch m.kind {
		case messages.ManagedFeatureExperiment:
			var result core.ExperimentList
			result, msg.Err = m.svc.ListRemoteConfigExperiments(context.Background(), project, update)
			msg.Template = result.Template
			msg.Experiments = result.Experiments
		case messages.ManagedFeaturePersonalization:
			var result core.PersonalizationList
			result, msg.Err = m.svc.ListRemoteConfigPersonalizations(context.Background(), project, update)
			msg.Template = result.Template
			msg.Personalizations = result.Personalizations
		case messages.ManagedFeatureRollout:
			var result core.RolloutList
			result, msg.Err = m.svc.ListRemoteConfigRollouts(context.Background(), project, update)
			msg.Template = result.Template
			msg.Rollouts = result.Rollouts
		}
		return msg
	}
}

func (m Model) startProjectLoads(indices []int, update bool) (Model, tea.Cmd) {
	commands := make([]tea.Cmd, 0, len(indices)+1)
	for _, index := range indices {
		if index < 0 || index >= len(m.projects) {
			continue
		}
		state := &m.projects[index]
		if (state.loading && !state.waitingTemplate) || (!update && state.loaded) {
			continue
		}
		if !update && !state.templateReady && state.err == nil {
			state.loading = state.err == nil
			state.waitingTemplate = state.err == nil
			continue
		}
		state.loading = true
		state.waitingTemplate = false
		state.err = nil
		commands = append(commands, m.loadProjectCmd(state.project, update))
	}
	m.syncVisible()
	if m.anyLoading() {
		commands = append(commands, m.spin.Tick)
	}
	return m, tea.Batch(commands...)
}
