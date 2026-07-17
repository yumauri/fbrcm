package core

import (
	"context"
	"encoding/json"

	"github.com/yumauri/fbrcm/core/draft"
)

func (s *Core) RenameParameter(ctx context.Context, projectID, groupKey, paramKey, nextParamKey string, publish bool) (*ParametersCache, *ParametersTree, bool, error) {
	return s.mutateDraft(ctx, projectID, publish, withMutationLog(draft.MutationSpec{
		UnchangedErr: "parameter not changed",
		Apply:        draft.RenameParameter(groupKey, paramKey, nextParamKey),
	}, "rename parameter", "project_id", projectID, "group", groupKey, "param", paramKey, "next_param", nextParamKey, "publish", publish))
}

func (s *Core) RenameGroup(ctx context.Context, projectID, groupKey, nextGroupKey string, publish bool) (*ParametersCache, *ParametersTree, bool, error) {
	return s.mutateDraft(ctx, projectID, publish, withMutationLog(draft.MutationSpec{
		UnchangedErr: "group not changed",
		Apply:        draft.RenameGroup(groupKey, nextGroupKey),
	}, "rename group", "project_id", projectID, "group", groupKey, "next_group", nextGroupKey, "publish", publish))
}

func (s *Core) EditGroupDetails(ctx context.Context, projectID string, edit GroupDetailsEdit, publish bool) (*ParametersCache, *ParametersTree, bool, error) {
	return s.mutateDraft(ctx, projectID, publish, withMutationLog(draft.MutationSpec{
		UnchangedErr: "group not changed",
		Apply:        draft.EditGroupDetails(edit),
	}, "edit group details", "project_id", projectID, "group", edit.Name, "next_group", edit.NextName, "publish", publish))
}

func (s *Core) MoveParameter(ctx context.Context, projectID, groupKey, paramKey, nextGroupKey string, publish bool) (*ParametersCache, *ParametersTree, bool, error) {
	return s.mutateDraft(ctx, projectID, publish, withMutationLog(draft.MutationSpec{
		UnchangedErr: "parameter not changed",
		Apply:        draft.MoveParameter(groupKey, paramKey, nextGroupKey),
	}, "move parameter", "project_id", projectID, "group", groupKey, "param", paramKey, "next_group", nextGroupKey, "publish", publish))
}

func (s *Core) EditParameterDetails(ctx context.Context, projectID string, edit ParameterDetailsEdit, publish bool) (*ParametersCache, *ParametersTree, bool, error) {
	return s.mutateDraft(ctx, projectID, publish, withMutationLog(draft.MutationSpec{
		UnchangedErr: "parameter not changed",
		Apply:        draft.EditParameterDetails(edit),
	}, "edit parameter details", "project_id", projectID, "group", edit.GroupKey, "param", edit.ParamKey, "next_group", edit.NextGroupKey, "next_param", edit.NextParamKey, "next_type", edit.NextValueType, "publish", publish))
}

func (s *Core) MoveGroup(ctx context.Context, projectID, groupKey, nextGroupKey string, publish bool) (*ParametersCache, *ParametersTree, bool, error) {
	return s.mutateDraft(ctx, projectID, publish, withMutationLog(draft.MutationSpec{
		UnchangedErr: "group not changed",
		Apply:        draft.MoveGroup(groupKey, nextGroupKey),
	}, "move group", "project_id", projectID, "group", groupKey, "next_group", nextGroupKey, "publish", publish))
}

func (s *Core) SetBooleanParameterValue(ctx context.Context, projectID, groupKey, paramKey, valueLabel string, nextValue, publish bool) (*ParametersCache, *ParametersTree, bool, error) {
	return s.mutateDraft(ctx, projectID, publish, withMutationLog(draft.MutationSpec{
		UnchangedErr: "parameter value not changed",
		Apply:        draft.SetBooleanParameterValue(groupKey, paramKey, valueLabel, nextValue),
	}, "set boolean parameter value", "project_id", projectID, "group", groupKey, "param", paramKey, "value_label", valueLabel, "next_value", nextValue, "publish", publish))
}

func (s *Core) SetNumberParameterValue(ctx context.Context, projectID, groupKey, paramKey, valueLabel, nextValue string, publish bool) (*ParametersCache, *ParametersTree, bool, error) {
	return s.mutateDraft(ctx, projectID, publish, withMutationLog(draft.MutationSpec{
		UnchangedErr: "parameter value not changed",
		Apply:        draft.SetNumberParameterValue(groupKey, paramKey, valueLabel, nextValue),
	}, "set number parameter value", "project_id", projectID, "group", groupKey, "param", paramKey, "value_label", valueLabel, "next_value", nextValue, "publish", publish))
}

func (s *Core) SetStringParameterValue(ctx context.Context, projectID, groupKey, paramKey, valueLabel, nextValue string, publish bool) (*ParametersCache, *ParametersTree, bool, error) {
	return s.mutateDraft(ctx, projectID, publish, withMutationLog(draft.MutationSpec{
		UnchangedErr: "parameter value not changed",
		Apply:        draft.SetStringParameterValue(groupKey, paramKey, valueLabel, nextValue),
	}, "set string parameter value", "project_id", projectID, "group", groupKey, "param", paramKey, "value_label", valueLabel, "publish", publish))
}

func (s *Core) SetJSONParameterValue(ctx context.Context, projectID, groupKey, paramKey, valueLabel, nextValue string, publish bool) (*ParametersCache, *ParametersTree, bool, error) {
	return s.mutateDraft(ctx, projectID, publish, withMutationLog(draft.MutationSpec{
		UnchangedErr: "parameter value not changed",
		Apply:        draft.SetJSONParameterValue(groupKey, paramKey, valueLabel, nextValue),
	}, "set json parameter value", "project_id", projectID, "group", groupKey, "param", paramKey, "value_label", valueLabel, "publish", publish))
}

func (s *Core) PreviewRenameParameter(projectID, groupKey, paramKey, nextParamKey string) (*ParametersCache, json.RawMessage, error) {
	return s.previewDraft(projectID, withMutationLog(draft.MutationSpec{
		UnchangedErr: "parameter not changed",
		Apply:        draft.RenameParameter(groupKey, paramKey, nextParamKey),
	}, "preview rename parameter", "project_id", projectID, "group", groupKey, "param", paramKey, "next_param", nextParamKey))
}

func (s *Core) PreviewRenameGroup(projectID, groupKey, nextGroupKey string) (*ParametersCache, json.RawMessage, error) {
	return s.previewDraft(projectID, withMutationLog(draft.MutationSpec{
		UnchangedErr: "group not changed",
		Apply:        draft.RenameGroup(groupKey, nextGroupKey),
	}, "preview rename group", "project_id", projectID, "group", groupKey, "next_group", nextGroupKey))
}

func (s *Core) PreviewEditGroupDetails(projectID string, edit GroupDetailsEdit) (*ParametersCache, json.RawMessage, error) {
	return s.previewDraft(projectID, withMutationLog(draft.MutationSpec{
		UnchangedErr: "group not changed",
		Apply:        draft.EditGroupDetails(edit),
	}, "preview edit group details", "project_id", projectID, "group", edit.Name, "next_group", edit.NextName))
}

func (s *Core) PreviewMoveParameter(projectID, groupKey, paramKey, nextGroupKey string) (*ParametersCache, json.RawMessage, error) {
	return s.previewDraft(projectID, withMutationLog(draft.MutationSpec{
		UnchangedErr: "parameter not changed",
		Apply:        draft.MoveParameter(groupKey, paramKey, nextGroupKey),
	}, "preview move parameter", "project_id", projectID, "group", groupKey, "param", paramKey, "next_group", nextGroupKey))
}

func (s *Core) PreviewEditParameterDetails(projectID string, edit ParameterDetailsEdit) (*ParametersCache, json.RawMessage, error) {
	return s.previewDraft(projectID, withMutationLog(draft.MutationSpec{
		Apply: draft.EditParameterDetails(edit),
	}, "preview edit parameter details", "project_id", projectID, "group", edit.GroupKey, "param", edit.ParamKey, "next_group", edit.NextGroupKey, "next_param", edit.NextParamKey, "next_type", edit.NextValueType))
}

func (s *Core) PreviewMoveGroup(projectID, groupKey, nextGroupKey string) (*ParametersCache, json.RawMessage, error) {
	return s.previewDraft(projectID, withMutationLog(draft.MutationSpec{
		UnchangedErr: "group not changed",
		Apply:        draft.MoveGroup(groupKey, nextGroupKey),
	}, "preview move group", "project_id", projectID, "group", groupKey, "next_group", nextGroupKey))
}

func (s *Core) PreviewSetBooleanParameterValue(projectID, groupKey, paramKey, valueLabel string, nextValue bool) (*ParametersCache, json.RawMessage, error) {
	return s.previewDraft(projectID, withMutationLog(draft.MutationSpec{
		UnchangedErr: "parameter value not changed",
		Apply:        draft.SetBooleanParameterValue(groupKey, paramKey, valueLabel, nextValue),
	}, "preview set boolean parameter value", "project_id", projectID, "group", groupKey, "param", paramKey, "value_label", valueLabel, "next_value", nextValue))
}

func (s *Core) PreviewSetNumberParameterValue(projectID, groupKey, paramKey, valueLabel, nextValue string) (*ParametersCache, json.RawMessage, error) {
	return s.previewDraft(projectID, withMutationLog(draft.MutationSpec{
		UnchangedErr: "parameter value not changed",
		Apply:        draft.SetNumberParameterValue(groupKey, paramKey, valueLabel, nextValue),
	}, "preview set number parameter value", "project_id", projectID, "group", groupKey, "param", paramKey, "value_label", valueLabel, "next_value", nextValue))
}

func (s *Core) PreviewSetStringParameterValue(projectID, groupKey, paramKey, valueLabel, nextValue string) (*ParametersCache, json.RawMessage, error) {
	return s.previewDraft(projectID, withMutationLog(draft.MutationSpec{
		UnchangedErr: "parameter value not changed",
		Apply:        draft.SetStringParameterValue(groupKey, paramKey, valueLabel, nextValue),
	}, "preview set string parameter value", "project_id", projectID, "group", groupKey, "param", paramKey, "value_label", valueLabel))
}

func (s *Core) PreviewSetJSONParameterValue(projectID, groupKey, paramKey, valueLabel, nextValue string) (*ParametersCache, json.RawMessage, error) {
	return s.previewDraft(projectID, withMutationLog(draft.MutationSpec{
		UnchangedErr: "parameter value not changed",
		Apply:        draft.SetJSONParameterValue(groupKey, paramKey, valueLabel, nextValue),
	}, "preview set json parameter value", "project_id", projectID, "group", groupKey, "param", paramKey, "value_label", valueLabel))
}
