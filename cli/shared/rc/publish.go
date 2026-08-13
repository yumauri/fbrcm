package rc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/yumauri/fbrcm/cli/progress"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	corehooks "github.com/yumauri/fbrcm/core/hooks"
	rcmutate "github.com/yumauri/fbrcm/core/rc/mutate"
)

// IsValidationError reports whether an error came from Firebase validation.
func IsValidationError(err error) bool {
	var validationErr *core.RemoteConfigValidationError
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

// RemoteMutationPlan retains selection provenance alongside the mutation.
type RemoteMutationPlan struct {
	Mutation         RemoteConfigMutation
	MatchedItemCount int
}

// RemoteConfigMutationPublishResult records the observable outcome of applying
// and attempting to publish one mutation.
type RemoteConfigMutationPublishResult struct {
	ChangedCount     int
	Retry            bool
	FailureStage     string
	Validated        bool
	ValidationSource string
	PublishedRaw     json.RawMessage
	PublishedVersion string
}

// ValidateAndPublishRemoteConfig validates, publishes, and reports whether callers should retry.
func ValidateAndPublishRemoteConfig(ctx context.Context, publisher RemoteConfigPublisher, projectID string, raw json.RawMessage, etag, operation string, errOut io.Writer) (bool, error) {
	result, err := validateAndPublishRemoteConfig(ctx, publisher, projectID, raw, etag, operation, errOut)
	return result.Retry, err
}

// ValidateAndPublishRemoteConfigResult returns validation provenance together
// with the publication outcome.
func ValidateAndPublishRemoteConfigResult(ctx context.Context, publisher RemoteConfigPublisher, projectID string, raw json.RawMessage, etag, operation string, errOut io.Writer) (RemoteConfigMutationPublishResult, error) {
	return validateAndPublishRemoteConfig(ctx, publisher, projectID, raw, etag, operation, errOut)
}

func validateAndPublishRemoteConfig(ctx context.Context, publisher RemoteConfigPublisher, projectID string, raw json.RawMessage, etag, operation string, errOut io.Writer) (RemoteConfigMutationPublishResult, error) {
	ctx = corehooks.WithOperation(ctx, operation)
	result := RemoteConfigMutationPublishResult{ValidationSource: core.ValidationSourceFirebase}
	progress.Start("Validating Remote Config for " + projectID + "…")
	if err := publisher.ValidateRemoteConfigWithETag(ctx, projectID, raw, etag); err != nil {
		if source, ok := core.RemoteConfigValidationSource(err); ok {
			result.ValidationSource = source
		}
		if IsRemoteConfigConflict(err) {
			writeRemoteConfigRetry(errOut, operation, projectID)
			result.Retry = true
			result.FailureStage = "validation"
			return result, nil
		}
		return result, &core.RemoteConfigValidationError{Source: result.ValidationSource, Err: err}
	}
	result.Validated = true
	if firebase.IsDryRun(ctx) {
		progress.Start("Previewing Remote Config for " + projectID + "…")
	} else {
		progress.Start("Publishing Remote Config for " + projectID + "…")
	}
	publishedRaw, _, err := publisher.PublishRemoteConfigWithETag(ctx, projectID, raw, etag)
	if err != nil {
		if IsRemoteConfigConflict(err) {
			writeRemoteConfigRetry(errOut, operation, projectID)
			result.Retry = true
			result.FailureStage = "publication"
			return result, nil
		}
		failureStage := ""
		var hookErr *corehooks.Error
		if errors.As(err, &hookErr) {
			failureStage = string(hookErr.Event) + "_hook"
		}
		result.PublishedRaw = publishedRaw
		result.FailureStage = failureStage
		return result, err
	}
	result.PublishedRaw = publishedRaw
	if !firebase.IsDryRun(ctx) {
		if published, parseErr := firebase.ParseRemoteConfig(publishedRaw); parseErr == nil {
			result.PublishedVersion = published.Version.VersionNumber
		}
	}
	return result, nil
}

// PublishProjectConfigMutation applies mutation and publishes the result, returning whether callers should retry.
func PublishProjectConfigMutation(ctx context.Context, publisher RemoteConfigPublisher, projectCfg *ProjectConfig, operation string, errOut io.Writer, mutate RemoteConfigMutation) (int, bool, error) {
	result, err := PublishProjectConfigMutationResult(ctx, publisher, projectCfg, operation, errOut, mutate)
	return result.ChangedCount, result.Retry, err
}

// PublishProjectConfigMutationResult applies one mutation and returns the
// machine-readable publication details used by batch commands.
func PublishProjectConfigMutationResult(ctx context.Context, publisher RemoteConfigPublisher, projectCfg *ProjectConfig, operation string, errOut io.Writer, mutate RemoteConfigMutation) (RemoteConfigMutationPublishResult, error) {
	localResult := RemoteConfigMutationPublishResult{ValidationSource: core.ValidationSourceLocal}
	if projectCfg == nil || projectCfg.Cache == nil {
		return localResult, &RemoteConfigPreparationError{Err: fmt.Errorf("project config is incomplete")}
	}

	changedCount, finalCfg, err := mutate(projectCfg.Config)
	localResult.ChangedCount = changedCount
	if err != nil {
		return localResult, &RemoteConfigPreparationError{Err: err}
	}
	if changedCount == 0 {
		localResult.Validated = true
		return localResult, nil
	}
	if err := rcmutate.EnsureOpaqueValuesUnchanged(projectCfg.Config, finalCfg); err != nil {
		return localResult, &RemoteConfigPreparationError{Err: err}
	}

	finalRaw, err := firebase.MarshalRemoteConfig(finalCfg)
	if err != nil {
		return localResult, &RemoteConfigPreparationError{Err: err}
	}
	result, err := validateAndPublishRemoteConfig(ctx, publisher, projectCfg.Project.ProjectID, finalRaw, projectCfg.Cache.ETag, operation, errOut)
	result.ChangedCount = changedCount
	if err != nil {
		var cacheErr *core.RemoteConfigPublishedCacheError
		if errors.As(err, &cacheErr) && len(result.PublishedRaw) == 0 {
			result.PublishedRaw = cacheErr.RemoteConfig
		}
		if !firebase.IsDryRun(ctx) && result.PublishedVersion == "" && len(result.PublishedRaw) > 0 {
			if published, parseErr := firebase.ParseRemoteConfig(result.PublishedRaw); parseErr == nil {
				result.PublishedVersion = published.Version.VersionNumber
			}
		}
		return result, err
	}
	return result, nil
}

func writeRemoteConfigRetry(out io.Writer, operation, projectID string) {
	if out == nil {
		return
	}
	_, _ = fmt.Fprintf(out, "remote config changed during %s; restarting project %s\n", operation, projectID)
}
