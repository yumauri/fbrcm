package core

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/firebase"
)

type VersionListOptions struct {
	Limit      int
	All        bool
	Before     string
	Since      time.Time
	Until      time.Time
	PageToken  string
	CachedOnly bool
}

type RemoteConfigVersionEntry struct {
	firebase.RemoteConfigVersion
	Current  bool      `json:"current"`
	Cached   bool      `json:"cached"`
	CachedAt time.Time `json:"cached_at,omitzero"`
	Size     int64     `json:"size,omitempty"`
	Path     string    `json:"path,omitempty"`
}

type RemoteConfigVersionList struct {
	Versions      []RemoteConfigVersionEntry `json:"versions"`
	NextPageToken string                     `json:"next_page_token,omitempty"`
}

type ResolvedRemoteConfigVersion struct {
	Version firebase.RemoteConfigVersion
	Cache   *ParametersCache
	Config  *firebase.RemoteConfig
	Cached  bool
}

type VersionPublishResult struct {
	PreviousVersion  string `json:"previous_version"`
	SourceVersion    string `json:"source_version"`
	PublishedVersion string `json:"published_version"`
	RemoteConfig     []byte `json:"-"`
	ETag             string `json:"-"`
}

func (s *Core) ListRemoteConfigVersions(ctx context.Context, projectID string, opts VersionListOptions) (RemoteConfigVersionList, error) {
	snapshots, err := config.ListParametersCacheSnapshotsForProject(projectID)
	if err != nil {
		return RemoteConfigVersionList{}, err
	}
	byVersion := make(map[string]config.ParametersCacheSnapshot, len(snapshots))
	for _, snapshot := range snapshots {
		byVersion[snapshot.Version] = snapshot
	}
	current := ""
	if cache, err := config.LoadParametersCache(projectID); err == nil {
		if cfg, parseErr := firebase.ParseRemoteConfig(cache.RemoteConfig); parseErr == nil {
			current = cfg.Version.VersionNumber
		}
	}
	if opts.CachedOnly {
		entries := make([]RemoteConfigVersionEntry, 0, len(snapshots))
		for _, snapshot := range snapshots {
			if opts.Before != "" && compareNumericVersion(snapshot.Version, opts.Before) > 0 {
				continue
			}
			if !opts.Since.IsZero() && snapshot.Cache.CachedAt.Before(opts.Since) {
				continue
			}
			if !opts.Until.IsZero() && !snapshot.Cache.CachedAt.Before(opts.Until) {
				continue
			}
			entry := RemoteConfigVersionEntry{RemoteConfigVersion: firebase.RemoteConfigVersion{VersionNumber: snapshot.Version}, Current: snapshot.Version == current, Cached: true, CachedAt: snapshot.Cache.CachedAt, Size: snapshot.Size, Path: snapshot.Path}
			if cfg, parseErr := firebase.ParseRemoteConfig(snapshot.Cache.RemoteConfig); parseErr == nil {
				entry.RemoteConfigVersion = cfg.Version
			}
			entries = append(entries, entry)
		}
		sortVersionEntries(entries)
		if !opts.All && opts.Limit > 0 && len(entries) > opts.Limit {
			entries = entries[:opts.Limit]
		}
		return RemoteConfigVersionList{Versions: entries}, nil
	}
	fb, err := s.firebaseServiceForProject(ctx, projectID)
	if err != nil {
		return RemoteConfigVersionList{}, err
	}
	remaining := opts.Limit
	if remaining <= 0 {
		remaining = 20
	}
	token := opts.PageToken
	result := RemoteConfigVersionList{}
	for {
		pageSize := remaining
		if opts.All {
			pageSize = 100
		}
		page, err := fb.ListRemoteConfigVersions(ctx, projectID, firebase.ListVersionsOptions{PageSize: pageSize, PageToken: token, EndVersionNumber: opts.Before, StartTime: formatOptionalTime(opts.Since), EndTime: formatOptionalTime(opts.Until)})
		if err != nil {
			return RemoteConfigVersionList{}, fmt.Errorf("firebase error: %w", err)
		}
		for _, version := range page.Versions {
			entry := RemoteConfigVersionEntry{RemoteConfigVersion: version, Current: version.VersionNumber == current}
			if snapshot, ok := byVersion[version.VersionNumber]; ok {
				entry.Cached, entry.CachedAt, entry.Size, entry.Path = true, snapshot.Cache.CachedAt, snapshot.Size, snapshot.Path
			}
			result.Versions = append(result.Versions, entry)
		}
		result.NextPageToken = page.NextPageToken
		if !opts.All || page.NextPageToken == "" {
			break
		}
		token = page.NextPageToken
	}
	if len(result.Versions) > 0 {
		for i := range result.Versions {
			result.Versions[i].Current = i == 0
		}
	}
	return result, nil
}

func (s *Core) GetRemoteConfigVersion(ctx context.Context, projectID, selector string, cachedOnly bool) (*ResolvedRemoteConfigVersion, error) {
	version, err := s.resolveVersionSelector(ctx, projectID, selector, cachedOnly)
	if err != nil {
		return nil, err
	}
	return s.getRemoteConfigVersionNumber(ctx, projectID, version, cachedOnly)
}

// GetRemoteConfigVersionPair resolves two selectors against the same Firebase
// history response when both are relative to the current version. This keeps a
// diff internally consistent and avoids listing version history twice.
func (s *Core) GetRemoteConfigVersionPair(ctx context.Context, projectID, fromSelector, toSelector string, cachedOnly bool) (*ResolvedRemoteConfigVersion, *ResolvedRemoteConfigVersion, error) {
	key := projectID + "\x00" + fromSelector + "\x00" + toSelector + "\x00" + strconv.FormatBool(cachedOnly)
	value, err, _ := s.versionHistory.Do(key, func() (any, error) {
		from, to, err := s.getRemoteConfigVersionPair(ctx, projectID, fromSelector, toSelector, cachedOnly)
		if err != nil {
			return nil, err
		}
		return remoteConfigVersionPair{from: from, to: to}, nil
	})
	if err != nil {
		return nil, nil, err
	}
	pair := value.(remoteConfigVersionPair)
	return pair.from, pair.to, nil
}

type remoteConfigVersionPair struct {
	from *ResolvedRemoteConfigVersion
	to   *ResolvedRemoteConfigVersion
}

func (s *Core) getRemoteConfigVersionPair(ctx context.Context, projectID, fromSelector, toSelector string, cachedOnly bool) (*ResolvedRemoteConfigVersion, *ResolvedRemoteConfigVersion, error) {
	if cachedOnly {
		from, err := s.GetRemoteConfigVersion(ctx, projectID, fromSelector, true)
		if err != nil {
			return nil, nil, err
		}
		to, err := s.GetRemoteConfigVersion(ctx, projectID, toSelector, true)
		return from, to, err
	}

	fromDistance, fromRelative, err := currentRelativeDistance(fromSelector)
	if err != nil {
		return nil, nil, err
	}
	toDistance, toRelative, err := currentRelativeDistance(toSelector)
	if err != nil {
		return nil, nil, err
	}
	if !fromRelative || !toRelative {
		from, err := s.GetRemoteConfigVersion(ctx, projectID, fromSelector, false)
		if err != nil {
			return nil, nil, err
		}
		to, err := s.GetRemoteConfigVersion(ctx, projectID, toSelector, false)
		return from, to, err
	}

	maxDistance := max(fromDistance, toDistance)
	fb, err := s.firebaseServiceForProject(ctx, projectID)
	if err != nil {
		return nil, nil, err
	}
	page, err := fb.ListRemoteConfigVersions(ctx, projectID, firebase.ListVersionsOptions{PageSize: maxDistance + 1})
	if err != nil {
		return nil, nil, err
	}
	if len(page.Versions) <= maxDistance {
		return nil, nil, fmt.Errorf("project %s has no Remote Config version %d publications before current", projectID, maxDistance)
	}
	from, err := s.getRemoteConfigVersionNumber(ctx, projectID, page.Versions[fromDistance].VersionNumber, false)
	if err != nil {
		return nil, nil, err
	}
	to, err := s.getRemoteConfigVersionNumber(ctx, projectID, page.Versions[toDistance].VersionNumber, false)
	return from, to, err
}

func (s *Core) getRemoteConfigVersionNumber(ctx context.Context, projectID, version string, cachedOnly bool) (*ResolvedRemoteConfigVersion, error) {
	cache, err := config.LoadParametersCacheVersion(projectID, version)
	if err == nil {
		cfg, parseErr := firebase.ParseCloneRemoteConfig(cache.RemoteConfig)
		if parseErr != nil {
			return nil, fmt.Errorf("decode cached Remote Config version %s: %w", version, parseErr)
		}
		if cfg.Version.VersionNumber != version {
			return nil, fmt.Errorf("cached Remote Config version mismatch: requested %s, got %s", version, cfg.Version.VersionNumber)
		}
		return &ResolvedRemoteConfigVersion{Version: cfg.Version, Cache: cache, Config: cfg, Cached: true}, nil
	}
	if !errors.Is(err, os.ErrNotExist) && !strings.Contains(err.Error(), "no such file") {
		return nil, err
	}
	if cachedOnly {
		return nil, fmt.Errorf("cached Remote Config version %s was not found for project %s", version, projectID)
	}
	fb, err := s.firebaseServiceForProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	raw, etag, err := fb.GetRemoteConfig(ctx, projectID, version)
	if err != nil {
		return nil, fmt.Errorf("firebase Remote Config version %s: %w", version, err)
	}
	cfg, err := firebase.ParseCloneRemoteConfig(raw)
	if err != nil {
		return nil, err
	}
	if cfg.Version.VersionNumber != version {
		return nil, fmt.Errorf("firebase returned version %s when version %s was requested", cfg.Version.VersionNumber, version)
	}
	cache = &config.ParametersCache{ETag: etag, CachedAt: time.Now().UTC(), RemoteConfig: raw}
	if !firebase.IsDryRun(ctx) {
		if err := config.SaveParametersCacheSnapshot(projectID, cache); err != nil {
			return nil, fmt.Errorf("cache historical version: %w", err)
		}
	}
	return &ResolvedRemoteConfigVersion{Version: cfg.Version, Cache: cache, Config: cfg}, nil
}

func currentRelativeDistance(selector string) (int, bool, error) {
	selector = strings.TrimSpace(selector)
	if selector == "current" || selector == "latest" {
		return 0, true, nil
	}
	if selector == "previous" {
		return 1, true, nil
	}
	_, distance, ok, err := parseRelativeVersionSelector(selector)
	return distance, ok, err
}

func (s *Core) RollbackRemoteConfig(ctx context.Context, projectID, sourceVersion string) (VersionPublishResult, error) {
	currentRaw, _, err := s.ExportRemoteConfig(ctx, projectID)
	if err != nil {
		return VersionPublishResult{}, err
	}
	current, err := firebase.ParseRemoteConfig(currentRaw)
	if err != nil {
		return VersionPublishResult{}, err
	}
	fb, err := s.firebaseServiceForProject(ctx, projectID)
	if err != nil {
		return VersionPublishResult{}, err
	}
	raw, etag, err := fb.RollbackRemoteConfig(ctx, projectID, sourceVersion)
	if err != nil {
		return VersionPublishResult{}, fmt.Errorf("firebase error: %w", err)
	}
	published, err := firebase.ParseRemoteConfig(raw)
	if err != nil {
		return VersionPublishResult{}, err
	}
	if !firebase.IsDryRun(ctx) {
		if err := config.SaveParametersCache(projectID, &config.ParametersCache{ETag: etag, CachedAt: time.Now().UTC(), RemoteConfig: raw}); err != nil {
			return VersionPublishResult{PreviousVersion: current.Version.VersionNumber, SourceVersion: sourceVersion, PublishedVersion: published.Version.VersionNumber, RemoteConfig: raw, ETag: etag}, fmt.Errorf("rollback succeeded but cache update failed: %w", err)
		}
	}
	return VersionPublishResult{PreviousVersion: current.Version.VersionNumber, SourceVersion: sourceVersion, PublishedVersion: published.Version.VersionNumber, RemoteConfig: raw, ETag: etag}, nil
}

func (s *Core) RestoreRemoteConfigVersion(ctx context.Context, projectID, sourceVersion string) (VersionPublishResult, error) {
	source, err := s.GetRemoteConfigVersion(ctx, projectID, sourceVersion, true)
	if err != nil {
		return VersionPublishResult{}, err
	}
	currentRaw, etag, err := s.ExportRemoteConfig(ctx, projectID)
	if err != nil {
		return VersionPublishResult{}, err
	}
	current, err := firebase.ParseRemoteConfig(currentRaw)
	if err != nil {
		return VersionPublishResult{}, err
	}
	clone, err := firebase.ParseCloneRemoteConfig(source.Cache.RemoteConfig)
	if err != nil {
		return VersionPublishResult{}, err
	}
	clone.Version = firebase.RemoteConfigVersion{}
	raw, err := firebase.MarshalRemoteConfig(clone)
	if err != nil {
		return VersionPublishResult{}, err
	}
	if err := s.ValidateRemoteConfigWithETag(ctx, projectID, raw, etag); err != nil {
		return VersionPublishResult{}, err
	}
	publishedRaw, nextETag, publishErr := s.PublishRemoteConfigWithETag(ctx, projectID, raw, etag)
	if publishErr != nil && (len(publishedRaw) == 0 || nextETag == "") {
		return VersionPublishResult{}, publishErr
	}
	if firebase.IsDryRun(ctx) {
		return VersionPublishResult{PreviousVersion: current.Version.VersionNumber, SourceVersion: sourceVersion}, nil
	}
	published, err := firebase.ParseRemoteConfig(publishedRaw)
	if err != nil {
		return VersionPublishResult{}, err
	}
	result := VersionPublishResult{PreviousVersion: current.Version.VersionNumber, SourceVersion: sourceVersion, PublishedVersion: published.Version.VersionNumber, RemoteConfig: publishedRaw, ETag: nextETag}
	return result, publishErr
}

func (s *Core) resolveVersionSelector(ctx context.Context, projectID, selector string, cachedOnly bool) (string, error) {
	selector = strings.TrimSpace(selector)
	if _, distance, ok, err := parseRelativeVersionSelector(selector); err != nil {
		return "", err
	} else if ok {
		return s.resolveRelativeVersion(ctx, projectID, distance, cachedOnly)
	}
	if selector == "current" || selector == "latest" {
		if cachedOnly {
			cache, err := config.LoadParametersCache(projectID)
			if err != nil {
				return "", err
			}
			cfg, err := firebase.ParseRemoteConfig(cache.RemoteConfig)
			if err != nil {
				return "", err
			}
			return cfg.Version.VersionNumber, nil
		}
		fb, err := s.firebaseServiceForProject(ctx, projectID)
		if err != nil {
			return "", err
		}
		latest, err := fb.GetLatestRemoteConfigVersion(ctx, projectID)
		if err != nil {
			return "", err
		}
		return latest.VersionNumber, nil
	}
	if selector == "previous" {
		return s.resolveRelativeVersion(ctx, projectID, 1, cachedOnly)
	}
	n, err := strconv.ParseInt(selector, 10, 64)
	if err != nil || n <= 0 {
		return "", fmt.Errorf("invalid Remote Config version %q", selector)
	}
	return selector, nil
}

func parseRelativeVersionSelector(selector string) (base string, distance int, ok bool, err error) {
	base, rawDistance, found := strings.Cut(selector, "~")
	if !found {
		return "", 0, false, nil
	}
	if base != "current" && base != "latest" {
		return "", 0, false, fmt.Errorf("invalid relative Remote Config version %q: use current~N or latest~N", selector)
	}
	if rawDistance == "" || strings.Contains(rawDistance, "~") {
		return "", 0, false, fmt.Errorf("invalid relative Remote Config version %q: distance must be a positive integer", selector)
	}
	distance64, parseErr := strconv.ParseInt(rawDistance, 10, 32)
	if parseErr != nil || distance64 <= 0 {
		return "", 0, false, fmt.Errorf("invalid relative Remote Config version %q: distance must be a positive integer", selector)
	}
	if distance64 > 299 {
		return "", 0, false, fmt.Errorf("relative Remote Config version distance %d exceeds Firebase's 300-version history limit", distance64)
	}
	return base, int(distance64), true, nil
}

func (s *Core) resolveRelativeVersion(ctx context.Context, projectID string, distance int, cachedOnly bool) (string, error) {
	if cachedOnly {
		return relativeCachedVersion(projectID, distance)
	}
	fb, err := s.firebaseServiceForProject(ctx, projectID)
	if err != nil {
		return "", err
	}
	page, err := fb.ListRemoteConfigVersions(ctx, projectID, firebase.ListVersionsOptions{PageSize: distance + 1})
	if err != nil {
		return "", err
	}
	if len(page.Versions) <= distance {
		return "", fmt.Errorf("project %s has no Remote Config version %d publications before current", projectID, distance)
	}
	return page.Versions[distance].VersionNumber, nil
}

func relativeCachedVersion(projectID string, distance int) (string, error) {
	currentCache, err := config.LoadParametersCache(projectID)
	if err != nil {
		return "", err
	}
	currentConfig, err := firebase.ParseRemoteConfig(currentCache.RemoteConfig)
	if err != nil {
		return "", err
	}
	current, err := strconv.ParseInt(currentConfig.Version.VersionNumber, 10, 64)
	if err != nil || current <= 0 {
		return "", fmt.Errorf("cached current Remote Config version %q is not numeric", currentConfig.Version.VersionNumber)
	}
	snapshots, err := config.ListParametersCacheSnapshotsForProject(projectID)
	if err != nil {
		return "", err
	}
	versions := make([]int64, 0, len(snapshots))
	for _, snapshot := range snapshots {
		version, parseErr := strconv.ParseInt(snapshot.Version, 10, 64)
		if parseErr == nil && version < current {
			versions = append(versions, version)
		}
	}
	slices.SortFunc(versions, func(left, right int64) int {
		if left > right {
			return -1
		}
		if left < right {
			return 1
		}
		return 0
	})
	if len(versions) < distance {
		return "", fmt.Errorf("project %s has no cached Remote Config version %d snapshots before current", projectID, distance)
	}
	return strconv.FormatInt(versions[distance-1], 10), nil
}

func formatOptionalTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}
func sortVersionEntries(entries []RemoteConfigVersionEntry) {
	slices.SortFunc(entries, func(a, b RemoteConfigVersionEntry) int {
		an, _ := strconv.ParseInt(a.VersionNumber, 10, 64)
		bn, _ := strconv.ParseInt(b.VersionNumber, 10, 64)
		if an > bn {
			return -1
		}
		if an < bn {
			return 1
		}
		return 0
	})
}

func compareNumericVersion(left, right string) int {
	l, lerr := strconv.ParseInt(strings.TrimSpace(left), 10, 64)
	r, rerr := strconv.ParseInt(strings.TrimSpace(right), 10, 64)
	if lerr != nil || rerr != nil {
		return strings.Compare(left, right)
	}
	if l < r {
		return -1
	}
	if l > r {
		return 1
	}
	return 0
}
