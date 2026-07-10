package parameters

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/filter"
	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
	corestyles "github.com/yumauri/fbrcm/core/styles"
	"github.com/yumauri/fbrcm/tui/components/jsoninput"
	"github.com/yumauri/fbrcm/tui/components/viewutil"
	"github.com/yumauri/fbrcm/tui/styles"
)

func (m Model) renderNode(node visibleNode, selected bool) string {
	switch node.kind {
	case nodeProject:
		return m.renderProjectNode(node, selected, false)
	case nodeGroup:
		return m.renderGroupNode(node, selected, false)
	case nodeParameter:
		return m.renderParameterNode(node, selected)
	case nodeValue:
		return m.renderValueNode(node, selected)
	default:
		return ""
	}
}

func (m Model) renderParameterNode(node visibleNode, selected bool) string {
	width := max(m.width-2, 0)
	layout := m.parameterRenderLayout()
	param := m.parameterByKey(node.projectID, node.groupKey, node.paramKey)
	if param == nil {
		return strings.Repeat(" ", width)
	}

	namePad := strings.Repeat(" ", max(layout.nameWidth-lipgloss.Width(param.Key), 0))
	style := parameterStyle
	if selected {
		if styles.NoColorEnabled() {
			style = lipgloss.NewStyle().Reverse(true)
		} else {
			style = style.Background(styles.PaletteBlueDeep).Foreground(styles.PaletteSlateBright)
		}
	}
	if isDeprecatedDescription(param.Description) {
		style = style.Strikethrough(true).Faint(true)
	}

	if node.expanded {
		left := lipgloss.NewStyle().Render("  ") + m.renderHighlightedParameterKey(param.Key, style, selected)
		if layout.mode == parameterRenderModeRegular && param.Description != "" {
			descStyle := descriptionStyle
			if selected {
				if styles.NoColorEnabled() {
					descStyle = lipgloss.NewStyle().Reverse(true).Italic(true)
				} else {
					descStyle = descStyle.Background(styles.PaletteBlueDeep).Foreground(styles.PaletteSlateBright)
				}
			}
			left += "  "
			left += descStyle.Render(param.Description)
		}
		if selected {
			prefix := parameterSelectionStyle().Render("  ")
			left = prefix + m.renderHighlightedParameterKey(param.Key, style, selected)
			if layout.mode == parameterRenderModeRegular && param.Description != "" {
				var descStyle lipgloss.Style
				if styles.NoColorEnabled() {
					descStyle = lipgloss.NewStyle().Reverse(true).Italic(true)
				} else {
					descStyle = descriptionStyle.Background(styles.PaletteBlueDeep).Foreground(styles.PaletteSlateBright)
				}
				left += parameterSelectionStyle().Render("  ")
				left += descStyle.Render(param.Description)
			}
			return fillSelectedLine(left, width, parameterSelectionStyle())
		}
		return viewutil.PadRight(left, width)
	}

	if layout.mode == parameterRenderModeNarrow {
		line := lipgloss.NewStyle().Render("  ") + m.renderHighlightedParameterKey(param.Key, style, selected)
		if selected {
			prefix := parameterSelectionStyle().Render("  ")
			line = prefix + m.renderHighlightedParameterKey(param.Key, style, selected)
			return fillSelectedLine(line, width, parameterSelectionStyle())
		}
		return viewutil.PadRight(line, width)
	}

	icon := "╌"
	if len(param.Values) > 1 {
		icon = "⌥"
	}
	prefixStyle := lipgloss.NewStyle()
	iconLineStyle := iconStyle
	separatorLineStyle := parameterSeparatorStyle
	if selected {
		if styles.NoColorEnabled() {
			prefixStyle = parameterSelectionStyle()
			iconLineStyle = parameterSelectionStyle()
			separatorLineStyle = parameterSelectionStyle()
		} else {
			prefixStyle = prefixStyle.Background(styles.PaletteBlueDeep)
			iconLineStyle = iconLineStyle.Background(styles.PaletteBlueDeep).Foreground(styles.PaletteSlateBright)
			separatorLineStyle = separatorLineStyle.Background(styles.PaletteBlueDeep).Foreground(styles.PaletteSlateBright)
		}
	}
	left := prefixStyle.Render("  ") + m.renderHighlightedParameterKey(param.Key, style, selected) + prefixStyle.Render(namePad)
	left += prefixStyle.Render(strings.Repeat(" ", 2)) + iconLineStyle.Render(icon)
	left += prefixStyle.Render(" ")
	line := left + m.renderCollapsedParameterValues(param.Values, separatorLineStyle, selected)
	if selected {
		return fillSelectedLine(line, width, parameterSelectionStyle())
	}
	return viewutil.PadRight(line, width)
}

func (m Model) renderValueNode(node visibleNode, selected bool) string {
	width := max(m.width-2, 0)
	layout := m.parameterRenderLayout()
	param := m.parameterByKey(node.projectID, node.groupKey, node.paramKey)
	if param == nil || node.valueIdx < 0 || node.valueIdx >= len(param.Values) {
		if node.label != "" {
			return viewutil.PadRight(parameterValueStyle.Render(node.label), width)
		}
		return strings.Repeat(" ", width)
	}

	value := param.Values[node.valueIdx]
	labelStyle := m.conditionStyle(value.Color)
	if value.Label == "default" {
		labelStyle = conditionDefaultStyle
	}

	conditionLabel := rcdisplay.FormatConditionLabel(value.Label)
	conditionWidth := parameterConditionWidth(param)
	connector := m.valueConnector(node, param)
	label := conditionLabel
	var tree string
	var fillerWidth int
	if layout.mode == parameterRenderModeNarrow {
		tree = leafLineStyle.Render(compactBranchGlyph(layout.paramStart, connector))
		fillerWidth = max(conditionWidth-lipgloss.Width(label)+1, 1)
	} else {
		leafOffset := 1
		if len(param.Values) == 1 {
			leafOffset = 2
		}
		leafOffset++
		leafValueStart := layout.valueStart + leafOffset
		labelStart := max(leafValueStart-conditionWidth-4, layout.paramStart+2)
		tree = leafLineStyle.Render(branchGlyph(layout.paramStart, labelStart, connector))
		fillerWidth = max(leafValueStart-labelStart-lipgloss.Width(label)-3, 1)
	}
	filler := strings.Repeat("╌", fillerWidth)
	valueRendered := m.renderParameterValue(value, selected)
	line := tree + " " + labelStyle.Render(label) + leafLineStyle.Render(" "+filler+" ") + valueRendered
	return viewutil.PadRight(line, width)
}

func (m Model) renderCollapsedParameterValues(values []core.ParametersValue, separatorStyle lipgloss.Style, selected bool) string {
	parts := make([]string, 0, max(0, len(values)*2-1))
	for i, value := range values {
		if i > 0 {
			parts = append(parts, separatorStyle.Render(" / "))
		}
		parts = append(parts, m.renderParameterValueWithBase(value, selected))
	}
	return strings.Join(parts, "")
}

func (m Model) renderParameterValue(value core.ParametersValue, selected bool) string {
	return m.renderParameterValueWithBase(value, selected)
}

func (m Model) renderParameterValueWithBase(value core.ParametersValue, selected bool) string {
	if value.Empty {
		style := corestyles.EmptyValueStyle()
		if selected {
			if styles.NoColorEnabled() {
				style = lipgloss.NewStyle().Reverse(true).Italic(true)
			} else {
				style = style.Background(styles.PaletteBlueDeep).Foreground(styles.PaletteSlateBright)
			}
		}
		return style.Render(value.Value)
	}
	if selected {
		return valueSelectionStyle().Render(value.Value)
	}
	if strings.EqualFold(strings.TrimSpace(value.ValueType), "json") {
		return jsoninput.HighlightJSONVisible(value.Value)
	}
	return corestyles.ValueTextStyle(value.Value, value.ValueType).Render(value.Value)
}

func (m Model) renderHighlightedParameterKey(text string, baseStyle lipgloss.Style, selected bool) string {
	query := m.filter.Value()
	if query == "" {
		return baseStyle.Render(text)
	}

	_, indices := filter.Match(text, query, m.filter.Mode())
	if len(indices) == 0 {
		return baseStyle.Render(text)
	}

	highlighted := indicesSet(indices)
	highlightStyle := baseStyle.Foreground(styles.PaletteYellow)
	if selected {
		highlightStyle = baseStyle.Foreground(styles.PaletteYellow)
	}

	var builder strings.Builder
	for i, r := range []rune(text) {
		style := baseStyle
		if highlighted[i] {
			style = highlightStyle
		}
		builder.WriteString(style.Render(string(r)))
	}
	return builder.String()
}

func (m Model) valueConnector(node visibleNode, param *core.ParametersEntry) string {
	if param == nil {
		return "last"
	}
	if len(param.Values) == 1 {
		return "single"
	}
	if node.valueIdx >= len(param.Values)-1 {
		return "last"
	}
	if node.valueIdx == 0 {
		return "first"
	}
	return "mid"
}

func branchGlyph(paramStart, labelStart int, connector string) string {
	totalWidth := max(labelStart-1, 1)
	if totalWidth <= paramStart {
		return strings.Repeat(" ", max(totalWidth-2, 0)) + "╰╌"
	}

	if connector == "first" {
		return strings.Repeat(" ", paramStart) + "╰" + strings.Repeat("╌", max(totalWidth-paramStart-3, 0)) + "┬╌"
	}
	if connector == "single" {
		return strings.Repeat(" ", paramStart) + "╰" + strings.Repeat("╌", max(totalWidth-paramStart-2, 0))
	}

	prefixWidth := max(totalWidth-2, 0)
	switch connector {
	case "mid":
		return strings.Repeat(" ", prefixWidth) + "├╌"
	default:
		return strings.Repeat(" ", prefixWidth) + "╰╌"
	}
}

func compactBranchGlyph(paramStart int, connector string) string {
	prefixWidth := max(paramStart, 0)
	switch connector {
	case "last", "single":
		return strings.Repeat(" ", prefixWidth) + "╰╌"
	default:
		return strings.Repeat(" ", prefixWidth) + "├╌"
	}
}

func parameterConditionWidth(param *core.ParametersEntry) int {
	width := lipgloss.Width("Default value")
	if param == nil {
		return width
	}
	for _, value := range param.Values {
		width = max(width, lipgloss.Width(rcdisplay.FormatConditionLabel(value.Label)))
	}
	return width
}

func (m Model) conditionStyle(color string) lipgloss.Style {
	return styles.PanelText.Foreground(styles.ConditionLipglossColor(color))
}
