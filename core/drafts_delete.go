package core

import (
	"context"
	"encoding/json"

	"github.com/yumauri/fbrcm/core/draft"
)

func (s *Core) DeleteParameter(ctx context.Context, projectID, groupKey, paramKey string, publish bool) (*ParametersCache, *ParametersTree, bool, error) {
	return s.mutateDraft(ctx, projectID, publish, draft.MutationSpec{
		UnchangedErr: "parameter not found",
		Apply:        draft.DeleteParameter(groupKey, paramKey),
	})
}

func (s *Core) DeleteGroup(ctx context.Context, projectID, groupKey string, publish bool) (*ParametersCache, *ParametersTree, bool, error) {
	return s.mutateDraft(ctx, projectID, publish, withMutationLog(draft.MutationSpec{
		UnchangedErr: "group not changed",
		Apply:        draft.DeleteGroup(groupKey),
	}, "delete group", "project_id", projectID, "group", groupKey, "publish", publish))
}

func (s *Core) DeleteConditionalValue(ctx context.Context, projectID, groupKey, paramKey, valueLabel string, publish bool) (*ParametersCache, *ParametersTree, bool, error) {
	return s.mutateDraft(ctx, projectID, publish, withMutationLog(draft.MutationSpec{
		UnchangedErr: "conditional value not changed",
		Apply:        draft.DeleteConditionalValue(groupKey, paramKey, valueLabel),
	}, "delete conditional value", "project_id", projectID, "group", groupKey, "param", paramKey, "value_label", valueLabel, "publish", publish))
}

func (s *Core) PreviewDeleteParameter(projectID, groupKey, paramKey string) (*ParametersCache, json.RawMessage, error) {
	return s.previewDraft(projectID, draft.MutationSpec{
		UnchangedErr: "parameter not found",
		Apply:        draft.DeleteParameter(groupKey, paramKey),
	})
}

func (s *Core) PreviewDeleteGroup(projectID, groupKey string) (*ParametersCache, json.RawMessage, error) {
	return s.previewDraft(projectID, withMutationLog(draft.MutationSpec{
		UnchangedErr: "group not changed",
		Apply:        draft.DeleteGroup(groupKey),
	}, "preview delete group", "project_id", projectID, "group", groupKey))
}

func (s *Core) PreviewDeleteConditionalValue(projectID, groupKey, paramKey, valueLabel string) (*ParametersCache, json.RawMessage, error) {
	return s.previewDraft(projectID, withMutationLog(draft.MutationSpec{
		UnchangedErr: "conditional value not changed",
		Apply:        draft.DeleteConditionalValue(groupKey, paramKey, valueLabel),
	}, "preview delete conditional value", "project_id", projectID, "group", groupKey, "param", paramKey, "value_label", valueLabel))
}
