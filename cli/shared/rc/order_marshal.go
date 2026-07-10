package rc

import (
	"bytes"
	"strings"

	"github.com/yumauri/fbrcm/core/firebase"
)

type objectEntry struct {
	key        string
	writeValue func()
}

// MarshalPrettyRemoteConfigWithOrder encodes Remote Config JSON while preserving known input order.
func MarshalPrettyRemoteConfigWithOrder(cfg *firebase.RemoteConfig, order RemoteConfigOrder) ([]byte, error) {
	if cfg == nil {
		return []byte("{}\n"), nil
	}

	var buf bytes.Buffer
	writeRemoteConfigObject(&buf, cfg, order, 0)
	buf.WriteByte('\n')
	return buf.Bytes(), nil
}

func writeRemoteConfigObject(buf *bytes.Buffer, cfg *firebase.RemoteConfig, order RemoteConfigOrder, indent int) {
	fields := map[string]objectEntry{
		"conditions": {
			key: "conditions",
			writeValue: func() {
				writeConditions(buf, cfg.Conditions, indent+1)
			},
		},
		"parameters": {
			key: "parameters",
			writeValue: func() {
				writeParametersMap(buf, cfg.Parameters, order.Parameters, order.ConditionalValues, "", indent+1)
			},
		},
		"parameterGroups": {
			key: "parameterGroups",
			writeValue: func() {
				writeGroups(buf, cfg.ParameterGroups, order, indent+1)
			},
		},
		"version": {
			key: "version",
			writeValue: func() {
				if len(order.VersionRaw) > 0 {
					buf.Write(order.VersionRaw)
					return
				}
				writeVersion(buf, cfg.Version, indent+1)
			},
		},
	}

	entries := make([]objectEntry, 0, 4)
	seen := make(map[string]struct{}, 4)
	for _, key := range order.TopLevel {
		if !remoteConfigFieldPresent(cfg, key) {
			continue
		}
		entry, ok := fields[key]
		if !ok {
			continue
		}
		entries = append(entries, entry)
		seen[key] = struct{}{}
	}
	for _, key := range []string{"conditions", "parameters", "parameterGroups", "version"} {
		if _, ok := seen[key]; ok || !remoteConfigFieldPresent(cfg, key) {
			continue
		}
		entries = append(entries, fields[key])
	}
	writeObject(buf, indent, entries)
}

func remoteConfigFieldPresent(cfg *firebase.RemoteConfig, key string) bool {
	switch key {
	case "conditions":
		return len(cfg.Conditions) > 0
	case "parameters":
		return len(cfg.Parameters) > 0
	case "parameterGroups":
		return len(cfg.ParameterGroups) > 0
	case "version":
		return strings.TrimSpace(cfg.Version.VersionNumber) != "" ||
			strings.TrimSpace(cfg.Version.UpdateTime) != "" ||
			strings.TrimSpace(cfg.Version.Description) != ""
	default:
		return false
	}
}

func writeObject(buf *bytes.Buffer, indent int, entries []objectEntry) {
	buf.WriteByte('{')
	if len(entries) == 0 {
		buf.WriteByte('}')
		return
	}
	for i, entry := range entries {
		buf.WriteByte('\n')
		writeIndent(buf, indent+1)
		writeJSONString(buf, entry.key)
		buf.WriteString(": ")
		entry.writeValue()
		if i < len(entries)-1 {
			buf.WriteByte(',')
		}
	}
	buf.WriteByte('\n')
	writeIndent(buf, indent)
	buf.WriteByte('}')
}
