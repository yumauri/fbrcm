package config

import (
	"errors"
	"os"
	"reflect"

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
	changed := false
	if cfg.PowerlineGlyphs == nil {
		enabled := true
		cfg.PowerlineGlyphs = &enabled
		changed = true
	}
	powerlineGlyphs = *cfg.PowerlineGlyphs
	merged := merge(DefaultKeyMap(), cfg.Keys)
	nextConfig := toConfigMap(merged)
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

// PowerlineGlyphsEnabled reports whether private-use Powerline separators
// should be used instead of standard Unicode triangle fallbacks.
func PowerlineGlyphsEnabled() bool { return powerlineGlyphs }

func merge(defaults KeyMap, configured map[string]map[string][]string) KeyMap {
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
	out := make(map[string]map[string][]string, len(m))
	for block, actions := range m {
		out[string(block)] = make(map[string][]string, len(actions))
		for action, keys := range actions {
			out[string(block)][string(action)] = append([]string(nil), keys...)
		}
	}
	return out
}
