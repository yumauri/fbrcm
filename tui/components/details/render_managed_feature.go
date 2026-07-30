package details

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
	"github.com/yumauri/fbrcm/tui/components/viewutil"
	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/styles"
)

func (m Model) renderManagedFeatureContentLines(width int) []string {
	data := m.managedFeatureData
	lines := make([]string, 0, 48)
	lines = appendStyledField(lines, width, "Project", rcdisplay.FormatProject(data.Project.Name, data.Project.ProjectID), projectValueStyle)
	switch data.Kind {
	case messages.ManagedFeatureExperiment:
		lines = appendExperimentDetails(lines, width, data.Template, data.Experiment)
	case messages.ManagedFeaturePersonalization:
		lines = appendPersonalizationDetails(lines, width, data.Template, data.Personalization)
	case messages.ManagedFeatureRollout:
		lines = appendRolloutDetails(lines, width, data.Template, data.Rollout)
	}
	return padLines(lines, width)
}

func appendExperimentDetails(
	lines []string,
	width int,
	template core.ManagedFeatureTemplate,
	experiment *core.ExperimentEntry,
) []string {
	if experiment == nil {
		return appendStyledField(lines, width, "A/B Test", "Unavailable", styles.SecondaryTitleError)
	}
	lines = appendStyledField(lines, width, "Type", "A/B Test", styles.PanelText)
	lines = appendStyledField(lines, width, "ID", firebase.ManagedFeatureID(experiment.Name), styles.PanelText)
	lines = appendStyledField(lines, width, "Name", managedFeatureValue(experiment.Definition.DisplayName), styles.PanelText)
	lines = appendStyledField(
		lines,
		width,
		"State",
		managedFeatureValue(experiment.State),
		styles.ManagedFeatureStatusStyle(experiment.State),
	)
	lines = appendStyledField(lines, width, "Service", managedFeatureValue(experiment.Definition.Service), styles.PanelText)
	lines = appendStyledField(lines, width, "Description", managedFeatureValue(experiment.Definition.Description), styles.PanelText)
	lines = appendStyledField(lines, width, "Started", managedFeatureTime(experiment.StartTime), styles.PanelText)
	lines = appendStyledField(lines, width, "Ended", managedFeatureTime(experiment.EndTime), styles.PanelText)
	lines = appendStyledField(lines, width, "Updated", managedFeatureTime(experiment.LastUpdateTime), styles.PanelText)
	lines = appendStyledField(lines, width, "Activation event", managedFeatureValue(experiment.Definition.Objectives.ActivationEvent.Event), styles.PanelText)
	lines = appendStyledField(lines, width, "ETag", managedFeatureValue(experiment.ETag), styles.PanelText)
	lines = appendTemplateDetails(lines, width, template)

	lines = append(lines, fieldTitle("Variants", false, false))
	if len(experiment.Definition.Variants) == 0 {
		lines = append(lines, styles.PanelMuted.Italic(true).Render("No variants exposed."), "")
	} else {
		for _, variant := range experiment.Definition.Variants {
			lines = append(lines, "  "+formatManagedVariant(variant))
		}
		lines = append(lines, "")
	}

	lines = append(lines, fieldTitle("Objectives", false, false))
	if len(experiment.Definition.Objectives.EventObjectives) == 0 {
		lines = append(lines, styles.PanelMuted.Italic(true).Render("No objectives exposed."))
	} else {
		for _, objective := range experiment.Definition.Objectives.EventObjectives {
			label := experimentObjectiveLabel(objective)
			if objective.IsPrimary {
				label += " · primary"
			}
			if objective.ABTOptimizationFunction != "" {
				label += " · " + objective.ABTOptimizationFunction
			}
			lines = append(lines, "  "+label)
		}
	}
	return appendExperimentBindings(lines, experiment.References)
}

func appendPersonalizationDetails(
	lines []string,
	width int,
	template core.ManagedFeatureTemplate,
	personalization *core.PersonalizationEntry,
) []string {
	if personalization == nil {
		return appendStyledField(lines, width, "Personalization", "Unavailable", styles.SecondaryTitleError)
	}
	lines = appendStyledField(lines, width, "Type", "Personalization", styles.PanelText)
	lines = appendStyledField(lines, width, "ID", personalization.ID, styles.PanelText)
	lines = appendTemplateDetails(lines, width, template)
	lines = appendBindings(lines, personalization.References, false)
	lines = append(lines, "", styles.PanelMuted.Italic(true).Render(
		"Firebase does not expose personalization value candidates or result metrics through this API.",
	))
	return lines
}

func appendRolloutDetails(
	lines []string,
	width int,
	template core.ManagedFeatureTemplate,
	rollout *core.RolloutEntry,
) []string {
	if rollout == nil {
		return appendStyledField(lines, width, "Rollout", "Unavailable", styles.SecondaryTitleError)
	}
	lines = appendStyledField(lines, width, "Type", "Rollout", styles.PanelText)
	lines = appendStyledField(lines, width, "ID", firebase.ManagedFeatureID(rollout.Name), styles.PanelText)
	lines = appendStyledField(lines, width, "Name", managedFeatureValue(rollout.Definition.DisplayName), styles.PanelText)
	lines = appendStyledField(
		lines,
		width,
		"State",
		managedFeatureValue(rollout.State),
		styles.ManagedFeatureStatusStyle(rollout.State),
	)
	lines = appendStyledField(lines, width, "Description", managedFeatureValue(rollout.Definition.Description), styles.PanelText)
	lines = appendStyledField(lines, width, "Created", managedFeatureTime(rollout.CreateTime), styles.PanelText)
	lines = appendStyledField(lines, width, "Started", managedFeatureTime(rollout.StartTime), styles.PanelText)
	lines = appendStyledField(lines, width, "Ended", managedFeatureTime(rollout.EndTime), styles.PanelText)
	lines = appendStyledField(lines, width, "Updated", managedFeatureTime(rollout.LastUpdateTime), styles.PanelText)
	lines = appendStyledField(lines, width, "Control variant", formatManagedVariant(rollout.Definition.ControlVariant), styles.PanelText)
	lines = appendStyledField(lines, width, "Enabled variant", formatManagedVariant(rollout.Definition.EnabledVariant), styles.PanelText)
	lines = appendStyledField(lines, width, "ETag", managedFeatureValue(rollout.ETag), styles.PanelText)
	lines = appendTemplateDetails(lines, width, template)
	return appendBindings(lines, rollout.References, true)
}

func appendTemplateDetails(lines []string, width int, template core.ManagedFeatureTemplate) []string {
	lines = appendStyledField(lines, width, "Remote Config version", managedFeatureValue(template.Version), styles.PanelText)
	return appendStyledField(lines, width, "Remote Config source", managedFeatureValue(template.Source), styles.PanelText)
}

func appendBindings(lines []string, references []core.ManagedValueReference, rollout bool) []string {
	lines = append(lines, fieldTitle(
		"Bindings · "+rcdisplay.FormatCount(len(references), "parameter value", "parameter values"),
		false,
		false,
	))
	if len(references) == 0 {
		return append(lines, styles.PanelMuted.Italic(true).Render("No template bindings found."))
	}
	for _, reference := range references {
		label := viewutil.RenderManagedReferenceIdentity(reference)
		if rollout {
			label += styles.ParameterSeparator.Render(" · ") +
				styles.PanelText.Render(managedFeatureOptionalValue(reference.Value))
			if reference.Percentage != nil {
				label += styles.ParameterSeparator.Render(" · ") +
					styles.SecondaryTitleCount.Render(strconv.FormatFloat(*reference.Percentage, 'f', -1, 64)+"%")
			}
		}
		lines = append(lines, "  "+label)
	}
	return lines
}

func appendExperimentBindings(lines []string, references []core.ManagedValueReference) []string {
	lines = append(lines, fieldTitle(
		"Bindings · "+rcdisplay.FormatCount(len(references), "parameter value", "parameter values"),
		false,
		false,
	))
	if len(references) == 0 {
		return append(lines, styles.PanelMuted.Italic(true).Render("No template bindings found."))
	}
	for _, reference := range references {
		label := viewutil.RenderManagedReferenceIdentity(reference)
		if reference.Percentage != nil {
			label += styles.ParameterSeparator.Render(" · exposure ") +
				styles.SecondaryTitleCount.Render(strconv.FormatFloat(*reference.Percentage, 'f', -1, 64)+"%")
		}
		lines = append(lines, "  "+label)
		for _, variant := range reference.Variants {
			lines = append(lines, "    "+formatManagedVariantValue(variant))
		}
	}
	return lines
}

func formatManagedVariantValue(variant core.ManagedVariantValue) string {
	switch {
	case variant.NoChange != nil && *variant.NoChange:
		return variant.ID + " = <no change>"
	case variant.Value != nil:
		return variant.ID + " = " + managedFeatureOptionalValue(variant.Value)
	default:
		return variant.ID + " = —"
	}
}

func managedFeatureOptionalValue(value *string) string {
	if value == nil {
		return "—"
	}
	if *value == "" {
		return `""`
	}
	return *value
}

func formatManagedVariant(variant firebase.ExperimentVariant) string {
	name := managedFeatureValue(variant.Name)
	if variant.Weight == 0 {
		return name
	}
	return fmt.Sprintf("%s (%d)", name, variant.Weight)
}

func experimentObjectiveLabel(objective firebase.ExperimentEventObjective) string {
	switch {
	case objective.CustomObjectiveDetails != nil:
		label := managedFeatureValue(objective.CustomObjectiveDetails.Event)
		if objective.CustomObjectiveDetails.CountType != "" {
			label += " · " + objective.CustomObjectiveDetails.CountType
		}
		return label
	case objective.SystemObjectiveDetails != nil:
		return managedFeatureValue(objective.SystemObjectiveDetails.Objective)
	default:
		return "—"
	}
}

func managedFeatureValue(value string) string {
	if strings.TrimSpace(value) == "" {
		return "—"
	}
	return value
}

func managedFeatureTime(value string) string {
	value = strings.TrimSpace(value)
	timestamp, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return managedFeatureValue(value)
	}
	if timestamp.Equal(time.Unix(0, 0)) {
		return "—"
	}
	return rcdisplay.FormatLocalDateTime(timestamp)
}
