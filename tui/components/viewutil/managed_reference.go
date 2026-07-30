package viewutil

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/tui/styles"
)

// ManagedReferencePathText returns the unstyled group and parameter identity.
func ManagedReferencePathText(reference core.ManagedValueReference) string {
	group := managedReferenceText(reference.Group)
	parameter := managedReferenceText(reference.Parameter)
	switch {
	case group == "":
		return parameter
	case parameter == "":
		return group
	default:
		return group + " / " + parameter
	}
}

// RenderManagedReferencePath renders group and parameter names with their
// standard Remote Config identity styles.
func RenderManagedReferencePath(reference core.ManagedValueReference) string {
	group := managedReferenceText(reference.Group)
	parameter := managedReferenceText(reference.Parameter)
	switch {
	case group == "":
		return renderManagedReferenceStyle(styles.ParameterName, parameter)
	case parameter == "":
		return renderManagedReferenceStyle(styles.ParameterGroup, group)
	default:
		return renderManagedReferenceStyle(styles.ParameterGroup, group) +
			renderManagedReferenceStyle(styles.ParameterSeparator, " / ") +
			renderManagedReferenceStyle(styles.ParameterName, parameter)
	}
}

// RenderManagedReferenceCondition renders a binding condition in its configured
// Remote Config color. Default values retain the standard empty-value style.
func RenderManagedReferenceCondition(reference core.ManagedValueReference) string {
	if reference.Default {
		return renderManagedReferenceStyle(styles.DetailsEmptyValue, "Default value")
	}
	condition := managedReferenceText(reference.Condition)
	return renderManagedReferenceStyle(
		styles.DetailsConditionValueStyle(reference.ConditionColor),
		condition,
	)
}

// RenderManagedReferenceIdentity renders the group, parameter, condition, and
// value type portions shared by managed-feature summaries and Details.
func RenderManagedReferenceIdentity(reference core.ManagedValueReference) string {
	return joinManagedReferenceMeta(
		RenderManagedReferencePath(reference),
		RenderManagedReferenceCondition(reference),
		renderManagedReferenceStyle(styles.PanelMuted, managedReferenceText(reference.ValueType)),
	)
}

func joinManagedReferenceMeta(values ...string) string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(ansi.Strip(value)) != "" {
			out = append(out, value)
		}
	}
	return strings.Join(out, renderManagedReferenceStyle(styles.ParameterSeparator, " · "))
}

func renderManagedReferenceStyle(style lipgloss.Style, value string) string {
	if value == "" || styles.NoColorEnabled() {
		return value
	}
	return style.Render(value)
}

func managedReferenceText(value string) string {
	return strings.Join(strings.Fields(value), " ")
}
