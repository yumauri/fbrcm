package app

import (
	"context"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	corelog "github.com/yumauri/fbrcm/core/log"
	"github.com/yumauri/fbrcm/tui/panels"
)

func (m *Model) openNumberInput() tea.Cmd {
	source := m.currentValueEditSource()
	anchor, ok := m.currentNumberValueAnchor()
	if !ok {
		return nil
	}
	m.closeOverlays()
	m.valueEditSource = source
	var cmd tea.Cmd
	m.numberInput, cmd = m.numberInput.Open(anchor.X, anchor.Y, anchor.Width, anchor.MaxWidth, anchor.CurrentValue)
	if m.valueEditSource == panels.Details {
		m.details = m.details.SetValuesInvalid(!m.numberInput.Valid())
	}
	return cmd
}

func (m *Model) closeNumberInput() {
	if !m.numberInput.IsOpen() {
		return
	}
	m.numberInput = m.numberInput.Close()
	m.valueEditSource = panels.None
	m.details = m.details.SetValuesInvalid(false)
}

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

func (m *Model) closeNumberInputIfOrphaned() {
	if !m.numberInput.IsOpen() {
		return
	}
	if _, ok := m.currentNumberValueAnchor(); ok {
		return
	}
	m.closeNumberInput()
}

func (m *Model) openNumberValueDialog(project core.Project, groupKey, paramKey, valueLabel, nextValue string) {
	m.openValueEditDialog(project, func() ([]string, error) {
		return m.numberValueDialogBody(project, groupKey, paramKey, valueLabel, nextValue)
	}, func(err error) {
		corelog.For("tui.number").Error("number value preview failed", "project_id", project.ProjectID, "group", groupKey, "param", paramKey, "value_label", valueLabel, "next_value", nextValue, "err", err)
	}, m.setNumberParameterValueCmd(project, groupKey, paramKey, valueLabel, nextValue, true), m.setNumberParameterValueCmd(project, groupKey, paramKey, valueLabel, nextValue, false))
}

func (m Model) numberValueDialogBody(project core.Project, groupKey, paramKey, valueLabel, nextValue string) ([]string, error) {
	return m.valueEditDialogBody(project, func() (*core.ParametersCache, []byte, error) {
		return m.svc.PreviewSetNumberParameterValue(project.ProjectID, groupKey, paramKey, valueLabel, nextValue)
	})
}

func (m Model) setNumberParameterValueCmd(project core.Project, groupKey, paramKey, valueLabel, nextValue string, publish bool) tea.Cmd {
	return m.runSetParameterValueCmd(project, groupKey, paramKey, valueLabel, publish, func(ctx context.Context) (*core.ParametersTree, bool, error) {
		_, tree, hasDraft, err := m.svc.SetNumberParameterValue(ctx, project.ProjectID, groupKey, paramKey, valueLabel, nextValue, publish)
		return tree, hasDraft, err
	})
}
