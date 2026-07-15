package styles

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	charmlog "charm.land/log/v2"

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

func NoColorEnabled() bool {
	return corestyles.NoColorEnabled()
}

// BorderStyle returns the CLI border style. CLI panels do not use active borders.
func BorderStyle(_ bool) lipgloss.Style {
	return BorderInactive
}

func ConditionLipglossColor(name string) color.Color {
	return corestyles.ConditionLipglossColor(name)
}

// RemoteConfigValueStyle returns the shared type-aware display style used by
// human-readable CLI values.
func RemoteConfigValueStyle(value, valueType string) lipgloss.Style {
	if strings.HasPrefix(value, "(empty ") && strings.HasSuffix(value, ")") {
		return corestyles.EmptyValueStyle()
	}
	return corestyles.ValueTextStyle(value, valueType)
}

func LogLevelLipglossColor(level charmlog.Level) color.Color {
	return corestyles.LogLevelLipglossColor(level)
}
