package shared

import (
	"sort"
	"strings"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/filter"
	"github.com/yumauri/fbrcm/core/firebase"
)

// ParamTarget identifies a parameter and the group it belongs to.
type ParamTarget struct {
	Key   string
	Group string
	Param firebase.RemoteConfigParam
}

// GroupOrDefault returns the target group or defaultLabel for root parameters.
func (t ParamTarget) GroupOrDefault(defaultLabel string) string {
	if strings.TrimSpace(t.Group) == "" {
		return defaultLabel
	}
	return t.Group
}

// CollectParamTargets returns root and grouped parameters in stable order.
func CollectParamTargets(cfg *firebase.RemoteConfig) []ParamTarget {
	if cfg == nil {
		return nil
	}

	out := make([]ParamTarget, 0, len(cfg.Parameters)+len(cfg.ParameterGroups))
	for _, key := range SortedStringKeys(cfg.Parameters) {
		out = append(out, ParamTarget{Key: key, Param: cfg.Parameters[key]})
	}
	for _, groupName := range SortedStringKeys(cfg.ParameterGroups) {
		group := cfg.ParameterGroups[groupName]
		for _, key := range SortedStringKeys(group.Parameters) {
			out = append(out, ParamTarget{Key: key, Group: groupName, Param: group.Parameters[key]})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if !strings.EqualFold(out[i].Key, out[j].Key) {
			return strings.ToLower(out[i].Key) < strings.ToLower(out[j].Key)
		}
		return strings.ToLower(out[i].Group) < strings.ToLower(out[j].Group)
	})
	return out
}

// CollectMatchingParamTargets filters parameters by name filters, search, and expression.
func CollectMatchingParamTargets(project core.Project, cfg *firebase.RemoteConfig, rawFilters []string, search ParameterSearch, compiledExpr *filter.Expression, defaultGroupLabel string) []ParamTarget {
	all := CollectParamTargets(cfg)
	filters := ParseFilters(rawFilters)

	filtered := make([]ParamTarget, 0, len(all))
	for _, target := range all {
		if !MatchAnyFilter(target.Key, filters) {
			continue
		}
		if !MatchParameterSearch(target.Key, target.Param, cfg, search) {
			continue
		}
		match, ok := MatchParameterByCompiledExpr(compiledExpr, project, cfg, target.Key, target.GroupOrDefault(defaultGroupLabel))
		if !ok || !match {
			continue
		}
		filtered = append(filtered, target)
	}
	return filtered
}

// RemoveParamSlot removes a parameter from the root or a group.
func RemoveParamSlot(cfg *firebase.RemoteConfig, key, groupName string) {
	if groupName == "" {
		delete(cfg.Parameters, key)
		return
	}
	group, ok := cfg.ParameterGroups[groupName]
	if !ok {
		return
	}
	delete(group.Parameters, key)
	if len(group.Parameters) == 0 {
		delete(cfg.ParameterGroups, groupName)
		return
	}
	cfg.ParameterGroups[groupName] = group
}

// SetParamSlot writes a parameter to the root or a group, creating containers as needed.
func SetParamSlot(cfg *firebase.RemoteConfig, key, groupName string, param firebase.RemoteConfigParam) {
	if groupName == "" {
		if cfg.Parameters == nil {
			cfg.Parameters = map[string]firebase.RemoteConfigParam{}
		}
		cfg.Parameters[key] = param
		return
	}
	if cfg.ParameterGroups == nil {
		cfg.ParameterGroups = map[string]firebase.RemoteConfigGroup{}
	}
	group := cfg.ParameterGroups[groupName]
	if group.Parameters == nil {
		group.Parameters = map[string]firebase.RemoteConfigParam{}
	}
	group.Parameters[key] = param
	cfg.ParameterGroups[groupName] = group
}

// ParamExists reports whether a parameter exists in the root or any group.
func ParamExists(cfg *firebase.RemoteConfig, key string) bool {
	if cfg == nil {
		return false
	}
	if _, ok := cfg.Parameters[key]; ok {
		return true
	}
	for _, group := range cfg.ParameterGroups {
		if _, ok := group.Parameters[key]; ok {
			return true
		}
	}
	return false
}

// ParamSlotExists reports whether a root or grouped parameter exists.
func ParamSlotExists(cfg *firebase.RemoteConfig, key, groupName string) bool {
	if cfg == nil {
		return false
	}
	if groupName == "" {
		_, ok := cfg.Parameters[key]
		return ok
	}
	group, ok := cfg.ParameterGroups[groupName]
	if !ok {
		return false
	}
	_, ok = group.Parameters[key]
	return ok
}
