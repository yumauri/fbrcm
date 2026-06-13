package app

import (
	"context"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	corelog "github.com/yumauri/fbrcm/core/log"
	dialogcmp "github.com/yumauri/fbrcm/tui/components/dialog"
	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/panels"
)

// openBoolPicker opens open bool picker for Model and returns the resulting state or error.
func (m *Model) openBoolPicker() tea.Cmd {
	if m.active == panels.Details {
		m.valueEditSource = panels.Details
	} else {
		m.valueEditSource = panels.Parameters
	}
	anchor, ok := m.currentBoolValueAnchor()
	if !ok {
		return nil
	}
	m.closeDialog(false)
	m.closeJSONInput()
	m.closeNumberInput()
	m.closeStringInput()
	m.closeRenameInput()
	m.closeMoveParam()
	m.boolPicker = m.boolPicker.Open(anchor.X, anchor.Y, anchor.Value)
	return nil
}

// closeBoolPicker closes close bool picker for Model and returns the resulting state or error.
func (m *Model) closeBoolPicker() {
	if !m.boolPicker.IsOpen() {
		return
	}
	m.boolPicker = m.boolPicker.Close()
	m.valueEditSource = panels.None
}

func (m *Model) submitBoolPicker() tea.Cmd {
	anchor, ok := m.currentBoolValueAnchor()
	if !ok {
		m.closeBoolPicker()
		return nil
	}
	nextValue, ok := m.boolPicker.Current()
	source := m.valueEditSource
	m.closeBoolPicker()
	if !ok {
		return nil
	}
	if source == panels.Details {
		nextRaw := "false"
		if nextValue {
			nextRaw = "true"
		}
		if nextRaw != anchor.CurrentValue {
			m.details = m.details.SetSelectedValue(nextRaw)
		}
		return nil
	}
	if nextValue == anchor.Value {
		return nil
	}
	if m.parameters.HasDraft(anchor.Project.ProjectID) {
		return m.setBooleanParameterValueCmd(anchor.Project, anchor.GroupKey, anchor.ParamKey, anchor.ValueLabel, nextValue, false)
	}
	m.openBoolValueDialog(anchor.Project, anchor.GroupKey, anchor.ParamKey, anchor.ValueLabel, nextValue)
	return nil
}

// closeBoolPickerIfOrphaned closes close bool picker if orphaned for Model and returns the resulting state or error.
func (m *Model) closeBoolPickerIfOrphaned() {
	if !m.boolPicker.IsOpen() {
		return
	}
	if _, ok := m.currentBoolValueAnchor(); ok {
		return
	}
	m.closeBoolPicker()
}

// openBoolValueDialog opens open bool value dialog for Model and returns the resulting state or error.
func (m *Model) openBoolValueDialog(project core.Project, groupKey, paramKey, valueLabel string, nextValue bool) {
	body, err := m.boolValueDialogBody(project, groupKey, paramKey, valueLabel, nextValue)
	if err != nil {
		corelog.For("tui.bool").Error("boolean value preview failed", "project_id", project.ProjectID, "group", groupKey, "param", paramKey, "value_label", valueLabel, "next_value", nextValue, "err", err)
		m.openErrorDialog("Edit Value Failed", project, err.Error())
		return
	}
	m.dialog = m.dialog.Open(dialogcmp.Config{
		Title: "Edit Value?",
		Body:  body,
		Buttons: []dialogcmp.Button{
			{Label: "Apply", Variant: dialogcmp.ButtonVariantDanger, OnPress: m.setBooleanParameterValueCmd(project, groupKey, paramKey, valueLabel, nextValue, true)},
			{Label: "Draft", Variant: dialogcmp.ButtonVariantAccent, OnPress: m.setBooleanParameterValueCmd(project, groupKey, paramKey, valueLabel, nextValue, false)},
			{Label: "Cancel", Variant: dialogcmp.ButtonVariantAccent, OnPress: dialogCanceledCmd()},
		},
	})
}

func (m Model) boolValueDialogBody(project core.Project, groupKey, paramKey, valueLabel string, nextValue bool) ([]string, error) {
	return m.valueEditDialogBody(project, func() (*core.ParametersCache, []byte, error) {
		return m.svc.PreviewSetBooleanParameterValue(project.ProjectID, groupKey, paramKey, valueLabel, nextValue)
	})
}

// setBooleanParameterValueCmd sets set boolean parameter value cmd for Model and returns the resulting state or error.
func (m Model) setBooleanParameterValueCmd(project core.Project, groupKey, paramKey, valueLabel string, nextValue, publish bool) tea.Cmd {
	return func() tea.Msg {
		_, stale := m.parameters.ProjectDraftState(project.ProjectID)
		_, tree, hasDraft, err := m.svc.SetBooleanParameterValue(context.Background(), project.ProjectID, groupKey, paramKey, valueLabel, nextValue, publish)
		if err != nil {
			return messages.ParametersLoadedMsg{Project: project, Err: err, HasDraft: m.parameters.HasDraft(project.ProjectID), StaleDraft: stale}
		}
		return m.valueEditLoadedMsg(project, groupKey, paramKey, tree, hasDraft, stale, publish)
	}
}
