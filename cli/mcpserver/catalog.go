// Package mcpserver adapts fbrcm's machine contract to the MCP transport.
package mcpserver

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/yumauri/fbrcm/cli/contract"
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
	set      string
	mutation bool
}

// Deliberately explicit: new CLI commands do not automatically become tools.
var catalog = map[string]catalogEntry{
	"projects.list": {"inspect", false}, "projects.diff": {"inspect", false},
	"project.show": {"inspect", false}, "get": {"inspect", false},
	"groups.list": {"inspect", false}, "conditions.list": {"inspect", false},
	"conditions.show": {"inspect", false}, "conditions.validate": {"inspect", false},
	"versions.list": {"inspect", false}, "versions.show": {"inspect", false}, "versions.diff": {"inspect", false},
	"experiments.list": {"inspect", false}, "experiments.show": {"inspect", false},
	"rollouts.list": {"inspect", false}, "rollouts.show": {"inspect", false},
	"personalizations.list": {"inspect", false}, "personalizations.show": {"inspect", false},
	"project.defaults": {"inspect", false},
	"add":              {"edit", true}, "update": {"edit", true}, "delete": {"edit", true}, "duplicate": {"edit", true},
	"groups.add": {"edit", true}, "groups.edit": {"edit", true}, "groups.rename": {"edit", true}, "groups.delete": {"edit", true},
	"conditions.add": {"edit", true}, "conditions.edit": {"edit", true}, "conditions.rename": {"edit", true},
	"conditions.move": {"edit", true}, "conditions.delete": {"edit", true},
	"draft.list": {"drafts", false}, "draft.show": {"drafts", false}, "draft.diff": {"drafts", false},
	"draft.change-note": {"drafts", true}, "draft.discard": {"drafts", true},
	"plan.show": {"plans", false}, "plan.validate": {"plans", false},
	"apply": {"publish", true}, "draft.publish": {"publish", true},
	"project.import": {"publish", true}, "project.export": {"publish", true}, "versions.export": {"publish", true},
	"projects.promote": {"publish", true}, "versions.rollback": {"publish", true},
	"versions.restore":   {"publish", true},
	"experiments.delete": {"publish", true}, "rollouts.delete": {"publish", true},
	"doctor": {"diagnostics", false},
}

func (o Options) allows(c contract.Capability) bool {
	entry, ok := catalog[c.ID]
	return ok && slices.Contains(o.Toolsets, entry.set) && (!entry.mutation || o.AllowWrites) && (!o.Stateless || c.Supports.Stateless)
}

func boundOption(name string) bool {
	return slices.Contains([]string{"profile", "stateless", "json", "timeout", "no-local-config", "help", "yes"}, strings.TrimPrefix(name, "--"))
}
