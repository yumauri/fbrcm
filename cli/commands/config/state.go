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
	Path      string
	Exists    bool
	Stored    *coreconfig.AppConfig
	Effective *coreconfig.AppConfig
	Migrated  *coreconfig.AppConfig
	Report    configValidationResult
}

type configValidationResult struct {
	Path     string             `json:"path"`
	Exists   bool               `json:"exists"`
	Valid    bool               `json:"valid"`
	Errors   []configDiagnostic `json:"errors"`
	Warnings []configDiagnostic `json:"warnings"`
}

func loadConfigState() (configState, error) {
	path := coreconfig.GetGlobalConfigFilePath()
	raw, err := os.ReadFile(path)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return configState{}, fmt.Errorf("read global config: %w", err)
		}
		return stateFromConfig(path, false, &coreconfig.AppConfig{}), nil
	}
	cfg, err := coreconfig.DecodeAppConfig(raw, true)
	if err != nil {
		return configState{}, fmt.Errorf("decode global config: %w", err)
	}
	return stateFromConfig(path, true, cfg), nil
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
	return configState{Path: path, Exists: exists, Stored: stored, Migrated: migrated, Effective: effective, Report: report}
}

func validateAppConfig(path string, exists bool, cfg *coreconfig.AppConfig) configValidationResult {
	report := configValidationResult{Path: path, Exists: exists, Valid: true, Errors: []configDiagnostic{}, Warnings: []configDiagnostic{}}
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
	if cfg == nil {
		cfg = &coreconfig.AppConfig{}
	}
	out := &coreconfig.AppConfig{Profile: cfg.Profile, Keys: tuiconfig.CloneConfigMap(cfg.Keys)}
	if cfg.PowerlineGlyphs != nil {
		value := *cfg.PowerlineGlyphs
		out.PowerlineGlyphs = &value
	}
	if out.Keys == nil {
		out.Keys = map[string]map[string][]string{}
	}
	return out
}

func configValue(state configState, key string) (any, string, error) {
	parts := strings.Split(strings.TrimSpace(key), ".")
	switch {
	case key == "profile":
		source := "configured"
		if strings.TrimSpace(state.Stored.Profile) == "" {
			source = "default"
		}
		return state.Effective.Profile, source, nil
	case key == "powerline_glyphs":
		source := "configured"
		if state.Stored.PowerlineGlyphs == nil {
			source = "default"
		}
		return *state.Effective.PowerlineGlyphs, source, nil
	case key == "keys":
		return state.Effective.Keys, keySource(state, parts), nil
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

func keySource(state configState, parts []string) string {
	configured := subtreeValue(state.Stored.Keys, parts[1:])
	if configured == nil {
		return "default"
	}
	migrated := subtreeValue(state.Migrated.Keys, parts[1:])
	if !reflect.DeepEqual(configured, migrated) {
		return "migrated"
	}
	return "configured"
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
