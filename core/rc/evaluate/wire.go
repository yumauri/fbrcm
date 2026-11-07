package evaluate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/tailscale/hujson"
)

// Firebase's Apple SDK uses a restricted JSON-like wire format with bare
// object keys and single-quoted strings. Normalize data only; never evaluate
// JavaScript or accept expressions, calls, or identifiers as values.
func normalizeFetchJSON(data []byte) ([]byte, error) {
	if standard, err := hujson.Standardize(data); err == nil {
		return standard, nil
	}
	var out bytes.Buffer
	for i := 0; i < len(data); {
		switch c := data[i]; {
		case c == '\'' || c == '"':
			start := i
			i++
			for i < len(data) && data[i] != c {
				if data[i] == '\\' {
					i++
				}
				i++
			}
			if i >= len(data) {
				return nil, fmt.Errorf("unterminated fetch string")
			}
			i++
			if c == '"' {
				out.Write(data[start:i])
				continue
			}
			var value []rune
			for content := string(data[start+1 : i-1]); content != ""; {
				r, _, rest, err := strconv.UnquoteChar(content, '\'')
				if err != nil {
					return nil, fmt.Errorf("invalid fetch string")
				}
				value = append(value, r)
				content = rest
			}
			encoded, _ := json.Marshal(string(value))
			out.Write(encoded)
		case c == '_' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z':
			start := i
			for i < len(data) && (data[i] == '_' || data[i] >= 'a' && data[i] <= 'z' || data[i] >= 'A' && data[i] <= 'Z' || data[i] >= '0' && data[i] <= '9') {
				i++
			}
			end := i
			for i < len(data) && (data[i] == ' ' || data[i] == '\t' || data[i] == '\r' || data[i] == '\n') {
				i++
			}
			if i < len(data) && data[i] == ':' {
				key, _ := json.Marshal(string(data[start:end]))
				out.Write(key)
				out.Write(data[end:i])
			} else {
				out.Write(data[start:i]) // Only JSON true, false, null can survive validation.
			}
		default:
			out.WriteByte(c)
			i++
		}
	}
	standard, err := hujson.Standardize(out.Bytes())
	if err != nil {
		return nil, fmt.Errorf("expected a fetch request object")
	}
	return standard, nil
}
