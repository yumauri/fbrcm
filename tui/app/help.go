package app

import (
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"

	tuiconfig "github.com/yumauri/fbrcm/tui/config"
	"github.com/yumauri/fbrcm/tui/panels"
	"github.com/yumauri/fbrcm/tui/styles"
)

const helpLineHeight = 1

type helpKeyMap struct {
	active          panels.ID
	keyboardCapture bool
	projectsMode    projectsPanelMode
	logsMode        logsPanelMode
	detailsVisible  bool
	conditionDetail bool
	conditionMove   bool
}

func newHelpModel() help.Model {
	m := help.New()
	m.ShortSeparator = " • "
	m.Styles.ShortKey = styles.FilterText
	m.Styles.ShortDesc = styles.PanelMuted
	m.Styles.ShortSeparator = styles.PanelMuted
	m.Styles.Ellipsis = styles.PanelMuted
	return m
}

func (k helpKeyMap) ShortHelp() []key.Binding {
	if k.conditionMove {
		return conditionMoveHelp()
	}
	common := []key.Binding{
		tuiconfig.Binding(tuiconfig.BlockGlobal, tuiconfig.ActionQuit, "quit"),
		tuiconfig.Binding(tuiconfig.BlockGlobal, tuiconfig.ActionHelp, "help"),
	}

	if k.keyboardCapture {
		return append(common, captureHelp()...)
	}

	switch k.active {
	case panels.Projects:
		return append(common, k.projectsHelp()...)
	case panels.Parameters:
		return append(common, parametersHelp()...)
	case panels.Conditions:
		return append(common, conditionsHelp()...)
	case panels.History:
		return append(common, historyHelp()...)
	case panels.Logs:
		return append(common, k.logsHelp()...)
	case panels.Details:
		if k.conditionDetail {
			return append(common, conditionDetailsHelp()...)
		}
		return append(common, detailsHelp()...)
	default:
		return common
	}
}

func conditionMoveHelp() []key.Binding {
	return []key.Binding{
		tuiconfig.Binding(tuiconfig.BlockMoveInput, tuiconfig.ActionUp, "move up"),
		tuiconfig.Binding(tuiconfig.BlockMoveInput, tuiconfig.ActionDown, "move down"),
		tuiconfig.Binding(tuiconfig.BlockMoveInput, tuiconfig.ActionSubmit, "place"),
		tuiconfig.Binding(tuiconfig.BlockMoveInput, tuiconfig.ActionCancel, "cancel"),
	}
}

func conditionsHelp() []key.Binding {
	return []key.Binding{
		tuiconfig.Binding(tuiconfig.BlockParameters, tuiconfig.ActionToggleMaximize, "maximize"),
		tuiconfig.Binding(tuiconfig.BlockConditions, tuiconfig.ActionRename, "rename"),
		tuiconfig.Binding(tuiconfig.BlockConditions, tuiconfig.ActionEdit, "expression"),
		tuiconfig.Binding(tuiconfig.BlockConditions, tuiconfig.ActionColor, "color"),
		tuiconfig.Binding(tuiconfig.BlockConditions, tuiconfig.ActionNew, "new"),
		tuiconfig.Binding(tuiconfig.BlockConditions, tuiconfig.ActionMove, "priority"),
		tuiconfig.Binding(tuiconfig.BlockConditions, tuiconfig.ActionDelete, "delete"),
		compoundBinding(ref(tuiconfig.BlockConditions, tuiconfig.ActionPublish), ref(tuiconfig.BlockConditions, tuiconfig.ActionPublishAll), "publish"),
		compoundBinding(ref(tuiconfig.BlockConditions, tuiconfig.ActionDiscard), ref(tuiconfig.BlockConditions, tuiconfig.ActionDiscardAll), "discard"),
		tuiconfig.Binding(tuiconfig.BlockConditions, tuiconfig.ActionOpenDetails, "details"),
		compoundBinding(ref(tuiconfig.BlockConditions, tuiconfig.ActionCopyName), ref(tuiconfig.BlockConditions, tuiconfig.ActionCopyPath), "copy"),
		compoundBinding(ref(tuiconfig.BlockParameters, tuiconfig.ActionReload), ref(tuiconfig.BlockParameters, tuiconfig.ActionReloadAll), "update"),
		filterBinding(),
	}
}

func conditionDetailsHelp() []key.Binding {
	return []key.Binding{
		tuiconfig.Binding(tuiconfig.BlockDetails, tuiconfig.ActionClose, "close"),
		tuiconfig.Binding(tuiconfig.BlockDetails, tuiconfig.ActionRename, "rename"),
		tuiconfig.Binding(tuiconfig.BlockDetails, tuiconfig.ActionEditValue, "expression"),
		tuiconfig.Binding(tuiconfig.BlockDetails, tuiconfig.ActionColor, "color"),
		tuiconfig.Binding(tuiconfig.BlockDetails, tuiconfig.ActionMove, "priority"),
		tuiconfig.Binding(tuiconfig.BlockDetails, tuiconfig.ActionDelete, "delete"),
		compoundBinding(ref(tuiconfig.BlockDetails, tuiconfig.ActionCopyName), ref(tuiconfig.BlockDetails, tuiconfig.ActionCopyPath), "copy"),
		tuiconfig.Binding(tuiconfig.BlockDetails, tuiconfig.ActionCopyValue, "copy expression"),
	}
}

func historyHelp() []key.Binding {
	return []key.Binding{
		tuiconfig.Binding(tuiconfig.BlockHistory, tuiconfig.ActionHistoryChanges, "changes only"),
		compoundBinding(ref(tuiconfig.BlockHistory, tuiconfig.ActionHistoryBothOlder), ref(tuiconfig.BlockHistory, tuiconfig.ActionHistoryBothNewer), "both versions"),
		tuiconfig.Binding(tuiconfig.BlockHistory, tuiconfig.ActionHistoryChoose, "choose versions"),
		tuiconfig.Binding(tuiconfig.BlockParameters, tuiconfig.ActionToggleMaximize, "maximize"),
		tuiconfig.Binding(tuiconfig.BlockParameters, tuiconfig.ActionToggle, "toggle"),
		compoundBinding(ref(tuiconfig.BlockParameters, tuiconfig.ActionCopyName), ref(tuiconfig.BlockParameters, tuiconfig.ActionCopyPath), "copy"),
		filterBinding(),
	}
}

func captureHelp() []key.Binding {
	return []key.Binding{
		tuiconfig.Binding(tuiconfig.BlockFilter, tuiconfig.ActionFilterCancel, "close filter"),
		tuiconfig.Binding(tuiconfig.BlockFilter, tuiconfig.ActionFilterApply, "apply"),
		filterBinding(),
	}
}

func (k helpKeyMap) projectsHelp() []key.Binding {
	modeLabel := "collapse"
	if k.projectsMode == projectsPanelModeCollapsed {
		modeLabel = "expand"
	}
	return []key.Binding{
		tuiconfig.Binding(tuiconfig.BlockProjects, tuiconfig.ActionToggleMode, modeLabel),
		tuiconfig.Binding(tuiconfig.BlockProjects, tuiconfig.ActionSelect, "select"),
		tuiconfig.Binding(tuiconfig.BlockProjects, tuiconfig.ActionMark, "mark"),
		tuiconfig.Binding(tuiconfig.BlockProjects, tuiconfig.ActionOpen, "open"),
		tuiconfig.Binding(tuiconfig.BlockProjects, tuiconfig.ActionRefresh, "update"),
		filterBinding(),
	}
}

func parametersHelp() []key.Binding {
	return []key.Binding{
		tuiconfig.Binding(tuiconfig.BlockParameters, tuiconfig.ActionToggleMaximize, "maximize"),
		tuiconfig.Binding(tuiconfig.BlockParameters, tuiconfig.ActionRename, "rename"),
		tuiconfig.Binding(tuiconfig.BlockParameters, tuiconfig.ActionEdit, "edit"),
		tuiconfig.Binding(tuiconfig.BlockParameters, tuiconfig.ActionNew, "new"),
		tuiconfig.Binding(tuiconfig.BlockParameters, tuiconfig.ActionDuplicate, "duplicate"),
		tuiconfig.Binding(tuiconfig.BlockParameters, tuiconfig.ActionMove, "move"),
		tuiconfig.Binding(tuiconfig.BlockParameters, tuiconfig.ActionToggle, "toggle"),
		tuiconfig.Binding(tuiconfig.BlockParameters, tuiconfig.ActionDelete, "delete"),
		compoundBinding(ref(tuiconfig.BlockParameters, tuiconfig.ActionPublish), ref(tuiconfig.BlockParameters, tuiconfig.ActionPublishAll), "publish"),
		compoundBinding(ref(tuiconfig.BlockParameters, tuiconfig.ActionDiscard), ref(tuiconfig.BlockParameters, tuiconfig.ActionDiscardAll), "discard"),
		compoundBinding(ref(tuiconfig.BlockParameters, tuiconfig.ActionCopyName), ref(tuiconfig.BlockParameters, tuiconfig.ActionCopyPath), "copy"),
		compoundBinding(ref(tuiconfig.BlockParameters, tuiconfig.ActionReload), ref(tuiconfig.BlockParameters, tuiconfig.ActionReloadAll), "update"),
		filterBinding(),
	}
}

func (k helpKeyMap) logsHelp() []key.Binding {
	modeLabel := "collapse"
	if k.logsMode == logsPanelModeCollapsed {
		modeLabel = "expand"
	}
	return []key.Binding{
		tuiconfig.Binding(tuiconfig.BlockLogs, tuiconfig.ActionToggleMode, modeLabel),
		compoundBinding(ref(tuiconfig.BlockLogs, tuiconfig.ActionLevelDown), ref(tuiconfig.BlockLogs, tuiconfig.ActionLevelUp), "level"),
		compoundBinding(ref(tuiconfig.BlockLogs, tuiconfig.ActionResizeShrink), ref(tuiconfig.BlockLogs, tuiconfig.ActionResizeGrow), "resize"),
	}
}

func detailsHelp() []key.Binding {
	return []key.Binding{
		tuiconfig.Binding(tuiconfig.BlockDetails, tuiconfig.ActionClose, "close"),
		tuiconfig.Binding(tuiconfig.BlockDetails, tuiconfig.ActionNew, "add conditional value"),
		tuiconfig.Binding(tuiconfig.BlockDetails, tuiconfig.ActionRename, "rename"),
		tuiconfig.Binding(tuiconfig.BlockDetails, tuiconfig.ActionEditValue, "edit"),
		tuiconfig.Binding(tuiconfig.BlockDetails, tuiconfig.ActionMove, "move"),
		tuiconfig.Binding(tuiconfig.BlockDetails, tuiconfig.ActionDelete, "delete"),
		compoundBinding(ref(tuiconfig.BlockDetails, tuiconfig.ActionCopyName), ref(tuiconfig.BlockDetails, tuiconfig.ActionCopyPath), "copy"),
		tuiconfig.Binding(tuiconfig.BlockDetails, tuiconfig.ActionCopyValue, "copy value"),
	}
}

func filterBinding() key.Binding {
	return multiBinding("filter",
		ref(tuiconfig.BlockFilter, tuiconfig.ActionFilterFuzzy),
		ref(tuiconfig.BlockFilter, tuiconfig.ActionFilterStartsWith),
		ref(tuiconfig.BlockFilter, tuiconfig.ActionFilterIncludes),
		ref(tuiconfig.BlockFilter, tuiconfig.ActionFilterExact),
	)
}

type helpRef struct {
	block  tuiconfig.Block
	action tuiconfig.Action
}

func ref(block tuiconfig.Block, action tuiconfig.Action) helpRef {
	return helpRef{block: block, action: action}
}

func compoundBinding(first, second helpRef, desc string) key.Binding {
	return multiBinding(desc, first, second)
}

func multiBinding(desc string, refs ...helpRef) key.Binding {
	var keys []string
	var labels []string
	for _, item := range refs {
		itemKeys := tuiconfig.Keys(item.block, item.action)
		if len(itemKeys) == 0 {
			continue
		}
		keys = append(keys, itemKeys...)
		labels = append(labels, tuiconfig.Label(item.block, item.action))
	}
	label := strings.Join(labels, "/")
	binding := key.NewBinding(key.WithKeys(keys...), key.WithHelp(label, desc))
	binding.SetEnabled(len(keys) > 0)
	return binding
}

func (k helpKeyMap) FullHelp() [][]key.Binding {
	groups := make([][]key.Binding, 0, len(helpPaletteBlockOrder))
	catalog := helpPaletteCatalog()
	for _, block := range helpPaletteBlockOrder {
		var bindings []key.Binding
		for _, item := range catalog {
			if item.block != block {
				continue
			}
			binding := tuiconfig.Binding(item.block, item.action, strings.ToLower(item.title))
			if binding.Enabled() {
				bindings = append(bindings, binding)
			}
		}
		if len(bindings) > 0 {
			groups = append(groups, bindings)
		}
	}
	return groups
}

func (m Model) helpView() string {
	if m.width <= 0 {
		return ""
	}

	h := m.help
	h.SetWidth(m.width)
	line := h.View(helpKeyMap{
		active:          m.active,
		keyboardCapture: m.keyboardCaptured(),
		projectsMode:    m.projectsMode,
		logsMode:        m.logsMode,
		detailsVisible:  m.detailsVisible,
		conditionDetail: m.details.IsCondition(),
		conditionMove:   m.conditions.MoveActive(),
	})

	return lipgloss.NewStyle().
		Width(m.width).
		MaxHeight(helpLineHeight).
		Foreground(styles.PaletteSlateDim).
		Render(line)
}
