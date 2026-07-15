package parameters

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/yumauri/fbrcm/core"
	rcdiff "github.com/yumauri/fbrcm/core/rc/diff"
	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
	"github.com/yumauri/fbrcm/tui/components/viewutil"
	"github.com/yumauri/fbrcm/tui/styles"
)

func (m Model) renderHistoryStackedNode(node visibleNode, selected bool) []string {
	if node.kind != nodeParameter && node.kind != nodeValue {
		return []string{m.renderNode(node, selected)}
	}
	if node.kind == nodeParameter {
		previous, current, kind := m.historyParameterPair(node.projectID, node.groupKey, node.paramKey)
		param := current
		if param == nil {
			param = previous
		}
		headerText := node.label
		if param != nil {
			headerText = param.Key
		}
		headerStyle := parameterStyle
		if background := historyChangeBackground(kind); background != nil {
			headerStyle = headerStyle.Background(background)
		}
		if param != nil && isDeprecatedDescription(param.Description) {
			headerStyle = headerStyle.Strikethrough(true).Faint(true)
		}
		header := "  " + m.renderHighlightedParameterKey(headerText, headerStyle, selected)
		if selected {
			header = styles.FillSelectedLine(parameterSelectionStyle().Render(header), m.viewportWidth(), parameterSelectionStyle())
		} else {
			header = viewutil.PadRight(parameterStyle.Render(header), m.viewportWidth())
		}
		if node.expanded {
			return []string{header}
		}
		return []string{header}
	}
	previous, current := m.historyValuePair(node.projectID, node.groupKey, node.paramKey, node.label)
	kind := m.historyValueKind(node.projectID, node.groupKey, node.paramKey, node.label)
	return m.renderStackedValuePair(node, previous, current, kind, selected)
}

func (m Model) historyStackedValueLineCount(node visibleNode) int {
	count := 0
	previous, current := m.historyValuePair(node.projectID, node.groupKey, node.paramKey, node.label)
	if previous != nil {
		count++
	}
	if current != nil {
		count++
	}
	return max(count, 1)
}

func (m Model) renderStackedValuePair(node visibleNode, previous, current *core.ParametersValue, kind rcdiff.ChangeKind, selected bool) []string {
	merged := m.historyMergedParameter(node.projectID, node.groupKey, node.paramKey)
	label := rcdisplay.FormatConditionLabel(node.label)
	labelStyle := conditionDefaultStyle
	if value := m.historyMergedValue(node.projectID, node.groupKey, node.paramKey, node.label); value != nil && value.Label != "default" {
		labelStyle = m.conditionStyle(value.Color)
	}
	connector := m.valueConnector(node, merged)
	paramStart := m.parameterRenderLayout().paramStart
	firstTree := compactBranchGlyph(paramStart, connector)
	tree := leafLineStyle.Render(firstTree)
	fillerWidth := max(parameterConditionWidth(merged)-lipgloss.Width(label)+1, 1)
	conditionPrefix := tree + " " + labelStyle.Render(label) + leafLineStyle.Render(" "+strings.Repeat("╌", fillerWidth)+" ")
	lines := make([]string, 0, 2)
	if previous != nil {
		lines = append(lines, m.renderStackedConditionVersion(conditionPrefix, m.stackedVersion(node.projectID, true), previous, kind, selected))
	}
	if current != nil {
		prefix := conditionPrefix
		if previous != nil {
			continuation := strings.Repeat(" ", paramStart)
			if connector != "last" && connector != "single" {
				continuation += "│"
			}
			prefix = leafLineStyle.Render(continuation) + strings.Repeat(" ", max(lipgloss.Width(conditionPrefix)-lipgloss.Width(continuation), 0))
		}
		lines = append(lines, m.renderStackedConditionVersion(prefix, m.stackedVersion(node.projectID, false), current, kind, selected))
	}
	return lines
}

func (m Model) renderStackedConditionVersion(prefix, version string, value *core.ParametersValue, kind rcdiff.ChangeKind, selected bool) string {
	if value == nil {
		return strings.Repeat(" ", m.viewportWidth())
	}
	valuePrefix := parameterValueStyle.Render(version+" ") + iconStyle.Render("╌") + " "
	remaining := max(m.viewportWidth()-lipgloss.Width(prefix)-lipgloss.Width(valuePrefix), 0)
	line := prefix + valuePrefix + m.renderHistoryTypedValue(value, kind, selected, remaining)
	return viewutil.PadRight(line, m.viewportWidth())
}

func (m Model) stackedVersion(projectID string, previous bool) string {
	history := m.histories[projectID]
	if previous {
		return "v" + history.previousVersion
	}
	return "v" + history.currentVersion
}
