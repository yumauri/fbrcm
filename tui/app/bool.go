package app

import (
	"context"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	corelog "github.com/yumauri/fbrcm/core/log"
	"github.com/yumauri/fbrcm/tui/panels"
)

func (m *Model) openBoolPicker() tea.Cmd {
	source := m.currentValueEditSource()
	anchor, ok := m.currentBoolValueAnchor()
	if !ok {
		return nil
	}
	m.closeOverlays()
	m.valueEditSource = source
	m.boolPicker = m.boolPicker.Open(anchor.X, anchor.Y, anchor.Value)
	return nil
}

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
		m.finishConditionalValueAdd()
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

func (m *Model) closeBoolPickerIfOrphaned() {
	if !m.boolPicker.IsOpen() {
		return
	}
	if _, ok := m.currentBoolValueAnchor(); ok {
		return
	}
	m.closeBoolPicker()
}

func (m *Model) openBoolValueDialog(project core.Project, groupKey, paramKey, valueLabel string, nextValue bool) {
	m.openValueEditDialog(project, func() ([]string, error) {
		return m.boolValueDialogBody(project, groupKey, paramKey, valueLabel, nextValue)
	}, func(err error) {
		corelog.For("tui.bool").Error("boolean value preview failed", "project_id", project.ProjectID, "group", groupKey, "param", paramKey, "value_label", valueLabel, "next_value", nextValue, "err", err)
	}, m.setBooleanParameterValueCmd(project, groupKey, paramKey, valueLabel, nextValue, true), m.setBooleanParameterValueCmd(project, groupKey, paramKey, valueLabel, nextValue, false))
}

func (m Model) boolValueDialogBody(project core.Project, groupKey, paramKey, valueLabel string, nextValue bool) ([]string, error) {
	return m.valueEditDialogBody(project, func() (*core.ParametersCache, []byte, error) {
		return m.svc.PreviewSetBooleanParameterValue(project.ProjectID, groupKey, paramKey, valueLabel, nextValue)
	})
}

func (m Model) setBooleanParameterValueCmd(project core.Project, groupKey, paramKey, valueLabel string, nextValue, publish bool) tea.Cmd {
	return m.runSetParameterValueCmd(project, groupKey, paramKey, valueLabel, publish, func(ctx context.Context) (*core.ParametersTree, bool, error) {
		_, tree, hasDraft, err := m.svc.SetBooleanParameterValue(ctx, project.ProjectID, groupKey, paramKey, valueLabel, nextValue, publish)
		return tree, hasDraft, err
	})
}
