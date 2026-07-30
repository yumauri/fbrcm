package app

import (
	"context"
	"fmt"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/tui/messages"
)

type draftMutationFunc func(context.Context) (*core.ParametersCache, *core.ParametersTree, bool, error)

func (m Model) draftMutationCmd(project core.Project, publish bool, selectGroupKey, selectParamKey string, closeDetails bool, run draftMutationFunc, changeNote ...string) tea.Cmd {
	return func() tea.Msg {
		_, stale := m.parameters.ProjectDraftState(project.ProjectID)
		ctx, err := tuiChangeNoteContext(changeNote...)
		if err != nil {
			return messages.ParametersLoadedMsg{Project: project, Err: err, HasDraft: m.parameters.HasDraft(project.ProjectID), StaleDraft: stale, CloseDetails: closeDetails}
		}
		_, tree, hasDraft, err := run(ctx)
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

func (m Model) deleteParameterCmd(project core.Project, groupKey, paramKey string, publish bool, closeDetails bool, changeNote ...string) tea.Cmd {
	return m.draftMutationCmd(project, publish, "", "", closeDetails, func(ctx context.Context) (*core.ParametersCache, *core.ParametersTree, bool, error) {
		return m.svc.DeleteParameter(ctx, project.ProjectID, groupKey, paramKey, publish)
	}, changeNote...)
}

func (m Model) deleteGroupCmd(project core.Project, groupKey string, publish, closeDetails bool, changeNote ...string) tea.Cmd {
	return m.draftMutationCmd(project, publish, "", "", closeDetails, func(ctx context.Context) (*core.ParametersCache, *core.ParametersTree, bool, error) {
		return m.svc.DeleteGroup(ctx, project.ProjectID, groupKey, publish)
	}, changeNote...)
}

// deleteConditionalValueCmd removes one conditional value.
func (m Model) deleteConditionalValueCmd(project core.Project, groupKey, paramKey, valueLabel string, publish bool, changeNote ...string) tea.Cmd {
	return m.draftMutationCmd(project, publish, groupKey, paramKey, false, func(ctx context.Context) (*core.ParametersCache, *core.ParametersTree, bool, error) {
		return m.svc.DeleteConditionalValue(ctx, project.ProjectID, groupKey, paramKey, valueLabel, publish)
	}, changeNote...)
}

func (m Model) publishDraftCmd(project core.Project, changeNote ...string) tea.Cmd {
	return func() tea.Msg {
		ctx, err := tuiChangeNoteContext(changeNote...)
		if err != nil {
			_, stale := m.parameters.ProjectDraftState(project.ProjectID)
			return messages.ParametersLoadedMsg{Project: project, Err: err, HasDraft: true, StaleDraft: stale}
		}
		_, tree, err := m.svc.PublishDraft(ctx, project.ProjectID)
		if err != nil {
			_, stale := m.parameters.ProjectDraftState(project.ProjectID)
			return messages.ParametersLoadedMsg{Project: project, Err: err, HasDraft: true, StaleDraft: stale}
		}
		return messages.ParametersLoadedMsg{Project: project, Tree: tree, Source: "firebase", CacheSource: "firebase", HasDraft: false}
	}
}

func (m Model) renameParameterCmd(project core.Project, groupKey, paramKey, nextParamKey string, publish bool, changeNote ...string) tea.Cmd {
	return m.draftMutationCmd(project, publish, "", "", false, func(ctx context.Context) (*core.ParametersCache, *core.ParametersTree, bool, error) {
		return m.svc.RenameParameter(ctx, project.ProjectID, groupKey, paramKey, nextParamKey, publish)
	}, changeNote...)
}

func (m Model) renameGroupCmd(project core.Project, groupKey, nextGroupKey string, publish bool, changeNote ...string) tea.Cmd {
	return m.draftMutationCmd(project, publish, "", "", false, func(ctx context.Context) (*core.ParametersCache, *core.ParametersTree, bool, error) {
		return m.svc.RenameGroup(ctx, project.ProjectID, groupKey, nextGroupKey, publish)
	}, changeNote...)
}

func (m Model) moveParameterCmd(project core.Project, groupKey, paramKey, nextGroupKey string, publish bool, changeNote ...string) tea.Cmd {
	return m.draftMutationCmd(project, publish, "", "", false, func(ctx context.Context) (*core.ParametersCache, *core.ParametersTree, bool, error) {
		return m.svc.MoveParameter(ctx, project.ProjectID, groupKey, paramKey, nextGroupKey, publish)
	}, changeNote...)
}

func (m Model) moveGroupCmd(project core.Project, groupKey, nextGroupKey string, publish bool, changeNote ...string) tea.Cmd {
	return m.draftMutationCmd(project, publish, "", "", false, func(ctx context.Context) (*core.ParametersCache, *core.ParametersTree, bool, error) {
		return m.svc.MoveGroup(ctx, project.ProjectID, groupKey, nextGroupKey, publish)
	}, changeNote...)
}

func (m Model) duplicateParameterNamedCmd(project core.Project, groupKey, paramKey, nextParamKey string, publish bool, changeNote ...string) tea.Cmd {
	return m.draftMutationCmd(project, publish, groupKey, nextParamKey, false, func(ctx context.Context) (*core.ParametersCache, *core.ParametersTree, bool, error) {
		return m.svc.DuplicateParameterNamed(ctx, project.ProjectID, groupKey, paramKey, nextParamKey, publish)
	}, changeNote...)
}

func (m Model) editParameterDetailsCmd(project core.Project, edit core.ParameterDetailsEdit, publish bool, closeDetails bool, selectSaved bool, changeNote ...string) tea.Cmd {
	return func() tea.Msg {
		_, stale := m.parameters.ProjectDraftState(project.ProjectID)
		ctx, err := tuiChangeNoteContext(changeNote...)
		if err != nil {
			return messages.ParametersLoadedMsg{Project: project, Err: err, HasDraft: m.parameters.HasDraft(project.ProjectID), StaleDraft: stale, CloseDetails: closeDetails}
		}
		_, tree, hasDraft, err := m.svc.EditParameterDetails(ctx, project.ProjectID, edit, publish)
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

func (m Model) editGroupDetailsCmd(project core.Project, edit core.GroupDetailsEdit, publish, closeDetails bool, changeNote ...string) tea.Cmd {
	return func() tea.Msg {
		_, stale := m.parameters.ProjectDraftState(project.ProjectID)
		ctx, err := tuiChangeNoteContext(changeNote...)
		if err != nil {
			return messages.ParametersLoadedMsg{Project: project, Err: err, HasDraft: m.parameters.HasDraft(project.ProjectID), StaleDraft: stale, CloseDetails: closeDetails}
		}
		_, tree, hasDraft, err := m.svc.EditGroupDetails(ctx, project.ProjectID, edit, publish)
		if err != nil {
			return messages.ParametersLoadedMsg{Project: project, Err: err, HasDraft: m.parameters.HasDraft(project.ProjectID), StaleDraft: stale, CloseDetails: closeDetails}
		}
		source := "draft"
		if publish {
			source = "firebase"
		}
		return messages.ParametersLoadedMsg{
			Project: project, Tree: tree, Source: source, CacheSource: "cache", HasDraft: hasDraft,
			StaleDraft: !publish && hasDraft && stale, SelectGroupKey: edit.NextName,
			CloseDetails: closeDetails, DetailsSaved: true,
		}
	}
}

func tuiChangeNoteContext(changeNote ...string) (context.Context, error) {
	ctx := context.Background()
	if len(changeNote) == 0 {
		return ctx, nil
	}
	if len(changeNote) > 1 {
		return nil, fmt.Errorf("expected at most one change note")
	}
	return firebase.WithChangeNote(ctx, changeNote[0])
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
