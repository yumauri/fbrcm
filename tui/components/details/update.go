package details

import (
	tea "charm.land/bubbletea/v2"

	tuiconfig "github.com/yumauri/fbrcm/tui/config"
	"github.com/yumauri/fbrcm/tui/messages"
)

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.active {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		k := msg.String()
		if m.data != nil || m.groupData != nil || m.conditionData != nil {
			switch {
			case tuiconfig.Matches(tuiconfig.BlockDetailsForm, tuiconfig.ActionDown, k):
				if !m.dropdownOpen {
					m.focusNextItem(1)
					m.refreshViewport()
					return m, nil
				}
			case tuiconfig.Matches(tuiconfig.BlockDetailsForm, tuiconfig.ActionUp, k):
				if !m.dropdownOpen {
					m.focusNextItem(-1)
					m.refreshViewport()
					return m, nil
				}
			case tuiconfig.Matches(tuiconfig.BlockDetailsForm, tuiconfig.ActionClose, k):
				if m.activeField != fieldNone {
					if m.dropdownOpen {
						m.closeDropdown()
						m.refreshViewport()
						return m, nil
					}
					m = m.DeactivateField()
					return m, nil
				}
				if m.ValueSelected() {
					m.selectedValue = -1
					m.refreshViewport()
					return m, nil
				}
				if m.UsageSelected() || m.AddConditionalValueSelected() {
					m.selectedUsage = -1
					m.selectedAddValue = false
					m.refreshViewport()
					return m, nil
				}
			}
		}
		if m.activeField != fieldNone {
			var cmd tea.Cmd
			switch m.activeField {
			case fieldGroup:
				if m.dropdownOpen {
					switch {
					case tuiconfig.Matches(tuiconfig.BlockDetailsForm, tuiconfig.ActionUp, k):
						m.moveDropdown(-1)
					case tuiconfig.Matches(tuiconfig.BlockDetailsForm, tuiconfig.ActionDown, k):
						m.moveDropdown(1)
					case tuiconfig.Matches(tuiconfig.BlockDetailsForm, tuiconfig.ActionSubmit, k):
						m.commitDropdown()
					default:
						if m.dropdownInputSelected() {
							m.groupInput, cmd = m.groupInput.Update(msg)
						}
					}
				} else {
					switch {
					case tuiconfig.Matches(tuiconfig.BlockDetailsForm, tuiconfig.ActionRight, k):
						m.openDropdown()
					case tuiconfig.Matches(tuiconfig.BlockDetailsForm, tuiconfig.ActionSubmit, k):
						if m.dropdownOpen {
							m.commitDropdown()
						} else {
							m = m.DeactivateField()
						}
					}
				}
			case fieldType:
				if m.dropdownOpen {
					switch {
					case tuiconfig.Matches(tuiconfig.BlockDetailsForm, tuiconfig.ActionUp, k):
						m.moveDropdown(-1)
					case tuiconfig.Matches(tuiconfig.BlockDetailsForm, tuiconfig.ActionDown, k):
						m.moveDropdown(1)
					case tuiconfig.Matches(tuiconfig.BlockDetailsForm, tuiconfig.ActionSubmit, k):
						m.commitDropdown()
					}
				} else {
					switch {
					case tuiconfig.Matches(tuiconfig.BlockDetailsForm, tuiconfig.ActionRight, k):
						m.openDropdown()
					case tuiconfig.Matches(tuiconfig.BlockDetailsForm, tuiconfig.ActionSubmit, k):
						m = m.DeactivateField()
					}
				}
			case fieldName:
				if tuiconfig.Matches(tuiconfig.BlockDetailsForm, tuiconfig.ActionSubmit, k) {
					m = m.DeactivateField()
				} else {
					m.nameInput, cmd = m.nameInput.Update(msg)
				}
			case fieldConditionPriority:
				if tuiconfig.Matches(tuiconfig.BlockDetailsForm, tuiconfig.ActionSubmit, k) {
					m = m.DeactivateField()
				} else {
					m, cmd = m.updatePriorityInput(msg)
				}
			case fieldConditionColor:
				if m.dropdownOpen {
					switch {
					case tuiconfig.Matches(tuiconfig.BlockDetailsForm, tuiconfig.ActionUp, k):
						m.moveDropdown(-1)
					case tuiconfig.Matches(tuiconfig.BlockDetailsForm, tuiconfig.ActionDown, k):
						m.moveDropdown(1)
					case tuiconfig.Matches(tuiconfig.BlockDetailsForm, tuiconfig.ActionSubmit, k):
						m.commitDropdown()
					}
				} else {
					switch {
					case tuiconfig.Matches(tuiconfig.BlockDetailsForm, tuiconfig.ActionRight, k):
						m.openDropdown()
					case tuiconfig.Matches(tuiconfig.BlockDetailsForm, tuiconfig.ActionSubmit, k):
						m = m.DeactivateField()
					}
				}
			case fieldDescription:
				if tuiconfig.Matches(tuiconfig.BlockDetailsForm, tuiconfig.ActionSubmit, k) {
					m = m.DeactivateField()
				} else {
					m.descInput, cmd = m.descInput.Update(msg)
					m.normalizeDescriptionInput()
				}
			}
			m.refreshViewport()
			return m, cmd
		}
		if m.ValueSelected() {
			switch {
			case tuiconfig.Matches(tuiconfig.BlockDetails, tuiconfig.ActionEditValue, k):
				return m, nil
			}
		}
		switch {
		case tuiconfig.Matches(tuiconfig.BlockDetailsForm, tuiconfig.ActionUp, k):
			m.viewport.ScrollUp(1)
		case tuiconfig.Matches(tuiconfig.BlockDetailsForm, tuiconfig.ActionDown, k):
			m.viewport.ScrollDown(1)
		case tuiconfig.Matches(tuiconfig.BlockDetailsForm, tuiconfig.ActionPageUp, k):
			m.viewport.PageUp()
		case tuiconfig.Matches(tuiconfig.BlockDetailsForm, tuiconfig.ActionPageDown, k):
			m.viewport.PageDown()
		case tuiconfig.Matches(tuiconfig.BlockDetailsForm, tuiconfig.ActionHome, k):
			m.viewport.GotoTop()
		case tuiconfig.Matches(tuiconfig.BlockDetailsForm, tuiconfig.ActionEnd, k):
			m.viewport.GotoBottom()
		}
	case tea.MouseWheelMsg:
		if !m.Contains(msg.Mouse().X, msg.Mouse().Y) {
			break
		}
		switch msg.Mouse().Button {
		case tea.MouseWheelUp:
			m.viewport.ScrollUp(1)
		case tea.MouseWheelDown:
			m.viewport.ScrollDown(1)
		}
	case tea.MouseClickMsg:
		if m.data == nil && m.groupData == nil && m.conditionData == nil {
			break
		}
		return m.handleMouseClick(msg)
	case tea.PasteMsg, tea.ClipboardMsg:
		var cmd tea.Cmd
		switch m.activeField {
		case fieldName:
			m.nameInput, cmd = m.nameInput.Update(msg)
		case fieldConditionPriority:
			m, cmd = m.updatePriorityInput(msg)
		case fieldDescription:
			m.descInput, cmd = m.descInput.Update(msg)
			m.normalizeDescriptionInput()
		case fieldGroup:
			if m.dropdownOpen && m.dropdownInputSelected() {
				m.groupInput, cmd = m.groupInput.Update(msg)
			}
		}
		m.refreshViewport()
		return m, cmd
	}

	return m, nil
}

func (m Model) handleMouseClick(msg tea.MouseClickMsg) (Model, tea.Cmd) {
	mouse := msg.Mouse()
	if m.dropdownOpen {
		if idx, ok := m.dropdownRowAt(mouse.X, mouse.Y); ok {
			m.dropdownIndex = idx
			rows := m.dropdownRows()
			if idx >= 0 && idx < len(rows) && rows[idx].Input {
				_ = m.groupInput.Focus()
				m.nameInput.Blur()
				m.descInput.Blur()
			} else {
				m.groupInput.Blur()
				m.commitDropdown()
			}
			m.refreshViewport()
			return m, nil
		}
		if m.dropdownCurrentContains(mouse.X, mouse.Y) {
			m.refreshViewport()
			return m, nil
		}
	}

	if idx, ok := m.valueAt(mouse.X, mouse.Y); ok {
		m.activeField = fieldNone
		m.selectedValue = idx
		m.selectedUsage = -1
		m.selectedAddValue = false
		m.nameInput.Blur()
		m.descInput.Blur()
		m.groupInput.Blur()
		m.dropdownOpen = false
		m.refreshViewport()
		return m, func() tea.Msg { return messages.DetailsValueEditRequestedMsg{} }
	}
	if idx, ok := m.usageAt(mouse.X, mouse.Y); ok {
		m.activeField = fieldNone
		m.selectedValue = -1
		m.selectedUsage = idx
		m.selectedAddValue = false
		m.nameInput.Blur()
		m.priorityInput.Blur()
		m.dropdownOpen = false
		m.refreshViewport()
		return m, nil
	}
	if m.addConditionalValueAt(mouse.X, mouse.Y) {
		return m, func() tea.Msg { return messages.DetailsAddConditionalValueRequestedMsg{} }
	}

	field, ok := m.fieldAt(mouse.X, mouse.Y)
	if !ok {
		return m, nil
	}
	m.activateField(field)
	m.positionCursorForClick(field, mouse.X, mouse.Y)
	if field == fieldGroup || field == fieldType || field == fieldConditionColor {
		m.openDropdown()
	}
	m.refreshViewport()
	return m, nil
}
