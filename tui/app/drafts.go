package app

import (
	"context"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	corelog "github.com/yumauri/fbrcm/core/log"
	dialogcmp "github.com/yumauri/fbrcm/tui/components/dialog"
	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/styles"
)

var (
	dialogParameterNameStyle = lipgloss.NewStyle().Bold(true).Foreground(styles.PaletteBlueBright)
	dialogProjectNameStyle   = lipgloss.NewStyle().Bold(true).Foreground(styles.PaletteError)
)

// closeDialog closes close dialog for Model and returns the resulting state or error.
func (m *Model) closeDialog(openNext bool) {
	if !m.dialog.IsOpen() {
		return
	}
	m.dialog = m.dialog.Close()
	if !openNext {
		m.dialogQueue = nil
		return
	}
	if len(m.dialogQueue) > 0 {
		next := m.dialogQueue[0]
		m.dialogQueue = m.dialogQueue[1:]
		m.openDraftDialog(next.project, next.mode, nil)
	}
}

// openDeleteDialog opens open delete dialog for Model and returns the resulting state or error.
func (m *Model) openDeleteDialog(project core.Project, groupKey, paramKey string, closeDetails bool) {
	body, ok := m.deleteDialogBody(project, groupKey, paramKey)
	if !ok {
		body = []string{
			"Project: " + dialogProjectNameStyle.Render(project.Name) + " (" + project.ProjectID + ")",
			"",
			"Delete parameter or draft changes?",
			"",
			"Parameter: " + dialogParameterNameStyle.Render(paramKey),
		}
	}
	m.dialog = m.dialog.Open(dialogcmp.Config{
		Title: "Delete Parameter?",
		Body:  body,
		Buttons: []dialogcmp.Button{
			{Label: "Delete", Variant: dialogcmp.ButtonVariantDanger, OnPress: m.deleteParameterCmd(project, groupKey, paramKey, true, closeDetails)},
			{Label: "Draft", Variant: dialogcmp.ButtonVariantAccent, OnPress: m.deleteParameterCmd(project, groupKey, paramKey, false, closeDetails)},
			{Label: "Cancel", Variant: dialogcmp.ButtonVariantAccent, OnPress: dialogCanceledCmd()},
		},
	})
}

// openDeleteGroupDialog opens open delete group dialog for Model and returns the resulting state or error.
func (m *Model) openDeleteGroupDialog(project core.Project, groupKey, groupLabel string) {
	body, ok := m.deleteGroupDialogBody(project, groupKey)
	if !ok {
		body = []string{
			"Project: " + dialogProjectNameStyle.Render(project.Name) + " (" + project.ProjectID + ")",
			"",
			"Delete group or draft changes?",
			"",
			"Group: " + dialogParameterNameStyle.Render(groupLabel),
		}
	}
	m.dialog = m.dialog.Open(dialogcmp.Config{
		Title: "Delete Group?",
		Body:  body,
		Buttons: []dialogcmp.Button{
			{Label: "Delete", Variant: dialogcmp.ButtonVariantDanger, OnPress: m.deleteGroupCmd(project, groupKey, true)},
			{Label: "Draft", Variant: dialogcmp.ButtonVariantAccent, OnPress: m.deleteGroupCmd(project, groupKey, false)},
			{Label: "Cancel", Variant: dialogcmp.ButtonVariantAccent, OnPress: dialogCanceledCmd()},
		},
	})
}

// openDeleteConditionalValueDialog opens delete conditional value dialog.
func (m *Model) openDeleteConditionalValueDialog(project core.Project, groupKey, paramKey, valueLabel string) {
	body, err := m.deleteConditionalValueDialogBody(project, groupKey, paramKey, valueLabel)
	if err != nil {
		corelog.For("tui.delete").Error("delete conditional value preview failed", "project_id", project.ProjectID, "group", groupKey, "param", paramKey, "value_label", valueLabel, "err", err)
		m.openErrorDialog("Delete Conditional Value Failed", project, err.Error())
		return
	}
	m.dialog = m.dialog.Open(dialogcmp.Config{
		Title: "Delete Conditional Value?",
		Body:  body,
		Buttons: []dialogcmp.Button{
			{Label: "Delete", Variant: dialogcmp.ButtonVariantDanger, OnPress: m.deleteConditionalValueCmd(project, groupKey, paramKey, valueLabel, true)},
			{Label: "Draft", Variant: dialogcmp.ButtonVariantAccent, OnPress: m.deleteConditionalValueCmd(project, groupKey, paramKey, valueLabel, false)},
			{Label: "Cancel", Variant: dialogcmp.ButtonVariantAccent, OnPress: dialogCanceledCmd()},
		},
	})
}

// openDraftDialog opens open draft dialog for Model and returns the resulting state or error.
func (m *Model) openDraftDialog(project core.Project, mode dialogMode, queue []pendingDialog) {
	body, ok := m.draftDialogBody(project, mode)
	if !ok {
		if len(queue) > 0 {
			next := queue[0]
			m.openDraftDialog(next.project, next.mode, queue[1:])
		}
		return
	}

	buttons := []dialogcmp.Button{
		{Label: "Discard", Variant: dialogcmp.ButtonVariantAccent, OnPress: m.discardDraftCmd(project)},
		{Label: "Cancel", Variant: dialogcmp.ButtonVariantAccent, OnPress: dialogCanceledCmd()},
	}
	title := "Discard Draft?"
	if mode == dialogModePublishDraft {
		title = "Publish Draft?"
		buttons = []dialogcmp.Button{
			{Label: "Publish", Variant: dialogcmp.ButtonVariantDanger, OnPress: m.publishDraftCmd(project)},
			{Label: "Discard", Variant: dialogcmp.ButtonVariantAccent, OnPress: m.discardDraftCmd(project)},
			{Label: "Cancel", Variant: dialogcmp.ButtonVariantAccent, OnPress: dialogCanceledCmd()},
		}
	}

	m.dialogQueue = append([]pendingDialog(nil), queue...)
	m.dialog = m.dialog.Open(dialogcmp.Config{
		Title:   title,
		Body:    body,
		Buttons: buttons,
	})
}

// openRenameDialog opens open rename dialog for Model and returns the resulting state or error.
func (m *Model) openRenameDialog(project core.Project, groupKey, paramKey, nextParamKey string) {
	body, err := m.renameDialogBody(project, groupKey, paramKey, nextParamKey)
	if err != nil {
		corelog.For("tui.rename").Error("rename preview failed", "project_id", project.ProjectID, "group", groupKey, "param", paramKey, "next_param", nextParamKey, "err", err)
		m.openErrorDialog("Rename Failed", project, err.Error())
		return
	}
	m.dialog = m.dialog.Open(dialogcmp.Config{
		Title: "Rename Parameter?",
		Body:  body,
		Buttons: []dialogcmp.Button{
			{Label: "Rename", Variant: dialogcmp.ButtonVariantDanger, OnPress: m.renameParameterCmd(project, groupKey, paramKey, nextParamKey, true)},
			{Label: "Draft", Variant: dialogcmp.ButtonVariantAccent, OnPress: m.renameParameterCmd(project, groupKey, paramKey, nextParamKey, false)},
			{Label: "Cancel", Variant: dialogcmp.ButtonVariantAccent, OnPress: dialogCanceledCmd()},
		},
	})
}

// openRenameGroupDialog opens open rename group dialog for Model and returns the resulting state or error.
func (m *Model) openRenameGroupDialog(project core.Project, groupKey, nextGroupKey string) {
	body, err := m.renameGroupDialogBody(project, groupKey, nextGroupKey)
	if err != nil {
		corelog.For("tui.rename").Error("rename group preview failed", "project_id", project.ProjectID, "group", groupKey, "next_group", nextGroupKey, "err", err)
		m.openErrorDialog("Rename Group Failed", project, err.Error())
		return
	}
	m.dialog = m.dialog.Open(dialogcmp.Config{
		Title: "Rename Group?",
		Body:  body,
		Buttons: []dialogcmp.Button{
			{Label: "Rename", Variant: dialogcmp.ButtonVariantDanger, OnPress: m.renameGroupCmd(project, groupKey, nextGroupKey, true)},
			{Label: "Draft", Variant: dialogcmp.ButtonVariantAccent, OnPress: m.renameGroupCmd(project, groupKey, nextGroupKey, false)},
			{Label: "Cancel", Variant: dialogcmp.ButtonVariantAccent, OnPress: dialogCanceledCmd()},
		},
	})
}

// openMoveDialog opens open move dialog for Model and returns the resulting state or error.
func (m *Model) openMoveDialog(project core.Project, groupKey, paramKey, nextGroupKey string) {
	body, err := m.moveDialogBody(project, groupKey, paramKey, nextGroupKey)
	if err != nil {
		corelog.For("tui.move").Error("move parameter preview failed", "project_id", project.ProjectID, "group", groupKey, "param", paramKey, "next_group", nextGroupKey, "err", err)
		m.openErrorDialog("Move Parameter Failed", project, err.Error())
		return
	}
	m.dialog = m.dialog.Open(dialogcmp.Config{
		Title: "Move Parameter?",
		Body:  body,
		Buttons: []dialogcmp.Button{
			{Label: "Move", Variant: dialogcmp.ButtonVariantDanger, OnPress: m.moveParameterCmd(project, groupKey, paramKey, nextGroupKey, true)},
			{Label: "Draft", Variant: dialogcmp.ButtonVariantAccent, OnPress: m.moveParameterCmd(project, groupKey, paramKey, nextGroupKey, false)},
			{Label: "Cancel", Variant: dialogcmp.ButtonVariantAccent, OnPress: dialogCanceledCmd()},
		},
	})
}

// openMoveGroupDialog opens open move group dialog for Model and returns the resulting state or error.
func (m *Model) openMoveGroupDialog(project core.Project, groupKey, nextGroupKey string) {
	body, err := m.moveGroupDialogBody(project, groupKey, nextGroupKey)
	if err != nil {
		corelog.For("tui.move").Error("move group preview failed", "project_id", project.ProjectID, "group", groupKey, "next_group", nextGroupKey, "err", err)
		m.openErrorDialog("Move Group Failed", project, err.Error())
		return
	}
	m.dialog = m.dialog.Open(dialogcmp.Config{
		Title: "Move Group?",
		Body:  body,
		Buttons: []dialogcmp.Button{
			{Label: "Move", Variant: dialogcmp.ButtonVariantDanger, OnPress: m.moveGroupCmd(project, groupKey, nextGroupKey, true)},
			{Label: "Draft", Variant: dialogcmp.ButtonVariantAccent, OnPress: m.moveGroupCmd(project, groupKey, nextGroupKey, false)},
			{Label: "Cancel", Variant: dialogcmp.ButtonVariantAccent, OnPress: dialogCanceledCmd()},
		},
	})
}

// openDuplicateDialog opens open duplicate dialog for Model and returns the resulting state or error.
func (m *Model) openDuplicateDialog(project core.Project, groupKey, paramKey, nextParamKey string) {
	body, err := m.duplicateDialogBody(project, groupKey, paramKey, nextParamKey)
	if err != nil {
		corelog.For("tui.duplicate").Error("duplicate parameter preview failed", "project_id", project.ProjectID, "group", groupKey, "param", paramKey, "next_param", nextParamKey, "err", err)
		m.openErrorDialog("Duplicate Failed", project, err.Error())
		return
	}
	m.dialog = m.dialog.Open(dialogcmp.Config{
		Title: "Duplicate Parameter?",
		Body:  body,
		Buttons: []dialogcmp.Button{
			{Label: "Duplicate", Variant: dialogcmp.ButtonVariantDanger, OnPress: m.duplicateParameterNamedCmd(project, groupKey, paramKey, nextParamKey, true)},
			{Label: "Draft", Variant: dialogcmp.ButtonVariantAccent, OnPress: m.duplicateParameterNamedCmd(project, groupKey, paramKey, nextParamKey, false)},
			{Label: "Cancel", Variant: dialogcmp.ButtonVariantAccent, OnPress: dialogCanceledCmd()},
		},
	})
}

// openEditDetailsDialog opens open edit details dialog for Model and returns the resulting state or error.
func (m *Model) openEditDetailsDialog(project core.Project, edit core.ParameterDetailsEdit, closeDetails bool, selectSaved bool) {
	body, err := m.editDetailsDialogBody(project, edit)
	if err != nil {
		corelog.For("tui.details").Error("edit parameter details preview failed", "project_id", project.ProjectID, "group", edit.GroupKey, "param", edit.ParamKey, "err", err)
		title := "Edit Parameter Failed"
		if edit.Create {
			title = "Create Parameter Failed"
		}
		m.openErrorDialog(title, project, err.Error())
		return
	}
	title := "Edit Parameter?"
	applyLabel := "Apply"
	if edit.Create {
		title = "Create Parameter?"
		applyLabel = "Create"
	}
	m.dialog = m.dialog.Open(dialogcmp.Config{
		Title: title,
		Body:  body,
		Buttons: []dialogcmp.Button{
			{Label: applyLabel, Variant: dialogcmp.ButtonVariantDanger, OnPress: m.editParameterDetailsCmd(project, edit, true, closeDetails, selectSaved)},
			{Label: "Draft", Variant: dialogcmp.ButtonVariantAccent, OnPress: m.editParameterDetailsCmd(project, edit, false, closeDetails, selectSaved)},
			{Label: "Cancel", Variant: dialogcmp.ButtonVariantAccent, OnPress: detailsEditCanceledCmd(closeDetails)},
		},
	})
}

// openInvalidDetailsDialog opens open invalid details dialog for Model and returns the resulting state or error.
func (m *Model) openInvalidDetailsDialog(project core.Project, reasons []string, closeDetails bool) {
	body := []string{
		"Project: " + dialogProjectNameStyle.Render(project.Name) + " (" + project.ProjectID + ")",
		"",
		"Details form is invalid.",
		"",
	}
	if len(reasons) == 0 {
		body = append(body, "Fix highlighted fields before applying changes.")
	} else {
		body = append(body, reasons...)
	}
	m.dialog = m.dialog.Open(dialogcmp.Config{
		Title: "Invalid Details?",
		Body:  body,
		Buttons: []dialogcmp.Button{
			{Label: "Fix", Variant: dialogcmp.ButtonVariantAccent, OnPress: detailsInvalidFixCmd()},
			{Label: "Discard", Variant: dialogcmp.ButtonVariantAccent, OnPress: detailsInvalidDiscardCmd(closeDetails)},
		},
	})
}

// openErrorDialog opens open error dialog for Model and returns the resulting state or error.
func (m *Model) openErrorDialog(title string, project core.Project, errText string) {
	body := []string{
		"Project: " + dialogProjectNameStyle.Render(project.Name) + " (" + project.ProjectID + ")",
		"",
		errText,
	}
	m.dialog = m.dialog.Open(dialogcmp.Config{
		Title: title,
		Body:  body,
		Buttons: []dialogcmp.Button{
			{Label: "Close", Variant: dialogcmp.ButtonVariantAccent, OnPress: dialogCanceledCmd()},
		},
	})
}

func (m Model) deleteDialogBody(project core.Project, groupKey, paramKey string) ([]string, bool) {
	cache, finalRaw, err := m.svc.PreviewDeleteParameter(project.ProjectID, groupKey, paramKey)
	if err != nil || cache == nil {
		return nil, false
	}

	currentCfg, err := firebase.ParseRemoteConfig(cache.RemoteConfig)
	if err != nil {
		return nil, false
	}
	finalCfg, err := firebase.ParseRemoteConfig(finalRaw)
	if err != nil {
		return nil, false
	}

	diffText, hasChanges := shared.RenderRemoteConfigDiff(currentCfg, finalCfg)
	lines := []string{
		"Project: " + dialogProjectNameStyle.Render(project.Name) + " (" + project.ProjectID + ")",
		"",
		"Delete parameter or draft changes?",
	}
	if !hasChanges {
		lines = append(lines, "", "Parameter: "+dialogParameterNameStyle.Render(paramKey))
		return lines, true
	}

	lines = append(lines, "")
	lines = append(lines, dialogDiffLines(diffText)...)
	return lines, true
}

func (m Model) deleteGroupDialogBody(project core.Project, groupKey string) ([]string, bool) {
	cache, finalRaw, err := m.svc.PreviewDeleteGroup(project.ProjectID, groupKey)
	if err != nil || cache == nil {
		return nil, false
	}

	currentCfg, err := firebase.ParseRemoteConfig(cache.RemoteConfig)
	if err != nil {
		return nil, false
	}
	finalCfg, err := firebase.ParseRemoteConfig(finalRaw)
	if err != nil {
		return nil, false
	}

	diffText, hasChanges := shared.RenderRemoteConfigDiff(currentCfg, finalCfg)
	lines := []string{
		"Project: " + dialogProjectNameStyle.Render(project.Name) + " (" + project.ProjectID + ")",
		"",
		"Delete group or draft changes?",
	}
	if !hasChanges {
		return lines, true
	}
	lines = append(lines, "")
	lines = append(lines, dialogDiffLines(diffText)...)
	return lines, true
}

func (m Model) deleteConditionalValueDialogBody(project core.Project, groupKey, paramKey, valueLabel string) ([]string, error) {
	cache, finalRaw, err := m.svc.PreviewDeleteConditionalValue(project.ProjectID, groupKey, paramKey, valueLabel)
	if err != nil || cache == nil {
		if err == nil {
			err = fmt.Errorf("parameters cache not found")
		}
		return nil, err
	}

	currentCfg, err := firebase.ParseRemoteConfig(cache.RemoteConfig)
	if err != nil {
		return nil, err
	}
	finalCfg, err := firebase.ParseRemoteConfig(finalRaw)
	if err != nil {
		return nil, err
	}

	diffText, hasChanges := shared.RenderRemoteConfigDiff(currentCfg, finalCfg)
	lines := []string{
		"Project: " + dialogProjectNameStyle.Render(project.Name) + " (" + project.ProjectID + ")",
		"",
		"Delete conditional value or draft changes?",
	}
	if !hasChanges {
		return nil, fmt.Errorf("conditional value not changed")
	}

	lines = append(lines, "")
	lines = append(lines, dialogDiffLines(diffText)...)
	return lines, nil
}

func (m Model) draftDialogBody(project core.Project, mode dialogMode) ([]string, bool) {
	cache, _, err := m.svc.InspectParametersCache(project.ProjectID)
	if err != nil || cache == nil {
		return nil, false
	}
	draftRaw, hasDraft, err := m.svc.LoadDraft(project.ProjectID)
	if err != nil || !hasDraft {
		return nil, false
	}

	currentCfg, err := firebase.ParseRemoteConfig(cache.RemoteConfig)
	if err != nil {
		return nil, false
	}
	draftCfg, err := firebase.ParseRemoteConfig(draftRaw)
	if err != nil {
		return nil, false
	}
	diffText, hasChanges := shared.RenderRemoteConfigDiff(currentCfg, draftCfg)
	if !hasChanges {
		diffText = "\nNo changes.\n"
	}

	lines := []string{
		"Project: " + dialogProjectNameStyle.Render(project.Name) + " (" + project.ProjectID + ")",
		"",
	}
	if mode == dialogModePublishDraft {
		lines = append(lines, "Publish draft changes?")
	} else {
		lines = append(lines, "Discard draft changes?")
	}
	lines = append(lines, "")
	lines = append(lines, dialogDiffLines(diffText)...)
	return lines, true
}

func (m Model) renameDialogBody(project core.Project, groupKey, paramKey, nextParamKey string) ([]string, error) {
	cache, finalRaw, err := m.svc.PreviewRenameParameter(project.ProjectID, groupKey, paramKey, nextParamKey)
	if err != nil || cache == nil {
		if err == nil {
			err = fmt.Errorf("parameters cache not found")
		}
		return nil, err
	}

	currentCfg, err := firebase.ParseRemoteConfig(cache.RemoteConfig)
	if err != nil {
		return nil, err
	}
	finalCfg, err := firebase.ParseRemoteConfig(finalRaw)
	if err != nil {
		return nil, err
	}

	diffText, hasChanges := shared.RenderRemoteConfigDiff(currentCfg, finalCfg)
	lines := []string{
		"Project: " + dialogProjectNameStyle.Render(project.Name) + " (" + project.ProjectID + ")",
		"",
		"Rename parameter or draft changes?",
	}
	if !hasChanges {
		return nil, fmt.Errorf("parameter not changed")
	}

	lines = append(lines, "")
	lines = append(lines, dialogDiffLines(diffText)...)
	return lines, nil
}

func (m Model) renameGroupDialogBody(project core.Project, groupKey, nextGroupKey string) ([]string, error) {
	cache, finalRaw, err := m.svc.PreviewRenameGroup(project.ProjectID, groupKey, nextGroupKey)
	if err != nil || cache == nil {
		if err == nil {
			err = fmt.Errorf("parameters cache not found")
		}
		return nil, err
	}

	currentCfg, err := firebase.ParseRemoteConfig(cache.RemoteConfig)
	if err != nil {
		return nil, err
	}
	finalCfg, err := firebase.ParseRemoteConfig(finalRaw)
	if err != nil {
		return nil, err
	}

	diffText, hasChanges := shared.RenderRemoteConfigDiff(currentCfg, finalCfg)
	lines := []string{
		"Project: " + dialogProjectNameStyle.Render(project.Name) + " (" + project.ProjectID + ")",
		"",
		"Rename group or draft changes?",
	}
	if !hasChanges {
		return nil, fmt.Errorf("group not changed")
	}

	lines = append(lines, "")
	lines = append(lines, dialogDiffLines(diffText)...)
	return lines, nil
}

// moveDialogBody moves move dialog body for Model and returns the resulting state or error.
func (m Model) moveDialogBody(project core.Project, groupKey, paramKey, nextGroupKey string) ([]string, error) {
	cache, finalRaw, err := m.svc.PreviewMoveParameter(project.ProjectID, groupKey, paramKey, nextGroupKey)
	if err != nil || cache == nil {
		if err == nil {
			err = fmt.Errorf("parameters cache not found")
		}
		return nil, err
	}

	currentCfg, err := firebase.ParseRemoteConfig(cache.RemoteConfig)
	if err != nil {
		return nil, err
	}
	finalCfg, err := firebase.ParseRemoteConfig(finalRaw)
	if err != nil {
		return nil, err
	}

	diffText, hasChanges := shared.RenderRemoteConfigDiff(currentCfg, finalCfg)
	lines := []string{
		"Project: " + dialogProjectNameStyle.Render(project.Name) + " (" + project.ProjectID + ")",
		"",
		"Move parameter or draft changes?",
	}
	if !hasChanges {
		return nil, fmt.Errorf("parameter not changed")
	}

	lines = append(lines, "")
	lines = append(lines, dialogDiffLines(diffText)...)
	return lines, nil
}

// moveGroupDialogBody moves move group dialog body for Model and returns the resulting state or error.
func (m Model) moveGroupDialogBody(project core.Project, groupKey, nextGroupKey string) ([]string, error) {
	cache, finalRaw, err := m.svc.PreviewMoveGroup(project.ProjectID, groupKey, nextGroupKey)
	if err != nil || cache == nil {
		if err == nil {
			err = fmt.Errorf("parameters cache not found")
		}
		return nil, err
	}

	currentCfg, err := firebase.ParseRemoteConfig(cache.RemoteConfig)
	if err != nil {
		return nil, err
	}
	finalCfg, err := firebase.ParseRemoteConfig(finalRaw)
	if err != nil {
		return nil, err
	}

	diffText, hasChanges := shared.RenderRemoteConfigDiff(currentCfg, finalCfg)
	lines := []string{
		"Project: " + dialogProjectNameStyle.Render(project.Name) + " (" + project.ProjectID + ")",
		"",
		"Move group or draft changes?",
	}
	if !hasChanges {
		return nil, fmt.Errorf("group not changed")
	}

	lines = append(lines, "")
	lines = append(lines, dialogDiffLines(diffText)...)
	return lines, nil
}

func (m Model) duplicateDialogBody(project core.Project, groupKey, paramKey, nextParamKey string) ([]string, error) {
	cache, finalRaw, err := m.svc.PreviewDuplicateParameter(project.ProjectID, groupKey, paramKey, nextParamKey)
	if err != nil || cache == nil {
		if err == nil {
			err = fmt.Errorf("parameters cache not found")
		}
		return nil, err
	}

	currentCfg, err := firebase.ParseRemoteConfig(cache.RemoteConfig)
	if err != nil {
		return nil, err
	}
	finalCfg, err := firebase.ParseRemoteConfig(finalRaw)
	if err != nil {
		return nil, err
	}

	diffText, hasChanges := shared.RenderRemoteConfigDiff(currentCfg, finalCfg)
	lines := []string{
		"Project: " + dialogProjectNameStyle.Render(project.Name) + " (" + project.ProjectID + ")",
		"",
		"Duplicate parameter or draft changes?",
	}
	if !hasChanges {
		return nil, fmt.Errorf("parameter not changed")
	}

	lines = append(lines, "")
	lines = append(lines, dialogDiffLines(diffText)...)
	return lines, nil
}

func (m Model) editDetailsDialogBody(project core.Project, edit core.ParameterDetailsEdit) ([]string, error) {
	cache, finalRaw, err := m.svc.PreviewEditParameterDetails(project.ProjectID, edit)
	if err != nil || cache == nil {
		if err == nil {
			err = fmt.Errorf("parameters cache not found")
		}
		return nil, err
	}

	currentRaw := cache.RemoteConfig
	if draftRaw, hasDraft, err := m.svc.LoadDraft(project.ProjectID); err != nil {
		return nil, err
	} else if hasDraft {
		currentRaw = draftRaw
	}
	currentCfg, err := firebase.ParseRemoteConfig(currentRaw)
	if err != nil {
		return nil, err
	}
	finalCfg, err := firebase.ParseRemoteConfig(finalRaw)
	if err != nil {
		return nil, err
	}

	diffText, hasChanges := shared.RenderRemoteConfigDiff(currentCfg, finalCfg)
	lines := []string{
		"Project: " + dialogProjectNameStyle.Render(project.Name) + " (" + project.ProjectID + ")",
		"",
		"Edit parameter or draft changes?",
	}
	if !hasChanges {
		return nil, fmt.Errorf("parameter not changed")
	}
	lines = append(lines, "")
	lines = append(lines, dialogDiffLines(diffText)...)
	return lines, nil
}

func dialogDiffLines(diffText string) []string {
	diffText = strings.Trim(diffText, "\n")
	if idx := strings.Index(diffText, "\n\nSummary:\n"); idx >= 0 {
		diffText = diffText[:idx]
	}
	if diffText == "" {
		return []string{"No changes."}
	}
	return strings.Split(diffText, "\n")
}

func (m Model) deleteParameterCmd(project core.Project, groupKey, paramKey string, publish bool, closeDetails bool) tea.Cmd {
	return func() tea.Msg {
		_, stale := m.parameters.ProjectDraftState(project.ProjectID)
		_, tree, hasDraft, err := m.svc.DeleteParameter(context.Background(), project.ProjectID, groupKey, paramKey, publish)
		if err != nil {
			return messages.ParametersLoadedMsg{Project: project, Err: err, HasDraft: m.parameters.HasDraft(project.ProjectID), StaleDraft: stale, CloseDetails: closeDetails}
		}
		source := "draft"
		if publish {
			source = "firebase"
		}
		return messages.ParametersLoadedMsg{
			Project:      project,
			Tree:         tree,
			Source:       source,
			CacheSource:  "cache",
			Err:          nil,
			CloseDetails: closeDetails,
			HasDraft:     hasDraft,
			StaleDraft:   !publish && hasDraft && stale,
			Revalidate:   false,
		}
	}
}

func (m Model) deleteGroupCmd(project core.Project, groupKey string, publish bool) tea.Cmd {
	return func() tea.Msg {
		_, stale := m.parameters.ProjectDraftState(project.ProjectID)
		_, tree, hasDraft, err := m.svc.DeleteGroup(context.Background(), project.ProjectID, groupKey, publish)
		if err != nil {
			return messages.ParametersLoadedMsg{Project: project, Err: err, HasDraft: m.parameters.HasDraft(project.ProjectID), StaleDraft: stale}
		}
		source := "draft"
		if publish {
			source = "firebase"
		}
		return messages.ParametersLoadedMsg{
			Project:     project,
			Tree:        tree,
			Source:      source,
			CacheSource: "cache",
			Err:         nil,
			HasDraft:    hasDraft,
			StaleDraft:  !publish && hasDraft && stale,
			Revalidate:  false,
		}
	}
}

// deleteConditionalValueCmd removes one conditional value.
func (m Model) deleteConditionalValueCmd(project core.Project, groupKey, paramKey, valueLabel string, publish bool) tea.Cmd {
	return func() tea.Msg {
		_, stale := m.parameters.ProjectDraftState(project.ProjectID)
		_, tree, hasDraft, err := m.svc.DeleteConditionalValue(context.Background(), project.ProjectID, groupKey, paramKey, valueLabel, publish)
		if err != nil {
			return messages.ParametersLoadedMsg{Project: project, Err: err, HasDraft: m.parameters.HasDraft(project.ProjectID), StaleDraft: stale}
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
			HasDraft:       hasDraft,
			StaleDraft:     !publish && hasDraft && stale,
			Revalidate:     false,
			SelectGroupKey: groupKey,
			SelectParamKey: paramKey,
		}
	}
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
	return func() tea.Msg {
		_, stale := m.parameters.ProjectDraftState(project.ProjectID)
		_, tree, hasDraft, err := m.svc.RenameParameter(context.Background(), project.ProjectID, groupKey, paramKey, nextParamKey, publish)
		if err != nil {
			return messages.ParametersLoadedMsg{Project: project, Err: err, HasDraft: m.parameters.HasDraft(project.ProjectID), StaleDraft: stale}
		}
		source := "draft"
		if publish {
			source = "firebase"
		}
		return messages.ParametersLoadedMsg{
			Project:     project,
			Tree:        tree,
			Source:      source,
			CacheSource: "cache",
			Err:         nil,
			HasDraft:    hasDraft,
			StaleDraft:  !publish && hasDraft && stale,
			Revalidate:  false,
		}
	}
}

func (m Model) renameGroupCmd(project core.Project, groupKey, nextGroupKey string, publish bool) tea.Cmd {
	return func() tea.Msg {
		_, stale := m.parameters.ProjectDraftState(project.ProjectID)
		_, tree, hasDraft, err := m.svc.RenameGroup(context.Background(), project.ProjectID, groupKey, nextGroupKey, publish)
		if err != nil {
			return messages.ParametersLoadedMsg{Project: project, Err: err, HasDraft: m.parameters.HasDraft(project.ProjectID), StaleDraft: stale}
		}
		source := "draft"
		if publish {
			source = "firebase"
		}
		return messages.ParametersLoadedMsg{
			Project:     project,
			Tree:        tree,
			Source:      source,
			CacheSource: "cache",
			Err:         nil,
			HasDraft:    hasDraft,
			StaleDraft:  !publish && hasDraft && stale,
			Revalidate:  false,
		}
	}
}

// moveParameterCmd moves move parameter cmd for Model and returns the resulting state or error.
func (m Model) moveParameterCmd(project core.Project, groupKey, paramKey, nextGroupKey string, publish bool) tea.Cmd {
	return func() tea.Msg {
		_, stale := m.parameters.ProjectDraftState(project.ProjectID)
		_, tree, hasDraft, err := m.svc.MoveParameter(context.Background(), project.ProjectID, groupKey, paramKey, nextGroupKey, publish)
		if err != nil {
			return messages.ParametersLoadedMsg{Project: project, Err: err, HasDraft: m.parameters.HasDraft(project.ProjectID), StaleDraft: stale}
		}
		source := "draft"
		if publish {
			source = "firebase"
		}
		return messages.ParametersLoadedMsg{
			Project:     project,
			Tree:        tree,
			Source:      source,
			CacheSource: "cache",
			Err:         nil,
			HasDraft:    hasDraft,
			StaleDraft:  !publish && hasDraft && stale,
			Revalidate:  false,
		}
	}
}

// moveGroupCmd moves move group cmd for Model and returns the resulting state or error.
func (m Model) moveGroupCmd(project core.Project, groupKey, nextGroupKey string, publish bool) tea.Cmd {
	return func() tea.Msg {
		_, stale := m.parameters.ProjectDraftState(project.ProjectID)
		_, tree, hasDraft, err := m.svc.MoveGroup(context.Background(), project.ProjectID, groupKey, nextGroupKey, publish)
		if err != nil {
			return messages.ParametersLoadedMsg{Project: project, Err: err, HasDraft: m.parameters.HasDraft(project.ProjectID), StaleDraft: stale}
		}
		source := "draft"
		if publish {
			source = "firebase"
		}
		return messages.ParametersLoadedMsg{
			Project:     project,
			Tree:        tree,
			Source:      source,
			CacheSource: "cache",
			Err:         nil,
			HasDraft:    hasDraft,
			StaleDraft:  !publish && hasDraft && stale,
			Revalidate:  false,
		}
	}
}

func (m Model) duplicateParameterNamedCmd(project core.Project, groupKey, paramKey, nextParamKey string, publish bool) tea.Cmd {
	return func() tea.Msg {
		_, stale := m.parameters.ProjectDraftState(project.ProjectID)
		_, tree, hasDraft, err := m.svc.DuplicateParameterNamed(context.Background(), project.ProjectID, groupKey, paramKey, nextParamKey, publish)
		if err != nil {
			return messages.ParametersLoadedMsg{Project: project, Err: err, HasDraft: m.parameters.HasDraft(project.ProjectID), StaleDraft: stale}
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
			HasDraft:       hasDraft,
			StaleDraft:     !publish && hasDraft && stale,
			Revalidate:     false,
			SelectGroupKey: groupKey,
			SelectParamKey: nextParamKey,
		}
	}
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

func dialogCanceledCmd() tea.Cmd {
	return func() tea.Msg {
		return messages.DialogCanceledMsg{}
	}
}

func detailsEditCanceledCmd(closeDetails bool) tea.Cmd {
	return func() tea.Msg {
		return messages.DetailsEditCanceledMsg{CloseDetails: closeDetails}
	}
}

func detailsInvalidFixCmd() tea.Cmd {
	return func() tea.Msg {
		return messages.DetailsInvalidFixMsg{}
	}
}

func detailsInvalidDiscardCmd(closeDetails bool) tea.Cmd {
	return func() tea.Msg {
		return messages.DetailsInvalidDiscardMsg{CloseDetails: closeDetails}
	}
}
