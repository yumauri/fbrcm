package details

import (
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/rc/evaluate"
	"github.com/yumauri/fbrcm/tui/contextview"
)

func (m Model) SetScenario(scenario *contextview.Scenario) Model {
	m.scenario = scenario
	m.deselectContextExcludedValue()
	if len(m.visibleConditionUsages()) == 0 {
		m.selectedUsage = -1
	}
	m.refreshViewport()
	return m
}

func (m *Model) deselectContextExcludedValue() {
	if m.ValueSelected() && !m.contextValueVisible(m.contextValues(), m.selectedValue) {
		m.selectedValue = -1
	}
}

func (m Model) conditionContextResult() evaluate.Result {
	if m.conditionData == nil || !m.scenario.AppliesTo(m.conditionData.Project) {
		return evaluate.Result{}
	}
	return m.scenario.Evaluate(m.conditionExpression)
}

func (m Model) visibleConditionUsages() []core.ConditionUsage {
	if m.conditionData == nil || !m.scenario.Visible(m.conditionContextResult().State) {
		return nil
	}
	return m.conditionData.Condition.Usages
}

func (m Model) contextValues() []contextview.Value {
	if m.data == nil {
		return nil
	}
	return m.scenario.Values(m.data.Project, m.data.Conditions, m.data.Parameter)
}

func (m Model) contextValueVisible(results []contextview.Value, index int) bool {
	if results == nil || results[index].Visible {
		return true
	}
	// Keep a just-added value available while the user is editing it.
	return !m.originalHasValue(m.data.Parameter.Values[index].Label)
}

func (m Model) visibleValueIndexes() []int {
	if m.data == nil {
		return nil
	}
	results := m.contextValues()
	indexes := make([]int, 0, len(m.data.Parameter.Values))
	for i := range m.data.Parameter.Values {
		if m.contextValueVisible(results, i) {
			indexes = append(indexes, i)
		}
	}
	return indexes
}
