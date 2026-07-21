package config

import (
	"errors"
	"os"
	"reflect"
	"slices"

	coreconfig "github.com/yumauri/fbrcm/core/config"
)

// Load reads global config, merges missing keys, writes complete map if needed.
func Load() (State, error) {
	cfg, err := coreconfig.LoadAppConfig()
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return State{}, err
		}
		cfg = &coreconfig.AppConfig{}
	}
	changed := MigrateAdminShortcuts(cfg.Keys)
	if cfg.PowerlineGlyphs == nil {
		enabled := true
		cfg.PowerlineGlyphs = &enabled
		changed = true
	}
	powerlineGlyphs = *cfg.PowerlineGlyphs
	merged := Merge(DefaultKeyMap(), cfg.Keys)
	nextConfig := ToConfigMap(merged)
	if !reflect.DeepEqual(cfg.Keys, nextConfig) {
		cfg.Keys = nextConfig
		changed = true
	}
	if changed {
		if err := coreconfig.SaveAppConfig(cfg); err != nil {
			return State{}, err
		}
	}
	active = validate(merged)
	logConflicts(active)
	return Current(), nil
}

// migrateAdminShortcuts updates generated defaults from releases that predate
// the chorded Accounts and Profiles shortcuts.
func migrateAdminShortcuts(configured map[string]map[string][]string) bool {
	return MigrateAdminShortcuts(configured)
}

// MigrateAdminShortcuts updates recognized legacy generated bindings in place.
func MigrateAdminShortcuts(configured map[string]map[string][]string) bool {
	global := configured[string(BlockGlobal)]
	if global == nil {
		return false
	}

	changed := false
	if slices.Equal(global[string(ActionAccounts)], []string{"A"}) {
		global[string(ActionAccounts)] = []string{"ctrl+a"}
		changed = true
	}
	if slices.Equal(global[string(ActionProfiles)], []string{"P"}) {
		global[string(ActionProfiles)] = []string{"ctrl+p"}
		changed = true
	}
	for _, block := range []Block{BlockParameters, BlockConditions} {
		actions := configured[string(block)]
		if actions != nil && slices.Equal(actions[string(ActionPublishAll)], []string{"ctrl+p"}) {
			actions[string(ActionPublishAll)] = []string{"P"}
			changed = true
		}
	}
	return changed
}

// PowerlineGlyphsEnabled reports whether private-use Powerline separators
// should be used instead of standard Unicode triangle fallbacks.
func PowerlineGlyphsEnabled() bool { return powerlineGlyphs }

func merge(defaults KeyMap, configured map[string]map[string][]string) KeyMap {
	return Merge(defaults, configured)
}

// Merge applies configured bindings over a complete default key map.
func Merge(defaults KeyMap, configured map[string]map[string][]string) KeyMap {
	out := Clone(defaults)
	for blockName, actions := range configured {
		block := Block(blockName)
		defaultActions, ok := defaults[block]
		if !ok {
			continue
		}
		for actionName, keys := range actions {
			action := Action(actionName)
			if _, ok := defaultActions[action]; !ok {
				continue
			}
			clean := cleanKeys(keys)
			if len(clean) > 0 {
				out[block][action] = clean
			}
		}
	}
	return out
}

func cleanKeys(keys []string) []string {
	out := make([]string, 0, len(keys))
	seen := map[string]struct{}{}
	for _, k := range keys {
		if k == "" {
			continue
		}
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, k)
	}
	return out
}

func toConfigMap(m KeyMap) map[string]map[string][]string {
	return ToConfigMap(m)
}

// ToConfigMap converts a typed key map to its persisted representation.
func ToConfigMap(m KeyMap) map[string]map[string][]string {
	out := make(map[string]map[string][]string, len(m))
	for block, actions := range m {
		out[string(block)] = make(map[string][]string, len(actions))
		for action, keys := range actions {
			out[string(block)][string(action)] = append([]string(nil), keys...)
		}
	}
	return out
}

// CloneConfigMap returns a deep copy of a persisted keybinding map.
func CloneConfigMap(configured map[string]map[string][]string) map[string]map[string][]string {
	out := make(map[string]map[string][]string, len(configured))
	for block, actions := range configured {
		out[block] = make(map[string][]string, len(actions))
		for action, keys := range actions {
			out[block][action] = append([]string(nil), keys...)
		}
	}
	return out
}
