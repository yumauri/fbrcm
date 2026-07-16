package app

import (
	"slices"
	"testing"

	"charm.land/bubbles/v2/key"

	"github.com/yumauri/fbrcm/tui/panels"
)

func TestShortHelpDescriptions(t *testing.T) {
	tests := []struct {
		name string
		keys helpKeyMap
		want []string
	}{
		{
			name: "keyboard capture",
			keys: helpKeyMap{keyboardCapture: true},
			want: []string{"quit", "close filter", "apply", "filter"},
		},
		{
			name: "projects expanded",
			keys: helpKeyMap{active: panels.Projects},
			want: []string{"quit", "collapse", "select", "mark", "open", "update", "filter"},
		},
		{
			name: "projects collapsed",
			keys: helpKeyMap{active: panels.Projects, projectsMode: projectsPanelModeCollapsed},
			want: []string{"quit", "expand", "select", "mark", "open", "update", "filter"},
		},
		{
			name: "parameters",
			keys: helpKeyMap{active: panels.Parameters},
			want: []string{"quit", "maximize", "rename", "edit", "new", "duplicate", "move", "toggle", "delete", "publish", "discard", "copy", "update", "filter"},
		},
		{
			name: "conditions",
			keys: helpKeyMap{active: panels.Conditions},
			want: []string{"quit", "maximize", "rename", "expression", "color", "new", "priority", "delete", "publish", "discard", "details", "copy", "update", "filter"},
		},
		{
			name: "condition move",
			keys: helpKeyMap{active: panels.Conditions, conditionMove: true},
			want: []string{"move up", "move down", "place", "cancel"},
		},
		{
			name: "condition details",
			keys: helpKeyMap{active: panels.Details, conditionDetail: true},
			want: []string{"quit", "close", "rename", "expression", "color", "priority", "delete", "copy", "copy expression"},
		},
		{
			name: "logs expanded",
			keys: helpKeyMap{active: panels.Logs},
			want: []string{"quit", "collapse", "level", "resize"},
		},
		{
			name: "logs collapsed",
			keys: helpKeyMap{active: panels.Logs, logsMode: logsPanelModeCollapsed},
			want: []string{"quit", "expand", "level", "resize"},
		},
		{
			name: "details",
			keys: helpKeyMap{active: panels.Details},
			want: []string{"quit", "close", "rename", "edit", "move", "delete", "copy", "copy value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := helpDescriptions(tt.keys.ShortHelp()); !slices.Equal(got, tt.want) {
				t.Fatalf("descriptions = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompoundHelpKeys(t *testing.T) {
	binding := parametersHelp()[8]
	help := binding.Help()

	if help.Desc != "publish" {
		t.Fatalf("desc = %q, want publish", help.Desc)
	}
	if help.Key != "p/P" {
		t.Fatalf("key label = %q, want p/P", help.Key)
	}
	if got := binding.Keys(); !slices.Equal(got, []string{"p", "P"}) {
		t.Fatalf("keys = %v, want [p P]", got)
	}
}

func helpDescriptions(bindings []key.Binding) []string {
	out := make([]string, 0, len(bindings))
	for _, binding := range bindings {
		out = append(out, binding.Help().Desc)
	}
	return out
}
