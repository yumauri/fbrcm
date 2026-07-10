package jsoninput

import (
	"charm.land/lipgloss/v2"
	"github.com/yumauri/fbrcm/tui/styles"
)

func cursorStyle() lipgloss.Style {
	if styles.NoColorEnabled() {
		return lipgloss.NewStyle().Reverse(true).Bold(true)
	}
	return lipgloss.NewStyle().Background(styles.PaletteYellow).Foreground(styles.PaletteBlueDeep).Bold(true)
}

func highlightJSONRune(r rune, inString, stringIsKey bool) string {
	switch {
	case inString && stringIsKey:
		return jsonKeyStyle().Render(string(r))
	case inString:
		return jsonStringStyle().Render(string(r))
	case r == '"':
		return jsonStringContextStyle(stringIsKey).Render(`"`)
	case r == '{' || r == '}' || r == '[' || r == ']' || r == ':' || r == ',':
		return jsonPunctuationStyle().Render(string(r))
	default:
		return jsonDefaultStyle().Render(string(r))
	}
}

func jsonDefaultStyle() lipgloss.Style {
	return styles.PanelText
}

func jsonPunctuationStyle() lipgloss.Style {
	return styles.PanelText
}

func jsonKeyStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(styles.ConditionLipglossColor("CYAN"))
}

func jsonStringStyle() lipgloss.Style {
	return styles.PanelText
}

func jsonStringContextStyle(key bool) lipgloss.Style {
	if key {
		return jsonKeyStyle()
	}
	return jsonStringStyle()
}

func jsonNumberStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(styles.PaletteBlueBright)
}

func jsonTrueStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(styles.ConditionLipglossColor("GREEN"))
}

func jsonFalseStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(styles.PaletteError)
}

func jsonNullStyle() lipgloss.Style {
	return styles.PanelMuted
}

func jsonTokenStyle(token string) lipgloss.Style {
	switch token {
	case "true":
		return jsonTrueStyle()
	case "false":
		return jsonFalseStyle()
	case "null":
		return jsonNullStyle()
	default:
		if token != "" && (token[0] == '-' || (token[0] >= '0' && token[0] <= '9')) {
			return jsonNumberStyle()
		}
		return jsonDefaultStyle()
	}
}
