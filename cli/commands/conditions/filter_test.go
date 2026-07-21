package conditions

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/core"
)

func TestFilterEntriesSupportsGetStylePrefixes(t *testing.T) {
	entries := []core.ConditionEntry{
		{Name: "production", Expression: "platform == 'android'"},
		{Name: "preview", Expression: "internal rollout"},
		{Name: "staff"},
	}

	tests := []struct {
		name    string
		filters []string
		search  string
		want    []string
	}{
		{name: "fuzzy prefix", filters: []string{"~pdn"}, want: []string{"production"}},
		{name: "unprefixed fuzzy", filters: []string{"pdn"}, want: []string{"production"}},
		{name: "starts with", filters: []string{"^pre"}, want: []string{"preview"}},
		{name: "includes", filters: []string{"/duct"}, want: []string{"production"}},
		{name: "exact", filters: []string{"=staff"}, want: []string{"staff"}},
		{name: "repeated filters are OR", filters: []string{"=staff", "^pre"}, want: []string{"preview", "staff"}},
		{name: "search remains AND", filters: []string{"~pdn"}, search: "android", want: []string{"production"}},
		{name: "search rejects filter match", filters: []string{"~pdn"}, search: "ios"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotEntries := filterEntries(entries, tc.filters, tc.search)
			got := make([]string, len(gotEntries))
			for i, entry := range gotEntries {
				got[i] = entry.Name
			}
			if !slices.Equal(got, tc.want) {
				t.Fatalf("filterEntries() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestFilterEntriesByExprUsesConditionContext(t *testing.T) {
	project := core.Project{Name: "Demo", ProjectID: "demo"}
	entries := []core.ConditionEntry{
		{Name: "unused", Priority: 1, TagColor: "RED"},
		{Name: "used", Priority: 2, TagColor: "BLUE", Usages: []core.ConditionUsage{{GroupLabel: "(root)", ParameterKey: "flag"}}},
	}

	tests := []struct {
		name string
		expr string
		want []string
	}{
		{name: "empty", want: []string{"unused", "used"}},
		{name: "unused", expr: `usage_count == 0`, want: []string{"unused"}},
		{name: "priority and color", expr: `priority == 2 && color == "BLUE"`, want: []string{"used"}},
		{name: "usage parameter", expr: `any(usages, #.parameter == "flag")`, want: []string{"used"}},
		{name: "project", expr: `project_id == "demo"`, want: []string{"unused", "used"}},
		{name: "invalid", expr: `priority ==`, want: nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotEntries := filterEntriesByExpr(project, entries, tc.expr)
			got := make([]string, len(gotEntries))
			for i, entry := range gotEntries {
				got[i] = entry.Name
			}
			if !slices.Equal(got, tc.want) {
				t.Fatalf("filterEntriesByExpr(%q) = %v, want %v", tc.expr, got, tc.want)
			}
		})
	}
}

func TestConditionFiltersComposeWithAnd(t *testing.T) {
	project := core.Project{Name: "Demo", ProjectID: "demo"}
	entries := []core.ConditionEntry{
		{Name: "unused", Expression: "release audience"},
		{Name: "used", Expression: "release audience", Usages: []core.ConditionUsage{{ParameterKey: "flag"}}},
		{Name: "other", Expression: "release audience", Usages: []core.ConditionUsage{{ParameterKey: "flag"}}},
	}

	filtered := filterEntries(entries, []string{"^u"}, "release")
	filtered = filterEntriesByExpr(project, filtered, `usage_count > 0`)
	if len(filtered) != 1 || filtered[0].Name != "used" {
		t.Fatalf("combined condition filters = %#v, want used", filtered)
	}
}

func TestConditionListJSONIsPlainConditionArray(t *testing.T) {
	items := []core.ConditionEntry{{Name: "staff", Expression: "true"}}
	raw, err := json.Marshal(items)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(raw), "[") || strings.Contains(string(raw), `"project"`) || strings.Contains(string(raw), `"version"`) || strings.Contains(string(raw), `"source"`) || strings.Contains(string(raw), `"has_draft"`) {
		t.Fatalf("condition list JSON = %s", raw)
	}
	if len(items) != 1 || items[0].Name != "staff" {
		t.Fatalf("condition list items = %#v", items)
	}
	emptyRaw, err := json.Marshal(make([]core.ConditionEntry, 0))
	if err != nil || string(emptyRaw) != "[]" {
		t.Fatalf("empty condition list JSON = %s, err = %v", emptyRaw, err)
	}
}
