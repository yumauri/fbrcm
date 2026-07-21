package core

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/yumauri/fbrcm/core/conditions"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/firebase"
	corelog "github.com/yumauri/fbrcm/core/log"
	"github.com/yumauri/fbrcm/core/parameters"
)

func (s *Core) InspectParametersCache(projectID string) (*ParametersCache, ParametersCacheState, error) {
	logger := corelog.For("core")
	logger.Debug("inspect parameters cache", "project_id", projectID)

	cache, err := config.LoadParametersCache(projectID)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			logger.Warn("parameters cache miss", "project_id", projectID)
			return nil, ParametersCacheMissing, nil
		}
		logger.Error("parameters cache load failed", "project_id", projectID, "err", err)
		return nil, ParametersCacheMissing, err
	}

	if _, err := firebase.ParseRemoteConfig(cache.RemoteConfig); err != nil {
		logger.Error("cached remote config decode failed", "project_id", projectID, "err", err)
		return nil, ParametersCacheMissing, fmt.Errorf("decode cached remote config: %w", err)
	}

	if cache.IsFresh(time.Now()) {
		logger.Info("parameters cache fresh", "project_id", projectID, "etag", cache.ETag)
		return cache, ParametersCacheFresh, nil
	}

	logger.Info("parameters cache stale", "project_id", projectID, "etag", cache.ETag)
	return cache, ParametersCacheStale, nil
}

// LoadCachedRemoteConfig loads a project's locally cached Remote Config.
// A missing cache returns a nil config without an error.
func (s *Core) LoadCachedRemoteConfig(projectID string) (*firebase.RemoteConfig, error) {
	cache, state, err := s.InspectParametersCache(projectID)
	if err != nil {
		return nil, err
	}
	if state == ParametersCacheMissing || cache == nil {
		return nil, nil
	}
	return firebase.ParseRemoteConfig(cache.RemoteConfig)
}

func (s *Core) GetParameters(ctx context.Context, projectID string, force bool) (*ParametersCache, string, error) {
	logger := corelog.For("core")
	logger.Debug("get parameters requested", "project_id", projectID, "force", force)

	if force {
		return s.fetchLatestParameters(ctx, projectID)
	}

	cache, state, err := s.InspectParametersCache(projectID)
	if err != nil {
		return nil, "", fmt.Errorf("inspect parameters cache: %w", err)
	}

	switch state {
	case ParametersCacheFresh:
		logger.Info("serving parameters from fresh cache", "project_id", projectID)
		return cache, "cache", nil
	case ParametersCacheStale:
		logger.Info("verifying stale parameters cache", "project_id", projectID)
		return s.verifyParameters(ctx, projectID, cache)
	default:
		logger.Warn("fetching parameters from firebase due to cache miss", "project_id", projectID)
		return s.fetchLatestParameters(ctx, projectID)
	}
}

func (s *Core) RevalidateParameters(ctx context.Context, projectID string) (*ParametersCache, string, error) {
	logger := corelog.For("core")
	logger.Debug("revalidate parameters requested", "project_id", projectID)

	cache, state, err := s.InspectParametersCache(projectID)
	if err != nil {
		return nil, "", fmt.Errorf("inspect parameters cache: %w", err)
	}

	if state == ParametersCacheMissing || cache == nil {
		logger.Warn("fetching parameters from firebase due to cache miss during revalidation", "project_id", projectID)
		return s.fetchLatestParameters(ctx, projectID)
	}

	return s.verifyParameters(ctx, projectID, cache)
}

func (s *Core) BuildParametersTree(cache *ParametersCache) (*ParametersTree, error) {
	if cache == nil {
		return nil, fmt.Errorf("parameters cache is nil")
	}

	remoteConfig, err := firebase.ParseRemoteConfig(cache.RemoteConfig)
	if err != nil {
		return nil, err
	}

	tree := parameters.BuildTree(remoteConfig, cache.CachedAt, cache.ETag)

	corelog.For("core").Debug("built parameters tree", "version", tree.Version, "group_count", len(tree.Groups))
	return tree, nil
}

func (s *Core) BuildConditionsTree(cache *ParametersCache) (*ConditionsTree, error) {
	if cache == nil {
		return nil, fmt.Errorf("parameters cache is nil")
	}

	remoteConfig, err := firebase.ParseRemoteConfig(cache.RemoteConfig)
	if err != nil {
		return nil, err
	}

	tree := conditions.BuildTree(remoteConfig, cache.CachedAt, cache.ETag)
	corelog.For("core").Debug("built conditions tree", "version", tree.Version, "condition_count", len(tree.Conditions))
	return tree, nil
}

func (s *Core) fetchParameters(ctx context.Context, projectID string) (*ParametersCache, string, error) {
	return s.fetchParametersVersion(ctx, projectID, "")
}

func (s *Core) fetchLatestParameters(ctx context.Context, projectID string) (*ParametersCache, string, error) {
	logger := corelog.For("core")
	fb, err := s.firebaseServiceForProject(ctx, projectID)
	if err != nil {
		return nil, "", err
	}

	latestVersion, err := fb.GetLatestRemoteConfigVersion(ctx, projectID)
	if err != nil {
		logger.Error("remote config version check failed", "project_id", projectID, "err", err)
		return nil, "", fmt.Errorf("firebase error: %w", err)
	}

	if latestVersion.VersionNumber != "" {
		cache, err := config.LoadParametersCacheVersion(projectID, latestVersion.VersionNumber)
		if err == nil {
			return s.refreshVerifiedCache(ctx, projectID, cache, latestVersion.VersionNumber)
		}
		if !errors.Is(err, os.ErrNotExist) {
			return nil, "", err
		}
	}

	if latestVersion.VersionNumber == "" || latestVersion.VersionNumber == "NA" {
		return s.fetchParameters(ctx, projectID)
	}
	return s.fetchParametersVersion(ctx, projectID, latestVersion.VersionNumber)
}

func (s *Core) fetchParametersVersion(ctx context.Context, projectID, version string) (*ParametersCache, string, error) {
	logger := corelog.For("core")
	logger.Info("fetch parameters from firebase", "project_id", projectID, "version", version)

	fb, err := s.firebaseServiceForProject(ctx, projectID)
	if err != nil {
		return nil, "", err
	}

	raw, etag, err := fb.GetRemoteConfig(ctx, projectID, version)
	if err != nil {
		logger.Error("firebase remote config fetch failed", "project_id", projectID, "err", err)
		return nil, "", fmt.Errorf("firebase error: %w", err)
	}

	if _, err := firebase.ParseRemoteConfig(raw); err != nil {
		logger.Error("firebase remote config decode failed", "project_id", projectID, "err", err)
		return nil, "", fmt.Errorf("decode firebase remote config: %w", err)
	}

	cache := &config.ParametersCache{
		ETag:         etag,
		CachedAt:     time.Now().UTC(),
		RemoteConfig: raw,
	}
	if firebase.IsDryRun(ctx) {
		logger.Warn("dry run, skip parameters cache save after fetch", "project_id", projectID, "etag", etag)
		return cache, "firebase", nil
	}
	if err := config.SaveParametersCache(projectID, cache); err != nil {
		logger.Error("save parameters cache failed", "project_id", projectID, "etag", etag, "err", err)
		return nil, "", fmt.Errorf("save parameters cache: %w", err)
	}

	logger.Info("parameters cached from firebase", "project_id", projectID, "etag", etag)
	return cache, "firebase", nil
}

func (s *Core) verifyParameters(ctx context.Context, projectID string, cache *ParametersCache) (*ParametersCache, string, error) {
	logger := corelog.For("core")
	logger.Info("verify parameters cache against firebase", "project_id", projectID, "etag", cache.ETag)

	remoteConfig, err := firebase.ParseRemoteConfig(cache.RemoteConfig)
	if err != nil {
		logger.Error("decode cached remote config failed during verification", "project_id", projectID, "err", err)
		return nil, "", fmt.Errorf("decode cached remote config: %w", err)
	}

	fb, err := s.firebaseServiceForProject(ctx, projectID)
	if err != nil {
		return nil, "", err
	}

	latestVersion, err := fb.GetLatestRemoteConfigVersion(ctx, projectID)
	if err != nil {
		logger.Error("remote config version check failed", "project_id", projectID, "err", err)
		return nil, "", fmt.Errorf("firebase error: %w", err)
	}

	if latestVersion.VersionNumber != "" && latestVersion.VersionNumber == remoteConfig.Version.VersionNumber {
		return s.refreshVerifiedCache(ctx, projectID, cache, latestVersion.VersionNumber)
	}

	if latestVersion.VersionNumber != "" {
		latestCache, err := config.LoadParametersCacheVersion(projectID, latestVersion.VersionNumber)
		if err == nil {
			logger.Info("parameters latest version already cached", "project_id", projectID, "version", latestVersion.VersionNumber)
			return s.refreshVerifiedCache(ctx, projectID, latestCache, latestVersion.VersionNumber)
		}
		if !errors.Is(err, os.ErrNotExist) {
			return nil, "", err
		}
	}

	logger.Info("parameters cache outdated; refetching", "project_id", projectID, "cached_version", remoteConfig.Version.VersionNumber, "latest_version", latestVersion.VersionNumber)
	if latestVersion.VersionNumber == "" || latestVersion.VersionNumber == "NA" {
		return s.fetchParameters(ctx, projectID)
	}
	return s.fetchParametersVersion(ctx, projectID, latestVersion.VersionNumber)
}

func (s *Core) refreshVerifiedCache(ctx context.Context, projectID string, cache *ParametersCache, version string) (*ParametersCache, string, error) {
	logger := corelog.For("core")
	refreshed := *cache
	refreshed.CachedAt = time.Now().UTC()
	if firebase.IsDryRun(ctx) {
		logger.Warn("dry run, skip parameters cache timestamp refresh", "project_id", projectID, "version", version)
		return &refreshed, "cache-verified", nil
	}
	if err := config.SaveParametersCache(projectID, &refreshed); err != nil {
		logger.Error("refresh parameters cache timestamp failed", "project_id", projectID, "err", err)
		return nil, "", fmt.Errorf("refresh parameters cache timestamp: %w", err)
	}
	logger.Info("parameters cache verified as current", "project_id", projectID, "version", version)
	return &refreshed, "cache-verified", nil
}

func ParametersStatusLabel(source string, cachedAt time.Time, hasTree bool, err error) string {
	if err != nil && hasTree {
		return "error"
	}
	switch source {
	case "firebase":
		return "fetch"
	case "cache", "cache-verified", "cache-stale":
		if time.Since(cachedAt) > 10*time.Minute {
			return "staled"
		}
		if time.Since(cachedAt) < time.Minute {
			return "fetch"
		}
		return "cached"
	default:
		return ""
	}
}
