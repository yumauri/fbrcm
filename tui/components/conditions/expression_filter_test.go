package conditions

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	coreconditions "github.com/yumauri/fbrcm/core/conditions"
	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/tui/components/filterbox"
)

func TestConditionsExpressionFilterUsesConditionContext(t *testing.T) {
	project := core.Project{Name: "Demo", ProjectID: "demo"}
	cfg := &firebase.RemoteConfig{
		Conditions: []firebase.RemoteConfigCondition{
			{Name: "used", Expression: "true"},
			{Name: "unused", Expression: "false"},
		},
		Parameters: map[string]firebase.RemoteConfigParam{
			"flag": {ConditionalValues: map[string]firebase.RemoteConfigValue{"used": {Value: "on"}}},
		},
	}
	tree := coreconditions.BuildTree(cfg, time.Time{}, "")
	box := filterbox.New()
	box.ActivateExpression()
	box, _ = box.Update(tea.PasteMsg{Content: `usage_count == 0`})
	box.Blur()
	m := New(nil)
	m.filter = box
	m.projects = []projectState{{project: project, tree: tree}}
	m.projectIndex[project.ProjectID] = 0
	m.syncVisible()

	var names []string
	for _, node := range m.visible {
		if node.kind == nodeCondition {
			names = append(names, node.conditionName)
		}
	}
	if len(names) != 1 || names[0] != "unused" {
		t.Fatalf("visible conditions = %v, want unused", names)
	}
}
