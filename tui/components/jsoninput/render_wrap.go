package jsoninput

import (
	"strings"
	"unicode"

	rw "github.com/mattn/go-runewidth"
	"github.com/rivo/uniseg"
)

type plainWrappedLine struct {
	text  string
	start int
}

func wrapPlainLine(value string, width int) []plainWrappedLine {
	if width <= 0 {
		return []plainWrappedLine{{}}
	}
	wrapped := textareaWrap([]rune(value), width)
	out := make([]plainWrappedLine, 0, len(wrapped))
	start := 0
	for _, part := range wrapped {
		text := string(part)
		if uniseg.StringWidth(text) > width {
			text = strings.TrimSuffix(text, " ")
		}
		out = append(out, plainWrappedLine{text: text, start: start})
		start += len(part)
	}
	if len(out) == 0 {
		return []plainWrappedLine{{}}
	}
	return out
}

func textareaWrap(runes []rune, width int) [][]rune {
	lines := [][]rune{{}}
	word := []rune{}
	row := 0
	spaces := 0

	for _, r := range runes {
		if unicode.IsSpace(r) {
			spaces++
		} else {
			word = append(word, r)
		}

		if spaces > 0 {
			if uniseg.StringWidth(string(lines[row]))+uniseg.StringWidth(string(word))+spaces > width {
				row++
				lines = append(lines, []rune{})
			}
			lines[row] = append(lines[row], word...)
			lines[row] = append(lines[row], repeatSpaces(spaces)...)
			spaces = 0
			word = nil
			continue
		}

		lastCharLen := rw.RuneWidth(word[len(word)-1])
		if uniseg.StringWidth(string(word))+lastCharLen > width {
			if len(lines[row]) > 0 {
				row++
				lines = append(lines, []rune{})
			}
			lines[row] = append(lines[row], word...)
			word = nil
		}
	}

	if uniseg.StringWidth(string(lines[row]))+uniseg.StringWidth(string(word))+spaces >= width {
		lines = append(lines, []rune{})
		lines[row+1] = append(lines[row+1], word...)
		spaces++
		lines[row+1] = append(lines[row+1], repeatSpaces(spaces)...)
	} else {
		lines[row] = append(lines[row], word...)
		spaces++
		lines[row] = append(lines[row], repeatSpaces(spaces)...)
	}

	return lines
}

func repeatSpaces(n int) []rune {
	return []rune(strings.Repeat(" ", n))
}
