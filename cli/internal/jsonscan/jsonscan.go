// Package jsonscan provides low-level helpers for scanning raw JSON bytes
// while preserving the original formatting.
package jsonscan

// ScanStringEnd returns the index of the closing quote of the JSON string that
// starts at body[start], honoring backslash escapes. ok is false when start is
// not a quote or the string is unterminated.
func ScanStringEnd(body []byte, start int) (end int, ok bool) {
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
