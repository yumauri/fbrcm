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
	Key      string `json:"key"`
	Previous any    `json:"previous"`
	Value    any    `json:"value"`
	Changed  bool   `json:"changed"`
}

type configResetResult struct {
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
			if !state.Report.Valid {
				return invalidConfigError(state.Report)
			}
			candidate := mutableConfig(state)
			previous, _, err := configValue(state, args[0])
			if err != nil {
				return err
			}
			value, err := setConfigValue(candidate, args[0], args[1:])
			if err != nil {
				return err
			}
			changed := !reflect.DeepEqual(previous, value)
			if changed {
				if err := validateAndSave(candidate); err != nil {
					return err
				}
			}
			result := configSetResult{Key: args[0], Previous: previous, Value: value, Changed: changed}
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
			candidate := mutableConfig(state)
			key := "preferences"
			if len(args) == 1 {
				key = args[0]
			}
			changed, err := resetConfigValue(candidate, state, key, len(args) == 0)
			if err != nil {
				return err
			}
			result := configResetResult{Key: key, Status: "unchanged", Changed: changed}
			if !changed {
				return writeResetResult(cmd, jsonOut, result)
			}
			if !yes {
				confirm := shared.NewConfirmation("Reset "+key+" to built-in defaults?", shared.ConfirmationOptions{Destructive: true})
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
			if err := validateAndSave(candidate); err != nil {
				return err
			}
			result.Status = "reset"
			return writeResetResult(cmd, jsonOut, result)
		},
	}
	shared.AddYesFlag(cmd, "Reset without confirmation")
	cmd.Flags().Bool("json", false, "Print reset result as JSON")
	return cmd
}

func resetConfigValue(candidate *coreconfig.AppConfig, state configState, key string, allPreferences bool) (bool, error) {
	defaults := tuiconfig.ToConfigMap(tuiconfig.DefaultKeyMap())
	if allPreferences {
		changed := !state.Report.Valid || candidate.PowerlineGlyphs == nil || !*candidate.PowerlineGlyphs || !reflect.DeepEqual(candidate.Keys, defaults)
		enabled := true
		candidate.PowerlineGlyphs = &enabled
		candidate.Keys = defaults
		return changed, nil
	}
	parts := strings.Split(strings.TrimSpace(key), ".")
	switch {
	case key == "profile":
		return false, fmt.Errorf("profile is managed by `fbrcm profile switch <name>`")
	case key == "powerline_glyphs":
		previous := *state.Effective.PowerlineGlyphs
		enabled := true
		candidate.PowerlineGlyphs = &enabled
		return !state.Report.Valid || previous != enabled, nil
	case key == "keys":
		changed := !state.Report.Valid || !reflect.DeepEqual(candidate.Keys, defaults)
		candidate.Keys = defaults
		return changed, nil
	case len(parts) == 2 && parts[0] == "keys":
		block := parts[1]
		if !tuiconfig.KnownBlock(block) {
			return false, fmt.Errorf("unknown keybinding block %q", block)
		}
		changed := !state.Report.Valid || !reflect.DeepEqual(candidate.Keys[block], defaults[block])
		candidate.Keys[block] = cloneActionMap(defaults[block])
		return changed, nil
	case len(parts) == 3 && parts[0] == "keys":
		block, action := parts[1], parts[2]
		if !tuiconfig.KnownBlock(block) {
			return false, fmt.Errorf("unknown keybinding block %q", block)
		}
		if !tuiconfig.KnownAction(block, action) {
			return false, fmt.Errorf("unknown action %q in block %q", action, block)
		}
		changed := !state.Report.Valid || !reflect.DeepEqual(candidate.Keys[block][action], defaults[block][action])
		candidate.Keys[block][action] = append([]string(nil), defaults[block][action]...)
		return changed, nil
	default:
		return false, fmt.Errorf("unknown or non-resettable config key %q", key)
	}
}

func mutableConfig(state configState) *coreconfig.AppConfig {
	candidate := cloneAppConfig(state.Effective)
	candidate.Profile = state.Stored.Profile
	return candidate
}

func validateAndSave(candidate *coreconfig.AppConfig) error {
	report := validateAppConfig(coreconfig.GetGlobalConfigFilePath(), true, candidate)
	if !report.Valid {
		return invalidConfigError(report)
	}
	return coreconfig.SaveAppConfig(candidate)
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

func cloneActionMap(actions map[string][]string) map[string][]string {
	out := make(map[string][]string, len(actions))
	for action, keys := range actions {
		out[action] = append([]string(nil), keys...)
	}
	return out
}
