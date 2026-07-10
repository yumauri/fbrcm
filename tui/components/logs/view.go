package logs

import (
	"strings"

	"charm.land/lipgloss/v2"
	charmlog "charm.land/log/v2"

	corelog "github.com/yumauri/fbrcm/core/log"
	"github.com/yumauri/fbrcm/tui/components/viewutil"
	"github.com/yumauri/fbrcm/tui/styles"
)

const (
	panelTitleKey   = "⁰"
	panelTitleLabel = "Logs"
)

func (m Model) View(active bool) string {
	body := strings.Split(m.viewport.View(), "\n")

	return renderLogsPanel(body, m.width, m.height, active, m.level, m.follow, m.statusFlashOn)
}

func renderLogsPanel(body []string, width, height int, active bool, currentLevel charmlog.Level, follow, flashStatus bool) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	borderStyle := styles.BorderStyle(active)
	top := renderTopBorder(width, borderStyle, active, currentLevel, follow, flashStatus)
	if height == 1 {
		return top
	}

	contentHeight := max(height-2, 0)

	lines := []string{top}
	for i := range contentHeight {
		line := ""
		if i < len(body) {
			line = body[i]
		}
		padding := max(width-lipgloss.Width(line), 0)
		lines = append(lines, line+strings.Repeat(" ", padding))
	}

	lines = append(lines, borderStyle.Render(strings.Repeat("─", width)))
	return strings.Join(lines, "\n")
}

func renderTopBorder(width int, borderStyle lipgloss.Style, active bool, currentLevel charmlog.Level, follow, flashStatus bool) string {
	leftPrefix := borderStyle.Render(viewutil.TruncatePlain("──", width))
	title, titleWidth := styles.PanelHeaderTitle(panelTitleKey, panelTitleLabel, active, max(width-lipgloss.Width(leftPrefix), 0))
	titleSep := borderStyle.Render("──")
	modeLabel := " scroll "
	if follow {
		modeLabel = " live "
	}
	mode := styles.PanelTitle.Render(modeLabel)
	if flashStatus {
		flashStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(styles.PaletteSlateBright).
			Background(styles.PaletteError)
		if styles.NoColorEnabled() {
			flashStyle = lipgloss.NewStyle().
				Bold(true).
				Reverse(true)
		}
		mode = flashStyle.Render(modeLabel)
	}
	modeSep := borderStyle.Render("──")

	levelSegment := renderLevelSegment(borderStyle, currentLevel)
	usedWidth := lipgloss.Width(leftPrefix) + titleWidth + lipgloss.Width(titleSep) + lipgloss.Width(levelSegment) + lipgloss.Width(mode) + lipgloss.Width(modeSep)
	if usedWidth >= width {
		top := leftPrefix + title + titleSep + levelSegment + mode + modeSep
		return truncateANSI(top, width)
	}

	fillWidth := width - usedWidth

	return leftPrefix +
		title +
		titleSep +
		levelSegment +
		borderStyle.Render(strings.Repeat("─", fillWidth)) +
		mode +
		modeSep
}

func renderLevelSegment(borderStyle lipgloss.Style, currentLevel charmlog.Level) string {
	levels := corelog.AvailableLevels()
	var b strings.Builder
	for i, level := range levels {
		if i > 0 {
			b.WriteString(borderStyle.Render("─"))
		}
		label := viewutil.TruncatePlain(levelLabel(level), 4)
		style := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(corelog.LevelColor(level))).
			Width(4)
		if level == currentLevel {
			style = selectedLevelStyle(style, level)
		}
		b.WriteString(style.Render(label))
	}
	return b.String()
}

func selectedLevelStyle(base lipgloss.Style, level charmlog.Level) lipgloss.Style {
	if styles.NoColorEnabled() {
		return base.Reverse(true)
	}

	return base.
		Background(styles.LogLevelLipglossColor(level)).
		Foreground(styles.PaletteSlateBright)
}

func levelLabel(level charmlog.Level) string {
	if level == corelog.SilentLevel {
		return "SLNT"
	}
	return strings.ToUpper(level.String())
}

func truncateANSI(value string, width int) string {
	if width <= 0 {
		return ""
	}
	return lipgloss.NewStyle().MaxWidth(width).Render(value)
}
