package projectio

import (
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/tui/components/buttonbar"
	moveparam "github.com/yumauri/fbrcm/tui/components/moveparam"
)

func (m Model) actionButtons() buttonbar.Model {
	var buttons []buttonbar.Button
	switch m.phase {
	case phaseImportFile:
		buttons = []buttonbar.Button{
			{Label: "Choose", Variant: buttonbar.VariantAccent},
			{Label: "Cancel", Variant: buttonbar.VariantAccent},
		}
	case phaseImportOptions:
		buttons = []buttonbar.Button{
			{Label: "Review Changes", Variant: buttonbar.VariantAccent},
			{Label: "Cancel", Variant: buttonbar.VariantAccent},
		}
	case phaseExportSource:
		buttons = []buttonbar.Button{
			{Label: "Continue", Variant: buttonbar.VariantAccent},
			{Label: "Cancel", Variant: buttonbar.VariantAccent},
		}
	case phaseExportPath:
		buttons = []buttonbar.Button{
			{Label: "Export", Variant: buttonbar.VariantAccent},
			{Label: "Cancel", Variant: buttonbar.VariantAccent},
		}
	case phaseDefaultsFormat:
		buttons = []buttonbar.Button{
			{Label: "Continue", Variant: buttonbar.VariantAccent},
			{Label: "Cancel", Variant: buttonbar.VariantAccent},
		}
	case phaseDefaultsPath:
		buttons = []buttonbar.Button{
			{Label: "Download", Variant: buttonbar.VariantAccent},
			{Label: "Cancel", Variant: buttonbar.VariantAccent},
		}
	}
	return buttonbar.New(buttons).SetSelected(m.buttonCursor).SetFocused(m.buttonsFocused)
}

func (m *Model) focusActionButtons() {
	m.buttonsFocused = true
	m.buttonCursor = min(m.buttonCursor, 1)
	for i := range m.optionInputs {
		m.optionInputs[i].Blur()
	}
}

func (m Model) updateFocusedActionButtons(key string) (Model, tea.Cmd) {
	buttons := m.actionButtons()
	switch key {
	case "left", "shift+tab":
		buttons.Move(-1)
		m.buttonCursor = buttons.Selected()
	case "right", "tab":
		buttons.Move(1)
		m.buttonCursor = buttons.Selected()
	case "up":
		m.buttonsFocused = false
		if m.phase == phaseImportOptions {
			m.optionCursor = optionRowCount - 1
			m.focusOptionInput()
		}
	case "enter", "space":
		return m.activateActionButton(m.buttonCursor)
	}
	return m, nil
}

func (m Model) activateActionButton(index int) (Model, tea.Cmd) {
	if index == 1 {
		return m.Close(), nil
	}
	if index != 0 {
		return m, nil
	}
	switch m.phase {
	case phaseImportFile:
		m.buttonsFocused = false
		return m.updateImportFile(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	case phaseImportOptions:
		cmd := m.importPlanCmd()
		m.workingFrom, m.phase = phaseImportOptions, phaseImportWorking
		return m, cmd
	case phaseExportSource:
		m.phase = phaseExportPath
		m.buttonsFocused = false
		return m, m.pathInput.Focus()
	case phaseExportPath:
		return m.submitExportPath()
	case phaseDefaultsFormat:
		m.phase = phaseDefaultsPath
		m.buttonsFocused = false
		m.pathInput.SetValue(defaultsPath(m.project.ProjectID, m.defaultsFormat))
		m.pathInput.CursorEnd()
		return m, m.pathInput.Focus()
	case phaseDefaultsPath:
		return m.submitDefaultsPath()
	default:
		return m, nil
	}
}

func (m Model) updateActionButtonsMouse(msg tea.Msg) (Model, tea.Cmd, bool) {
	var x, y int
	clicked := false
	switch msg := msg.(type) {
	case tea.MouseClickMsg:
		mouse := msg.Mouse()
		if mouse.Button != tea.MouseLeft {
			return m, nil, false
		}
		x, y, clicked = mouse.X, mouse.Y, true
	case tea.MouseMotionMsg:
		mouse := msg.Mouse()
		x, y = mouse.X, mouse.Y
	default:
		return m, nil, false
	}
	index, ok := m.actionButtonIndexAt(x, y)
	if !ok {
		return m, nil, false
	}
	m.buttonsFocused, m.buttonCursor = true, index
	if clicked {
		next, cmd := m.activateActionButton(index)
		return next, cmd, true
	}
	return m, nil, true
}

func (m *Model) openCurrentOptionSelector() {
	var (
		label    string
		options  []moveparam.Option
		selected int
		kind     optionSelectorKind
	)
	switch m.optionCursor {
	case optionStrategy:
		label = "Strategy    "
		kind = optionSelectorStrategy
		options = []moveparam.Option{
			{Key: string(core.ProjectImportMerge), Label: "Merge (keep current conflicts)"},
			{Key: string(core.ProjectImportReplace), Label: "Replace entire config"},
		}
		if m.strategy == core.ProjectImportReplace {
			selected = 1
		}
	case optionConditions:
		label = "Conditions  "
		kind = optionSelectorConditions
		options = m.conditionSelectorOptions()
		switch m.conditionPolicy {
		case core.ProjectImportKeepPortableConditions:
			selected = 1
		case core.ProjectImportRemoveAllConditions:
			selected = 2
		}
	default:
		return
	}
	x, y := m.optionSelectorAnchor(m.optionCursor)
	m.optionSelector = m.optionSelector.OpenOptions(x, y, label, options, selected)
	m.optionSelectorKind = kind
	m.focusOptionInput()
}

func (m Model) conditionSelectorOptions() []moveparam.Option {
	return []moveparam.Option{
		{Key: string(core.ProjectImportKeepConditions), Label: m.conditionPolicyLabel(core.ProjectImportKeepConditions)},
		{Key: string(core.ProjectImportKeepPortableConditions), Label: m.conditionPolicyLabel(core.ProjectImportKeepPortableConditions)},
		{Key: string(core.ProjectImportRemoveAllConditions), Label: m.conditionPolicyLabel(core.ProjectImportRemoveAllConditions)},
	}
}

func (m Model) updateOptionSelector(msg tea.Msg) (Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "esc":
		m.optionSelector = m.optionSelector.Close()
		m.optionSelectorKind = optionSelectorNone
	case "up", "k", "ctrl+k":
		return m, m.optionSelector.Move(-1)
	case "down", "j", "ctrl+j":
		return m, m.optionSelector.Move(1)
	case "enter":
		option, selected := m.optionSelector.Current()
		if !selected {
			return m, nil
		}
		switch m.optionSelectorKind {
		case optionSelectorStrategy:
			m.strategy = core.ProjectImportStrategy(option.Key)
		case optionSelectorConditions:
			m.conditionPolicy = core.ProjectConditionPolicy(option.Key)
		}
		m.optionSelector = m.optionSelector.Close()
		m.optionSelectorKind = optionSelectorNone
	default:
		m.optionSelector.Typeahead(key.String(), time.Now())
	}
	return m, nil
}

func (m Model) OptionSelectorOpen() bool               { return m.optionSelector.IsOpen() }
func (m Model) OptionSelectorHeaderView() string       { return m.optionSelector.HeaderView() }
func (m Model) OptionSelectorListView() string         { return m.optionSelector.ListView() }
func (m Model) OptionSelectorPosition() (int, int)     { return m.optionSelector.Position() }
func (m Model) OptionSelectorListPosition() (int, int) { return m.optionSelector.ListPosition() }
