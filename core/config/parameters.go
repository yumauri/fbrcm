package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	corelog "github.com/yumauri/fbrcm/core/log"
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

type ParametersCacheSnapshot struct {
	ProjectID string
	Version   string
	Path      string
	Size      int64
	Cache     *ParametersCache
}

var createParametersCacheSymlink = os.Symlink

func GetParametersCacheDirPath() string {
	return filepath.Join(GetCacheDirPath(), "remote-config")
}

func GetParametersCachePath(projectID string) string {
	return filepath.Join(GetParametersCacheDirPath(), projectID+".json")
}

func GetParametersCacheVersionPath(projectID, version string) string {
	return filepath.Join(GetParametersCacheDirPath(), projectID+"."+version+".json")
}

func LoadParametersCache(projectID string) (*ParametersCache, error) {
	logger := corelog.For("config")
	path := GetParametersCachePath(projectID)
	logger.Debug("read current parameters cache", "project_id", projectID, "path", path)

	legacy, legacyErr := loadLegacyOrPointerParametersCache(projectID)
	if legacyErr == nil && legacy.pointer {
		return legacy.cache, nil
	}
	if legacyErr != nil && !errors.Is(legacyErr, os.ErrNotExist) {
		return nil, legacyErr
	}

	snapshot, latestErr := latestParametersCacheSnapshot(projectID)
	if latestErr == nil && snapshot != nil {
		logger.Info("loaded latest versioned parameters cache", "project_id", projectID, "path", snapshot.Path, "etag", snapshot.Cache.ETag, "version", snapshot.Version)
		return snapshot.Cache, nil
	}
	if latestErr != nil {
		return nil, latestErr
	}
	if legacyErr == nil && legacy.cache != nil {
		return legacy.cache, nil
	}

	logger.Warn("parameters cache miss", "project_id", projectID, "path", path)
	return nil, fmt.Errorf("read parameters cache: %w", os.ErrNotExist)
}

func LoadParametersCacheVersion(projectID, version string) (*ParametersCache, error) {
	path := GetParametersCacheVersionPath(projectID, version)
	logger := corelog.For("config")
	logger.Debug("read versioned parameters cache", "project_id", projectID, "version", version, "path", path)

	cache, err := readParametersCacheFile(projectID, path)
	if err != nil {
		return nil, err
	}
	logger.Info("loaded versioned parameters cache", "project_id", projectID, "path", path, "etag", cache.ETag, "version", parametersCacheVersion(cache.RemoteConfig))
	return cache, nil
}

func SaveParametersCache(projectID string, cache *ParametersCache) error {
	version := parametersCacheVersion(cache.RemoteConfig)
	if !isNumericVersion(version) {
		return saveLegacyParametersCache(projectID, cache)
	}

	path := GetParametersCacheVersionPath(projectID, version)
	logger := corelog.For("config")
	if err := EnsurePrivateDir(filepath.Dir(path)); err != nil {
		return fmt.Errorf("create parameters cache dir: %w", err)
	}

	logger.Debug("write versioned parameters cache", "project_id", projectID, "version", version, "path", path, "etag", cache.ETag)
	if err := writeJSONFile(path, cache); err != nil {
		if isEncodeError(err) {
			return fmt.Errorf("encode parameters cache: %w", err)
		}
		logger.Error("write parameters cache failed", "project_id", projectID, "path", path, "err", err)
		return fmt.Errorf("write parameters cache: %w", err)
	}
	if err := updateParametersCachePointer(projectID, path, cache); err != nil {
		return err
	}

	logger.Info("saved versioned parameters cache", "project_id", projectID, "path", path, "etag", cache.ETag, "version", version)
	return nil
}

func ListParametersCacheSnapshots() ([]ParametersCacheSnapshot, error) {
	dir := GetParametersCacheDirPath()
	files, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []ParametersCacheSnapshot{}, nil
		}
		return nil, fmt.Errorf("read cache dir: %w", err)
	}

	snapshots := make([]ParametersCacheSnapshot, 0, len(files))
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		projectID, version, ok := parseParametersCacheSnapshotName(file.Name())
		if !ok {
			continue
		}
		path := filepath.Join(dir, file.Name())
		info, err := file.Info()
		if err != nil {
			return nil, fmt.Errorf("stat cache file %s: %w", path, err)
		}
		cache, err := readParametersCacheFile(projectID, path)
		if err != nil {
			return nil, err
		}
		snapshots = append(snapshots, ParametersCacheSnapshot{
			ProjectID: projectID,
			Version:   version,
			Path:      path,
			Size:      info.Size(),
			Cache:     cache,
		})
	}
	return snapshots, nil
}

type loadedParametersCache struct {
	cache   *ParametersCache
	pointer bool
}

func loadLegacyOrPointerParametersCache(projectID string) (loadedParametersCache, error) {
	path := GetParametersCachePath(projectID)
	logger := corelog.For("config")
	info, err := os.Lstat(path)
	if err != nil {
		if isNotExist(err) {
			logger.Warn("parameters cache miss", "project_id", projectID, "path", path)
		} else {
			logger.Error("stat parameters cache failed", "project_id", projectID, "path", path, "err", err)
		}
		return loadedParametersCache{}, fmt.Errorf("read parameters cache: %w", err)
	}

	cache, err := readParametersCacheFile(projectID, path)
	if err != nil {
		return loadedParametersCache{}, err
	}
	logger.Info("loaded parameters cache", "project_id", projectID, "path", path, "etag", cache.ETag, "version", parametersCacheVersion(cache.RemoteConfig))
	return loadedParametersCache{
		cache:   cache,
		pointer: info.Mode()&os.ModeSymlink != 0,
	}, nil
}

func readParametersCacheFile(projectID, path string) (*ParametersCache, error) {
	logger := corelog.For("config")

	var cache ParametersCache
	if err := readJSONFile(path, &cache); err != nil {
		if isNotExist(err) {
			logger.Warn("parameters cache miss", "project_id", projectID, "path", path)
		} else if !isDecodeError(err) {
			logger.Error("read parameters cache failed", "project_id", projectID, "path", path, "err", err)
		} else {
			logger.Error("decode parameters cache failed", "project_id", projectID, "path", path, "err", err)
		}
		if isDecodeError(err) {
			return nil, fmt.Errorf("decode parameters cache: %w", err)
		}
		return nil, fmt.Errorf("read parameters cache: %w", err)
	}

	return &cache, nil
}

func saveLegacyParametersCache(projectID string, cache *ParametersCache) error {
	path := GetParametersCachePath(projectID)
	logger := corelog.For("config")
	if err := EnsurePrivateDir(filepath.Dir(path)); err != nil {
		return fmt.Errorf("create parameters cache dir: %w", err)
	}
	if info, err := os.Lstat(path); err == nil && info.Mode()&os.ModeSymlink != 0 {
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("remove parameters cache pointer: %w", err)
		}
	}

	logger.Debug("write legacy parameters cache", "project_id", projectID, "path", path, "etag", cache.ETag)
	if err := writeJSONFile(path, cache); err != nil {
		if isEncodeError(err) {
			return fmt.Errorf("encode parameters cache: %w", err)
		}
		logger.Error("write parameters cache failed", "project_id", projectID, "path", path, "err", err)
		return fmt.Errorf("write parameters cache: %w", err)
	}

	logger.Info("saved parameters cache", "project_id", projectID, "path", path, "etag", cache.ETag)
	return nil
}

func latestParametersCacheSnapshot(projectID string) (*ParametersCacheSnapshot, error) {
	snapshots, err := ListParametersCacheSnapshots()
	if err != nil {
		return nil, err
	}

	var latest *ParametersCacheSnapshot
	var latestNumber int64
	for i := range snapshots {
		snapshot := &snapshots[i]
		if snapshot.ProjectID != projectID {
			continue
		}
		number, ok := parseNumericVersion(snapshot.Version)
		if !ok {
			continue
		}
		if latest == nil || number > latestNumber {
			latest = snapshot
			latestNumber = number
		}
	}
	return latest, nil
}

func updateParametersCachePointer(projectID, targetPath string, cache *ParametersCache) error {
	pointerPath := GetParametersCachePath(projectID)
	logger := corelog.For("config")
	if err := os.Remove(pointerPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove parameters cache pointer: %w", err)
	}

	target := filepath.Base(targetPath)
	if err := createParametersCacheSymlink(target, pointerPath); err == nil {
		logger.Debug("updated parameters cache pointer", "project_id", projectID, "path", pointerPath, "target", target)
		return nil
	} else {
		logger.Warn("parameters cache symlink failed; writing pointer copy", "project_id", projectID, "path", pointerPath, "target", target, "err", err)
	}

	if err := writeJSONFile(pointerPath, cache); err != nil {
		if isEncodeError(err) {
			return fmt.Errorf("encode parameters cache pointer copy: %w", err)
		}
		return fmt.Errorf("write parameters cache pointer copy: %w", err)
	}
	return nil
}

func parseParametersCacheSnapshotName(name string) (projectID, version string, ok bool) {
	if filepath.Ext(name) != ".json" {
		return "", "", false
	}
	base := strings.TrimSuffix(name, ".json")
	dot := strings.LastIndex(base, ".")
	if dot <= 0 || dot == len(base)-1 {
		return "", "", false
	}
	projectID = base[:dot]
	version = base[dot+1:]
	if !isNumericVersion(version) {
		return "", "", false
	}
	return projectID, version, true
}

func isNumericVersion(version string) bool {
	_, ok := parseNumericVersion(version)
	return ok
}

func parseNumericVersion(version string) (int64, bool) {
	version = strings.TrimSpace(version)
	if version == "" {
		return 0, false
	}
	n, err := strconv.ParseInt(version, 10, 64)
	if err != nil || n <= 0 {
		return 0, false
	}
	return n, true
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
