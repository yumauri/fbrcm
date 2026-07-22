package core

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/yumauri/fbrcm/core/draft"
	"github.com/yumauri/fbrcm/core/firebase"
	corelog "github.com/yumauri/fbrcm/core/log"
)

type (
	ParameterDetailsEdit       = draft.ParameterDetailsEdit
	ParameterValueEdit         = draft.ParameterValueEdit
	DraftRecord                = draft.Record
	DraftPublishPlan           = draft.PublishPlan
	DraftPublishedCleanupError = draft.PublishedCleanupError
)

func NormalizeRemoteConfigGroupKey(groupKey string) string {
	return draft.NormalizeGroupKey(groupKey)
}

func MergeDraftWithLatest(baseRaw, draftRaw, latestRaw json.RawMessage) (json.RawMessage, bool, error) {
	return draft.MergeWithLatest(baseRaw, draftRaw, latestRaw)
}

func (s *Core) draftDeps() draft.Deps {
	return draft.Deps{
		GetParameters:                s.GetParameters,
		InspectParametersCache:       s.inspectParametersCacheOrError,
		ValidateRemoteConfigWithETag: s.ValidateRemoteConfigWithETag,
		PublishRemoteConfigWithETag:  s.PublishRemoteConfigWithETag,
	}
}

func (s *Core) inspectParametersCacheOrError(projectID string) (*ParametersCache, error) {
	cache, _, err := s.InspectParametersCache(projectID)
	if err != nil {
		return nil, err
	}
	return cache, nil
}

func (s *Core) mutateDraft(ctx context.Context, projectID string, publish bool, spec draft.MutationSpec) (*ParametersCache, *ParametersTree, bool, error) {
	result, hasDraftBefore, err := draft.Mutate(ctx, s.draftDeps(), projectID, publish, spec)
	if err != nil && result == nil {
		return nil, nil, hasDraftBefore, err
	}
	tree, treeErr := s.treeFromMutateResult(result)
	if err != nil {
		return result.Cache, tree, result.HasDraft, err
	}
	return result.Cache, tree, result.HasDraft, treeErr
}

func (s *Core) previewDraft(projectID string, spec draft.MutationSpec) (*ParametersCache, json.RawMessage, error) {
	return draft.Preview(s.draftDeps(), projectID, spec)
}

func (s *Core) treeFromMutateResult(result *draft.MutateResult) (*ParametersTree, error) {
	if result.Published {
		return s.BuildParametersTree(result.Cache)
	}
	return s.BuildParametersTreeFromRaw(result.FinalRaw, result.Cache.CachedAt, result.Cache.ETag)
}

func withMutationLog(spec draft.MutationSpec, action string, fields ...any) draft.MutationSpec {
	apply := spec.Apply
	spec.Apply = func(cfg *firebase.RemoteConfig) error {
		if err := apply(cfg); err != nil {
			args := append(fields, "err", err)
			corelog.For("core").Error(action+" failed", args...)
			return err
		}
		return nil
	}
	return spec
}

func (s *Core) ListDraftProjectIDs() ([]string, error) {
	return draft.ListProjectIDs()
}

func (s *Core) LoadDraft(projectID string) (json.RawMessage, bool, error) {
	return draft.Load(projectID)
}

func (s *Core) LoadDraftRecord(projectID string) (*draft.Record, bool, error) {
	return draft.LoadRecord(projectID)
}

func (s *Core) HasDraft(projectID string) (bool, error) {
	_, ok, err := draft.LoadRecord(projectID)
	return ok, err
}

func (s *Core) SaveDraft(projectID string, raw json.RawMessage) error {
	return draft.Save(projectID, raw)
}

func (s *Core) DeleteDraft(projectID string) error {
	return draft.Delete(projectID)
}

func (s *Core) PrepareDraftPublish(ctx context.Context, projectID string) (*DraftPublishPlan, error) {
	return draft.PreparePublish(ctx, s.draftDeps(), projectID)
}

func (s *Core) ExecuteDraftPublish(ctx context.Context, projectID string, plan *DraftPublishPlan) (*ParametersCache, *ParametersTree, error) {
	cache, _, publishErr := draft.ExecutePublish(ctx, s.draftDeps(), projectID, plan)
	if publishErr != nil && cache == nil {
		return nil, nil, publishErr
	}
	tree, treeErr := s.BuildParametersTree(cache)
	if treeErr != nil {
		return cache, nil, treeErr
	}
	return cache, tree, publishErr
}

func (s *Core) BuildParametersTreeFromRaw(raw json.RawMessage, cachedAt time.Time, etag string) (*ParametersTree, error) {
	return s.BuildParametersTree(&ParametersCache{
		ETag:         etag,
		CachedAt:     cachedAt,
		RemoteConfig: raw,
	})
}

func (s *Core) BuildConditionsTreeFromRaw(raw json.RawMessage, cachedAt time.Time, etag string) (*ConditionsTree, error) {
	return s.BuildConditionsTree(&ParametersCache{
		ETag:         etag,
		CachedAt:     cachedAt,
		RemoteConfig: raw,
	})
}

func (s *Core) BuildDraftAwareParametersTree(projectID string, cache *ParametersCache) (*ParametersTree, bool, error) {
	draftRaw, hasDraft, err := s.LoadDraft(projectID)
	if err != nil {
		return nil, false, err
	}
	if hasDraft {
		tree, err := s.BuildParametersTreeFromRaw(draftRaw, cache.CachedAt, cache.ETag)
		return tree, true, err
	}
	tree, err := s.BuildParametersTree(cache)
	return tree, false, err
}

func (s *Core) BuildDraftAwareConditionsTree(projectID string, cache *ParametersCache) (*ConditionsTree, bool, error) {
	if cache == nil {
		return nil, false, fmt.Errorf("parameters cache is nil")
	}
	draftRaw, hasDraft, err := s.LoadDraft(projectID)
	if err != nil {
		return nil, false, err
	}
	if hasDraft {
		tree, err := s.BuildConditionsTreeFromRaw(draftRaw, cache.CachedAt, cache.ETag)
		return tree, true, err
	}
	tree, err := s.BuildConditionsTree(cache)
	return tree, false, err
}
