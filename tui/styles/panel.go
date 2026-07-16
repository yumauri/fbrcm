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

	PanelBorderInactive = lipgloss.NewStyle().
				Foreground(PaletteSlateDark)

	PanelBorderActive = lipgloss.NewStyle().
				Foreground(PaletteBlueBright)

	PanelTitle = lipgloss.NewStyle().
			Foreground(PaletteSlateBright)

	PanelTitleInactiveTab = lipgloss.NewStyle().
				Foreground(PaletteSlateDark)

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

	TreeProjectName = PanelText.Bold(true).Foreground(PaletteError)
	TreeProjectID   = PanelMuted

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

func PanelHeaderTitle(key, title string, active bool, maxWidth int) (string, int) {
	text := truncateHeaderText(" "+key+title+" ", maxWidth)
	width := lipgloss.Width(text)
	if text == "" {
		return "", width
	}
	if active {
		return TitleStyle(true).Render(text), width
	}
	if !strings.Contains(text, key) {
		return PanelTitle.Render(text), width
	}

	before, after, _ := strings.Cut(text, key)
	return PanelTitle.Render(before) + FilterText.Render(key) + PanelTitle.Render(after), width
}

// PanelHeaderTab renders a tab title independently from panel focus. The
// selected tab keeps the normal title style while an unselected sibling is
// muted; shortcut hints retain their yellow filter-key style.
func PanelHeaderTab(key, title string, selected, focused bool, maxWidth int) (string, int) {
	text := truncateHeaderText(" "+key+title+" ", maxWidth)
	width := lipgloss.Width(text)
	if text == "" {
		return "", width
	}
	if selected && focused {
		return TitleStyle(true).Render(text), width
	}
	if !strings.Contains(text, key) {
		if selected {
			return PanelTitle.Render(text), width
		}
		return PanelTitleInactiveTab.Render(text), width
	}
	before, after, _ := strings.Cut(text, key)
	labelStyle := PanelTitleInactiveTab
	if selected {
		labelStyle = PanelTitle
	}
	return labelStyle.Render(before) + FilterText.Render(key) + labelStyle.Render(after), width
}

func truncateHeaderText(text string, width int) string {
	if width <= 0 {
		return ""
	}

	var b strings.Builder
	used := 0
	for _, r := range text {
		rw := lipgloss.Width(string(r))
		if used+rw > width {
			break
		}
		b.WriteRune(r)
		used += rw
	}
	return b.String()
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

// TreeProjectSelectionStyle returns the shared full-row project selection style.
func TreeProjectSelectionStyle() lipgloss.Style {
	if NoColorEnabled() {
		return lipgloss.NewStyle().Bold(true).Reverse(true)
	}
	return lipgloss.NewStyle().Background(PaletteError).Foreground(PaletteSlateBright)
}

// TreeItemSelectionStyle returns the shared full-row item selection style.
func TreeItemSelectionStyle() lipgloss.Style {
	if NoColorEnabled() {
		return lipgloss.NewStyle().Reverse(true)
	}
	return lipgloss.NewStyle().Background(PaletteBlueDeep).Foreground(PaletteSlateBright)
}

// RenderDraftBadge renders the shared project-row draft marker.
func RenderDraftBadge(label string, selected bool) string {
	if selected {
		return lipgloss.NewStyle().Foreground(PaletteError).Render(label)
	}
	return lipgloss.NewStyle().
		Background(PaletteError).
		Foreground(PaletteSlateBright).
		Padding(0, 1).
		Render(label)
}

// FillSelectedLine clips and fills a selected tree row to the available width.
func FillSelectedLine(line string, width int, fillStyle lipgloss.Style) string {
	clipped := lipgloss.NewStyle().MaxWidth(width).Render(line)
	padding := max(width-lipgloss.Width(clipped), 0)
	if padding == 0 {
		return clipped
	}
	return clipped + fillStyle.Render(strings.Repeat(" ", padding))
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
