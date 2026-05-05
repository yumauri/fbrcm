package styles

import (
	"image/color"

	"charm.land/lipgloss/v2"
	charmlog "github.com/charmbracelet/log"

	corestyles "fbrcm/core/styles"
)

var (
	PaletteBlueBright  = corestyles.ColorBlueBright
	PaletteBlueDeep    = corestyles.ColorBlueDeep
	PaletteOrange      = corestyles.ColorOrange
	PaletteYellow      = corestyles.ColorYellow
	PaletteGold        = corestyles.ColorGold
	PaletteSlateBright = corestyles.ColorSlateBright
	PaletteSlate       = corestyles.ColorSlate
	PaletteSlateDim    = corestyles.ColorSlateDim
	PaletteSlateDark   = corestyles.ColorSlateDark
	PaletteError       = corestyles.ColorError

	PanelBorderInactive = lipgloss.NewStyle().
				Foreground(PaletteSlateDark)

	PanelBorderActive = lipgloss.NewStyle().
				Foreground(PaletteBlueBright)

	PanelTitle = lipgloss.NewStyle().
			Foreground(PaletteSlateBright)

	PanelTitleActive = lipgloss.NewStyle().
				Bold(true).
				Foreground(PaletteYellow).
				Background(PaletteBlueDeep)

	PanelBody = lipgloss.NewStyle().
			Foreground(PaletteSlateDim)

	PanelMuted = lipgloss.NewStyle().
			Foreground(PaletteSlateDim)

	PanelText = lipgloss.NewStyle().
			Foreground(PaletteSlateBright)

	FilterText = lipgloss.NewStyle().
			Foreground(PaletteYellow)

	ProjectFocused = lipgloss.NewStyle().
			Background(PaletteBlueDeep)

	ProjectSelected = lipgloss.NewStyle().
			Background(PaletteSlateDark)

	ProjectFocusedSelected = lipgloss.NewStyle().
				Bold(true).
				Background(PaletteBlueBright)

	ScrollbarThumb = lipgloss.NewStyle().
			Foreground(PaletteYellow)

	SecondaryTitleSpinner = lipgloss.NewStyle().
				Foreground(PaletteOrange)

	SecondaryTitleCount = lipgloss.NewStyle().
				Foreground(PaletteGold)

	SecondaryTitleError = lipgloss.NewStyle().
				Bold(true).
				Foreground(PaletteError)
)

func BorderStyle(active bool) lipgloss.Style {
	if active {
		return PanelBorderActive
	}

	return PanelBorderInactive
}

func TitleStyle(active bool) lipgloss.Style {
	if !active {
		return PanelTitle
	}

	if NoColorEnabled() {
		return lipgloss.NewStyle().
			Bold(true).
			Reverse(true)
	}

	return PanelTitleActive
}

func ProjectStateStyle(cursor, selected bool) lipgloss.Style {
	switch {
	case NoColorEnabled() && cursor && selected:
		return lipgloss.NewStyle().Bold(true).Underline(true).Reverse(true)
	case NoColorEnabled() && cursor:
		return lipgloss.NewStyle().Bold(true).Reverse(true)
	case NoColorEnabled() && selected:
		return lipgloss.NewStyle().Underline(true).Reverse(true)
	case cursor && selected:
		return ProjectFocusedSelected
	case cursor:
		return ProjectFocused
	case selected:
		return ProjectSelected
	default:
		return lipgloss.NewStyle()
	}
}

func NoColorEnabled() bool {
	return corestyles.NoColorEnabled()
}

func ConditionLipglossColor(name string) color.Color {
	return corestyles.ConditionLipglossColor(name)
}

func LogLevelLipglossColor(level charmlog.Level) color.Color {
	return corestyles.LogLevelLipglossColor(level)
}
