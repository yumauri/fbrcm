package config

import (
	"slices"
	"testing"
)

func TestValidateConfiguredKeysReturnsAllSchemaAndBindingErrors(t *testing.T) {
	diagnostics := ValidateConfiguredKeys(map[string]map[string][]string{
		"missing": {"action": {"x"}},
		"projects": {
			"missing": {"x"},
			"refresh": {"ctrl+r", "ctrl+r", "ctrl++r"},
		},
	})

	var codes []string
	for _, diagnostic := range diagnostics {
		codes = append(codes, diagnostic.Code)
	}
	for _, want := range []string{"unknown_block", "unknown_action", "duplicate_binding", "invalid_binding"} {
		if !slices.Contains(codes, want) {
			t.Fatalf("diagnostic codes = %v, missing %q", codes, want)
		}
	}
}

func TestValidateConfiguredKeysDetectsEffectiveDefaultConflict(t *testing.T) {
	diagnostics := ValidateConfiguredKeys(map[string]map[string][]string{
		"projects": {"refresh": {"enter"}},
	})
	if !slices.ContainsFunc(diagnostics, func(d Diagnostic) bool { return d.Code == "keybinding_conflict" }) {
		t.Fatalf("diagnostics = %+v, want conflict", diagnostics)
	}
}

func TestValidateConfiguredKeysAcceptsSupportedTerminalKeys(t *testing.T) {
	for _, key := range []string{"x", "?", "space", "enter", "ctrl+r", "shift+tab", "alt+left", "f12", "meta+enter", "super+kp1"} {
		if !validKeyName(key) {
			t.Errorf("validKeyName(%q) = false", key)
		}
	}
}
