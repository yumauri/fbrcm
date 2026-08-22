package viewutil

import (
	"strings"

	"github.com/rivo/uniseg"
)

// CropTextBeforeRender returns the leading text that can occupy at most the
// requested number of wrapped display lines. It stops scanning as soon as the
// budget is exhausted so callers can bound styling and syntax-highlighting
// work for very large values.
func CropTextBeforeRender(value string, firstWidth, continuationWidth, lineBudget int) (string, bool) {
	if lineBudget <= 0 {
		return value, false
	}
	firstWidth = max(firstWidth, 1)
	continuationWidth = max(continuationWidth, 1)

	var out strings.Builder
	graphemes := uniseg.NewGraphemes(value)
	line := 0
	lineWidth := 0
	lineLimit := firstWidth
	for graphemes.Next() {
		cluster := graphemes.Str()
		if cluster == "\r" {
			continue
		}
		if cluster == "\n" {
			_, end := graphemes.Positions()
			if line+1 >= lineBudget {
				return out.String(), end < len(value)
			}
			out.WriteByte('\n')
			line++
			lineWidth = 0
			lineLimit = continuationWidth
			continue
		}

		clusterWidth := graphemes.Width()
		if lineWidth > 0 && lineWidth+clusterWidth > lineLimit {
			if line+1 >= lineBudget {
				return out.String(), true
			}
			line++
			lineWidth = 0
			lineLimit = continuationWidth
		}
		out.WriteString(cluster)
		lineWidth += clusterWidth
	}
	return out.String(), false
}
