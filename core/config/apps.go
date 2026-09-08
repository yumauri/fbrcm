package config

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	corelog "github.com/yumauri/fbrcm/core/log"
)

const (
	AppsCacheFormatVersion = 1
	AppsCacheTTL           = time.Hour
)

type AppsCacheRecord struct {
	FormatVersion     int             `json:"format_version"`
	Kind              string          `json:"kind"`
	ProjectID         string          `json:"project_id"`
	AppID             string          `json:"app_id,omitempty"`
	CachedAt          time.Time       `json:"cached_at"`
	SuggestedFilename string          `json:"suggested_filename,omitempty"`
	MediaType         string          `json:"media_type,omitempty"`
	Payload           json.RawMessage `json:"payload"`
}

type AppsCacheEntry struct {
	Kind      string
	ProjectID string
	AppID     string
	CachedAt  time.Time
	Path      string
	Size      int64
}

func (r *AppsCacheRecord) IsFresh(now time.Time) bool {
	return r != nil && !r.CachedAt.IsZero() && now.Sub(r.CachedAt) < AppsCacheTTL
}

func GetAppsCacheDirPath() string { return filepath.Join(GetCacheDirPath(), "apps") }

func GetAppsIndexCachePath(projectID string) string {
	return filepath.Join(GetAppsCacheDirPath(), projectID, "index.json")
}

func GetAppDetailsCachePath(projectID, appID string) string {
	return filepath.Join(GetAppsCacheDirPath(), projectID, "details", appCacheKey(appID)+".json")
}

func GetAppConfigCachePath(projectID, appID string) string {
	return filepath.Join(GetAppsCacheDirPath(), projectID, "configs", appCacheKey(appID)+".json")
}

func GetAppConfigArtifactPath(projectID, appID string) string {
	return filepath.Join(GetAppsCacheDirPath(), projectID, "configs", appCacheKey(appID)+".bin")
}

func LoadAppsIndexCache(projectID string) (*AppsCacheRecord, error) {
	if err := validateAppsCacheProjectID(projectID); err != nil {
		return nil, err
	}
	path := GetAppsIndexCachePath(projectID)
	record, err := loadAppsCacheRecord(path, "index", projectID, "")
	if err == nil {
		corelog.For("config").Info("loaded applications cache", "project_id", projectID, "path", path, "cached_at", record.CachedAt)
	}
	return record, err
}

func SaveAppsIndexCache(projectID string, cachedAt time.Time, payload []byte) error {
	if err := validateAppsCacheProjectID(projectID); err != nil {
		return err
	}
	return saveAppsCacheRecord(GetAppsIndexCachePath(projectID), AppsCacheRecord{FormatVersion: AppsCacheFormatVersion, Kind: "index", ProjectID: projectID, CachedAt: cachedAt, Payload: payload})
}

func LoadAppDetailsCache(projectID, appID string) (*AppsCacheRecord, error) {
	if err := validateAppsCacheProjectID(projectID); err != nil {
		return nil, err
	}
	path := GetAppDetailsCachePath(projectID, appID)
	record, err := loadAppsCacheRecord(path, "details", projectID, appID)
	if err == nil {
		corelog.For("config").Info("loaded application details cache", "project_id", projectID, "app_id", appID, "path", path, "cached_at", record.CachedAt)
	}
	return record, err
}

func SaveAppDetailsCache(projectID, appID string, cachedAt time.Time, payload []byte) error {
	if err := validateAppsCacheProjectID(projectID); err != nil {
		return err
	}
	return saveAppsCacheRecord(GetAppDetailsCachePath(projectID, appID), AppsCacheRecord{FormatVersion: AppsCacheFormatVersion, Kind: "details", ProjectID: projectID, AppID: appID, CachedAt: cachedAt, Payload: payload})
}

func LoadAppConfigCache(projectID, appID string) (*AppsCacheRecord, []byte, error) {
	if err := validateAppsCacheProjectID(projectID); err != nil {
		return nil, nil, err
	}
	record, err := loadAppsCacheRecord(GetAppConfigCachePath(projectID, appID), "config", projectID, appID)
	if err != nil {
		return nil, nil, err
	}
	contents, err := os.ReadFile(GetAppConfigArtifactPath(projectID, appID))
	if err != nil {
		corelog.For("config").Error("read application configuration cache artifact failed", "project_id", projectID, "app_id", appID, "path", GetAppConfigArtifactPath(projectID, appID), "err", err)
		return nil, nil, fmt.Errorf("read app config cache artifact: %w", err)
	}
	corelog.For("config").Info("loaded application configuration cache", "project_id", projectID, "app_id", appID, "path", GetAppConfigCachePath(projectID, appID), "artifact_path", GetAppConfigArtifactPath(projectID, appID), "cached_at", record.CachedAt)
	return record, contents, nil
}

func SaveAppConfigCache(projectID, appID string, cachedAt time.Time, suggestedFilename, mediaType string, payload, contents []byte) error {
	if err := validateAppsCacheProjectID(projectID); err != nil {
		return err
	}
	metadataPath := GetAppConfigCachePath(projectID, appID)
	if err := EnsurePrivateDir(filepath.Dir(metadataPath)); err != nil {
		return fmt.Errorf("create app config cache dir: %w", err)
	}
	if err := WritePrivateFileAtomic(GetAppConfigArtifactPath(projectID, appID), contents); err != nil {
		return fmt.Errorf("write app config cache artifact: %w", err)
	}
	record := AppsCacheRecord{FormatVersion: AppsCacheFormatVersion, Kind: "config", ProjectID: projectID, AppID: appID, CachedAt: cachedAt, SuggestedFilename: suggestedFilename, MediaType: mediaType, Payload: payload}
	if err := saveAppsCacheRecord(metadataPath, record); err != nil {
		return err
	}
	return nil
}

func saveAppsCacheRecord(path string, record AppsCacheRecord) error {
	logger := corelog.For("config")
	logger.Debug("write "+appsCacheLabel(record.Kind)+" cache", appsCacheLogFields(record.ProjectID, record.AppID, path, "cached_at", record.CachedAt)...)
	if err := EnsurePrivateDir(filepath.Dir(path)); err != nil {
		logger.Error("create applications cache directory failed", appsCacheLogFields(record.ProjectID, record.AppID, filepath.Dir(path), "err", err)...)
		return fmt.Errorf("create apps cache dir: %w", err)
	}
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		logger.Error("encode "+appsCacheLabel(record.Kind)+" cache failed", appsCacheLogFields(record.ProjectID, record.AppID, path, "err", err)...)
		return fmt.Errorf("encode apps cache: %w", err)
	}
	if err := WritePrivateFileAtomic(path, append(data, '\n')); err != nil {
		logger.Error("write "+appsCacheLabel(record.Kind)+" cache failed", appsCacheLogFields(record.ProjectID, record.AppID, path, "err", err)...)
		return fmt.Errorf("write apps cache: %w", err)
	}
	logger.Info("saved "+appsCacheLabel(record.Kind)+" cache", appsCacheLogFields(record.ProjectID, record.AppID, path, "cached_at", record.CachedAt)...)
	return nil
}

func loadAppsCacheRecord(path, kind, projectID, appID string) (*AppsCacheRecord, error) {
	logger := corelog.For("config")
	logger.Debug("read "+appsCacheLabel(kind)+" cache", appsCacheLogFields(projectID, appID, path)...)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			logger.Warn(appsCacheLabel(kind)+" cache miss", appsCacheLogFields(projectID, appID, path)...)
		} else {
			logger.Error("read "+appsCacheLabel(kind)+" cache failed", appsCacheLogFields(projectID, appID, path, "err", err)...)
		}
		return nil, fmt.Errorf("read apps cache: %w", err)
	}
	var record AppsCacheRecord
	if err := json.Unmarshal(data, &record); err != nil {
		logger.Error("decode "+appsCacheLabel(kind)+" cache failed", appsCacheLogFields(projectID, appID, path, "err", err)...)
		return nil, fmt.Errorf("decode apps cache: %w", err)
	}
	if record.FormatVersion != AppsCacheFormatVersion || record.Kind != kind || record.ProjectID != projectID || record.AppID != appID || record.CachedAt.IsZero() {
		logger.Error("decode "+appsCacheLabel(kind)+" cache failed", appsCacheLogFields(projectID, appID, path, "err", "invalid cache metadata")...)
		return nil, fmt.Errorf("decode apps cache: invalid %s cache metadata", kind)
	}
	return &record, nil
}

func appsCacheLogFields(projectID, appID, path string, extra ...any) []any {
	fields := make([]any, 0, 6+len(extra))
	fields = append(fields, "project_id", projectID)
	if appID != "" {
		fields = append(fields, "app_id", appID)
	}
	fields = append(fields, "path", path)
	return append(fields, extra...)
}

func ListAppsCacheEntries() ([]AppsCacheEntry, error) {
	root := GetAppsCacheDirPath()
	var entries []AppsCacheEntry
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var record AppsCacheRecord
		if err := json.Unmarshal(data, &record); err != nil {
			return fmt.Errorf("decode apps cache %s: %w", path, err)
		}
		validAppID := (record.Kind == "index" && record.AppID == "") || (record.Kind != "index" && record.AppID != "")
		if record.FormatVersion != AppsCacheFormatVersion || (record.Kind != "index" && record.Kind != "details" && record.Kind != "config") || record.ProjectID == "" || record.CachedAt.IsZero() || !validAppID {
			return fmt.Errorf("decode apps cache %s: invalid cache metadata", path)
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		size := info.Size()
		if record.Kind == "config" {
			if artifactInfo, err := os.Stat(strings.TrimSuffix(path, ".json") + ".bin"); err == nil {
				size += artifactInfo.Size()
			}
		}
		entries = append(entries, AppsCacheEntry{Kind: record.Kind, ProjectID: record.ProjectID, AppID: record.AppID, CachedAt: record.CachedAt, Path: path, Size: size})
		return nil
	})
	if errors.Is(err, os.ErrNotExist) {
		return []AppsCacheEntry{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("list apps cache: %w", err)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })
	return entries, nil
}

func PruneAppsCacheForProject(projectID string, appIDs []string) error {
	if err := validateAppsCacheProjectID(projectID); err != nil {
		return err
	}
	valid := make(map[string]struct{}, len(appIDs))
	for _, appID := range appIDs {
		valid[appID] = struct{}{}
	}
	entries, err := ListAppsCacheEntries()
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.ProjectID != projectID || entry.AppID == "" {
			continue
		}
		if _, ok := valid[entry.AppID]; ok {
			continue
		}
		if err := os.Remove(entry.Path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("prune apps cache: %w", err)
		}
		if entry.Kind == "config" {
			_ = os.Remove(strings.TrimSuffix(entry.Path, ".json") + ".bin")
		}
	}
	return nil
}

func DeleteAppsCacheForProject(projectID string) error {
	if err := validateAppsCacheProjectID(projectID); err != nil {
		return err
	}
	return os.RemoveAll(filepath.Join(GetAppsCacheDirPath(), projectID))
}

func ClearAppsCache() error { return os.RemoveAll(GetAppsCacheDirPath()) }

func appCacheKey(appID string) string {
	sum := sha256.Sum256([]byte(appID))
	return hex.EncodeToString(sum[:])
}

func validateAppsCacheProjectID(projectID string) error {
	if err := ValidatePhysicalProjectID(projectID); err != nil {
		return fmt.Errorf("invalid apps cache project: %w", err)
	}
	return nil
}

func appsCacheLabel(kind string) string {
	switch kind {
	case "index":
		return "applications"
	case "details":
		return "application details"
	case "config":
		return "application configuration"
	default:
		return "applications"
	}
}
