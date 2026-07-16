package display

import "fmt"

func FormatConditionMoveImpact(crossedConditions, affectedParameters int) string {
	return fmt.Sprintf(
		"Priority impact: crosses %s and can change the winning value for %s.",
		FormatCount(crossedConditions, "condition", "conditions"),
		FormatCount(affectedParameters, "parameter", "parameters"),
	)
}

func FormatConditionDeleteImpact(conditionalValues, removedParameters int) string {
	return fmt.Sprintf(
		"Deletion impact: removes %s; %s will have no remaining value and will also be removed.",
		FormatCount(conditionalValues, "conditional value", "conditional values"),
		FormatCount(removedParameters, "parameter", "parameters"),
	)
}
