package filter

import (
	"encoding/json"
	"strings"
	"unicode"

	"github.com/itchyny/gojq"
)

// compileJQ compiles and caches a gojq query.
func compileJQ(query string) (*gojq.Code, error) {
	if cached, ok := jqCodeCache.Load(query); ok {
		if err, ok := cached.(error); ok {
			return nil, err
		}
		return cached.(*gojq.Code), nil
	}
	parsed, err := gojq.Parse(query)
	if err != nil {
		jqCodeCache.Store(query, err)
		return nil, err
	}
	code, err := gojq.Compile(parsed)
	if err != nil {
		jqCodeCache.Store(query, err)
		return nil, err
	}
	jqCodeCache.Store(query, code)
	return code, nil
}

// jqResultsForValue runs code against value and returns successful results.
func jqResultsForValue(value any, code *gojq.Code) []any {
	if values, ok := value.(anyValue); ok {
		results := make([]any, 0, len(values.values))
		for _, item := range values.values {
			results = append(results, jqResultsForValue(item, code)...)
		}
		return results
	}
	input, ok := jqInputForValue(value)
	if !ok {
		return nil
	}
	results := []any{}
	iter := code.Run(input)
	for {
		result, ok := iter.Next()
		if !ok {
			return results
		}
		if _, ok := result.(error); ok {
			return results
		}
		results = append(results, result)
	}
}

// jqInputForValue converts an expression value into gojq input.
func jqInputForValue(value any) (any, bool) {
	text, ok := value.(string)
	if !ok {
		return value, true
	}
	var input any
	if err := json.Unmarshal([]byte(text), &input); err != nil {
		return nil, false
	}
	return input, true
}

// jqResultsAreBool reports whether all results are booleans.
func jqResultsAreBool(results []any) bool {
	for _, result := range results {
		if _, ok := result.(bool); !ok {
			return false
		}
	}
	return true
}

// prepareJQExpressions wraps unquoted jq(...) arguments as string literals.
func prepareJQExpressions(raw string) string {
	var out strings.Builder
	for i := 0; i < len(raw); {
		if isExprStringStart(raw[i]) {
			next := copyExprString(&out, raw, i)
			i = next
			continue
		}
		jqStart, paren := findJQCall(raw, i)
		if jqStart != i {
			out.WriteByte(raw[i])
			i++
			continue
		}
		closeParen := findMatchingParen(raw, paren)
		if closeParen < 0 {
			out.WriteString(raw[i:])
			break
		}
		out.WriteString(raw[jqStart : paren+1])
		arg := strings.TrimSpace(raw[paren+1 : closeParen])
		if arg == "" || isExprStringStart(arg[0]) {
			out.WriteString(raw[paren+1 : closeParen])
		} else {
			encoded, _ := json.Marshal(arg)
			out.Write(encoded)
		}
		out.WriteByte(')')
		i = closeParen + 1
	}
	return out.String()
}

// findJQCall reports whether a jq call starts at pos and returns its opening parenthesis.
func findJQCall(raw string, pos int) (int, int) {
	if pos > 0 && isIdentifierRune(rune(raw[pos-1])) {
		return -1, -1
	}
	if !strings.HasPrefix(raw[pos:], "jq") {
		return -1, -1
	}
	next := pos + len("jq")
	if next < len(raw) && isIdentifierRune(rune(raw[next])) {
		return -1, -1
	}
	for next < len(raw) && unicode.IsSpace(rune(raw[next])) {
		next++
	}
	if next >= len(raw) || raw[next] != '(' {
		return -1, -1
	}
	return pos, next
}

// findMatchingParen finds the closing parenthesis for open.
func findMatchingParen(raw string, open int) int {
	depth := 0
	for i := open; i < len(raw); {
		if isExprStringStart(raw[i]) {
			i = skipExprString(raw, i)
			continue
		}
		switch raw[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return i
			}
		}
		i++
	}
	return -1
}

// copyExprString copies a string literal and returns the next position.
func copyExprString(out *strings.Builder, raw string, start int) int {
	next := skipExprString(raw, start)
	out.WriteString(raw[start:next])
	return next
}

// skipExprString skips an expr or jq string literal.
func skipExprString(raw string, start int) int {
	quote := raw[start]
	for i := start + 1; i < len(raw); i++ {
		if raw[i] == '\\' {
			i++
			continue
		}
		if raw[i] == quote {
			return i + 1
		}
	}
	return len(raw)
}

// isExprStringStart reports whether b starts a string literal.
func isExprStringStart(b byte) bool {
	return b == '"' || b == '\'' || b == '`'
}

// isIdentifierRune reports whether r can be part of an identifier.
func isIdentifierRune(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}
