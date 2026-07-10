package draft

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/yumauri/fbrcm/core/config"
	corelog "github.com/yumauri/fbrcm/core/log"
)

// Deps supplies draft pipeline callbacks that require Core services.
type Deps struct {
	GetParameters               func(ctx context.Context, projectID string, force bool) (*config.ParametersCache, string, error)
	InspectParametersCache      func(projectID string) (*config.ParametersCache, error)
	PublishRemoteConfigWithETag func(ctx context.Context, projectID string, raw json.RawMessage, etag string) (json.RawMessage, string, error)
}

// MutateResult is the outcome of a draft mutation before tree building.
type MutateResult struct {
	Cache     *config.ParametersCache
	FinalRaw  json.RawMessage
	HasDraft  bool
	Published bool
}

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

	if err := Save(projectID, finalRaw); err != nil {
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
	logger := corelog.For("core")
	cache, _, err := deps.GetParameters(ctx, projectID, false)
	if err != nil {
		return nil, nil, err
	}
	draftRaw, hasDraft, err := Load(projectID)
	if err != nil {
		return nil, nil, err
	}
	if !hasDraft {
		return nil, nil, fmt.Errorf("draft not found")
	}

	updatedRaw, nextETag, err := deps.PublishRemoteConfigWithETag(ctx, projectID, draftRaw, cache.ETag)
	if err != nil {
		return nil, nil, err
	}
	if err := Delete(projectID); err != nil {
		logger.Warn("remove draft after publish failed", "project_id", projectID, "err", err)
	}

	updatedCache := &config.ParametersCache{
		ETag:         nextETag,
		CachedAt:     time.Now().UTC(),
		RemoteConfig: updatedRaw,
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

	draftRaw, hasDraft, err := Load(projectID)
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

	if previousCache == nil || bytes.Equal(previousCache.RemoteConfig, cache.RemoteConfig) {
		return &RefreshOutcome{
			Cache:    cache,
			Source:   "draft",
			HasDraft: true,
			DraftRaw: draftRaw,
			UseDraft: true,
		}, nil
	}

	mergedRaw, hasChanges, err := MergeWithLatest(previousCache.RemoteConfig, draftRaw, cache.RemoteConfig)
	if err != nil {
		logger.Error("merge draft with latest failed", "project_id", projectID, "err", err)
		return &RefreshOutcome{
			Cache:      previousCache,
			Source:     "draft-stale",
			HasDraft:   true,
			StaleDraft: true,
			DraftRaw:   draftRaw,
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
	if err := Save(projectID, mergedRaw); err != nil {
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
