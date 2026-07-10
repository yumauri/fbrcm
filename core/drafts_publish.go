package core

import (
	"context"
	"fmt"

	"github.com/yumauri/fbrcm/core/draft"
)

func (s *Core) PublishDraft(ctx context.Context, projectID string) (*ParametersCache, *ParametersTree, error) {
	cache, _, err := draft.PublishExistingDraft(ctx, s.draftDeps(), projectID)
	if err != nil {
		return nil, nil, err
	}
	tree, err := s.BuildParametersTree(cache)
	return cache, tree, err
}

func (s *Core) DiscardDraft(ctx context.Context, projectID string) (*ParametersCache, *ParametersTree, error) {
	cache, _, err := s.GetParameters(ctx, projectID, false)
	if err != nil {
		return nil, nil, err
	}
	if err := draft.Delete(projectID); err != nil {
		return nil, nil, err
	}
	tree, err := s.BuildParametersTree(cache)
	return cache, tree, err
}

func (s *Core) RefreshDraftAwareParameters(ctx context.Context, projectID string, previousCache *ParametersCache) (*ParametersCache, *ParametersTree, string, bool, bool, error) {
	if previousCache == nil {
		var err error
		previousCache, _, err = s.InspectParametersCache(projectID)
		if err != nil {
			return nil, nil, "", false, false, fmt.Errorf("inspect parameters cache: %w", err)
		}
	}

	outcome, err := draft.RefreshDraftAware(ctx, s.draftDeps(), projectID, previousCache)
	if err != nil {
		return nil, nil, "", false, false, err
	}

	if !outcome.HasDraft {
		tree, err := s.BuildParametersTree(outcome.Cache)
		return outcome.Cache, tree, outcome.Source, false, false, err
	}

	var tree *ParametersTree
	var treeErr error
	if outcome.UseDraft {
		tree, treeErr = s.BuildParametersTreeFromRaw(outcome.DraftRaw, outcome.Cache.CachedAt, outcome.Cache.ETag)
	} else {
		tree, treeErr = s.BuildParametersTree(outcome.Cache)
	}
	return outcome.Cache, tree, outcome.Source, outcome.HasDraft, outcome.StaleDraft, treeErr
}
