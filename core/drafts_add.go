package core

import (
	"context"
	"encoding/json"

	"github.com/yumauri/fbrcm/core/draft"
)

func (s *Core) DuplicateParameter(ctx context.Context, projectID, groupKey, paramKey string) (*ParametersCache, *ParametersTree, bool, string, error) {
	apply, nextKey := draft.DuplicateParameterAutoNamed(groupKey, paramKey)
	cache, tree, hasDraft, err := s.mutateDraft(ctx, projectID, false, draft.MutationSpec{Apply: apply})
	return cache, tree, hasDraft, nextKey(), err
}

func (s *Core) DuplicateParameterNamed(ctx context.Context, projectID, groupKey, paramKey, nextParamKey string, publish bool) (*ParametersCache, *ParametersTree, bool, error) {
	return s.mutateDraft(ctx, projectID, publish, withMutationLog(draft.MutationSpec{
		Apply: draft.DuplicateParameterNamed(groupKey, paramKey, nextParamKey),
	}, "duplicate parameter", "project_id", projectID, "group", groupKey, "param", paramKey, "next_param", nextParamKey, "publish", publish))
}

func (s *Core) PreviewDuplicateParameter(projectID, groupKey, paramKey, nextParamKey string) (*ParametersCache, json.RawMessage, error) {
	return s.previewDraft(projectID, withMutationLog(draft.MutationSpec{
		UnchangedErr: "parameter not changed",
		Apply:        draft.DuplicateParameterNamed(groupKey, paramKey, nextParamKey),
	}, "preview duplicate parameter", "project_id", projectID, "group", groupKey, "param", paramKey, "next_param", nextParamKey))
}
