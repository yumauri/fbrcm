package config

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	coreconfig "github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/ops/shared"
	tuiconfig "github.com/yumauri/fbrcm/tui/config"
)

type configShowResult struct {
	Scope        string                `json:"scope" contract:"enum=effective|global|local"`
	Path         string                `json:"path"`
	Exists       bool                  `json:"exists"`
	GlobalPath   string                `json:"global_path,omitempty"`
	GlobalExists bool                  `json:"global_exists,omitempty"`
	LocalPath    string                `json:"local_path,omitempty"`
	LocalExists  bool                  `json:"local_exists,omitempty"`
	Config       *coreconfig.AppConfig `json:"config"`
}

type configValueResult struct {
	Key    string          `json:"key"`
	Value  json.RawMessage `json:"value"`
	Source string          `json:"source" contract:"enum=absent|default|global|local|mixed|migrated"`
}

func newShowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show [key]",
		Short: "Show effective global configuration",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonOut, err := cmd.Flags().GetBool("json")
			if err != nil {
				return err
			}
			state, err := loadConfigState()
			if err != nil {
				return err
			}
			scope, err := readConfigScope(cmd, scopeEffective, scopeEffective, scopeGlobal, scopeLocal)
			if err != nil {
				return err
			}
			selected, path, exists := scopedConfig(state, scope)
			if len(args) == 1 {
				value, source, err := scopedConfigValue(state, scope, args[0])
				if err != nil {
					return err
				}
				if jsonOut {
					result, err := newConfigValueResult(args[0], value, source)
					if err != nil {
						return err
					}
					return shared.WriteJSON(cmd, result)
				}
				return writeConfigValue(cmd, args[0], value)
			}
			if jsonOut {
				return shared.WriteJSON(cmd, configShowResult{Scope: scope, Path: path, Exists: exists, GlobalPath: state.GlobalPath, GlobalExists: state.GlobalExists, LocalPath: state.LocalPath, LocalExists: state.LocalExists, Config: selected})
			}
			raw, err := coreconfig.MarshalAppConfig(selected)
			if err != nil {
				return fmt.Errorf("encode effective config: %w", err)
			}
			_, err = cmd.OutOrStdout().Write(raw)
			return err
		},
	}
	cmd.Flags().Bool("json", false, "Print configuration as JSON")
	addScopeFlag(cmd, scopeEffective)
	return cmd
}

func newConfigValueResult(key string, value any, source string) (configValueResult, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return configValueResult{}, fmt.Errorf("encode config value %s: %w", key, err)
	}
	return configValueResult{Key: key, Value: raw, Source: source}, nil
}

func scopedConfig(state configState, scope string) (*coreconfig.AppConfig, string, bool) {
	switch scope {
	case scopeGlobal:
		return state.Global, state.GlobalPath, state.GlobalExists
	case scopeLocal:
		return state.Local, state.LocalPath, state.LocalExists
	default:
		return state.Effective, "effective", state.GlobalExists || state.LocalExists
	}
}

func scopedConfigValue(state configState, scope, key string) (any, string, error) {
	if scope == scopeEffective {
		return configValue(state, key)
	}
	cfg, _, _ := scopedConfig(state, scope)
	parts := strings.Split(strings.TrimSpace(key), ".")
	source := scope
	switch {
	case key == "profile":
		if strings.TrimSpace(cfg.Profile) == "" {
			return nil, "absent", nil
		}
		return cfg.Profile, source, nil
	case key == "theme":
		if cfg.Theme == "" {
			return nil, "absent", nil
		}
		return cfg.Theme, source, nil
	case key == "powerline_glyphs":
		if cfg.PowerlineGlyphs == nil {
			return nil, "absent", nil
		}
		return *cfg.PowerlineGlyphs, source, nil
	case key == "keys":
		return cfg.Keys, source, nil
	case key == "network":
		if cfg.Network == nil {
			return nil, "absent", nil
		}
		return cfg.Network, source, nil
	case key == "hooks":
		if cfg.Hooks == nil {
			return nil, "absent", nil
		}
		return cfg.Hooks, source, nil
	case key == "projects":
		if cfg.Projects == nil {
			return nil, "absent", nil
		}
		return cfg.Projects, source, nil
	case key == "projects.aliases":
		aliases := coreconfig.CloneProjectAliases(cfg)
		if len(aliases) == 0 {
			return nil, "absent", nil
		}
		return aliases, source, nil
	case len(parts) == 3 && parts[0] == "projects" && parts[1] == "aliases":
		if err := coreconfig.ValidateProjectAliasName(parts[2]); err != nil {
			return nil, "", shared.InvalidArgument(err)
		}
		value, ok := coreconfig.CloneProjectAliases(cfg)[parts[2]]
		if !ok {
			return nil, "absent", nil
		}
		return value, source, nil
	case len(parts) == 2 && parts[0] == "hooks":
		if cfg.Hooks == nil {
			return nil, "absent", nil
		}
		switch parts[1] {
		case "timeout":
			if strings.TrimSpace(cfg.Hooks.Timeout) == "" {
				return nil, "absent", nil
			}
			return cfg.Hooks.Timeout, source, nil
		case "pre_publish":
			if cfg.Hooks.PrePublish == nil {
				return nil, "absent", nil
			}
			return append([]string(nil), cfg.Hooks.PrePublish...), source, nil
		case "post_publish":
			if cfg.Hooks.PostPublish == nil {
				return nil, "absent", nil
			}
			return append([]string(nil), cfg.Hooks.PostPublish...), source, nil
		default:
			return nil, "", shared.InvalidArgument(fmt.Errorf("unknown hook key %q", parts[1]))
		}
	case len(parts) == 2 && parts[0] == "network":
		if cfg.Network == nil {
			return nil, "absent", nil
		}
		switch parts[1] {
		case "max_concurrent_requests":
			if cfg.Network.MaxConcurrentRequests == nil {
				return nil, "absent", nil
			}
			return *cfg.Network.MaxConcurrentRequests, source, nil
		case "requests_per_minute":
			if cfg.Network.RequestsPerMinute == nil {
				return nil, "absent", nil
			}
			return *cfg.Network.RequestsPerMinute, source, nil
		case "rate_limit_cooldown":
			if strings.TrimSpace(cfg.Network.RateLimitCooldown) == "" {
				return nil, "absent", nil
			}
			return cfg.Network.RateLimitCooldown, source, nil
		case "retry":
			if !retryConfigPresent(cfg.Network.Retry) {
				return nil, "absent", nil
			}
			return cfg.Network.Retry, source, nil
		default:
			return nil, "", shared.InvalidArgument(fmt.Errorf("unknown network key %q", parts[1]))
		}
	case len(parts) == 3 && parts[0] == "network" && parts[1] == "retry":
		if cfg.Network == nil || cfg.Network.Retry == nil {
			return nil, "absent", nil
		}
		retry := cfg.Network.Retry
		switch parts[2] {
		case "max_attempts":
			if retry.MaxAttempts == nil {
				return nil, "absent", nil
			}
			return *retry.MaxAttempts, source, nil
		case "base_delay":
			if strings.TrimSpace(retry.BaseDelay) == "" {
				return nil, "absent", nil
			}
			return retry.BaseDelay, source, nil
		case "max_delay":
			if strings.TrimSpace(retry.MaxDelay) == "" {
				return nil, "absent", nil
			}
			return retry.MaxDelay, source, nil
		case "jitter_percent":
			if retry.JitterPercent == nil {
				return nil, "absent", nil
			}
			return *retry.JitterPercent, source, nil
		default:
			return nil, "", shared.InvalidArgument(fmt.Errorf("unknown network retry key %q", parts[2]))
		}
	case len(parts) == 2 && parts[0] == "keys":
		if !tuiconfig.KnownBlock(parts[1]) {
			return nil, "", shared.InvalidArgument(fmt.Errorf("unknown keybinding block %q", parts[1]))
		}
		return cfg.Keys[parts[1]], source, nil
	case len(parts) == 3 && parts[0] == "keys":
		if !tuiconfig.KnownAction(parts[1], parts[2]) {
			return nil, "", shared.InvalidArgument(fmt.Errorf("unknown action %q in block %q", parts[2], parts[1]))
		}
		value, ok := cfg.Keys[parts[1]][parts[2]]
		if !ok {
			return nil, "absent", nil
		}
		return append([]string(nil), value...), source, nil
	default:
		return nil, "", shared.InvalidArgument(fmt.Errorf("unknown config key %q", key))
	}
}

func writeConfigValue(cmd *cobra.Command, key string, value any) error {
	switch value := value.(type) {
	case nil:
		return nil
	case string:
		_, err := fmt.Fprintln(cmd.OutOrStdout(), value)
		return err
	case bool:
		_, err := fmt.Fprintln(cmd.OutOrStdout(), value)
		return err
	default:
		nested := value
		parts := strings.Split(key, ".")
		for _, part := range slices.Backward(parts) {
			nested = map[string]any{part: nested}
		}
		raw, err := coreconfig.MarshalTOML(nested)
		if err != nil {
			return fmt.Errorf("encode selected config: %w", err)
		}
		_, err = cmd.OutOrStdout().Write(raw)
		return err
	}
}
