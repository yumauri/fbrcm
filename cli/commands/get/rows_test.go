package get

import (
	"reflect"
	"testing"
	"time"

	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/filter"
	"github.com/yumauri/fbrcm/core/firebase"
)

func TestFlattenParametersOrdersGroupsAndRootParams(t *testing.T) {
	cachedAt := time.Date(2026, 6, 13, 12, 0, 0, 0, time.UTC)
	cfg := &firebase.RemoteConfig{
		Conditions: []firebase.RemoteConfigCondition{
			{Name: "ga", TagColor: "GREEN"},
			{Name: "beta", TagColor: "BLUE"},
		},
		ParameterGroups: map[string]firebase.RemoteConfigGroup{
			"Zeta": {
				Parameters: map[string]firebase.RemoteConfigParam{
					"shared_key": {
						Description: " grouped description ",
						ConditionalValues: map[string]firebase.RemoteConfigValue{
							"beta": {Value: "beta-value"},
							"ga":   {Value: "ga-value"},
						},
					},
				},
			},
		},
		Parameters: map[string]firebase.RemoteConfigParam{
			"alpha_key": {
				DefaultValue: &firebase.RemoteConfigValue{Value: "alpha"},
				ValueType:    "NUMBER",
			},
			"shared_key": {
				DefaultValue: &firebase.RemoteConfigValue{Value: "root should be hidden"},
			},
		},
		Version: firebase.RemoteConfigVersion{VersionNumber: " 42 "},
	}

	rows := flattenParameters(core.Project{Name: "Project A", ProjectID: "project-a"}, cfg, cachedAt, "cache", "", nil, shared.NewParameterSearch(""))

	if got, want := rowKeys(rows), []string{"shared_key", "alpha_key"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("row keys = %#v, want %#v", got, want)
	}
	grouped := rows[0]
	if grouped.Group != "Zeta" {
		t.Fatalf("group = %q, want Zeta", grouped.Group)
	}
	if grouped.Description != "grouped description" {
		t.Fatalf("description = %q, want trimmed grouped description", grouped.Description)
	}
	if grouped.Type != "string" {
		t.Fatalf("empty value type = %q, want string", grouped.Type)
	}
	if grouped.Version != "42" {
		t.Fatalf("version = %q, want 42", grouped.Version)
	}
	if got, want := conditionNames(grouped.Conditions), []string{"ga", "beta"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("condition names = %#v, want %#v", got, want)
	}
	if len(grouped.ValueLines) != 2 || grouped.ValueLines[0].Label != "ga" || grouped.ValueLines[1].Label != "beta" {
		t.Fatalf("value line labels = %#v, want condition order ga,beta", valueLineLabels(grouped.ValueLines))
	}
	if rows[1].Group != defaultGroupLabel || rows[1].Type != "NUMBER" {
		t.Fatalf("root row = %#v, want root NUMBER row", rows[1])
	}
	if rows[1].CachedAt != cachedAt || rows[1].Status != "cache" {
		t.Fatalf("cache metadata = %s/%q, want %s/cache", rows[1].CachedAt, rows[1].Status, cachedAt)
	}
}

func TestFilterHelpers(t *testing.T) {
	rows := []parameterRow{
		{Key: "alpha"},
		{Key: "beta"},
		{Key: "alphabet"},
	}

	filtered := filterParameterRows(rows, []string{"=alpha"})
	if got, want := rowKeys(filtered), []string{"alpha"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("exact filtered keys = %#v, want %#v", got, want)
	}

	mode, query := parseFilter("=alpha")
	if mode != filter.ModeExact || query != "alpha" {
		t.Fatalf("parseFilter exact = %v/%q, want exact/alpha", mode, query)
	}
	mode, query = parseFilter(" alpha ")
	if mode != filter.ModeFuzzy || query != "alpha" {
		t.Fatalf("parseFilter fuzzy = %v/%q, want fuzzy/alpha", mode, query)
	}

	for _, fn := range []struct {
		name  string
		exact func([]string) bool
	}{
		{name: "project", exact: singleExactProjectFilter},
		{name: "parameter", exact: singleExactParameterFilter},
	} {
		if !fn.exact([]string{"=alpha"}) {
			t.Fatalf("%s exact single filter = false, want true", fn.name)
		}
		if fn.exact([]string{"=alpha", "=beta"}) {
			t.Fatalf("%s repeated exact filter = true, want false", fn.name)
		}
		if fn.exact([]string{"alpha"}) {
			t.Fatalf("%s fuzzy filter = true, want false", fn.name)
		}
		if fn.exact([]string{"", "  "}) {
			t.Fatalf("%s empty filters = true, want false", fn.name)
		}
	}
}

func TestSortParameterRows(t *testing.T) {
	rows := []parameterRow{
		{Project: "Beta", ProjectID: "project-b", Group: defaultGroupLabel, Key: "zeta"},
		{Project: "", ProjectID: "project-c", Group: defaultGroupLabel, Key: "alpha"},
		{Project: "alpha", ProjectID: "project-a2", Group: "B", Key: "beta"},
		{Project: "Alpha", ProjectID: "project-a1", Group: "B", Key: "alpha"},
		{Project: "Alpha", ProjectID: "project-a1", Group: "A", Key: "zeta"},
	}

	sortParameterRows(rows)

	got := rowIDs(rows)
	want := []string{
		"project-a1/A/zeta",
		"project-a1/B/alpha",
		"project-a2/B/beta",
		"project-b/(root)/zeta",
		"project-c/(root)/alpha",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("sorted rows = %#v, want %#v", got, want)
	}
}

func TestBuildTableRowsAddsMissingProjects(t *testing.T) {
	projects := []loadedProjectParameters{
		{project: core.Project{Name: "Project A", ProjectID: "project-a"}, status: "cache"},
		{project: core.Project{Name: "Project B", ProjectID: "project-b"}, status: "missing"},
	}
	rows := []parameterRow{
		{Project: "Project A", ProjectID: "project-a", Group: defaultGroupLabel, Key: "flag", ValueLines: []valueLine{{Label: "Default value", Value: "on"}}},
	}

	tableRows := buildTableRows(projects, rows)

	if len(tableRows) != 2 {
		t.Fatalf("table row count = %d, want 2", len(tableRows))
	}
	if tableRows[0].ProjectID != "project-a" || tableRows[0].Key != "flag" {
		t.Fatalf("first table row = %#v, want original project-a row", tableRows[0])
	}
	missing := tableRows[1]
	if missing.Project != "Project B" || missing.ProjectID != "project-b" || missing.Status != "missing" {
		t.Fatalf("missing row metadata = %#v, want Project B/project-b/missing", missing)
	}
	if len(missing.ValueLines) != 1 || !missing.ValueLines[0].Missing || missing.ValueLines[0].Label != "Missing values" {
		t.Fatalf("missing value lines = %#v, want single Missing values line", missing.ValueLines)
	}
}

func rowKeys(rows []parameterRow) []string {
	keys := make([]string, len(rows))
	for i, row := range rows {
		keys[i] = row.Key
	}
	return keys
}

func conditionNames(conditions []parameterConditionJSON) []string {
	names := make([]string, len(conditions))
	for i, condition := range conditions {
		names[i] = condition.Name
	}
	return names
}

func valueLineLabels(lines []valueLine) []string {
	labels := make([]string, len(lines))
	for i, line := range lines {
		labels[i] = line.Label
	}
	return labels
}

func rowIDs(rows []parameterRow) []string {
	ids := make([]string, len(rows))
	for i, row := range rows {
		ids[i] = row.ProjectID + "/" + row.Group + "/" + row.Key
	}
	return ids
}
