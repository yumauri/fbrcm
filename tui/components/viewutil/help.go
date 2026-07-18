package viewutil

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"

	"github.com/yumauri/fbrcm/tui/styles"
)

// ShortHelpView renders the shared one-line TUI help style.
func ShortHelpView(width int, bindings ...key.Binding) string {
	m := help.New()
	m.ShortSeparator = " • "
	m.Styles.ShortKey = styles.FilterText
	m.Styles.ShortDesc = styles.PanelMuted
	m.Styles.ShortSeparator = styles.PanelMuted
	m.Styles.Ellipsis = styles.PanelMuted
	m.SetWidth(width)
	return m.ShortHelpView(bindings)
}

// HelpBinding creates a display binding for a literal key label.
func HelpBinding(keyLabel, description string) key.Binding {
	return key.NewBinding(
		key.WithKeys(keyLabel),
		key.WithHelp(keyLabel, description),
	)
}
