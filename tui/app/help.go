package app

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"

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
		key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
	}

	if k.keyboardCapture {
		return append(common,
			key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "close filter")),
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "apply")),
			key.NewBinding(key.WithKeys("~", "^", "/", "="), key.WithHelp("~ ^ / =", "filter")),
		)
	}

	switch k.active {
	case panels.Projects:
		modeLabel := "collapse"
		if k.projectsMode == projectsPanelModeCollapsed {
			modeLabel = "expand"
		}
		return append(common,
			key.NewBinding(key.WithKeys("f9"), key.WithHelp("F9", modeLabel)),
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
			key.NewBinding(key.WithKeys("space"), key.WithHelp("space", "mark")),
			key.NewBinding(key.WithKeys("o"), key.WithHelp("o", "open")),
			key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
			key.NewBinding(key.WithKeys("~", "^", "/", "="), key.WithHelp("~ ^ / =", "filter")),
		)
	case panels.Parameters:
		return append(common,
			key.NewBinding(key.WithKeys("f11"), key.WithHelp("F11", "maximize")),
			key.NewBinding(key.WithKeys("f2"), key.WithHelp("F2", "rename")),
			key.NewBinding(key.WithKeys("f4"), key.WithHelp("F4", "edit")),
			key.NewBinding(key.WithKeys("shift+f4"), key.WithHelp("S-F4", "new")),
			key.NewBinding(key.WithKeys("f5"), key.WithHelp("F5", "duplicate")),
			key.NewBinding(key.WithKeys("f6"), key.WithHelp("F6", "move")),
			key.NewBinding(key.WithKeys("space"), key.WithHelp("space", "toggle")),
			key.NewBinding(key.WithKeys("f8"), key.WithHelp("F8", "delete")),
			key.NewBinding(key.WithKeys("p", "P"), key.WithHelp("p/P", "publish")),
			key.NewBinding(key.WithKeys("d", "D"), key.WithHelp("d/D", "discard")),
			key.NewBinding(key.WithKeys("y", "Y"), key.WithHelp("y/Y", "copy")),
			key.NewBinding(key.WithKeys("r", "R"), key.WithHelp("r/R", "reload")),
			key.NewBinding(key.WithKeys("~", "^", "/", "="), key.WithHelp("~ ^ / =", "filter")),
		)
	case panels.Logs:
		modeLabel := "collapse"
		if k.logsMode == logsPanelModeCollapsed {
			modeLabel = "expand"
		}
		return append(common,
			key.NewBinding(key.WithKeys("f9"), key.WithHelp("F9", modeLabel)),
			key.NewBinding(key.WithKeys("[", "]"), key.WithHelp("[/]", "level")),
			key.NewBinding(key.WithKeys("-", "_", "=", "+"), key.WithHelp("-/+", "resize")),
		)
	case panels.Details:
		return append(common,
			key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "close")),
			key.NewBinding(key.WithKeys("f8"), key.WithHelp("F8", "delete")),
		)
	default:
		return common
	}
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
