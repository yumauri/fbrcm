package config

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/shared"
	coreconfig "github.com/yumauri/fbrcm/core/config"
	tuiconfig "github.com/yumauri/fbrcm/tui/config"
)

type configSetResult struct {
	Scope    string `json:"scope"`
	Key      string `json:"key"`
	Previous any    `json:"previous"`
	Value    any    `json:"value"`
	Changed  bool   `json:"changed"`
}

type configResetResult struct {
	Scope   string `json:"scope"`
	Key     string `json:"key"`
	Status  string `json:"status"`
	Changed bool   `json:"changed"`
}

func newSetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <key> <value>...",
		Short: "Set a global configuration value",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonOut, err := cmd.Flags().GetBool("json")
			if err != nil {
				return err
			}
			state, err := loadConfigState()
			if err != nil {
				return err
			}
			scope, err := readConfigScope(cmd, scopeGlobal, scopeGlobal, scopeLocal)
			if err != nil {
				return err
			}
			candidate := mutationCandidate(state, scope)
			previous, _, err := scopedConfigValue(state, scope, args[0])
			if err != nil {
				return err
			}
			value, err := setConfigValue(candidate, args[0], args[1:])
			if err != nil {
				return err
			}
			changed := !reflect.DeepEqual(previous, value)
			if changed {
				if err := validateAndSaveScoped(state, candidate, scope); err != nil {
					return err
				}
			}
			result := configSetResult{Scope: scope, Key: args[0], Previous: previous, Value: value, Changed: changed}
			if jsonOut {
				return shared.WriteJSON(cmd, result)
			}
			verb := "unchanged"
			if changed {
				verb = "updated"
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s %s: %s\n", verb, args[0], formatConfigValue(value))
			return err
		},
	}
	cmd.Flags().Bool("json", false, "Print update result as JSON")
	addScopeFlag(cmd, scopeGlobal)
	return cmd
}

func setConfigValue(cfg *coreconfig.AppConfig, key string, values []string) (any, error) {
	parts := strings.Split(strings.TrimSpace(key), ".")
	switch {
	case key == "profile":
		return nil, fmt.Errorf("profile is managed by `fbrcm profile switch <name>`")
	case key == "powerline_glyphs":
		if len(values) != 1 {
			return nil, fmt.Errorf("powerline_glyphs requires exactly one boolean value")
		}
		value, err := strconv.ParseBool(values[0])
		if err != nil || values[0] != "true" && values[0] != "false" {
			return nil, fmt.Errorf("powerline_glyphs must be true or false")
		}
		cfg.PowerlineGlyphs = &value
		return value, nil
	case len(parts) == 3 && parts[0] == "keys":
		block, action := parts[1], parts[2]
		if !tuiconfig.KnownBlock(block) {
			return nil, fmt.Errorf("unknown keybinding block %q", block)
		}
		if !tuiconfig.KnownAction(block, action) {
			return nil, fmt.Errorf("unknown action %q in block %q", action, block)
		}
		if len(values) == 0 {
			return nil, fmt.Errorf("keybinding requires at least one key")
		}
		if cfg.Keys[block] == nil {
			cfg.Keys[block] = map[string][]string{}
		}
		cfg.Keys[block][action] = append([]string(nil), values...)
		return append([]string(nil), values...), nil
	default:
		return nil, fmt.Errorf("unknown or non-settable config key %q", key)
	}
}

func newResetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset [key]",
		Short: "Reset global configuration to built-in defaults",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			yes, err := cmd.Flags().GetBool("yes")
			if err != nil {
				return err
			}
			jsonOut, err := cmd.Flags().GetBool("json")
			if err != nil {
				return err
			}
			state, err := loadConfigState()
			if err != nil {
				return err
			}
			scope, err := readConfigScope(cmd, scopeGlobal, scopeGlobal, scopeLocal)
			if err != nil {
				return err
			}
			candidate := mutationCandidate(state, scope)
			key := "preferences"
			if len(args) == 1 {
				key = args[0]
			}
			changed, err := resetConfigValue(candidate, key, len(args) == 0)
			if err != nil {
				return err
			}
			result := configResetResult{Scope: scope, Key: key, Status: "unchanged", Changed: changed}
			if !changed {
				return writeResetResult(cmd, jsonOut, result)
			}
			if !yes {
				confirm := shared.NewConfirmation(fmt.Sprintf("Remove %s override from %s configuration?", key, scope), shared.ConfirmationOptions{Destructive: true})
				confirm.Input = cmd.InOrStdin()
				confirm.Output = cmd.ErrOrStderr()
				ok, err := confirm.RunPrompt()
				if err != nil {
					return err
				}
				if !ok {
					result.Status = "canceled"
					result.Changed = false
					return writeResetResult(cmd, jsonOut, result)
				}
			}
			if err := validateAndSaveScoped(state, candidate, scope); err != nil {
				return err
			}
			result.Status = "reset"
			return writeResetResult(cmd, jsonOut, result)
		},
	}
	shared.AddYesFlag(cmd, "Reset without confirmation")
	cmd.Flags().Bool("json", false, "Print reset result as JSON")
	addScopeFlag(cmd, scopeGlobal)
	return cmd
}

func resetConfigValue(candidate *coreconfig.AppConfig, key string, allPreferences bool) (bool, error) {
	if allPreferences {
		changed := candidate.PowerlineGlyphs != nil || len(candidate.Keys) > 0
		candidate.PowerlineGlyphs = nil
		candidate.Keys = map[string]map[string][]string{}
		return changed, nil
	}
	parts := strings.Split(strings.TrimSpace(key), ".")
	switch {
	case key == "profile":
		return false, fmt.Errorf("profile is managed by `fbrcm profile switch <name>`")
	case key == "powerline_glyphs":
		changed := candidate.PowerlineGlyphs != nil
		candidate.PowerlineGlyphs = nil
		return changed, nil
	case key == "keys":
		changed := len(candidate.Keys) > 0
		candidate.Keys = map[string]map[string][]string{}
		return changed, nil
	case len(parts) == 2 && parts[0] == "keys":
		block := parts[1]
		if !tuiconfig.KnownBlock(block) {
			return false, fmt.Errorf("unknown keybinding block %q", block)
		}
		_, changed := candidate.Keys[block]
		delete(candidate.Keys, block)
		return changed, nil
	case len(parts) == 3 && parts[0] == "keys":
		block, action := parts[1], parts[2]
		if !tuiconfig.KnownBlock(block) {
			return false, fmt.Errorf("unknown keybinding block %q", block)
		}
		if !tuiconfig.KnownAction(block, action) {
			return false, fmt.Errorf("unknown action %q in block %q", action, block)
		}
		actions := candidate.Keys[block]
		_, changed := actions[action]
		delete(actions, action)
		if len(actions) == 0 {
			delete(candidate.Keys, block)
		}
		return changed, nil
	default:
		return false, fmt.Errorf("unknown or non-resettable config key %q", key)
	}
}

func mutableConfig(state configState) *coreconfig.AppConfig {
	return cloneAppConfig(state.Global)
}

func mutationCandidate(state configState, scope string) *coreconfig.AppConfig {
	if scope == scopeLocal {
		return cloneAppConfig(state.Local)
	}
	return mutableConfig(state)
}

func validateAndSaveScoped(state configState, candidate *coreconfig.AppConfig, scope string) error {
	if err := validateScopedCandidate(state, candidate, scope); err != nil {
		return err
	}
	if scope == scopeGlobal {
		return coreconfig.SaveAppConfig(candidate)
	}
	raw, err := coreconfig.MarshalAppConfig(candidate)
	if err != nil {
		return fmt.Errorf("encode local config: %w", err)
	}
	return coreconfig.SaveLocalAppConfigRaw(state.LocalPath, raw)
}

func validateScopedCandidate(state configState, candidate *coreconfig.AppConfig, scope string) error {
	path := state.GlobalPath
	global, local := candidate, state.Local
	if scope == scopeLocal {
		path = state.LocalPath
		global, local = state.Global, candidate
	}
	report := validateAppConfig(path, true, candidate)
	if !report.Valid {
		return invalidConfigError(report)
	}
	merged, err := coreconfig.MergeAppConfigs(global, local)
	if err != nil {
		return err
	}
	report = validateAppConfig("effective configuration", true, merged)
	if !report.Valid {
		return invalidConfigError(report)
	}
	return nil
}

func invalidConfigError(report configValidationResult) error {
	parts := make([]string, 0, len(report.Errors))
	for _, diagnostic := range report.Errors {
		parts = append(parts, diagnosticKey(diagnostic)+": "+diagnostic.Message)
	}
	return fmt.Errorf("config is invalid: %s", strings.Join(parts, "; "))
}

func writeResetResult(cmd *cobra.Command, jsonOut bool, result configResetResult) error {
	if jsonOut {
		return shared.WriteJSON(cmd, result)
	}
	_, err := fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", result.Status, result.Key)
	return err
}

func formatConfigValue(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprint(value)
	}
	return string(raw)
}
