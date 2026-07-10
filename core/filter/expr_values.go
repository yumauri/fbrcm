package filter

import (
	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/core/strfold"
	"strings"
)

func applyParameterScope(env expressionEnv, param firebase.RemoteConfigParam) expressionEnv {
	env.Default = defaultRemoteConfigValueForExpr(param.DefaultValue, param.ValueType)
	env.Value = anyRemoteConfigValuesForExpr(param)
	env.Conditionals = make(map[string]any, len(param.ConditionalValues))
	for name, value := range param.ConditionalValues {
		env.Conditionals[name] = remoteConfigValueForExpr(value, param.ValueType)
	}
	return env
}

func anyRemoteConfigValuesForExpr(param firebase.RemoteConfigParam) anyValue {
	out := anyValue{values: make([]any, 0, len(param.ConditionalValues)+1), valueType: strings.ToUpper(strings.TrimSpace(param.ValueType))}
	if param.DefaultValue != nil {
		out.values = append(out.values, defaultRemoteConfigValueForExpr(param.DefaultValue, param.ValueType))
	}
	for _, name := range strfold.SortedKeys(param.ConditionalValues) {
		out.values = append(out.values, remoteConfigValueForExpr(param.ConditionalValues[name], param.ValueType))
	}
	return out
}

func defaultRemoteConfigValueForExpr(value *firebase.RemoteConfigValue, valueType string) any {
	if value == nil {
		return nil
	}
	return remoteConfigValueForExpr(*value, valueType)
}

func remoteConfigValueForExpr(value firebase.RemoteConfigValue, valueType string) any {
	switch {
	case value.UseInAppDefault:
		return "<in-app default>"
	case len(value.PersonalizationValue) > 0:
		return "<personalization>"
	case len(value.RolloutValue) > 0:
		return "<rollout>"
	}
	raw := strings.TrimSpace(value.Value)
	switch strings.ToUpper(strings.TrimSpace(valueType)) {
	case "BOOLEAN":
		switch raw {
		case "true":
			return true
		case "false":
			return false
		}
	case "NUMBER":
		if number, ok := exprParseJSONNumber(raw); ok {
			return number
		}
	}
	return value.Value
}
