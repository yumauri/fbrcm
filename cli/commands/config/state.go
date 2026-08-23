package config

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"slices"
	"strings"

	"github.com/yumauri/fbrcm/cli/shared"
	coreconfig "github.com/yumauri/fbrcm/core/config"
	tuiconfig "github.com/yumauri/fbrcm/tui/config"
)

type configDiagnostic = tuiconfig.Diagnostic

type configState struct {
	Path         string
	Exists       bool
	Stored       *coreconfig.AppConfig
	GlobalPath   string
	GlobalExists bool
	Global       *coreconfig.AppConfig
	LocalPath    string
	LocalExists  bool
	Local        *coreconfig.AppConfig
	Merged       *coreconfig.AppConfig
	Effective    *coreconfig.AppConfig
	Migrated     *coreconfig.AppConfig
	Report       configValidationResult
}

type configValidationResult struct {
	Path     string             `json:"path"`
	Exists   bool               `json:"exists"`
	Valid    bool               `json:"valid"`
	Errors   []configDiagnostic `json:"errors"`
	Warnings []configDiagnostic `json:"warnings"`
}

func loadConfigState() (configState, error) {
	resolved, err := coreconfig.ResolveAppConfig()
	if err != nil {
		return configState{}, err
	}
	state := stateFromConfig("effective", resolved.Global.Exists || resolved.Local.Exists, resolved.Effective)
	state.GlobalPath = resolved.Global.Path
	state.GlobalExists = resolved.Global.Exists
	state.Global = cloneAppConfig(resolved.Global.Config)
	state.LocalPath = resolved.Local.Path
	state.LocalExists = resolved.Local.Exists
	state.Local = cloneAppConfig(resolved.Local.Config)
	state.Path = resolved.Global.Path
	state.Exists = resolved.Global.Exists
	state.Stored = state.Global
	state.Merged = cloneAppConfig(resolved.Effective)
	state.Report = validateAppConfig("effective configuration", state.GlobalExists || state.LocalExists, state.Merged)
	return state, nil
}

func stateFromConfig(path string, exists bool, stored *coreconfig.AppConfig) configState {
	stored = cloneAppConfig(stored)
	migrated := cloneAppConfig(stored)
	tuiconfig.MigrateAdminShortcuts(migrated.Keys)
	effective := cloneAppConfig(migrated)
	if strings.TrimSpace(effective.Profile) == "" {
		effective.Profile = coreconfig.DefaultProfileName
	}
	if effective.PowerlineGlyphs == nil {
		enabled := true
		effective.PowerlineGlyphs = &enabled
	}
	if effective.Network == nil {
		effective.Network = &coreconfig.NetworkConfig{}
	}
	if effective.Network.MaxConcurrentRequests == nil {
		value := coreconfig.DefaultMaxConcurrentRequests
		effective.Network.MaxConcurrentRequests = &value
	}
	if effective.Network.RequestsPerMinute == nil {
		value := coreconfig.DefaultRequestsPerMinute
		effective.Network.RequestsPerMinute = &value
	}
	if strings.TrimSpace(effective.Network.RateLimitCooldown) == "" {
		effective.Network.RateLimitCooldown = coreconfig.DefaultRateLimitCooldown.String()
	}
	if effective.Network.Retry == nil {
		effective.Network.Retry = &coreconfig.RetryConfig{}
	}
	if effective.Network.Retry.MaxAttempts == nil {
		value := coreconfig.DefaultRetryMaxAttempts
		effective.Network.Retry.MaxAttempts = &value
	}
	if strings.TrimSpace(effective.Network.Retry.BaseDelay) == "" {
		effective.Network.Retry.BaseDelay = coreconfig.DefaultRetryBaseDelay.String()
	}
	if strings.TrimSpace(effective.Network.Retry.MaxDelay) == "" {
		effective.Network.Retry.MaxDelay = coreconfig.DefaultRetryMaxDelay.String()
	}
	if effective.Network.Retry.JitterPercent == nil {
		value := coreconfig.DefaultRetryJitterPercent
		effective.Network.Retry.JitterPercent = &value
	}
	effective.Keys = tuiconfig.ToConfigMap(tuiconfig.Merge(tuiconfig.DefaultKeyMap(), effective.Keys))
	report := validateAppConfig(path, exists, stored)
	return configState{Path: path, Exists: exists, Stored: stored, GlobalPath: path, GlobalExists: exists, Global: stored, Local: &coreconfig.AppConfig{Keys: map[string]map[string][]string{}}, Merged: stored, Migrated: migrated, Effective: effective, Report: report}
}

func validateAppConfig(path string, exists bool, cfg *coreconfig.AppConfig) configValidationResult {
	report := configValidationResult{Path: path, Exists: exists, Valid: true, Errors: []configDiagnostic{}, Warnings: []configDiagnostic{}}
	if path == coreconfig.GetGlobalConfigFilePath() && len(coreconfig.CloneProjectAliases(cfg)) > 0 {
		report.Errors = append(report.Errors, configDiagnostic{
			Severity: "error",
			Code:     "repository_scope_required",
			Key:      "projects.aliases",
			Message:  "project aliases are repository-scoped; move them to .fbrcm.toml",
		})
	}
	if profile := strings.TrimSpace(cfg.Profile); profile != "" {
		if err := coreconfig.ValidateProfileName(cfg.Profile); err != nil {
			report.Errors = append(report.Errors, configDiagnostic{Severity: "error", Code: "invalid_profile", Key: "profile", Message: shared.SafeErrorText(err)})
		} else if !coreconfig.ProfileExists(profile) {
			report.Errors = append(report.Errors, configDiagnostic{Severity: "error", Code: "missing_profile", Key: "profile", Message: fmt.Sprintf("profile %q does not exist", profile)})
		}
	}
	for _, diagnostic := range tuiconfig.ValidateConfiguredKeys(cfg.Keys) {
		if diagnostic.Severity == "warning" {
			report.Warnings = append(report.Warnings, diagnostic)
		} else {
			report.Errors = append(report.Errors, diagnostic)
		}
	}
	sortDiagnostics(report.Errors)
	sortDiagnostics(report.Warnings)
	report.Valid = len(report.Errors) == 0
	return report
}

func decodeConfigForValidation(path string) (configState, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return stateFromConfig(path, false, &coreconfig.AppConfig{}), nil
		}
		return configState{}, err
	}
	cfg, err := coreconfig.DecodeAppConfig(raw, true)
	if err != nil {
		diagnostic := configDiagnostic{Severity: "error", Code: "toml_decode", Key: "", Message: shared.SafeErrorText(err)}
		return configState{Path: path, Exists: true, Report: configValidationResult{
			Path: path, Exists: true, Valid: false, Errors: []configDiagnostic{diagnostic}, Warnings: []configDiagnostic{},
		}}, nil
	}
	return stateFromConfig(path, true, cfg), nil
}

func cloneAppConfig(cfg *coreconfig.AppConfig) *coreconfig.AppConfig {
	out := coreconfig.CloneAppConfig(cfg)
	out.Keys = tuiconfig.CloneConfigMap(out.Keys)
	if out.Keys == nil {
		out.Keys = map[string]map[string][]string{}
	}
	return out
}

func configValue(state configState, key string) (any, string, error) {
	parts := strings.Split(strings.TrimSpace(key), ".")
	switch {
	case key == "profile":
		source := scalarSource(strings.TrimSpace(state.Local.Profile) != "", strings.TrimSpace(state.Global.Profile) != "")
		return state.Effective.Profile, source, nil
	case key == "powerline_glyphs":
		source := scalarSource(state.Local.PowerlineGlyphs != nil, state.Global.PowerlineGlyphs != nil)
		return *state.Effective.PowerlineGlyphs, source, nil
	case key == "keys":
		return state.Effective.Keys, keySource(state, parts), nil
	case key == "network":
		return state.Effective.Network, networkSource(state, ""), nil
	case key == "hooks":
		return state.Effective.Hooks, hookSource(state, ""), nil
	case key == "projects":
		return state.Effective.Projects, projectAliasSource(state, ""), nil
	case key == "projects.aliases":
		return coreconfig.CloneProjectAliases(state.Effective), projectAliasSource(state, ""), nil
	case len(parts) == 3 && parts[0] == "projects" && parts[1] == "aliases":
		if err := coreconfig.ValidateProjectAliasName(parts[2]); err != nil {
			return nil, "", shared.InvalidArgument(err)
		}
		value, ok := coreconfig.CloneProjectAliases(state.Effective)[parts[2]]
		if !ok {
			return nil, "default", nil
		}
		return value, projectAliasSource(state, parts[2]), nil
	case len(parts) == 2 && parts[0] == "hooks":
		if state.Effective.Hooks == nil {
			return nil, "default", nil
		}
		switch parts[1] {
		case "timeout":
			return state.Effective.Hooks.Timeout, hookSource(state, parts[1]), nil
		case "pre_publish":
			return append([]string(nil), state.Effective.Hooks.PrePublish...), hookSource(state, parts[1]), nil
		case "post_publish":
			return append([]string(nil), state.Effective.Hooks.PostPublish...), hookSource(state, parts[1]), nil
		default:
			return nil, "", shared.InvalidArgument(fmt.Errorf("unknown hook key %q", parts[1]))
		}
	case len(parts) == 2 && parts[0] == "network":
		switch parts[1] {
		case "max_concurrent_requests":
			return *state.Effective.Network.MaxConcurrentRequests, networkSource(state, parts[1]), nil
		case "requests_per_minute":
			return *state.Effective.Network.RequestsPerMinute, networkSource(state, parts[1]), nil
		case "rate_limit_cooldown":
			return state.Effective.Network.RateLimitCooldown, networkSource(state, parts[1]), nil
		case "retry":
			return state.Effective.Network.Retry, networkSource(state, parts[1]), nil
		default:
			return nil, "", shared.InvalidArgument(fmt.Errorf("unknown network key %q", parts[1]))
		}
	case len(parts) == 3 && parts[0] == "network" && parts[1] == "retry":
		retry := state.Effective.Network.Retry
		switch parts[2] {
		case "max_attempts":
			return *retry.MaxAttempts, networkSource(state, "retry."+parts[2]), nil
		case "base_delay":
			return retry.BaseDelay, networkSource(state, "retry."+parts[2]), nil
		case "max_delay":
			return retry.MaxDelay, networkSource(state, "retry."+parts[2]), nil
		case "jitter_percent":
			return *retry.JitterPercent, networkSource(state, "retry."+parts[2]), nil
		default:
			return nil, "", shared.InvalidArgument(fmt.Errorf("unknown network retry key %q", parts[2]))
		}
	case len(parts) == 2 && parts[0] == "keys":
		if !tuiconfig.KnownBlock(parts[1]) {
			return nil, "", shared.InvalidArgument(fmt.Errorf("unknown keybinding block %q", parts[1]))
		}
		return state.Effective.Keys[parts[1]], keySource(state, parts), nil
	case len(parts) == 3 && parts[0] == "keys":
		if !tuiconfig.KnownBlock(parts[1]) {
			return nil, "", shared.InvalidArgument(fmt.Errorf("unknown keybinding block %q", parts[1]))
		}
		if !tuiconfig.KnownAction(parts[1], parts[2]) {
			return nil, "", shared.InvalidArgument(fmt.Errorf("unknown action %q in block %q", parts[2], parts[1]))
		}
		return append([]string(nil), state.Effective.Keys[parts[1]][parts[2]]...), keySource(state, parts), nil
	default:
		return nil, "", shared.InvalidArgument(fmt.Errorf("unknown config key %q", key))
	}
}

func networkSource(state configState, key string) string {
	has := func(cfg *coreconfig.AppConfig) bool {
		if cfg == nil || cfg.Network == nil {
			return false
		}
		switch key {
		case "max_concurrent_requests":
			return cfg.Network.MaxConcurrentRequests != nil
		case "requests_per_minute":
			return cfg.Network.RequestsPerMinute != nil
		case "rate_limit_cooldown":
			return strings.TrimSpace(cfg.Network.RateLimitCooldown) != ""
		case "retry":
			return retryConfigPresent(cfg.Network.Retry)
		case "retry.max_attempts":
			return cfg.Network.Retry != nil && cfg.Network.Retry.MaxAttempts != nil
		case "retry.base_delay":
			return cfg.Network.Retry != nil && strings.TrimSpace(cfg.Network.Retry.BaseDelay) != ""
		case "retry.max_delay":
			return cfg.Network.Retry != nil && strings.TrimSpace(cfg.Network.Retry.MaxDelay) != ""
		case "retry.jitter_percent":
			return cfg.Network.Retry != nil && cfg.Network.Retry.JitterPercent != nil
		default:
			return cfg.Network.MaxConcurrentRequests != nil || cfg.Network.RequestsPerMinute != nil ||
				strings.TrimSpace(cfg.Network.RateLimitCooldown) != "" || retryConfigPresent(cfg.Network.Retry)
		}
	}
	local, global := has(state.Local), has(state.Global)
	if (key == "" || key == "retry") && local && global {
		return "mixed"
	}
	return scalarSource(local, global)
}

func retryConfigPresent(retry *coreconfig.RetryConfig) bool {
	return retry != nil && (retry.MaxAttempts != nil || strings.TrimSpace(retry.BaseDelay) != "" ||
		strings.TrimSpace(retry.MaxDelay) != "" || retry.JitterPercent != nil)
}

func projectAliasSource(state configState, alias string) string {
	aliases := coreconfig.CloneProjectAliases(state.Local)
	if alias == "" {
		if len(aliases) > 0 {
			return "local"
		}
		return "default"
	}
	if _, ok := aliases[alias]; ok {
		return "local"
	}
	return "default"
}

func hookSource(state configState, key string) string {
	has := func(cfg *coreconfig.AppConfig) bool {
		if cfg == nil || cfg.Hooks == nil {
			return false
		}
		switch key {
		case "timeout":
			return strings.TrimSpace(cfg.Hooks.Timeout) != ""
		case "pre_publish":
			return cfg.Hooks.PrePublish != nil
		case "post_publish":
			return cfg.Hooks.PostPublish != nil
		default:
			return true
		}
	}
	return scalarSource(has(state.Local), has(state.Global))
}

func keySource(state configState, parts []string) string {
	local := subtreeValue(state.Local.Keys, parts[1:])
	global := subtreeValue(state.Global.Keys, parts[1:])
	source := scalarSource(local != nil, global != nil)
	if len(parts) < 3 && local != nil && global != nil {
		source = "mixed"
	}
	configured := subtreeValue(state.Merged.Keys, parts[1:])
	migrated := subtreeValue(state.Migrated.Keys, parts[1:])
	if configured != nil && !reflect.DeepEqual(configured, migrated) {
		return "migrated"
	}
	return source
}

func scalarSource(local, global bool) string {
	if local {
		return "local"
	}
	if global {
		return "global"
	}
	return "default"
}

func subtreeValue(keys map[string]map[string][]string, parts []string) any {
	if len(parts) == 0 {
		if len(keys) == 0 {
			return nil
		}
		return keys
	}
	actions, ok := keys[parts[0]]
	if !ok {
		return nil
	}
	if len(parts) == 1 {
		return actions
	}
	value, ok := actions[parts[1]]
	if !ok {
		return nil
	}
	return value
}

func sortDiagnostics(diagnostics []configDiagnostic) {
	slices.SortFunc(diagnostics, func(left, right configDiagnostic) int {
		if left.Key != right.Key {
			return strings.Compare(left.Key, right.Key)
		}
		if left.Code != right.Code {
			return strings.Compare(left.Code, right.Code)
		}
		return strings.Compare(left.Message, right.Message)
	})
}
