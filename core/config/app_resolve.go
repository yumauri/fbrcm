package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/pelletier/go-toml/v2"

	"github.com/yumauri/fbrcm/core/env"
)

const LocalConfigFileName = ".fbrcm.toml"

type AppConfigLayer struct {
	Path   string
	Exists bool
	Config *AppConfig
	values map[string]any
}

type AppConfigResolution struct {
	Global    AppConfigLayer
	Local     AppConfigLayer
	Effective *AppConfig
}

var (
	localConfigMu       sync.RWMutex
	localConfigDisabled bool
)

// SetLocalConfigDisabled controls local config discovery for the current
// process. FBRCM_NO_LOCAL_CONFIG also disables discovery when non-empty.
func SetLocalConfigDisabled(disabled bool) {
	localConfigMu.Lock()
	localConfigDisabled = disabled
	localConfigMu.Unlock()
	resetPaths()
}

func LocalConfigDisabled() bool {
	localConfigMu.RLock()
	disabled := localConfigDisabled
	localConfigMu.RUnlock()
	if disabled {
		return true
	}
	_, disabled = env.LookupTrimmed(env.NoLocalConfig)
	return disabled
}

// FindLocalConfig searches startDir and each ancestor through the filesystem
// root. If no file exists, candidate is startDir/.fbrcm.toml.
func FindLocalConfig(startDir string) (candidate string, found bool, err error) {
	return findAncestorFile(startDir, LocalConfigFileName)
}

func findAncestorFile(startDir, name string) (candidate string, found bool, err error) {
	if strings.TrimSpace(startDir) == "" {
		return "", false, fmt.Errorf("file search start directory is empty")
	}
	startDir, err = filepath.Abs(startDir)
	if err != nil {
		return "", false, fmt.Errorf("resolve file search start directory: %w", err)
	}
	info, err := os.Stat(startDir)
	if err != nil {
		return "", false, fmt.Errorf("inspect file search start directory %s: %w", startDir, err)
	}
	if !info.IsDir() {
		return "", false, fmt.Errorf("file search start path is not a directory: %s", startDir)
	}
	first := filepath.Join(startDir, name)
	for dir := startDir; ; dir = filepath.Dir(dir) {
		path := filepath.Join(dir, name)
		info, statErr := os.Stat(path)
		switch {
		case statErr == nil:
			if !info.Mode().IsRegular() {
				return "", false, fmt.Errorf("repository file is not a regular file: %s", path)
			}
			return path, true, nil
		case errors.Is(statErr, os.ErrNotExist):
			// Continue to the parent.
		default:
			return "", false, fmt.Errorf("inspect repository file %s: %w", path, statErr)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
	}
	return first, false, nil
}

func GetLocalConfigFilePath() (string, bool, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", false, fmt.Errorf("resolve current working directory: %w", err)
	}
	return FindLocalConfig(cwd)
}

func ResolveAppConfig() (AppConfigResolution, error) {
	globalPath := GetGlobalConfigFilePath()
	global, err := loadAppConfigLayer(globalPath, "global")
	if err != nil {
		return AppConfigResolution{}, err
	}
	if err := RejectGlobalProjectAliases(global.Config); err != nil {
		return AppConfigResolution{}, invalidConfiguration(globalPath, "validation", fmt.Errorf("decode global config %s: %w", globalPath, err))
	}

	local := AppConfigLayer{Config: &AppConfig{}, values: map[string]any{}}
	if !LocalConfigDisabled() {
		localPath, found, findErr := GetLocalConfigFilePath()
		if findErr != nil {
			return AppConfigResolution{}, findErr
		}
		local.Path = localPath
		if found {
			local, err = loadAppConfigLayer(localPath, "local")
			if err != nil {
				return AppConfigResolution{}, err
			}
		}
	}

	merged := cloneTOMLMap(global.values)
	deepMergeTOML(merged, local.values)
	raw, err := toml.Marshal(merged)
	if err != nil {
		return AppConfigResolution{}, fmt.Errorf("encode merged config: %w", err)
	}
	effective, err := DecodeAppConfig(raw, true)
	if err != nil {
		return AppConfigResolution{}, invalidConfiguration("effective", "validation", fmt.Errorf("decode merged config: %w", err))
	}
	return AppConfigResolution{Global: global, Local: local, Effective: effective}, nil
}

// MergeAppConfigs deeply overlays local on global while preserving absent
// scalar fields. Built-in defaults are intentionally not applied here.
func MergeAppConfigs(global, local *AppConfig) (*AppConfig, error) {
	if err := RejectGlobalProjectAliases(global); err != nil {
		return nil, err
	}
	globalValues, err := appConfigValues(global)
	if err != nil {
		return nil, fmt.Errorf("encode global config values: %w", err)
	}
	localValues, err := appConfigValues(local)
	if err != nil {
		return nil, fmt.Errorf("encode local config values: %w", err)
	}
	deepMergeTOML(globalValues, localValues)
	raw, err := toml.Marshal(globalValues)
	if err != nil {
		return nil, fmt.Errorf("encode merged config: %w", err)
	}
	return DecodeAppConfig(raw, true)
}

func appConfigValues(cfg *AppConfig) (map[string]any, error) {
	if cfg == nil {
		cfg = &AppConfig{}
	}
	raw, err := MarshalAppConfig(cfg)
	if err != nil {
		return nil, err
	}
	values := map[string]any{}
	if err := toml.Unmarshal(raw, &values); err != nil {
		return nil, err
	}
	return values, nil
}

func loadAppConfigLayer(path, label string) (AppConfigLayer, error) {
	layer := AppConfigLayer{Path: path, Config: &AppConfig{}, values: map[string]any{}}
	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return layer, nil
		}
		return AppConfigLayer{}, fmt.Errorf("read %s config %s: %w", label, path, err)
	}
	cfg, err := DecodeAppConfig(raw, true)
	if err != nil {
		return AppConfigLayer{}, invalidConfiguration(path, "decoding", fmt.Errorf("decode %s config %s: %w", label, path, err))
	}
	values := map[string]any{}
	if err := toml.Unmarshal(raw, &values); err != nil {
		return AppConfigLayer{}, invalidConfiguration(path, "decoding", fmt.Errorf("decode %s config values %s: %w", label, path, err))
	}
	layer.Exists = true
	layer.Config = cfg
	layer.values = values
	return layer, nil
}

func cloneTOMLMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for key, value := range in {
		if nested, ok := value.(map[string]any); ok {
			out[key] = cloneTOMLMap(nested)
		} else {
			out[key] = value
		}
	}
	return out
}

func deepMergeTOML(base, overlay map[string]any) {
	for key, value := range overlay {
		overlayTable, overlayIsTable := value.(map[string]any)
		baseTable, baseIsTable := base[key].(map[string]any)
		if overlayIsTable && baseIsTable {
			deepMergeTOML(baseTable, overlayTable)
			continue
		}
		if overlayIsTable {
			base[key] = cloneTOMLMap(overlayTable)
			continue
		}
		base[key] = value
	}
}
