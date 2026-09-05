package cache

import (
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/firebase"
	rctarget "github.com/yumauri/fbrcm/core/rc/target"
	"github.com/yumauri/fbrcm/core/strfold"
)

type cacheEntry struct {
	Kind      string     `json:"kind" contract:"enum=remote-config|apps-index|app-details|app-config"`
	ProjectID string     `json:"project_id"`
	Project   string     `json:"project"`
	Resource  string     `json:"resource,omitempty"`
	Version   string     `json:"version,omitempty"`
	Size      int64      `json:"size"`
	CachedAt  *time.Time `json:"cached_at"`
	Path      string     `json:"path"`
}

func loadCacheEntries(kind string) ([]cacheEntry, error) {
	projectNames := loadProjectNames()
	var entries []cacheEntry
	if kind == "all" || kind == "remote-config" {
		remoteConfigEntries, err := loadParametersCacheEntries(projectNames)
		if err != nil {
			return nil, err
		}
		entries = append(entries, remoteConfigEntries...)
	}
	if kind == "all" || kind == "apps" {
		appEntries, err := loadAppsCacheEntries(projectNames)
		if err != nil {
			return nil, err
		}
		entries = append(entries, appEntries...)
	}
	sortCacheEntries(entries)
	return entries, nil
}

func sortCacheEntries(entries []cacheEntry) {
	slices.SortFunc(entries, func(left, right cacheEntry) int {
		if cmp := strfold.CompareFolded(left.ProjectID, right.ProjectID); cmp != 0 {
			return cmp
		}
		if cmp := compareVersionsDesc(left.Version, right.Version); cmp != 0 {
			return cmp
		}
		if cmp := strfold.CompareFolded(left.Kind, right.Kind); cmp != 0 {
			return cmp
		}
		if cmp := strfold.CompareFolded(left.Resource, right.Resource); cmp != 0 {
			return cmp
		}
		return strfold.Compare(left.ProjectID, right.ProjectID)
	})
}

func loadParametersCacheEntries(projectNames map[string]string) ([]cacheEntry, error) {
	snapshots, err := config.ListParametersCacheSnapshots()
	if err != nil {
		return nil, err
	}

	entries := make([]cacheEntry, 0, len(snapshots))
	for _, snapshot := range snapshots {
		target, targetErr := rctarget.Parse(snapshot.ProjectID)
		if targetErr != nil {
			return nil, targetErr
		}
		version := snapshot.Version
		if remoteConfig, err := firebase.ParseRemoteConfig(snapshot.Cache.RemoteConfig); err == nil {
			version = remoteConfig.Version.VersionNumber
		}
		cachedAt := snapshot.Cache.CachedAt
		entries = append(entries, cacheEntry{
			Kind:      "remote-config",
			ProjectID: snapshot.ProjectID,
			Project:   projectNames[target.ProjectID],
			Version:   version,
			CachedAt:  &cachedAt,
			Size:      snapshot.Size,
			Path:      snapshot.Path,
		})
	}
	return entries, nil
}

func loadAppsCacheEntries(projectNames map[string]string) ([]cacheEntry, error) {
	cached, err := config.ListAppsCacheEntries()
	if err != nil {
		return nil, err
	}
	entries := make([]cacheEntry, 0, len(cached))
	for _, item := range cached {
		kind := "app-" + item.Kind
		if item.Kind == "index" {
			kind = "apps-index"
		}
		cachedAt := item.CachedAt
		entries = append(entries, cacheEntry{Kind: kind, ProjectID: item.ProjectID, Project: projectNames[item.ProjectID], Resource: item.AppID, Size: item.Size, CachedAt: &cachedAt, Path: item.Path})
	}
	return entries, nil
}

func compareVersionsDesc(left, right string) int {
	leftN, leftOK := parseCacheVersion(left)
	rightN, rightOK := parseCacheVersion(right)
	if leftOK && rightOK && leftN != rightN {
		if leftN > rightN {
			return -1
		}
		return 1
	}
	if leftOK != rightOK {
		if leftOK {
			return -1
		}
		return 1
	}
	return strfold.Compare(right, left)
}

func parseCacheVersion(version string) (int64, bool) {
	n, err := strconv.ParseInt(strings.TrimSpace(version), 10, 64)
	if err != nil || n <= 0 {
		return 0, false
	}
	return n, true
}

func loadProjectNames() map[string]string {
	projects, err := config.LoadProjects()
	if err != nil {
		return nil
	}

	names := make(map[string]string, len(projects))
	for _, project := range projects {
		names[project.ProjectID] = project.Name
	}
	return names
}
