package styles

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	charmlog "charm.land/log/v2"

	"github.com/yumauri/fbrcm/core"
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
	PaletteSuccess     = corestyles.ColorSuccess
	FirebaseAndroid    = lipgloss.Color(corestyles.FirebaseAndroid)
	FirebaseIOS        = lipgloss.Color(corestyles.FirebaseIOS)
	FirebaseWeb        = lipgloss.Color(corestyles.FirebaseWeb)

	ColorRowStripe         = corestyles.ColorRowStripe
	ColorInactiveSelection = corestyles.ColorInactiveSelection
	ColorAdded             = corestyles.ColorAdded
	ColorRemoved           = corestyles.ColorRemoved
	ColorChanged           = corestyles.ColorChanged
	ColorNote              = corestyles.ColorNote

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

// FirebaseAppPlatformColor returns Firebase's display color for a platform.
func FirebaseAppPlatformColor(platform core.AppPlatform) color.Color {
	switch platform {
	case core.AppPlatformAndroid:
		return FirebaseAndroid
	case core.AppPlatformIOS:
		return FirebaseIOS
	case core.AppPlatformWeb:
		return FirebaseWeb
	default:
		return PaletteSlateDim
	}
}

// RenderFirebaseAppPlatforms applies base to value and replaces the foreground
// of every platform segment in an embedded Firebase App ID with that platform's
// Firebase color. It deliberately leaves machine-readable and NO_COLOR output
// untouched.
func RenderFirebaseAppPlatforms(value string, base lipgloss.Style) string {
	if NoColorEnabled() {
		return value
	}
	spans := core.FirebaseAppIDPlatformSpans(value)
	if len(spans) == 0 {
		return base.Render(value)
	}
	var b strings.Builder
	start := 0
	for _, span := range spans {
		b.WriteString(base.Render(value[start:span.Start]))
		b.WriteString(base.Foreground(FirebaseAppPlatformColor(span.Platform)).Render(value[span.Start:span.End]))
		start = span.End
	}
	b.WriteString(base.Render(value[start:]))
	return b.String()
}

// RemoteConfigValueStyle returns the shared type-aware display style used by
// human-readable CLI values.
func RemoteConfigValueStyle(value, valueType string) lipgloss.Style {
	if value == "(in-app default)" || strings.HasPrefix(value, "(empty ") && strings.HasSuffix(value, ")") {
		return corestyles.EmptyValueStyle()
	}
	return corestyles.ValueTextStyle(value, valueType)
}

func LogLevelLipglossColor(level charmlog.Level) color.Color {
	return corestyles.LogLevelLipglossColor(level)
}
