package mutate

import "github.com/yumauri/fbrcm/core/firebase"

// Slot is one parameter in one group.
type Slot struct {
	Group string
	Param firebase.RemoteConfigParam
}

// SlotKey returns the composite slot identity group\x00paramKey.
func SlotKey(group, paramKey string) string {
	return group + "\x00" + paramKey
}

// SlotKeyParam extracts the parameter key from a composite slot key.
func SlotKeyParam(key string) string {
	for i := 0; i < len(key); i++ {
		if key[i] == 0 {
			return key[i+1:]
		}
	}
	return key
}

// SlotKeyGroup extracts the group name from a composite slot key.
func SlotKeyGroup(key string) string {
	for i := 0; i < len(key); i++ {
		if key[i] == 0 {
			return key[:i]
		}
	}
	return ""
}

// SlotDisplayKey formats a composite slot key for error messages.
func SlotDisplayKey(key string) string {
	group := SlotKeyGroup(key)
	param := SlotKeyParam(key)
	if group == "" {
		return param
	}
	return group + "/" + param
}

// CollectParamSlots enumerates all parameter slots keyed by SlotKey(group, paramKey).
func CollectParamSlots(cfg *firebase.RemoteConfig) map[string]Slot {
	out := make(map[string]Slot)
	if cfg == nil {
		return out
	}
	for key, param := range cfg.Parameters {
		out[SlotKey("", key)] = Slot{Group: "", Param: param}
	}
	for groupName, group := range cfg.ParameterGroups {
		for key, param := range group.Parameters {
			out[SlotKey(groupName, key)] = Slot{Group: groupName, Param: param}
		}
	}
	return out
}

// SetParamSlot writes a slot back, initializing Parameters and ParameterGroups maps when nil.
func SetParamSlot(cfg *firebase.RemoteConfig, paramKey string, slot Slot) {
	if slot.Group == "" {
		if cfg.Parameters == nil {
			cfg.Parameters = map[string]firebase.RemoteConfigParam{}
		}
		cfg.Parameters[paramKey] = slot.Param
		return
	}

	if cfg.ParameterGroups == nil {
		cfg.ParameterGroups = map[string]firebase.RemoteConfigGroup{}
	}
	group := cfg.ParameterGroups[slot.Group]
	if group.Parameters == nil {
		group.Parameters = map[string]firebase.RemoteConfigParam{}
	}
	group.Parameters[paramKey] = slot.Param
	cfg.ParameterGroups[slot.Group] = group
}

// RemoveParamSlot deletes one parameter from root or a named group.
func RemoveParamSlot(cfg *firebase.RemoteConfig, paramKey, groupName string) {
	if groupName == "" {
		delete(cfg.Parameters, paramKey)
		return
	}
	group, ok := cfg.ParameterGroups[groupName]
	if !ok {
		return
	}
	delete(group.Parameters, paramKey)
	if len(group.Parameters) == 0 {
		delete(cfg.ParameterGroups, groupName)
		return
	}
	cfg.ParameterGroups[groupName] = group
}
