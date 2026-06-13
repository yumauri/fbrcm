package shared

import "encoding/json"

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
		end, ok := scanJSONStringEnd(body, start)
		if !ok {
			return nil, 0, false
		}
		return &orderedJSONNode{kind: '"', raw: append([]byte(nil), body[start:end+1]...)}, end + 1, true
	default:
		end, ok := scanPrimitiveEnd(body, start)
		if !ok {
			return nil, 0, false
		}
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
		keyEnd, ok := scanJSONStringEnd(body, keyStart)
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

func scanJSONStringEnd(body []byte, start int) (int, bool) {
	if start >= len(body) || body[start] != '"' {
		return 0, false
	}
	escaped := false
	for i := start + 1; i < len(body); i++ {
		switch {
		case escaped:
			escaped = false
		case body[i] == '\\':
			escaped = true
		case body[i] == '"':
			return i, true
		}
	}
	return 0, false
}

func scanPrimitiveEnd(body []byte, start int) (int, bool) {
	for i := start; i < len(body); i++ {
		switch body[i] {
		case ',', '}', ']', ' ', '\n', '\r', '\t':
			return i, true
		}
	}
	return len(body), true
}

func unquoteJSONString(raw []byte) (string, error) {
	var out string
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", err
	}
	return out, nil
}
