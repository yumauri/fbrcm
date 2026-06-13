package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/firebase"
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
	sort.Slice(entries, func(i, j int) bool {
		left := strings.ToLower(entries[i].ProjectID)
		right := strings.ToLower(entries[j].ProjectID)
		if left == right {
			if entries[i].Draft != entries[j].Draft {
				return !entries[i].Draft
			}
			return entries[i].ProjectID < entries[j].ProjectID
		}
		return left < right
	})
}

func loadParametersCacheEntries(projectNames map[string]string) ([]cacheEntry, error) {
	dir := config.GetParametersCacheDirPath()
	files, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []cacheEntry{}, nil
		}
		return nil, fmt.Errorf("read cache dir: %w", err)
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
			return nil, fmt.Errorf("stat cache file %s: %w", path, err)
		}

		cache, err := config.LoadParametersCache(projectID)
		if err != nil {
			return nil, err
		}

		version := ""
		if remoteConfig, err := firebase.ParseRemoteConfig(cache.RemoteConfig); err == nil {
			version = remoteConfig.Version.VersionNumber
		}
		cachedAt := cache.CachedAt
		entries = append(entries, cacheEntry{
			ProjectID: projectID,
			Project:   projectNames[projectID],
			Version:   version,
			CachedAt:  &cachedAt,
			Size:      info.Size(),
			Path:      path,
		})
	}
	return entries, nil
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
