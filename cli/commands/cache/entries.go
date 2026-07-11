package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/core/strfold"
)

type cacheEntry struct {
	ProjectID string     `json:"project_id"`
	Project   string     `json:"project"`
	Version   string     `json:"version"`
	Size      int64      `json:"size"`
	CachedAt  *time.Time `json:"cached_at"`
	Draft     bool       `json:"draft"`
	Path      string     `json:"path"`
}

func loadCacheEntries() ([]cacheEntry, error) {
	projectNames := loadProjectNames()
	entries, err := loadParametersCacheEntries(projectNames)
	if err != nil {
		return nil, err
	}
	draftEntries, err := loadDraftEntries(projectNames)
	if err != nil {
		return nil, err
	}
	entries = append(entries, draftEntries...)
	sortCacheEntries(entries)
	return entries, nil
}

func sortCacheEntries(entries []cacheEntry) {
	slices.SortFunc(entries, func(left, right cacheEntry) int {
		if cmp := strfold.CompareFolded(left.ProjectID, right.ProjectID); cmp != 0 {
			return cmp
		}
		if left.Draft != right.Draft {
			if !left.Draft {
				return -1
			}
			return 1
		}
		if !left.Draft {
			if cmp := compareVersionsDesc(left.Version, right.Version); cmp != 0 {
				return cmp
			}
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
		version := snapshot.Version
		if remoteConfig, err := firebase.ParseRemoteConfig(snapshot.Cache.RemoteConfig); err == nil {
			version = remoteConfig.Version.VersionNumber
		}
		cachedAt := snapshot.Cache.CachedAt
		entries = append(entries, cacheEntry{
			ProjectID: snapshot.ProjectID,
			Project:   projectNames[snapshot.ProjectID],
			Version:   version,
			CachedAt:  &cachedAt,
			Size:      snapshot.Size,
			Path:      snapshot.Path,
		})
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

func loadDraftEntries(projectNames map[string]string) ([]cacheEntry, error) {
	dir := config.GetDraftsDirPath()
	files, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []cacheEntry{}, nil
		}
		return nil, fmt.Errorf("read drafts dir: %w", err)
	}

	entries := make([]cacheEntry, 0, len(files))
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		projectID := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
		path := filepath.Join(dir, file.Name())
		info, err := file.Info()
		if err != nil {
			return nil, fmt.Errorf("stat draft file %s: %w", path, err)
		}

		raw, err := config.LoadDraft(projectID)
		if err != nil {
			return nil, err
		}

		version := ""
		if remoteConfig, err := firebase.ParseRemoteConfig(raw); err == nil {
			version = remoteConfig.Version.VersionNumber
		}

		entries = append(entries, cacheEntry{
			ProjectID: projectID,
			Project:   projectNames[projectID],
			Version:   version,
			Size:      info.Size(),
			Draft:     true,
			Path:      path,
		})
	}
	return entries, nil
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
