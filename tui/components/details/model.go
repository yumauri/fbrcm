package details

import (
	"bytes"
	"encoding/json"
	"math"
	"strings"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/charmbracelet/x/ansi"

	"fbrcm/core"
	corestyles "fbrcm/core/styles"
	jsoninput "fbrcm/tui/components/jsoninput"
	"fbrcm/tui/components/parameters"
	"fbrcm/tui/messages"
	"fbrcm/tui/styles"
)

const panelTitle = "[3] Details"

var (
	labelStyle             = styles.PanelMuted
	projectValueStyle      = styles.PanelText.Bold(true).Foreground(styles.PaletteError)
	groupValueStyle        = styles.PanelText.Bold(true).Foreground(styles.PaletteYellow)
	parameterKeyStyle      = styles.PanelBody.Foreground(styles.PaletteBlueBright)
	selectedValueStyle     = lipgloss.NewStyle().Background(styles.PaletteBlueDeep).Foreground(styles.PaletteSlateBright)
	conditionDefaultStyle  = styles.PanelMuted.Italic(true)
	fieldDirtyStyle        = styles.PanelMuted.Bold(true).Underline(true)
	fieldInvalidStyle      = lipgloss.NewStyle().Foreground(styles.PaletteError)
	fieldInvalidDirtyStyle = lipgloss.NewStyle().
				Foreground(styles.PaletteError).
				Bold(true).
				Underline(true)
)

// selectedDropdownFieldStyle selects selected dropdown field style and returns the resulting value or error.
func selectedDropdownFieldStyle() lipgloss.Style {
	if styles.NoColorEnabled() {
		return lipgloss.NewStyle().Reverse(true)
	}
	return selectedValueStyle
}

type fieldID int

const (
	fieldNone fieldID = iota
	fieldGroup
	fieldName
	fieldType
	fieldDescription
)

var typeOptions = []string{"STRING", "BOOLEAN", "NUMBER", "JSON"}

// Model holds model state used by the details package.
type Model struct {
	// x stores x for Model.
	x int
	// y stores y for Model.
	y int
	// width stores width for Model.
	width int
	// height stores height for Model.
	height int
	// active stores active for Model.
	active bool
	// bridgeActive stores bridge active for Model.
	bridgeActive bool
	// viewport stores viewport for Model.
	viewport viewport.Model
	// data stores data for Model.
	data *messages.ParameterViewData
	// activeField stores active field for Model.
	activeField fieldID
	// dropdownOpen stores dropdown open for Model.
	dropdownOpen bool
	// dropdownIndex stores dropdown index for Model.
	dropdownIndex int
	// groupKey stores group key for Model.
	groupKey string
	// groupLabel stores group label for Model.
	groupLabel string
	// typeValue stores type value for Model.
	typeValue string
	// nameInput stores name input for Model.
	nameInput textinput.Model
	// descInput stores desc input for Model.
	descInput textarea.Model
	// groupInput stores group input for Model.
	groupInput textinput.Model
	// selectedValue stores selected value for Model.
	selectedValue int
	// valuesInvalid stores values invalid for Model.
	valuesInvalid bool
	// originalParam stores original param for Model.
	originalParam core.ParametersEntry
}

// New constructs new and returns the resulting value or error.
func New() Model {
	vp := viewport.New(
		viewport.WithWidth(1),
		viewport.WithHeight(1),
	)
	vp.SoftWrap = true
	return Model{
		viewport:   vp,
		nameInput:  newTextInput(),
		descInput:  newDescriptionInput(),
		groupInput: newGroupInput(),
	}
}

// Init initializes init for Model and returns the resulting state or error.
func (m Model) Init() tea.Cmd {
	return nil
}

// SetBounds sets bounds for Model and returns the resulting state or error.
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

// SetActive sets active for Model and returns the resulting state or error.
func (m Model) SetActive(active bool) Model {
	m.active = active
	return m
}

// SetBridgeActive sets bridge active for Model and returns the resulting state or error.
func (m Model) SetBridgeActive(active bool) Model {
	m.bridgeActive = active
	return m
}

// SetData sets data for Model and returns the resulting state or error.
func (m Model) SetData(data *messages.ParameterViewData) Model {
	m.data = cloneViewData(data)
	m.activeField = fieldNone
	m.dropdownOpen = false
	m.dropdownIndex = 0
	m.nameInput = newTextInput()
	m.descInput = newDescriptionInput()
	m.groupInput = newGroupInput()
	m.groupKey = ""
	m.groupLabel = "(root)"
	m.typeValue = "STRING"
	m.selectedValue = -1
	m.valuesInvalid = false
	if m.data != nil {
		m.originalParam = cloneParameterEntry(m.data.Parameter)
		m.selectedValue = m.data.SelectedValueIdx
		m.nameInput.SetValue(m.data.Parameter.Key)
		m.descInput.SetValue(m.data.Parameter.Description)
		m.groupKey = m.data.GroupKey
		m.groupLabel = m.data.GroupLabel
		for _, group := range m.data.Groups {
			if group.Key == m.data.GroupKey {
				m.groupLabel = group.Label
				break
			}
		}
		currentType := strings.ToUpper(parameterType(m.data.Parameter))
		for _, option := range typeOptions {
			if option == currentType {
				m.typeValue = option
				break
			}
		}
	}
	if m.data == nil {
		m.originalParam = core.ParametersEntry{}
	}
	m.refreshViewport()
	return m
}

// Data handles data for Model and returns the resulting state or error.
func (m Model) Data() *messages.ParameterViewData {
	return m.data
}

// FieldActive handles field active for Model and returns the resulting state or error.
func (m Model) FieldActive() bool {
	return m.activeField != fieldNone
}

// ValueSelected handles value selected for Model and returns the resulting state or error.
func (m Model) ValueSelected() bool {
	return m.activeField == fieldNone && m.selectedValue >= 0 && m.data != nil && m.selectedValue < len(m.data.Parameter.Values)
}

// SetValuesInvalid sets values invalid for Model and returns the resulting state or error.
func (m Model) SetValuesInvalid(invalid bool) Model {
	if m.valuesInvalid == invalid {
		return m
	}
	m.valuesInvalid = invalid
	m.refreshViewport()
	return m
}

// SetSelectedValue sets selected value for Model and returns the resulting state or error.
func (m Model) SetSelectedValue(nextRaw string) Model {
	if !m.ValueSelected() {
		return m
	}
	value := &m.data.Parameter.Values[m.selectedValue]
	value.RawValue = nextRaw
	value.Value = displayRawValue(nextRaw, m.selectedType())
	value.ValueType = m.selectedType()
	value.Empty = nextRaw == ""
	m.refreshViewport()
	return m
}

// Dirty handles dirty for Model and returns the resulting state or error.
func (m Model) Dirty() bool {
	return m.data != nil && m.hasChanges()
}

// Invalid handles invalid for Model and returns the resulting state or error.
func (m Model) Invalid() bool {
	return m.invalidName() || m.invalidValues()
}

// InvalidReasons handles invalid reasons for Model and returns the resulting state or error.
func (m Model) InvalidReasons() []string {
	if m.data == nil {
		return nil
	}
	reasons := make([]string, 0, 2)
	if m.invalidName() {
		nextKey := strings.TrimSpace(m.nameInput.Value())
		if nextKey == "" {
			reasons = append(reasons, "Parameter name is empty.")
		} else {
			reasons = append(reasons, "Parameter name already exists in this project.")
		}
	}
	if m.invalidValues() {
		reasons = append(reasons, "One or more values are invalid for selected type "+m.selectedType()+".")
	}
	return reasons
}

// Edit handles edit for Model and returns the resulting state or error.
func (m Model) Edit() (core.ParameterDetailsEdit, bool) {
	if m.data == nil || !m.hasChanges() {
		return core.ParameterDetailsEdit{}, false
	}
	return core.ParameterDetailsEdit{
		Create:          m.data.Parameter.Key == "",
		GroupKey:        m.data.GroupKey,
		ParamKey:        m.data.Parameter.Key,
		NextGroupKey:    m.selectedGroupKey(),
		NextParamKey:    strings.TrimSpace(m.nameInput.Value()),
		NextValueType:   m.selectedType(),
		NextDescription: m.descInput.Value(),
		ValueEdits:      m.valueEdits(),
	}, true
}

// MarkSaved handles mark saved for Model and returns the resulting state or error.
func (m Model) MarkSaved() Model {
	if m.data == nil {
		return m
	}
	edit, ok := m.Edit()
	if !ok {
		return m
	}
	m.data.GroupKey = edit.NextGroupKey
	m.data.GroupLabel = m.groupLabel
	m.data.Parameter.Key = edit.NextParamKey
	m.data.Parameter.Description = edit.NextDescription
	for i := range m.data.Parameter.Values {
		m.data.Parameter.Values[i].ValueType = edit.NextValueType
	}
	m.originalParam = cloneParameterEntry(m.data.Parameter)
	m.activeField = fieldNone
	m.refreshViewport()
	return m
}

// DeactivateField handles deactivate field for Model and returns the resulting state or error.
func (m Model) DeactivateField() Model {
	m.activeField = fieldNone
	m.dropdownOpen = false
	m.dropdownIndex = 0
	m.nameInput.Blur()
	m.descInput.Blur()
	m.groupInput.Blur()
	m.refreshViewport()
	return m
}

// ActivateName handles activate name for Model and returns the resulting state or error.
func (m Model) ActivateName() (Model, tea.Cmd) {
	m.activateField(fieldName)
	m.refreshViewport()
	return m, m.nameInput.Focus()
}

// DropdownOpen handles dropdown open for Model and returns the resulting state or error.
func (m Model) DropdownOpen() bool {
	return m.dropdownOpen && (m.activeField == fieldGroup || m.activeField == fieldType)
}

// DropdownPosition handles dropdown position for Model and returns the resulting state or error.
func (m Model) DropdownPosition() (int, int) {
	return m.DropdownListPosition()
}

// DropdownCurrentPosition handles dropdown current position for Model and returns the resulting state or error.
func (m Model) DropdownCurrentPosition() (int, int) {
	fieldLine := m.fieldValueLine(m.activeField)
	return m.x + 1, m.y + fieldLine - m.viewport.YOffset()
}

// DropdownListPosition handles dropdown list position for Model and returns the resulting state or error.
func (m Model) DropdownListPosition() (int, int) {
	x, y := m.DropdownCurrentPosition()
	return x + lipgloss.Width(m.dropdownCurrentLabel()) + 3, y - m.dropdownIndex
}

// DropdownCurrentView handles dropdown current view for Model and returns the resulting state or error.
func (m Model) DropdownCurrentView() string {
	if !m.DropdownOpen() {
		return ""
	}
	value := m.dropdownCurrentLabel()
	width := max(lipgloss.Width(value), 1)
	return strings.Join([]string{
		dropdownBorderStyle.Render("╭" + strings.Repeat("─", width+2) + "╮"),
		dropdownBorderStyle.Render("│ ") + m.dropdownCurrentStyle().Render(padRight(value, width)) + dropdownBorderStyle.Render(" │"),
		dropdownBorderStyle.Render("╰" + strings.Repeat("─", width+2) + "╯"),
	}, "\n")
}

// DropdownListView handles dropdown list view for Model and returns the resulting state or error.
func (m Model) DropdownListView() string {
	if !m.DropdownOpen() {
		return ""
	}
	rows := m.dropdownRows()
	if len(rows) == 0 {
		return ""
	}
	width := 1
	for _, row := range rows {
		width = max(width, lipgloss.Width(row.Label))
	}
	if m.activeField == fieldGroup {
		width = max(width, lipgloss.Width(strings.TrimSpace(m.groupInput.Value()))+1)
		width = max(width, lipgloss.Width(m.groupInput.Placeholder))
	}
	lines := make([]string, 0, len(rows)+2)
	topLeft := "╭"
	if m.dropdownIndex == 0 {
		topLeft = "─"
	}
	bottomLeft := "╰"
	if m.dropdownIndex == len(rows)-1 {
		bottomLeft = "─"
	}
	lines = append(lines, dropdownBorderStyle.Render(topLeft+strings.Repeat("─", width+2)+"╮"))
	input := m.groupInput
	for i, row := range rows {
		left := dropdownBorderStyle.Render("│ ")
		switch i {
		case m.dropdownIndex:
			left = dropdownBorderStyle.Render("▸ ")
		case m.dropdownIndex - 1:
			left = dropdownBorderStyle.Render("╯ ")
		case m.dropdownIndex + 1:
			left = dropdownBorderStyle.Render("╮ ")
		}
		content := ""
		if row.Input {
			if i == m.dropdownIndex {
				input.SetWidth(max(width-1, 1))
				content = padRight(input.View(), width)
			} else if value := strings.TrimSpace(m.groupInput.Value()); value != "" {
				content = dropdownOptionStyle(false).Render(padRight(value, width))
			} else {
				content = styles.PanelMuted.Render(padRight(input.Placeholder, width))
			}
		} else {
			content = dropdownOptionStyle(i == m.dropdownIndex).Render(padRight(row.Label, width))
		}
		lines = append(lines, left+content+dropdownBorderStyle.Render(" │"))
	}
	lines = append(lines, dropdownBorderStyle.Render(bottomLeft+strings.Repeat("─", width+2)+"╯"))
	return strings.Join(lines, "\n")
}

// Bounds handles bounds for Model and returns the resulting state or error.
func (m Model) Bounds() (int, int, int, int) {
	return m.x, m.y, m.width, m.height
}

// Contains handles contains for Model and returns the resulting state or error.
func (m Model) Contains(x, y int) bool {
	if m.width <= 0 || m.height <= 0 {
		return false
	}
	return x >= m.x && x < m.x+m.width && y >= m.y && y < m.y+m.height
}

// Update updates update for Model and returns the resulting state or error.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.active {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.data != nil {
			switch msg.String() {
			case "down":
				if !m.dropdownOpen {
					m.focusNextItem(1)
					m.refreshViewport()
					return m, nil
				}
			case "up":
				if !m.dropdownOpen {
					m.focusNextItem(-1)
					m.refreshViewport()
					return m, nil
				}
			case "esc":
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
			}
		}
		if m.activeField != fieldNone {
			var cmd tea.Cmd
			switch m.activeField {
			case fieldGroup:
				if m.dropdownOpen {
					switch msg.String() {
					case "up":
						m.moveDropdown(-1)
					case "down":
						m.moveDropdown(1)
					case "enter":
						m.commitDropdown()
					default:
						if m.dropdownInputSelected() {
							m.groupInput, cmd = m.groupInput.Update(msg)
						}
					}
				} else {
					switch msg.String() {
					case "right":
						m.openDropdown(1)
					case "enter":
						if m.dropdownOpen {
							m.commitDropdown()
						} else {
							m = m.DeactivateField()
						}
					}
				}
			case fieldType:
				if m.dropdownOpen {
					switch msg.String() {
					case "up":
						m.moveDropdown(-1)
					case "down":
						m.moveDropdown(1)
					case "enter":
						m.commitDropdown()
					}
				} else {
					switch msg.String() {
					case "right":
						m.openDropdown(1)
					case "enter":
						m = m.DeactivateField()
					}
				}
			case fieldName:
				m.nameInput, cmd = m.nameInput.Update(msg)
			case fieldDescription:
				if msg.String() != "enter" {
					m.descInput, cmd = m.descInput.Update(msg)
					m.normalizeDescriptionInput()
				}
			}
			m.refreshViewport()
			return m, cmd
		}
		if m.ValueSelected() {
			switch msg.String() {
			case "right", "f4":
				return m, nil
			}
		}
		switch msg.String() {
		case "up", "k":
			m.viewport.ScrollUp(1)
		case "down", "j":
			m.viewport.ScrollDown(1)
		case "pgup", "h":
			m.viewport.PageUp()
		case "pgdown", "l":
			m.viewport.PageDown()
		case "home":
			m.viewport.GotoTop()
		case "end":
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
		if m.data == nil {
			break
		}
		return m.handleMouseClick(msg)
	case tea.PasteMsg, tea.ClipboardMsg:
		var cmd tea.Cmd
		switch m.activeField {
		case fieldName:
			m.nameInput, cmd = m.nameInput.Update(msg)
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

// handleMouseClick handles handle mouse click for Model and returns the resulting state or error.
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
		m.nameInput.Blur()
		m.descInput.Blur()
		m.groupInput.Blur()
		m.dropdownOpen = false
		m.refreshViewport()
		return m, func() tea.Msg { return messages.DetailsValueEditRequestedMsg{} }
	}

	field, ok := m.fieldAt(mouse.X, mouse.Y)
	if !ok {
		return m, nil
	}
	m.activateField(field)
	m.positionCursorForClick(field, mouse.X, mouse.Y)
	if field == fieldGroup || field == fieldType {
		m.openDropdown(1)
	}
	m.refreshViewport()
	return m, nil
}

// refreshViewport handles refresh viewport for Model and returns the resulting state or error.
func (m *Model) refreshViewport() {
	width := max(m.width-5, 1)
	m.nameInput.SetWidth(max(width-2, 1))
	m.resizeDescriptionInput(width)
	m.viewport.SetWidth(width)
	m.viewport.SetHeight(max(m.height-2, 1))
	m.viewport.SetContentLines(m.renderContentLines())
}

// renderContentLines renders render content lines for Model and returns the resulting state or error.
func (m Model) renderContentLines() []string {
	width := max(m.width-5, 1)
	if m.data == nil {
		return padLines([]string{
			"Press Enter on parameter",
			"to open details panel.",
		}, width)
	}

	lines := make([]string, 0, 32)
	lines = appendStyledField(lines, width, "Project", displayProject(m.data.Project), projectValueStyle)
	lines = appendEditableField(lines, width, "Group", m.renderGroupField(), m.fieldChanged(fieldGroup), false)
	lines = appendEditableField(lines, width, "Name", m.renderNameField(), m.fieldChanged(fieldName), m.invalidName())
	lines = appendEditableField(lines, width, "Type", m.renderTypeField(), m.fieldChanged(fieldType), false)
	lines = appendEditableField(lines, width, "Description", m.renderDescriptionField(), m.fieldChanged(fieldDescription), false)

	valuesTitle := fieldTitle("Values", m.valueChanged(), m.invalidValues())
	lines = append(lines, valuesTitle)
	for i, value := range m.data.Parameter.Values {
		prefix := "  "
		if m.activeField == fieldNone && i == m.selectedValue {
			prefix = "▸ "
		}

		conditionStyle := m.conditionStyle(value.Color)
		if value.Label == "default" {
			conditionStyle = conditionDefaultStyle
		}

		label := prefix + displayConditionLabel(value.Label)
		if m.activeField == fieldNone && i == m.selectedValue {
			label = selectedValueStyle.Render(ansi.Truncate(label, width, ""))
		} else {
			label = conditionStyle.Render(ansi.Truncate(label, width, ""))
		}
		lines = append(lines, label)

		valueLines := m.renderValueLines(value, max(width-4, 1))
		for _, line := range valueLines {
			lines = append(lines, ansi.Truncate("    "+line, width, ""))
		}
		lines = append(lines, "")
	}

	if len(m.data.Parameter.Values) == 0 {
		lines = append(lines, "No values.")
	}

	return padLines(lines, width)
}

// appendStyledField handles append styled field and returns the resulting value or error.
func appendStyledField(lines []string, width int, label, value string, style lipgloss.Style) []string {
	lines = append(lines, labelStyle.Render(label))
	for _, line := range wrappedLines(value, width) {
		lines = append(lines, style.Render(ansi.Truncate(line, width, "")))
	}
	lines = append(lines, "")
	return lines
}

// appendEditableField handles append editable field and returns the resulting value or error.
func appendEditableField(lines []string, width int, label, value string, dirty, invalid bool) []string {
	labelText := fieldTitle(label, dirty, invalid)
	lines = append(lines, labelText)
	for line := range strings.SplitSeq(value, "\n") {
		lines = append(lines, ansi.Truncate(line, width, ""))
	}
	lines = append(lines, "")
	return lines
}

// fieldTitle handles field title and returns the resulting value or error.
func fieldTitle(label string, dirty, invalid bool) string {
	switch {
	case invalid && dirty:
		return fieldInvalidDirtyStyle.Render(label)
	case invalid:
		return fieldInvalidStyle.Render(label)
	case dirty:
		return fieldDirtyStyle.Render(label)
	default:
		return labelStyle.Render(label)
	}
}

// wrappedLines handles wrapped lines and returns the resulting value or error.
func wrappedLines(value string, width int) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return []string{"-"}
	}
	rendered := lipgloss.NewStyle().Width(width).Render(value)
	return strings.Split(rendered, "\n")
}

// wrapLine handles wrap line and returns the resulting value or error.
func wrapLine(value string, width int) []string {
	if width <= 0 {
		return []string{""}
	}
	if value == "" {
		return []string{""}
	}
	wrapped := ansi.Hardwrap(value, width, true)
	parts := strings.Split(wrapped, "\n")
	if len(parts) == 0 {
		return []string{""}
	}
	return parts
}

// padLines handles pad lines and returns the resulting value or error.
func padLines(lines []string, width int) []string {
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		out = append(out, line+strings.Repeat(" ", max(width-lipgloss.Width(line), 0)))
	}
	return out
}

// padRight handles pad right and returns the resulting value or error.
func padRight(value string, width int) string {
	return value + strings.Repeat(" ", max(width-lipgloss.Width(value), 0))
}

// displayProject handles display project and returns the resulting value or error.
func displayProject(project core.Project) string {
	if strings.TrimSpace(project.Name) == "" {
		return project.ProjectID
	}
	return project.Name + " (" + project.ProjectID + ")"
}

// displayConditionLabel handles display condition label and returns the resulting value or error.
func displayConditionLabel(label string) string {
	if label == "default" {
		return "Default value"
	}
	return label
}

// displayRawValue handles display raw value and returns the resulting value or error.
func displayRawValue(value, valueType string) string {
	if value == "" {
		valueType = strings.ToLower(strings.TrimSpace(valueType))
		if valueType == "" {
			valueType = "string"
		}
		return "(empty " + valueType + ")"
	}
	return strings.ReplaceAll(value, "\n", "\\n")
}

// cloneParameterEntry handles clone parameter entry and returns the resulting value or error.
func cloneParameterEntry(param core.ParametersEntry) core.ParametersEntry {
	param.Values = append([]core.ParametersValue(nil), param.Values...)
	return param
}

// cloneViewData handles clone view data and returns the resulting value or error.
func cloneViewData(data *messages.ParameterViewData) *messages.ParameterViewData {
	if data == nil {
		return nil
	}
	next := *data
	next.Groups = append([]messages.ParameterGroupOption(nil), data.Groups...)
	next.ParameterKeys = append([]string(nil), data.ParameterKeys...)
	next.Parameter = cloneParameterEntry(data.Parameter)
	return &next
}

// parameterType handles parameter type and returns the resulting value or error.
func parameterType(param core.ParametersEntry) string {
	for _, value := range param.Values {
		if strings.TrimSpace(value.ValueType) != "" {
			return value.ValueType
		}
	}
	return "unspecified"
}

// focusNextItem focuses focus next item for Model and returns the resulting state or error.
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

// activateField handles activate field for Model and returns the resulting state or error.
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

// fieldAt handles field at for Model and returns the resulting state or error.
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

// fieldVisualHeight handles field visual height for Model and returns the resulting state or error.
func (m Model) fieldVisualHeight(field fieldID) int {
	if field == fieldDescription {
		return m.descriptionVisualHeight()
	}
	return 1
}

// fieldScreenY handles field screen y for Model and returns the resulting state or error.
func (m Model) fieldScreenY(field fieldID) int {
	return m.y + 1 + m.fieldValueLine(field) - m.viewport.YOffset()
}

// valueAt handles value at for Model and returns the resulting state or error.
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

// positionCursorForClick handles position cursor for click for Model and returns the resulting state or error.
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

// ensureSelectionVisible handles ensure selection visible for Model and returns the resulting state or error.
func (m *Model) ensureSelectionVisible() {
	line := -1
	if m.activeField != fieldNone {
		line = m.fieldValueLine(m.activeField)
	} else if m.selectedValue >= 0 {
		line = m.valueConditionLine(m.selectedValue)
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

// dropdownCurrentContains handles dropdown current contains for Model and returns the resulting state or error.
func (m Model) dropdownCurrentContains(x, y int) bool {
	currentX, currentY := m.DropdownCurrentPosition()
	width := lipgloss.Width(m.dropdownCurrentLabel()) + 4
	return x >= currentX && x < currentX+width && y >= currentY && y < currentY+3
}

// dropdownRowAt handles dropdown row at for Model and returns the resulting state or error.
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

// CurrentBoolValueAnchor handles current bool value anchor for Model and returns the resulting state or error.
func (m Model) CurrentBoolValueAnchor() (parameters.BoolValueAnchor, bool) {
	value, ok := m.currentSelectedPlainValue("boolean")
	if !ok {
		return parameters.BoolValueAnchor{}, false
	}
	x, y := m.valueEditorPosition()
	return parameters.BoolValueAnchor{
		Project:      m.data.Project,
		GroupKey:     m.data.GroupKey,
		ParamKey:     m.data.Parameter.Key,
		ValueLabel:   value.Label,
		Value:        strings.EqualFold(strings.TrimSpace(value.RawValue), "true"),
		CurrentValue: value.RawValue,
		X:            x + 2,
		Y:            y,
	}, true
}

// CurrentNumberValueAnchor handles current number value anchor for Model and returns the resulting state or error.
func (m Model) CurrentNumberValueAnchor() (parameters.NumberValueAnchor, bool) {
	value, ok := m.currentSelectedPlainValue("number")
	if !ok {
		return parameters.NumberValueAnchor{}, false
	}
	currentValue := strings.TrimSpace(value.RawValue)
	x, y := m.valueEditorPosition()
	return parameters.NumberValueAnchor{
		Project:      m.data.Project,
		GroupKey:     m.data.GroupKey,
		ParamKey:     m.data.Parameter.Key,
		ValueLabel:   value.Label,
		CurrentValue: currentValue,
		X:            x + 2,
		Y:            y - 1,
		Width:        max(lipgloss.Width(currentValue), 3),
		MaxWidth:     max(m.width-5, 3),
	}, true
}

// CurrentStringValueAnchor handles current string value anchor for Model and returns the resulting state or error.
func (m Model) CurrentStringValueAnchor(_ int) (parameters.StringValueAnchor, bool) {
	value, ok := m.currentSelectedPlainValue("string")
	if !ok {
		return parameters.StringValueAnchor{}, false
	}
	currentValue := value.RawValue
	x, y := m.valueEditorPosition()
	editorX := x + 2
	width := max(m.width-(editorX-m.x)-2, 15)
	return parameters.StringValueAnchor{
		Project:      m.data.Project,
		GroupKey:     m.data.GroupKey,
		ParamKey:     m.data.Parameter.Key,
		ValueLabel:   value.Label,
		CurrentValue: currentValue,
		X:            editorX,
		Y:            y - 1,
		Width:        width,
		MaxWidth:     width + 2,
		FullWidth:    false,
		Expanded:     strings.Contains(currentValue, "\n"),
	}, true
}

// CurrentJSONValueAnchor handles current jsonvalue anchor for Model and returns the resulting state or error.
func (m Model) CurrentJSONValueAnchor() (parameters.JSONValueAnchor, bool) {
	value, ok := m.currentSelectedPlainValue("json")
	if !ok {
		return parameters.JSONValueAnchor{}, false
	}
	return parameters.JSONValueAnchor{
		Project:      m.data.Project,
		GroupKey:     m.data.GroupKey,
		ParamKey:     m.data.Parameter.Key,
		ValueLabel:   value.Label,
		CurrentValue: value.RawValue,
	}, true
}

// currentSelectedPlainValue handles current selected plain value for Model and returns the resulting state or error.
func (m Model) currentSelectedPlainValue(valueType string) (core.ParametersValue, bool) {
	if !m.ValueSelected() {
		return core.ParametersValue{}, false
	}
	value := m.data.Parameter.Values[m.selectedValue]
	if !value.Plain {
		return core.ParametersValue{}, false
	}
	selectedType := strings.TrimSpace(strings.ToLower(m.selectedType()))
	if selectedType == "" {
		selectedType = "string"
	}
	if selectedType != valueType {
		return core.ParametersValue{}, false
	}
	return value, true
}

// valueEditorPosition handles value editor position for Model and returns the resulting state or error.
func (m Model) valueEditorPosition() (int, int) {
	line := m.valueConditionLine(m.selectedValue) + 1
	return m.x + 3, m.y + 1 + line - m.viewport.YOffset()
}

// renderGroupField renders render group field for Model and returns the resulting state or error.
func (m Model) renderGroupField() string {
	value := m.groupLabel
	if m.activeField == fieldGroup {
		return selectedDropdownFieldStyle().Render(value)
	}
	return groupValueStyle.Render(value)
}

// renderNameField renders render name field for Model and returns the resulting state or error.
func (m Model) renderNameField() string {
	if m.activeField == fieldName {
		return styles.FilterText.Render(m.nameInput.View())
	}
	return parameterKeyStyle.Render(strings.TrimSpace(m.nameInput.Value()))
}

// renderTypeField renders render type field for Model and returns the resulting state or error.
func (m Model) renderTypeField() string {
	value := m.typeValue
	if m.activeField == fieldType {
		return selectedDropdownFieldStyle().Render(value)
	}
	return styles.PanelText.Render(value)
}

// renderDescriptionField renders render description field for Model and returns the resulting state or error.
func (m Model) renderDescriptionField() string {
	width := m.descriptionTextWidth()
	if m.activeField == fieldDescription {
		return m.renderActiveDescription(width)
	}
	rawValue := m.descInput.Value()
	value := rawValue
	if value == "" {
		value = "No description."
	}
	segments := descriptionWrapSegments(value, width)
	lines := make([]string, 0, len(segments))
	for _, segment := range segments {
		if rawValue == "" {
			lines = append(lines, styles.PanelMuted.Italic(true).Render(segment.text))
		} else {
			lines = append(lines, styles.PanelText.Render(segment.text))
		}
	}
	return strings.Join(lines, "\n")
}

// resizeDescriptionInput handles resize description input for Model and returns the resulting state or error.
func (m *Model) resizeDescriptionInput(width int) {
	inputWidth := m.descriptionTextWidth()
	m.descInput.SetWidth(inputWidth)
	m.descInput.SetHeight(m.descriptionVisualHeightForWidth(inputWidth))
}

// normalizeDescriptionInput handles normalize description input for Model and returns the resulting state or error.
func (m *Model) normalizeDescriptionInput() {
	value := singleLineValue(m.descInput.Value())
	pos := m.descInput.Column()
	maxPos := len([]rune(value))
	if value != m.descInput.Value() {
		m.descInput.SetValue(value)
	}
	if pos > maxPos {
		m.descInput.SetCursorColumn(maxPos)
	}
}

// descriptionVisualHeight handles description visual height for Model and returns the resulting state or error.
func (m Model) descriptionVisualHeight() int {
	width := m.descriptionTextWidth()
	return m.descriptionVisualHeightForWidth(width)
}

// descriptionTextWidth handles description text width for Model and returns the resulting state or error.
func (m Model) descriptionTextWidth() int {
	return max(m.width-6, 1)
}

// descriptionVisualHeightForWidth handles description visual height for width for Model and returns the resulting state or error.
func (m Model) descriptionVisualHeightForWidth(width int) int {
	value := m.descInput.Value()
	if value == "" {
		value = "No description."
	}
	return max(len(descriptionWrapSegments(value, width)), 1)
}

// renderActiveDescription renders render active description for Model and returns the resulting state or error.
func (m Model) renderActiveDescription(width int) string {
	value := m.descInput.Value()
	segments := descriptionWrapSegments(value, width)
	cursor := m.descInput.Column()
	lines := make([]string, 0, len(segments))
	cursorLine, cursorCol := wrappedCursorPosition(segments, cursor)
	for i, segment := range segments {
		if i == cursorLine {
			lines = append(lines, renderWithCursor(segment.text, cursorCol, width))
		} else {
			lines = append(lines, styles.FilterText.Render(padRight(segment.text, width)))
		}
	}
	return strings.Join(lines, "\n")
}

// descriptionSegment holds description segment state used by the details package.
type descriptionSegment struct {
	// text stores text for descriptionSegment.
	text string
	// start, end store start end values for descriptionSegment.
	start, end int
}

// descriptionWrapSegments handles description wrap segments and returns the resulting value or error.
func descriptionWrapSegments(value string, width int) []descriptionSegment {
	if width <= 0 {
		return []descriptionSegment{{text: ""}}
	}
	if value == "" {
		return []descriptionSegment{{text: ""}}
	}
	wrapped := ansi.Wordwrap(value, width, " ")
	parts := strings.Split(wrapped, "\n")
	if len(parts) == 0 {
		return []descriptionSegment{{text: ""}}
	}
	valueRunes := []rune(value)
	pos := 0
	segments := make([]descriptionSegment, 0, len(parts))
	for _, part := range parts {
		partRunes := []rune(part)
		for len(partRunes) > 0 && pos < len(valueRunes) && valueRunes[pos] != partRunes[0] {
			pos++
		}
		start := pos
		for _, r := range partRunes {
			for pos < len(valueRunes) && valueRunes[pos] != r {
				pos++
			}
			if pos < len(valueRunes) {
				pos++
			}
		}
		segments = append(segments, descriptionSegment{text: part, start: start, end: pos})
	}
	for pos < len(valueRunes) && valueRunes[pos] == ' ' {
		if len(segments) == 0 || lipgloss.Width(segments[len(segments)-1].text) >= width {
			segments = append(segments, descriptionSegment{text: "", start: pos, end: pos})
		}
		last := &segments[len(segments)-1]
		last.text += " "
		pos++
		last.end = pos
	}
	return segments
}

// wrappedOffsetForClick handles wrapped offset for click and returns the resulting value or error.
func wrappedOffsetForClick(segments []descriptionSegment, line, col int) int {
	if len(segments) == 0 {
		return 0
	}
	line = min(max(line, 0), len(segments)-1)
	segment := segments[line]
	return segment.start + min(max(col, 0), len([]rune(segment.text)))
}

// wrappedCursorPosition handles wrapped cursor position and returns the resulting value or error.
func wrappedCursorPosition(segments []descriptionSegment, cursor int) (int, int) {
	if len(segments) == 0 {
		return 0, 0
	}
	for i, segment := range segments {
		if cursor >= segment.start && cursor <= segment.end {
			return i, min(max(cursor-segment.start, 0), len([]rune(segment.text)))
		}
		if cursor < segment.start {
			return i, 0
		}
	}
	last := len(segments) - 1
	return last, len([]rune(segments[last].text))
}

// renderWithCursor renders render with cursor and returns the resulting value or error.
func renderWithCursor(value string, cursorCol, width int) string {
	runes := []rune(value)
	cursorCol = min(max(cursorCol, 0), len(runes))
	before := styles.FilterText.Render(string(runes[:cursorCol]))
	cursorChar := " "
	after := ""
	if cursorCol < len(runes) {
		cursorChar = string(runes[cursorCol])
		after = string(runes[cursorCol+1:])
	}
	rendered := before + descriptionCursorStyle().Render(styles.FilterText.Render(cursorChar)) + styles.FilterText.Render(after)
	return padRight(rendered, width)
}

// descriptionCursorStyle handles description cursor style and returns the resulting value or error.
func descriptionCursorStyle() lipgloss.Style {
	if styles.NoColorEnabled() {
		return lipgloss.NewStyle().Reverse(true).Bold(true)
	}
	return lipgloss.NewStyle().Background(styles.PaletteYellow).Foreground(styles.PaletteBlueDeep).Bold(true)
}

// singleLineValue handles single line value and returns the resulting value or error.
func singleLineValue(value string) string {
	return strings.Join(strings.FieldsFunc(value, func(r rune) bool {
		return r == '\n' || r == '\r'
	}), " ")
}

// selectedGroupKey selects selected group key for Model and returns the resulting state or error.
func (m Model) selectedGroupKey() string {
	return m.groupKey
}

// selectedType selects selected type for Model and returns the resulting state or error.
func (m Model) selectedType() string {
	return m.typeValue
}

// fieldChanged handles field changed for Model and returns the resulting state or error.
func (m Model) fieldChanged(field fieldID) bool {
	if m.data == nil {
		return false
	}
	switch field {
	case fieldGroup:
		return core.NormalizeRemoteConfigGroupKey(m.selectedGroupKey()) != core.NormalizeRemoteConfigGroupKey(m.data.GroupKey)
	case fieldName:
		return strings.TrimSpace(m.nameInput.Value()) != m.data.Parameter.Key
	case fieldType:
		return m.selectedType() != strings.ToUpper(parameterType(m.data.Parameter))
	case fieldDescription:
		return m.descInput.Value() != m.data.Parameter.Description
	default:
		return false
	}
}

// valueChanged handles value changed for Model and returns the resulting state or error.
func (m Model) valueChanged() bool {
	return len(m.valueEdits()) > 0
}

// valueEdits handles value edits for Model and returns the resulting state or error.
func (m Model) valueEdits() []core.ParameterValueEdit {
	if m.data == nil {
		return nil
	}
	original := make(map[string]string, len(m.originalParam.Values))
	for _, value := range m.originalParam.Values {
		original[value.Label] = value.RawValue
	}
	edits := make([]core.ParameterValueEdit, 0)
	for _, value := range m.data.Parameter.Values {
		if original[value.Label] == value.RawValue {
			continue
		}
		edits = append(edits, core.ParameterValueEdit{
			Label:     value.Label,
			NextValue: value.RawValue,
		})
	}
	return edits
}

// invalidName handles invalid name for Model and returns the resulting state or error.
func (m Model) invalidName() bool {
	if m.data == nil {
		return false
	}
	nextKey := strings.TrimSpace(m.nameInput.Value())
	if nextKey == "" {
		return true
	}
	for _, key := range m.data.ParameterKeys {
		if key == m.data.Parameter.Key {
			continue
		}
		if key == nextKey {
			return true
		}
	}
	return false
}

// invalidValues handles invalid values for Model and returns the resulting state or error.
func (m Model) invalidValues() bool {
	if m.valuesInvalid {
		return true
	}
	if m.data == nil {
		return false
	}
	valueType := m.selectedType()
	for _, value := range m.data.Parameter.Values {
		if !value.Plain {
			continue
		}
		if !validRawValueForType(value.RawValue, valueType) {
			return true
		}
	}
	return false
}

// validRawValueForType handles valid raw value for type and returns the resulting value or error.
func validRawValueForType(value, valueType string) bool {
	value = strings.TrimSpace(value)
	switch strings.ToUpper(strings.TrimSpace(valueType)) {
	case "STRING", "":
		return true
	case "BOOLEAN":
		return value == "true" || value == "false"
	case "NUMBER":
		var number float64
		if err := json.Unmarshal([]byte(value), &number); err != nil {
			return false
		}
		return !math.IsInf(number, 0) && !math.IsNaN(number)
	case "JSON":
		return json.Valid([]byte(value))
	default:
		return false
	}
}

// hasChanges reports has changes for Model and returns the resulting state or error.
func (m Model) hasChanges() bool {
	return m.fieldChanged(fieldGroup) || m.fieldChanged(fieldName) || m.fieldChanged(fieldType) || m.fieldChanged(fieldDescription) || m.valueChanged()
}

// dropdownRow holds dropdown row state used by the details package.
type dropdownRow struct {
	// Key stores key for dropdownRow.
	Key string
	// Label stores label for dropdownRow.
	Label string
	// Input stores input for dropdownRow.
	Input bool
}

// dropdownRows handles dropdown rows for Model and returns the resulting state or error.
func (m Model) dropdownRows() []dropdownRow {
	switch m.activeField {
	case fieldGroup:
		if m.data == nil {
			return nil
		}
		out := make([]dropdownRow, 0, len(m.data.Groups)+1)
		root := dropdownRow{}
		hasRoot := false
		for _, group := range m.data.Groups {
			if core.NormalizeRemoteConfigGroupKey(group.Key) == core.NormalizeRemoteConfigGroupKey(m.groupKey) {
				continue
			}
			if core.NormalizeRemoteConfigGroupKey(group.Key) == "" {
				root = dropdownRow{Key: group.Key, Label: group.Label}
				hasRoot = true
				continue
			}
			out = append(out, dropdownRow{Key: group.Key, Label: group.Label})
		}
		out = append(out, dropdownRow{Input: true, Label: m.groupInput.Placeholder})
		if hasRoot {
			out = append(out, root)
		}
		return out
	case fieldType:
		out := make([]dropdownRow, 0, len(typeOptions)-1)
		for _, option := range typeOptions {
			if option == m.typeValue {
				continue
			}
			out = append(out, dropdownRow{Key: option, Label: option})
		}
		return out
	default:
		return nil
	}
}

// fieldValueLine handles field value line for Model and returns the resulting state or error.
func (m Model) fieldValueLine(field fieldID) int {
	if m.data == nil {
		return 0
	}
	width := max(m.width-5, 1)
	line := 0
	line += 1 + len(wrappedLines(displayProject(m.data.Project), width)) + 1
	if field == fieldGroup {
		return line + 1
	}
	line += 3
	if field == fieldName {
		return line + 1
	}
	line += 3
	if field == fieldType {
		return line + 1
	}
	line += 3
	if field == fieldDescription {
		return line + 1
	}
	return 0
}

// valuesTitleLine handles values title line for Model and returns the resulting state or error.
func (m Model) valuesTitleLine() int {
	return m.fieldValueLine(fieldDescription) + m.descriptionVisualHeight() + 1
}

// valueConditionLine handles value condition line for Model and returns the resulting state or error.
func (m Model) valueConditionLine(index int) int {
	if m.data == nil {
		return 0
	}
	width := max(m.width-5, 1)
	line := m.valuesTitleLine() + 1
	for i, value := range m.data.Parameter.Values {
		if i == index {
			return line
		}
		line += 1 + len(wrappedLines(value.Value, max(width-4, 1))) + 1
	}
	return line
}

// dropdownCurrentLabel handles dropdown current label for Model and returns the resulting state or error.
func (m Model) dropdownCurrentLabel() string {
	switch m.activeField {
	case fieldGroup:
		return m.groupLabel
	case fieldType:
		return m.typeValue
	default:
		return ""
	}
}

// dropdownCurrentStyle handles dropdown current style for Model and returns the resulting state or error.
func (m Model) dropdownCurrentStyle() lipgloss.Style {
	switch m.activeField {
	case fieldGroup:
		return groupValueStyle
	default:
		return styles.PanelText
	}
}

// openDropdown opens open dropdown for Model and returns the resulting state or error.
func (m *Model) openDropdown(delta int) {
	rows := m.dropdownRows()
	if len(rows) == 0 {
		return
	}
	m.dropdownOpen = true
	m.dropdownIndex = 0
	if delta < 0 {
		m.dropdownIndex = len(rows) - 1
	}
	if rows[m.dropdownIndex].Input {
		_ = m.groupInput.Focus()
	} else {
		m.groupInput.Blur()
	}
}

// closeDropdown closes close dropdown for Model and returns the resulting state or error.
func (m *Model) closeDropdown() {
	m.dropdownOpen = false
	m.dropdownIndex = 0
	m.groupInput = newGroupInput()
}

// moveDropdown moves move dropdown for Model and returns the resulting state or error.
func (m *Model) moveDropdown(delta int) {
	rows := m.dropdownRows()
	if len(rows) == 0 {
		return
	}
	m.dropdownIndex = (m.dropdownIndex + delta + len(rows)) % len(rows)
	if rows[m.dropdownIndex].Input {
		_ = m.groupInput.Focus()
	} else {
		m.groupInput.Blur()
	}
}

// commitDropdown handles commit dropdown for Model and returns the resulting state or error.
func (m *Model) commitDropdown() {
	rows := m.dropdownRows()
	if len(rows) == 0 || m.dropdownIndex < 0 || m.dropdownIndex >= len(rows) {
		return
	}
	row := rows[m.dropdownIndex]
	if row.Input {
		value := strings.TrimSpace(m.groupInput.Value())
		if value == "" {
			return
		}
		m.groupKey = value
		m.groupLabel = value
	} else {
		switch m.activeField {
		case fieldGroup:
			m.groupKey = row.Key
			m.groupLabel = row.Label
		case fieldType:
			m.typeValue = row.Key
		}
	}
	m.closeDropdown()
}

// dropdownInputSelected handles dropdown input selected for Model and returns the resulting state or error.
func (m Model) dropdownInputSelected() bool {
	rows := m.dropdownRows()
	return m.dropdownIndex >= 0 && m.dropdownIndex < len(rows) && rows[m.dropdownIndex].Input
}

var dropdownBorderStyle = lipgloss.NewStyle().Foreground(styles.PaletteBlueBright)

// dropdownOptionStyle handles dropdown option style and returns the resulting value or error.
func dropdownOptionStyle(selected bool) lipgloss.Style {
	if !selected {
		return styles.PanelText
	}
	if styles.NoColorEnabled() {
		return lipgloss.NewStyle().Bold(true).Reverse(true)
	}
	return styles.PanelText.Bold(true).Foreground(styles.PaletteGold)
}

// textinputStyles handles textinput styles and returns the resulting value or error.
func textinputStyles() textinput.Styles {
	inputStyles := textinput.DefaultDarkStyles()
	valueStyle := styles.FilterText
	inputStyles.Focused.Text = valueStyle
	inputStyles.Focused.Prompt = valueStyle
	inputStyles.Focused.Placeholder = valueStyle
	inputStyles.Focused.Suggestion = valueStyle
	inputStyles.Blurred.Text = valueStyle
	inputStyles.Blurred.Prompt = valueStyle
	inputStyles.Blurred.Placeholder = valueStyle
	inputStyles.Blurred.Suggestion = valueStyle
	inputStyles.Cursor.Color = styles.PaletteYellow
	return inputStyles
}

// textareaStyles handles textarea styles and returns the resulting value or error.
func textareaStyles() textarea.Styles {
	s := textarea.DefaultStyles(true)
	textStyle := styles.FilterText
	s.Focused.Text = textStyle
	s.Focused.Prompt = lipgloss.NewStyle()
	s.Focused.Placeholder = styles.PanelMuted
	s.Focused.LineNumber = lipgloss.NewStyle()
	s.Focused.CursorLineNumber = lipgloss.NewStyle()
	s.Focused.CursorLine = textStyle
	s.Focused.EndOfBuffer = lipgloss.NewStyle()
	s.Blurred.Text = textStyle
	s.Blurred.Prompt = lipgloss.NewStyle()
	s.Blurred.Placeholder = styles.PanelMuted
	s.Blurred.LineNumber = lipgloss.NewStyle()
	s.Blurred.CursorLineNumber = lipgloss.NewStyle()
	s.Blurred.CursorLine = lipgloss.NewStyle()
	s.Blurred.EndOfBuffer = lipgloss.NewStyle()
	s.Cursor.Color = styles.PaletteYellow
	return s
}

// newTextInput constructs new text input and returns the resulting value or error.
func newTextInput() textinput.Model {
	input := textinput.New()
	input.Prompt = ""
	input.SetStyles(textinputStyles())
	input.Blur()
	return input
}

// newDescriptionInput constructs new description input and returns the resulting value or error.
func newDescriptionInput() textarea.Model {
	input := textarea.New()
	input.Prompt = ""
	input.ShowLineNumbers = false
	input.EndOfBufferCharacter = ' '
	input.SetVirtualCursor(false)
	input.SetStyles(textareaStyles())
	input.Blur()
	return input
}

// newGroupInput constructs new group input and returns the resulting value or error.
func newGroupInput() textinput.Model {
	input := newTextInput()
	input.Placeholder = "New group"
	return input
}

// conditionStyle handles condition style for Model and returns the resulting state or error.
func (m Model) conditionStyle(color string) lipgloss.Style {
	return styles.PanelText.Foreground(styles.ConditionLipglossColor(color))
}

// valueTextStyle handles value text style for Model and returns the resulting state or error.
func (m Model) valueTextStyle(value core.ParametersValue) lipgloss.Style {
	if value.Empty {
		return corestyles.EmptyValueStyle()
	}
	return corestyles.ValueTextStyle(value.Value, value.ValueType)
}

// renderValueLines renders render value lines for Model and returns the resulting state or error.
func (m Model) renderValueLines(value core.ParametersValue, width int) []string {
	if value.Empty {
		return []string{corestyles.EmptyValueStyle().Render(value.Value)}
	}
	switch strings.TrimSpace(strings.ToLower(value.ValueType)) {
	case "json":
		return renderJSONValueLines(value.RawValue, width)
	case "string", "":
		if strings.Contains(value.RawValue, "\n") {
			return renderPlainValueLines(value.RawValue, width, corestyles.ValueTextStyle(value.RawValue, value.ValueType))
		}
	}
	return renderPlainValueLines(value.Value, width, m.valueTextStyle(value))
}

// renderPlainValueLines renders render plain value lines and returns the resulting value or error.
func renderPlainValueLines(value string, width int, style lipgloss.Style) []string {
	lines := make([]string, 0)
	for part := range strings.SplitSeq(value, "\n") {
		for _, line := range wrapLine(part, width) {
			lines = append(lines, style.Render(line))
		}
	}
	if len(lines) == 0 {
		return []string{style.Render("")}
	}
	return lines
}

// renderJSONValueLines renders render jsonvalue lines and returns the resulting value or error.
func renderJSONValueLines(value string, width int) []string {
	var out bytes.Buffer
	if err := json.Indent(&out, []byte(value), "", "  "); err != nil {
		return renderPlainValueLines(value, width, corestyles.ValueTextStyle(value, "json"))
	}
	lines := strings.Split(out.String(), "\n")
	rendered := make([]string, 0, len(lines))
	for _, line := range lines {
		highlighted := jsoninput.HighlightJSONVisible(line)
		rendered = append(rendered, wrapRenderedLine(highlighted, width)...)
	}
	return rendered
}

// wrapRenderedLine handles wrap rendered line and returns the resulting value or error.
func wrapRenderedLine(value string, width int) []string {
	if width <= 0 {
		return []string{""}
	}
	if lipgloss.Width(value) <= width {
		return []string{value}
	}
	lines := make([]string, 0)
	remaining := value
	for lipgloss.Width(remaining) > width {
		part := ansi.Truncate(remaining, width, "")
		lines = append(lines, part)
		remaining = ansi.Cut(remaining, lipgloss.Width(part), lipgloss.Width(remaining))
	}
	lines = append(lines, remaining)
	return lines
}
