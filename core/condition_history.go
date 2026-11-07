package core

import (
	"context"
	"fmt"

	"github.com/yumauri/fbrcm/core/firebase"
	rcdiff "github.com/yumauri/fbrcm/core/rc/diff"
)

// ConditionHistoryOptions controls retained history inspection for a condition.
type ConditionHistoryOptions = RemoteConfigHistoryOptions

// RemoteConfigConditionLookupError identifies an exact condition name absent
// from every inspected retained publication.
type RemoteConfigConditionLookupError struct {
	ProjectID string
	Condition string
	Err       error
}

func (e *RemoteConfigConditionLookupError) Error() string { return e.Err.Error() }
func (e *RemoteConfigConditionLookupError) Unwrap() error { return e.Err }

// ConditionHistoryChange attributes a condition diff to the publication that
// introduced it, compared with the immediately preceding retained publication.
type ConditionHistoryChange struct {
	PreviousVersion string
	Version         firebase.RemoteConfigVersion
	Change          rcdiff.ConditionChange
}

// ConditionHistoryBoundary describes the condition at the retention boundary.
type ConditionHistoryBoundary struct {
	Version string
	Present bool
}

// ConditionHistory is a newest-first log for one exact condition name.
type ConditionHistory struct {
	Condition           string
	AtVersion           string
	Changes             []ConditionHistoryChange
	ScannedVersionCount int
	HistoryExhausted    bool
	Boundary            *ConditionHistoryBoundary
}

// GetRemoteConfigConditionHistory traces definition and evaluation-position
// changes. Parameter usages are independent of the condition definition.
// Renames appear as removal/addition because condition identity is its name.
func (s *Core) GetRemoteConfigConditionHistory(ctx context.Context, projectID, condition string, opts ConditionHistoryOptions) (ConditionHistory, error) {
	scan, err := scanRemoteConfigHistory(ctx, s, projectID, opts,
		func(cfg *firebase.RemoteConfig) bool {
			_, ok := firebase.ConditionIndex(cfg, condition)
			return ok
		},
		func(newer, older *firebase.RemoteConfig) (rcdiff.ConditionChange, rcdiff.ChangeKind, bool) {
			for _, change := range rcdiff.CompareRemoteConfigs(older, newer).Conditions {
				if change.Name == condition {
					return change, change.Kind, true
				}
			}
			return rcdiff.ConditionChange{}, "", false
		},
	)
	if err != nil {
		return ConditionHistory{}, err
	}
	if !scan.Seen {
		return ConditionHistory{}, &RemoteConfigConditionLookupError{
			ProjectID: projectID, Condition: condition,
			Err: fmt.Errorf("condition %q was not found in retained Remote Config history for project %s", condition, projectID),
		}
	}
	result := ConditionHistory{
		Condition: condition, AtVersion: scan.AtVersion,
		ScannedVersionCount: scan.ScannedVersionCount, HistoryExhausted: scan.HistoryExhausted,
		Changes: make([]ConditionHistoryChange, 0, len(scan.Changes)),
	}
	for _, change := range scan.Changes {
		result.Changes = append(result.Changes, ConditionHistoryChange(change))
	}
	if scan.Boundary != nil {
		result.Boundary = &ConditionHistoryBoundary{Version: scan.Boundary.Version, Present: scan.Boundary.Present}
	}
	return result, nil
}
