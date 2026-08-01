package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/firebase"
	corehooks "github.com/yumauri/fbrcm/core/hooks"
	corelog "github.com/yumauri/fbrcm/core/log"
)

const (
	ValidationSourceLocal    = "local"
	ValidationSourceFirebase = "firebase"
)

// RemoteConfigValidationError identifies whether validation failed before or
// after the candidate reached Firebase.
type RemoteConfigValidationError struct {
	Source string
	Err    error
}

func (e *RemoteConfigValidationError) Error() string { return e.Err.Error() }
func (e *RemoteConfigValidationError) Unwrap() error { return e.Err }

// RemoteConfigValidationSource returns validation failure provenance.
func RemoteConfigValidationSource(err error) (string, bool) {
	var validationErr *RemoteConfigValidationError
	if !errors.As(err, &validationErr) {
		return "", false
	}
	return validationErr.Source, true
}

// RemoteConfigPublishedCacheError reports that Firebase accepted a Remote
// Config publish, but the returned config could not be persisted locally.
// RemoteConfig and ETag describe the successfully published remote state.
type RemoteConfigPublishedCacheError struct {
	ProjectID    string
	RemoteConfig json.RawMessage
	ETag         string
	Err          error
}

func (e *RemoteConfigPublishedCacheError) Error() string {
	return fmt.Sprintf("remote config was published for %s but the local cache update failed: %v", e.ProjectID, e.Err)
}

func (e *RemoteConfigPublishedCacheError) Unwrap() error { return e.Err }

// RemoteConfigPublishedHookError reports that Firebase accepted a publish but
// a post_publish hook failed. RemoteConfig and ETag identify the committed
// state; CacheErr is also populated when the local cache update failed.
type RemoteConfigPublishedHookError struct {
	ProjectID    string
	RemoteConfig json.RawMessage
	ETag         string
	HookErr      error
	CacheErr     error
}

func (e *RemoteConfigPublishedHookError) Error() string {
	if e.CacheErr != nil {
		return fmt.Sprintf("remote config was published for %s but the local cache update and post_publish hook failed: %v; %v", e.ProjectID, e.CacheErr, e.HookErr)
	}
	return fmt.Sprintf("remote config was published for %s but a post_publish hook failed: %v", e.ProjectID, e.HookErr)
}

func (e *RemoteConfigPublishedHookError) Unwrap() []error {
	errors := []error{e.HookErr}
	if e.CacheErr != nil {
		errors = append(errors, &RemoteConfigPublishedCacheError{ProjectID: e.ProjectID, RemoteConfig: e.RemoteConfig, ETag: e.ETag, Err: e.CacheErr})
	}
	return errors
}

func (s *Core) ExportRemoteConfig(ctx context.Context, projectID string) (json.RawMessage, string, error) {
	logger := corelog.For("core")
	logger.Info("export remote config requested", "project_id", projectID)

	fb, err := s.firebaseServiceForProject(ctx, projectID)
	if err != nil {
		return nil, "", err
	}

	raw, etag, err := fb.GetRemoteConfig(ctx, projectID)
	if err != nil {
		logger.Error("firebase remote config export failed", "project_id", projectID, "err", err)
		return nil, "", fmt.Errorf("firebase error: %w", err)
	}

	return raw, etag, nil
}

// DownloadRemoteConfigDefaults downloads application defaults for a project.
func (s *Core) DownloadRemoteConfigDefaults(ctx context.Context, projectID string, format firebase.DefaultsFormat) ([]byte, error) {
	logger := corelog.For("core")
	logger.Info("download remote config defaults requested", "project_id", projectID, "format", format)

	fb, err := s.firebaseServiceForProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	defaults, err := fb.DownloadRemoteConfigDefaults(ctx, projectID, format)
	if err != nil {
		logger.Error("firebase remote config defaults download failed", "project_id", projectID, "format", format, "err", err)
		return nil, fmt.Errorf("firebase error: %w", err)
	}
	return defaults, nil
}

func (s *Core) ValidateRemoteConfigWithETag(ctx context.Context, projectID string, raw json.RawMessage, etag string) error {
	logger := corelog.For("core")
	logger.Info("validate remote config with etag requested", "project_id", projectID, "etag", etag)

	fb, err := s.firebaseServiceForProject(ctx, projectID)
	if err != nil {
		return err
	}

	changeNote, changeNoteSet := firebase.ChangeNoteFromContext(ctx)
	var updateRaw []byte
	if changeNoteSet {
		updateRaw, err = firebase.PrepareRemoteConfigUpdate(raw, changeNote)
	} else {
		updateRaw, err = firebase.PrepareRemoteConfigUpdate(raw)
	}
	if err != nil {
		logger.Error("remote config validation payload decode failed", "project_id", projectID, "err", err)
		return &RemoteConfigValidationError{Source: ValidationSourceLocal, Err: fmt.Errorf("decode remote config: %w", err)}
	}

	if err := fb.ValidateRemoteConfig(ctx, projectID, updateRaw, etag); err != nil {
		logger.Error("firebase remote config validation failed", "project_id", projectID, "err", err)
		return &RemoteConfigValidationError{Source: ValidationSourceFirebase, Err: fmt.Errorf("firebase error: %w", err)}
	}

	return nil
}

func (s *Core) PublishRemoteConfigWithETag(ctx context.Context, projectID string, raw json.RawMessage, etag string) (json.RawMessage, string, error) {
	logger := corelog.For("core")
	logger.Info("publish remote config with etag requested", "project_id", projectID, "etag", etag)

	fb, err := s.firebaseServiceForProject(ctx, projectID)
	if err != nil {
		return nil, "", err
	}

	changeNote, changeNoteSet := firebase.ChangeNoteFromContext(ctx)
	var updateRaw []byte
	if changeNoteSet {
		updateRaw, err = firebase.PrepareRemoteConfigUpdate(raw, changeNote)
	} else {
		updateRaw, err = firebase.PrepareRemoteConfigUpdate(raw)
	}
	if err != nil {
		logger.Error("remote config publish payload decode failed", "project_id", projectID, "err", err)
		return nil, "", fmt.Errorf("decode remote config: %w", err)
	}

	var currentRaw json.RawMessage
	if current, loadErr := config.LoadParametersCache(projectID); loadErr == nil && current.ETag == etag {
		currentRaw = current.RemoteConfig
	}
	hookSession, err := s.preparePublicationHooks(ctx, projectID, currentRaw, updateRaw)
	if err != nil {
		return nil, "", err
	}
	defer hookSession.Close()
	if err := hookSession.Run(ctx, corehooks.PrePublish, nil); err != nil {
		return nil, "", err
	}

	updatedRaw, nextETag, err := fb.UpdateRemoteConfig(ctx, projectID, updateRaw, etag)
	if err != nil {
		logger.Error("firebase remote config publish failed", "project_id", projectID, "err", err)
		return nil, "", fmt.Errorf("firebase error: %w", err)
	}

	cache := &config.ParametersCache{
		ETag:         nextETag,
		CachedAt:     time.Now().UTC(),
		RemoteConfig: updatedRaw,
	}
	if firebase.IsDryRun(ctx) {
		logger.Warn("dry run, skip parameters cache update after publish", "project_id", projectID, "etag", nextETag)
		return updatedRaw, nextETag, nil
	}
	var cacheErr error
	if err := config.SaveParametersCache(projectID, cache); err != nil {
		logger.Error("save parameters cache after publish failed", "project_id", projectID, "etag", nextETag, "err", err)
		cacheErr = err
	}
	if err := hookSession.Run(ctx, corehooks.PostPublish, updatedRaw); err != nil {
		return updatedRaw, nextETag, &RemoteConfigPublishedHookError{ProjectID: projectID, RemoteConfig: updatedRaw, ETag: nextETag, HookErr: err, CacheErr: cacheErr}
	}
	if cacheErr != nil {
		return updatedRaw, nextETag, &RemoteConfigPublishedCacheError{ProjectID: projectID, RemoteConfig: updatedRaw, ETag: nextETag, Err: cacheErr}
	}

	return updatedRaw, nextETag, nil
}
