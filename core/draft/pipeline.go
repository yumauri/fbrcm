package draft

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/firebase"
	corelog "github.com/yumauri/fbrcm/core/log"
)

// Deps supplies draft pipeline callbacks that require Core services.
type Deps struct {
	GetParameters                func(ctx context.Context, projectID string, force bool) (*config.ParametersCache, string, error)
	InspectParametersCache       func(projectID string) (*config.ParametersCache, error)
	ValidateRemoteConfigWithETag func(ctx context.Context, projectID string, raw json.RawMessage, etag string) error
	PublishRemoteConfigWithETag  func(ctx context.Context, projectID string, raw json.RawMessage, etag string) (json.RawMessage, string, error)
}

// MutateResult is the outcome of a draft mutation before tree building.
type MutateResult struct {
	Cache     *config.ParametersCache
	FinalRaw  json.RawMessage
	HasDraft  bool
	Published bool
}

type PublishPlan struct {
	Draft      *Record
	Latest     *config.ParametersCache
	Candidate  json.RawMessage
	HasChanges bool
	Rebased    bool
}

type PublishedCleanupError struct {
	Err error
}

func (e *PublishedCleanupError) Error() string {
	return fmt.Sprintf("draft was published but local cleanup failed: %v", e.Err)
}

func (e *PublishedCleanupError) Unwrap() error { return e.Err }

func Mutate(ctx context.Context, deps Deps, projectID string, publish bool, spec MutationSpec) (*MutateResult, bool, error) {
	cache, _, err := deps.GetParameters(ctx, projectID, false)
	if err != nil {
		return nil, false, err
	}

	currentRaw, hasDraft, err := currentDraftRaw(projectID, cache.RemoteConfig)
	if err != nil {
		return nil, false, err
	}

	finalRaw, err := BuildMutatedRemoteConfig(currentRaw, spec.UnchangedErr, spec.Apply)
	if err != nil {
		return nil, hasDraft, err
	}

	return writeMutationResult(ctx, deps, projectID, cache, finalRaw, hasDraft, publish)
}

func Preview(deps Deps, projectID string, spec MutationSpec) (*config.ParametersCache, json.RawMessage, error) {
	cache, err := deps.InspectParametersCache(projectID)
	if err != nil {
		return nil, nil, err
	}
	if cache == nil {
		return nil, nil, fmt.Errorf("parameters cache not found")
	}

	currentRaw, _, err := currentDraftRaw(projectID, cache.RemoteConfig)
	if err != nil {
		return nil, nil, err
	}

	finalRaw, err := BuildMutatedRemoteConfig(currentRaw, spec.UnchangedErr, spec.Apply)
	if err != nil {
		return nil, nil, err
	}
	return cache, finalRaw, nil
}

func writeMutationResult(ctx context.Context, deps Deps, projectID string, cache *config.ParametersCache, finalRaw json.RawMessage, hasDraft bool, publish bool) (*MutateResult, bool, error) {
	logger := corelog.For("core")
	if publish {
		validated := false
		if hasDraft {
			stored, ok, err := LoadRecord(projectID)
			if err != nil {
				return nil, hasDraft, err
			}
			if !ok {
				return nil, hasDraft, fmt.Errorf("draft not found")
			}
			latest, _, err := deps.GetParameters(ctx, projectID, true)
			if err != nil {
				return nil, hasDraft, err
			}
			mergedRaw, changed, err := MergeWithLatest(stored.BaseRemoteConfig, finalRaw, latest.RemoteConfig)
			if err != nil {
				return nil, hasDraft, err
			}
			if !changed {
				if err := Delete(projectID); err != nil {
					return nil, hasDraft, err
				}
				return &MutateResult{Cache: latest, FinalRaw: latest.RemoteConfig}, false, nil
			}
			if deps.ValidateRemoteConfigWithETag != nil {
				if err := deps.ValidateRemoteConfigWithETag(ctx, projectID, mergedRaw, latest.ETag); err != nil {
					return nil, hasDraft, err
				}
				validated = true
			}
			finalRaw, cache = mergedRaw, latest
		}
		if !validated && deps.ValidateRemoteConfigWithETag != nil {
			if err := deps.ValidateRemoteConfigWithETag(ctx, projectID, finalRaw, cache.ETag); err != nil {
				return nil, hasDraft, err
			}
		}
		updatedRaw, nextETag, err := deps.PublishRemoteConfigWithETag(ctx, projectID, finalRaw, cache.ETag)
		if err != nil {
			return nil, hasDraft, err
		}
		if err := Delete(projectID); err != nil {
			logger.Warn("remove draft after publish failed", "project_id", projectID, "err", err)
		}
		updatedCache := &config.ParametersCache{
			ETag:         nextETag,
			CachedAt:     time.Now().UTC(),
			RemoteConfig: updatedRaw,
		}
		return &MutateResult{
			Cache:     updatedCache,
			FinalRaw:  updatedRaw,
			HasDraft:  false,
			Published: true,
		}, false, nil
	}

	if err := SaveWithBase(projectID, cache, finalRaw); err != nil {
		return nil, hasDraft, err
	}
	return &MutateResult{
		Cache:    cache,
		FinalRaw: finalRaw,
		HasDraft: true,
	}, true, nil
}

// PublishExistingDraft publishes the on-disk draft for one project.
func PublishExistingDraft(ctx context.Context, deps Deps, projectID string) (*config.ParametersCache, json.RawMessage, error) {
	plan, err := PreparePublish(ctx, deps, projectID)
	if err != nil {
		return nil, nil, err
	}
	return ExecutePublish(ctx, deps, projectID, plan)
}

func PreparePublish(ctx context.Context, deps Deps, projectID string) (*PublishPlan, error) {
	stored, hasDraft, err := LoadRecord(projectID)
	if err != nil {
		return nil, err
	}
	if !hasDraft {
		return nil, fmt.Errorf("draft not found")
	}
	cache, _, err := deps.GetParameters(ctx, projectID, true)
	if err != nil {
		return nil, err
	}

	candidateRaw, hasChanges, err := MergeWithLatest(stored.BaseRemoteConfig, stored.RemoteConfig, cache.RemoteConfig)
	if err != nil {
		return nil, err
	}
	baseCfg, err := firebase.ParseRemoteConfig(stored.BaseRemoteConfig)
	if err != nil {
		return nil, err
	}
	latestCfg, err := firebase.ParseRemoteConfig(cache.RemoteConfig)
	if err != nil {
		return nil, err
	}
	if !hasChanges {
		candidateRaw = cache.RemoteConfig
	}
	return &PublishPlan{
		Draft:      stored,
		Latest:     cache,
		Candidate:  candidateRaw,
		HasChanges: hasChanges,
		Rebased:    baseCfg.Version.VersionNumber != latestCfg.Version.VersionNumber,
	}, nil
}

func ExecutePublish(ctx context.Context, deps Deps, projectID string, plan *PublishPlan) (*config.ParametersCache, json.RawMessage, error) {
	logger := corelog.For("core")
	if plan == nil || plan.Draft == nil || plan.Latest == nil {
		return nil, nil, fmt.Errorf("draft publish plan is incomplete")
	}
	current, ok, err := LoadRecord(projectID)
	if err != nil {
		return nil, nil, err
	}
	if !ok || !current.UpdatedAt.Equal(plan.Draft.UpdatedAt) {
		return nil, nil, fmt.Errorf("draft changed during preview; rerun the command")
	}
	if !plan.HasChanges {
		if !firebase.IsDryRun(ctx) {
			if err := Delete(projectID); err != nil {
				return nil, nil, err
			}
		}
		return plan.Latest, plan.Latest.RemoteConfig, nil
	}
	if deps.ValidateRemoteConfigWithETag != nil {
		if err := deps.ValidateRemoteConfigWithETag(ctx, projectID, plan.Candidate, plan.Latest.ETag); err != nil {
			return nil, nil, err
		}
	}
	if firebase.IsDryRun(ctx) {
		return &config.ParametersCache{ETag: plan.Latest.ETag, CachedAt: plan.Latest.CachedAt, RemoteConfig: plan.Candidate}, plan.Candidate, nil
	}

	updatedRaw, nextETag, err := deps.PublishRemoteConfigWithETag(ctx, projectID, plan.Candidate, plan.Latest.ETag)
	if err != nil {
		return nil, nil, err
	}
	updatedCache := &config.ParametersCache{
		ETag:         nextETag,
		CachedAt:     time.Now().UTC(),
		RemoteConfig: updatedRaw,
	}
	if err := Delete(projectID); err != nil {
		logger.Warn("remove draft after publish failed", "project_id", projectID, "err", err)
		return updatedCache, updatedRaw, &PublishedCleanupError{Err: err}
	}
	return updatedCache, updatedRaw, nil
}

// RefreshOutcome describes how a draft-aware refresh resolved.
type RefreshOutcome struct {
	Cache      *config.ParametersCache
	Source     string
	HasDraft   bool
	StaleDraft bool
	DraftRaw   json.RawMessage
	UseDraft   bool
}

func RefreshDraftAware(ctx context.Context, deps Deps, projectID string, previousCache *config.ParametersCache) (*RefreshOutcome, error) {
	logger := corelog.For("core")
	cache, source, err := deps.GetParameters(ctx, projectID, true)
	if err != nil {
		return nil, err
	}

	stored, hasDraft, err := LoadRecord(projectID)
	if err != nil {
		return nil, err
	}
	if !hasDraft {
		return &RefreshOutcome{
			Cache:    cache,
			Source:   source,
			HasDraft: false,
		}, nil
	}

	latestCfg, parseErr := firebase.ParseRemoteConfig(cache.RemoteConfig)
	if parseErr != nil {
		return nil, parseErr
	}
	if stored.BaseVersion == latestCfg.Version.VersionNumber {
		return &RefreshOutcome{
			Cache:    cache,
			Source:   "draft",
			HasDraft: true,
			DraftRaw: stored.RemoteConfig,
			UseDraft: true,
		}, nil
	}

	mergedRaw, hasChanges, err := MergeWithLatest(stored.BaseRemoteConfig, stored.RemoteConfig, cache.RemoteConfig)
	if err != nil {
		logger.Error("merge draft with latest failed", "project_id", projectID, "err", err)
		return &RefreshOutcome{
			Cache:      previousCache,
			Source:     "draft-stale",
			HasDraft:   true,
			StaleDraft: true,
			DraftRaw:   stored.RemoteConfig,
			UseDraft:   true,
		}, nil
	}
	if !hasChanges {
		if err := Delete(projectID); err != nil {
			logger.Warn("remove obsolete draft failed", "project_id", projectID, "err", err)
		}
		return &RefreshOutcome{
			Cache:    cache,
			Source:   source,
			HasDraft: false,
		}, nil
	}
	if err := SaveRebased(projectID, cache, mergedRaw, stored.CreatedAt); err != nil {
		return nil, err
	}
	return &RefreshOutcome{
		Cache:    cache,
		Source:   "draft",
		HasDraft: true,
		DraftRaw: mergedRaw,
		UseDraft: true,
	}, nil
}
