package jsoninput

import (
	"fmt"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/yumauri/fbrcm/tui/styles"
)

type jsonContainer struct {
	kind           rune
	expectingKey   bool
	expectingValue bool
}

func highlightJSONRanges(value string, ranges []JSONRange) []string {
	if len(ranges) == 0 {
		return nil
	}
	runes := []rune(value)
	normalized := make([]JSONRange, len(ranges))
	for i, r := range ranges {
		if r.Start < 0 {
			r.Start = 0
		}
		if r.End < r.Start {
			r.End = r.Start
		}
		if r.Start > len(runes) {
			r.Start = len(runes)
		}
		if r.End > len(runes) {
			r.End = len(runes)
		}
		normalized[i] = r
	}

	builders := make([]strings.Builder, len(normalized))
	rangeIndex := 0
	advanceRange := func(pos int) {
		for rangeIndex < len(normalized) && pos >= normalized[rangeIndex].End {
			rangeIndex++
		}
	}
	writeVisible := func(pos int, rendered string) {
		advanceRange(pos)
		if rangeIndex >= len(normalized) {
			return
		}
		r := normalized[rangeIndex]
		if pos >= r.Start && pos < r.End {
			builders[rangeIndex].WriteString(rendered)
		}
	}

	inString := false
	escaped := false
	stringIsKey := false
	stack := make([]jsonContainer, 0, 8)
	lit := strings.Builder{}
	litStart := 0
	flushLit := func() {
		if lit.Len() == 0 {
			return
		}
		token := lit.String()
		tokenRunes := []rune(token)
		style := jsonTokenStyle(token)
		for offset, r := range tokenRunes {
			pos := litStart + offset
			advanceRange(pos)
			rendered := style.Render(string(r))
			if rangeIndex < len(normalized) && pos == normalized[rangeIndex].CursorCol {
				rendered = cursorStyle().Render(rendered)
			}
			writeVisible(pos, rendered)
		}
		lit.Reset()
	}
	writeStyledRune := func(pos int, r rune, style lipgloss.Style) {
		advanceRange(pos)
		rendered := style.Render(string(r))
		if rangeIndex < len(normalized) && pos == normalized[rangeIndex].CursorCol {
			rendered = cursorStyle().Render(rendered)
		}
		writeVisible(pos, rendered)
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

	for i, r := range runes {
		if rangeIndex >= len(normalized) {
			break
		}
		if inString {
			if escaped {
				writeStyledRune(i, r, jsonStringContextStyle(stringIsKey))
				escaped = false
				continue
			}
			if r == '\\' {
				writeStyledRune(i, r, jsonStringContextStyle(stringIsKey))
				escaped = true
				continue
			}
			writeStyledRune(i, r, jsonStringContextStyle(stringIsKey))
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
			writeStyledRune(i, r, jsonStringContextStyle(stringIsKey))
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
			writeStyledRune(i, r, jsonPunctuationStyle())
		case ' ', '\n', '\t', '\r':
			flushLit()
			writeStyledRune(i, r, jsonDefaultStyle())
		default:
			if lit.Len() == 0 {
				litStart = i
				markValueStarted()
			}
			lit.WriteRune(r)
		}
	}
	flushLit()
	out := make([]string, len(builders))
	for i := range builders {
		r := normalized[i]
		if r.CursorCol == r.End || (r.CursorCol == len(runes) && r.CursorCol >= r.Start && r.CursorCol <= r.End) {
			builders[i].WriteString(cursorStyle().Render(" "))
		}
		out[i] = builders[i].String()
	}
	return out
}

func renderLineNumber(n, total int, active bool) string {
	digits := max(len(strconv.Itoa(max(total, 1))), 1)
	style := styles.PanelMuted
	if active {
		style = styles.PanelText.Bold(true)
	}
	return style.Render(fmt.Sprintf("%*d ", digits, n))
}

func lineNumberGutter(total int) int {
	return max(len(strconv.Itoa(max(total, 1))), 1) + 1
}

func padHighlighted(value string, width int) string {
	return value + strings.Repeat(" ", max(width-lipgloss.Width(value), 0))
}

func HighlightJSONRanges(value string, ranges []JSONRange) []string {
	return highlightJSONRanges(value, ranges)
}
