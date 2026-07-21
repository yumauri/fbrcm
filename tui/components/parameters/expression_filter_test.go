package parameters

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	coreparameters "github.com/yumauri/fbrcm/core/parameters"
	"github.com/yumauri/fbrcm/tui/components/filterbox"
)

func TestParametersExpressionFilterMatchesTypedValues(t *testing.T) {
	project := core.Project{Name: "Demo", ProjectID: "demo"}
	cfg := &firebase.RemoteConfig{Parameters: map[string]firebase.RemoteConfigParam{
		"enabled":  {ValueType: "BOOLEAN", DefaultValue: &firebase.RemoteConfigValue{Value: "true"}},
		"disabled": {ValueType: "BOOLEAN", DefaultValue: &firebase.RemoteConfigValue{Value: "false"}},
	}}
	tree := coreparameters.BuildTree(cfg, timeZero, "")
	m := expressionParameterModel(project, tree, `default == true`)

	parameters := expressionVisibleParameterKeys(m.visible)
	if len(parameters) != 1 || parameters[0] != "enabled" {
		t.Fatalf("visible parameters = %v, want enabled", parameters)
	}
}

func TestHistoryExpressionFilterUsesPreviousConfigForRemovedParameter(t *testing.T) {
	project := core.Project{Name: "Demo", ProjectID: "demo"}
	previous := coreparameters.BuildTree(&firebase.RemoteConfig{Parameters: map[string]firebase.RemoteConfigParam{
		"removed": {ValueType: "STRING", DefaultValue: &firebase.RemoteConfigValue{Value: "old"}},
	}}, timeZero, "")
	current := coreparameters.BuildTree(&firebase.RemoteConfig{Parameters: map[string]firebase.RemoteConfigParam{}}, timeZero, "")
	state := buildHistoryState(historyState{previous: previous, current: current})
	m := New(nil)
	m.history = true
	m.projects = []projectState{{project: project, tree: current}}
	m.histories[project.ProjectID] = state
	m.filter = expressionFilter(`default == "old"`)
	for _, group := range state.merged.Groups {
		m.groupExpanded[m.groupKey(project.ProjectID, group.Key)] = true
	}
	m.syncVisible()

	parameters := expressionVisibleParameterKeys(m.visible)
	if len(parameters) != 1 || parameters[0] != "removed" {
		t.Fatalf("visible history parameters = %v, want removed", parameters)
	}
}

var timeZero = time.Time{}

func expressionParameterModel(project core.Project, tree *core.ParametersTree, expression string) Model {
	m := New(nil)
	m.projects = []projectState{{project: project, tree: tree}}
	m.projectIndex[project.ProjectID] = 0
	m.filter = expressionFilter(expression)
	for _, group := range tree.Groups {
		m.groupExpanded[m.groupKey(project.ProjectID, group.Key)] = true
	}
	m.syncVisible()
	return m
}

func expressionFilter(expression string) filterbox.Model {
	box := filterbox.New()
	box.ActivateExpression()
	box, _ = box.Update(tea.PasteMsg{Content: expression})
	box.Blur()
	return box
}

func expressionVisibleParameterKeys(nodes []visibleNode) []string {
	var keys []string
	for _, node := range nodes {
		if node.kind == nodeParameter {
			keys = append(keys, node.paramKey)
		}
	}
	return keys
}
