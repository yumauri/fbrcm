package config

import (
	"fmt"
	"os"
	"path/filepath"

	corelog "github.com/yumauri/fbrcm/core/log"
)

type AppConfig struct {
	Profile         string                         `toml:"profile" json:"profile"`
	PowerlineGlyphs *bool                          `toml:"powerline_glyphs,omitempty" json:"powerline_glyphs"`
	Keys            map[string]map[string][]string `toml:"keys" json:"keys"`
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

// LoadAppConfigStrict reads global config and rejects unknown TOML fields.
func LoadAppConfigStrict() (*AppConfig, error) {
	path := GetGlobalConfigFilePath()
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cfg, err := DecodeAppConfig(raw, true)
	if err != nil {
		return nil, fmt.Errorf("decode global config: %w", err)
	}
	return cfg, nil
}

// DecodeAppConfig decodes global TOML, optionally rejecting unknown fields.
func DecodeAppConfig(raw []byte, strict bool) (*AppConfig, error) {
	cfg := &AppConfig{}
	if err := decodeTOMLWithOptions(raw, cfg, strict); err != nil {
		return nil, err
	}
	return cfg, nil
}

// MarshalAppConfig encodes global config as TOML.
func MarshalAppConfig(cfg *AppConfig) ([]byte, error) {
	if cfg == nil {
		cfg = &AppConfig{}
	}
	return MarshalTOML(cfg)
}

// MarshalTOML encodes a configuration value as TOML.
func MarshalTOML(value any) ([]byte, error) {
	return encodeTOML(value)
}

func SaveAppConfig(cfg *AppConfig) error {
	if cfg == nil {
		cfg = &AppConfig{}
	}
	if err := EnsurePrivateDir(GetConfigRootDirPath()); err != nil {
		return fmt.Errorf("create config root: %w", err)
	}

	path := GetGlobalConfigFilePath()
	data, err := encodeTOML(cfg)
	if err != nil {
		return fmt.Errorf("encode global config: %w", err)
	}
	if err := WritePrivateFileAtomic(path, data); err != nil {
		return fmt.Errorf("write global config: %w", err)
	}
	return nil
}

// SaveAppConfigRaw atomically writes already validated global TOML.
func SaveAppConfigRaw(raw []byte) error {
	if err := EnsurePrivateDir(GetConfigRootDirPath()); err != nil {
		return fmt.Errorf("create config root: %w", err)
	}
	if err := WritePrivateFileAtomic(GetGlobalConfigFilePath(), raw); err != nil {
		return fmt.Errorf("write global config: %w", err)
	}
	return nil
}
