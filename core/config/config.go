package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"

	corelog "github.com/yumauri/fbrcm/core/log"
)

// AppConfig holds app config state used by the config package.
type AppConfig struct {
	// Profile stores profile for AppConfig.
	Profile string `toml:"profile"`
}

// GetGlobalConfigFilePath gets global config file path and returns the resulting value or error.
func GetGlobalConfigFilePath() string {
	return filepath.Join(GetConfigRootDirPath(), "config.toml")
}

// LoadAppConfig loads app config and returns the resulting value or error.
func LoadAppConfig() (*AppConfig, error) {
	path := GetGlobalConfigFilePath()
	logger := corelog.For("config")
	logger.Debug("read global config", "path", path)

	data, err := os.ReadFile(path)
	if err != nil {
		logger.Debug("read global config failed", "path", path, "err", err)
		return nil, err
	}

	cfg := &AppConfig{}
	if err := toml.Unmarshal(data, cfg); err != nil {
		logger.Debug("decode global config failed", "path", path, "err", err)
		return nil, fmt.Errorf("decode global config: %w", err)
	}
	logger.Debug("loaded global config", "path", path, "profile", cfg.Profile)
	return cfg, nil
}

// SaveAppConfig saves app config and returns the resulting value or error.
func SaveAppConfig(cfg *AppConfig) error {
	if cfg == nil {
		cfg = &AppConfig{}
	}
	if err := EnsurePrivateDir(GetConfigRootDirPath()); err != nil {
		return fmt.Errorf("create config root: %w", err)
	}

	data, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("encode global config: %w", err)
	}
	path := GetGlobalConfigFilePath()
	if err := os.WriteFile(path, data, PrivateFileMode); err != nil {
		return fmt.Errorf("write global config: %w", err)
	}
	if err := EnsurePrivateFile(path); err != nil {
		return fmt.Errorf("chmod global config: %w", err)
	}
	return nil
}
