package config

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"slices"
	"strings"

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
			report.Errors = append(report.Errors, configDiagnostic{Severity: "error", Code: "invalid_profile", Key: "profile", Message: err.Error()})
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
		diagnostic := configDiagnostic{Severity: "error", Code: "toml_decode", Key: "", Message: err.Error()}
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
	case key == "hooks":
		return state.Effective.Hooks, hookSource(state, ""), nil
	case key == "projects":
		return state.Effective.Projects, projectAliasSource(state, ""), nil
	case key == "projects.aliases":
		return coreconfig.CloneProjectAliases(state.Effective), projectAliasSource(state, ""), nil
	case len(parts) == 3 && parts[0] == "projects" && parts[1] == "aliases":
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
			return nil, "", fmt.Errorf("unknown hook key %q", parts[1])
		}
	case len(parts) == 2 && parts[0] == "keys":
		if !tuiconfig.KnownBlock(parts[1]) {
			return nil, "", fmt.Errorf("unknown keybinding block %q", parts[1])
		}
		return state.Effective.Keys[parts[1]], keySource(state, parts), nil
	case len(parts) == 3 && parts[0] == "keys":
		if !tuiconfig.KnownBlock(parts[1]) {
			return nil, "", fmt.Errorf("unknown keybinding block %q", parts[1])
		}
		if !tuiconfig.KnownAction(parts[1], parts[2]) {
			return nil, "", fmt.Errorf("unknown action %q in block %q", parts[2], parts[1])
		}
		return append([]string(nil), state.Effective.Keys[parts[1]][parts[2]]...), keySource(state, parts), nil
	default:
		return nil, "", fmt.Errorf("unknown config key %q", key)
	}
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
