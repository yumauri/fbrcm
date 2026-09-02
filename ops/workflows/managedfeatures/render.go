package managedfeatures

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
	clistyles "github.com/yumauri/fbrcm/internal/terminal/styles"
	"github.com/yumauri/fbrcm/ops/shared"
)

func renderExperimentsTable(experiments []core.ExperimentEntry) string {
	return renderExperimentsTableAtWidth(experiments, shared.TerminalWidth())
}

func renderExperimentsTableAtWidth(experiments []core.ExperimentEntry, terminalWidth int) string {
	return renderExperimentsTableAt(experiments, terminalWidth, time.Now())
}

func renderExperimentsTableAt(experiments []core.ExperimentEntry, terminalWidth int, now time.Time) string {
	headers := []string{"ID", "Name", "Parameter", "Condition", "Exposure", "Last update", "State"}
	var rows [][]string
	var rowStyles []managedFeatureTableRowStyle
	widths := shared.HeaderWidths(headers)
	for _, experiment := range experiments {
		id := firebase.ManagedFeatureID(experiment.Name)
		if len(experiment.References) == 0 {
			row := []string{
				id, emptyDash(experiment.Definition.DisplayName), "—", "—", "—",
				formatManagedFeatureRelativeTime(experiment.LastUpdateTime, now), emptyDash(experiment.State),
			}
			shared.UpdateTableWidths(widths, row)
			rows = append(rows, row)
			rowStyles = append(rowStyles, managedFeatureTableRowStyle{})
			continue
		}
		for _, reference := range experiment.References {
			row := []string{
				id,
				emptyDash(experiment.Definition.DisplayName),
				reference.Parameter,
				referenceValueLabel(reference),
				formatPercentage(reference.Percentage),
				formatManagedFeatureRelativeTime(experiment.LastUpdateTime, now),
				emptyDash(experiment.State),
			}
			shared.UpdateTableWidths(widths, row)
			rows = append(rows, row)
			rowStyles = append(rowStyles, managedFeatureTableRowStyle{
				hasParameter: true, hasCondition: !reference.Default && strings.TrimSpace(reference.Condition) != "",
				conditionColor: reference.ConditionColor,
			})
		}
	}
	fitFlexibleColumns(widths, terminalWidth, 1, 2, 3, 5)
	truncateHeaders(headers, widths)
	truncateColumns(rows, widths, 1, 2, 3, 5)
	return shared.StyledTable(headers, rows, widths, map[int]bool{4: true}, managedFeatureTableStyle(rowStyles, 2, 3))
}

func renderRolloutsTable(rollouts []core.RolloutEntry) string {
	return renderRolloutsTableAtWidth(rollouts, shared.TerminalWidth())
}

func renderRolloutsTableAtWidth(rollouts []core.RolloutEntry, terminalWidth int) string {
	return renderRolloutsTableAt(rollouts, terminalWidth, time.Now())
}

func renderRolloutsTableAt(rollouts []core.RolloutEntry, terminalWidth int, now time.Time) string {
	headers := []string{"ID", "Name", "Parameter", "Condition", "Percentage", "Last update", "State"}
	var rows [][]string
	var rowStyles []managedFeatureTableRowStyle
	widths := shared.HeaderWidths(headers)
	for _, rollout := range rollouts {
		id := firebase.ManagedFeatureID(rollout.Name)
		if len(rollout.References) == 0 {
			row := []string{
				id,
				emptyDash(rollout.Definition.DisplayName),
				"—",
				"—",
				"—",
				formatManagedFeatureRelativeTime(rollout.LastUpdateTime, now),
				emptyDash(rollout.State),
			}
			shared.UpdateTableWidths(widths, row)
			rows = append(rows, row)
			rowStyles = append(rowStyles, managedFeatureTableRowStyle{})
			continue
		}
		for _, reference := range rollout.References {
			row := []string{
				id,
				emptyDash(rollout.Definition.DisplayName),
				reference.Parameter,
				referenceValueLabel(reference),
				formatPercentage(reference.Percentage),
				formatManagedFeatureRelativeTime(rollout.LastUpdateTime, now),
				emptyDash(rollout.State),
			}
			shared.UpdateTableWidths(widths, row)
			rows = append(rows, row)
			rowStyles = append(rowStyles, managedFeatureTableRowStyle{
				hasParameter:   true,
				hasCondition:   !reference.Default && strings.TrimSpace(reference.Condition) != "",
				conditionColor: reference.ConditionColor,
			})
		}
	}
	fitFlexibleColumns(widths, terminalWidth, 1, 2, 3, 5)
	truncateHeaders(headers, widths)
	truncateColumns(rows, widths, 1, 2, 3, 5)
	return shared.StyledTable(headers, rows, widths, map[int]bool{4: true}, managedFeatureTableStyle(rowStyles, 2, 3))
}

func renderPersonalizationsTable(personalizations []core.PersonalizationEntry) string {
	return renderPersonalizationsTableAtWidth(personalizations, shared.TerminalWidth())
}

func renderPersonalizationsTableAtWidth(personalizations []core.PersonalizationEntry, terminalWidth int) string {
	headers := []string{"ID", "Group", "Parameter", "Condition"}
	var rows [][]string
	var rowStyles []managedFeatureTableRowStyle
	widths := shared.HeaderWidths(headers)
	for _, personalization := range personalizations {
		for _, reference := range personalization.References {
			row := []string{personalization.ID, reference.Group, reference.Parameter, referenceValueLabel(reference)}
			shared.UpdateTableWidths(widths, row)
			rows = append(rows, row)
			rowStyles = append(rowStyles, managedFeatureTableRowStyle{
				hasParameter:   true,
				hasCondition:   !reference.Default && strings.TrimSpace(reference.Condition) != "",
				conditionColor: reference.ConditionColor,
			})
		}
	}
	fitFlexibleColumns(widths, terminalWidth, 0, 2, 3, 1)
	truncateHeaders(headers, widths)
	truncateColumns(rows, widths, 0, 1, 2, 3)
	return shared.StyledTable(headers, rows, widths, nil, managedFeatureTableStyle(rowStyles, 2, 3))
}

type managedFeatureTableRowStyle struct {
	hasParameter   bool
	hasCondition   bool
	conditionColor string
}

func managedFeatureTableStyle(rows []managedFeatureTableRowStyle, parameterColumn, conditionColumn int) func(int, int, lipgloss.Style) lipgloss.Style {
	return func(row, col int, style lipgloss.Style) lipgloss.Style {
		if row < 0 || row >= len(rows) {
			return style
		}
		switch {
		case col == parameterColumn && rows[row].hasParameter:
			return style.Foreground(clistyles.PaletteBlueBright)
		case col == conditionColumn && rows[row].hasCondition:
			return style.Foreground(clistyles.ConditionLipglossColor(rows[row].conditionColor))
		default:
			return style
		}
	}
}

func renderExperimentDetails(experiment core.ExperimentEntry) string {
	return renderExperimentDetailsAtWidth(experiment, shared.TerminalWidth())
}

func renderExperimentDetailsAtWidth(experiment core.ExperimentEntry, terminalWidth int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "ID: %s\n", firebase.ManagedFeatureID(experiment.Name))
	fmt.Fprintf(&b, "Name: %s\n", emptyDash(experiment.Definition.DisplayName))
	fmt.Fprintf(&b, "State: %s\n", emptyDash(experiment.State))
	fmt.Fprintf(&b, "Service: %s\n", emptyDash(experiment.Definition.Service))
	fmt.Fprintf(&b, "Description: %s\n", emptyDash(experiment.Definition.Description))
	fmt.Fprintf(&b, "Started: %s\n", formatManagedFeatureTime(experiment.StartTime))
	fmt.Fprintf(&b, "Ended: %s\n", formatManagedFeatureTime(experiment.EndTime))
	fmt.Fprintf(&b, "Updated: %s\n", formatManagedFeatureTime(experiment.LastUpdateTime))
	fmt.Fprintf(&b, "Activation event: %s\n", emptyDash(experiment.Definition.Objectives.ActivationEvent.Event))
	fmt.Fprintf(&b, "Variants: %s\n\n", rcdisplay.FormatCount(len(experiment.Definition.Variants), "variant", "variants"))
	b.WriteString(renderExperimentVariantsTableAtWidth(experiment.Definition.Variants, terminalWidth))
	fmt.Fprintf(&b, "\n\nObjectives: %s\n\n", rcdisplay.FormatCount(len(experiment.Definition.Objectives.EventObjectives), "objective", "objectives"))
	b.WriteString(renderExperimentObjectivesTableAtWidth(experiment.Definition.Objectives.EventObjectives, terminalWidth))
	fmt.Fprintf(&b, "\n\nBindings: %s\n\n", rcdisplay.FormatCount(len(experiment.References), "parameter value", "parameter values"))
	b.WriteString(renderExperimentReferencesTableAtWidth(experiment.References, terminalWidth))
	return b.String()
}

func renderExperimentVariantsTableAtWidth(variants []firebase.ExperimentVariant, terminalWidth int) string {
	headers := []string{"Name", "Weight"}
	rows := make([][]string, 0, len(variants))
	widths := shared.HeaderWidths(headers)
	for _, variant := range variants {
		row := []string{emptyDash(variant.Name), strconv.Itoa(variant.Weight)}
		shared.UpdateTableWidths(widths, row)
		rows = append(rows, row)
	}
	fitFlexibleColumns(widths, terminalWidth, 0)
	truncateHeaders(headers, widths)
	truncateColumns(rows, widths, 0)
	return shared.StyledTable(headers, rows, widths, map[int]bool{1: true}, nil)
}

func renderExperimentObjectivesTableAtWidth(objectives []firebase.ExperimentEventObjective, terminalWidth int) string {
	headers := []string{"Primary", "Objective", "Optimization"}
	rows := make([][]string, 0, len(objectives))
	widths := shared.HeaderWidths(headers)
	for _, objective := range objectives {
		row := []string{strconv.FormatBool(objective.IsPrimary), experimentObjectiveName(objective), emptyDash(objective.ABTOptimizationFunction)}
		shared.UpdateTableWidths(widths, row)
		rows = append(rows, row)
	}
	fitFlexibleColumns(widths, terminalWidth, 1, 2)
	truncateHeaders(headers, widths)
	truncateColumns(rows, widths, 1, 2)
	return shared.StyledTable(headers, rows, widths, nil, nil)
}

func renderRolloutDetails(rollout core.RolloutEntry) string {
	return renderRolloutDetailsAtWidth(rollout, shared.TerminalWidth())
}

func renderRolloutDetailsAtWidth(rollout core.RolloutEntry, terminalWidth int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "ID: %s\n", firebase.ManagedFeatureID(rollout.Name))
	fmt.Fprintf(&b, "Name: %s\n", emptyDash(rollout.Definition.DisplayName))
	fmt.Fprintf(&b, "State: %s\n", emptyDash(rollout.State))
	fmt.Fprintf(&b, "Description: %s\n", emptyDash(rollout.Definition.Description))
	fmt.Fprintf(&b, "Created: %s\n", formatManagedFeatureTime(rollout.CreateTime))
	fmt.Fprintf(&b, "Started: %s\n", formatManagedFeatureTime(rollout.StartTime))
	fmt.Fprintf(&b, "Ended: %s\n", formatManagedFeatureTime(rollout.EndTime))
	fmt.Fprintf(&b, "Updated: %s\n", formatManagedFeatureTime(rollout.LastUpdateTime))
	fmt.Fprintf(&b, "Control variant: %s\n", formatExperimentVariant(rollout.Definition.ControlVariant))
	fmt.Fprintf(&b, "Enabled variant: %s\n", formatExperimentVariant(rollout.Definition.EnabledVariant))
	fmt.Fprintf(&b, "Bindings: %s\n\n", rcdisplay.FormatCount(len(rollout.References), "parameter value", "parameter values"))
	b.WriteString(renderManagedValueReferencesTableAtWidth(rollout.References, true, terminalWidth))
	return b.String()
}

func renderPersonalizationDetails(personalization core.PersonalizationEntry) string {
	return renderPersonalizationDetailsAtWidth(personalization, shared.TerminalWidth())
}

func renderPersonalizationDetailsAtWidth(personalization core.PersonalizationEntry, terminalWidth int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "ID: %s\n", personalization.ID)
	fmt.Fprintf(&b, "Bindings: %s\n\n", rcdisplay.FormatCount(len(personalization.References), "parameter value", "parameter values"))
	b.WriteString(renderManagedValueReferencesTableAtWidth(personalization.References, false, terminalWidth))
	b.WriteString("\n\nFirebase does not expose personalization value candidates or result metrics through this API.")
	return b.String()
}

func renderManagedValueReferencesTableAtWidth(references []core.ManagedValueReference, includeRollout bool, terminalWidth int) string {
	headers := []string{"Group", "Parameter", "Condition", "Type"}
	if includeRollout {
		headers = append(headers, "Value", "Traffic")
	}
	rows := make([][]string, 0, len(references))
	widths := shared.HeaderWidths(headers)
	for _, reference := range references {
		row := []string{reference.Group, reference.Parameter, referenceValueLabel(reference), reference.ValueType}
		if includeRollout {
			row = append(row, formatOptionalValue(reference.Value), formatPercentage(reference.Percentage))
		}
		shared.UpdateTableWidths(widths, row)
		rows = append(rows, row)
	}
	rightAligned := map[int]bool{}
	if includeRollout {
		rightAligned[5] = true
	}
	flexible := []int{1, 2, 0}
	if includeRollout {
		flexible = append(flexible, 4)
	}
	fitFlexibleColumns(widths, terminalWidth, flexible...)
	truncateHeaders(headers, widths)
	truncateColumns(rows, widths, flexible...)
	styles := make([]managedFeatureTableRowStyle, len(references))
	for index, reference := range references {
		styles[index] = managedFeatureTableRowStyle{
			hasParameter: true, hasCondition: !reference.Default && strings.TrimSpace(reference.Condition) != "",
			conditionColor: reference.ConditionColor,
		}
	}
	return shared.StyledTable(headers, rows, widths, rightAligned, managedFeatureTableStyle(styles, 1, 2))
}

func renderExperimentReferencesTableAtWidth(references []core.ManagedValueReference, terminalWidth int) string {
	headers := []string{"Group", "Parameter", "Condition", "Type", "Exposure", "Variants"}
	rows := make([][]string, 0, len(references))
	widths := shared.HeaderWidths(headers)
	styles := make([]managedFeatureTableRowStyle, 0, len(references))
	for _, reference := range references {
		row := []string{
			reference.Group,
			reference.Parameter,
			referenceValueLabel(reference),
			reference.ValueType,
			formatPercentage(reference.Percentage),
			formatManagedVariantValues(reference.Variants),
		}
		shared.UpdateTableWidths(widths, row)
		rows = append(rows, row)
		styles = append(styles, managedFeatureTableRowStyle{
			hasParameter: true, hasCondition: !reference.Default && strings.TrimSpace(reference.Condition) != "",
			conditionColor: reference.ConditionColor,
		})
	}
	fitFlexibleColumns(widths, terminalWidth, 5, 1, 2, 0)
	truncateHeaders(headers, widths)
	truncateColumns(rows, widths, 0, 1, 2, 5)
	return shared.StyledTable(headers, rows, widths, map[int]bool{4: true}, managedFeatureTableStyle(styles, 1, 2))
}

func formatManagedVariantValues(variants []core.ManagedVariantValue) string {
	if len(variants) == 0 {
		return "—"
	}
	values := make([]string, 0, len(variants))
	for _, variant := range variants {
		switch {
		case variant.NoChange != nil && *variant.NoChange:
			values = append(values, variant.ID+"=<no change>")
		case variant.Value != nil:
			values = append(values, variant.ID+"="+formatOptionalValue(variant.Value))
		default:
			values = append(values, variant.ID+"=—")
		}
	}
	return strings.Join(values, ", ")
}

func formatExperimentVariant(variant firebase.ExperimentVariant) string {
	name := emptyDash(variant.Name)
	if variant.Weight == 0 {
		return name
	}
	return fmt.Sprintf("%s (%d)", name, variant.Weight)
}

func experimentObjectiveName(objective firebase.ExperimentEventObjective) string {
	switch {
	case objective.CustomObjectiveDetails != nil:
		event := emptyDash(objective.CustomObjectiveDetails.Event)
		if objective.CustomObjectiveDetails.CountType != "" {
			return event + " · " + objective.CustomObjectiveDetails.CountType
		}
		return event
	case objective.SystemObjectiveDetails != nil:
		return emptyDash(objective.SystemObjectiveDetails.Objective)
	default:
		return "—"
	}
}

func referenceValueLabel(reference core.ManagedValueReference) string {
	if reference.Default {
		return "Default value"
	}
	return emptyDash(reference.Condition)
}

func formatPercentage(value *float64) string {
	if value == nil {
		return "—"
	}
	return strconv.FormatFloat(*value, 'f', -1, 64) + "%"
}

func formatOptionalValue(value *string) string {
	if value == nil {
		return "—"
	}
	if *value == "" {
		return `""`
	}
	return *value
}

func formatManagedFeatureTime(value string) string {
	value = strings.TrimSpace(value)
	if strings.HasPrefix(value, "1970-01-01T00:00:00") {
		return "—"
	}
	return emptyDash(value)
}

func formatManagedFeatureRelativeTime(value string, now time.Time) string {
	return rcdisplay.FormatRelativeTime(value, now)
}

func emptyDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "—"
	}
	return value
}

func fitFlexibleColumns(widths []int, terminalWidth int, flexible ...int) {
	if terminalWidth <= 0 {
		return
	}
	for shared.TableWidth(widths) > terminalWidth {
		selected := -1
		for _, index := range flexible {
			minimum := 1
			if widths[index] <= minimum {
				continue
			}
			if selected < 0 || widths[index]-minimum > widths[selected]-1 {
				selected = index
			}
		}
		if selected < 0 {
			return
		}
		widths[selected]--
	}
}

func truncateHeaders(headers []string, widths []int) {
	for index := range headers {
		headers[index] = ansi.Truncate(headers[index], widths[index], "…")
	}
}

func truncateColumns(rows [][]string, widths []int, columns ...int) {
	for row := range rows {
		for _, column := range columns {
			rows[row][column] = ansi.Truncate(rows[row][column], widths[column], "…")
		}
	}
}
