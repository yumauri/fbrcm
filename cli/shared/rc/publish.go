package rc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/yumauri/fbrcm/core/firebase"
)

// RemoteConfigValidationError identifies a project failure that occurred
// before a publish request was sent.
type RemoteConfigValidationError struct{ Err error }

func (e *RemoteConfigValidationError) Error() string { return e.Err.Error() }
func (e *RemoteConfigValidationError) Unwrap() error { return e.Err }

// IsValidationError reports whether an error came from Firebase validation.
func IsValidationError(err error) bool {
	var validationErr *RemoteConfigValidationError
	return errors.As(err, &validationErr)
}

// RemoteConfigPreparationError identifies a local candidate construction
// failure before Firebase validation or publication.
type RemoteConfigPreparationError struct{ Err error }

func (e *RemoteConfigPreparationError) Error() string { return e.Err.Error() }
func (e *RemoteConfigPreparationError) Unwrap() error { return e.Err }

// IsPreparationError reports whether a publish candidate could not be built.
func IsPreparationError(err error) bool {
	var preparationErr *RemoteConfigPreparationError
	return errors.As(err, &preparationErr)
}

// RemoteConfigPublisher validates and publishes Remote Config payloads.
type RemoteConfigPublisher interface {
	ValidateRemoteConfigWithETag(ctx context.Context, projectID string, raw json.RawMessage, etag string) error
	PublishRemoteConfigWithETag(ctx context.Context, projectID string, raw json.RawMessage, etag string) (json.RawMessage, string, error)
}

// RemoteConfigMutation applies a command-specific change to a cloned config.
type RemoteConfigMutation func(current *firebase.RemoteConfig) (changedCount int, finalCfg *firebase.RemoteConfig, err error)

// ValidateAndPublishRemoteConfig validates, publishes, and reports whether callers should retry.
func ValidateAndPublishRemoteConfig(ctx context.Context, publisher RemoteConfigPublisher, projectID string, raw json.RawMessage, etag, operation string, errOut io.Writer) (bool, error) {
	if err := publisher.ValidateRemoteConfigWithETag(ctx, projectID, raw, etag); err != nil {
		if IsRemoteConfigConflict(err) {
			writeRemoteConfigRetry(errOut, operation, projectID)
			return true, nil
		}
		return false, &RemoteConfigValidationError{Err: err}
	}
	if _, _, err := publisher.PublishRemoteConfigWithETag(ctx, projectID, raw, etag); err != nil {
		if IsRemoteConfigConflict(err) {
			writeRemoteConfigRetry(errOut, operation, projectID)
			return true, nil
		}
		return false, err
	}
	return false, nil
}

// PublishProjectConfigMutation applies mutation and publishes the result, returning whether callers should retry.
func PublishProjectConfigMutation(ctx context.Context, publisher RemoteConfigPublisher, projectCfg *ProjectConfig, operation string, errOut io.Writer, mutate RemoteConfigMutation) (int, bool, error) {
	if projectCfg == nil || projectCfg.Cache == nil {
		return 0, false, &RemoteConfigPreparationError{Err: fmt.Errorf("project config is incomplete")}
	}

	changedCount, finalCfg, err := mutate(projectCfg.Config)
	if err != nil {
		return 0, false, &RemoteConfigPreparationError{Err: err}
	}
	if changedCount == 0 {
		return 0, false, nil
	}

	finalRaw, err := firebase.MarshalRemoteConfig(finalCfg)
	if err != nil {
		return 0, false, &RemoteConfigPreparationError{Err: err}
	}
	retry, err := ValidateAndPublishRemoteConfig(ctx, publisher, projectCfg.Project.ProjectID, finalRaw, projectCfg.Cache.ETag, operation, errOut)
	if err != nil {
		return changedCount, false, err
	}
	return changedCount, retry, nil
}

func writeRemoteConfigRetry(out io.Writer, operation, projectID string) {
	if out == nil {
		return
	}
	_, _ = fmt.Fprintf(out, "remote config changed during %s; restarting project %s\n", operation, projectID)
}
