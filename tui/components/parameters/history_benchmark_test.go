package parameters

import (
	"fmt"
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/core"
)

func BenchmarkHistoryLargeConfigNavigation(b *testing.B) {
	const parameterCount, valuesPerParameter = 500, 12
	makeTree := func(suffix string) *core.ParametersTree {
		params := make([]core.ParametersEntry, 0, parameterCount)
		for i := range parameterCount {
			values := make([]core.ParametersValue, 0, valuesPerParameter)
			for j := range valuesPerParameter {
				values = append(values, core.ParametersValue{Label: fmt.Sprintf("condition-%02d", j), Value: strings.Repeat("large-value-"+suffix, 160), ValueType: "STRING"})
			}
			params = append(params, core.ParametersEntry{Key: fmt.Sprintf("parameter_%04d", i), Values: values})
		}
		return &core.ParametersTree{Version: suffix, Groups: []core.ParametersGroup{{Key: "group", Label: "group", Parameters: params}}}
	}

	project := core.Project{ProjectID: "large", Name: "Large"}
	previous, current := makeTree("1"), makeTree("2")
	m := New(nil)
	m.projects = []projectState{{project: project, tree: current}}
	m.projectIndex[project.ProjectID] = 0
	m.groupExpanded[m.groupKey(project.ProjectID, "group")] = true
	m.histories[project.ProjectID] = buildHistoryState(historyState{previous: previous, current: current, previousVersion: "1", currentVersion: "2"})
	m = m.SetBounds(0, 0, 180, 50).SetHistory(true)
	m.setAllParametersExpanded(true)
	m.cursor = len(m.visible) / 2
	m.ensureCursorVisible()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.moveCursor(1)
		_ = m.View(true)
	}
}
