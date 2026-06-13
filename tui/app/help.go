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

// helpKeyMap holds help key map state used by the app package.
type helpKeyMap struct {
	// active stores active for helpKeyMap.
	active panels.ID
	// keyboardCapture stores keyboard capture for helpKeyMap.
	keyboardCapture bool
	// projectsMode stores projects mode for helpKeyMap.
	projectsMode projectsPanelMode
	// logsMode stores logs mode for helpKeyMap.
	logsMode logsPanelMode
	// detailsVisible stores details visible for helpKeyMap.
	detailsVisible bool
}

// newHelpModel constructs new help model and returns the resulting value or error.
func newHelpModel() help.Model {
	m := help.New()
	m.ShortSeparator = " • "
	m.Styles.ShortKey = styles.FilterText
	m.Styles.ShortDesc = styles.PanelMuted
	m.Styles.ShortSeparator = styles.PanelMuted
	m.Styles.Ellipsis = styles.PanelMuted
	return m
}

// ShortHelp handles short help for helpKeyMap and returns the resulting state or error.
func (k helpKeyMap) ShortHelp() []key.Binding {
	common := []key.Binding{
		tuiconfig.Binding(tuiconfig.BlockGlobal, tuiconfig.ActionQuit, "quit"),
	}

	if k.keyboardCapture {
		return append(common,
			tuiconfig.Binding(tuiconfig.BlockFilter, tuiconfig.ActionFilterCancel, "close filter"),
			tuiconfig.Binding(tuiconfig.BlockFilter, tuiconfig.ActionFilterApply, "apply"),
			filterBinding(),
		)
	}

	switch k.active {
	case panels.Projects:
		modeLabel := "collapse"
		if k.projectsMode == projectsPanelModeCollapsed {
			modeLabel = "expand"
		}
		return append(common,
			tuiconfig.Binding(tuiconfig.BlockProjects, tuiconfig.ActionToggleMode, modeLabel),
			tuiconfig.Binding(tuiconfig.BlockProjects, tuiconfig.ActionSelect, "select"),
			tuiconfig.Binding(tuiconfig.BlockProjects, tuiconfig.ActionMark, "mark"),
			tuiconfig.Binding(tuiconfig.BlockProjects, tuiconfig.ActionOpen, "open"),
			tuiconfig.Binding(tuiconfig.BlockProjects, tuiconfig.ActionRefresh, "refresh"),
			filterBinding(),
		)
	case panels.Parameters:
		return append(common,
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
			compoundBinding(ref(tuiconfig.BlockParameters, tuiconfig.ActionReload), ref(tuiconfig.BlockParameters, tuiconfig.ActionReloadAll), "reload"),
			filterBinding(),
		)
	case panels.Logs:
		modeLabel := "collapse"
		if k.logsMode == logsPanelModeCollapsed {
			modeLabel = "expand"
		}
		return append(common,
			tuiconfig.Binding(tuiconfig.BlockLogs, tuiconfig.ActionToggleMode, modeLabel),
			compoundBinding(ref(tuiconfig.BlockLogs, tuiconfig.ActionLevelDown), ref(tuiconfig.BlockLogs, tuiconfig.ActionLevelUp), "level"),
			compoundBinding(ref(tuiconfig.BlockLogs, tuiconfig.ActionResizeShrink), ref(tuiconfig.BlockLogs, tuiconfig.ActionResizeGrow), "resize"),
		)
	case panels.Details:
		return append(common,
			tuiconfig.Binding(tuiconfig.BlockDetails, tuiconfig.ActionClose, "close"),
			tuiconfig.Binding(tuiconfig.BlockDetails, tuiconfig.ActionRename, "rename"),
			tuiconfig.Binding(tuiconfig.BlockDetails, tuiconfig.ActionEditValue, "edit"),
			tuiconfig.Binding(tuiconfig.BlockDetails, tuiconfig.ActionMove, "move"),
			tuiconfig.Binding(tuiconfig.BlockDetails, tuiconfig.ActionDelete, "delete"),
			compoundBinding(ref(tuiconfig.BlockDetails, tuiconfig.ActionCopyName), ref(tuiconfig.BlockDetails, tuiconfig.ActionCopyPath), "copy"),
			tuiconfig.Binding(tuiconfig.BlockDetails, tuiconfig.ActionCopyValue, "copy value"),
		)
	default:
		return common
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

// FullHelp handles full help for helpKeyMap and returns the resulting state or error.
func (k helpKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{k.ShortHelp()}
}

// helpView handles help view for Model and returns the resulting state or error.
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
	})

	return lipgloss.NewStyle().
		Width(m.width).
		MaxHeight(helpLineHeight).
		Foreground(styles.PaletteSlateDim).
		Render(line)
}
