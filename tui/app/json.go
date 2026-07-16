package app

import (
	"context"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	corelog "github.com/yumauri/fbrcm/core/log"
	"github.com/yumauri/fbrcm/tui/panels"
)

func (m *Model) openJSONInput() tea.Cmd {
	source := m.currentValueEditSource()
	anchor, ok := m.currentJSONValueAnchor()
	if !ok {
		return nil
	}
	m.closeOverlays()
	m.valueEditSource = source
	var cmd tea.Cmd
	m.jsonInput, cmd = m.jsonInput.Open(m.width, m.height, anchor.CurrentValue)
	return cmd
}

func (m *Model) closeJSONInput() {
	if !m.jsonInput.IsOpen() {
		return
	}
	m.jsonInput = m.jsonInput.Close()
	m.valueEditSource = panels.None
}

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
		if source == panels.Details {
			m.finishConditionalValueAdd()
		}
		return nil
	}
	if source == panels.Details {
		m.details = m.details.SetSelectedValue(nextValue)
		m.finishConditionalValueAdd()
		return nil
	}
	if m.parameters.HasDraft(anchor.Project.ProjectID) {
		return m.setJSONParameterValueCmd(anchor.Project, anchor.GroupKey, anchor.ParamKey, anchor.ValueLabel, nextValue, false)
	}
	m.openJSONValueDialog(anchor.Project, anchor.GroupKey, anchor.ParamKey, anchor.ValueLabel, nextValue)
	return nil
}

func (m *Model) closeJSONInputIfOrphaned() {
	if !m.jsonInput.IsOpen() {
		return
	}
	if _, ok := m.currentJSONValueAnchor(); ok {
		return
	}
	m.closeJSONInput()
}

func (m *Model) openJSONValueDialog(project core.Project, groupKey, paramKey, valueLabel, nextValue string) {
	m.openValueEditDialog(project, func() ([]string, error) {
		return m.jsonValueDialogBody(project, groupKey, paramKey, valueLabel, nextValue)
	}, func(err error) {
		corelog.For("tui.json").Error("json value preview failed", "project_id", project.ProjectID, "group", groupKey, "param", paramKey, "value_label", valueLabel, "err", err)
	}, m.setJSONParameterValueCmd(project, groupKey, paramKey, valueLabel, nextValue, true), m.setJSONParameterValueCmd(project, groupKey, paramKey, valueLabel, nextValue, false))
}

func (m Model) jsonValueDialogBody(project core.Project, groupKey, paramKey, valueLabel, nextValue string) ([]string, error) {
	return m.valueEditDialogBody(project, func() (*core.ParametersCache, []byte, error) {
		return m.svc.PreviewSetJSONParameterValue(project.ProjectID, groupKey, paramKey, valueLabel, nextValue)
	})
}

func (m Model) setJSONParameterValueCmd(project core.Project, groupKey, paramKey, valueLabel, nextValue string, publish bool) tea.Cmd {
	return m.runSetParameterValueCmd(project, groupKey, paramKey, valueLabel, publish, func(ctx context.Context) (*core.ParametersTree, bool, error) {
		_, tree, hasDraft, err := m.svc.SetJSONParameterValue(ctx, project.ProjectID, groupKey, paramKey, valueLabel, nextValue, publish)
		return tree, hasDraft, err
	})
}
