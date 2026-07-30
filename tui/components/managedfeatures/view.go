package managedfeatures

import (
	"strconv"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/core"
	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
	"github.com/yumauri/fbrcm/tui/components/viewutil"
	"github.com/yumauri/fbrcm/tui/components/workspaceheader"
	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/styles"
)

func (m Model) ViewWithBorder(active, borderActive bool) string {
	return m.viewWithWorkspacePreview(active, borderActive, false, 0)
}

// ViewWithWorkspacePreview renders a menu title in the workspace border row
// without storing transient app-level menu state in the managed-feature model.
func (m Model) ViewWithWorkspacePreview(active, borderActive bool, tab int) string {
	return m.viewWithWorkspacePreview(active, borderActive, true, tab)
}

func (m Model) viewWithWorkspacePreview(active, borderActive, previewOpen bool, previewTab int) string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}
	var footer []string
	if m.filterEnabled() {
		footer = m.filter.View(max(m.width-1, 1), active, m.visibleEntityCount())
	}
	borderStyle := styles.BorderStyle(borderActive)
	var titles string
	var titleWidth int
	if previewOpen {
		titles, titleWidth = workspaceheader.RenderMenuWithRightReserve(m.width, m.tabIndex(), previewTab, active, borderStyle, m.headerRightReserve)
	} else {
		titles, titleWidth = workspaceheader.RenderWithRightReserve(m.width, m.tabIndex(), active, borderStyle, m.headerRightReserve)
	}
	prefixWidth := min(2, m.width)
	lines := []string{
		borderStyle.Render("╭"+strings.Repeat("─", max(prefixWidth-1, 0))) +
			titles +
			borderStyle.Render(strings.Repeat("─", max(m.width-prefixWidth-titleWidth-1, 0))+"╮"),
	}
	body := m.bodyLines()
	innerWidth := max(m.width-2, 0)
	for row := range max(m.height-2-len(footer), 0) {
		line := ""
		if row < len(body) {
			line = body[row]
		}
		lines = append(lines, borderStyle.Render("│")+viewutil.PadRight(ansi.Truncate(line, innerWidth, ""), innerWidth)+borderStyle.Render("│"))
	}
	for index, line := range footer {
		left := "│"
		if index == 0 {
			left = "├"
		}
		lines = append(lines, borderStyle.Render(left)+line)
	}
	lines = append(lines, borderStyle.Render("╰"+strings.Repeat("─", max(m.width-2, 0))+"╯"))
	return strings.Join(lines, "\n")
}

func (m Model) tabIndex() int {
	switch m.kind {
	case messages.ManagedFeatureExperiment:
		return 3
	case messages.ManagedFeaturePersonalization:
		return 4
	case messages.ManagedFeatureRollout:
		return 5
	default:
		return 0
	}
}

func (m Model) bodyLines() []string {
	width := max(m.width-2, 1)
	if len(m.visible) == 0 {
		lines := []string{
			"Select a client template in Projects.",
			"",
			"Managed features are not available for server templates.",
		}
		for index := range lines {
			lines[index] = viewutil.PadRight(ansi.Truncate(lines[index], width, ""), width)
		}
		return lines
	}
	rows := make([]string, 0, m.visibleHeight())
	for index, node := range m.visible {
		rows = append(rows, m.renderNodeLines(node, index == m.cursor, width)...)
	}
	start := min(max(m.offset, 0), len(rows))
	end := min(start+m.contentHeight(), len(rows))
	return rows[start:end]
}

func (m Model) renderNode(node visibleNode, selected bool, width int) string {
	lines := m.renderNodeLines(node, selected, width)
	if len(lines) == 0 {
		return strings.Repeat(" ", width)
	}
	return lines[0]
}

func (m Model) renderNodeLines(node visibleNode, selected bool, width int) []string {
	if node.kind == nodeGap {
		return []string{strings.Repeat(" ", width)}
	}
	projectIndex, ok := m.projectIndex[node.projectID]
	if !ok {
		return []string{strings.Repeat(" ", width)}
	}
	project := m.projects[projectIndex]
	var lines []string
	if node.kind == nodeProject {
		lines = []string{m.renderProject(project, width)}
	} else {
		lines = m.renderEntity(project, node.index, width)
	}
	for index, line := range lines {
		lines[index] = m.styleNodeLine(line, node.kind, selected, width)
	}
	return lines
}

func (m Model) styleNodeLine(line string, kind nodeKind, selected bool, width int) string {
	if !selected {
		return viewutil.PadRight(ansi.Truncate(line, width, "…"), width)
	}
	selection := styles.TreeItemSelectionStyle()
	if kind == nodeProject {
		selection = styles.TreeProjectSelectionStyle()
	}
	return styles.FillSelectedLine(selection.Render(ansi.Truncate(ansi.Strip(line), width, "…")), width, selection)
}

func (m Model) renderProject(project projectState, width int) string {
	name := styles.TreeProjectName.Render(project.project.Name)
	id := styles.TreeProjectID.Render(project.project.ProjectID)
	meta := ""
	switch {
	case project.loading:
		meta = m.spin.View()
	case project.err != nil:
		meta = styles.SecondaryTitleError.Render("error: " + singleLineMeta(project.err.Error()))
	default:
		count := m.entityCount(project)
		if m.filterEnabled() {
			count = m.visibleProjectEntityCount(project.project.ProjectID)
		}
		version := ""
		if project.template.Version != "" {
			version = "RC v" + project.template.Version
		}
		meta = styles.PanelMuted.Render(joinMeta(
			version,
			rcdisplay.FormatCount(count, m.entityNoun(), m.entityNoun()+"s"),
		))
	}
	left := name + " " + id
	gap := max(width-lipgloss.Width(left)-lipgloss.Width(meta), 1)
	return ansi.Truncate(left+strings.Repeat(" ", gap)+meta, width, "…")
}

func (m Model) renderEntity(project projectState, index, width int) []string {
	if index < 0 || index >= m.entityCount(project) {
		return []string{"", ""}
	}
	id := m.entityID(project, index)
	label := id
	marker := ""
	state := ""
	meta := ""
	switch m.kind {
	case messages.ManagedFeatureExperiment:
		entity := project.experiments[index]
		if entity.Definition.DisplayName != "" {
			label = entity.Definition.DisplayName
		}
		state = entity.State
		marker = managedFeatureStateMarker(state)
		bindings := rcdisplay.FormatCount(len(entity.References), "binding", "bindings")
		if len(entity.References) == 0 {
			bindings = "bindings not exposed"
		}
		meta = joinRenderedMeta(
			renderManagedFeature(styles.PanelMuted, managedFeatureListID(id)),
			renderManagedFeature(styles.PanelMuted, bindings),
			renderManagedFeature(styles.PanelMuted, prefixedRelativeMeta("updated", entity.LastUpdateTime)),
		)
	case messages.ManagedFeaturePersonalization:
		entity := project.personalizations[index]
		marker = "◆"
		meta = personalizationReferencesMeta(entity.References)
	case messages.ManagedFeatureRollout:
		entity := project.rollouts[index]
		if entity.Definition.DisplayName != "" {
			label = entity.Definition.DisplayName
		}
		state = entity.State
		marker = managedFeatureStateMarker(state)
		meta = rolloutReferencesMeta(id, entity.References)
	}
	title := renderEntityTitle(marker, label, state, width)
	summary := "    "
	if meta != "" {
		summary += meta
	}
	return []string{title, summary}
}

func (m Model) entityNoun() string {
	switch m.kind {
	case messages.ManagedFeatureExperiment:
		return "A/B test"
	case messages.ManagedFeaturePersonalization:
		return "personalization"
	case messages.ManagedFeatureRollout:
		return "rollout"
	default:
		return "item"
	}
}

func joinMeta(values ...string) string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			out = append(out, value)
		}
	}
	return strings.Join(out, " · ")
}

func prefixedRelativeMeta(label, value string) string {
	relative := rcdisplay.FormatRelativeTime(value, time.Now())
	if relative == "—" {
		return ""
	}
	return label + " " + relative
}

func singleLineMeta(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func renderEntityTitle(marker, label, state string, width int) string {
	title := "  "
	if marker != "" {
		markerStyle := styles.PanelMuted
		if state != "" {
			markerStyle = styles.ManagedFeatureStatusStyle(state)
		}
		title += renderManagedFeature(markerStyle, marker) + " "
	}
	title += renderManagedFeature(styles.PanelText, label)
	if state != "" {
		title += renderManagedFeature(styles.ParameterSeparator, " · ") +
			renderManagedFeature(styles.ManagedFeatureStatusStyle(state), state)
	}
	return ansi.Truncate(title, width, "…")
}

func managedFeatureStateMarker(state string) string {
	state = strings.ToUpper(strings.TrimSpace(state))
	switch state {
	case "":
		return ""
	case "RUNNING":
		return "●"
	case "PENDING":
		return "◌"
	case "DONE", "EXPIRED":
		return "○"
	default:
		return "•"
	}
}

func renderManagedFeature(style lipgloss.Style, value string) string {
	if value == "" || styles.NoColorEnabled() {
		return value
	}
	return style.Render(value)
}

func joinRenderedMeta(values ...string) string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(ansi.Strip(value)) != "" {
			out = append(out, value)
		}
	}
	return strings.Join(out, renderManagedFeature(styles.ParameterSeparator, " · "))
}

func personalizationReferencesMeta(references []core.ManagedValueReference) string {
	switch len(references) {
	case 0:
		return renderManagedFeature(styles.PanelMuted, "no parameter values exposed")
	case 1:
		reference := references[0]
		return viewutil.RenderManagedReferenceIdentity(reference)
	default:
		return joinRenderedMeta(
			renderManagedFeature(styles.PanelMuted, rcdisplay.FormatCount(len(references), "parameter value", "parameter values")),
			managedReferenceNames(references, 2),
		)
	}
}

func rolloutReferencesMeta(id string, references []core.ManagedValueReference) string {
	if len(references) == 0 {
		return joinRenderedMeta(
			renderManagedFeature(styles.PanelMuted, managedFeatureListID(id)),
			renderManagedFeature(styles.PanelMuted, "no parameter values exposed"),
		)
	}
	percentage := managedReferencePercentage(references[0])
	if len(references) == 1 {
		reference := references[0]
		assignment := viewutil.RenderManagedReferencePath(reference)
		if reference.Value != nil {
			assignment += renderManagedFeature(styles.ParameterSeparator, " = ") +
				renderManagedFeature(styles.PanelText, managedReferenceValue(reference.Value))
		}
		if percentage != "" {
			assignment = renderManagedFeature(styles.SecondaryTitleCount, percentage) +
				renderManagedFeature(styles.ParameterSeparator, " → ") + assignment
		}
		return joinRenderedMeta(
			assignment,
			renderManagedFeature(styles.PanelMuted, managedFeatureListID(id)),
		)
	}
	return joinRenderedMeta(
		renderManagedFeature(styles.SecondaryTitleCount, percentage),
		renderManagedFeature(styles.PanelMuted, rcdisplay.FormatCount(len(references), "parameter value", "parameter values")),
		managedReferenceNames(references, 2),
		renderManagedFeature(styles.PanelMuted, managedFeatureListID(id)),
	)
}

func managedReferenceNames(references []core.ManagedValueReference, limit int) string {
	names := make([]string, 0, min(len(references), limit))
	seen := make(map[string]struct{}, len(references))
	for _, reference := range references {
		name := viewutil.ManagedReferencePathText(reference)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		if len(names) < limit {
			names = append(names, viewutil.RenderManagedReferencePath(reference))
		}
	}
	remaining := len(seen) - len(names)
	if remaining > 0 {
		names = append(names, renderManagedFeature(styles.PanelMuted, "+"+strconv.Itoa(remaining)+" more"))
	}
	return strings.Join(names, renderManagedFeature(styles.ParameterSeparator, ", "))
}

func managedReferencePercentage(reference core.ManagedValueReference) string {
	if reference.Percentage == nil {
		return ""
	}
	return strconv.FormatFloat(*reference.Percentage, 'f', -1, 64) + "%"
}

func managedReferenceValue(value *string) string {
	if value == nil {
		return ""
	}
	if *value == "" {
		return `""`
	}
	return singleLineMeta(*value)
}

func managedFeatureListID(id string) string {
	if id == "" {
		return ""
	}
	return "#" + id
}
