package jsoninput

import "strings"

func highlightJSONVisible(value string) string {
	var out strings.Builder
	inString := false
	escaped := false
	stringIsKey := false
	stack := make([]jsonContainer, 0, 8)
	lit := strings.Builder{}
	flushLit := func() {
		if lit.Len() == 0 {
			return
		}
		token := lit.String()
		style := jsonTokenStyle(token)
		out.WriteString(style.Render(token))
		lit.Reset()
	}
	currentStringIsKey := func() bool {
		if len(stack) == 0 {
			return false
		}
		top := stack[len(stack)-1]
		return top.kind == '{' && top.expectingKey
	}
	markValueStarted := func() {
		if len(stack) == 0 {
			return
		}
		top := &stack[len(stack)-1]
		if top.kind == '{' && top.expectingValue {
			top.expectingValue = false
		}
	}
	for _, r := range value {
		if inString {
			if escaped {
				out.WriteString(highlightJSONRune(r, true, stringIsKey))
				escaped = false
				continue
			}
			if r == '\\' {
				out.WriteString(highlightJSONRune(r, true, stringIsKey))
				escaped = true
				continue
			}
			out.WriteString(highlightJSONRune(r, true, stringIsKey))
			if r == '"' {
				inString = false
			}
			continue
		}
		switch r {
		case '"':
			flushLit()
			inString = true
			stringIsKey = currentStringIsKey()
			if stringIsKey {
				stack[len(stack)-1].expectingKey = false
			} else {
				markValueStarted()
			}
			out.WriteString(jsonStringContextStyle(stringIsKey).Render(`"`))
		case '{', '}', '[', ']', ':', ',':
			flushLit()
			switch r {
			case '{':
				markValueStarted()
				stack = append(stack, jsonContainer{kind: '{', expectingKey: true})
			case '[':
				markValueStarted()
				stack = append(stack, jsonContainer{kind: '['})
			case '}', ']':
				if len(stack) > 0 {
					stack = stack[:len(stack)-1]
				}
			case ':':
				if len(stack) > 0 && stack[len(stack)-1].kind == '{' {
					stack[len(stack)-1].expectingValue = true
				}
			case ',':
				if len(stack) > 0 && stack[len(stack)-1].kind == '{' {
					stack[len(stack)-1].expectingKey = true
					stack[len(stack)-1].expectingValue = false
				}
			}
			out.WriteString(highlightJSONRune(r, false, false))
		case ' ', '\n', '\t', '\r':
			flushLit()
			out.WriteRune(r)
		default:
			if lit.Len() == 0 {
				markValueStarted()
			}
			lit.WriteRune(r)
		}
	}
	flushLit()
	return out.String()
}

func HighlightJSONVisible(value string) string {
	return highlightJSONVisible(value)
}
