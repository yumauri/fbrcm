package project

import (
	"bytes"
	"strconv"
	"unicode"
)

func normalizeExportJSON(body []byte) []byte {
	body = bytes.ReplaceAll(body, []byte(`\u003c`), []byte("<"))
	body = bytes.ReplaceAll(body, []byte(`\u003e`), []byte(">"))
	body = bytes.ReplaceAll(body, []byte(`\u0026`), []byte("&"))
	body = reorderConditionalValuesKeysNumericFirst(body)
	return body
}

func trimTrailingLineBreaks(body []byte) []byte {
	return bytes.TrimRight(body, "\r\n")
}

type jsonObjectMember struct {
	key string
	raw []byte
}

func reorderConditionalValuesKeysNumericFirst(body []byte) []byte {
	const needle = `"conditionalValues"`

	var out bytes.Buffer
	cursor := 0
	for {
		rel := bytes.Index(body[cursor:], []byte(needle))
		if rel < 0 {
			out.Write(body[cursor:])
			return out.Bytes()
		}

		keyPos := cursor + rel
		out.Write(body[cursor:keyPos])

		colonPos := skipJSONWhitespace(body, keyPos+len(needle))
		if colonPos >= len(body) || body[colonPos] != ':' {
			out.Write(body[keyPos : keyPos+len(needle)])
			cursor = keyPos + len(needle)
			continue
		}

		objStart := skipJSONWhitespace(body, colonPos+1)
		if objStart >= len(body) || body[objStart] != '{' {
			out.Write(body[keyPos:objStart])
			cursor = objStart
			continue
		}

		objEnd, ok := scanJSONObjectEnd(body, objStart)
		if !ok {
			out.Write(body[keyPos:])
			return out.Bytes()
		}

		out.Write(body[keyPos : objStart+1])
		out.Write(reorderJSONObjectMembers(body[objStart+1 : objEnd]))
		out.WriteByte('}')
		cursor = objEnd + 1
	}
}

func reorderJSONObjectMembers(body []byte) []byte {
	type segment struct {
		member jsonObjectMember
		sep    []byte
	}

	prefixEnd := skipJSONWhitespace(body, 0)
	if prefixEnd >= len(body) {
		return body
	}
	if body[prefixEnd] == '}' {
		return body
	}

	prefix := append([]byte(nil), body[:prefixEnd]...)
	segments := make([]segment, 0)
	suffix := []byte(nil)
	pos := prefixEnd
	for {
		keyStart := pos
		keyEnd, ok := scanJSONStringEnd(body, keyStart)
		if !ok {
			return body
		}
		key, err := strconv.Unquote(string(body[keyStart : keyEnd+1]))
		if err != nil {
			return body
		}

		colonPos := skipJSONWhitespace(body, keyEnd+1)
		if colonPos >= len(body) || body[colonPos] != ':' {
			return body
		}
		valueStart := skipJSONWhitespace(body, colonPos+1)
		valueEnd, ok := scanJSONValueEnd(body, valueStart)
		if !ok {
			return body
		}

		next := valueEnd
		if next < len(body) && body[next] == ',' {
			next++
			nextKeyStart := skipJSONWhitespace(body, next)
			segments = append(segments, segment{
				member: jsonObjectMember{
					key: key,
					raw: append([]byte(nil), body[keyStart:valueEnd]...),
				},
				sep: append([]byte(nil), body[valueEnd:nextKeyStart]...),
			})
			pos = nextKeyStart
			continue
		}

		segments = append(segments, segment{
			member: jsonObjectMember{
				key: key,
				raw: append([]byte(nil), body[keyStart:valueEnd]...),
			},
		})
		suffix = append([]byte(nil), body[valueEnd:]...)
		break
	}

	firstNumeric := -1
	for i, seg := range segments {
		if isDigitsOnly(seg.member.key) {
			firstNumeric = i
			break
		}
	}
	if firstNumeric <= 0 {
		return body
	}

	reordered := make([]jsonObjectMember, 0, len(segments))
	for _, seg := range segments {
		if isDigitsOnly(seg.member.key) {
			reordered = append(reordered, seg.member)
		}
	}
	for _, seg := range segments {
		if !isDigitsOnly(seg.member.key) {
			reordered = append(reordered, seg.member)
		}
	}

	sep := []byte(nil)
	for _, seg := range segments {
		if len(seg.sep) > 0 {
			sep = seg.sep
			break
		}
	}

	var out bytes.Buffer
	out.Write(prefix)
	for i, member := range reordered {
		if i > 0 && len(sep) > 0 {
			out.Write(sep)
		}
		out.Write(member.raw)
		if i < len(reordered)-1 && len(sep) == 0 {
			out.WriteByte(',')
		}
	}
	out.Write(suffix)
	return out.Bytes()
}

func skipJSONWhitespace(body []byte, pos int) int {
	for pos < len(body) && unicode.IsSpace(rune(body[pos])) {
		pos++
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

func scanJSONObjectEnd(body []byte, start int) (int, bool) {
	if start >= len(body) || body[start] != '{' {
		return 0, false
	}
	depth := 0
	inString := false
	escaped := false
	for i := start; i < len(body); i++ {
		ch := body[i]
		if inString {
			switch {
			case escaped:
				escaped = false
			case ch == '\\':
				escaped = true
			case ch == '"':
				inString = false
			}
			continue
		}
		switch ch {
		case '"':
			inString = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return i, true
			}
		}
	}
	return 0, false
}

func scanJSONValueEnd(body []byte, start int) (int, bool) {
	if start >= len(body) {
		return 0, false
	}
	switch body[start] {
	case '"':
		end, ok := scanJSONStringEnd(body, start)
		if !ok {
			return 0, false
		}
		return end + 1, true
	case '{':
		end, ok := scanJSONObjectEnd(body, start)
		if !ok {
			return 0, false
		}
		return end + 1, true
	case '[':
		depth := 0
		inString := false
		escaped := false
		for i := start; i < len(body); i++ {
			ch := body[i]
			if inString {
				switch {
				case escaped:
					escaped = false
				case ch == '\\':
					escaped = true
				case ch == '"':
					inString = false
				}
				continue
			}
			switch ch {
			case '"':
				inString = true
			case '[':
				depth++
			case ']':
				depth--
				if depth == 0 {
					return i + 1, true
				}
			}
		}
		return 0, false
	default:
		for i := start; i < len(body); i++ {
			switch body[i] {
			case ',', '}', ']':
				return i, true
			}
		}
		return len(body), true
	}
}

func isDigitsOnly(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
