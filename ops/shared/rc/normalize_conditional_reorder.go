package rc

import (
	"bytes"
)

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
		member, valueEnd, ok := parseJSONObjectMember(body, pos)
		if !ok {
			return body
		}
		next := valueEnd
		if next < len(body) && body[next] == ',' {
			next++
			nextKeyStart := skipJSONWhitespaceExport(body, next)
			segments = append(segments, segment{
				member: member,
				sep:    append([]byte(nil), body[valueEnd:nextKeyStart]...),
			})
			pos = nextKeyStart
			continue
		}
		segments = append(segments, segment{member: member})
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
	writeJSONObjectMembers(&out, reordered, sep)
	out.Write(suffix)
	return out.Bytes()
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
