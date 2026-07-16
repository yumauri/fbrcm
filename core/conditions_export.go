package core

import "github.com/yumauri/fbrcm/core/conditions"

type (
	ConditionsTree        = conditions.Tree
	ConditionEntry        = conditions.Entry
	ConditionUsage        = conditions.Usage
	ConditionParameterRef = conditions.ParameterRef
	ConditionDeleteImpact = conditions.DeleteImpact
	ConditionMoveImpact   = conditions.MoveImpact
	ConditionDefinition   = conditions.Definition
	ConditionEdit         = conditions.Edit
	ConditionDetailsEdit  = conditions.DetailsEdit
)

var ConditionDisplayColors = conditions.DisplayColors

func NormalizeConditionName(name string) (string, error) {
	return conditions.NormalizeName(name)
}

func NormalizeConditionExpression(expression string) (string, error) {
	return conditions.NormalizeExpression(expression)
}

func NormalizeConditionTagColor(color string) (string, error) {
	return conditions.NormalizeTagColor(color)
}
