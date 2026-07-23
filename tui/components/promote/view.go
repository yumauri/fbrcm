package promote

import (
	"fmt"
	"reflect"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/rivo/uniseg"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	coreparameters "github.com/yumauri/fbrcm/core/parameters"
	rcdiff "github.com/yumauri/fbrcm/core/rc/diff"
	rcpromote "github.com/yumauri/fbrcm/core/rc/promote"
	corestyles "github.com/yumauri/fbrcm/core/styles"
	"github.com/yumauri/fbrcm/tui/components/jsoninput"
	"github.com/yumauri/fbrcm/tui/components/viewutil"
	tuiconfig "github.com/yumauri/fbrcm/tui/config"
	"github.com/yumauri/fbrcm/tui/styles"
)

const (
	targetFirstOptionRow     = 1
	promoteHeaderHeight      = 4
	detailValueMaxLineBudget = 5
	detailValueMinLineBudget = 1
)

type detailRenderOptions struct {
	width           int
	valueLineBudget int
	hideUnchanged   bool
}

type parameterHiddenFields struct {
	valueType        bool
	description      bool
	group            bool
	conditionalValue map[string]bool
	defaultValue     bool
	values           bool
}

type conditionHiddenFields struct {
	position   bool
	expression bool
	color      bool
}

var (
	targetBorderStyle = lipgloss.NewStyle().Foreground(styles.PaletteBlueBright)
	targetTitleStyle  = lipgloss.NewStyle().Bold(true).Foreground(styles.PaletteGold)
)

// TargetPosition places the popup so its selected option remains on the
// source project's row as the target list moves.
func (m Model) TargetPosition() (int, int) {
	return m.x - 2, m.targetRow - targetFirstOptionRow - m.pickerCursor
}

// SourcePosition places the source frame one row above the project name. Its
// omitted left edge represents the part of the overlay clipped by the screen.
func (m Model) SourcePosition() (int, int) {
	return 0, m.targetRow - 1
}

// SourceView renders the source project as an independent overlay. It does
// not alter Projects panel content, selection state, or viewport geometry.
func (m Model) SourceView() string {
	if !m.TargetPickerOpen() || m.x <= 1 {
		return ""
	}
	width := m.x - 1
	contentWidth := max(width-1, 0)
	nameStyle, idStyle := styles.PanelText, styles.PanelMuted
	source := m.pickerSource
	if source.Disabled {
		nameStyle = nameStyle.Foreground(styles.PanelTitleInactiveTab.GetForeground())
		idStyle = idStyle.Foreground(styles.PanelTitleInactiveTab.GetForeground())
	}
	id := "  " + source.ProjectID
	if source.Disabled {
		id += " · disabled"
	}
	return strings.Join([]string{
		targetBorderStyle.Render(strings.Repeat("─", contentWidth) + "╮"),
		sourceProjectLine(" "+source.Name, contentWidth, nameStyle),
		sourceProjectLine(id, contentWidth, idStyle),
		targetBorderStyle.Render(strings.Repeat("─", contentWidth) + "╯"),
	}, "\n")
}

func sourceProjectLine(value string, width int, style lipgloss.Style) string {
	value = ansi.Truncate(value, width, "…")
	return style.Render(viewutil.PadRight(value, width)) + targetBorderStyle.Render("│")
}

func (m Model) TargetView() string {
	if !m.TargetPickerOpen() || m.width <= 0 {
		return ""
	}
	optionWidth := m.targetOptionWidth()
	titleText := " Promote to… "
	title := targetTitleStyle.Render(titleText)
	titleWidth := lipgloss.Width(titleText)
	topLeft := "╭"
	if len(m.candidates) == 0 || m.pickerCursor == 0 {
		topLeft = "─"
	}
	lines := []string{
		targetBorderStyle.Render(topLeft+"─") + title + targetBorderStyle.Render(strings.Repeat("─", max(optionWidth+1-titleWidth, 0))+"╮"),
	}
	if len(m.candidates) == 0 {
		lines = append(lines, targetOptionLine("No matching target projects", optionWidth, false, 1, false))
	} else {
		for index, project := range m.candidates {
			selected := index == m.pickerCursor
			label := targetProjectLabel(project, selected)
			lines = append(lines, targetOptionLine(label, optionWidth, selected, index-m.pickerCursor, true))
		}
	}
	lines = append(lines, m.targetFilterLine(optionWidth))
	bottomLeft := "╰"
	if len(m.candidates) == 0 || m.pickerCursor == len(m.candidates)-1 {
		bottomLeft = "─"
	}
	lines = append(lines, targetBorderStyle.Render(bottomLeft+strings.Repeat("─", optionWidth+2)+"╯"))
	return strings.Join(lines, "\n")
}

func (m Model) targetOptionWidth() int {
	width := max(lipgloss.Width(m.targetInput.Prompt+m.targetInput.Placeholder)+1, 28)
	for _, project := range m.candidates {
		label := projectName(project)
		if project.Disabled {
			label += " · cached only"
		}
		width = max(width, lipgloss.Width(label)+2)
	}
	return min(width, max(m.width-4, 1))
}

func (m Model) targetFilterLine(width int) string {
	input := m.targetInput
	input.SetWidth(max(width-lipgloss.Width(input.Prompt), 1))
	value := ansi.Truncate(input.View(), width, "")
	value = viewutil.PadRight(value, width)
	left := "│ "
	switch len(m.candidates) - m.pickerCursor {
	case 0:
		left = "  "
	case 1:
		left = "  "
	case 2:
		left = "╮ "
	}
	return targetBorderStyle.Render(left) + value + targetBorderStyle.Render(" │")
}

func targetProjectLabel(project core.Project, selected bool) string {
	nameStyle := styles.SelectionListOptionStyle(selected)
	if strings.TrimSpace(project.Name) == "" || project.Name == project.ProjectID {
		label := styles.PanelMuted.Render(project.ProjectID)
		if project.Disabled {
			label += styles.PanelMuted.Render(" · cached only")
		}
		return label
	}
	label := nameStyle.Render(project.Name) + styles.PanelMuted.Render(" ("+project.ProjectID+")")
	if project.Disabled {
		label += styles.PanelMuted.Render(" · cached only")
	}
	return label
}

func targetOptionLine(label string, width int, selected bool, relative int, styled bool) string {
	left := targetBorderStyle.Render("│ ")
	if selected {
		left = targetBorderStyle.Render("▸ ")
	} else if relative == -1 {
		left = targetBorderStyle.Render("╯ ")
	} else if relative == 1 {
		left = "  "
	} else if relative == 2 {
		left = targetBorderStyle.Render("╮ ")
	}
	label = ansi.Truncate(label, width, "…")
	content := viewutil.PadRight(label, width)
	if !styled {
		content = styles.SelectionListOptionStyle(selected).Render(content)
	}
	return left + content + targetBorderStyle.Render(" │")
}

func panelTitleKey() string {
	return tuiconfig.ActionKeyHint(tuiconfig.BlockGlobal, tuiconfig.ActionFocusPromote)
}

func (m Model) ViewWithBorder(active, borderActive bool) string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}
	border := styles.BorderStyle(borderActive)
	inner := max(m.width-2, 0)
	title, titleWidth := styles.PanelHeaderTitle(panelTitleKey(), "Promote Remote Config", active, max(inner-2, 0))
	top := border.Render("╭─") + title + border.Render(strings.Repeat("─", max(inner-titleWidth-1, 0))+"╮")
	footer := m.filter.View(max(m.width-1, 1), active, m.visibleCount())
	bodyHeight := max(m.height-2-len(footer), 0)
	body := m.bodyLines(inner, bodyHeight)
	lines := []string{top}
	for i := range bodyHeight {
		line := ""
		if i < len(body) {
			line = ansi.Truncate(body[i], inner, "")
		}
		lines = append(lines, border.Render("│")+viewutil.PadRight(line, inner)+border.Render("│"))
	}
	for i, line := range footer {
		left := "│"
		if i == 0 {
			left = "├"
		}
		lines = append(lines, border.Render(left)+line)
	}
	lines = append(lines, border.Render("╰"+strings.Repeat("─", inner)+"╯"))
	return strings.Join(lines, "\n")
}

func (m Model) visibleCount() int {
	if m.TargetPickerOpen() {
		return len(m.candidates)
	}
	return len(m.visible)
}

func (m Model) bodyLines(width, height int) []string {
	switch m.phase {
	case phaseLoading:
		return m.headerLines(width)
	case phaseReview:
		return m.reviewLines(width, height)
	default:
		return nil
	}
}

func (m Model) reviewLines(width, height int) []string {
	if m.plan == nil {
		lines := []string{"", " Promotion unavailable."}
		if m.err != nil {
			lines = append(lines, "", " "+m.err.Error())
		}
		return lines
	}
	header := m.headerLines(width)
	contentHeight := max(height-len(header), 1)
	leftWidth, rightWidth := m.promotionColumnWidths(width)
	left := m.changeLines(leftWidth, contentHeight)
	right := strings.Split(m.detail.View(), "\n")
	rows := make([]string, 0, contentHeight)
	separator := styles.BorderStyle(false).Render("│")
	for i := range contentHeight {
		l, r := "", ""
		if i < len(left) {
			l = left[i]
		}
		if i < len(right) {
			r = right[i]
		}
		l = ansi.Truncate(l, leftWidth, "…")
		r = ansi.Truncate(r, rightWidth, "")
		rows = append(rows, viewutil.PadRight(l, leftWidth)+separator+viewutil.PadRight(r, rightWidth))
	}
	return append(header, rows...)
}

func (m Model) headerLines(width int) []string {
	sourceProject := m.source
	targetProject := m.target
	sourceState := styles.PanelMuted.Render(" …")
	targetState := styles.PanelMuted.Render(" …")
	status := styles.PanelMuted.Render(" Loading Remote Config snapshots…")
	if m.plan != nil {
		sourceProject = m.plan.Source.Project
		targetProject = m.plan.Target.Project
		sourceState = styles.PanelMuted.Render(" " + snapshotState(m.plan.Source))
		targetState = styles.PanelMuted.Render(" " + snapshotState(m.plan.Target))
		status = m.summaryLine()
	}

	connectorWidth := lipgloss.Width("🬭🬿")
	sourceWidth, targetWidth := promotionHeaderBlockWidths(
		sourceProject, targetProject,
		lipgloss.Width(sourceState), lipgloss.Width(targetState),
		width-connectorWidth-1,
	)
	sourceLabel := " " + headerProjectLabel(sourceProject, max(sourceWidth-1, 0))
	targetLabel := " " + headerProjectLabel(targetProject, max(targetWidth-1, 0))
	projects := promotionHeaderLine(
		sourceLabel,
		targetLabel,
		sourceWidth, width, "🬭🬿",
	)
	states := promotionHeaderLine(
		sourceState,
		targetState,
		sourceWidth, width, "🬂🭚",
	)
	return []string{projects, states, ansi.Truncate(status, width, "…"), styles.BorderStyle(false).Render(strings.Repeat("─", width))}
}

func (m Model) summaryLine() string {
	summary := m.plan.Plan.Diff.TotalSummary()
	selected, required := 0, 0
	if m.preview != nil {
		selected, required = len(m.preview.Requested), len(m.preview.Required)
	}
	prune := "OFF"
	if m.prune {
		prune = changeStyle(rcdiff.ChangeRemoved).Render("ON")
	}
	counts := " " + changeStyle(rcdiff.ChangeAdded).Render(fmt.Sprintf("+%d to add", summary.Added)) +
		" · " + changeStyle(rcdiff.ChangeChanged).Render(fmt.Sprintf("~%d to update", summary.Changed)) +
		" · " + changeStyle(rcdiff.ChangeRemoved).Render(fmt.Sprintf("-%d target-only", summary.Removed)) +
		fmt.Sprintf(" · %d selected · %d required · prune %s", selected, required, prune)
	if m.err != nil {
		counts = " Error: " + m.err.Error()
	}
	return counts
}

func promotionHeaderLine(left, right string, sourceWidth, width int, connector string) string {
	connectorWidth := lipgloss.Width(connector)
	left = viewutil.PadRight(ansi.Truncate(left, sourceWidth, "…"), sourceWidth)
	rightWidth := max(width-sourceWidth-connectorWidth-1, 0)
	right = viewutil.PadRight(ansi.Truncate(right, rightWidth, "…"), rightWidth)
	return left + " " + styles.BorderStyle(false).Render(connector) + right
}

func headerProjectLabel(project core.Project, width int) string {
	if width <= 0 {
		return ""
	}
	nameStyle := styles.PanelText
	idStyle := styles.PanelMuted
	if project.Disabled {
		inactive := styles.PanelTitleInactiveTab.GetForeground()
		nameStyle = nameStyle.Foreground(inactive)
		idStyle = idStyle.Foreground(inactive)
	}
	name := strings.TrimSpace(project.Name)
	if name == "" || name == project.ProjectID {
		return nameStyle.Render(ansi.Truncate(project.ProjectID, width, "…"))
	}
	full := name + " (" + project.ProjectID + ")"
	if lipgloss.Width(full) <= width {
		return nameStyle.Render(name) + idStyle.Render(" ("+project.ProjectID+")")
	}
	nameWidth := lipgloss.Width(name)
	if nameWidth+2 >= width {
		return nameStyle.Render(ansi.Truncate(name, width, "…"))
	}
	idWidth := width - nameWidth - 2
	return nameStyle.Render(name) + idStyle.Render(" ("+ansi.Truncate(project.ProjectID, idWidth, "…"))
}

func promotionHeaderBlockWidths(source, target core.Project, sourceStateWidth, targetStateWidth, available int) (int, int) {
	sourceDesired := max(headerProjectLabelWidth(source)+1, sourceStateWidth)
	targetDesired := max(headerProjectLabelWidth(target)+1, targetStateWidth)
	if sourceDesired+targetDesired <= available {
		return sourceDesired, targetDesired
	}
	if available <= 0 {
		return 0, 0
	}

	sourceWidth := min(sourceDesired, headerProjectMinimumWidth(source)+1)
	targetWidth := min(targetDesired, headerProjectMinimumWidth(target)+1)
	if sourceWidth+targetWidth > available {
		sourceWidth = min(sourceDesired, available/2)
		targetWidth = min(targetDesired, available-sourceWidth)
	}
	remaining := max(available-sourceWidth-targetWidth, 0)
	for remaining > 0 && (sourceWidth < sourceDesired || targetWidth < targetDesired) {
		if sourceWidth < sourceDesired {
			sourceWidth++
			remaining--
		}
		if remaining > 0 && targetWidth < targetDesired {
			targetWidth++
			remaining--
		}
	}
	return sourceWidth, targetWidth
}

func headerProjectLabelWidth(project core.Project) int {
	name := strings.TrimSpace(project.Name)
	if name == "" || name == project.ProjectID {
		return lipgloss.Width(project.ProjectID)
	}
	return lipgloss.Width(name + " (" + project.ProjectID + ")")
}

func headerProjectMinimumWidth(project core.Project) int {
	name := strings.TrimSpace(project.Name)
	if name == "" || name == project.ProjectID {
		return min(headerProjectLabelWidth(project), 1)
	}
	return min(headerProjectLabelWidth(project), lipgloss.Width(name)+3)
}

func snapshotState(snapshot core.ProjectPromotionSnapshot) string {
	state := snapshot.Source
	if state == "draft" && snapshot.StaleDraft {
		state += " stale"
	}
	return state + " v" + displayVersion(snapshot)
}

func (m Model) promotionColumnWidths(width int) (int, int) {
	const (
		minChangesWidth = 27
		minDetailsWidth = 27
	)
	if width < minChangesWidth+minDetailsWidth+1 {
		leftWidth := max(min(width/2, width-1), 1)
		return leftWidth, max(width-leftWidth-1, 1)
	}

	leftWidth := max(m.changeColumnNaturalWidth(), minChangesWidth)
	leftWidth = min(leftWidth, width-minDetailsWidth-1)
	return leftWidth, max(width-leftWidth-1, 1)
}

func (m Model) changeColumnNaturalWidth() int {
	width := lipgloss.Width(" Changes")
	for _, item := range m.visible {
		hint := m.changeHint(item)
		conditionMarker := ""
		if item.Kind == rcdiff.ItemCondition {
			conditionMarker = "● "
		}
		row := " [ ] " + changeSymbol(item.Change) + " " + conditionMarker + changeItemLabel(item) + hint
		width = max(width, lipgloss.Width(row))
		width = max(width, lipgloss.Width(" "+itemKindLabel(item.Kind)))
	}
	return width
}

func (m Model) changeLines(width, height int) []string {
	lines := []string{styles.DetailsLabel.Render(" Changes"), ""}
	start := min(m.offset, len(m.visible))
	end := min(start+max(height-2, 0), len(m.visible))
	lastKind := rcdiff.ItemKind("")
	for index := start; index < end; index++ {
		item := m.visible[index]
		if item.Kind != lastKind {
			if lastKind != "" {
				lines = append(lines, "")
			}
			lines = append(lines, styles.DetailsLabel.Render(" "+itemKindLabel(item.Kind)))
			lastKind = item.Kind
		}
		lines = append(lines, m.renderChangeRow(item, index == m.cursor, width))
	}
	if len(m.visible) == 0 {
		lines = append(lines, " No matching changes")
	}
	for len(lines) < height {
		lines = append(lines, "")
	}
	return lines[:height]
}

func (m Model) renderChangeRow(item rcpromote.Item, selected bool, width int) string {
	required := m.preview != nil && m.preview.Effective[item.ID] && !m.preview.Requested[item.ID]
	mark := "[ ]"
	if required {
		mark = "[•]"
	} else if m.requested[item.ID] {
		mark = "[✓]"
	}

	markerStyle := styles.DetailsValue
	diffStyle := changeStyle(item.Change)
	spaceStyle := styles.DetailsValue
	selectable := m.changeSelectable(item)
	if !selectable {
		markerStyle = styles.DetailsLabel.Faint(true)
		diffStyle = diffStyle.Faint(true)
		spaceStyle = spaceStyle.Faint(true)
	}

	line := selectedChangeSegment(spaceStyle, " ", selected) +
		selectedChangeSegment(markerStyle, mark, selected) +
		selectedChangeSegment(spaceStyle, " ", selected) +
		selectedChangeSegment(diffStyle, changeSymbol(item.Change)+" ", selected)
	if item.Kind == rcdiff.ItemCondition {
		colorStyle := styles.DetailsConditionValueStyle(m.conditionItemColor(item))
		if !selectable {
			colorStyle = colorStyle.Faint(true)
		}
		line += selectedChangeSegment(colorStyle, "●", selected) +
			selectedChangeSegment(diffStyle, " "+changeItemLabel(item), selected)
	} else {
		line += selectedChangeSegment(diffStyle, changeItemLabel(item), selected)
	}
	line += selectedChangeSegment(markerStyle, m.changeHint(item), selected)

	if selected {
		line = ansi.Truncate(line, width, "")
		padding := max(width-lipgloss.Width(line), 0)
		return line + styles.TreeItemSelectionStyle().Render(strings.Repeat(" ", padding))
	}
	return line
}

func selectedChangeSegment(style lipgloss.Style, value string, selected bool) string {
	if !selected {
		return style.Render(value)
	}
	if styles.NoColorEnabled() {
		return style.Reverse(true).Render(value)
	}
	return style.Background(styles.TreeItemSelectionStyle().GetBackground()).Render(value)
}

func (m Model) conditionItemColor(item rcpromote.Item) string {
	if m.plan == nil || item.Kind != rcdiff.ItemCondition {
		return ""
	}
	change := conditionChange(m.plan.Plan.Diff, item.ID.Name)
	if change.Final != nil {
		return change.Final.TagColor
	}
	if change.Current != nil {
		return change.Current.TagColor
	}
	return ""
}

func (m Model) changeHint(item rcpromote.Item) string {
	switch {
	case item.Change == rcdiff.ChangeRemoved && !m.prune:
		return " kept"
	case m.preview != nil && m.preview.Effective[item.ID] && !m.preview.Requested[item.ID]:
		return " required"
	default:
		return ""
	}
}

func (m Model) changeSelectable(item rcpromote.Item) bool {
	return item.Change != rcdiff.ChangeRemoved || m.prune
}

func (m *Model) syncLayout() {
	m.syncDetail()
}

func (m *Model) syncDetail() {
	innerWidth := max(m.width-2, 1)
	_, rightWidth := m.promotionColumnWidths(innerWidth)
	detailHeight := max(m.height-2-promoteHeaderHeight, 1)
	m.detail.SetWidth(rightWidth)
	m.detail.SetHeight(detailHeight)
	m.detail.SetContent(strings.Join(m.detailLines(), "\n"))
	m.detail.GotoTop()
}

func (m Model) detailLines() []string {
	width := max(m.detail.Width(), 1)
	height := max(m.detail.Height(), 1)
	for budget := detailValueMaxLineBudget; budget >= detailValueMinLineBudget; budget-- {
		lines := m.detailLinesWithOptions(detailRenderOptions{
			width:           width,
			valueLineBudget: budget,
		})
		if len(lines) <= height {
			return lines
		}
	}
	return m.detailLinesWithOptions(detailRenderOptions{
		width:           width,
		valueLineBudget: detailValueMinLineBudget,
		hideUnchanged:   true,
	})
}

func (m Model) detailLinesWithOptions(options detailRenderOptions) []string {
	item, ok := m.CurrentItem()
	if !ok || m.plan == nil {
		lines := []string{
			detailSection(" ", "Selected change"),
			"",
		}
		return append(lines, detailWrappedText(" ", "Move to a change to inspect it.", styles.DetailsEmptyValue, options.width)...)
	}
	identity := " " + styles.DetailsValue.Render(item.Label)
	switch item.Kind {
	case rcdiff.ItemParameter:
		identity = renderParameterIdentity(item.ID)
	case rcdiff.ItemCondition, rcdiff.ItemGroupDescription:
		identity = " " + styles.DetailsValue.Render(item.ID.Name)
	}
	lines := []string{
		detailSection(" ", "Selected change"),
		"",
	}
	lines = append(lines, viewutil.WrapRenderedLine(identity, options.width, 3)...)
	lines = append(lines,
		" "+changeStyle(item.Change).Render(changeActionLabel(item.Change, m.prune)),
		"",
	)
	switch item.Kind {
	case rcdiff.ItemParameter:
		change := parameterChange(m.plan.Plan.Diff, item.ID)
		lines = append(lines, parameterDetail(
			change,
			m.plan.Source.Project,
			m.plan.Target.Project,
			configConditions(m.plan.Plan.Source),
			configConditions(m.plan.Plan.Target),
			options,
		)...)
	case rcdiff.ItemCondition:
		change := conditionChange(m.plan.Plan.Diff, item.ID.Name)
		lines = append(lines, conditionDetail(change, m.plan.Source.Project, m.plan.Target.Project, options)...)
	case rcdiff.ItemGroupDescription:
		change := groupChange(m.plan.Plan.Diff, item.ID.Name)
		lines = append(lines, groupDescriptionDetail(change, m.plan.Source.Project, m.plan.Target.Project, options)...)
	}
	lines = append(lines, "")
	if item.Change == rcdiff.ChangeRemoved && !m.prune {
		lines = append(lines, detailWrappedText(" ", "Target-only item will be kept.", styles.DetailsEmptyValue, options.width)...)
		lines = append(lines, detailWrappedText(" ", "Enable pruning to make it selectable.", styles.DetailsEmptyValue, options.width)...)
	} else if m.preview != nil && m.preview.Effective[item.ID] && !m.preview.Requested[item.ID] {
		lines = append(lines, detailWrappedText(" ", "Automatically required by another selected change.", styles.DetailsEmptyValue, options.width)...)
	} else if m.requested[item.ID] {
		lines = append(lines, detailWrappedText(" ", "Selected for promotion.", styles.DetailsEmptyValue, options.width)...)
	} else {
		lines = append(lines, detailWrappedText(" ", "Not selected; target remains unchanged.", styles.DetailsEmptyValue, options.width)...)
	}
	return lines
}

func parameterDetail(
	change rcdiff.ParameterChange,
	sourceProject, targetProject core.Project,
	sourceConditions, targetConditions []firebase.RemoteConfigCondition,
	options detailRenderOptions,
) []string {
	hidden := hiddenParameterFields(change, options.hideUnchanged)
	lines := parameterSide(sourceProject, change.Group, change.Final, sourceConditions, hidden, options)
	lines = append(lines, detailTransitionLines()...)
	lines = append(lines, parameterSide(targetProject, parameterTargetGroup(change), change.Current, targetConditions, hidden, options)...)
	return lines
}

func parameterSide(
	project core.Project,
	group string,
	param *firebase.RemoteConfigParam,
	conditions []firebase.RemoteConfigCondition,
	hidden parameterHiddenFields,
	options detailRenderOptions,
) []string {
	if param == nil {
		return []string{
			detailProjectSection(project, options.width),
			"   " + styles.DetailsEmptyValue.Render("—"),
		}
	}
	lines := []string{detailProjectSection(project, options.width)}
	if !hidden.valueType {
		lines = append(lines, detailFieldLines("   ", "type", param.ValueType, styles.DetailsValue, options.width)...)
	}
	if !hidden.description {
		lines = append(lines, detailFieldLines("   ", "description", param.Description, styles.DetailsValue, options.width)...)
	}
	if !hidden.group {
		lines = append(lines, detailFieldLines("   ", "group", group, styles.ParameterGroup, options.width)...)
	}
	if hidden.values {
		return lines
	}

	valueLines := make([]string, 0)
	for _, condition := range coreparameters.OrderedConditionalKeys(param.ConditionalValues, conditions) {
		if hidden.conditionalValue[condition] {
			continue
		}
		value := param.ConditionalValues[condition]
		valueLines = append(valueLines, parameterValueLines(
			condition,
			conditionColor(conditions, condition),
			value,
			param.ValueType,
			options,
		)...)
	}
	if param.DefaultValue != nil && !hidden.defaultValue {
		valueLines = append(valueLines, parameterValueLines("default", "", *param.DefaultValue, param.ValueType, options)...)
	}
	if len(valueLines) > 0 {
		lines = append(lines, "   "+styles.DetailsLabel.Render("values:"))
		lines = append(lines, valueLines...)
	} else {
		lines = append(lines,
			"   "+styles.DetailsLabel.Render("values:"),
			"     "+styles.DetailsEmptyValue.Render("—"),
		)
	}
	return lines
}

func hiddenParameterFields(change rcdiff.ParameterChange, hide bool) parameterHiddenFields {
	hidden := parameterHiddenFields{conditionalValue: map[string]bool{}}
	if !hide || change.Final == nil || change.Current == nil {
		return hidden
	}
	hidden.valueType = change.Final.ValueType == change.Current.ValueType
	hidden.description = change.Final.Description == change.Current.Description
	hidden.group = change.Group == parameterTargetGroup(change)
	for condition, sourceValue := range change.Final.ConditionalValues {
		targetValue, ok := change.Current.ConditionalValues[condition]
		hidden.conditionalValue[condition] = ok && reflect.DeepEqual(sourceValue, targetValue)
	}
	if change.Final.DefaultValue != nil && change.Current.DefaultValue != nil {
		hidden.defaultValue = reflect.DeepEqual(*change.Final.DefaultValue, *change.Current.DefaultValue)
	}
	hidden.values = reflect.DeepEqual(change.Final.DefaultValue, change.Current.DefaultValue) &&
		reflect.DeepEqual(change.Final.ConditionalValues, change.Current.ConditionalValues)
	return hidden
}

func conditionDetail(
	change rcdiff.ConditionChange,
	sourceProject, targetProject core.Project,
	options detailRenderOptions,
) []string {
	hidden := hiddenConditionFields(change, options.hideUnchanged)
	lines := conditionSide(sourceProject, change.FinalPosition, change.Final, hidden, options)
	lines = append(lines, detailTransitionLines()...)
	lines = append(lines, conditionSide(targetProject, change.PreviousPosition, change.Current, hidden, options)...)
	return lines
}

func conditionSide(
	project core.Project,
	position int,
	condition *firebase.RemoteConfigCondition,
	hidden conditionHiddenFields,
	options detailRenderOptions,
) []string {
	if condition == nil {
		return []string{
			detailProjectSection(project, options.width),
			"   " + styles.DetailsEmptyValue.Render("—"),
		}
	}
	lines := []string{detailProjectSection(project, options.width)}
	if !hidden.position {
		lines = append(lines, detailFieldLines("   ", "position", fmt.Sprintf("%d", position), styles.ParameterName, options.width)...)
	}
	if !hidden.expression {
		lines = append(lines, detailFieldLines("   ", "expression", condition.Expression, styles.DetailsValue, options.width)...)
	}
	if !hidden.color {
		lines = append(lines, detailFieldLines(
			"   ",
			"color",
			viewutil.ConditionColorValue(condition.TagColor),
			styles.DetailsConditionValueStyle(condition.TagColor),
			options.width,
		)...)
	}
	return lines
}

func hiddenConditionFields(change rcdiff.ConditionChange, hide bool) conditionHiddenFields {
	if !hide || change.Final == nil || change.Current == nil {
		return conditionHiddenFields{}
	}
	return conditionHiddenFields{
		position:   change.FinalPosition == change.PreviousPosition,
		expression: change.Final.Expression == change.Current.Expression,
		color:      change.Final.TagColor == change.Current.TagColor,
	}
}

func groupDescriptionDetail(
	change rcdiff.GroupDescriptionChange,
	sourceProject, targetProject core.Project,
	options detailRenderOptions,
) []string {
	showDescription := !options.hideUnchanged || change.Kind != rcdiff.ChangeChanged || change.Final != change.Current
	lines := groupDescriptionSide(
		sourceProject,
		change.Final,
		change.Kind != rcdiff.ChangeRemoved,
		showDescription,
		options.width,
	)
	lines = append(lines, detailTransitionLines()...)
	lines = append(lines, groupDescriptionSide(
		targetProject,
		change.Current,
		change.Kind != rcdiff.ChangeAdded,
		showDescription,
		options.width,
	)...)
	return lines
}

func groupDescriptionSide(project core.Project, description string, exists, showDescription bool, width int) []string {
	lines := []string{detailProjectSection(project, width)}
	if !exists {
		return append(lines, "   "+styles.DetailsEmptyValue.Render("—"))
	}
	if showDescription {
		lines = append(lines, detailFieldLines("   ", "description", description, styles.DetailsValue, width)...)
	}
	return lines
}

func detailTransitionLines() []string {
	style := styles.BorderStyle(false)
	return []string{
		"   " + style.Render("🬞🬏"),
		"   " + style.Render("🭣🭘"),
	}
}

func detailProjectSection(project core.Project, width int) string {
	return " " + headerProjectLabel(project, max(width-1, 0))
}

func detailSection(indent, label string) string {
	return indent + styles.DetailsLabel.Render(label)
}

func detailFieldLines(indent, label, value string, valueStyle lipgloss.Style, width int) []string {
	return detailRenderedFieldLines(
		indent,
		label,
		styles.DetailsLabel,
		renderDetailValue(value, valueStyle),
		width,
		0,
		false,
	)
}

func renderDetailValue(value string, valueStyle lipgloss.Style) string {
	if strings.TrimSpace(value) == "" {
		return styles.DetailsEmptyValue.Render("—")
	}
	return valueStyle.Render(value)
}

func parameterValueLines(
	label, colorName string,
	value firebase.RemoteConfigValue,
	valueType string,
	options detailRenderOptions,
) []string {
	labelStyle := styles.DetailsConditionValueStyle(colorName)
	if label == "default" {
		labelStyle = styles.DetailsEmptyValue
	}
	text := formatValue(value)
	indent := "     "
	rendered, cropped := renderParameterValue(
		text,
		valueType,
		max(options.width-lipgloss.Width(indent+label+": "), 1),
		max(options.width-lipgloss.Width(indent)-2, 1),
		options.valueLineBudget,
	)
	return detailRenderedFieldLines(
		indent,
		label,
		labelStyle,
		rendered,
		options.width,
		options.valueLineBudget,
		cropped,
	)
}

func detailRenderedFieldLines(
	indent, label string,
	labelStyle lipgloss.Style,
	renderedValue string,
	width, lineBudget int,
	forceCrop bool,
) []string {
	prefix := indent + labelStyle.Render(label+": ")
	continuationIndent := lipgloss.Width(indent) + 2
	valueLines := strings.Split(renderedValue, "\n")
	lines := make([]string, 0, len(valueLines))
	for index, valueLine := range valueLines {
		line := strings.Repeat(" ", continuationIndent) + valueLine
		wrapIndent := continuationIndent + detailLeadingSpaceWidth(ansi.Strip(valueLine))
		if index == 0 {
			line = prefix + valueLine
			wrapIndent = continuationIndent
		}
		lines = append(lines, viewutil.WrapRenderedLine(line, width, wrapIndent)...)
	}
	return cropDetailLines(lines, lineBudget, width, forceCrop)
}

func detailWrappedText(indent, value string, style lipgloss.Style, width int) []string {
	line := indent + style.Render(value)
	return viewutil.WrapRenderedLine(line, width, lipgloss.Width(indent)+2)
}

func cropDetailLines(lines []string, budget, width int, force bool) []string {
	if budget <= 0 || len(lines) <= budget && !force {
		return lines
	}
	lines = append([]string(nil), lines[:min(len(lines), budget)]...)
	ellipsis := styles.DetailsEmptyValue.Render("…")
	lines[len(lines)-1] = ansi.Truncate(lines[len(lines)-1], max(width-lipgloss.Width(ellipsis), 0), "") + ellipsis
	return lines
}

func detailLeadingSpaceWidth(value string) int {
	return len(value) - len(strings.TrimLeft(value, " "))
}

func renderParameterValue(value, valueType string, firstWidth, continuationWidth, lineBudget int) (string, bool) {
	if strings.TrimSpace(value) == "" {
		return styles.DetailsEmptyValue.Render("—"), false
	}
	fragment, cropped := cropValueBeforeRender(value, firstWidth, continuationWidth, lineBudget)
	if strings.EqualFold(strings.TrimSpace(valueType), "json") {
		return jsoninput.HighlightJSONVisible(fragment), cropped
	}
	return corestyles.ValueTextStyle(value, valueType).Render(fragment), cropped
}

func cropValueBeforeRender(value string, firstWidth, continuationWidth, lineBudget int) (string, bool) {
	if lineBudget <= 0 {
		return value, false
	}
	firstWidth = max(firstWidth, 1)
	continuationWidth = max(continuationWidth, 1)

	var out strings.Builder
	graphemes := uniseg.NewGraphemes(value)
	line := 0
	lineWidth := 0
	lineLimit := firstWidth
	for graphemes.Next() {
		cluster := graphemes.Str()
		if cluster == "\r" {
			continue
		}
		if cluster == "\n" {
			_, end := graphemes.Positions()
			if line+1 >= lineBudget {
				return out.String(), end < len(value)
			}
			out.WriteByte('\n')
			line++
			lineWidth = 0
			lineLimit = continuationWidth
			continue
		}

		clusterWidth := graphemes.Width()
		if lineWidth > 0 && lineWidth+clusterWidth > lineLimit {
			if line+1 >= lineBudget {
				return out.String(), true
			}
			line++
			lineWidth = 0
			lineLimit = continuationWidth
		}
		out.WriteString(cluster)
		lineWidth += clusterWidth
	}
	return out.String(), false
}

func conditionColor(conditions []firebase.RemoteConfigCondition, name string) string {
	for _, condition := range conditions {
		if condition.Name == name {
			return condition.TagColor
		}
	}
	return ""
}

func configConditions(config *firebase.RemoteConfig) []firebase.RemoteConfigCondition {
	if config == nil {
		return nil
	}
	return config.Conditions
}

func parameterChange(result rcdiff.Result, id rcpromote.ItemID) rcdiff.ParameterChange {
	for _, change := range result.Parameters {
		if change.Key == id.Name && change.Group == id.Group {
			return change
		}
	}
	return rcdiff.ParameterChange{}
}

func conditionChange(result rcdiff.Result, name string) rcdiff.ConditionChange {
	for _, change := range result.Conditions {
		if change.Name == name {
			return change
		}
	}
	return rcdiff.ConditionChange{}
}

func groupChange(result rcdiff.Result, name string) rcdiff.GroupDescriptionChange {
	for _, change := range result.GroupDescriptions {
		if change.Group == name {
			return change
		}
	}
	return rcdiff.GroupDescriptionChange{}
}

func formatValue(value firebase.RemoteConfigValue) string {
	return value.Value
}

func projectName(project core.Project) string {
	if strings.TrimSpace(project.Name) == "" || project.Name == project.ProjectID {
		return project.ProjectID
	}
	return project.Name + " (" + project.ProjectID + ")"
}

func displayVersion(snapshot core.ProjectPromotionSnapshot) string {
	if snapshot.Source == "draft" && snapshot.DraftVersion != "" {
		return snapshot.DraftVersion
	}
	if snapshot.Version == "" {
		return "?"
	}
	return snapshot.Version
}

func itemKindLabel(kind rcdiff.ItemKind) string {
	switch kind {
	case rcdiff.ItemCondition:
		return "Conditions"
	case rcdiff.ItemGroupDescription:
		return "Groups"
	default:
		return "Parameters"
	}
}

func changeSymbol(kind rcdiff.ChangeKind) string {
	switch kind {
	case rcdiff.ChangeAdded:
		return "+"
	case rcdiff.ChangeRemoved:
		return "-"
	default:
		return "~"
	}
}

func changeActionLabel(kind rcdiff.ChangeKind, prune bool) string {
	switch kind {
	case rcdiff.ChangeAdded:
		return "+ TO ADD"
	case rcdiff.ChangeRemoved:
		if !prune {
			return "× TARGET-ONLY"
		}
		return "- TO REMOVE"
	default:
		return "~ TO UPDATE"
	}
}

func changeItemLabel(item rcpromote.Item) string {
	if item.Kind == rcdiff.ItemParameter || item.Kind == rcdiff.ItemCondition || item.Kind == rcdiff.ItemGroupDescription {
		return item.ID.Name
	}
	return item.Label
}

func renderParameterIdentity(id rcpromote.ItemID) string {
	name := styles.ParameterName.Render(id.Name)
	if id.Group == "" {
		return " " + name
	}
	return " " + styles.ParameterGroup.Render(id.Group) + styles.ParameterSeparator.Render(" / ") + name
}

func changeStyle(kind rcdiff.ChangeKind) lipgloss.Style {
	color := styles.PaletteChanged
	switch kind {
	case rcdiff.ChangeAdded:
		color = styles.PaletteAdded
	case rcdiff.ChangeRemoved:
		color = styles.PaletteRemoved
	}
	return lipgloss.NewStyle().Foreground(color)
}
