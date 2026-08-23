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
	Network         *NetworkConfig                 `toml:"network,omitempty" json:"network,omitempty"`
	Hooks           *HooksConfig                   `toml:"hooks,omitempty" json:"hooks,omitempty"`
	Projects        *ProjectsConfig                `toml:"projects,omitempty" json:"projects,omitempty"`
}

const (
	DefaultMaxConcurrentRequests = 5
	MaxConcurrentRequests        = 64
	DefaultRequestsPerMinute     = 0
	MaxRequestsPerMinute         = 60_000
	DefaultRateLimitCooldown     = 30 * time.Second
	DefaultRetryMaxAttempts      = 5
	MaxRetryAttempts             = 10
	DefaultRetryBaseDelay        = time.Second
	DefaultRetryMaxDelay         = 10 * time.Second
	DefaultRetryJitterPercent    = 50
)

// NetworkConfig controls pacing and recovery for outbound API requests.
// RequestsPerMinute is a pointer so a local zero can explicitly disable a
// nonzero global limit in the deeply overlaid repository configuration.
type NetworkConfig struct {
	MaxConcurrentRequests *int         `toml:"max_concurrent_requests,omitempty" json:"max_concurrent_requests,omitempty"`
	RequestsPerMinute     *int         `toml:"requests_per_minute,omitempty" json:"requests_per_minute,omitempty"`
	RateLimitCooldown     string       `toml:"rate_limit_cooldown,omitempty" json:"rate_limit_cooldown,omitempty"`
	Retry                 *RetryConfig `toml:"retry,omitempty" json:"retry,omitempty"`
}

// RetryConfig controls retries for replayable requests after transient
// failures. MaxAttempts includes the initial request.
type RetryConfig struct {
	MaxAttempts   *int   `toml:"max_attempts,omitempty" json:"max_attempts,omitempty"`
	BaseDelay     string `toml:"base_delay,omitempty" json:"base_delay,omitempty"`
	MaxDelay      string `toml:"max_delay,omitempty" json:"max_delay,omitempty"`
	JitterPercent *int   `toml:"jitter_percent,omitempty" json:"jitter_percent,omitempty"`
}

func (c *NetworkConfig) EffectiveMaxConcurrentRequests() int {
	if c == nil || c.MaxConcurrentRequests == nil {
		return DefaultMaxConcurrentRequests
	}
	return *c.MaxConcurrentRequests
}

// EffectiveRequestsPerMinute returns the configured request rate. Zero leaves
// proactive pacing disabled while 429 cooldown coordination remains active.
func (c *NetworkConfig) EffectiveRequestsPerMinute() int {
	if c == nil || c.RequestsPerMinute == nil {
		return DefaultRequestsPerMinute
	}
	return *c.RequestsPerMinute
}

// EffectiveRateLimitCooldown returns the fallback cooldown used when a 429
// response does not include Retry-After.
func (c *NetworkConfig) EffectiveRateLimitCooldown() (time.Duration, error) {
	if c == nil || strings.TrimSpace(c.RateLimitCooldown) == "" {
		return DefaultRateLimitCooldown, nil
	}
	delay, err := time.ParseDuration(strings.TrimSpace(c.RateLimitCooldown))
	if err != nil {
		return 0, fmt.Errorf("network.rate_limit_cooldown must be a duration such as 30s or 1m: %w", err)
	}
	if delay <= 0 {
		return 0, fmt.Errorf("network.rate_limit_cooldown must be positive")
	}
	return delay, nil
}

func (c *RetryConfig) EffectiveMaxAttempts() int {
	if c == nil || c.MaxAttempts == nil {
		return DefaultRetryMaxAttempts
	}
	return *c.MaxAttempts
}

func (c *RetryConfig) EffectiveBaseDelay() (time.Duration, error) {
	if c == nil || strings.TrimSpace(c.BaseDelay) == "" {
		return DefaultRetryBaseDelay, nil
	}
	return parsePositiveNetworkDuration("network.retry.base_delay", c.BaseDelay)
}

func (c *RetryConfig) EffectiveMaxDelay() (time.Duration, error) {
	if c == nil || strings.TrimSpace(c.MaxDelay) == "" {
		return DefaultRetryMaxDelay, nil
	}
	return parsePositiveNetworkDuration("network.retry.max_delay", c.MaxDelay)
}

func (c *RetryConfig) EffectiveJitterPercent() int {
	if c == nil || c.JitterPercent == nil {
		return DefaultRetryJitterPercent
	}
	return *c.JitterPercent
}

func parsePositiveNetworkDuration(key, raw string) (time.Duration, error) {
	delay, err := time.ParseDuration(strings.TrimSpace(raw))
	if err != nil {
		return 0, fmt.Errorf("%s must be a duration such as 500ms or 10s: %w", key, err)
	}
	if delay <= 0 {
		return 0, fmt.Errorf("%s must be positive", key)
	}
	return delay, nil
}

func validateNetworkConfig(network *NetworkConfig) error {
	if network == nil {
		return nil
	}
	maxConcurrentRequests := network.EffectiveMaxConcurrentRequests()
	if maxConcurrentRequests < 1 || maxConcurrentRequests > MaxConcurrentRequests {
		return fmt.Errorf("network.max_concurrent_requests must be between 1 and %d", MaxConcurrentRequests)
	}
	requestsPerMinute := network.EffectiveRequestsPerMinute()
	if requestsPerMinute < 0 || requestsPerMinute > MaxRequestsPerMinute {
		return fmt.Errorf("network.requests_per_minute must be between 0 and %d", MaxRequestsPerMinute)
	}
	if _, err := network.EffectiveRateLimitCooldown(); err != nil {
		return err
	}
	retry := network.Retry
	maxAttempts := retry.EffectiveMaxAttempts()
	if maxAttempts < 1 || maxAttempts > MaxRetryAttempts {
		return fmt.Errorf("network.retry.max_attempts must be between 1 and %d", MaxRetryAttempts)
	}
	baseDelay, err := retry.EffectiveBaseDelay()
	if err != nil {
		return err
	}
	maxDelay, err := retry.EffectiveMaxDelay()
	if err != nil {
		return err
	}
	if maxDelay < baseDelay {
		return fmt.Errorf("network.retry.max_delay must be greater than or equal to network.retry.base_delay")
	}
	jitterPercent := retry.EffectiveJitterPercent()
	if jitterPercent < 0 || jitterPercent > 100 {
		return fmt.Errorf("network.retry.jitter_percent must be between 0 and 100")
	}
	return nil
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
		return nil, invalidConfiguration(path, "decoding", fmt.Errorf("decode global config %s: %w", path, err))
	}
	if err := RejectGlobalProjectAliases(cfg); err != nil {
		return nil, invalidConfiguration(path, "validation", fmt.Errorf("decode global config %s: %w", path, err))
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
	if err := validateNetworkConfig(cfg.Network); err != nil {
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
	clearSessionAppConfigResolution()
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
	clearSessionAppConfigResolution()
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
	clearSessionAppConfigResolution()
	return nil
}
