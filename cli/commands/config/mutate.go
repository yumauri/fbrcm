package config

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	coreconfig "github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/ops/shared"
	tuiconfig "github.com/yumauri/fbrcm/tui/config"
)

type configSetResult struct {
	Scope    string          `json:"scope" contract:"enum=global|local"`
	Key      string          `json:"key"`
	Previous json.RawMessage `json:"previous"`
	Value    json.RawMessage `json:"value"`
	Changed  bool            `json:"changed"`
}

type configResetResult struct {
	Scope   string `json:"scope" contract:"enum=global|local"`
	Key     string `json:"key"`
	Status  string `json:"status" contract:"enum=unchanged|reset"`
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
			if err := requireLocalProjectAliasScope(args[0], scope); err != nil {
				return err
			}
			candidate := mutationCandidate(state, scope)
			previous, _, err := scopedConfigValue(state, scope, args[0])
			if err != nil {
				return err
			}
			value, err := setConfigValue(candidate, args[0], args[1:])
			if err != nil {
				return shared.InvalidArgument(err)
			}
			changed := !reflect.DeepEqual(previous, value)
			if changed {
				if err := validateAndSaveScoped(state, candidate, scope); err != nil {
					return err
				}
			}
			result, err := newConfigSetResult(scope, args[0], previous, value, changed)
			if err != nil {
				return err
			}
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

func newConfigSetResult(scope, key string, previous, value any, changed bool) (configSetResult, error) {
	previousRaw, err := json.Marshal(previous)
	if err != nil {
		return configSetResult{}, fmt.Errorf("encode previous config value %s: %w", key, err)
	}
	valueRaw, err := json.Marshal(value)
	if err != nil {
		return configSetResult{}, fmt.Errorf("encode config value %s: %w", key, err)
	}
	return configSetResult{Scope: scope, Key: key, Previous: previousRaw, Value: valueRaw, Changed: changed}, nil
}

func setConfigValue(cfg *coreconfig.AppConfig, key string, values []string) (any, error) {
	parts := strings.Split(strings.TrimSpace(key), ".")
	switch {
	case key == "profile":
		return nil, fmt.Errorf("profile is managed by `fbrcm profile switch <name>`")
	case key == "theme":
		if len(values) != 1 {
			return nil, fmt.Errorf("theme requires exactly one theme name")
		}
		if err := coreconfig.ValidateThemeName(values[0]); err != nil {
			return nil, err
		}
		cfg.Theme = values[0]
		return values[0], nil
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
	case len(parts) == 2 && parts[0] == "network" && parts[1] == "max_concurrent_requests":
		value, err := parseBoundedNetworkInteger(values, key, 1, coreconfig.MaxConcurrentRequests)
		if err != nil {
			return nil, err
		}
		ensureNetworkConfig(cfg).MaxConcurrentRequests = &value
		return value, nil
	case len(parts) == 2 && parts[0] == "network" && parts[1] == "requests_per_minute":
		if len(values) != 1 {
			return nil, fmt.Errorf("network.requests_per_minute requires exactly one integer value")
		}
		value, err := strconv.Atoi(values[0])
		if err != nil || value < 0 || value > coreconfig.MaxRequestsPerMinute {
			return nil, fmt.Errorf("network.requests_per_minute must be between 0 and %d", coreconfig.MaxRequestsPerMinute)
		}
		if cfg.Network == nil {
			cfg.Network = &coreconfig.NetworkConfig{}
		}
		cfg.Network.RequestsPerMinute = &value
		return value, nil
	case len(parts) == 2 && parts[0] == "network" && parts[1] == "rate_limit_cooldown":
		if len(values) != 1 {
			return nil, fmt.Errorf("network.rate_limit_cooldown requires exactly one duration value")
		}
		candidate := &coreconfig.NetworkConfig{RateLimitCooldown: values[0]}
		if _, err := candidate.EffectiveRateLimitCooldown(); err != nil {
			return nil, err
		}
		if cfg.Network == nil {
			cfg.Network = &coreconfig.NetworkConfig{}
		}
		cfg.Network.RateLimitCooldown = values[0]
		return values[0], nil
	case len(parts) == 3 && parts[0] == "network" && parts[1] == "retry" && parts[2] == "max_attempts":
		value, err := parseBoundedNetworkInteger(values, key, 1, coreconfig.MaxRetryAttempts)
		if err != nil {
			return nil, err
		}
		ensureRetryConfig(cfg).MaxAttempts = &value
		return value, nil
	case len(parts) == 3 && parts[0] == "network" && parts[1] == "retry" && parts[2] == "base_delay":
		if len(values) != 1 {
			return nil, fmt.Errorf("%s requires exactly one duration value", key)
		}
		candidate := &coreconfig.RetryConfig{BaseDelay: values[0]}
		if _, err := candidate.EffectiveBaseDelay(); err != nil {
			return nil, err
		}
		ensureRetryConfig(cfg).BaseDelay = values[0]
		return values[0], nil
	case len(parts) == 3 && parts[0] == "network" && parts[1] == "retry" && parts[2] == "max_delay":
		if len(values) != 1 {
			return nil, fmt.Errorf("%s requires exactly one duration value", key)
		}
		candidate := &coreconfig.RetryConfig{MaxDelay: values[0]}
		if _, err := candidate.EffectiveMaxDelay(); err != nil {
			return nil, err
		}
		ensureRetryConfig(cfg).MaxDelay = values[0]
		return values[0], nil
	case len(parts) == 3 && parts[0] == "network" && parts[1] == "retry" && parts[2] == "jitter_percent":
		value, err := parseBoundedNetworkInteger(values, key, 0, 100)
		if err != nil {
			return nil, err
		}
		ensureRetryConfig(cfg).JitterPercent = &value
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
	case len(parts) == 3 && parts[0] == "projects" && parts[1] == "aliases":
		if len(values) != 1 {
			return nil, fmt.Errorf("project alias requires exactly one physical project ID")
		}
		alias, projectID := parts[2], values[0]
		if err := coreconfig.ValidateProjectAliasName(alias); err != nil {
			return nil, err
		}
		if err := coreconfig.ValidateProjectAliasProjectID(projectID); err != nil {
			return nil, fmt.Errorf("project alias %q: %w", alias, err)
		}
		if _, _, err := coreconfig.SetProjectAlias(cfg, alias, projectID); err != nil {
			return nil, err
		}
		return projectID, nil
	default:
		return nil, fmt.Errorf("unknown or non-settable config key %q", key)
	}
}

func parseBoundedNetworkInteger(values []string, key string, minimum, maximum int) (int, error) {
	if len(values) != 1 {
		return 0, fmt.Errorf("%s requires exactly one integer value", key)
	}
	value, err := strconv.Atoi(values[0])
	if err != nil || value < minimum || value > maximum {
		return 0, fmt.Errorf("%s must be between %d and %d", key, minimum, maximum)
	}
	return value, nil
}

func ensureNetworkConfig(cfg *coreconfig.AppConfig) *coreconfig.NetworkConfig {
	if cfg.Network == nil {
		cfg.Network = &coreconfig.NetworkConfig{}
	}
	return cfg.Network
}

func ensureRetryConfig(cfg *coreconfig.AppConfig) *coreconfig.RetryConfig {
	network := ensureNetworkConfig(cfg)
	if network.Retry == nil {
		network.Retry = &coreconfig.RetryConfig{}
	}
	return network.Retry
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
			keyArg := ""
			if len(args) == 1 {
				keyArg = args[0]
			}
			if err := requireLocalProjectAliasScope(keyArg, scope); err != nil {
				return err
			}
			candidate := mutationCandidate(state, scope)
			key := "preferences"
			if len(args) == 1 {
				key = args[0]
			}
			changed, err := resetConfigValue(candidate, key, len(args) == 0)
			if err != nil {
				return shared.InvalidArgument(err)
			}
			result := configResetResult{Scope: scope, Key: key, Status: "unchanged", Changed: changed}
			if !changed {
				return writeResetResult(cmd, jsonOut, result)
			}
			if !yes {
				if err := shared.RequireYesInMachineMode(cmd, yes, "removing the configuration override", true); err != nil {
					return err
				}
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

func requireLocalProjectAliasScope(key, scope string) error {
	if scope == scopeGlobal && (key == "projects.aliases" || strings.HasPrefix(key, "projects.aliases.")) {
		return shared.InvalidArgument(fmt.Errorf("projects.aliases: project aliases are repository-scoped; use --scope local"))
	}
	return nil
}

func resetConfigValue(candidate *coreconfig.AppConfig, key string, allPreferences bool) (bool, error) {
	if allPreferences {
		changed := candidate.Theme != "" || candidate.PowerlineGlyphs != nil || len(candidate.Keys) > 0 || candidate.Network != nil
		candidate.Theme = ""
		candidate.PowerlineGlyphs = nil
		candidate.Keys = map[string]map[string][]string{}
		candidate.Network = nil
		return changed, nil
	}
	parts := strings.Split(strings.TrimSpace(key), ".")
	switch {
	case key == "profile":
		return false, fmt.Errorf("profile is managed by `fbrcm profile switch <name>`")
	case key == "theme":
		changed := candidate.Theme != ""
		candidate.Theme = ""
		return changed, nil
	case key == "powerline_glyphs":
		changed := candidate.PowerlineGlyphs != nil
		candidate.PowerlineGlyphs = nil
		return changed, nil
	case key == "keys":
		changed := len(candidate.Keys) > 0
		candidate.Keys = map[string]map[string][]string{}
		return changed, nil
	case key == "network":
		changed := candidate.Network != nil
		candidate.Network = nil
		return changed, nil
	case len(parts) == 2 && parts[0] == "network" && parts[1] == "max_concurrent_requests":
		if candidate.Network == nil || candidate.Network.MaxConcurrentRequests == nil {
			return false, nil
		}
		candidate.Network.MaxConcurrentRequests = nil
		pruneNetworkConfig(candidate)
		return true, nil
	case len(parts) == 2 && parts[0] == "network" && parts[1] == "requests_per_minute":
		if candidate.Network == nil || candidate.Network.RequestsPerMinute == nil {
			return false, nil
		}
		candidate.Network.RequestsPerMinute = nil
		pruneNetworkConfig(candidate)
		return true, nil
	case len(parts) == 2 && parts[0] == "network" && parts[1] == "rate_limit_cooldown":
		if candidate.Network == nil || strings.TrimSpace(candidate.Network.RateLimitCooldown) == "" {
			return false, nil
		}
		candidate.Network.RateLimitCooldown = ""
		pruneNetworkConfig(candidate)
		return true, nil
	case len(parts) == 2 && parts[0] == "network" && parts[1] == "retry":
		if candidate.Network == nil || !retryConfigPresent(candidate.Network.Retry) {
			return false, nil
		}
		candidate.Network.Retry = nil
		pruneNetworkConfig(candidate)
		return true, nil
	case len(parts) == 3 && parts[0] == "network" && parts[1] == "retry":
		if candidate.Network == nil || candidate.Network.Retry == nil {
			return false, nil
		}
		retry := candidate.Network.Retry
		changed := false
		switch parts[2] {
		case "max_attempts":
			changed = retry.MaxAttempts != nil
			retry.MaxAttempts = nil
		case "base_delay":
			changed = strings.TrimSpace(retry.BaseDelay) != ""
			retry.BaseDelay = ""
		case "max_delay":
			changed = strings.TrimSpace(retry.MaxDelay) != ""
			retry.MaxDelay = ""
		case "jitter_percent":
			changed = retry.JitterPercent != nil
			retry.JitterPercent = nil
		default:
			return false, fmt.Errorf("unknown network retry key %q", parts[2])
		}
		pruneNetworkConfig(candidate)
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
	case key == "projects.aliases":
		changed := len(coreconfig.CloneProjectAliases(candidate)) > 0
		if candidate.Projects != nil {
			candidate.Projects.Aliases = nil
			candidate.Projects = nil
		}
		return changed, nil
	case len(parts) == 3 && parts[0] == "projects" && parts[1] == "aliases":
		if err := coreconfig.ValidateProjectAliasName(parts[2]); err != nil {
			return false, err
		}
		if candidate.Projects == nil || candidate.Projects.Aliases == nil {
			return false, nil
		}
		_, changed := candidate.Projects.Aliases[parts[2]]
		delete(candidate.Projects.Aliases, parts[2])
		if len(candidate.Projects.Aliases) == 0 {
			candidate.Projects = nil
		}
		return changed, nil
	default:
		return false, fmt.Errorf("unknown or non-resettable config key %q", key)
	}
}

func pruneNetworkConfig(cfg *coreconfig.AppConfig) {
	if cfg.Network == nil {
		return
	}
	if !retryConfigPresent(cfg.Network.Retry) {
		cfg.Network.Retry = nil
	}
	if cfg.Network.MaxConcurrentRequests == nil && cfg.Network.RequestsPerMinute == nil &&
		strings.TrimSpace(cfg.Network.RateLimitCooldown) == "" && cfg.Network.Retry == nil {
		cfg.Network = nil
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
	return &shared.ValidationError{Code: "configuration.invalid", Source: "configuration", Stage: "validation", Err: fmt.Errorf("config is invalid: %s", strings.Join(parts, "; "))}
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
