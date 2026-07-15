package conditions

import (
	"slices"
	"testing"

	"github.com/yumauri/fbrcm/core"
)

func TestFilterEntriesSupportsGetStylePrefixes(t *testing.T) {
	entries := []core.ConditionEntry{
		{Name: "production", Expression: "platform == 'android'"},
		{Name: "preview", Description: "Internal rollout"},
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
