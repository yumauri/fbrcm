package styles

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	charmlog "github.com/charmbracelet/log"

	"fbrcm/core/env"
)

const (
	URLColor = "117"

	PaletteBlueBright  = "#8FA8C7"
	PaletteBlueDeep    = "#556B84"
	PaletteOrange      = "#C8A27E"
	PaletteYellow      = "#D8C6A0"
	PaletteGold        = "#BFA77A"
	PaletteSlateBright = "#D7D9DE"
	PaletteSlate       = "#BEC3CC"
	PaletteSlateDim    = "#959CA8"
	PaletteSlateDark   = "#5A6270"
	PaletteError       = "#C58A8A"

	PaletteRowStripe = "#121417"

	PaletteConditionCyan       = "#61D6E8"
	PaletteConditionDeepOrange = "#FF8A5B"
	PaletteConditionGreen      = "#7FD38B"
	PaletteConditionIndigo     = "#8AA2FF"
	PaletteConditionLime       = "#C1D96F"
	PaletteConditionPink       = "#F38DB5"
	PaletteConditionPurple     = "#B58CFF"
	PaletteConditionTeal       = "#58D1C9"

	PaletteAdded   = "42"
	PaletteRemoved = "203"
	PaletteChanged = "221"
	PaletteNote    = "245"

	DebugLevelColor   = "63"
	InfoLevelColor    = "86"
	WarnLevelColor    = "192"
	ErrorLevelColor   = "204"
	FatalLevelColor   = "134"
	DefaultLevelColor = "255"
)

var (
	ColorURL = lipgloss.Color(URLColor)

	ColorBlueBright  = lipgloss.Color(PaletteBlueBright)
	ColorBlueDeep    = lipgloss.Color(PaletteBlueDeep)
	ColorOrange      = lipgloss.Color(PaletteOrange)
	ColorYellow      = lipgloss.Color(PaletteYellow)
	ColorGold        = lipgloss.Color(PaletteGold)
	ColorSlateBright = lipgloss.Color(PaletteSlateBright)
	ColorSlate       = lipgloss.Color(PaletteSlate)
	ColorSlateDim    = lipgloss.Color(PaletteSlateDim)
	ColorSlateDark   = lipgloss.Color(PaletteSlateDark)
	ColorError       = lipgloss.Color(PaletteError)

	ColorRowStripe = lipgloss.Color(PaletteRowStripe)

	ColorConditionCyan       = lipgloss.Color(PaletteConditionCyan)
	ColorConditionDeepOrange = lipgloss.Color(PaletteConditionDeepOrange)
	ColorConditionGreen      = lipgloss.Color(PaletteConditionGreen)
	ColorConditionIndigo     = lipgloss.Color(PaletteConditionIndigo)
	ColorConditionLime       = lipgloss.Color(PaletteConditionLime)
	ColorConditionPink       = lipgloss.Color(PaletteConditionPink)
	ColorConditionPurple     = lipgloss.Color(PaletteConditionPurple)
	ColorConditionTeal       = lipgloss.Color(PaletteConditionTeal)

	ColorAdded   = lipgloss.Color(PaletteAdded)
	ColorRemoved = lipgloss.Color(PaletteRemoved)
	ColorChanged = lipgloss.Color(PaletteChanged)
	ColorNote    = lipgloss.Color(PaletteNote)
)

func NoColorEnabled() bool {
	return env.NoColorEnabled()
}

func LogLevelColor(level charmlog.Level) string {
	switch level {
	case charmlog.DebugLevel:
		return DebugLevelColor
	case charmlog.InfoLevel:
		return InfoLevelColor
	case charmlog.WarnLevel:
		return WarnLevelColor
	case charmlog.ErrorLevel:
		return ErrorLevelColor
	case charmlog.FatalLevel:
		return FatalLevelColor
	default:
		return DefaultLevelColor
	}
}

func ConditionColor(name string) string {
	switch strings.ToUpper(strings.TrimSpace(name)) {
	case "BLUE":
		return PaletteBlueBright
	case "BROWN", "ORANGE":
		return PaletteOrange
	case "CYAN":
		return PaletteConditionCyan
	case "DEEP_ORANGE":
		return PaletteConditionDeepOrange
	case "GREEN":
		return PaletteConditionGreen
	case "INDIGO":
		return PaletteConditionIndigo
	case "LIME":
		return PaletteConditionLime
	case "PINK":
		return PaletteConditionPink
	case "PURPLE":
		return PaletteConditionPurple
	case "TEAL":
		return PaletteConditionTeal
	default:
		return PaletteBlueBright
	}
}

func LogLevelLipglossColor(level charmlog.Level) color.Color {
	return lipgloss.Color(LogLevelColor(level))
}

func ConditionLipglossColor(name string) color.Color {
	switch strings.ToUpper(strings.TrimSpace(name)) {
	case "BLUE":
		return ColorBlueBright
	case "BROWN", "ORANGE":
		return ColorOrange
	case "CYAN":
		return ColorConditionCyan
	case "DEEP_ORANGE":
		return ColorConditionDeepOrange
	case "GREEN":
		return ColorConditionGreen
	case "INDIGO":
		return ColorConditionIndigo
	case "LIME":
		return ColorConditionLime
	case "PINK":
		return ColorConditionPink
	case "PURPLE":
		return ColorConditionPurple
	case "TEAL":
		return ColorConditionTeal
	default:
		return ColorBlueBright
	}
}
