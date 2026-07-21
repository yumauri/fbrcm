package viewutil

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

const (
	PopupPaddingLeft  = 2
	PopupPaddingRight = 1
	PopupPaddingTop   = 1
)

// PopupInnerWidth returns the width between popup borders for a content width.
func PopupInnerWidth(contentWidth int) int {
	return PopupPaddingLeft + max(contentWidth, 0) + PopupPaddingRight
}

// PopupContentLine adds the standard horizontal popup padding and fits content.
func PopupContentLine(content string, contentWidth int) string {
	contentWidth = max(contentWidth, 0)
	content = ansi.Truncate(content, contentWidth, "")
	content += strings.Repeat(" ", max(contentWidth-lipgloss.Width(content), 0))
	return strings.Repeat(" ", PopupPaddingLeft) + content + strings.Repeat(" ", PopupPaddingRight)
}
