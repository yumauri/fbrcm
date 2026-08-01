package hooks

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
)

const TrustEnvironment = "FBRCM_HOOK_TRUST"

type Event string

const (
	PrePublish  Event = "pre_publish"
	PostPublish Event = "post_publish"
)

// Resolution describes the effective hook configuration and its provenance.
type Resolution struct {
	Hooks            config.HooksConfig `json:"hooks"`
	Timeout          time.Duration      `json:"-"`
	GlobalPath       string             `json:"global_path"`
	LocalPath        string             `json:"local_path,omitempty"`
	LocalExists      bool               `json:"local_exists"`
	LocalHooks       bool               `json:"local_hooks"`
	Fingerprint      string             `json:"fingerprint,omitempty"`
	Trusted          bool               `json:"trusted"`
	TrustEnvironment bool               `json:"trust_environment"`
	prePublishLocal  bool
	postPublishLocal bool
}

func Resolve() (Resolution, error) {
	resolved, err := config.ResolveAppConfig()
	if err != nil {
		return Resolution{}, err
	}
	result := Resolution{GlobalPath: resolved.Global.Path, LocalPath: resolved.Local.Path, LocalExists: resolved.Local.Exists}
	if resolved.Effective.Hooks != nil {
		result.Hooks = *resolved.Effective.Hooks
		result.Hooks.PrePublish = append([]string(nil), resolved.Effective.Hooks.PrePublish...)
		result.Hooks.PostPublish = append([]string(nil), resolved.Effective.Hooks.PostPublish...)
	}
	result.Timeout, err = result.Hooks.HookTimeout()
	if err != nil {
		return Resolution{}, err
	}
	result.LocalHooks = resolved.Local.Exists && resolved.Local.Config.Hooks != nil &&
		(resolved.Local.Config.Hooks.PrePublish != nil || resolved.Local.Config.Hooks.PostPublish != nil)
	if resolved.Local.Exists && resolved.Local.Config.Hooks != nil {
		result.prePublishLocal = resolved.Local.Config.Hooks.PrePublish != nil
		result.postPublishLocal = resolved.Local.Config.Hooks.PostPublish != nil
	}
	if !result.LocalHooks {
		result.Trusted = true
		return result, nil
	}
	result.Fingerprint, err = fingerprint(result.LocalPath, result.Hooks)
	if err != nil {
		return Resolution{}, err
	}
	if expected, ok := env.LookupTrimmed(TrustEnvironment); ok {
		result.TrustEnvironment = true
		result.Trusted = strings.EqualFold(expected, result.Fingerprint)
		return result, nil
	}
	store, err := loadTrustStore()
	if err != nil {
		return Resolution{}, err
	}
	canonical, err := canonicalPath(result.LocalPath)
	if err != nil {
		return Resolution{}, err
	}
	result.Trusted = store.Trusted[canonical] == result.Fingerprint
	return result, nil
}

func (r Resolution) Commands(event Event) []string {
	switch event {
	case PrePublish:
		return append([]string(nil), r.Hooks.PrePublish...)
	case PostPublish:
		return append([]string(nil), r.Hooks.PostPublish...)
	default:
		return nil
	}
}

func (r Resolution) Source(event Event) (path string, local bool) {
	switch event {
	case PrePublish:
		local = r.prePublishLocal
	case PostPublish:
		local = r.postPublishLocal
	}
	if local {
		return r.LocalPath, true
	}
	return r.GlobalPath, false
}

func (r Resolution) RequireTrust(event Event) error {
	if len(r.Commands(event)) == 0 {
		return nil
	}
	_, local := r.Source(event)
	if !local || r.Trusted {
		return nil
	}
	if r.TrustEnvironment {
		return fmt.Errorf("local hooks in %s do not match %s; expected fingerprint %s", r.LocalPath, TrustEnvironment, r.Fingerprint)
	}
	return fmt.Errorf("local hooks in %s are not trusted; review them and run `fbrcm hooks trust` (fingerprint %s)", r.LocalPath, r.Fingerprint)
}

type trustStore struct {
	Trusted map[string]string `json:"trusted"`
}

func trustStorePath() string { return filepath.Join(config.GetConfigRootDirPath(), "hook-trust.json") }

func loadTrustStore() (trustStore, error) {
	store := trustStore{Trusted: map[string]string{}}
	raw, err := os.ReadFile(trustStorePath())
	if errors.Is(err, os.ErrNotExist) {
		return store, nil
	}
	if err != nil {
		return trustStore{}, fmt.Errorf("read hook trust store: %w", err)
	}
	if err := json.Unmarshal(raw, &store); err != nil {
		return trustStore{}, fmt.Errorf("decode hook trust store: %w", err)
	}
	if store.Trusted == nil {
		store.Trusted = map[string]string{}
	}
	return store, nil
}

func saveTrustStore(store trustStore) error {
	if err := config.EnsurePrivateDir(config.GetConfigRootDirPath()); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	return config.WritePrivateFileAtomic(trustStorePath(), raw)
}

func TrustCurrent() (Resolution, error) {
	resolution, err := Resolve()
	if err != nil {
		return Resolution{}, err
	}
	if !resolution.LocalHooks {
		return Resolution{}, fmt.Errorf("the effective configuration does not define local hooks")
	}
	canonical, err := canonicalPath(resolution.LocalPath)
	if err != nil {
		return Resolution{}, err
	}
	store, err := loadTrustStore()
	if err != nil {
		return Resolution{}, err
	}
	store.Trusted[canonical] = resolution.Fingerprint
	if err := saveTrustStore(store); err != nil {
		return Resolution{}, err
	}
	resolution.Trusted = true
	return resolution, nil
}

func UntrustCurrent() (Resolution, bool, error) {
	resolution, err := Resolve()
	if err != nil {
		return Resolution{}, false, err
	}
	if !resolution.LocalExists {
		return resolution, false, nil
	}
	canonical, err := canonicalPath(resolution.LocalPath)
	if err != nil {
		return Resolution{}, false, err
	}
	store, err := loadTrustStore()
	if err != nil {
		return Resolution{}, false, err
	}
	_, changed := store.Trusted[canonical]
	delete(store.Trusted, canonical)
	if changed {
		if err := saveTrustStore(store); err != nil {
			return Resolution{}, false, err
		}
	}
	resolution.Trusted = !resolution.LocalHooks
	return resolution, changed, nil
}

func fingerprint(path string, hooks config.HooksConfig) (string, error) {
	canonical, err := canonicalPath(path)
	if err != nil {
		return "", err
	}
	raw, err := json.Marshal(struct {
		Path  string             `json:"path"`
		Hooks config.HooksConfig `json:"hooks"`
	}{canonical, hooks})
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(raw)
	return hex.EncodeToString(digest[:]), nil
}

func canonicalPath(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	canonical, err := filepath.EvalSymlinks(absolute)
	if err == nil {
		return canonical, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return filepath.Clean(absolute), nil
	}
	return "", err
}
