package config

import (
	"fmt"
	"path/filepath"

	corelog "github.com/yumauri/fbrcm/core/log"
)

type AppConfig struct {
	Profile         string                         `toml:"profile"`
	PowerlineGlyphs *bool                          `toml:"powerline_glyphs,omitempty"`
	Keys            map[string]map[string][]string `toml:"keys"`
}

func GetGlobalConfigFilePath() string {
	return filepath.Join(GetConfigRootDirPath(), "config.toml")
}

func LoadAppConfig() (*AppConfig, error) {
	path := GetGlobalConfigFilePath()
	logger := corelog.For("config")
	logger.Debug("read global config", "path", path)

	cfg := &AppConfig{}
	if err := readTOMLFile(path, cfg); err != nil {
		logger.Debug("read global config failed", "path", path, "err", err)
		if isDecodeError(err) {
			return nil, fmt.Errorf("decode global config: %w", err)
		}
		return nil, err
	}
	logger.Debug("loaded global config", "path", path, "profile", cfg.Profile)
	return cfg, nil
}

func SaveAppConfig(cfg *AppConfig) error {
	if cfg == nil {
		cfg = &AppConfig{}
	}
	if err := EnsurePrivateDir(GetConfigRootDirPath()); err != nil {
		return fmt.Errorf("create config root: %w", err)
	}

	path := GetGlobalConfigFilePath()
	if err := writeTOMLFile(path, cfg); err != nil {
		if isEncodeError(err) {
			return fmt.Errorf("encode global config: %w", err)
		}
		return fmt.Errorf("write global config: %w", err)
	}
	return nil
}
