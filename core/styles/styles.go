package styles

import (
	"image/color"
	"maps"
	"slices"
	"strings"
	"sync/atomic"

	"charm.land/lipgloss/v2"
	charmlog "charm.land/log/v2"

	"github.com/yumauri/fbrcm/core/env"
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

const (
	TokenURL                      = "url"
	TokenPrimary                  = "primary"
	TokenSelection                = "selection"
	TokenSecondary                = "secondary"
	TokenHighlight                = "highlight"
	TokenHighlightMuted           = "highlight_muted"
	TokenText                     = "text"
	TokenTextSoft                 = "text_soft"
	TokenTextMuted                = "text_muted"
	TokenTextDim                  = "text_dim"
	TokenError                    = "error"
	TokenSuccess                  = "success"
	TokenRowStripe                = "row_stripe"
	TokenConditionBlue            = "condition_blue"
	TokenConditionBrown           = "condition_brown"
	TokenConditionCyan            = "condition_cyan"
	TokenConditionDeepOrange      = "condition_deep_orange"
	TokenConditionGreen           = "condition_green"
	TokenConditionIndigo          = "condition_indigo"
	TokenConditionLime            = "condition_lime"
	TokenConditionOrange          = "condition_orange"
	TokenConditionPink            = "condition_pink"
	TokenConditionPurple          = "condition_purple"
	TokenConditionTeal            = "condition_teal"
	TokenDiffAdded                = "diff_added"
	TokenDiffRemoved              = "diff_removed"
	TokenDiffChanged              = "diff_changed"
	TokenDiffNote                 = "diff_note"
	TokenHistoryAddedBackground   = "history_added_background"
	TokenHistoryRemovedBackground = "history_removed_background"
	TokenHistoryChangedBackground = "history_changed_background"
	TokenInactiveSelection        = "inactive_selection"
	TokenOfflineForeground        = "offline_foreground"
	TokenOfflineBackground        = "offline_background"
	TokenLogDebug                 = "log_debug"
	TokenLogInfo                  = "log_info"
	TokenLogWarn                  = "log_warn"
	TokenLogError                 = "log_error"
	TokenLogFatal                 = "log_fatal"
	TokenLogSilent                = "log_silent"
	TokenLogDefault               = "log_default"
	TokenLogoStart                = "logo_start"
	TokenLogoMiddle               = "logo_middle"
	TokenLogoEnd                  = "logo_end"
)

// Palette is a complete set of semantic application colors. Palette values
// are immutable after publication through ApplyPalette.
type Palette map[string]string

var builtInPalette = Palette{
	TokenURL: URLColor,

	TokenPrimary:        PaletteBlueBright,
	TokenSelection:      PaletteBlueDeep,
	TokenSecondary:      PaletteOrange,
	TokenHighlight:      PaletteYellow,
	TokenHighlightMuted: PaletteGold,
	TokenText:           PaletteSlateBright,
	TokenTextSoft:       PaletteSlate,
	TokenTextMuted:      PaletteSlateDim,
	TokenTextDim:        PaletteSlateDark,
	TokenError:          PaletteError,
	TokenSuccess:        PaletteConditionGreen,
	TokenRowStripe:      PaletteRowStripe,

	TokenConditionBlue:       PaletteBlueBright,
	TokenConditionBrown:      PaletteOrange,
	TokenConditionCyan:       PaletteConditionCyan,
	TokenConditionDeepOrange: PaletteConditionDeepOrange,
	TokenConditionGreen:      PaletteConditionGreen,
	TokenConditionIndigo:     PaletteConditionIndigo,
	TokenConditionLime:       PaletteConditionLime,
	TokenConditionOrange:     PaletteOrange,
	TokenConditionPink:       PaletteConditionPink,
	TokenConditionPurple:     PaletteConditionPurple,
	TokenConditionTeal:       PaletteConditionTeal,

	TokenDiffAdded:   PaletteAdded,
	TokenDiffRemoved: PaletteRemoved,
	TokenDiffChanged: PaletteChanged,
	TokenDiffNote:    PaletteNote,

	TokenHistoryAddedBackground:   "#315A46",
	TokenHistoryRemovedBackground: "#68434A",
	TokenHistoryChangedBackground: "#665A38",
	TokenInactiveSelection:        "#343A43",
	TokenOfflineForeground:        "15",
	TokenOfflineBackground:        "196",

	TokenLogDebug:   DebugLevelColor,
	TokenLogInfo:    InfoLevelColor,
	TokenLogWarn:    WarnLevelColor,
	TokenLogError:   ErrorLevelColor,
	TokenLogFatal:   FatalLevelColor,
	TokenLogSilent:  PaletteNote,
	TokenLogDefault: DefaultLevelColor,

	TokenLogoStart:  "#FFC400",
	TokenLogoMiddle: "#FF9100",
	TokenLogoEnd:    "#DD2C00",
}

var currentPalette atomic.Value

func init() {
	currentPalette.Store(DefaultPalette())
}

// DefaultPalette returns a detached copy of the built-in palette.
func DefaultPalette() Palette {
	return maps.Clone(builtInPalette)
}

// CurrentPalette returns a detached copy of the active palette.
func CurrentPalette() Palette {
	return maps.Clone(currentPalette.Load().(Palette))
}

// ApplyPalette atomically publishes a palette. Missing values retain their
// built-in defaults so callers can safely apply a partial palette.
func ApplyPalette(palette Palette) {
	complete := DefaultPalette()
	maps.Copy(complete, palette)
	currentPalette.Store(complete)
}

func ResetPalette() {
	currentPalette.Store(DefaultPalette())
}

func SupportedTokens() []string {
	tokens := slices.Collect(maps.Keys(builtInPalette))
	slices.Sort(tokens)
	return tokens
}

// PreviewTokens returns the representative semantic colors used by compact
// theme previews in both human interfaces.
func PreviewTokens() []string {
	return []string{
		TokenPrimary,
		TokenSelection,
		TokenSecondary,
		TokenHighlight,
		TokenText,
		TokenTextMuted,
		TokenSuccess,
		TokenError,
	}
}

func ColorValue(token string) string {
	return currentPalette.Load().(Palette)[token]
}

type paletteColor struct{ token string }

func newPaletteColor(token string) color.Color { return &paletteColor{token: token} }

func (c *paletteColor) RGBA() (r, g, b, a uint32) {
	return lipgloss.Color(ColorValue(c.token)).RGBA()
}

func (c *paletteColor) String() string { return ColorValue(c.token) }

var (
	ColorURL = newPaletteColor(TokenURL)

	ColorBlueBright  = newPaletteColor(TokenPrimary)
	ColorBlueDeep    = newPaletteColor(TokenSelection)
	ColorOrange      = newPaletteColor(TokenSecondary)
	ColorYellow      = newPaletteColor(TokenHighlight)
	ColorGold        = newPaletteColor(TokenHighlightMuted)
	ColorSlateBright = newPaletteColor(TokenText)
	ColorSlate       = newPaletteColor(TokenTextSoft)
	ColorSlateDim    = newPaletteColor(TokenTextMuted)
	ColorSlateDark   = newPaletteColor(TokenTextDim)
	ColorError       = newPaletteColor(TokenError)
	ColorSuccess     = newPaletteColor(TokenSuccess)

	ColorRowStripe = newPaletteColor(TokenRowStripe)

	ColorConditionBlue       = newPaletteColor(TokenConditionBlue)
	ColorConditionBrown      = newPaletteColor(TokenConditionBrown)
	ColorConditionCyan       = newPaletteColor(TokenConditionCyan)
	ColorConditionDeepOrange = newPaletteColor(TokenConditionDeepOrange)
	ColorConditionGreen      = newPaletteColor(TokenConditionGreen)
	ColorConditionIndigo     = newPaletteColor(TokenConditionIndigo)
	ColorConditionLime       = newPaletteColor(TokenConditionLime)
	ColorConditionOrange     = newPaletteColor(TokenConditionOrange)
	ColorConditionPink       = newPaletteColor(TokenConditionPink)
	ColorConditionPurple     = newPaletteColor(TokenConditionPurple)
	ColorConditionTeal       = newPaletteColor(TokenConditionTeal)

	ColorAdded   = newPaletteColor(TokenDiffAdded)
	ColorRemoved = newPaletteColor(TokenDiffRemoved)
	ColorChanged = newPaletteColor(TokenDiffChanged)
	ColorNote    = newPaletteColor(TokenDiffNote)

	ColorHistoryAddedBackground   = newPaletteColor(TokenHistoryAddedBackground)
	ColorHistoryRemovedBackground = newPaletteColor(TokenHistoryRemovedBackground)
	ColorHistoryChangedBackground = newPaletteColor(TokenHistoryChangedBackground)
	ColorInactiveSelection        = newPaletteColor(TokenInactiveSelection)
	ColorOfflineForeground        = newPaletteColor(TokenOfflineForeground)
	ColorOfflineBackground        = newPaletteColor(TokenOfflineBackground)
	ColorLogoStart                = newPaletteColor(TokenLogoStart)
	ColorLogoMiddle               = newPaletteColor(TokenLogoMiddle)
	ColorLogoEnd                  = newPaletteColor(TokenLogoEnd)
)

func NoColorEnabled() bool {
	return env.NoColorEnabled()
}

func LogLevelColor(level charmlog.Level) string {
	switch level {
	case charmlog.DebugLevel:
		return ColorValue(TokenLogDebug)
	case charmlog.InfoLevel:
		return ColorValue(TokenLogInfo)
	case charmlog.WarnLevel:
		return ColorValue(TokenLogWarn)
	case charmlog.ErrorLevel:
		return ColorValue(TokenLogError)
	case charmlog.FatalLevel:
		return ColorValue(TokenLogFatal)
	default:
		return ColorValue(TokenLogDefault)
	}
}

func SilentLevelColor() string { return ColorValue(TokenLogSilent) }

func LogLevelLipglossColor(level charmlog.Level) color.Color {
	return lipgloss.Color(LogLevelColor(level))
}

func ConditionLipglossColor(name string) color.Color {
	return conditionPaletteColor(name)
}

func conditionPaletteColor(name string) color.Color {
	switch strings.ToUpper(strings.TrimSpace(name)) {
	case "BLUE":
		return ColorConditionBlue
	case "BROWN":
		return ColorConditionBrown
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
	case "ORANGE":
		return ColorConditionOrange
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

func EmptyValueStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(ColorSlateDim).Italic(true)
}

func ValueTextStyle(value, valueType string) lipgloss.Style {
	switch strings.TrimSpace(strings.ToLower(valueType)) {
	case "boolean":
		if strings.EqualFold(value, "true") {
			return lipgloss.NewStyle().Foreground(ConditionLipglossColor("GREEN"))
		}
		if strings.EqualFold(value, "false") {
			return lipgloss.NewStyle().Foreground(ColorError)
		}
	case "number":
		return lipgloss.NewStyle().Foreground(ColorBlueBright)
	case "json":
		return lipgloss.NewStyle().Foreground(ConditionLipglossColor("CYAN"))
	}
	return lipgloss.NewStyle().Foreground(ColorSlateBright)
}
