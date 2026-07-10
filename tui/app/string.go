package app

import (
	"context"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	corelog "github.com/yumauri/fbrcm/core/log"
	"github.com/yumauri/fbrcm/tui/panels"
)

func (m *Model) openStringInput() tea.Cmd {
	source := m.currentValueEditSource()
	anchor, ok := m.currentStringValueAnchor()
	if !ok {
		return nil
	}
	m.closeOverlays()
	m.valueEditSource = source
	var cmd tea.Cmd
	m.stringInput, cmd = m.stringInput.Open(anchor.X, anchor.Y, anchor.Width, anchor.MaxWidth, m.width, m.height, anchor.CurrentValue, anchor.FullWidth, anchor.Expanded)
	return cmd
}

func (m *Model) closeStringInput() {
	if !m.stringInput.IsOpen() {
		return
	}
	m.stringInput = m.stringInput.Close()
	m.valueEditSource = panels.None
}

func (m *Model) toggleStringInputMode() tea.Cmd {
	var cmd tea.Cmd
	m.stringInput, cmd = m.stringInput.ToggleExpanded()
	return cmd
}

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

func (m *Model) closeStringInputIfOrphaned() {
	if !m.stringInput.IsOpen() {
		return
	}
	if _, ok := m.currentStringValueAnchor(); ok {
		return
	}
	m.closeStringInput()
}

func (m *Model) openStringValueDialog(project core.Project, groupKey, paramKey, valueLabel, nextValue string) {
	m.openValueEditDialog(project, func() ([]string, error) {
		return m.stringValueDialogBody(project, groupKey, paramKey, valueLabel, nextValue)
	}, func(err error) {
		corelog.For("tui.string").Error("string value preview failed", "project_id", project.ProjectID, "group", groupKey, "param", paramKey, "value_label", valueLabel, "err", err)
	}, m.setStringParameterValueCmd(project, groupKey, paramKey, valueLabel, nextValue, true), m.setStringParameterValueCmd(project, groupKey, paramKey, valueLabel, nextValue, false))
}

func (m Model) stringValueDialogBody(project core.Project, groupKey, paramKey, valueLabel, nextValue string) ([]string, error) {
	return m.valueEditDialogBody(project, func() (*core.ParametersCache, []byte, error) {
		return m.svc.PreviewSetStringParameterValue(project.ProjectID, groupKey, paramKey, valueLabel, nextValue)
	})
}

func (m Model) setStringParameterValueCmd(project core.Project, groupKey, paramKey, valueLabel, nextValue string, publish bool) tea.Cmd {
	return m.runSetParameterValueCmd(project, groupKey, paramKey, valueLabel, publish, func(ctx context.Context) (*core.ParametersTree, bool, error) {
		_, tree, hasDraft, err := m.svc.SetStringParameterValue(ctx, project.ProjectID, groupKey, paramKey, valueLabel, nextValue, publish)
		return tree, hasDraft, err
	})
}
