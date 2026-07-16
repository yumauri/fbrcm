package app

import (
	tea "charm.land/bubbletea/v2"

	moveparam "github.com/yumauri/fbrcm/tui/components/moveparam"
	"github.com/yumauri/fbrcm/tui/styles"
)

func (m *Model) openAddConditionalValue() tea.Cmd {
	data := m.details.Data()
	if data == nil {
		return nil
	}
	conditions := m.details.AvailableConditions()
	if len(conditions) == 0 {
		message := "All project conditions already have values for this parameter."
		if len(data.Conditions) == 0 {
			message = "Project has no conditions."
		}
		m.openErrorDialog("Add Conditional Value Failed", data.Project, message)
		return nil
	}
	x, y, ok := m.details.ConditionalValuePickerPosition()
	if !ok {
		return nil
	}
	options := make([]moveparam.Option, 0, len(conditions))
	for _, condition := range conditions {
		foreground := styles.ConditionLipglossColor(condition.Color)
		if condition.Color == "" {
			foreground = styles.PaletteSlateBright
		}
		options = append(options, moveparam.Option{
			Key:                    condition.Name,
			Label:                  "● " + condition.Name,
			Foreground:             foreground,
			KeepForegroundOnSelect: true,
		})
	}
	m.closeOverlays()
	m.conditionalAdd = &conditionalValueAddSession{}
	m.moveParam = m.moveParam.OpenOptions(x, y, "+ Add conditional value", options, 0)
	return nil
}

func (m *Model) submitConditionalValueOption() (tea.Cmd, bool) {
	if m.conditionalAdd == nil || m.conditionalAdd.condition != "" {
		return nil, false
	}
	option, ok := m.moveParam.Current()
	m.moveParam = m.moveParam.Close()
	if !ok {
		m.conditionalAdd = nil
		return nil, true
	}
	next, added := m.details.AddConditionalValue(option.Key)
	if !added {
		m.conditionalAdd = nil
		return nil, true
	}
	m.details = next
	m.conditionalAdd.condition = option.Key
	return m.openDetailsValueEditor(), true
}

func (m *Model) cancelConditionalValueAdd() {
	if m.conditionalAdd == nil {
		return
	}
	if m.conditionalAdd.condition != "" {
		m.details = m.details.RemoveAddedConditionalValue(m.conditionalAdd.condition)
	}
	m.conditionalAdd = nil
}

func (m *Model) finishConditionalValueAdd() {
	m.conditionalAdd = nil
}
