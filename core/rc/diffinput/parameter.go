// Package diffinput adapts Remote Config entities to generic dictionary diffs.
package diffinput

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/yumauri/fbrcm/core/dictdiff"
	"github.com/yumauri/fbrcm/core/firebase"
)

// ParameterEntityName formats the entity name used by generic dictionary diff
// renderers.
func ParameterEntityName(group, key string) string {
	if strings.TrimSpace(group) == "" {
		return "Property: " + key
	}
	return "Property: " + group + " / " + key
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

// Value preserves Remote Config value semantics while choosing a generic
// comparison hint.
func Value(value firebase.RemoteConfigValue, valueType string) dictdiff.Value {
	switch {
	case value.UseInAppDefault:
		return dictdiff.Enum("in-app default")
	case len(value.PersonalizationValue) > 0:
		return dictdiff.JSON(string(value.PersonalizationValue))
	case len(value.RolloutValue) > 0:
		return dictdiff.JSON(string(value.RolloutValue))
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
