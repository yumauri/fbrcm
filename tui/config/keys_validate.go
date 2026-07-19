package config

import (
	"slices"
	"strings"
	"unicode"

	"charm.land/bubbles/v2/key"

	corefilter "github.com/yumauri/fbrcm/core/filter"
	corelog "github.com/yumauri/fbrcm/core/log"
	"github.com/yumauri/fbrcm/core/strfold"
)

// Current returns active keymap state.
func Current() State {
	return State{
		keys:     Clone(active.keys),
		disabled: cloneDisabled(active.disabled),
	}
}

// Matches reports whether key triggers action in active keymap.
func Matches(block Block, action Action, k string) bool {
	return active.Matches(block, action, k)
}

// Matches reports whether key triggers action.
func (s State) Matches(block Block, action Action, k string) bool {
	if s.Disabled(block, action) {
		return false
	}
	return slices.Contains(s.keys[block][action], k)
}

// Disabled reports whether action is disabled due to conflict.
func Disabled(block Block, action Action) bool {
	return active.Disabled(block, action)
}

// Disabled reports whether action is disabled due to conflict.
func (s State) Disabled(block Block, action Action) bool {
	_, ok := s.disabled[actionRef{block: block, action: action}]
	return ok
}

// Keys returns keys for action.
func Keys(block Block, action Action) []string {
	if Disabled(block, action) {
		return nil
	}
	return append([]string(nil), active.keys[block][action]...)
}

// Binding returns help binding for action.
func Binding(block Block, action Action, desc string) key.Binding {
	keys := Keys(block, action)
	binding := key.NewBinding(key.WithKeys(keys...), key.WithHelp(Label(block, action), desc))
	binding.SetEnabled(len(keys) > 0)
	return binding
}

// Label returns compact help label for action.
func Label(block Block, action Action) string {
	return strings.Join(Keys(block, action), "/")
}

// ActionKeyHint returns a compact title hint for the action's primary key.
func ActionKeyHint(block Block, action Action) string {
	keys := Keys(block, action)
	if len(keys) == 0 {
		return ""
	}
	return KeyHint(keys[0])
}

// KeyHint renders compact panel-title hints. Single-character keys use
// superscripts, and Ctrl plus a single character uses a modifier circumflex
// followed by the same superscript (for example, ctrl+t becomes ˆᵗ).
func KeyHint(key string) string {
	if suffix, ok := strings.CutPrefix(key, "ctrl+"); ok {
		if hint, ok := superscriptHint(suffix); ok {
			return "ˆ" + hint
		}
		return key
	}
	if hint, ok := superscriptHint(key); ok {
		return hint
	}
	return key
}

func superscriptHint(key string) (string, bool) {
	runes := []rune(key)
	if len(runes) != 1 {
		return "", false
	}
	if hint, ok := superscriptKeyHints[runes[0]]; ok {
		return hint, true
	}
	if hint, ok := superscriptKeyHints[unicode.ToLower(runes[0])]; ok {
		return hint, true
	}
	return "", false
}

var superscriptKeyHints = map[rune]string{
	'?': "ˀ",
	'0': "⁰", '1': "¹", '2': "²", '3': "³", '4': "⁴",
	'5': "⁵", '6': "⁶", '7': "⁷", '8': "⁸", '9': "⁹",
	'a': "ᵃ", 'b': "ᵇ", 'c': "ᶜ", 'd': "ᵈ", 'e': "ᵉ",
	'f': "ᶠ", 'g': "ᵍ", 'h': "ʰ", 'i': "ⁱ", 'j': "ʲ",
	'k': "ᵏ", 'l': "ˡ", 'm': "ᵐ", 'n': "ⁿ", 'o': "ᵒ",
	'p': "ᵖ", 'r': "ʳ", 's': "ˢ", 't': "ᵗ", 'u': "ᵘ",
	'v': "ᵛ", 'w': "ʷ", 'x': "ˣ", 'y': "ʸ", 'z': "ᶻ",
	'A': "ᴬ", 'B': "ᴮ", 'D': "ᴰ", 'E': "ᴱ", 'G': "ᴳ",
	'H': "ᴴ", 'I': "ᴵ", 'J': "ᴶ", 'K': "ᴷ", 'L': "ᴸ",
	'M': "ᴹ", 'N': "ᴺ", 'O': "ᴼ", 'P': "ᴾ", 'R': "ᴿ",
	'T': "ᵀ", 'U': "ᵁ", 'V': "ⱽ", 'W': "ᵂ",
}

// FilterModeForKey returns filter mode configured for key.
func FilterModeForKey(k string) (corefilter.Mode, bool) {
	switch {
	case Matches(BlockFilter, ActionFilterFuzzy, k):
		return corefilter.ModeFuzzy, true
	case Matches(BlockFilter, ActionFilterStartsWith, k):
		return corefilter.ModeStartsWith, true
	case Matches(BlockFilter, ActionFilterIncludes, k):
		return corefilter.ModeIncludes, true
	case Matches(BlockFilter, ActionFilterExact, k):
		return corefilter.ModeExact, true
	default:
		return "", false
	}
}

// Clone returns deep copy of keymap.
func Clone(m KeyMap) KeyMap {
	out := make(KeyMap, len(m))
	for block, actions := range m {
		out[block] = make(map[Action][]string, len(actions))
		for action, keys := range actions {
			out[block][action] = append([]string(nil), keys...)
		}
	}
	return out
}

func validate(keys KeyMap) State {
	disabled := map[actionRef]struct{}{}
	for block, actions := range keys {
		byKey := map[string][]Action{}
		for action, actionKeys := range actions {
			for _, k := range actionKeys {
				byKey[k] = append(byKey[k], action)
			}
		}
		for _, conflictActions := range byKey {
			if len(conflictActions) < 2 {
				continue
			}
			for _, action := range conflictActions {
				disabled[actionRef{block: block, action: action}] = struct{}{}
			}
		}
	}
	return State{keys: Clone(keys), disabled: disabled}
}

func logConflicts(state State) {
	logger := corelog.For("tui.config")
	for _, conflict := range conflicts(state) {
		logger.Error("keybinding conflict", "block", conflict.block, "key", conflict.key, "actions", strings.Join(conflict.actions, ","))
	}
}

type conflict struct {
	block   Block
	key     string
	actions []string
}

func conflicts(state State) []conflict {
	var out []conflict
	for block, actions := range state.keys {
		byKey := map[string][]Action{}
		for action, keys := range actions {
			if !state.Disabled(block, action) {
				continue
			}
			for _, k := range keys {
				byKey[k] = append(byKey[k], action)
			}
		}
		for k, conflictActions := range byKey {
			conflictActions = slices.DeleteFunc(conflictActions, func(action Action) bool {
				count := 0
				for _, other := range actions {
					if slices.Contains(other, k) {
						count++
					}
				}
				return count < 2
			})
			if len(conflictActions) < 2 {
				continue
			}
			actionNames := make([]string, 0, len(conflictActions))
			for _, action := range conflictActions {
				actionNames = append(actionNames, string(action))
			}
			sortStringsFold(actionNames)
			out = append(out, conflict{block: block, key: k, actions: actionNames})
		}
	}
	slices.SortFunc(out, func(left, right conflict) int {
		if cmp := compareStringsFold(string(left.block), string(right.block)); cmp != 0 {
			return cmp
		}
		return compareStringsFold(left.key, right.key)
	})
	return out
}

func sortStringsFold(values []string) {
	strfold.Sort(values)
}

func compareStringsFold(left, right string) int {
	return strfold.Compare(left, right)
}

func cloneDisabled(disabled map[actionRef]struct{}) map[actionRef]struct{} {
	out := make(map[actionRef]struct{}, len(disabled))
	for ref := range disabled {
		out[ref] = struct{}{}
	}
	return out
}
