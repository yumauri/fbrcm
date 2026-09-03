// Package mcpserver adapts fbrcm's machine contract to the MCP transport.
package mcpserver

import (
	"fmt"
	"slices"
	"time"

	"github.com/yumauri/fbrcm/ops"
	"github.com/yumauri/fbrcm/ops/contract"
)

// Options are immutable for the lifetime of one server process.
type Options struct {
	Profile        string
	Stateless      bool
	NoLocalConfig  bool
	AllowWrites    bool
	AllowHooks     bool
	Toolsets       []string
	Confirmation   string
	BrowserAuth    string
	RequestTimeout time.Duration
	AuthTimeout    time.Duration
}

func (o Options) Validate() error {
	if o.Stateless && o.Profile != "" {
		return fmt.Errorf("--profile cannot be used with --stateless")
	}
	if o.Stateless && o.AllowHooks {
		return fmt.Errorf("--allow-hooks cannot be used with --stateless")
	}
	if !slices.Contains([]string{"host", "none"}, o.Confirmation) {
		return fmt.Errorf("--confirmation must be host or none")
	}
	if !slices.Contains([]string{"auto", "never"}, o.BrowserAuth) {
		return fmt.Errorf("--browser-auth must be auto or never")
	}
	if o.RequestTimeout <= 0 || o.AuthTimeout <= 0 {
		return fmt.Errorf("--request-timeout and --auth-timeout must be greater than zero")
	}
	if len(o.Toolsets) == 0 {
		return fmt.Errorf("--toolsets must select at least one toolset")
	}
	for _, set := range o.Toolsets {
		if !slices.Contains([]string{"inspect", "edit", "drafts", "plans", "publish", "diagnostics"}, set) {
			return fmt.Errorf("unknown toolset %q", set)
		}
	}
	return nil
}

type catalogEntry struct {
	name     string
	set      string
	mutation bool
}

// Deliberately explicit: new CLI commands do not automatically become tools.
// Keys are shared operation IDs; names belong only to the MCP frontend.
var catalog = map[string]catalogEntry{
	"projects.list":         {"projects.list", "inspect", false},
	"projects.diff":         {"projects.diff", "inspect", false},
	"project.show":          {"project.show", "inspect", false},
	"get":                   {"parameters.get", "inspect", false},
	"groups.list":           {"groups.list", "inspect", false},
	"conditions.list":       {"conditions.list", "inspect", false},
	"conditions.show":       {"conditions.show", "inspect", false},
	"conditions.validate":   {"conditions.validate", "inspect", false},
	"versions.list":         {"versions.list", "inspect", false},
	"versions.show":         {"versions.show", "inspect", false},
	"versions.diff":         {"versions.diff", "inspect", false},
	"experiments.list":      {"experiments.list", "inspect", false},
	"experiments.show":      {"experiments.show", "inspect", false},
	"rollouts.list":         {"rollouts.list", "inspect", false},
	"rollouts.show":         {"rollouts.show", "inspect", false},
	"personalizations.list": {"personalizations.list", "inspect", false},
	"personalizations.show": {"personalizations.show", "inspect", false},
	"project.defaults":      {"project.defaults", "inspect", false},

	"add":               {"parameters.add", "edit", true},
	"update":            {"parameters.update", "edit", true},
	"delete":            {"parameters.delete", "edit", true},
	"duplicate":         {"parameters.duplicate", "edit", true},
	"groups.add":        {"groups.add", "edit", true},
	"groups.edit":       {"groups.edit", "edit", true},
	"groups.rename":     {"groups.rename", "edit", true},
	"groups.delete":     {"groups.delete", "edit", true},
	"conditions.add":    {"conditions.add", "edit", true},
	"conditions.edit":   {"conditions.edit", "edit", true},
	"conditions.rename": {"conditions.rename", "edit", true},
	"conditions.move":   {"conditions.move", "edit", true},
	"conditions.delete": {"conditions.delete", "edit", true},

	"draft.list":        {"draft.list", "drafts", false},
	"draft.show":        {"draft.show", "drafts", false},
	"draft.diff":        {"draft.diff", "drafts", false},
	"draft.change-note": {"draft.change-note", "drafts", true},
	"draft.discard":     {"draft.discard", "drafts", true},

	"plan.show":     {"plan.show", "plans", false},
	"plan.validate": {"plan.validate", "plans", false},

	"apply":              {"plan.apply", "publish", true},
	"draft.publish":      {"draft.publish", "publish", true},
	"project.import":     {"project.import", "publish", true},
	"project.export":     {"project.export", "publish", true},
	"versions.export":    {"versions.export", "publish", true},
	"projects.promote":   {"projects.promote", "publish", true},
	"versions.rollback":  {"versions.rollback", "publish", true},
	"versions.restore":   {"versions.restore", "publish", true},
	"experiments.delete": {"experiments.delete", "publish", true},
	"rollouts.delete":    {"rollouts.delete", "publish", true},

	"doctor": {"diagnostics.doctor", "diagnostics", false},
}

func (o Options) allows(c contract.Capability) bool {
	entry, ok := catalog[c.ID]
	return ok && slices.Contains(o.Toolsets, entry.set) && (!entry.mutation || o.AllowWrites) && (!o.Stateless || c.Supports.Stateless)
}

func boundOption(name string) bool {
	return ops.BoundOption(name)
}
