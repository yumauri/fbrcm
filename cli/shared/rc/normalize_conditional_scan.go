package rc

import (
	"bytes"
	"strconv"
	"unicode"

	"github.com/yumauri/fbrcm/cli/internal/jsonscan"
)

type jsonObjectMember struct {
	key string
	raw []byte
}

func skipJSONWhitespaceExport(body []byte, pos int) int {
	for pos < len(body) && unicode.IsSpace(rune(body[pos])) {
		pos++
	}
	return pos
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
		end, ok := jsonscan.ScanStringEnd(body, start)
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

func parseJSONObjectMember(body []byte, pos int) (member jsonObjectMember, next int, ok bool) {
	keyStart := pos
	keyEnd, ok := jsonscan.ScanStringEnd(body, keyStart)
	if !ok {
		return jsonObjectMember{}, 0, false
	}
	key, err := strconv.Unquote(string(body[keyStart : keyEnd+1]))
	if err != nil {
		return jsonObjectMember{}, 0, false
	}
	colonPos := skipJSONWhitespaceExport(body, keyEnd+1)
	if colonPos >= len(body) || body[colonPos] != ':' {
		return jsonObjectMember{}, 0, false
	}
	valueStart := skipJSONWhitespaceExport(body, colonPos+1)
	valueEnd, ok := scanJSONValueEndExport(body, valueStart)
	if !ok {
		return jsonObjectMember{}, 0, false
	}
	return jsonObjectMember{
		key: key,
		raw: append([]byte(nil), body[keyStart:valueEnd]...),
	}, valueEnd, true
}

func writeJSONObjectMembers(out *bytes.Buffer, members []jsonObjectMember, sep []byte) {
	for i, member := range members {
		if i > 0 {
			if len(sep) > 0 {
				out.Write(sep)
			} else {
				out.WriteByte(',')
			}
		}
		out.Write(member.raw)
	}
}
