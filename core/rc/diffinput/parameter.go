// Package diffinput adapts Remote Config entities to generic dictionary diffs.
package diffinput

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/yumauri/fbrcm/core/dictdiff"
	"github.com/yumauri/fbrcm/core/firebase"
	rcdiff "github.com/yumauri/fbrcm/core/rc/diff"
)

// ParameterEntityName formats the entity name used by generic dictionary diff
// renderers.
func ParameterEntityName(group, key string) string {
	if strings.TrimSpace(group) == "" {
		return "Parameter: " + key
	}
	return "Parameter: " + group + " / " + key
}

// Parameter prepares one Remote Config parameter as a generic dictionary.
func Parameter(group string, parameter *firebase.RemoteConfigParam) dictdiff.Dictionary {
	if parameter == nil {
		return dictdiff.Dictionary{}
	}
	properties := dictdiff.Dictionary{
		"type":        dictdiff.Enum(parameter.ValueType),
		"description": dictdiff.String(parameter.Description),
		"group":       dictdiff.Enum(group),
	}
	for condition, value := range parameter.ConditionalValues {
		properties["value · "+condition] = Value(value, parameter.ValueType)
	}
	if parameter.DefaultValue != nil {
		properties["value · default"] = Value(*parameter.DefaultValue, parameter.ValueType)
	}
	return properties
}

// ParameterChange prepares one Remote Config parameter change as a generic
// dictionary diff. Names describe the left and right versions when provided.
func ParameterChange(change rcdiff.ParameterChange, leftName, rightName string) dictdiff.Input {
	leftGroup := change.Group
	leftKey := change.Key
	if change.PreviousKey != "" {
		leftGroup = change.PreviousGroup
		leftKey = change.PreviousKey
	}
	displayGroup := change.Group
	if change.Final == nil {
		displayGroup = leftGroup
	}
	leftProperties := Parameter(leftGroup, change.Current)
	rightProperties := Parameter(change.Group, change.Final)
	if leftKey != change.Key {
		leftProperties["name"] = dictdiff.Enum(leftKey)
		rightProperties["name"] = dictdiff.Enum(change.Key)
	}
	return dictdiff.Input{
		EntityName: ParameterEntityName(displayGroup, change.Key),
		Left: dictdiff.NamedDictionary{
			Name:       leftName,
			Properties: leftProperties,
		},
		Right: dictdiff.NamedDictionary{
			Name:       rightName,
			Properties: rightProperties,
		},
	}
}

// Value preserves Remote Config value semantics while choosing a generic
// comparison hint.
func Value(value firebase.RemoteConfigValue, valueType string) dictdiff.Value {
	switch {
	case value.UseInAppDefault:
		return dictdiff.Enum("in-app default")
	case len(value.PersonalizationValue) > 0:
		return dictdiff.JSON(string(value.PersonalizationValue))
	case len(value.ExperimentValue) > 0:
		return dictdiff.JSON(string(value.ExperimentValue))
	case len(value.RolloutValue) > 0:
		return dictdiff.JSON(string(value.RolloutValue))
	case value.UnknownValueOption != "":
		raw, err := json.Marshal(map[string]json.RawMessage{value.UnknownValueOption: value.UnknownValue})
		if err == nil {
			return dictdiff.JSON(string(raw))
		}
		return dictdiff.String(value.UnknownValueOption)
	}
	switch strings.ToLower(strings.TrimSpace(valueType)) {
	case "boolean":
		parsed, err := strconv.ParseBool(value.Value)
		if err == nil {
			return dictdiff.Boolean(parsed)
		}
		return dictdiff.Enum(value.Value)
	case "number":
		return dictdiff.Number(json.Number(value.Value))
	case "json":
		return dictdiff.JSON(value.Value)
	}
	return dictdiff.String(value.Value)
}
