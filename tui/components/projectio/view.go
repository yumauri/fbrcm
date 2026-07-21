package projectio

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	rcdiff "github.com/yumauri/fbrcm/core/rc/diff"
	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
	"github.com/yumauri/fbrcm/tui/components/viewutil"
	"github.com/yumauri/fbrcm/tui/styles"
)

var (
	ioBorder = lipgloss.NewStyle().Foreground(styles.PaletteBlueBright)
	ioError  = lipgloss.NewStyle().Foreground(styles.PaletteError)
)

func (m Model) View() string {
	if !m.IsOpen() || m.width <= 0 || m.height <= 0 {
		return ""
	}
	inner := m.contentWidth()
	title := "Project Remote Config"
	var body []string
	switch m.phase {
	case phaseImportFile:
		title, body = "Import Remote Config · Choose JSON", m.importFileLines(inner)
	case phaseImportOptions:
		title, body = "Import Remote Config · Options", m.importOptionLines(inner)
	case phaseImportConflicts:
		title, body = "Import Remote Config · Resolve Conflicts", m.importConflictLines(inner)
	case phaseImportWorking:
		title, body = "Import Remote Config · Preparing", []string{viewutil.ProjectLine(m.project), "", "Preparing import preview…", "", helpLine("esc", "cancel")}
	case phaseExportSource:
		title, body = "Export Remote Config · Source", m.exportSourceLines()
	case phaseExportPath:
		title, body = "Export Remote Config · Destination", m.exportPathLines()
	case phaseDefaultsFormat:
		title, body = "Download Application Defaults · Format", m.defaultsFormatLines()
	case phaseDefaultsPath:
		title, body = "Download Application Defaults · Destination", m.defaultsPathLines()
	}
	return renderCard(title, body, inner)
}

func (m Model) importFileLines(width int) []string {
	return m.appendActionButtons(m.importFileBaseLines(width), width)
}

func (m Model) importFileBaseLines(width int) []string {
	lines := []string{
		"Choose a Remote Config JSON file.",
		"",
		viewutil.ProjectLine(m.project),
		"",
		viewutil.IndentLines(strings.TrimRight(m.picker.View(), "\n"), 1),
	}
	if m.errorText != "" {
		lines = append(lines, "", ioError.Render(m.errorText))
	}
	return lines
}

func (m Model) importOptionLines(width int) []string {
	return m.appendActionButtons(m.importOptionBaseLines(width), width)
}

func (m Model) importOptionBaseLines(width int) []string {
	typeName := "raw Remote Config"
	if m.summary.WrappedCache {
		typeName = "fbrcm cache"
	}
	lines := []string{
		viewutil.ProjectLine(m.project),
		styles.PanelMuted.Render("File: ") + styles.PanelText.Render(m.sourcePath),
		styles.PanelMuted.Render("Source: ") + styles.PanelText.Render(strings.Join([]string{
			typeName,
			rcdisplay.FormatCount(m.summary.Parameters(), "parameter", "parameters"),
			rcdisplay.FormatCount(m.summary.Groups, "group", "groups"),
			rcdisplay.FormatCount(m.summary.Conditions, "condition", "conditions"),
		}, " · ")),
		"",
		styles.PanelMuted.Render("Comma-separate groups and filters; leave optional fields blank for the full file."),
		"",
	}
	strategy := "Merge (keep current conflicts)"
	if m.importOptions().Strategy == core.ProjectImportReplace {
		strategy = "Replace entire config"
	}
	conditions := m.conditionPolicyLabel(m.conditionPolicy)
	rows := []string{
		"Strategy       " + strategy,
		"Groups         " + m.optionInputs[0].View(),
		"Filters        " + m.optionInputs[1].View(),
		"Search         " + m.optionInputs[2].View(),
		"Expression     " + m.optionInputs[3].View(),
		"Conditions     " + conditions,
	}
	for index, row := range rows {
		lines = append(lines, viewutil.SelectorLine(row, index == m.optionCursor))
	}
	if m.errorText != "" {
		lines = append(lines, "", ioError.Render(m.errorText))
	}
	return lines
}

func (m Model) conditionPolicyLabel(policy core.ProjectConditionPolicy) string {
	switch policy {
	case core.ProjectImportKeepPortableConditions:
		return fmt.Sprintf(
			"Keep portable conditions only (%d kept · %d removed)",
			m.summary.PortableConditions(),
			m.summary.NonPortableConditions,
		)
	case core.ProjectImportRemoveAllConditions:
		return fmt.Sprintf("Remove all conditions (%d removed)", m.summary.Conditions)
	default:
		return fmt.Sprintf("Keep all conditions (%d kept)", m.summary.Conditions)
	}
}

func (m Model) importConflictLines(width int) []string {
	lines := []string{viewutil.ProjectLine(m.project), "", rcdisplay.FormatCount(len(m.conflicts), "merge conflict", "merge conflicts") + ". Choose the value for each item.", ""}
	rows := min(len(m.conflicts), max(m.height-15, 4))
	start := max(min(m.conflictCursor-rows+1, len(m.conflicts)-rows), 0)
	for index := start; index < min(start+rows, len(m.conflicts)); index++ {
		conflict := m.conflicts[index]
		choice := "keep current"
		if m.resolutions[conflict.ID] == core.ProjectImportUseImported {
			choice = "use imported"
		}
		line := conflict.Label + "  ·  " + choice
		lines = append(lines, viewutil.SelectorLine(ansi.Truncate(line, width-2, "…"), index == m.conflictCursor))
	}
	selected := m.conflicts[m.conflictCursor]
	lines = append(lines, "",
		styles.PanelMuted.Render("Current: ")+ansi.Truncate(rcdiff.RenderConflictChoiceValue(selected.Current), width-9, "…"),
		styles.PanelMuted.Render("Import:  ")+ansi.Truncate(rcdiff.RenderConflictChoiceValue(selected.Import), width-9, "…"),
	)
	lines = append(lines, "", helpLine("←", "current", "→", "import", "C/I", "all", "enter", "review", "esc", "cancel"))
	return lines
}

func (m Model) exportSourceLines() []string {
	return m.appendActionButtons(m.exportSourceBaseLines(), m.contentWidth())
}

func (m Model) exportSourceBaseLines() []string {
	rows := []string{"Published Remote Config", "Local draft"}
	lines := []string{viewutil.ProjectLine(m.project), "", "Choose which Remote Config to export:", ""}
	for index, row := range rows {
		selected := index == 1 && m.exportDraft || index == 0 && !m.exportDraft
		lines = append(lines, viewutil.SelectorLine(row, selected))
	}
	return lines
}

func (m Model) exportPathLines() []string {
	return m.appendActionButtons(m.exportPathBaseLines(), m.contentWidth())
}

func (m Model) exportPathBaseLines() []string {
	source := "published Remote Config"
	if m.exportDraft {
		source = "local draft"
	}
	lines := []string{viewutil.ProjectLine(m.project), styles.PanelMuted.Render("Source: ") + styles.PanelText.Render(source), "", "Destination:", "  " + m.pathInput.View()}
	if m.errorText != "" {
		lines = append(lines, "", ioError.Render(m.errorText))
	}
	return lines
}

func (m Model) defaultsFormatLines() []string {
	return m.appendActionButtons(m.defaultsFormatBaseLines(), m.contentWidth())
}

func (m Model) defaultsFormatBaseLines() []string {
	labels := map[firebase.DefaultsFormat]string{
		firebase.DefaultsFormatJSON:  "JSON · Web",
		firebase.DefaultsFormatXML:   "XML · Android",
		firebase.DefaultsFormatPlist: "plist · Apple",
	}
	lines := []string{viewutil.ProjectLine(m.project), "", "Choose the application defaults format:", ""}
	for _, format := range defaultsFormats() {
		lines = append(lines, viewutil.SelectorLine(labels[format], format == m.defaultsFormat))
	}
	return lines
}

func (m Model) defaultsPathLines() []string {
	return m.appendActionButtons(m.defaultsPathBaseLines(), m.contentWidth())
}

func (m Model) defaultsPathBaseLines() []string {
	lines := []string{
		viewutil.ProjectLine(m.project),
		styles.PanelMuted.Render("Format: ") + styles.PanelText.Render(strings.ToLower(string(m.defaultsFormat))),
		"",
		"Destination:",
		"  " + m.pathInput.View(),
	}
	if m.errorText != "" {
		lines = append(lines, "", ioError.Render(m.errorText))
	}
	return lines
}

func renderCard(title string, body []string, inner int) string {
	frameInner := viewutil.PopupInnerWidth(inner)
	renderedTitle, titleWidth := styles.PanelHeaderTitle("", title, true, max(frameInner-1, 0))
	lines := []string{ioBorder.Render("╭─") + renderedTitle + ioBorder.Render(strings.Repeat("─", max(frameInner-titleWidth-1, 0))+"╮")}
	for range viewutil.PopupPaddingTop {
		lines = append(lines, ioBorder.Render("│")+viewutil.PopupContentLine("", inner)+ioBorder.Render("│"))
	}
	for _, raw := range body {
		for line := range strings.SplitSeq(raw, "\n") {
			lines = append(lines, ioBorder.Render("│")+viewutil.PopupContentLine(line, inner)+ioBorder.Render("│"))
		}
	}
	lines = append(lines, ioBorder.Render("╰"+strings.Repeat("─", frameInner)+"╯"))
	return strings.Join(lines, "\n")
}

func (m Model) appendActionButtons(lines []string, inner int) []string {
	buttons := m.actionButtons().View()
	if buttons == "" {
		return lines
	}
	leftPadding := strings.Repeat(" ", max(inner-viewWidth(buttons), 0))
	rendered := strings.Split(buttons, "\n")
	for i := range rendered {
		rendered[i] = leftPadding + rendered[i]
	}
	return append(lines, "", strings.Join(rendered, "\n"))
}

func (m Model) actionButtonIndexAt(x, y int) (int, bool) {
	buttons := m.actionButtons()
	if buttons.View() == "" {
		return -1, false
	}
	buttonX, buttonY := m.actionButtonOrigin()
	return buttons.IndexAt(x-buttonX, y-buttonY)
}

func (m Model) actionButtonOrigin() (int, int) {
	inner := m.contentWidth()
	var base []string
	switch m.phase {
	case phaseImportFile:
		base = m.importFileBaseLines(inner)
	case phaseImportOptions:
		base = m.importOptionBaseLines(inner)
	case phaseExportSource:
		base = m.exportSourceBaseLines()
	case phaseExportPath:
		base = m.exportPathBaseLines()
	case phaseDefaultsFormat:
		base = m.defaultsFormatBaseLines()
	case phaseDefaultsPath:
		base = m.defaultsPathBaseLines()
	default:
		return 0, 0
	}
	cardX, cardY := m.Position()
	buttons := m.actionButtons().View()
	return cardX + 1 + viewutil.PopupPaddingLeft + max(inner-viewWidth(buttons), 0), cardY + 2 + viewutil.PopupPaddingTop + physicalLines(base)
}

func (m Model) optionSelectorAnchor(row int) (int, int) {
	cardX, cardY := m.Position()
	const optionRowsStart = 6
	rowY := cardY + 1 + viewutil.PopupPaddingTop + optionRowsStart + row
	return cardX + 1 + viewutil.PopupPaddingLeft, rowY - 1
}

func (m Model) contentWidth() int {
	maximum := min(max(m.width-14, 1), 88)
	if m.phase != phaseImportFile {
		return min(max(maximum, 54), 88)
	}
	title := "Import Remote Config · Choose JSON"
	natural := lipgloss.Width(" " + title + " ")
	for _, raw := range m.importFileBaseLines(maximum) {
		for line := range strings.SplitSeq(raw, "\n") {
			natural = max(natural, viewWidth(line))
		}
	}
	natural = max(natural, viewWidth(m.actionButtons().View()))
	return min(max(natural, 1), maximum)
}

func physicalLines(lines []string) int {
	height := 0
	for _, line := range lines {
		height += len(strings.Split(line, "\n"))
	}
	return height
}

func helpLine(items ...string) string {
	parts := make([]string, 0, len(items)/2)
	for i := 0; i+1 < len(items); i += 2 {
		parts = append(parts, styles.FilterText.Render(items[i])+styles.PanelMuted.Render(" "+items[i+1]))
	}
	return strings.Join(parts, styles.PanelMuted.Render("  •  "))
}

func viewWidth(value string) int  { return lipgloss.Width(value) }
func viewHeight(value string) int { return lipgloss.Height(value) }
