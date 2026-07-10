package app

import (
	"context"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/tui/messages"
)

type draftMutationFunc func(context.Context) (*core.ParametersCache, *core.ParametersTree, bool, error)

func (m Model) draftMutationCmd(project core.Project, publish bool, selectGroupKey, selectParamKey string, closeDetails bool, run draftMutationFunc) tea.Cmd {
	return func() tea.Msg {
		_, stale := m.parameters.ProjectDraftState(project.ProjectID)
		_, tree, hasDraft, err := run(context.Background())
		if err != nil {
			return messages.ParametersLoadedMsg{Project: project, Err: err, HasDraft: m.parameters.HasDraft(project.ProjectID), StaleDraft: stale, CloseDetails: closeDetails}
		}
		source := "draft"
		if publish {
			source = "firebase"
		}
		return messages.ParametersLoadedMsg{
			Project:        project,
			Tree:           tree,
			Source:         source,
			CacheSource:    "cache",
			Err:            nil,
			CloseDetails:   closeDetails,
			HasDraft:       hasDraft,
			StaleDraft:     !publish && hasDraft && stale,
			Revalidate:     false,
			SelectGroupKey: selectGroupKey,
			SelectParamKey: selectParamKey,
		}
	}
}

func (m Model) deleteParameterCmd(project core.Project, groupKey, paramKey string, publish bool, closeDetails bool) tea.Cmd {
	return m.draftMutationCmd(project, publish, "", "", closeDetails, func(ctx context.Context) (*core.ParametersCache, *core.ParametersTree, bool, error) {
		return m.svc.DeleteParameter(ctx, project.ProjectID, groupKey, paramKey, publish)
	})
}

func (m Model) deleteGroupCmd(project core.Project, groupKey string, publish bool) tea.Cmd {
	return m.draftMutationCmd(project, publish, "", "", false, func(ctx context.Context) (*core.ParametersCache, *core.ParametersTree, bool, error) {
		return m.svc.DeleteGroup(ctx, project.ProjectID, groupKey, publish)
	})
}

// deleteConditionalValueCmd removes one conditional value.
func (m Model) deleteConditionalValueCmd(project core.Project, groupKey, paramKey, valueLabel string, publish bool) tea.Cmd {
	return m.draftMutationCmd(project, publish, groupKey, paramKey, false, func(ctx context.Context) (*core.ParametersCache, *core.ParametersTree, bool, error) {
		return m.svc.DeleteConditionalValue(ctx, project.ProjectID, groupKey, paramKey, valueLabel, publish)
	})
}

func (m Model) publishDraftCmd(project core.Project) tea.Cmd {
	return func() tea.Msg {
		_, tree, err := m.svc.PublishDraft(context.Background(), project.ProjectID)
		if err != nil {
			_, stale := m.parameters.ProjectDraftState(project.ProjectID)
			return messages.ParametersLoadedMsg{Project: project, Err: err, HasDraft: true, StaleDraft: stale}
		}
		return messages.ParametersLoadedMsg{Project: project, Tree: tree, Source: "firebase", CacheSource: "firebase", HasDraft: false}
	}
}

func (m Model) renameParameterCmd(project core.Project, groupKey, paramKey, nextParamKey string, publish bool) tea.Cmd {
	return m.draftMutationCmd(project, publish, "", "", false, func(ctx context.Context) (*core.ParametersCache, *core.ParametersTree, bool, error) {
		return m.svc.RenameParameter(ctx, project.ProjectID, groupKey, paramKey, nextParamKey, publish)
	})
}

func (m Model) renameGroupCmd(project core.Project, groupKey, nextGroupKey string, publish bool) tea.Cmd {
	return m.draftMutationCmd(project, publish, "", "", false, func(ctx context.Context) (*core.ParametersCache, *core.ParametersTree, bool, error) {
		return m.svc.RenameGroup(ctx, project.ProjectID, groupKey, nextGroupKey, publish)
	})
}

func (m Model) moveParameterCmd(project core.Project, groupKey, paramKey, nextGroupKey string, publish bool) tea.Cmd {
	return m.draftMutationCmd(project, publish, "", "", false, func(ctx context.Context) (*core.ParametersCache, *core.ParametersTree, bool, error) {
		return m.svc.MoveParameter(ctx, project.ProjectID, groupKey, paramKey, nextGroupKey, publish)
	})
}

func (m Model) moveGroupCmd(project core.Project, groupKey, nextGroupKey string, publish bool) tea.Cmd {
	return m.draftMutationCmd(project, publish, "", "", false, func(ctx context.Context) (*core.ParametersCache, *core.ParametersTree, bool, error) {
		return m.svc.MoveGroup(ctx, project.ProjectID, groupKey, nextGroupKey, publish)
	})
}

func (m Model) duplicateParameterNamedCmd(project core.Project, groupKey, paramKey, nextParamKey string, publish bool) tea.Cmd {
	return m.draftMutationCmd(project, publish, groupKey, nextParamKey, false, func(ctx context.Context) (*core.ParametersCache, *core.ParametersTree, bool, error) {
		return m.svc.DuplicateParameterNamed(ctx, project.ProjectID, groupKey, paramKey, nextParamKey, publish)
	})
}

func (m Model) editParameterDetailsCmd(project core.Project, edit core.ParameterDetailsEdit, publish bool, closeDetails bool, selectSaved bool) tea.Cmd {
	return func() tea.Msg {
		_, stale := m.parameters.ProjectDraftState(project.ProjectID)
		_, tree, hasDraft, err := m.svc.EditParameterDetails(context.Background(), project.ProjectID, edit, publish)
		if err != nil {
			return messages.ParametersLoadedMsg{Project: project, Err: err, HasDraft: m.parameters.HasDraft(project.ProjectID), StaleDraft: stale, CloseDetails: closeDetails}
		}
		source := "draft"
		if publish {
			source = "firebase"
		}
		msg := messages.ParametersLoadedMsg{
			Project:      project,
			Tree:         tree,
			Source:       source,
			CacheSource:  "cache",
			Err:          nil,
			HasDraft:     hasDraft,
			StaleDraft:   !publish && hasDraft && stale,
			Revalidate:   false,
			CloseDetails: closeDetails,
			DetailsSaved: true,
		}
		if selectSaved {
			msg.SelectGroupKey = edit.NextGroupKey
			msg.SelectParamKey = edit.NextParamKey
		}
		return msg
	}
}

func (m Model) discardDraftCmd(project core.Project) tea.Cmd {
	return func() tea.Msg {
		_, tree, err := m.svc.DiscardDraft(context.Background(), project.ProjectID)
		if err != nil {
			_, stale := m.parameters.ProjectDraftState(project.ProjectID)
			return messages.ParametersLoadedMsg{Project: project, Err: err, HasDraft: true, StaleDraft: stale}
		}
		return messages.ParametersLoadedMsg{Project: project, Tree: tree, Source: "cache", CacheSource: "cache", HasDraft: false}
	}
}
