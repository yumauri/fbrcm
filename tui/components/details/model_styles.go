package details

import (
	"charm.land/lipgloss/v2"

	tuiconfig "github.com/yumauri/fbrcm/tui/config"
	"github.com/yumauri/fbrcm/tui/styles"
)

const (
	panelTitleLabel = "Details"
)

func panelTitleKey() string {
	return tuiconfig.ActionKeyHint(tuiconfig.BlockGlobal, tuiconfig.ActionFocusDetails)
}

var (
	labelStyle             = styles.DetailsLabel
	projectValueStyle      = styles.DetailsProjectValue
	groupValueStyle        = styles.ParameterGroup
	parameterKeyStyle      = styles.ParameterName
	selectedValueStyle     = lipgloss.NewStyle().Background(styles.PaletteBlueDeep).Foreground(styles.PaletteSlateBright)
	conditionDefaultStyle  = styles.DetailsEmptyValue
	fieldDirtyStyle        = styles.PanelMuted.Bold(true).Underline(true)
	fieldInvalidStyle      = lipgloss.NewStyle().Foreground(styles.PaletteError)
	fieldInvalidDirtyStyle = lipgloss.NewStyle().
				Foreground(styles.PaletteError).
				Bold(true).
				Underline(true)
)

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
	fieldConditionPriority
	fieldConditionColor
)

var typeOptions = []string{"STRING", "BOOLEAN", "NUMBER", "JSON"}
