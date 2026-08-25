// Package about provides the shared application identity displayed by the CLI
// and TUI.
package about

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"

	corestyles "github.com/yumauri/fbrcm/core/styles"
)

const (
	Name = "fbrcm"
	Logo = `▄─┐▄
█- █─▄ ▄─▄ ▄── ▄─▄─▄
▀  ▀─▀ ▀   ▀── ▀ ▀ ▀`

	Author = "Victor Didenko <yumaa.verdin@gmail.com> (https://yumaa.name)"
)

// BuildInfo identifies one fbrcm build.
type BuildInfo struct {
	Version string
	Commit  string
	Date    string
}

type logoColor struct {
	red   int
	green int
	blue  int
}

type logoColorStop struct {
	position float64
	color    logoColor
}

func logoGradient() []logoColorStop {
	return []logoColorStop{
		{position: 0, color: themeLogoColor(corestyles.ColorLogoStart)},
		{position: 0.20, color: themeLogoColor(corestyles.ColorLogoMiddle)},
		{position: 1, color: themeLogoColor(corestyles.ColorLogoEnd)},
	}
}

func themeLogoColor(value color.Color) logoColor {
	red, green, blue, _ := value.RGBA()
	return logoColor{red: int(red >> 8), green: int(green >> 8), blue: int(blue >> 8)}
}

// Metadata returns the version value used by Cobra after the application name.
func (i BuildInfo) Metadata() string {
	return fmt.Sprintf("%s (commit %s, built %s)", i.Version, i.Commit, i.Date)
}

// Text returns the complete CLI-equivalent About text with a trailing newline.
func (i BuildInfo) Text(color bool) string {
	return RenderLogo(color) + "\n" + Name + " " + i.Metadata() + "\n" + Author + "\n"
}

// RenderLogo returns the shared text logo, optionally styled with its Firebase
// yellow, orange, and red gradient.
func RenderLogo(color bool) string {
	if !color {
		return Logo
	}

	lines := strings.Split(Logo, "\n")
	width := 0
	for _, line := range lines {
		width = max(width, len([]rune(line)))
	}

	var rendered strings.Builder
	for lineIndex, line := range lines {
		if lineIndex > 0 {
			rendered.WriteByte('\n')
		}
		for column, char := range []rune(line) {
			if char == ' ' {
				rendered.WriteRune(char)
				continue
			}
			color := interpolateLogoColor(column, width)
			style := lipgloss.NewStyle().Foreground(lipgloss.Color(fmt.Sprintf("#%02X%02X%02X", color.red, color.green, color.blue)))
			rendered.WriteString(style.Render(string(char)))
		}
	}
	return rendered.String()
}

func interpolateLogoColor(column, width int) logoColor {
	gradient := logoGradient()
	if width <= 1 || column <= 0 {
		return gradient[0].color
	}
	if column >= width-1 {
		return gradient[len(gradient)-1].color
	}

	position := float64(column) / float64(width-1)
	segment := 0
	for segment+1 < len(gradient)-1 && position > gradient[segment+1].position {
		segment++
	}
	left, right := gradient[segment], gradient[segment+1]
	fraction := (position - left.position) / (right.position - left.position)
	return logoColor{
		red:   interpolateLogoChannel(left.color.red, right.color.red, fraction),
		green: interpolateLogoChannel(left.color.green, right.color.green, fraction),
		blue:  interpolateLogoChannel(left.color.blue, right.color.blue, fraction),
	}
}

func interpolateLogoChannel(left, right int, fraction float64) int {
	return int(float64(left) + float64(right-left)*fraction + 0.5)
}
