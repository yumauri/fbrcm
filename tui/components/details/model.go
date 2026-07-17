package details

import (
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/tui/messages"
)

type Model struct {
	x                   int
	y                   int
	width               int
	height              int
	active              bool
	bridgeActive        bool
	viewport            viewport.Model
	data                *messages.ParameterViewData
	conditionData       *messages.ConditionViewData
	activeField         fieldID
	dropdownOpen        bool
	dropdownIndex       int
	groupKey            string
	groupLabel          string
	typeValue           string
	nameInput           textinput.Model
	descInput           textarea.Model
	groupInput          textinput.Model
	priorityInput       textinput.Model
	conditionColor      string
	conditionExpression string
	selectedValue       int
	selectedUsage       int
	selectedAddValue    bool
	valuesInvalid       bool
	originalParam       core.ParametersEntry
	originalCondition   core.ConditionEntry
}

func New() Model {
	vp := viewport.New(
		viewport.WithWidth(1),
		viewport.WithHeight(1),
	)
	vp.SoftWrap = true
	return Model{
		viewport:      vp,
		nameInput:     newTextInput(),
		descInput:     newDescriptionInput(),
		groupInput:    newGroupInput(),
		priorityInput: newTextInput(),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) SetBounds(x, y, width, height int) Model {
	if m.x == x && m.y == y && m.width == width && m.height == height {
		return m
	}
	m.x = x
	m.y = y
	m.width = width
	m.height = height
	m.refreshViewport()
	return m
}

func (m Model) SetActive(active bool) Model {
	m.active = active
	return m
}

// ResetScroll resets details viewport scroll position.
func (m Model) ResetScroll() Model {
	m.viewport.GotoTop()
	return m
}

func (m Model) SetBridgeActive(active bool) Model {
	m.bridgeActive = active
	return m
}

func (m Model) Data() *messages.ParameterViewData {
	return m.data
}

func (m Model) ConditionData() *messages.ConditionViewData {
	return m.conditionData
}

func (m Model) IsCondition() bool { return m.conditionData != nil }

func (m Model) FieldActive() bool {
	return m.activeField != fieldNone
}

// TextInputActive reports whether active field should receive printable key strokes.
func (m Model) TextInputActive() bool {
	return m.activeField == fieldName || m.activeField == fieldDescription || m.activeField == fieldConditionPriority
}

func (m Model) ValueSelected() bool {
	return m.activeField == fieldNone && m.selectedValue >= 0 && m.data != nil && m.selectedValue < len(m.data.Parameter.Values)
}

func (m Model) UsageSelected() bool {
	return m.activeField == fieldNone && m.selectedUsage >= 0 && m.conditionData != nil && m.selectedUsage < len(m.conditionData.Condition.Usages)
}

func (m Model) AddConditionalValueSelected() bool {
	return m.activeField == fieldNone && m.selectedAddValue && m.data != nil && len(m.AvailableConditions()) > 0
}
