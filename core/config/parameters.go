package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	corelog "fbrcm/core/log"
)

const ParametersCacheTTL = 10 * time.Minute

type ParametersCache struct {
	ETag         string          `json:"etag"`
	CachedAt     time.Time       `json:"cached_at"`
	RemoteConfig json.RawMessage `json:"remote_config"`
}

type parametersCacheVersionEnvelope struct {
	Version struct {
		VersionNumber string `json:"versionNumber"`
	} `json:"version"`
}

func GetParametersCacheDirPath() string {
	return filepath.Join(GetCacheDirPath(), "remote-config")
}

func GetParametersCachePath(projectID string) string {
	return filepath.Join(GetParametersCacheDirPath(), projectID+".json")
}

func LoadParametersCache(projectID string) (*ParametersCache, error) {
	path := GetParametersCachePath(projectID)
	logger := corelog.For("config")
	logger.Debug("read parameters cache", "project_id", projectID, "path", path)

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			logger.Warn("parameters cache miss", "project_id", projectID, "path", path)
			return nil, fmt.Errorf("read parameters cache: %w", err)
		}
		logger.Error("read parameters cache failed", "project_id", projectID, "path", path, "err", err)
		return nil, fmt.Errorf("read parameters cache: %w", err)
	}

	var cache ParametersCache
	if err := json.Unmarshal(data, &cache); err != nil {
		logger.Error("decode parameters cache failed", "project_id", projectID, "path", path, "err", err)
		return nil, fmt.Errorf("decode parameters cache: %w", err)
	}

	logger.Info("loaded parameters cache", "project_id", projectID, "path", path, "etag", cache.ETag, "version", parametersCacheVersion(cache.RemoteConfig))
	return &cache, nil
}

func SaveParametersCache(projectID string, cache *ParametersCache) error {
	path := GetParametersCachePath(projectID)
	logger := corelog.For("config")
	if err := EnsurePrivateDir(filepath.Dir(path)); err != nil {
		return fmt.Errorf("create parameters cache dir: %w", err)
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("encode parameters cache: %w", err)
	}
	data = append(data, '\n')

	logger.Debug("write parameters cache", "project_id", projectID, "path", path, "etag", cache.ETag)
	if err := os.WriteFile(path, data, PrivateFileMode); err != nil {
		logger.Error("write parameters cache failed", "project_id", projectID, "path", path, "err", err)
		return fmt.Errorf("write parameters cache: %w", err)
	}
	if err := EnsurePrivateFile(path); err != nil {
		return fmt.Errorf("chmod parameters cache: %w", err)
	}

	logger.Info("saved parameters cache", "project_id", projectID, "path", path, "etag", cache.ETag)
	return nil
}

func (c *ParametersCache) IsFresh(now time.Time) bool {
	if c == nil || c.CachedAt.IsZero() {
		return false
	}

	return now.Sub(c.CachedAt) < ParametersCacheTTL
}

func PurgeParametersCache() error {
	path := GetParametersCacheDirPath()
	logger := corelog.For("config")
	logger.Debug("remove parameters cache dir", "path", path)
	if err := os.RemoveAll(path); err != nil {
		logger.Error("remove parameters cache dir failed", "path", path, "err", err)
		return fmt.Errorf("remove parameters cache dir: %w", err)
	}

	logger.Info("parameters cache dir removed", "path", path)
	return nil
}

func parametersCacheVersion(raw json.RawMessage) string {
	if len(raw) == 0 {
		return "NA"
	}

	var envelope parametersCacheVersionEnvelope
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return ""
	}
	if envelope.Version.VersionNumber == "" {
		return "NA"
	}
	return envelope.Version.VersionNumber
}
