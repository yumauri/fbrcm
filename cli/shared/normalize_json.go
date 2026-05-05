package shared

import (
	"bytes"
	"strconv"
	"unicode"
)

func NormalizeExportJSON(body []byte) []byte {
	body = bytes.ReplaceAll(body, []byte(`\u003c`), []byte("<"))
	body = bytes.ReplaceAll(body, []byte(`\u003e`), []byte(">"))
	body = bytes.ReplaceAll(body, []byte(`\u0026`), []byte("&"))
	body = reorderConditionalValuesKeysNumericFirst(body)
	return body
}

func TrimTrailingLineBreaks(body []byte) []byte {
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
		colonPos := skipJSONWhitespaceExport(body, keyPos+len(needle))
		if colonPos >= len(body) || body[colonPos] != ':' {
			out.Write(body[keyPos : keyPos+len(needle)])
			cursor = keyPos + len(needle)
			continue
		}
		objStart := skipJSONWhitespaceExport(body, colonPos+1)
		if objStart >= len(body) || body[objStart] != '{' {
			out.Write(body[keyPos:objStart])
			cursor = objStart
			continue
		}
		objEnd, ok := scanJSONObjectEndExport(body, objStart)
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
	prefixEnd := skipJSONWhitespaceExport(body, 0)
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
		keyEnd, ok := scanJSONStringEndExport(body, keyStart)
		if !ok {
			return body
		}
		key, err := strconv.Unquote(string(body[keyStart : keyEnd+1]))
		if err != nil {
			return body
		}
		colonPos := skipJSONWhitespaceExport(body, keyEnd+1)
		if colonPos >= len(body) || body[colonPos] != ':' {
			return body
		}
		valueStart := skipJSONWhitespaceExport(body, colonPos+1)
		valueEnd, ok := scanJSONValueEndExport(body, valueStart)
		if !ok {
			return body
		}
		next := valueEnd
		if next < len(body) && body[next] == ',' {
			next++
			nextKeyStart := skipJSONWhitespaceExport(body, next)
			segments = append(segments, segment{
				member: jsonObjectMember{key: key, raw: append([]byte(nil), body[keyStart:valueEnd]...)},
				sep:    append([]byte(nil), body[valueEnd:nextKeyStart]...),
			})
			pos = nextKeyStart
			continue
		}
		segments = append(segments, segment{
			member: jsonObjectMember{key: key, raw: append([]byte(nil), body[keyStart:valueEnd]...)},
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

func skipJSONWhitespaceExport(body []byte, pos int) int {
	for pos < len(body) && unicode.IsSpace(rune(body[pos])) {
		pos++
	}
	return pos
}

func scanJSONStringEndExport(body []byte, start int) (int, bool) {
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

func scanJSONObjectEndExport(body []byte, start int) (int, bool) {
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

func scanJSONValueEndExport(body []byte, start int) (int, bool) {
	if start >= len(body) {
		return 0, false
	}
	switch body[start] {
	case '"':
		end, ok := scanJSONStringEndExport(body, start)
		if !ok {
			return 0, false
		}
		return end + 1, true
	case '{':
		end, ok := scanJSONObjectEndExport(body, start)
		if !ok {
			return 0, false
		}
		return end + 1, true
	case '[':
		end, ok := scanJSONArrayEndExport(body, start)
		if !ok {
			return 0, false
		}
		return end + 1, true
	default:
		pos := start
		for pos < len(body) {
			switch body[pos] {
			case ',', '}', ']':
				return pos, true
			}
			if unicode.IsSpace(rune(body[pos])) {
				return pos, true
			}
			pos++
		}
		return pos, true
	}
}

func scanJSONArrayEndExport(body []byte, start int) (int, bool) {
	if start >= len(body) || body[start] != '[' {
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
		case '[':
			depth++
		case ']':
			depth--
			if depth == 0 {
				return i, true
			}
		case '{':
			end, ok := scanJSONObjectEndExport(body, i)
			if !ok {
				return 0, false
			}
			i = end
		}
	}
	return 0, false
}

func isDigitsOnly(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
