package groups

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/core"
)

func TestGroupsCommandSurface(t *testing.T) {
	cmd := New(nil)
	want := map[string][]string{
		"list":   {"project", "filter", "search", "update", "json"},
		"add":    {"project", "description", "dry-run", "draft", "yes"},
		"edit":   {"project", "description", "no-description", "dry-run", "draft", "yes"},
		"rename": {"project", "dry-run", "draft", "yes"},
		"delete": {"project", "dry-run", "draft", "yes"},
	}
	if len(cmd.Commands()) != len(want) {
		t.Fatalf("subcommands = %d, want %d", len(cmd.Commands()), len(want))
	}
	for name, flags := range want {
		subcommand, _, err := cmd.Find([]string{name})
		if err != nil {
			t.Fatalf("find %s: %v", name, err)
		}
		for _, flag := range flags {
			found := subcommand.Flags().Lookup(flag)
			if found == nil {
				t.Errorf("groups %s missing --%s", name, flag)
			} else if flag == "project" && found.Shorthand != "p" {
				t.Errorf("groups %s --project shorthand = %q, want p", name, found.Shorthand)
			}
		}
	}
}

func TestGroupsCommandUsesGroupOnlyPositionals(t *testing.T) {
	cmd := New(nil)
	wantUse := map[string]string{
		"list": "list", "add": "add <name>", "edit": "edit <group>",
		"rename": "rename <group> <new-name>", "delete": "delete <group>",
	}
	for name, want := range wantUse {
		subcommand, _, err := cmd.Find([]string{name})
		if err != nil {
			t.Fatal(err)
		}
		if subcommand.Use != want {
			t.Errorf("groups %s Use = %q, want %q", name, subcommand.Use, want)
		}
	}
}

func TestGroupsJSONIncludesProjectContextPerGroup(t *testing.T) {
	raw, err := json.Marshal(groupsJSON([]projectGroup{{
		Project: core.Project{Name: "Demo", ProjectID: "demo"}, Version: "7", Source: "draft", HasDraft: true,
		Group: core.ParametersGroup{Key: "checkout", Description: "Checkout flags", Parameters: []core.ParametersEntry{{Key: "enabled"}}},
	}}))
	if err != nil {
		t.Fatal(err)
	}
	got := string(raw)
	for _, want := range []string{`"project":"Demo"`, `"project_id":"demo"`, `"version":"7"`, `"source":"draft"`, `"has_draft":true`, `"name":"checkout"`, `"description":"Checkout flags"`, `"parameter_count":1`} {
		if !strings.Contains(got, want) {
			t.Fatalf("multi-project JSON missing %s: %s", want, got)
		}
	}
	if strings.Contains(got, `"Key"`) {
		t.Fatalf("JSON leaked presentation model fields: %s", got)
	}
}
