package app

import (
	"fmt"
	"maps"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core/draft"
	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/core/rc/evaluate"
	rctarget "github.com/yumauri/fbrcm/core/rc/target"
	"github.com/yumauri/fbrcm/ops/workflows/explain"
	"github.com/yumauri/fbrcm/tui/components/textview"
	tuiconfig "github.com/yumauri/fbrcm/tui/config"
	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/panels"
)

func (m Model) explainContext() (Model, tea.Cmd, bool) {
	data := m.contextExplanationData()
	if data == nil {
		return m, nil, false
	}
	render, err := m.prepareContextExplanation(data)
	if err != nil {
		render = func(int) string { return err.Error() }
	}
	m.explainView = textview.Open(m.width, m.height, "Explain "+data.Parameter.Key+" with Context", render)
	return m, nil, true
}

func (m Model) prepareContextExplanation(data *messages.ParameterViewData) (func(int) string, error) {
	cfg, source := m.parameters.TemplateSnapshot(data.Project.ProjectID)
	if cfg == nil {
		return nil, fmt.Errorf("no Remote Config template is loaded for %s", data.Project.ProjectID)
	}
	key := data.Parameter.Key
	if m.active == panels.Details {
		if edit, ok := m.details.Edit(); ok {
			copy, err := firebase.CloneRemoteConfig(cfg)
			if err != nil {
				return nil, err
			}
			if err := draft.EditParameterDetails(edit)(copy); err != nil {
				return nil, fmt.Errorf("explain pending parameter edit: %w", err)
			}
			cfg, key, source = copy, edit.NextParamKey, "pending Details edits"
		}
	}
	c := m.requestContext
	input := explain.TemplateInput{
		Project: data.Project.ProjectID, Parameter: key, Source: source, Config: cfg,
	}
	target, err := rctarget.Parse(data.Project.ProjectID)
	if c.capture != nil && err == nil && c.capture.TemplateKind() == target.Kind &&
		c.capture.MatchesProject(target.ProjectID, data.Project.ProjectNumber) {
		input.Exchange, input.Capture = c.capture, c.captureSource
	}
	for _, fact := range c.facts {
		if fact.Source == manualContextSource {
			input.Assumptions = true
			break
		}
	}
	return explain.PrepareHumanReport(input, &evaluate.Context{Facts: maps.Clone(c.facts), Capture: c.capture})
}

func (m Model) updateExplainView(msg tea.Msg) (Model, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.updateWindowSize(msg)
		m.explainView = m.explainView.SetSize(msg.Width, msg.Height)
		return m, nil, true
	case tea.KeyMsg:
		if tuiconfig.Matches(tuiconfig.BlockGlobal, tuiconfig.ActionQuit, msg.String()) {
			m.explainView = textview.Model{}
			return m, m.requestQuit(), true
		}
	case tea.MouseMsg:
	default:
		return m, nil, false
	}
	var cmd tea.Cmd
	m.explainView, cmd = m.explainView.Update(msg)
	return m, cmd, true
}
