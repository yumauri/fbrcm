package promote

import (
	"github.com/yumauri/fbrcm/core/dictdiff"
	rcdiff "github.com/yumauri/fbrcm/core/rc/diff"
	rcdiffinput "github.com/yumauri/fbrcm/core/rc/diffinput"
	rcpromote "github.com/yumauri/fbrcm/core/rc/promote"
)

// DiffInput prepares the selected promotion entity for the generic dictionary
// diff viewer. Rendering and comparison remain independent of promotion types.
func (m Model) DiffInput(item rcpromote.Item) (dictdiff.Input, bool) {
	if m.plan == nil {
		return dictdiff.Input{}, false
	}
	input := dictdiff.Input{
		EntityName: item.ID.Name,
		Left: dictdiff.NamedDictionary{
			Name: "Current target: " + projectName(m.plan.Target.Project),
		},
		Right: dictdiff.NamedDictionary{
			Name: "Promotion source: " + projectName(m.plan.Source.Project),
		},
	}
	switch item.Kind {
	case rcdiff.ItemParameter:
		change := parameterChange(m.plan.Plan.Diff, item.ID)
		input.EntityName = rcdiffinput.ParameterEntityName(
			change.Group,
			change.Key,
		)
		input.Left.Properties = rcdiffinput.Parameter(parameterTargetGroup(change), change.Current)
		input.Right.Properties = rcdiffinput.Parameter(change.Group, change.Final)
	case rcdiff.ItemCondition:
		change := conditionChange(m.plan.Plan.Diff, item.ID.Name)
		input.EntityName = "Condition: " + change.Name
		input.Left.Properties = rcdiffinput.Condition(change.PreviousPosition, change.Current)
		input.Right.Properties = rcdiffinput.Condition(change.FinalPosition, change.Final)
	case rcdiff.ItemGroupDescription:
		change := groupChange(m.plan.Plan.Diff, item.ID.Name)
		input.EntityName = "Group: " + change.Group
		input.Left.Properties = rcdiffinput.Group(change.Current, change.Kind != rcdiff.ChangeAdded)
		input.Right.Properties = rcdiffinput.Group(change.Final, change.Kind != rcdiff.ChangeRemoved)
	default:
		return dictdiff.Input{}, false
	}
	return input, true
}

func parameterTargetGroup(change rcdiff.ParameterChange) string {
	if change.PreviousGroup != "" {
		return change.PreviousGroup
	}
	if change.Current != nil && change.PreviousKey == "" {
		return change.Group
	}
	return change.PreviousGroup
}
