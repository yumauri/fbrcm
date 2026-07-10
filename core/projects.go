package core

import (
	"slices"
	"strings"
	"time"

	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/filter"
	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/core/strfold"
)

func toConfigProjects(projects []firebase.Project) []config.Project {
	out := make([]config.Project, len(projects))
	for i, p := range projects {
		out[i] = config.Project{
			Name:          p.Name,
			ProjectID:     p.ProjectID,
			ProjectNumber: p.ProjectNumber,
			State:         p.State,
			ETag:          p.ETag,
			UpdatedAt:     p.UpdateTime,
		}
	}
	return out
}

func mergeProjects(existing, incoming []config.Project, defaultAuthID string, authIDs []string, onlyAuthID string, now time.Time) []config.Project {
	byID := make(map[string]config.Project, len(existing))
	for _, project := range existing {
		byID[project.ProjectID] = project
	}

	updatedAt := now.Format(time.RFC3339)
	mergedByID := make(map[string]config.Project, len(existing)+len(incoming))
	if onlyAuthID != "" {
		for _, project := range existing {
			mergedByID[project.ProjectID] = project
		}
	}
	for _, project := range incoming {
		if previous, ok := byID[project.ProjectID]; ok {
			project.AuthID = chooseProjectAuth(previous.AuthID, project.DiscoveredBy, defaultAuthID, authIDs)
			if sameProject(previous, project) {
				project.SyncedAt = previous.SyncedAt
			} else {
				project.SyncedAt = updatedAt
			}
		} else {
			project.AuthID = chooseProjectAuth("", project.DiscoveredBy, defaultAuthID, authIDs)
			project.SyncedAt = updatedAt
		}
		mergedByID[project.ProjectID] = project
	}

	merged := make([]config.Project, 0, len(mergedByID))
	for _, project := range mergedByID {
		merged = append(merged, project)
	}
	strfold.SortProjects(merged, func(p config.Project) string { return p.Name }, func(p config.Project) string { return p.ProjectID })
	return merged
}

func sameProject(left, right config.Project) bool {
	authSame := left.AuthID == right.AuthID &&
		strings.Join(left.DiscoveredBy, "\x00") == strings.Join(right.DiscoveredBy, "\x00")
	if left.ETag != "" && right.ETag != "" {
		return left.ETag == right.ETag && authSame
	}
	if left.UpdatedAt != "" && right.UpdatedAt != "" {
		return left.UpdatedAt == right.UpdatedAt && authSame
	}

	return left.Name == right.Name &&
		left.ProjectNumber == right.ProjectNumber &&
		left.State == right.State &&
		authSame
}

func chooseProjectAuth(previous string, discovered []string, defaultAuthID string, authIDs []string) string {
	if contains(discovered, previous) {
		return previous
	}
	if contains(discovered, defaultAuthID) {
		return defaultAuthID
	}
	for _, authID := range authIDs {
		if contains(discovered, authID) {
			return authID
		}
	}
	if len(discovered) > 0 {
		return discovered[0]
	}
	return previous
}

func authOrder(auth []config.AuthEntry) []string {
	out := make([]string, 0, len(auth))
	for _, entry := range auth {
		out = append(out, entry.ID)
	}
	return out
}

func appendUnique(values []string, value string) []string {
	if contains(values, value) {
		return values
	}
	return append(values, value)
}

func contains(values []string, value string) bool {
	return slices.Contains(values, value)
}

func matchProjectFilter(project Project, rawFilters []string) bool {
	if len(rawFilters) == 0 {
		return true
	}
	for _, raw := range rawFilters {
		mode, query := filter.ParseModePrefixedQuery(raw)
		if query == "" {
			continue
		}
		nameMatch, _ := filter.Match(project.Name, query, mode)
		idMatch, _ := filter.Match(project.ProjectID, query, mode)
		if nameMatch || idMatch {
			return true
		}
	}
	return false
}
