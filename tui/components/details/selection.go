package details

import (
	"charm.land/lipgloss/v2"
	"github.com/yumauri/fbrcm/tui/components/parameters"
)

// SelectedRawValue returns selected value raw content.
func (m Model) SelectedRawValue() (string, bool) {
	if !m.ValueSelected() {
		return "", false
	}
	return m.data.Parameter.Values[m.selectedValue].RawValue, true
}

// CurrentConditionalValueAnchor returns selected conditional value deletion target.
func (m Model) CurrentConditionalValueAnchor() (parameters.ConditionalValueAnchor, bool) {
	if !m.ValueSelected() {
		return parameters.ConditionalValueAnchor{}, false
	}
	value := m.data.Parameter.Values[m.selectedValue]
	if value.Label == "" || value.Label == "default" {
		return parameters.ConditionalValueAnchor{}, false
	}
	return parameters.ConditionalValueAnchor{Project: m.data.Project, GroupKey: m.data.GroupKey, ParamKey: m.data.Parameter.Key, ValueLabel: value.Label}, true
}

func (m Model) DropdownOpen() bool {
	return m.dropdownOpen && (m.activeField == fieldGroup || m.activeField == fieldType || m.activeField == fieldConditionColor)
}
func (m Model) DropdownCurrentPosition() (int, int) {
	fieldLine := m.fieldValueLine(m.activeField)
	return m.x + 1, m.y + fieldLine - m.viewport.YOffset()
}
func (m Model) DropdownListPosition() (int, int) {
	x, y := m.DropdownCurrentPosition()
	return x + lipgloss.Width(m.dropdownCurrentLabel()) + 3, y - m.dropdownIndex
}
func (m Model) Bounds() (int, int, int, int) { return m.x, m.y, m.width, m.height }
func (m Model) Contains(x, y int) bool {
	return m.width > 0 && m.height > 0 && x >= m.x && x < m.x+m.width && y >= m.y && y < m.y+m.height
}

func leadingSpaceWidth(value string) int {
	width := 0
	for _, r := range value {
		if r != ' ' {
			break
		}
		width++
	}
	return width
}
