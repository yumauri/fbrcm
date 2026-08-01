package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	corelog "github.com/yumauri/fbrcm/core/log"
)

type AppConfig struct {
	Profile         string                         `toml:"profile,omitempty" json:"profile"`
	PowerlineGlyphs *bool                          `toml:"powerline_glyphs,omitempty" json:"powerline_glyphs"`
	Keys            map[string]map[string][]string `toml:"keys,omitempty" json:"keys"`
	Hooks           *HooksConfig                   `toml:"hooks,omitempty" json:"hooks,omitempty"`
	Projects        *ProjectsConfig                `toml:"projects,omitempty" json:"projects,omitempty"`
}

// ProjectsConfig contains repository-scoped project selection metadata.
type ProjectsConfig struct {
	Aliases map[string]string `toml:"aliases,omitempty" json:"aliases,omitempty"`
}

const DefaultHookTimeout = 5 * time.Minute

// HooksConfig configures commands around Remote Config publication.
type HooksConfig struct {
	Timeout     string   `toml:"timeout,omitempty" json:"timeout,omitempty"`
	PrePublish  []string `toml:"pre_publish,omitempty" json:"pre_publish,omitempty"`
	PostPublish []string `toml:"post_publish,omitempty" json:"post_publish,omitempty"`
}

// HookTimeout parses the configured per-command timeout.
func (c *HooksConfig) HookTimeout() (time.Duration, error) {
	if c == nil || strings.TrimSpace(c.Timeout) == "" {
		return DefaultHookTimeout, nil
	}
	timeout, err := time.ParseDuration(strings.TrimSpace(c.Timeout))
	if err != nil {
		return 0, fmt.Errorf("hooks.timeout must be a duration such as 30s or 2m: %w", err)
	}
	if timeout <= 0 {
		return 0, fmt.Errorf("hooks.timeout must be positive")
	}
	return timeout, nil
}

func validateHooksConfig(hooks *HooksConfig) error {
	if hooks == nil {
		return nil
	}
	if _, err := hooks.HookTimeout(); err != nil {
		return err
	}
	for name, commands := range map[string][]string{"pre_publish": hooks.PrePublish, "post_publish": hooks.PostPublish} {
		for i, command := range commands {
			if strings.TrimSpace(command) == "" {
				return fmt.Errorf("hooks.%s[%d] must not be empty", name, i)
			}
		}
	}
	return nil
}

func GetGlobalConfigFilePath() string {
	return filepath.Join(GetConfigRootDirPath(), "config.toml")
}

// LoadGlobalAppConfig reads only the user-wide configuration file.
func LoadGlobalAppConfig() (*AppConfig, error) {
	path := GetGlobalConfigFilePath()
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cfg, err := DecodeAppConfig(raw, true)
	if err != nil {
		return nil, fmt.Errorf("decode global config %s: %w", path, err)
	}
	if err := RejectGlobalProjectAliases(cfg); err != nil {
		return nil, fmt.Errorf("decode global config %s: %w", path, err)
	}
	return cfg, nil
}

// LoadAppConfig resolves the user-wide configuration and the nearest local
// .fbrcm.toml overlay. It returns os.ErrNotExist when neither file exists.
func LoadAppConfig() (*AppConfig, error) {
	logger := corelog.For("config")
	resolved, err := ResolveAppConfig()
	if err != nil {
		return nil, err
	}
	if !resolved.Global.Exists && !resolved.Local.Exists {
		return nil, os.ErrNotExist
	}
	logger.Debug("loaded effective config", "global_path", resolved.Global.Path, "local_path", resolved.Local.Path, "local_exists", resolved.Local.Exists, "profile", resolved.Effective.Profile)
	return resolved.Effective, nil
}

// LoadAppConfigStrict reads global config and rejects unknown TOML fields.
func LoadAppConfigStrict() (*AppConfig, error) {
	return LoadGlobalAppConfig()
}

// DecodeAppConfig decodes global TOML, optionally rejecting unknown fields.
func DecodeAppConfig(raw []byte, strict bool) (*AppConfig, error) {
	cfg := &AppConfig{}
	if err := decodeTOMLWithOptions(raw, cfg, strict); err != nil {
		return nil, err
	}
	if err := validateHooksConfig(cfg.Hooks); err != nil {
		return nil, err
	}
	if err := ValidateProjectAliases(projectAliases(cfg)); err != nil {
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
	if err := RejectGlobalProjectAliases(cfg); err != nil {
		return err
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

// SaveLocalAppConfigRaw atomically writes already validated local TOML. New
// repository configuration files use ordinary shared-file permissions, while
// existing permissions are preserved.
func SaveLocalAppConfigRaw(path string, raw []byte) error {
	mode := os.FileMode(0o644)
	if info, err := os.Stat(path); err == nil {
		mode = info.Mode().Perm()
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat local config %s: %w", path, err)
	}
	if err := writeFileAtomicMode(path, raw, mode); err != nil {
		return fmt.Errorf("write local config %s: %w", path, err)
	}
	return nil
}
