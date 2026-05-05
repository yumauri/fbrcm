package logs

import (
	"strings"

	"charm.land/lipgloss/v2"
	charmlog "github.com/charmbracelet/log"

	corelog "fbrcm/core/log"
	"fbrcm/tui/styles"
)

const panelTitle = "[0] Logs"

func (m Model) View(active bool) string {
	body := strings.Split(m.viewport.View(), "\n")

	return renderLogsPanel(body, m.width, m.height, active, m.level, m.follow)
}

func renderLogsPanel(body []string, width, height int, active bool, currentLevel charmlog.Level, follow bool) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	borderStyle := styles.BorderStyle(active)
	contentHeight := max(height-2, 0)
	top := renderTopBorder(width, borderStyle, styles.TitleStyle(active), currentLevel, follow)

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

func renderTopBorder(width int, borderStyle, titleStyle lipgloss.Style, currentLevel charmlog.Level, follow bool) string {
	leftPrefix := borderStyle.Render(truncatePlain("──", width))
	title := titleStyle.Render(truncatePlain(" "+panelTitle+" ", max(width-lipgloss.Width(leftPrefix), 0)))
	titleSep := borderStyle.Render("──")
	modeLabel := " scroll "
	if follow {
		modeLabel = " live "
	}
	mode := styles.PanelTitle.Render(modeLabel)
	modeSep := borderStyle.Render("──")

	levelSegment := renderLevelSegment(borderStyle, currentLevel)
	usedWidth := lipgloss.Width(leftPrefix) + lipgloss.Width(title) + lipgloss.Width(titleSep) + lipgloss.Width(levelSegment) + lipgloss.Width(mode) + lipgloss.Width(modeSep)
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
		label := truncatePlain(levelLabel(level), 4)
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

func truncatePlain(value string, width int) string {
	if width <= 0 {
		return ""
	}

	runes := []rune(value)
	if len(runes) <= width {
		return value
	}

	return string(runes[:width])
}

func truncateANSI(value string, width int) string {
	if width <= 0 {
		return ""
	}
	return lipgloss.NewStyle().MaxWidth(width).Render(value)
}
