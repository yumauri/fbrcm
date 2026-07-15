package mutate

import "github.com/yumauri/fbrcm/core/firebase"

// DropUnknownConditionReferences removes conditional values whose condition names
// are not present in cfg.Conditions and drops parameters left without values.
// Groups are preserved even when those removals leave them empty.
func DropUnknownConditionReferences(cfg *firebase.RemoteConfig) {
	allowed := make(map[string]struct{}, len(cfg.Conditions))
	for _, condition := range cfg.Conditions {
		allowed[condition.Name] = struct{}{}
	}
	cfg.Parameters = stripUnknownConditionRefs(cfg.Parameters, allowed)
	for groupName, group := range cfg.ParameterGroups {
		group.Parameters = stripUnknownConditionRefs(group.Parameters, allowed)
		cfg.ParameterGroups[groupName] = group
	}
	NormalizeEmptyParameterMaps(cfg)
}

func stripUnknownConditionRefs(params map[string]firebase.RemoteConfigParam, allowed map[string]struct{}) map[string]firebase.RemoteConfigParam {
	if len(params) == 0 {
		return nil
	}
	out := make(map[string]firebase.RemoteConfigParam, len(params))
	for key, param := range params {
		if len(param.ConditionalValues) > 0 {
			filtered := make(map[string]firebase.RemoteConfigValue, len(param.ConditionalValues))
			for cond, value := range param.ConditionalValues {
				if _, ok := allowed[cond]; !ok {
					continue
				}
				filtered[cond] = value
			}
			if len(filtered) > 0 {
				param.ConditionalValues = filtered
			} else {
				param.ConditionalValues = nil
			}
		}
		if param.DefaultValue == nil && len(param.ConditionalValues) == 0 {
			continue
		}
		out[key] = param
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// NormalizeEmptyParameterMaps normalizes empty parameter maps without removing
// groups. Empty groups are valid Remote Config entities and may carry metadata.
func NormalizeEmptyParameterMaps(cfg *firebase.RemoteConfig) {
	if cfg == nil {
		return
	}
	for groupName, group := range cfg.ParameterGroups {
		if len(group.Parameters) == 0 {
			group.Parameters = nil
			cfg.ParameterGroups[groupName] = group
		}
	}
	if len(cfg.ParameterGroups) == 0 {
		cfg.ParameterGroups = nil
	}
	if len(cfg.Parameters) == 0 {
		cfg.Parameters = nil
	}
}
