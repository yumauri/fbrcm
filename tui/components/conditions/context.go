package conditions

import (
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/rc/evaluate"
	"github.com/yumauri/fbrcm/tui/contextview"
)

func (m Model) SetScenario(scenario *contextview.Scenario) Model {
	m.scenario = scenario
	m.syncVisible()
	return m
}

func (m Model) contextResult(project core.Project, condition core.ConditionEntry) evaluate.Result {
	if !m.scenario.AppliesTo(project) {
		return evaluate.Result{}
	}
	return m.scenario.Evaluate(condition.Expression)
}
