package styles

import (
	"charm.land/lipgloss/v2"

	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
)

// EvaluationStatusStyle is the shared color for an evaluation symbol.
func EvaluationStatusStyle(symbol string) lipgloss.Style {
	style := lipgloss.NewStyle()
	if NoColorEnabled() {
		return style
	}
	switch symbol {
	case "×":
		return style.Foreground(PaletteError)
	case "✓":
		return style.Foreground(PaletteSuccess)
	case "◆":
		return style.Foreground(PaletteBlueBright)
	default:
		return style
	}
}

// OutcomeNoteStyle applies the shared value outcome colors, with a caller
// supplied style for other labels such as unresolved conditions.
func OutcomeNoteStyle(label string, fallback lipgloss.Style) lipgloss.Style {
	if NoColorEnabled() {
		return lipgloss.NewStyle()
	}
	switch label {
	case rcdisplay.OutcomeDoesNotMatch, rcdisplay.OutcomeNotUsed:
		return EvaluationStatusStyle("×")
	case rcdisplay.OutcomeLowerPriority:
		return EvaluationStatusStyle("✓")
	case rcdisplay.OutcomeSelected:
		return lipgloss.NewStyle().Foreground(PaletteSlateBright).Background(ColorHistoryAddedBackground)
	default:
		return fallback
	}
}
