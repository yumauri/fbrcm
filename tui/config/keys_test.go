package config

import "testing"

func TestMatchesDefaultKeyMap(t *testing.T) {
	tests := []struct {
		block  Block
		action Action
		key    string
		want   bool
	}{
		{BlockGlobal, ActionQuit, "q", true},
		{BlockGlobal, ActionHelp, "?", true},
		{BlockGlobal, ActionQuit, "Q", false},
		{BlockGlobal, ActionFocusConditions, "3", true},
		{BlockGlobal, ActionFocusHistory, "4", true},
		{BlockGlobal, ActionFocusDetails, "5", true},
		{BlockConditions, ActionColor, "c", true},
		{BlockFilter, ActionFilterFuzzy, "~", true},
		{BlockParameters, ActionPublish, "p", true},
		{BlockJSONInput, ActionSave, "ctrl+s", true},
		{BlockJSONInput, ActionSave, "ctrl+enter", true},
	}

	for _, tc := range tests {
		if got := Matches(tc.block, tc.action, tc.key); got != tc.want {
			t.Errorf("Matches(%q, %q, %q) = %v, want %v", tc.block, tc.action, tc.key, got, tc.want)
		}
	}
}

func TestKeyHint(t *testing.T) {
	tests := map[string]string{
		"0": "⁰", "1": "¹", "2": "²", "3": "³", "4": "⁴",
		"5": "⁵", "6": "⁶", "7": "⁷", "8": "⁸", "9": "⁹",
		"a": "ᵃ", "b": "ᵇ", "c": "ᶜ", "d": "ᵈ", "e": "ᵉ",
		"f": "ᶠ", "g": "ᵍ", "h": "ʰ", "i": "ⁱ", "j": "ʲ",
		"k": "ᵏ", "l": "ˡ", "m": "ᵐ", "n": "ⁿ", "o": "ᵒ",
		"p": "ᵖ", "q": "q", "r": "ʳ", "s": "ˢ", "t": "ᵗ",
		"u": "ᵘ", "v": "ᵛ", "w": "ʷ", "x": "ˣ", "y": "ʸ", "z": "ᶻ",
		"Q":      "Q",
		"ctrl+x": "ctrl+x",
		"tab":    "tab",
	}
	for key, want := range tests {
		if got := KeyHint(key); got != want {
			t.Errorf("KeyHint(%q) = %q, want %q", key, got, want)
		}
	}
}

func TestActionKeyHintUsesActivePrimaryBinding(t *testing.T) {
	previous := active
	t.Cleanup(func() { active = previous })
	keys := Clone(DefaultKeyMap())
	keys[BlockGlobal][ActionFocusParameters] = []string{"x", "2"}
	active = validate(keys)

	if got := ActionKeyHint(BlockGlobal, ActionFocusParameters); got != "ˣ" {
		t.Fatalf("ActionKeyHint = %q, want ˣ", got)
	}
}

func TestValidateDisablesConflictingActions(t *testing.T) {
	keys := Clone(DefaultKeyMap())
	keys[BlockProjects][ActionRefresh] = []string{"u", "enter"}
	keys[BlockProjects][ActionSelect] = []string{"enter"}

	state := validate(keys)

	if !state.Disabled(BlockProjects, ActionRefresh) {
		t.Fatal("expected ActionRefresh to be disabled due to enter conflict")
	}
	if !state.Disabled(BlockProjects, ActionSelect) {
		t.Fatal("expected ActionSelect to be disabled due to enter conflict")
	}
	if state.Matches(BlockProjects, ActionRefresh, "u") {
		t.Fatal("expected disabled action not to match")
	}
	if state.Matches(BlockProjects, ActionSelect, "enter") {
		t.Fatal("expected disabled action not to match conflicting key")
	}
	if !state.Matches(BlockProjects, ActionOpen, "o") {
		t.Fatal("expected non-conflicting action to still match")
	}
}

func TestFilterModeForKey(t *testing.T) {
	if mode, ok := FilterModeForKey("~"); !ok || mode != "fuzzy" {
		t.Fatalf("FilterModeForKey(~) = %q, %v; want fuzzy, true", mode, ok)
	}
	if _, ok := FilterModeForKey("q"); ok {
		t.Fatal("expected quit key not to map to a filter mode")
	}
}
