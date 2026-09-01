package core

import (
	"context"
	"crypto/sha256"
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

type publicationPreHooksCompleteKey struct{}

// PublicationEnvironment identifies the execution policy and effective hook
// definition against which a publication candidate was prepared.
type PublicationEnvironment struct {
	Policy               string
	HooksEnabled         bool
	HookDefinitionSHA256 string
}

// PublicationEnvironmentForContext resolves the effective publication policy
// without executing hooks.
func PublicationEnvironmentForContext(ctx context.Context) (PublicationEnvironment, error) {
	policy := ExecutionPolicyFromContext(ctx)
	if !policy.RunHooks {
		return PublicationEnvironment{Policy: "stateless"}, nil
	}
	resolution, err := corehooks.Resolve()
	if err != nil {
		return PublicationEnvironment{}, err
	}
	enabled := len(resolution.Commands(corehooks.PrePublish)) > 0 || len(resolution.Commands(corehooks.PostPublish)) > 0
	environment := PublicationEnvironment{Policy: "stateful", HooksEnabled: enabled}
	if !enabled {
		return environment, nil
	}
	preSource, preLocal := resolution.Source(corehooks.PrePublish)
	postSource, postLocal := resolution.Source(corehooks.PostPublish)
	raw, err := json.Marshal(struct {
		Hooks      config.HooksConfig `json:"hooks"`
		PreSource  string             `json:"pre_source"`
		PreLocal   bool               `json:"pre_local"`
		PostSource string             `json:"post_source"`
		PostLocal  bool               `json:"post_local"`
	}{
		Hooks:      resolution.Hooks,
		PreSource:  preSource,
		PreLocal:   preLocal,
		PostSource: postSource,
		PostLocal:  postLocal,
	})
	if err != nil {
		return PublicationEnvironment{}, fmt.Errorf("encode publication hooks: %w", err)
	}
	digest := sha256.Sum256(raw)
	environment.HookDefinitionSHA256 = fmt.Sprintf("%x", digest[:])
	return environment, nil
}

// WithPublicationPreHooksComplete marks a context whose exact candidate has
// already passed its pre-publish hook gate. It is intended only for the commit
// stage of a plan apply operation.
func WithPublicationPreHooksComplete(ctx context.Context) context.Context {
	return context.WithValue(ctx, publicationPreHooksCompleteKey{}, true)
}

func publicationPreHooksComplete(ctx context.Context) bool {
	complete, _ := ctx.Value(publicationPreHooksCompleteKey{}).(bool)
	return complete
}

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

// ValidatePublicationCandidate performs Firebase validation and the trusted
// pre-publish hook gate for one exact base/candidate pair without publishing.
func (s *Core) ValidatePublicationCandidate(ctx context.Context, projectID string, current, candidate json.RawMessage, etag, operation string) error {
	ctx = corehooks.WithOperation(ctx, operation)
	if err := s.ValidateRemoteConfigWithETag(ctx, projectID, candidate, etag); err != nil {
		return err
	}
	changeNote, changeNoteSet := firebase.ChangeNoteFromContext(ctx)
	var updateRaw []byte
	var err error
	if changeNoteSet {
		updateRaw, err = firebase.PrepareRemoteConfigUpdate(candidate, changeNote)
	} else {
		updateRaw, err = firebase.PrepareRemoteConfigUpdate(candidate)
	}
	if err != nil {
		return &RemoteConfigValidationError{Source: ValidationSourceLocal, Err: fmt.Errorf("decode remote config: %w", err)}
	}
	hookSession, err := s.preparePublicationHooks(ctx, projectID, current, updateRaw)
	if err != nil {
		return err
	}
	if hookSession == nil {
		return nil
	}
	defer hookSession.Close()
	return hookSession.Run(ctx, corehooks.PrePublish, nil)
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

	policy := ExecutionPolicyFromContext(ctx)
	var currentRaw json.RawMessage
	if policy.ReadLocalState {
		if current, loadErr := config.LoadParametersCache(projectID); loadErr == nil && current.ETag == etag {
			currentRaw = current.RemoteConfig
		}
	}
	hookSession, err := s.preparePublicationHooks(ctx, projectID, currentRaw, updateRaw)
	if err != nil {
		return nil, "", err
	}
	if hookSession != nil {
		defer hookSession.Close()
		if !publicationPreHooksComplete(ctx) {
			if err := hookSession.Run(ctx, corehooks.PrePublish, nil); err != nil {
				return nil, "", err
			}
		}
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
	if policy.WriteLocalState {
		if err := config.SaveParametersCache(projectID, cache); err != nil {
			logger.Error("save parameters cache after publish failed", "project_id", projectID, "etag", nextETag, "err", err)
			cacheErr = err
		}
	} else {
		logger.Debug("execution policy skips parameters cache update after publish", "project_id", projectID, "etag", nextETag)
	}
	if hookSession != nil {
		if err := hookSession.Run(ctx, corehooks.PostPublish, updatedRaw); err != nil {
			return updatedRaw, nextETag, &RemoteConfigPublishedHookError{ProjectID: projectID, RemoteConfig: updatedRaw, ETag: nextETag, HookErr: err, CacheErr: cacheErr}
		}
	}
	if cacheErr != nil {
		return updatedRaw, nextETag, &RemoteConfigPublishedCacheError{ProjectID: projectID, RemoteConfig: updatedRaw, ETag: nextETag, Err: cacheErr}
	}

	return updatedRaw, nextETag, nil
}
