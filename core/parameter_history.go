package core

import (
	"context"
	"fmt"

	"github.com/yumauri/fbrcm/core/firebase"
	rcdiff "github.com/yumauri/fbrcm/core/rc/diff"
)

// ParameterHistoryOptions controls how much retained Firebase history is
// inspected for one parameter.
type ParameterHistoryOptions = RemoteConfigHistoryOptions

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
	scan, err := scanRemoteConfigHistory(ctx, s, projectID, opts,
		func(cfg *firebase.RemoteConfig) bool { _, _, ok := remoteConfigParameter(cfg, parameter); return ok },
		func(newer, older *firebase.RemoteConfig) (rcdiff.ParameterChange, rcdiff.ChangeKind, bool) {
			change, found := remoteConfigParameterChange(newer, older, parameter)
			return change, change.Kind, found
		},
	)
	if err != nil {
		return ParameterHistory{}, err
	}
	if !scan.Seen {
		return ParameterHistory{}, &RemoteConfigParameterLookupError{
			ProjectID: projectID, Parameter: parameter,
			Err: fmt.Errorf("parameter %q was not found in retained Remote Config history for project %s", parameter, projectID),
		}
	}
	result := ParameterHistory{Parameter: parameter, AtVersion: scan.AtVersion,
		ScannedVersionCount: scan.ScannedVersionCount, HistoryExhausted: scan.HistoryExhausted,
		Changes: make([]ParameterHistoryChange, 0, len(scan.Changes)),
	}
	if group, _, ok := remoteConfigParameter(scan.AtConfig, parameter); ok {
		result.AtGroup = &group
	}
	for _, change := range scan.Changes {
		result.Changes = append(result.Changes, ParameterHistoryChange(change))
	}
	if scan.Boundary != nil {
		group, _, _ := remoteConfigParameter(scan.BoundaryConfig, parameter)
		result.Boundary = &ParameterHistoryBoundary{Version: scan.Boundary.Version, Present: scan.Boundary.Present, Group: group}
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
