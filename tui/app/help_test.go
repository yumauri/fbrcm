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
			want: []string{"quit", "help", "close filter", "apply", "filter"},
		},
		{
			name: "projects expanded",
			keys: helpKeyMap{active: panels.Projects, canBindAuth: true},
			want: []string{"quit", "help", "collapse", "select", "mark", "bind auth", "open", "update", "filter"},
		},
		{
			name: "projects collapsed",
			keys: helpKeyMap{active: panels.Projects, projectsMode: projectsPanelModeCollapsed, canBindAuth: true},
			want: []string{"quit", "help", "expand", "select", "mark", "bind auth", "open", "update", "filter"},
		},
		{
			name: "parameters",
			keys: helpKeyMap{active: panels.Parameters},
			want: []string{"quit", "help", "maximize", "rename", "edit", "new", "duplicate", "move", "toggle", "delete", "publish", "discard", "copy", "update", "filter"},
		},
		{
			name: "conditions",
			keys: helpKeyMap{active: panels.Conditions},
			want: []string{"quit", "help", "maximize", "rename", "expression", "color", "new", "priority", "delete", "publish", "discard", "details", "copy", "update", "filter"},
		},
		{
			name: "condition move",
			keys: helpKeyMap{active: panels.Conditions, conditionMove: true},
			want: []string{"move up", "move down", "place", "cancel"},
		},
		{
			name: "condition details",
			keys: helpKeyMap{active: panels.Details, conditionDetail: true},
			want: []string{"quit", "help", "close", "rename", "expression", "color", "priority", "delete", "copy", "copy expression"},
		},
		{
			name: "group details",
			keys: helpKeyMap{active: panels.Details, groupDetail: true},
			want: []string{"quit", "help", "close", "rename", "delete", "copy"},
		},
		{
			name: "logs expanded",
			keys: helpKeyMap{active: panels.Logs},
			want: []string{"quit", "help", "collapse", "level", "resize"},
		},
		{
			name: "logs collapsed",
			keys: helpKeyMap{active: panels.Logs, logsMode: logsPanelModeCollapsed},
			want: []string{"quit", "help", "expand", "level", "resize"},
		},
		{
			name: "details",
			keys: helpKeyMap{active: panels.Details},
			want: []string{"quit", "help", "close", "add conditional value", "rename", "edit", "move", "delete", "copy", "copy value"},
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

func TestProjectHelpOmitsBindAuthenticationWithoutMultipleIdentities(t *testing.T) {
	got := helpDescriptions(helpKeyMap{active: panels.Projects}.ShortHelp())
	if slices.Contains(got, "bind auth") {
		t.Fatalf("single-auth project help includes bind action: %v", got)
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

func TestFullHelpIncludesActionsOmittedFromFooter(t *testing.T) {
	full := helpKeyMap{active: panels.Projects}.FullHelp()
	if len(full) <= 1 {
		t.Fatalf("full help groups = %d, want grouped actions", len(full))
	}
	var descriptions []string
	for _, group := range full {
		descriptions = append(descriptions, helpDescriptions(group)...)
	}
	for _, want := range []string{"focus logs", "collapse all", "blank line", "format"} {
		if !slices.Contains(descriptions, want) {
			t.Errorf("full help descriptions do not contain %q", want)
		}
	}
}

func helpDescriptions(bindings []key.Binding) []string {
	out := make([]string, 0, len(bindings))
	for _, binding := range bindings {
		out = append(out, binding.Help().Desc)
	}
	return out
}
