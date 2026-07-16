package core

import (
	"context"
	"encoding/json"

	"github.com/yumauri/fbrcm/core/conditions"
	"github.com/yumauri/fbrcm/core/draft"
	"github.com/yumauri/fbrcm/core/firebase"
)

func (s *Core) AddCondition(ctx context.Context, projectID string, definition ConditionDefinition, priority int, publish bool) (*ParametersCache, *ParametersTree, bool, error) {
	return s.mutateCondition(ctx, projectID, publish, "add condition", definition.Name, draft.MutationSpec{Apply: func(cfg *firebase.RemoteConfig) error {
		return conditions.Add(cfg, definition, priority)
	}})
}

func (s *Core) EditCondition(ctx context.Context, projectID, name string, edit ConditionEdit, publish bool) (*ParametersCache, *ParametersTree, bool, error) {
	return s.mutateCondition(ctx, projectID, publish, "edit condition", name, draft.MutationSpec{Apply: func(cfg *firebase.RemoteConfig) error {
		return conditions.EditDefinition(cfg, name, edit)
	}})
}

func (s *Core) EditConditionDetails(ctx context.Context, projectID string, edit ConditionDetailsEdit, publish bool) (*ParametersCache, *ParametersTree, bool, error) {
	return s.mutateCondition(ctx, projectID, publish, "edit condition details", edit.Name, draft.MutationSpec{Apply: func(cfg *firebase.RemoteConfig) error {
		return conditions.EditDetails(cfg, edit)
	}})
}

func (s *Core) RenameCondition(ctx context.Context, projectID, name, nextName string, publish bool) (*ParametersCache, *ParametersTree, bool, error) {
	return s.mutateCondition(ctx, projectID, publish, "rename condition", name, draft.MutationSpec{Apply: func(cfg *firebase.RemoteConfig) error {
		return conditions.Rename(cfg, name, nextName)
	}})
}

func (s *Core) MoveCondition(ctx context.Context, projectID, name string, priority int, publish bool) (*ParametersCache, *ParametersTree, bool, error) {
	return s.mutateCondition(ctx, projectID, publish, "move condition", name, draft.MutationSpec{Apply: func(cfg *firebase.RemoteConfig) error {
		return conditions.Move(cfg, name, priority)
	}})
}

func (s *Core) DeleteCondition(ctx context.Context, projectID, name string, publish bool) (*ParametersCache, *ParametersTree, bool, error) {
	return s.mutateCondition(ctx, projectID, publish, "delete condition", name, draft.MutationSpec{Apply: func(cfg *firebase.RemoteConfig) error {
		return conditions.Delete(cfg, name)
	}})
}

func (s *Core) PreviewAddCondition(projectID string, definition ConditionDefinition, priority int) (*ParametersCache, json.RawMessage, error) {
	return s.previewCondition(projectID, "preview add condition", definition.Name, draft.MutationSpec{Apply: func(cfg *firebase.RemoteConfig) error {
		return conditions.Add(cfg, definition, priority)
	}})
}

func (s *Core) PreviewEditCondition(projectID, name string, edit ConditionEdit) (*ParametersCache, json.RawMessage, error) {
	return s.previewCondition(projectID, "preview edit condition", name, draft.MutationSpec{Apply: func(cfg *firebase.RemoteConfig) error {
		return conditions.EditDefinition(cfg, name, edit)
	}})
}

func (s *Core) PreviewEditConditionDetails(projectID string, edit ConditionDetailsEdit) (*ParametersCache, json.RawMessage, error) {
	return s.previewCondition(projectID, "preview edit condition details", edit.Name, draft.MutationSpec{Apply: func(cfg *firebase.RemoteConfig) error {
		return conditions.EditDetails(cfg, edit)
	}})
}

func (s *Core) PreviewRenameCondition(projectID, name, nextName string) (*ParametersCache, json.RawMessage, error) {
	return s.previewCondition(projectID, "preview rename condition", name, draft.MutationSpec{Apply: func(cfg *firebase.RemoteConfig) error {
		return conditions.Rename(cfg, name, nextName)
	}})
}

func (s *Core) PreviewMoveCondition(projectID, name string, priority int) (*ParametersCache, json.RawMessage, error) {
	return s.previewCondition(projectID, "preview move condition", name, draft.MutationSpec{Apply: func(cfg *firebase.RemoteConfig) error {
		return conditions.Move(cfg, name, priority)
	}})
}

func (s *Core) PreviewDeleteCondition(projectID, name string) (*ParametersCache, json.RawMessage, error) {
	return s.previewCondition(projectID, "preview delete condition", name, draft.MutationSpec{Apply: func(cfg *firebase.RemoteConfig) error {
		return conditions.Delete(cfg, name)
	}})
}

func (s *Core) mutateCondition(ctx context.Context, projectID string, publish bool, action, name string, spec draft.MutationSpec) (*ParametersCache, *ParametersTree, bool, error) {
	return s.mutateDraft(ctx, projectID, publish, withMutationLog(spec, action, "project_id", projectID, "condition", name, "publish", publish))
}

func (s *Core) previewCondition(projectID, action, name string, spec draft.MutationSpec) (*ParametersCache, json.RawMessage, error) {
	return s.previewDraft(projectID, withMutationLog(spec, action, "project_id", projectID, "condition", name))
}
