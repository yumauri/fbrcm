package styles

import (
	"image/color"

	"charm.land/lipgloss/v2"
	charmlog "github.com/charmbracelet/log"

	corestyles "github.com/yumauri/fbrcm/core/styles"
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

	ColorRowStripe = corestyles.ColorRowStripe
	ColorAdded     = corestyles.ColorAdded
	ColorRemoved   = corestyles.ColorRemoved
	ColorChanged   = corestyles.ColorChanged
	ColorNote      = corestyles.ColorNote

	PanelMuted = lipgloss.NewStyle().
			Foreground(PaletteSlateDim)

	PanelText = lipgloss.NewStyle().
			Foreground(PaletteSlateBright)

	BorderInactive = lipgloss.NewStyle().
			Foreground(PaletteSlateDim)
)

// NoColorEnabled handles no color enabled and returns the resulting value or error.
func NoColorEnabled() bool {
	return corestyles.NoColorEnabled()
}

// BorderStyle handles border style and returns the resulting value or error.
func BorderStyle(_ bool) lipgloss.Style {
	return BorderInactive
}

// ConditionLipglossColor handles condition lipgloss color and returns the resulting value or error.
func ConditionLipglossColor(name string) color.Color {
	return corestyles.ConditionLipglossColor(name)
}

// LogLevelLipglossColor handles log level lipgloss color and returns the resulting value or error.
func LogLevelLipglossColor(level charmlog.Level) color.Color {
	return corestyles.LogLevelLipglossColor(level)
}
