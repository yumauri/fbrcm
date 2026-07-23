package parameters

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/dictdiff"
	"github.com/yumauri/fbrcm/core/firebase"
	coreparameters "github.com/yumauri/fbrcm/core/parameters"
)

func TestHistoryEnterRequestsGenericPropertyDiff(t *testing.T) {
	project := core.Project{ProjectID: "demo", Name: "Demo"}
	previous := historyDiffTree("1", "Shopping cart", "old description", "true")
	current := historyDiffTree("2", "Shopping cart", "new description", "false")
	m := New(nil)
	m.projects = []projectState{{project: project, tree: current}}
	m.projectIndex[project.ProjectID] = 0
	m.groupExpanded[m.groupKey(project.ProjectID, "WEB")] = true
	m.histories[project.ProjectID] = buildHistoryState(historyState{
		previous: previous, current: current,
		previousVersion: "1", currentVersion: "2",
	})
	m = m.SetHistory(true)
	for index, node := range m.visible {
		if node.kind == nodeParameter && node.paramKey == "checkout_enabled" {
			m.cursor = index
			break
		}
	}

	next, cmd, handled := m.updateHistoryKey(tea.KeyPressMsg{}, "enter")
	if !handled || cmd == nil {
		t.Fatalf("history Enter = handled:%v cmd:%v", handled, cmd)
	}
	m = next
	message := cmd()
	request, ok := message.(HistoryDiffRequestedMsg)
	if !ok {
		t.Fatalf("history Enter message = %T, want HistoryDiffRequestedMsg", message)
	}
	input := request.Input
	if request.Project.ProjectID != project.ProjectID ||
		input.EntityName != "Property: WEB / checkout_enabled" ||
		input.Left.Name != "Earlier version: v1" ||
		input.Right.Name != "Later version: v2" {
		t.Fatalf("history diff identity = project:%#v entity:%q maps:%q -> %q",
			request.Project, input.EntityName, input.Left.Name, input.Right.Name)
	}
	if got := input.Left.Properties["description"].Raw; got != "old description" {
		t.Fatalf("earlier description = %q, want old description", got)
	}
	if got := input.Right.Properties["description"].Raw; got != "new description" {
		t.Fatalf("later description = %q, want new description", got)
	}
	result, err := dictdiff.Compare(input)
	if err != nil {
		t.Fatal(err)
	}
	value := historyDiffProperty(t, result, "value · default")
	row := value.Chunks[0].Rows[0]
	if row.Left.Segments[0].Kind != dictdiff.LineChanged ||
		row.Right.Segments[0].Kind != dictdiff.LineChanged {
		t.Fatalf("history boolean diff is not atomic: %#v", row)
	}
}

func historyDiffTree(version, groupDescription, parameterDescription, value string) *core.ParametersTree {
	return coreparameters.BuildTree(&firebase.RemoteConfig{
		Version: firebase.RemoteConfigVersion{VersionNumber: version},
		ParameterGroups: map[string]firebase.RemoteConfigGroup{
			"WEB": {
				Description: groupDescription,
				Parameters: map[string]firebase.RemoteConfigParam{
					"checkout_enabled": {
						ValueType:   "BOOLEAN",
						Description: parameterDescription,
						DefaultValue: &firebase.RemoteConfigValue{
							Value: value,
						},
					},
				},
			},
		},
	}, time.Time{}, "")
}

func historyDiffProperty(t *testing.T, result dictdiff.Result, name string) dictdiff.Property {
	t.Helper()
	for _, property := range result.Properties {
		if property.Name == name {
			return property
		}
	}
	t.Fatalf("property %q not found in %#v", name, result.Properties)
	return dictdiff.Property{}
}
