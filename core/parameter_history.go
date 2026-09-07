package core

import (
	"context"
	"fmt"
	"strconv"

	"github.com/yumauri/fbrcm/core/firebase"
	rcdiff "github.com/yumauri/fbrcm/core/rc/diff"
)

// ParameterHistoryOptions controls how much retained Firebase history is
// inspected for one parameter.
type ParameterHistoryOptions struct {
	At    string
	Limit int
	All   bool
}

// RemoteConfigParameterLookupError identifies a parameter that does not occur
// anywhere in the inspected retained history.
type RemoteConfigParameterLookupError struct {
	ProjectID string
	Parameter string
	Err       error
}

func (e *RemoteConfigParameterLookupError) Error() string { return e.Err.Error() }
func (e *RemoteConfigParameterLookupError) Unwrap() error { return e.Err }

// ParameterHistoryChange is one publication that directly changed a
// parameter. PreviousVersion is the immediately preceding retained
// publication, including when unchanged publications are omitted from the
// returned change log.
type ParameterHistoryChange struct {
	PreviousVersion string
	Version         firebase.RemoteConfigVersion
	Change          rcdiff.ParameterChange
}

// ParameterHistoryBoundary describes the parameter state in the oldest
// retained publication after a complete scan.
type ParameterHistoryBoundary struct {
	Version string
	Present bool
	Group   string
}

// ParameterHistory is a newest-first parameter change log derived from
// adjacent retained Remote Config publications.
type ParameterHistory struct {
	Parameter           string
	AtVersion           string
	AtGroup             *string
	Changes             []ParameterHistoryChange
	ScannedVersionCount int
	HistoryExhausted    bool
	Boundary            *ParameterHistoryBoundary
}

// GetRemoteConfigParameterHistory scans retained Firebase history from the
// selected publication toward older publications. It compares every adjacent
// publication and returns only direct changes to parameter.
func (s *Core) GetRemoteConfigParameterHistory(ctx context.Context, projectID, parameter string, opts ParameterHistoryOptions) (ParameterHistory, error) {
	selector := opts.At
	if selector == "" {
		selector = "current"
	}
	if opts.Limit < 1 {
		return ParameterHistory{}, fmt.Errorf("parameter history limit must be greater than zero")
	}

	distance, relative, err := currentRelativeDistance(selector)
	if err != nil {
		return ParameterHistory{}, invalidVersionSelector(projectID, selector, err)
	}
	var number int64
	if !relative {
		number, err = strconv.ParseInt(selector, 10, 64)
		if err != nil || number <= 0 {
			return ParameterHistory{}, invalidVersionSelector(projectID, selector, fmt.Errorf("invalid Remote Config version %q", selector))
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
			return ParameterHistory{}, err
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
				return ParameterHistory{}, versionNotFound(projectID, selector, fmt.Errorf("project %s has no retained Remote Config versions", projectID))
			}
			if relative {
				return ParameterHistory{}, versionNotFound(projectID, selector, fmt.Errorf("project %s has no Remote Config version %d publications before current", projectID, distance))
			}
			return ParameterHistory{}, versionNotFound(projectID, selector, fmt.Errorf("remote config version %s was not found in retained history for project %s", selector, projectID))
		}
	}

	result := ParameterHistory{
		Parameter: parameter,
		AtVersion: versions[anchor].VersionNumber,
		Changes:   make([]ParameterHistoryChange, 0),
	}

	newer, err := s.GetRemoteConfigVersion(ctx, projectID, versions[anchor].VersionNumber, false)
	if err != nil {
		return ParameterHistory{}, err
	}
	result.ScannedVersionCount = 1
	if group, _, ok := remoteConfigParameter(newer.Config, parameter); ok {
		result.AtGroup = &group
	}
	seen := result.AtGroup != nil

	cursor := anchor
	for {
		if cursor+1 >= len(versions) {
			if historyExhausted {
				break
			}
			if err := loadNextPage(); err != nil {
				return ParameterHistory{}, err
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
			return ParameterHistory{}, getErr
		}
		result.ScannedVersionCount++

		change, found := remoteConfigParameterChange(newer.Config, older.Config, parameter)
		if found {
			seen = true
			if change.Kind != rcdiff.ChangeUnchanged {
				result.Changes = append(result.Changes, ParameterHistoryChange{
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
		group, _, present := remoteConfigParameter(newer.Config, parameter)
		result.Boundary = &ParameterHistoryBoundary{Version: versions[cursor].VersionNumber, Present: present, Group: group}
	}
	if !seen {
		return ParameterHistory{}, &RemoteConfigParameterLookupError{
			ProjectID: projectID,
			Parameter: parameter,
			Err:       fmt.Errorf("parameter %q was not found in retained Remote Config history for project %s", parameter, projectID),
		}
	}
	return result, nil
}

func remoteConfigParameterChange(newer, older *firebase.RemoteConfig, parameter string) (rcdiff.ParameterChange, bool) {
	result := rcdiff.CompareRemoteConfigs(older, newer)
	for _, change := range result.Parameters {
		if change.Key == parameter || change.PreviousKey == parameter {
			return change, true
		}
	}
	return rcdiff.ParameterChange{}, false
}

func remoteConfigParameter(cfg *firebase.RemoteConfig, parameter string) (string, firebase.RemoteConfigParam, bool) {
	if cfg == nil {
		return "", firebase.RemoteConfigParam{}, false
	}
	if value, ok := cfg.Parameters[parameter]; ok {
		return "", value, true
	}
	for groupName, group := range cfg.ParameterGroups {
		if value, ok := group.Parameters[parameter]; ok {
			return groupName, value, true
		}
	}
	return "", firebase.RemoteConfigParam{}, false
}
