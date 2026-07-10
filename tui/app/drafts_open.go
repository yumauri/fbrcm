package app

import (
	"github.com/yumauri/fbrcm/core"
	corelog "github.com/yumauri/fbrcm/core/log"
	dialogcmp "github.com/yumauri/fbrcm/tui/components/dialog"
)

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

func (m *Model) openPreviewDialog(project core.Project, title, errorTitle string, bodyFn dialogBodyFunc, logErr func(error), buttons []dialogcmp.Button) {
	body, err := bodyFn()
	if err != nil {
		logErr(err)
		m.openErrorDialog(errorTitle, project, err.Error())
		return
	}
	m.dialog = m.dialog.Open(dialogcmp.Config{
		Title:   title,
		Body:    body,
		Buttons: buttons,
	})
}

// openDeleteConditionalValueDialog opens delete conditional value dialog.
func (m *Model) openDeleteConditionalValueDialog(project core.Project, groupKey, paramKey, valueLabel string) {
	m.openPreviewDialog(project, "Delete Conditional Value?", "Delete Conditional Value Failed", func() ([]string, error) {
		return m.deleteConditionalValueDialogBody(project, groupKey, paramKey, valueLabel)
	}, func(err error) {
		corelog.For("tui.delete").Error("delete conditional value preview failed", "project_id", project.ProjectID, "group", groupKey, "param", paramKey, "value_label", valueLabel, "err", err)
	}, []dialogcmp.Button{
		{Label: "Delete", Variant: dialogcmp.ButtonVariantDanger, OnPress: m.deleteConditionalValueCmd(project, groupKey, paramKey, valueLabel, true)},
		{Label: "Draft", Variant: dialogcmp.ButtonVariantAccent, OnPress: m.deleteConditionalValueCmd(project, groupKey, paramKey, valueLabel, false)},
		{Label: "Cancel", Variant: dialogcmp.ButtonVariantAccent, OnPress: dialogCanceledCmd()},
	})
}

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

func (m *Model) openRenameDialog(project core.Project, groupKey, paramKey, nextParamKey string) {
	m.openPreviewDialog(project, "Rename Parameter?", "Rename Failed", func() ([]string, error) {
		return m.renameDialogBody(project, groupKey, paramKey, nextParamKey)
	}, func(err error) {
		corelog.For("tui.rename").Error("rename preview failed", "project_id", project.ProjectID, "group", groupKey, "param", paramKey, "next_param", nextParamKey, "err", err)
	}, []dialogcmp.Button{
		{Label: "Rename", Variant: dialogcmp.ButtonVariantDanger, OnPress: m.renameParameterCmd(project, groupKey, paramKey, nextParamKey, true)},
		{Label: "Draft", Variant: dialogcmp.ButtonVariantAccent, OnPress: m.renameParameterCmd(project, groupKey, paramKey, nextParamKey, false)},
		{Label: "Cancel", Variant: dialogcmp.ButtonVariantAccent, OnPress: dialogCanceledCmd()},
	})
}

func (m *Model) openRenameGroupDialog(project core.Project, groupKey, nextGroupKey string) {
	m.openPreviewDialog(project, "Rename Group?", "Rename Group Failed", func() ([]string, error) {
		return m.renameGroupDialogBody(project, groupKey, nextGroupKey)
	}, func(err error) {
		corelog.For("tui.rename").Error("rename group preview failed", "project_id", project.ProjectID, "group", groupKey, "next_group", nextGroupKey, "err", err)
	}, []dialogcmp.Button{
		{Label: "Rename", Variant: dialogcmp.ButtonVariantDanger, OnPress: m.renameGroupCmd(project, groupKey, nextGroupKey, true)},
		{Label: "Draft", Variant: dialogcmp.ButtonVariantAccent, OnPress: m.renameGroupCmd(project, groupKey, nextGroupKey, false)},
		{Label: "Cancel", Variant: dialogcmp.ButtonVariantAccent, OnPress: dialogCanceledCmd()},
	})
}

func (m *Model) openMoveDialog(project core.Project, groupKey, paramKey, nextGroupKey string) {
	m.openPreviewDialog(project, "Move Parameter?", "Move Parameter Failed", func() ([]string, error) {
		return m.moveDialogBody(project, groupKey, paramKey, nextGroupKey)
	}, func(err error) {
		corelog.For("tui.move").Error("move parameter preview failed", "project_id", project.ProjectID, "group", groupKey, "param", paramKey, "next_group", nextGroupKey, "err", err)
	}, []dialogcmp.Button{
		{Label: "Move", Variant: dialogcmp.ButtonVariantDanger, OnPress: m.moveParameterCmd(project, groupKey, paramKey, nextGroupKey, true)},
		{Label: "Draft", Variant: dialogcmp.ButtonVariantAccent, OnPress: m.moveParameterCmd(project, groupKey, paramKey, nextGroupKey, false)},
		{Label: "Cancel", Variant: dialogcmp.ButtonVariantAccent, OnPress: dialogCanceledCmd()},
	})
}

func (m *Model) openMoveGroupDialog(project core.Project, groupKey, nextGroupKey string) {
	m.openPreviewDialog(project, "Move Group?", "Move Group Failed", func() ([]string, error) {
		return m.moveGroupDialogBody(project, groupKey, nextGroupKey)
	}, func(err error) {
		corelog.For("tui.move").Error("move group preview failed", "project_id", project.ProjectID, "group", groupKey, "next_group", nextGroupKey, "err", err)
	}, []dialogcmp.Button{
		{Label: "Move", Variant: dialogcmp.ButtonVariantDanger, OnPress: m.moveGroupCmd(project, groupKey, nextGroupKey, true)},
		{Label: "Draft", Variant: dialogcmp.ButtonVariantAccent, OnPress: m.moveGroupCmd(project, groupKey, nextGroupKey, false)},
		{Label: "Cancel", Variant: dialogcmp.ButtonVariantAccent, OnPress: dialogCanceledCmd()},
	})
}

func (m *Model) openDuplicateDialog(project core.Project, groupKey, paramKey, nextParamKey string) {
	m.openPreviewDialog(project, "Duplicate Parameter?", "Duplicate Failed", func() ([]string, error) {
		return m.duplicateDialogBody(project, groupKey, paramKey, nextParamKey)
	}, func(err error) {
		corelog.For("tui.duplicate").Error("duplicate parameter preview failed", "project_id", project.ProjectID, "group", groupKey, "param", paramKey, "next_param", nextParamKey, "err", err)
	}, []dialogcmp.Button{
		{Label: "Duplicate", Variant: dialogcmp.ButtonVariantDanger, OnPress: m.duplicateParameterNamedCmd(project, groupKey, paramKey, nextParamKey, true)},
		{Label: "Draft", Variant: dialogcmp.ButtonVariantAccent, OnPress: m.duplicateParameterNamedCmd(project, groupKey, paramKey, nextParamKey, false)},
		{Label: "Cancel", Variant: dialogcmp.ButtonVariantAccent, OnPress: dialogCanceledCmd()},
	})
}

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
