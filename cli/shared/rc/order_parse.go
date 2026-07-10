package rc

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/yumauri/fbrcm/cli/internal/jsonscan"
)

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

func parseOrderedJSONValue(body []byte, start int) (*orderedJSONNode, int, bool) {
	start = skipJSONWhitespace(body, start)
	if start >= len(body) {
		return nil, 0, false
	}
	switch body[start] {
	case '{':
		return parseOrderedJSONObject(body, start)
	case '[':
		return parseOrderedJSONArray(body, start)
	case '"':
		end, ok := jsonscan.ScanStringEnd(body, start)
		if !ok {
			return nil, 0, false
		}
		return &orderedJSONNode{kind: '"', raw: append([]byte(nil), body[start:end+1]...)}, end + 1, true
	default:
		end := scanPrimitiveEnd(body, start)
		return &orderedJSONNode{kind: 'v', raw: append([]byte(nil), body[start:end]...)}, end, true
	}
}

func parseOrderedJSONObject(body []byte, start int) (*orderedJSONNode, int, bool) {
	if start >= len(body) || body[start] != '{' {
		return nil, 0, false
	}
	node := &orderedJSONNode{kind: '{'}
	pos := skipJSONWhitespace(body, start+1)
	if pos < len(body) && body[pos] == '}' {
		node.raw = append([]byte(nil), body[start:pos+1]...)
		return node, pos + 1, true
	}
	for {
		keyStart := skipJSONWhitespace(body, pos)
		keyEnd, ok := jsonscan.ScanStringEnd(body, keyStart)
		if !ok {
			return nil, 0, false
		}
		key, err := unquoteJSONString(body[keyStart : keyEnd+1])
		if err != nil {
			return nil, 0, false
		}
		colon := skipJSONWhitespace(body, keyEnd+1)
		if colon >= len(body) || body[colon] != ':' {
			return nil, 0, false
		}
		value, next, ok := parseOrderedJSONValue(body, colon+1)
		if !ok {
			return nil, 0, false
		}
		node.members = append(node.members, orderedJSONMember{key: key, value: value})
		pos = skipJSONWhitespace(body, next)
		if pos >= len(body) {
			return nil, 0, false
		}
		if body[pos] == '}' {
			node.raw = append([]byte(nil), body[start:pos+1]...)
			return node, pos + 1, true
		}
		if body[pos] != ',' {
			return nil, 0, false
		}
		pos++
	}
}

func parseOrderedJSONArray(body []byte, start int) (*orderedJSONNode, int, bool) {
	if start >= len(body) || body[start] != '[' {
		return nil, 0, false
	}
	node := &orderedJSONNode{kind: '['}
	pos := skipJSONWhitespace(body, start+1)
	if pos < len(body) && body[pos] == ']' {
		node.raw = append([]byte(nil), body[start:pos+1]...)
		return node, pos + 1, true
	}
	for {
		value, next, ok := parseOrderedJSONValue(body, pos)
		if !ok {
			return nil, 0, false
		}
		node.items = append(node.items, value)
		pos = skipJSONWhitespace(body, next)
		if pos >= len(body) {
			return nil, 0, false
		}
		if body[pos] == ']' {
			node.raw = append([]byte(nil), body[start:pos+1]...)
			return node, pos + 1, true
		}
		if body[pos] != ',' {
			return nil, 0, false
		}
		pos++
	}
}

func skipJSONWhitespace(body []byte, pos int) int {
	for pos < len(body) {
		switch body[pos] {
		case ' ', '\n', '\r', '\t':
			pos++
		default:
			return pos
		}
	}
	return pos
}

func scanPrimitiveEnd(body []byte, start int) int {
	for i := start; i < len(body); i++ {
		switch body[i] {
		case ',', '}', ']', ' ', '\n', '\r', '\t':
			return i
		}
	}
	return len(body)
}

func unquoteJSONString(raw []byte) (string, error) {
	var out string
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", err
	}
	return out, nil
}
