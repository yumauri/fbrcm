package app

import (
	"context"
	"fmt"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	corelog "github.com/yumauri/fbrcm/core/log"
	dialogcmp "github.com/yumauri/fbrcm/tui/components/dialog"
	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/panels"
)

// openNumberInput opens open number input for Model and returns the resulting state or error.
func (m *Model) openNumberInput() tea.Cmd {
	if m.active == panels.Details {
		m.valueEditSource = panels.Details
	} else {
		m.valueEditSource = panels.Parameters
	}
	anchor, ok := m.currentNumberValueAnchor()
	if !ok {
		return nil
	}
	m.closeDialog(false)
	m.closeJSONInput()
	m.closeBoolPicker()
	m.closeStringInput()
	m.closeMoveParam()
	m.closeRenameInput()
	var cmd tea.Cmd
	m.numberInput, cmd = m.numberInput.Open(anchor.X, anchor.Y, anchor.Width, anchor.MaxWidth, anchor.CurrentValue)
	if m.valueEditSource == panels.Details {
		m.details = m.details.SetValuesInvalid(!m.numberInput.Valid())
	}
	return cmd
}

// closeNumberInput closes close number input for Model and returns the resulting state or error.
func (m *Model) closeNumberInput() {
	if !m.numberInput.IsOpen() {
		return
	}
	m.numberInput = m.numberInput.Close()
	m.valueEditSource = panels.None
	m.details = m.details.SetValuesInvalid(false)
}

// submitNumberInput handles submit number input for Model and returns the resulting state or error.
func (m *Model) submitNumberInput() tea.Cmd {
	anchor, ok := m.currentNumberValueAnchor()
	if !ok {
		m.closeNumberInput()
		return nil
	}
	if !m.numberInput.Valid() {
		return nil
	}
	nextValue := m.numberInput.Value()
	source := m.valueEditSource
	m.closeNumberInput()
	if nextValue == anchor.CurrentValue {
		return nil
	}
	if source == panels.Details {
		m.details = m.details.SetSelectedValue(nextValue)
		return nil
	}
	if m.parameters.HasDraft(anchor.Project.ProjectID) {
		return m.setNumberParameterValueCmd(anchor.Project, anchor.GroupKey, anchor.ParamKey, anchor.ValueLabel, nextValue, false)
	}
	m.openNumberValueDialog(anchor.Project, anchor.GroupKey, anchor.ParamKey, anchor.ValueLabel, nextValue)
	return nil
}

// closeNumberInputIfOrphaned closes close number input if orphaned for Model and returns the resulting state or error.
func (m *Model) closeNumberInputIfOrphaned() {
	if !m.numberInput.IsOpen() {
		return
	}
	if _, ok := m.currentNumberValueAnchor(); ok {
		return
	}
	m.closeNumberInput()
}

// openNumberValueDialog opens open number value dialog for Model and returns the resulting state or error.
func (m *Model) openNumberValueDialog(project core.Project, groupKey, paramKey, valueLabel, nextValue string) {
	body, err := m.numberValueDialogBody(project, groupKey, paramKey, valueLabel, nextValue)
	if err != nil {
		corelog.For("tui.number").Error("number value preview failed", "project_id", project.ProjectID, "group", groupKey, "param", paramKey, "value_label", valueLabel, "next_value", nextValue, "err", err)
		m.openErrorDialog("Edit Value Failed", project, err.Error())
		return
	}
	m.dialog = m.dialog.Open(dialogcmp.Config{
		Title: "Edit Value?",
		Body:  body,
		Buttons: []dialogcmp.Button{
			{Label: "Apply", Variant: dialogcmp.ButtonVariantDanger, OnPress: m.setNumberParameterValueCmd(project, groupKey, paramKey, valueLabel, nextValue, true)},
			{Label: "Draft", Variant: dialogcmp.ButtonVariantAccent, OnPress: m.setNumberParameterValueCmd(project, groupKey, paramKey, valueLabel, nextValue, false)},
			{Label: "Cancel", Variant: dialogcmp.ButtonVariantAccent, OnPress: dialogCanceledCmd()},
		},
	})
}

// numberValueDialogBody handles number value dialog body for Model and returns the resulting state or error.
func (m Model) numberValueDialogBody(project core.Project, groupKey, paramKey, valueLabel, nextValue string) ([]string, error) {
	cache, finalRaw, err := m.svc.PreviewSetNumberParameterValue(project.ProjectID, groupKey, paramKey, valueLabel, nextValue)
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
		"Edit value or draft changes?",
	}
	if !hasChanges {
		return nil, fmt.Errorf("parameter value not changed")
	}
	lines = append(lines, "")
	lines = append(lines, dialogDiffLines(diffText)...)
	return lines, nil
}

// setNumberParameterValueCmd sets set number parameter value cmd for Model and returns the resulting state or error.
func (m Model) setNumberParameterValueCmd(project core.Project, groupKey, paramKey, valueLabel, nextValue string, publish bool) tea.Cmd {
	return func() tea.Msg {
		_, stale := m.parameters.ProjectDraftState(project.ProjectID)
		_, tree, hasDraft, err := m.svc.SetNumberParameterValue(context.Background(), project.ProjectID, groupKey, paramKey, valueLabel, nextValue, publish)
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
