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

// openJSONInput opens open jsoninput for Model and returns the resulting state or error.
func (m *Model) openJSONInput() tea.Cmd {
	if m.active == panels.Details {
		m.valueEditSource = panels.Details
	} else {
		m.valueEditSource = panels.Parameters
	}
	anchor, ok := m.currentJSONValueAnchor()
	if !ok {
		return nil
	}
	m.closeDialog(false)
	m.closeBoolPicker()
	m.closeNumberInput()
	m.closeStringInput()
	m.closeMoveParam()
	m.closeRenameInput()
	var cmd tea.Cmd
	m.jsonInput, cmd = m.jsonInput.Open(m.width, m.height, anchor.CurrentValue)
	return cmd
}

// closeJSONInput closes close jsoninput for Model and returns the resulting state or error.
func (m *Model) closeJSONInput() {
	if !m.jsonInput.IsOpen() {
		return
	}
	m.jsonInput = m.jsonInput.Close()
	m.valueEditSource = panels.None
}

// submitJSONInput handles submit jsoninput for Model and returns the resulting state or error.
func (m *Model) submitJSONInput() tea.Cmd {
	anchor, ok := m.currentJSONValueAnchor()
	if !ok {
		m.closeJSONInput()
		return nil
	}
	nextValue, valid := m.jsonInput.CompactedValue()
	if !valid {
		return nil
	}
	source := m.valueEditSource
	m.closeJSONInput()
	if nextValue == anchor.CurrentValue {
		return nil
	}
	if source == panels.Details {
		m.details = m.details.SetSelectedValue(nextValue)
		return nil
	}
	if m.parameters.HasDraft(anchor.Project.ProjectID) {
		return m.setJSONParameterValueCmd(anchor.Project, anchor.GroupKey, anchor.ParamKey, anchor.ValueLabel, nextValue, false)
	}
	m.openJSONValueDialog(anchor.Project, anchor.GroupKey, anchor.ParamKey, anchor.ValueLabel, nextValue)
	return nil
}

// closeJSONInputIfOrphaned closes close jsoninput if orphaned for Model and returns the resulting state or error.
func (m *Model) closeJSONInputIfOrphaned() {
	if !m.jsonInput.IsOpen() {
		return
	}
	if _, ok := m.currentJSONValueAnchor(); ok {
		return
	}
	m.closeJSONInput()
}

// openJSONValueDialog opens open jsonvalue dialog for Model and returns the resulting state or error.
func (m *Model) openJSONValueDialog(project core.Project, groupKey, paramKey, valueLabel, nextValue string) {
	body, err := m.jsonValueDialogBody(project, groupKey, paramKey, valueLabel, nextValue)
	if err != nil {
		corelog.For("tui.json").Error("json value preview failed", "project_id", project.ProjectID, "group", groupKey, "param", paramKey, "value_label", valueLabel, "err", err)
		m.openErrorDialog("Edit Value Failed", project, err.Error())
		return
	}
	m.dialog = m.dialog.Open(dialogcmp.Config{
		Title: "Edit Value?",
		Body:  body,
		Buttons: []dialogcmp.Button{
			{Label: "Apply", Variant: dialogcmp.ButtonVariantDanger, OnPress: m.setJSONParameterValueCmd(project, groupKey, paramKey, valueLabel, nextValue, true)},
			{Label: "Draft", Variant: dialogcmp.ButtonVariantAccent, OnPress: m.setJSONParameterValueCmd(project, groupKey, paramKey, valueLabel, nextValue, false)},
			{Label: "Cancel", Variant: dialogcmp.ButtonVariantAccent, OnPress: dialogCanceledCmd()},
		},
	})
}

// jsonValueDialogBody handles json value dialog body for Model and returns the resulting state or error.
func (m Model) jsonValueDialogBody(project core.Project, groupKey, paramKey, valueLabel, nextValue string) ([]string, error) {
	cache, finalRaw, err := m.svc.PreviewSetJSONParameterValue(project.ProjectID, groupKey, paramKey, valueLabel, nextValue)
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

// setJSONParameterValueCmd sets set jsonparameter value cmd for Model and returns the resulting state or error.
func (m Model) setJSONParameterValueCmd(project core.Project, groupKey, paramKey, valueLabel, nextValue string, publish bool) tea.Cmd {
	return func() tea.Msg {
		_, stale := m.parameters.ProjectDraftState(project.ProjectID)
		_, tree, hasDraft, err := m.svc.SetJSONParameterValue(context.Background(), project.ProjectID, groupKey, paramKey, valueLabel, nextValue, publish)
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
