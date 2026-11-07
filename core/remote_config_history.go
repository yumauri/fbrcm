package core

import (
	"context"
	"fmt"
	"strconv"

	"github.com/yumauri/fbrcm/core/firebase"
	rcdiff "github.com/yumauri/fbrcm/core/rc/diff"
)

// RemoteConfigHistoryOptions controls the starting publication and number of
// entity changes inspected in retained Firebase history.
type RemoteConfigHistoryOptions struct {
	At    string
	Limit int
	All   bool
}

type remoteConfigHistoryChange[T any] struct {
	PreviousVersion string
	Version         firebase.RemoteConfigVersion
	Change          T
}

type remoteConfigHistoryBoundary struct {
	Version string
	Present bool
}

type remoteConfigHistory[T any] struct {
	AtVersion           string
	AtConfig            *firebase.RemoteConfig
	Changes             []remoteConfigHistoryChange[T]
	ScannedVersionCount int
	HistoryExhausted    bool
	Boundary            *remoteConfigHistoryBoundary
	BoundaryConfig      *firebase.RemoteConfig
	Seen                bool
}

// scanRemoteConfigHistory shares lazy pagination, selector resolution and
// adjacent-publication traversal across all entity history views.
func scanRemoteConfigHistory[T any](ctx context.Context, s *Core, projectID string, opts RemoteConfigHistoryOptions,
	present func(*firebase.RemoteConfig) bool,
	compare func(newer, older *firebase.RemoteConfig) (T, rcdiff.ChangeKind, bool),
) (remoteConfigHistory[T], error) {
	selector := opts.At
	if selector == "" {
		selector = "current"
	}
	if opts.Limit < 1 {
		return remoteConfigHistory[T]{}, fmt.Errorf("remote config history limit must be greater than zero")
	}

	distance, relative, err := currentRelativeDistance(selector)
	if err != nil {
		return remoteConfigHistory[T]{}, invalidVersionSelector(projectID, selector, err)
	}
	var number int64
	if !relative {
		number, err = strconv.ParseInt(selector, 10, 64)
		if err != nil || number <= 0 {
			return remoteConfigHistory[T]{}, invalidVersionSelector(projectID, selector, fmt.Errorf("invalid Remote Config version %q", selector))
		}
	}

	versions := make([]RemoteConfigVersionEntry, 0, 100)
	nextPageToken := ""
	historyExhausted := false
	loadNextPage := func() error {
		page, listErr := s.ListRemoteConfigVersions(ctx, projectID, VersionListOptions{Limit: 100, PageToken: nextPageToken})
		if listErr != nil {
			return listErr
		}
		versions = append(versions, page.Versions...)
		historyExhausted = page.NextPageToken == ""
		nextPageToken = page.NextPageToken
		return nil
	}

	anchor := -1
	searchFrom := 0
	for anchor < 0 {
		if err := loadNextPage(); err != nil {
			return remoteConfigHistory[T]{}, err
		}
		if relative {
			if distance < len(versions) {
				anchor = distance
			}
		} else {
			for index := searchFrom; index < len(versions); index++ {
				candidate, parseErr := strconv.ParseInt(versions[index].VersionNumber, 10, 64)
				if parseErr == nil && candidate == number {
					anchor = index
					break
				}
			}
			searchFrom = len(versions)
		}
		if anchor < 0 && historyExhausted {
			if len(versions) == 0 {
				return remoteConfigHistory[T]{}, versionNotFound(projectID, selector, fmt.Errorf("project %s has no retained Remote Config versions", projectID))
			}
			if relative {
				return remoteConfigHistory[T]{}, versionNotFound(projectID, selector, fmt.Errorf("project %s has no Remote Config version %d publications before current", projectID, distance))
			}
			return remoteConfigHistory[T]{}, versionNotFound(projectID, selector, fmt.Errorf("remote config version %s was not found in retained history for project %s", selector, projectID))
		}
	}

	result := remoteConfigHistory[T]{
		AtVersion: versions[anchor].VersionNumber,
		Changes:   make([]remoteConfigHistoryChange[T], 0),
	}

	newer, err := s.GetRemoteConfigVersion(ctx, projectID, versions[anchor].VersionNumber, false)
	if err != nil {
		return remoteConfigHistory[T]{}, err
	}
	result.ScannedVersionCount = 1
	result.AtConfig = newer.Config
	result.Seen = present(newer.Config)

	cursor := anchor
	for {
		if cursor+1 >= len(versions) {
			if historyExhausted {
				break
			}
			if err := loadNextPage(); err != nil {
				return remoteConfigHistory[T]{}, err
			}
			if cursor+1 >= len(versions) {
				if historyExhausted {
					break
				}
				continue
			}
		}

		olderEntry := versions[cursor+1]
		older, getErr := s.GetRemoteConfigVersion(ctx, projectID, olderEntry.VersionNumber, false)
		if getErr != nil {
			return remoteConfigHistory[T]{}, getErr
		}
		result.ScannedVersionCount++

		change, kind, found := compare(newer.Config, older.Config)
		if found {
			result.Seen = true
			if kind != rcdiff.ChangeUnchanged {
				result.Changes = append(result.Changes, remoteConfigHistoryChange[T]{
					PreviousVersion: olderEntry.VersionNumber,
					Version:         versions[cursor].RemoteConfigVersion,
					Change:          change,
				})
			}
		}
		newer = older
		cursor++

		if !opts.All && len(result.Changes) >= opts.Limit {
			break
		}
	}

	result.HistoryExhausted = historyExhausted && cursor == len(versions)-1
	if result.HistoryExhausted {
		result.Boundary = &remoteConfigHistoryBoundary{Version: versions[cursor].VersionNumber, Present: present(newer.Config)}
		result.BoundaryConfig = newer.Config
	}
	return result, nil
}
