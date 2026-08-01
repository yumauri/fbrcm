package config

import (
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/shared"
	coreconfig "github.com/yumauri/fbrcm/core/config"
	tuiconfig "github.com/yumauri/fbrcm/tui/config"
)

type configShowResult struct {
	Scope        string                `json:"scope"`
	Path         string                `json:"path"`
	Exists       bool                  `json:"exists"`
	GlobalPath   string                `json:"global_path,omitempty"`
	GlobalExists bool                  `json:"global_exists,omitempty"`
	LocalPath    string                `json:"local_path,omitempty"`
	LocalExists  bool                  `json:"local_exists,omitempty"`
	Config       *coreconfig.AppConfig `json:"config"`
}

type configValueResult struct {
	Key    string `json:"key"`
	Value  any    `json:"value"`
	Source string `json:"source"`
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
					return shared.WriteJSON(cmd, configValueResult{Key: args[0], Value: value, Source: source})
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
	case key == "powerline_glyphs":
		if cfg.PowerlineGlyphs == nil {
			return nil, "absent", nil
		}
		return *cfg.PowerlineGlyphs, source, nil
	case key == "keys":
		return cfg.Keys, source, nil
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
			return nil, "", fmt.Errorf("unknown hook key %q", parts[1])
		}
	case len(parts) == 2 && parts[0] == "keys":
		if !tuiconfig.KnownBlock(parts[1]) {
			return nil, "", fmt.Errorf("unknown keybinding block %q", parts[1])
		}
		return cfg.Keys[parts[1]], source, nil
	case len(parts) == 3 && parts[0] == "keys":
		if !tuiconfig.KnownAction(parts[1], parts[2]) {
			return nil, "", fmt.Errorf("unknown action %q in block %q", parts[2], parts[1])
		}
		value, ok := cfg.Keys[parts[1]][parts[2]]
		if !ok {
			return nil, "absent", nil
		}
		return append([]string(nil), value...), source, nil
	default:
		return nil, "", fmt.Errorf("unknown config key %q", key)
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
