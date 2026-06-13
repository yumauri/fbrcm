package minsize

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

const (
	MinWidth  = 80
	MinHeight = 20
)

func View(width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	message := strings.Join([]string{
		"Terminal too small",
		fmt.Sprintf("Minimum size: %dx%d", MinWidth, MinHeight),
	}, "\n")
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, message)
}
