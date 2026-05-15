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

// openStringInput opens open string input for Model and returns the resulting state or error.
func (m *Model) openStringInput() tea.Cmd {
	if m.active == panels.Details {
		m.valueEditSource = panels.Details
	} else {
		m.valueEditSource = panels.Parameters
	}
	anchor, ok := m.currentStringValueAnchor()
	if !ok {
		return nil
	}
	m.closeDialog(false)
	m.closeJSONInput()
	m.closeBoolPicker()
	m.closeNumberInput()
	m.closeMoveParam()
	m.closeRenameInput()
	var cmd tea.Cmd
	m.stringInput, cmd = m.stringInput.Open(anchor.X, anchor.Y, anchor.Width, anchor.MaxWidth, m.width, m.height, anchor.CurrentValue, anchor.FullWidth, anchor.Expanded)
	return cmd
}

// closeStringInput closes close string input for Model and returns the resulting state or error.
func (m *Model) closeStringInput() {
	if !m.stringInput.IsOpen() {
		return
	}
	m.stringInput = m.stringInput.Close()
	m.valueEditSource = panels.None
}

// toggleStringInputMode toggles toggle string input mode for Model and returns the resulting state or error.
func (m *Model) toggleStringInputMode() tea.Cmd {
	var cmd tea.Cmd
	m.stringInput, cmd = m.stringInput.ToggleExpanded()
	return cmd
}

// submitStringInput handles submit string input for Model and returns the resulting state or error.
func (m *Model) submitStringInput() tea.Cmd {
	anchor, ok := m.currentStringValueAnchor()
	if !ok {
		m.closeStringInput()
		return nil
	}
	nextValue := m.stringInput.Value()
	source := m.valueEditSource
	m.closeStringInput()
	if nextValue == anchor.CurrentValue {
		return nil
	}
	if source == panels.Details {
		m.details = m.details.SetSelectedValue(nextValue)
		return nil
	}
	if m.parameters.HasDraft(anchor.Project.ProjectID) {
		return m.setStringParameterValueCmd(anchor.Project, anchor.GroupKey, anchor.ParamKey, anchor.ValueLabel, nextValue, false)
	}
	m.openStringValueDialog(anchor.Project, anchor.GroupKey, anchor.ParamKey, anchor.ValueLabel, nextValue)
	return nil
}

// closeStringInputIfOrphaned closes close string input if orphaned for Model and returns the resulting state or error.
func (m *Model) closeStringInputIfOrphaned() {
	if !m.stringInput.IsOpen() {
		return
	}
	if _, ok := m.currentStringValueAnchor(); ok {
		return
	}
	m.closeStringInput()
}

// openStringValueDialog opens open string value dialog for Model and returns the resulting state or error.
func (m *Model) openStringValueDialog(project core.Project, groupKey, paramKey, valueLabel, nextValue string) {
	body, err := m.stringValueDialogBody(project, groupKey, paramKey, valueLabel, nextValue)
	if err != nil {
		corelog.For("tui.string").Error("string value preview failed", "project_id", project.ProjectID, "group", groupKey, "param", paramKey, "value_label", valueLabel, "err", err)
		m.openErrorDialog("Edit Value Failed", project, err.Error())
		return
	}
	m.dialog = m.dialog.Open(dialogcmp.Config{
		Title: "Edit Value?",
		Body:  body,
		Buttons: []dialogcmp.Button{
			{Label: "Apply", Variant: dialogcmp.ButtonVariantDanger, OnPress: m.setStringParameterValueCmd(project, groupKey, paramKey, valueLabel, nextValue, true)},
			{Label: "Draft", Variant: dialogcmp.ButtonVariantAccent, OnPress: m.setStringParameterValueCmd(project, groupKey, paramKey, valueLabel, nextValue, false)},
			{Label: "Cancel", Variant: dialogcmp.ButtonVariantAccent, OnPress: dialogCanceledCmd()},
		},
	})
}

// stringValueDialogBody handles string value dialog body for Model and returns the resulting state or error.
func (m Model) stringValueDialogBody(project core.Project, groupKey, paramKey, valueLabel, nextValue string) ([]string, error) {
	cache, finalRaw, err := m.svc.PreviewSetStringParameterValue(project.ProjectID, groupKey, paramKey, valueLabel, nextValue)
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

// setStringParameterValueCmd sets set string parameter value cmd for Model and returns the resulting state or error.
func (m Model) setStringParameterValueCmd(project core.Project, groupKey, paramKey, valueLabel, nextValue string, publish bool) tea.Cmd {
	return func() tea.Msg {
		_, stale := m.parameters.ProjectDraftState(project.ProjectID)
		_, tree, hasDraft, err := m.svc.SetStringParameterValue(context.Background(), project.ProjectID, groupKey, paramKey, valueLabel, nextValue, publish)
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
