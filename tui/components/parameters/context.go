package parameters

import (
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/tui/contextview"
)

func (m Model) SetScenario(scenario *contextview.Scenario) Model {
	m.scenario = scenario
	if !m.history {
		m.applyFilter()
	}
	return m
}

// ReferencedConditionExpressions uses the loaded templates before any view filters.
func (m Model) ReferencedConditionExpressions() []string {
	var expressions []string
	for _, project := range m.projects {
		for _, condition := range project.tree.RemoteConfig().ReferencedConditions() {
			expressions = append(expressions, condition.Expression)
		}
	}
	return expressions
}

func (m Model) contextResults(projectID string, param *core.ParametersEntry) []contextview.Value {
	if m.scenario == nil || m.history {
		return nil
	}
	project := m.projectByID(projectID)
	if project == nil || project.tree == nil {
		return nil
	}
	return m.scenario.Values(project.project, project.tree.Conditions, *param)
}

func (m Model) contextVisibleValues(projectID string, param *core.ParametersEntry) []core.ParametersValue {
	results := m.contextResults(projectID, param)
	if results == nil {
		return param.Values
	}
	values := make([]core.ParametersValue, 0, len(param.Values))
	for i, value := range param.Values {
		if results[i].Visible {
			values = append(values, value)
		}
	}
	return values
}

func (m Model) contextValueDecoration(node visibleNode, param *core.ParametersEntry) (string, string) {
	results := m.contextResults(node.projectID, param)
	if results == nil {
		return m.valueConnector(node, param), ""
	}
	first, last := -1, -1
	for i, result := range results {
		if result.Visible {
			if first < 0 {
				first = i
			}
			last = i
		}
	}
	connector := "mid"
	switch {
	case first == last:
		connector = "single"
	case node.valueIdx == first:
		connector = "first"
	case node.valueIdx == last:
		connector = "last"
	}
	return connector, results[node.valueIdx].Status
}
