package shared

import (
	"bytes"
	"fmt"

	"github.com/yumauri/fbrcm/core/firebase"
)

// RemoteConfigOrder preserves the member order of an input Remote Config JSON file.
type RemoteConfigOrder struct {
	TopLevel          []string
	Parameters        []string
	Groups            []string
	GroupParameters   map[string][]string
	ConditionalValues map[string][]string
	VersionRaw        []byte
}

type objectEntry struct {
	key        string
	writeValue func()
}

type orderedJSONNode struct {
	kind    byte
	members []orderedJSONMember
	items   []*orderedJSONNode
	raw     []byte
}

type orderedJSONMember struct {
	key   string
	value *orderedJSONNode
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

// ParseRemoteConfigOrder parses the member order of a Remote Config JSON document.
func ParseRemoteConfigOrder(raw []byte) (RemoteConfigOrder, error) {
	body := bytes.TrimSpace(raw)
	root, next, ok := parseOrderedJSONValue(body, 0)
	if !ok || next != len(body) {
		return RemoteConfigOrder{}, fmt.Errorf("invalid json")
	}
	if root == nil || root.kind != '{' {
		return RemoteConfigOrder{}, fmt.Errorf("root json value is not object")
	}

	order := RemoteConfigOrder{
		GroupParameters:   make(map[string][]string),
		ConditionalValues: make(map[string][]string),
	}
	for _, member := range root.members {
		order.TopLevel = append(order.TopLevel, member.key)
		switch member.key {
		case "parameters":
			order.Parameters = objectMemberOrder(member.value)
			collectConditionalValueOrders(member.value, "", order.ConditionalValues)
		case "parameterGroups":
			order.Groups = objectMemberOrder(member.value)
			for _, groupMember := range member.value.members {
				for _, field := range groupMember.value.members {
					if field.key != "parameters" {
						continue
					}
					order.GroupParameters[groupMember.key] = objectMemberOrder(field.value)
					collectConditionalValueOrders(field.value, groupMember.key, order.ConditionalValues)
				}
			}
		case "version":
			order.VersionRaw = append([]byte(nil), member.value.raw...)
		}
	}
	return order, nil
}

func objectMemberOrder(node *orderedJSONNode) []string {
	if node == nil || node.kind != '{' {
		return nil
	}
	keys := make([]string, 0, len(node.members))
	for _, member := range node.members {
		keys = append(keys, member.key)
	}
	return keys
}

func collectConditionalValueOrders(node *orderedJSONNode, groupName string, out map[string][]string) {
	if node == nil || node.kind != '{' {
		return
	}
	for _, paramMember := range node.members {
		for _, field := range paramMember.value.members {
			if field.key != "conditionalValues" {
				continue
			}
			out[orderPath(groupName, paramMember.key)] = objectMemberOrder(field.value)
		}
	}
}

func orderPath(groupName, paramKey string) string {
	if groupName == "" {
		return paramKey
	}
	return groupName + "\x00" + paramKey
}
