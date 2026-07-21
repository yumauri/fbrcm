package config

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Diagnostic describes one keybinding configuration problem.
type Diagnostic struct {
	Severity string `json:"severity"`
	Code     string `json:"code"`
	Key      string `json:"key"`
	Message  string `json:"message"`
}

// ValidateConfiguredKeys validates persisted bindings against the current TUI
// schema and returns every problem in stable order.
func ValidateConfiguredKeys(configured map[string]map[string][]string) []Diagnostic {
	defaults := DefaultKeyMap()
	migrated := CloneConfigMap(configured)
	didMigrate := MigrateAdminShortcuts(migrated)
	var diagnostics []Diagnostic
	for blockName, actions := range configured {
		block := Block(blockName)
		defaultActions, ok := defaults[block]
		if !ok {
			diagnostics = append(diagnostics, Diagnostic{
				Severity: "error", Code: "unknown_block", Key: "keys." + blockName,
				Message: fmt.Sprintf("unknown keybinding block %q", blockName),
			})
			continue
		}
		for actionName, keys := range actions {
			action := Action(actionName)
			path := "keys." + blockName + "." + actionName
			if _, ok := defaultActions[action]; !ok {
				diagnostics = append(diagnostics, Diagnostic{
					Severity: "error", Code: "unknown_action", Key: path,
					Message: fmt.Sprintf("unknown action %q in block %q", actionName, blockName),
				})
				continue
			}
			if len(keys) == 0 {
				diagnostics = append(diagnostics, Diagnostic{
					Severity: "error", Code: "empty_binding", Key: path,
					Message: "binding list cannot be empty; reset the action to restore its default",
				})
				continue
			}
			seen := map[string]struct{}{}
			for _, keyName := range keys {
				if _, ok := seen[keyName]; ok {
					diagnostics = append(diagnostics, Diagnostic{
						Severity: "error", Code: "duplicate_binding", Key: path,
						Message: fmt.Sprintf("binding %q is listed more than once", keyName),
					})
					continue
				}
				seen[keyName] = struct{}{}
				if !validKeyName(keyName) {
					diagnostics = append(diagnostics, Diagnostic{
						Severity: "error", Code: "invalid_binding", Key: path,
						Message: fmt.Sprintf("unsupported key name %q", keyName),
					})
				}
			}
		}
	}

	merged := Merge(defaults, migrated)
	for _, conflict := range conflicts(validate(merged)) {
		diagnostics = append(diagnostics, Diagnostic{
			Severity: "error",
			Code:     "keybinding_conflict",
			Key:      "keys." + string(conflict.block),
			Message:  fmt.Sprintf("key %q conflicts between actions %s", conflict.key, strings.Join(conflict.actions, ", ")),
		})
	}

	if didMigrate {
		diagnostics = append(diagnostics, Diagnostic{
			Severity: "warning", Code: "legacy_bindings", Key: "keys",
			Message: "legacy generated administration bindings will be migrated to current defaults",
		})
	}

	slices.SortFunc(diagnostics, func(left, right Diagnostic) int {
		if left.Severity != right.Severity {
			return strings.Compare(left.Severity, right.Severity)
		}
		if left.Key != right.Key {
			return strings.Compare(left.Key, right.Key)
		}
		if left.Code != right.Code {
			return strings.Compare(left.Code, right.Code)
		}
		return strings.Compare(left.Message, right.Message)
	})
	return diagnostics
}

func validKeyName(name string) bool {
	if name == " " || name == "space" {
		return true
	}
	if name == "" || strings.TrimSpace(name) != name || strings.ContainsAny(name, "\t\r\n") {
		return false
	}
	if utf8.RuneCountInString(name) == 1 {
		r, _ := utf8.DecodeRuneInString(name)
		return unicode.IsPrint(r)
	}

	parts := strings.Split(name, "+")
	if len(parts) > 1 {
		modifiers := map[string]struct{}{}
		for _, modifier := range parts[:len(parts)-1] {
			if !supportedModifier(modifier) {
				return false
			}
			if _, duplicate := modifiers[modifier]; duplicate {
				return false
			}
			modifiers[modifier] = struct{}{}
		}
		name = parts[len(parts)-1]
		if name == "" {
			return false
		}
		if utf8.RuneCountInString(name) == 1 {
			r, _ := utf8.DecodeRuneInString(name)
			return unicode.IsPrint(r)
		}
	}

	if number, err := strconv.Atoi(strings.TrimPrefix(name, "f")); err == nil && strings.HasPrefix(name, "f") && number >= 1 && number <= 63 {
		return true
	}
	switch name {
	case "enter", "esc", "escape", "tab", "backspace", "delete", "insert",
		"up", "down", "left", "right", "pgup", "pgdown", "home", "end", "space",
		"begin", "find", "select", "kpenter", "kpequal", "kpmul", "kpplus", "kpcomma",
		"kpminus", "kpperiod", "kpdiv", "kp0", "kp1", "kp2", "kp3", "kp4", "kp5",
		"kp6", "kp7", "kp8", "kp9", "kpsep", "kpup", "kpdown", "kpleft", "kpright",
		"kppgup", "kppgdown", "kphome", "kpend", "kpinsert", "kpdelete", "kpbegin",
		"capslock", "scrolllock", "numlock", "printscreen", "pause", "menu", "mediaplay",
		"mediapause", "mediaplaypause", "mediareverse", "mediastop", "mediafastforward",
		"mediarewind", "medianext", "mediaprev", "mediarecord", "lowervol", "raisevol", "mute",
		"leftshift", "leftalt", "leftctrl", "leftsuper", "lefthyper", "leftmeta", "rightshift",
		"rightalt", "rightctrl", "rightsuper", "righthyper", "rightmeta", "isolevel3shift",
		"isolevel5shift":
		return true
	default:
		return false
	}
}

func supportedModifier(name string) bool {
	switch name {
	case "ctrl", "alt", "shift", "meta", "hyper", "super", "capslock", "scrolllock", "numlock":
		return true
	default:
		return false
	}
}

// KnownBlock reports whether block is configurable.
func KnownBlock(block string) bool {
	_, ok := DefaultKeyMap()[Block(block)]
	return ok
}

// KnownAction reports whether action belongs to block.
func KnownAction(block, action string) bool {
	actions, ok := DefaultKeyMap()[Block(block)]
	if !ok {
		return false
	}
	_, ok = actions[Action(action)]
	return ok
}
