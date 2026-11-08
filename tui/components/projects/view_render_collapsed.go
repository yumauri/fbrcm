package projects

import (
	"charm.land/lipgloss/v2"

	tuiconfig "github.com/yumauri/fbrcm/tui/config"
	"github.com/yumauri/fbrcm/tui/styles"
)

var (
	panelTitleLabel = "Projects"

	itemStyle = styles.PanelText
	metaStyle = styles.PanelMuted
)

func panelTitleKey() string {
	return tuiconfig.ActionKeyHint(tuiconfig.BlockGlobal, tuiconfig.ActionFocusProjects)
}

type scrollbarState struct {
	visible    bool
	thumbStart int
	thumbEnd   int
}

type secondaryTitleState struct {
	text  string
	style lipgloss.Style
}
