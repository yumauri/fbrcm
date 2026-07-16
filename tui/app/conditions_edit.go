package app

import (
	"context"
	"encoding/json"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
	dialogcmp "github.com/yumauri/fbrcm/tui/components/dialog"
	moveparam "github.com/yumauri/fbrcm/tui/components/moveparam"
	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/panels"
	"github.com/yumauri/fbrcm/tui/styles"
)

type conditionPreviewFunc func() (*core.ParametersCache, json.RawMessage, error)
type conditionApplyFunc func(bool) tea.Cmd

func (m *Model) openNewConditionInput() tea.Cmd {
	project, x, y, ok := m.conditions.CurrentProjectAnchor()
	if !ok {
		return nil
	}
	m.closeOverlays()
	m.conditionEdit = &conditionEditSession{mode: conditionAddName, project: project, creating: true}
	var cmd tea.Cmd
	x += 4
	m.renameInput, cmd = m.renameInput.Open(x, y, 20, max(m.width-x-2, 22), "")
	return cmd
}

func (m *Model) openConditionRenameInput() tea.Cmd {
	anchor, ok := m.conditions.CurrentEditAnchor()
	if !ok {
		return nil
	}
	m.closeOverlays()
	m.conditionEdit = &conditionEditSession{mode: conditionRename, project: anchor.Project, originalName: anchor.Condition.Name, name: anchor.Condition.Name}
	var cmd tea.Cmd
	x, y := anchor.NameOverlayPosition()
	m.renameInput, cmd = m.renameInput.Open(x, y, anchor.Width, anchor.MaxWidth, anchor.Condition.Name)
	return cmd
}

func (m *Model) openConditionExpressionInput() tea.Cmd {
	data, ok := m.conditionDataForAction()
	if !ok {
		return nil
	}
	m.closeOverlays()
	mode := conditionExpression
	expression := data.Condition.Expression
	if m.active == panels.Details && m.detailsVisible && m.details.IsCondition() {
		mode = conditionDetailsExpression
		if edit, ok := m.details.ConditionEdit(); ok {
			expression = edit.NextExpression
		}
	}
	m.conditionEdit = &conditionEditSession{mode: mode, project: data.Project, originalName: data.Condition.Name, name: data.Condition.Name}
	var cmd tea.Cmd
	m.stringInput, cmd = m.stringInput.Open(0, 0, 20, m.width, m.width, m.height, expression, true, true)
	return cmd
}

func (m *Model) openConditionColorPicker() {
	data, ok := m.conditionDataForAction()
	if !ok {
		return
	}
	anchor, _ := m.conditions.CurrentEditAnchor()
	options := make([]moveparam.Option, 0, len(core.ConditionDisplayColors)+1)
	options = append(options, moveparam.Option{Key: "", Label: "No color", KeepForegroundOnSelect: true})
	selected := 0
	for index, color := range core.ConditionDisplayColors {
		options = append(options, moveparam.Option{
			Key:                    color,
			Label:                  "● " + conditionColorLabel(color),
			Foreground:             styles.ConditionLipglossColor(color),
			KeepForegroundOnSelect: true,
		})
		if color == data.Condition.TagColor {
			selected = index + 1
		}
	}
	m.closeOverlays()
	m.conditionEdit = &conditionEditSession{mode: conditionColor, project: data.Project, originalName: data.Condition.Name, name: data.Condition.Name, currentColor: data.Condition.TagColor}
	x, y := anchor.NameOverlayPosition()
	m.moveParam = m.moveParam.OpenOptions(x, y, data.Condition.Name, options, selected)
}

func (m *Model) startConditionMove() {
	data, ok := m.conditionDataForAction()
	if !ok {
		return
	}
	m.closeOverlays()
	if !m.conditions.StartMove() {
		return
	}
	if m.active == panels.Details {
		m.closeDetailsPanel()
	}
	m.conditionEdit = &conditionEditSession{mode: conditionMove, project: data.Project, originalName: data.Condition.Name, name: data.Condition.Name}
}

func (m *Model) submitConditionRenameInput() tea.Cmd {
	session := m.conditionEdit
	if session == nil || (session.mode != conditionAddName && session.mode != conditionRename) {
		return nil
	}
	name, err := core.NormalizeConditionName(m.renameInput.Value())
	if err != nil {
		m.openErrorDialog("Invalid Condition Name", session.project, err.Error())
		return nil
	}
	if session.mode == conditionAddName {
		session.name = name
		session.mode = conditionExpression
		m.conditionEdit = session
		m.closeRenameInput()
		var cmd tea.Cmd
		m.stringInput, cmd = m.stringInput.Open(0, 0, 20, m.width, m.width, m.height, "", true, true)
		return cmd
	}
	if name == session.originalName {
		m.conditionEdit = nil
		m.closeRenameInput()
		return nil
	}
	m.conditionEdit = nil
	m.closeRenameInput()
	return m.startConditionMutation(
		session.project, "Rename Condition?", "Rename Condition Failed", "Rename condition or draft changes?", "Rename",
		func() (*core.ParametersCache, json.RawMessage, error) {
			return m.svc.PreviewRenameCondition(session.project.ProjectID, session.originalName, name)
		},
		func(publish bool) tea.Cmd {
			return m.renameConditionCmd(session.project, session.originalName, name, publish)
		},
	)
}

func (m *Model) cancelConditionMove() {
	m.conditions.CancelMove()
	if m.conditionEdit != nil && m.conditionEdit.mode == conditionMove {
		m.conditionEdit = nil
	}
}

func (m *Model) submitConditionMove() tea.Cmd {
	session := m.conditionEdit
	priority, changed, ok := m.conditions.FinishMove()
	m.conditionEdit = nil
	if session == nil || session.mode != conditionMove || !ok || !changed {
		return nil
	}
	impact, err := m.conditions.CurrentMoveImpact(priority)
	if err != nil {
		m.openErrorDialog("Move Condition Failed", session.project, err.Error())
		return nil
	}
	return m.startConditionMutation(
		session.project, "Move Condition?", "Move Condition Failed", "Move condition or draft changes?", "Move",
		func() (*core.ParametersCache, json.RawMessage, error) {
			return m.svc.PreviewMoveCondition(session.project.ProjectID, session.originalName, priority)
		},
		func(publish bool) tea.Cmd {
			return m.moveConditionCmd(session.project, session.originalName, priority, publish)
		},
		rcdisplay.FormatConditionMoveImpact(len(impact.CrossedConditions), len(impact.AffectedParameters)),
	)
}

func (m *Model) submitConditionExpressionInput() tea.Cmd {
	session := m.conditionEdit
	if session == nil || (session.mode != conditionExpression && session.mode != conditionDetailsExpression) {
		return nil
	}
	expression, err := core.NormalizeConditionExpression(m.stringInput.Value())
	if err != nil {
		m.openErrorDialog("Invalid Condition Expression", session.project, err.Error())
		return nil
	}
	if session.mode == conditionDetailsExpression {
		m.conditionEdit = nil
		m.closeStringInput()
		m.details = m.details.SetConditionExpression(expression)
		return nil
	}
	m.conditionEdit = nil
	m.closeStringInput()
	if session.creating {
		definition := core.ConditionDefinition{Name: session.name, Expression: expression}
		return m.startConditionMutation(
			session.project, "Add Condition?", "Add Condition Failed", "Add condition or draft changes?", "Add",
			func() (*core.ParametersCache, json.RawMessage, error) {
				return m.svc.PreviewAddCondition(session.project.ProjectID, definition, 0)
			},
			func(publish bool) tea.Cmd { return m.addConditionCmd(session.project, definition, 0, publish) },
		)
	}
	data, ok := m.conditionDataForAction()
	if ok && expression == data.Condition.Expression {
		return nil
	}
	edit := core.ConditionEdit{Expression: &expression}
	return m.startConditionMutation(
		session.project, "Edit Condition?", "Edit Condition Failed", "Edit condition or draft changes?", "Apply",
		func() (*core.ParametersCache, json.RawMessage, error) {
			return m.svc.PreviewEditCondition(session.project.ProjectID, session.originalName, edit)
		},
		func(publish bool) tea.Cmd {
			return m.editConditionCmd(session.project, session.originalName, edit, publish)
		},
	)
}

func (m *Model) submitConditionOption() (tea.Cmd, bool) {
	session := m.conditionEdit
	if session == nil || session.mode != conditionColor {
		return nil, false
	}
	option, ok := m.moveParam.Current()
	m.conditionEdit = nil
	m.closeMoveParam()
	if !ok {
		return nil, true
	}
	color := option.Key
	if color == session.currentColor {
		return nil, true
	}
	edit := core.ConditionEdit{TagColor: &color}
	return m.startConditionMutation(
		session.project, "Edit Condition Color?", "Edit Condition Failed", "Edit condition color or draft changes?", "Apply",
		func() (*core.ParametersCache, json.RawMessage, error) {
			return m.svc.PreviewEditCondition(session.project.ProjectID, session.originalName, edit)
		},
		func(publish bool) tea.Cmd {
			return m.editConditionCmd(session.project, session.originalName, edit, publish)
		},
	), true
}

func (m *Model) requestDeleteCondition() tea.Cmd {
	data, ok := m.conditionDataForAction()
	if !ok {
		return nil
	}
	impact, err := m.conditions.CurrentDeleteImpact()
	if err != nil {
		m.openErrorDialog("Delete Condition Failed", data.Project, err.Error())
		return nil
	}
	return m.startConditionMutation(
		data.Project, "Delete Condition?", "Delete Condition Failed", "Delete condition and its conditional values?", "Delete",
		func() (*core.ParametersCache, json.RawMessage, error) {
			return m.svc.PreviewDeleteCondition(data.Project.ProjectID, data.Condition.Name)
		},
		func(publish bool) tea.Cmd { return m.deleteConditionCmd(data.Project, data.Condition.Name, publish) },
		rcdisplay.FormatConditionDeleteImpact(len(impact.Usages), len(impact.RemovedParameters)),
	)
}

func (m *Model) startConditionMutation(project core.Project, title, errorTitle, prompt, applyLabel string, preview conditionPreviewFunc, apply conditionApplyFunc, impact ...string) tea.Cmd {
	if m.parameters.HasDraft(project.ProjectID) || m.conditions.HasDraft(project.ProjectID) {
		return apply(false)
	}
	body := func() ([]string, error) {
		lines, err := m.previewDialogBody(project, prompt, "condition not changed", previewDialogFunc(preview), nil)
		if err != nil {
			return nil, err
		}
		if len(impact) > 0 && impact[0] != "" {
			lines = append(lines, "", impact[0])
		}
		return lines, nil
	}
	m.openPreviewDialog(project, title, errorTitle, body, func(error) {}, []dialogcmp.Button{
		{Label: applyLabel, Variant: dialogcmp.ButtonVariantDanger, OnPress: apply(true)},
		{Label: "Draft", Variant: dialogcmp.ButtonVariantAccent, OnPress: apply(false)},
		{Label: "Cancel", Variant: dialogcmp.ButtonVariantAccent, OnPress: dialogCanceledCmd()},
	})
	return nil
}

func (m Model) conditionMutationCmd(project core.Project, selectConditionName string, publish bool, run draftMutationFunc) tea.Cmd {
	return func() tea.Msg {
		_, stale := m.parameters.ProjectDraftState(project.ProjectID)
		_, tree, hasDraft, err := run(context.Background())
		if err != nil {
			return messages.ParametersLoadedMsg{Project: project, Err: err, HasDraft: m.parameters.HasDraft(project.ProjectID), StaleDraft: stale, SelectConditionName: selectConditionName}
		}
		source := "draft"
		if publish {
			source = "firebase"
		}
		return messages.ParametersLoadedMsg{Project: project, Tree: tree, Source: source, CacheSource: "cache", HasDraft: hasDraft, StaleDraft: !publish && hasDraft && stale, SelectConditionName: selectConditionName}
	}
}

func (m Model) addConditionCmd(project core.Project, definition core.ConditionDefinition, priority int, publish bool) tea.Cmd {
	return m.conditionMutationCmd(project, definition.Name, publish, func(ctx context.Context) (*core.ParametersCache, *core.ParametersTree, bool, error) {
		return m.svc.AddCondition(ctx, project.ProjectID, definition, priority, publish)
	})
}

func (m Model) editConditionCmd(project core.Project, name string, edit core.ConditionEdit, publish bool) tea.Cmd {
	return m.conditionMutationCmd(project, name, publish, func(ctx context.Context) (*core.ParametersCache, *core.ParametersTree, bool, error) {
		return m.svc.EditCondition(ctx, project.ProjectID, name, edit, publish)
	})
}

func (m Model) renameConditionCmd(project core.Project, name, nextName string, publish bool) tea.Cmd {
	return m.conditionMutationCmd(project, nextName, publish, func(ctx context.Context) (*core.ParametersCache, *core.ParametersTree, bool, error) {
		return m.svc.RenameCondition(ctx, project.ProjectID, name, nextName, publish)
	})
}

func (m Model) moveConditionCmd(project core.Project, name string, priority int, publish bool) tea.Cmd {
	return m.conditionMutationCmd(project, name, publish, func(ctx context.Context) (*core.ParametersCache, *core.ParametersTree, bool, error) {
		return m.svc.MoveCondition(ctx, project.ProjectID, name, priority, publish)
	})
}

func (m Model) deleteConditionCmd(project core.Project, name string, publish bool) tea.Cmd {
	return m.conditionMutationCmd(project, "", publish, func(ctx context.Context) (*core.ParametersCache, *core.ParametersTree, bool, error) {
		return m.svc.DeleteCondition(ctx, project.ProjectID, name, publish)
	})
}

func (m Model) conditionDataForAction() (*messages.ConditionViewData, bool) {
	if m.active == panels.Details && m.detailsVisible {
		if data := m.details.ConditionData(); data != nil {
			return data, true
		}
	}
	return m.conditions.CurrentCondition()
}

func (m Model) conditionDetailsMutationCmd(project core.Project, edit core.ConditionDetailsEdit, publish, closeDetails bool) tea.Cmd {
	return func() tea.Msg {
		_, stale := m.parameters.ProjectDraftState(project.ProjectID)
		_, tree, hasDraft, err := m.svc.EditConditionDetails(context.Background(), project.ProjectID, edit, publish)
		if err != nil {
			return messages.ParametersLoadedMsg{
				Project: project, Err: err, HasDraft: m.parameters.HasDraft(project.ProjectID), StaleDraft: stale, CloseDetails: closeDetails,
			}
		}
		source := "draft"
		if publish {
			source = "firebase"
		}
		return messages.ParametersLoadedMsg{
			Project:             project,
			Tree:                tree,
			Source:              source,
			CacheSource:         "cache",
			HasDraft:            hasDraft,
			StaleDraft:          !publish && hasDraft && stale,
			SelectConditionName: edit.NextName,
			CloseDetails:        closeDetails,
			DetailsSaved:        true,
		}
	}
}

func (m *Model) openConditionDetailsDialog(project core.Project, edit core.ConditionDetailsEdit, closeDetails bool) {
	m.openPreviewDialog(project, "Edit Condition?", "Edit Condition Failed", func() ([]string, error) {
		return m.previewDialogBody(project, "Edit condition or draft changes?", "condition not changed", func() (*core.ParametersCache, json.RawMessage, error) {
			return m.svc.PreviewEditConditionDetails(project.ProjectID, edit)
		}, func(cache *core.ParametersCache) (json.RawMessage, error) {
			if draftRaw, hasDraft, err := m.svc.LoadDraft(project.ProjectID); err != nil {
				return nil, err
			} else if hasDraft {
				return draftRaw, nil
			}
			return cache.RemoteConfig, nil
		})
	}, func(error) {}, []dialogcmp.Button{
		{Label: "Apply", Variant: dialogcmp.ButtonVariantDanger, OnPress: m.conditionDetailsMutationCmd(project, edit, true, closeDetails)},
		{Label: "Draft", Variant: dialogcmp.ButtonVariantAccent, OnPress: m.conditionDetailsMutationCmd(project, edit, false, closeDetails)},
		{Label: "Cancel", Variant: dialogcmp.ButtonVariantAccent, OnPress: detailsEditCanceledCmd(closeDetails)},
	})
}

func conditionColorLabel(color string) string {
	if color == "" {
		return "No color"
	}
	return strings.ReplaceAll(color, "_", " ")
}
