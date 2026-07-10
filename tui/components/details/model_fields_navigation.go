package details

import (
	"strings"

	"charm.land/lipgloss/v2"
)

func (m *Model) focusNextItem(delta int) {
	if m.data == nil {
		return
	}
	m.nameInput.Blur()
	m.descInput.Blur()
	m.dropdownOpen = false
	m.dropdownIndex = 0
	m.groupInput.Blur()
	fields := []fieldID{fieldGroup, fieldName, fieldType, fieldDescription}
	total := len(fields) + len(m.data.Parameter.Values)
	if total == 0 {
		m.activeField = fieldNone
		m.selectedValue = -1
		return
	}
	idx := -1
	for i, field := range fields {
		if field == m.activeField {
			idx = i
			break
		}
	}
	if idx < 0 && m.selectedValue >= 0 {
		idx = len(fields) + m.selectedValue
	}
	if idx < 0 && delta < 0 && len(m.data.Parameter.Values) > 0 {
		idx = total
	}
	idx = (idx + delta + total) % total
	if idx < len(fields) {
		m.activeField = fields[idx]
		m.selectedValue = -1
	} else {
		m.activeField = fieldNone
		m.selectedValue = idx - len(fields)
	}
	if m.activeField == fieldName {
		_ = m.nameInput.Focus()
	}
	if m.activeField == fieldDescription {
		_ = m.descInput.Focus()
	}
	m.ensureSelectionVisible()
}

func (m *Model) activateField(field fieldID) {
	m.activeField = field
	m.selectedValue = -1
	m.dropdownOpen = false
	m.dropdownIndex = 0
	m.nameInput.Blur()
	m.descInput.Blur()
	m.groupInput.Blur()
	if field == fieldName {
		_ = m.nameInput.Focus()
	}
	if field == fieldDescription {
		_ = m.descInput.Focus()
	}
}

func (m Model) fieldAt(x, y int) (fieldID, bool) {
	if !m.Contains(x, y) {
		return fieldNone, false
	}
	fields := []fieldID{fieldGroup, fieldName, fieldType, fieldDescription}
	for _, field := range fields {
		if y >= m.fieldScreenY(field) && y < m.fieldScreenY(field)+m.fieldVisualHeight(field) {
			return field, true
		}
	}
	return fieldNone, false
}

func (m Model) fieldVisualHeight(field fieldID) int {
	if field == fieldDescription {
		return m.descriptionVisualHeight()
	}
	return 1
}

func (m Model) fieldScreenY(field fieldID) int {
	return m.y + 1 + m.fieldValueLine(field) - m.viewport.YOffset()
}

func (m Model) valueAt(_, y int) (int, bool) {
	if m.data == nil {
		return 0, false
	}
	for i := range m.data.Parameter.Values {
		if y == m.y+1+m.valueConditionLine(i)-m.viewport.YOffset() {
			return i, true
		}
	}
	return 0, false
}

func (m *Model) positionCursorForClick(field fieldID, x, y int) {
	contentX := m.x + 2
	col := max(x-contentX, 0)
	switch field {
	case fieldName:
		m.nameInput.SetCursor(min(col, len([]rune(m.nameInput.Value()))))
	case fieldDescription:
		line := max(y-m.fieldScreenY(fieldDescription), 0)
		width := m.descriptionTextWidth()
		offset := wrappedOffsetForClick(descriptionWrapSegments(m.descInput.Value(), width), line, col)
		m.descInput.SetCursorColumn(min(offset, len([]rune(m.descInput.Value()))))
	}
}

func (m *Model) ensureSelectionVisible() {
	line := -1
	if m.activeField != fieldNone {
		if m.activeField == fieldGroup {
			m.viewport.GotoTop()
			return
		}
		line = m.fieldValueLine(m.activeField)
	} else if m.selectedValue >= 0 {
		m.ensureValueVisible(m.selectedValue)
		return
	}
	if line < 0 {
		return
	}
	top := m.viewport.YOffset()
	bottom := top + m.viewport.Height() - 1
	switch {
	case line < top:
		m.viewport.SetYOffset(line)
	case line > bottom:
		m.viewport.SetYOffset(max(line-m.viewport.Height()+1, 0))
	}
}

// ensureSelectedBlockVisible keeps selected details content in view after rerender.
func (m *Model) ensureSelectedBlockVisible() {
	if m.activeField == fieldGroup {
		m.viewport.GotoTop()
		return
	}
	if m.activeField != fieldNone {
		m.ensureSelectionVisible()
		return
	}
	if m.selectedValue >= 0 {
		m.ensureValueVisible(m.selectedValue)
	}
}

// ensureValueVisible adjusts scroll so selected condition and value are visible when possible.
func (m *Model) ensureValueVisible(index int) {
	if m.data == nil || index < 0 || index >= len(m.data.Parameter.Values) {
		return
	}
	start := m.valueConditionLine(index)
	end := m.valueEndLine(index)
	height := max(m.viewport.Height(), 1)
	if end-start+1 > height {
		m.viewport.SetYOffset(start)
		return
	}
	top := m.viewport.YOffset()
	bottom := top + height - 1
	switch {
	case start < top:
		m.viewport.SetYOffset(start)
	case end > bottom:
		m.viewport.SetYOffset(max(end-height+1, 0))
	}
}

func (m Model) dropdownCurrentContains(x, y int) bool {
	currentX, currentY := m.DropdownCurrentPosition()
	width := lipgloss.Width(m.dropdownCurrentLabel()) + 4
	return x >= currentX && x < currentX+width && y >= currentY && y < currentY+3
}

func (m Model) dropdownRowAt(x, y int) (int, bool) {
	if !m.DropdownOpen() {
		return 0, false
	}
	rows := m.dropdownRows()
	if len(rows) == 0 {
		return 0, false
	}
	listX, listY := m.DropdownListPosition()
	width := 1
	for _, row := range rows {
		width = max(width, lipgloss.Width(row.Label))
	}
	if m.activeField == fieldGroup {
		width = max(width, lipgloss.Width(strings.TrimSpace(m.groupInput.Value()))+1)
		width = max(width, lipgloss.Width(m.groupInput.Placeholder))
	}
	if x < listX || x >= listX+width+4 {
		return 0, false
	}
	idx := y - listY - 1
	if idx < 0 || idx >= len(rows) {
		return 0, false
	}
	return idx, true
}
