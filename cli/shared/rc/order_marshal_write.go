package rc

import (
	"bytes"
	"encoding/json"

	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/core/strfold"
)

func writeConditions(buf *bytes.Buffer, conditions []firebase.RemoteConfigCondition, indent int) {
	buf.WriteByte('[')
	if len(conditions) == 0 {
		buf.WriteByte(']')
		return
	}
	for i, condition := range conditions {
		buf.WriteByte('\n')
		writeIndent(buf, indent+1)
		writeCondition(buf, condition, indent+1)
		if i < len(conditions)-1 {
			buf.WriteByte(',')
		}
	}
	buf.WriteByte('\n')
	writeIndent(buf, indent)
	buf.WriteByte(']')
}

func writeCondition(buf *bytes.Buffer, condition firebase.RemoteConfigCondition, indent int) {
	entries := make([]objectEntry, 0, 3)
	if condition.Name != "" {
		value := condition.Name
		entries = append(entries, objectEntry{key: "name", writeValue: func() { writeJSONString(buf, value) }})
	}
	if condition.Expression != "" {
		value := condition.Expression
		entries = append(entries, objectEntry{key: "expression", writeValue: func() { writeJSONString(buf, value) }})
	}
	if condition.TagColor != "" {
		value := condition.TagColor
		entries = append(entries, objectEntry{key: "tagColor", writeValue: func() { writeJSONString(buf, value) }})
	}
	writeObject(buf, indent, entries)
}

func writeGroups(buf *bytes.Buffer, groups map[string]firebase.RemoteConfigGroup, order RemoteConfigOrder, indent int) {
	keys := orderedKeys(groups, order.Groups)
	entries := make([]objectEntry, 0, len(keys))
	for _, key := range keys {
		groupName := key
		group := groups[key]
		entries = append(entries, objectEntry{
			key: groupName,
			writeValue: func() {
				writeGroup(buf, groupName, group, order, indent+1)
			},
		})
	}
	writeObject(buf, indent, entries)
}

func writeGroup(buf *bytes.Buffer, groupName string, group firebase.RemoteConfigGroup, order RemoteConfigOrder, indent int) {
	entries := make([]objectEntry, 0, 2)
	if group.Description != "" {
		value := group.Description
		entries = append(entries, objectEntry{key: "description", writeValue: func() { writeJSONString(buf, value) }})
	}
	if len(group.Parameters) > 0 {
		params := group.Parameters
		paramOrder := order.GroupParameters[groupName]
		entries = append(entries, objectEntry{
			key: "parameters",
			writeValue: func() {
				writeParametersMap(buf, params, paramOrder, order.ConditionalValues, groupName, indent+1)
			},
		})
	}
	writeObject(buf, indent, entries)
}

func writeParametersMap(buf *bytes.Buffer, params map[string]firebase.RemoteConfigParam, order []string, conditionalOrders map[string][]string, groupName string, indent int) {
	keys := orderedKeys(params, order)
	entries := make([]objectEntry, 0, len(keys))
	for _, key := range keys {
		paramKey := key
		param := params[key]
		condOrder := conditionalOrders[orderPath(groupName, paramKey)]
		entries = append(entries, objectEntry{
			key: paramKey,
			writeValue: func() {
				writeParam(buf, param, condOrder, indent+1)
			},
		})
	}
	writeObject(buf, indent, entries)
}

func writeParam(buf *bytes.Buffer, param firebase.RemoteConfigParam, conditionalOrder []string, indent int) {
	entries := make([]objectEntry, 0, 4)
	if param.DefaultValue != nil {
		value := *param.DefaultValue
		entries = append(entries, objectEntry{
			key: "defaultValue",
			writeValue: func() {
				writeRemoteConfigValue(buf, value, indent+1)
			},
		})
	}
	if len(param.ConditionalValues) > 0 {
		values := param.ConditionalValues
		entries = append(entries, objectEntry{
			key: "conditionalValues",
			writeValue: func() {
				writeConditionalValues(buf, values, conditionalOrder, indent+1)
			},
		})
	}
	if param.Description != "" {
		value := param.Description
		entries = append(entries, objectEntry{key: "description", writeValue: func() { writeJSONString(buf, value) }})
	}
	if param.ValueType != "" {
		value := param.ValueType
		entries = append(entries, objectEntry{key: "valueType", writeValue: func() { writeJSONString(buf, value) }})
	}
	writeObject(buf, indent, entries)
}

func writeConditionalValues(buf *bytes.Buffer, values map[string]firebase.RemoteConfigValue, order []string, indent int) {
	keys := orderedKeys(values, order)
	entries := make([]objectEntry, 0, len(keys))
	for _, key := range keys {
		condition := key
		value := values[key]
		entries = append(entries, objectEntry{
			key: condition,
			writeValue: func() {
				writeRemoteConfigValue(buf, value, indent+1)
			},
		})
	}
	writeObject(buf, indent, entries)
}

func writeRemoteConfigValue(buf *bytes.Buffer, value firebase.RemoteConfigValue, indent int) {
	entries := make([]objectEntry, 0, 4)
	if value.Value != "" || (!value.UseInAppDefault && len(value.PersonalizationValue) == 0 && len(value.RolloutValue) == 0) {
		raw := value.Value
		entries = append(entries, objectEntry{key: "value", writeValue: func() { writeJSONString(buf, raw) }})
	}
	if value.UseInAppDefault {
		entries = append(entries, objectEntry{key: "useInAppDefault", writeValue: func() { buf.WriteString("true") }})
	}
	if len(value.PersonalizationValue) > 0 {
		raw := append([]byte(nil), value.PersonalizationValue...)
		entries = append(entries, objectEntry{key: "personalizationValue", writeValue: func() { buf.Write(NormalizeJSONEscapes(bytes.TrimSpace(raw))) }})
	}
	if len(value.RolloutValue) > 0 {
		raw := append([]byte(nil), value.RolloutValue...)
		entries = append(entries, objectEntry{key: "rolloutValue", writeValue: func() { buf.Write(NormalizeJSONEscapes(bytes.TrimSpace(raw))) }})
	}
	writeObject(buf, indent, entries)
}

func writeVersion(buf *bytes.Buffer, version firebase.RemoteConfigVersion, indent int) {
	entries := make([]objectEntry, 0, 3)
	if version.VersionNumber != "" {
		value := version.VersionNumber
		entries = append(entries, objectEntry{key: "versionNumber", writeValue: func() { writeJSONString(buf, value) }})
	}
	if version.UpdateTime != "" {
		value := version.UpdateTime
		entries = append(entries, objectEntry{key: "updateTime", writeValue: func() { writeJSONString(buf, value) }})
	}
	if version.Description != "" {
		value := version.Description
		entries = append(entries, objectEntry{key: "description", writeValue: func() { writeJSONString(buf, value) }})
	}
	writeObject(buf, indent, entries)
}

func writeIndent(buf *bytes.Buffer, indent int) {
	for range indent {
		buf.WriteString("  ")
	}
}

func writeJSONString(buf *bytes.Buffer, value string) {
	encoded, _ := json.Marshal(value)
	buf.Write(NormalizeJSONEscapes(encoded))
}

func orderedKeys[T any](items map[string]T, preferred []string) []string {
	keys := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, key := range preferred {
		if _, ok := items[key]; !ok {
			continue
		}
		keys = append(keys, key)
		seen[key] = struct{}{}
	}
	rest := make([]string, 0, len(items)-len(keys))
	for key := range items {
		if _, ok := seen[key]; ok {
			continue
		}
		rest = append(rest, key)
	}
	strfold.Sort(rest)
	return append(keys, rest...)
}
