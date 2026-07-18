// Package inputstyles centralizes the shared bubbles text input/textarea
// styling and constructors used by the TUI value editors and filters.
package inputstyles

import (
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"

	"github.com/yumauri/fbrcm/tui/styles"
)

// TextInput returns the shared single-line text input styling.
func TextInput() textinput.Styles {
	inputStyles := textinput.DefaultDarkStyles()
	valueStyle := styles.FilterText
	inputStyles.Focused.Text = valueStyle
	inputStyles.Focused.Prompt = valueStyle
	inputStyles.Focused.Placeholder = valueStyle
	inputStyles.Focused.Suggestion = valueStyle
	inputStyles.Blurred.Text = valueStyle
	inputStyles.Blurred.Prompt = valueStyle
	inputStyles.Blurred.Placeholder = valueStyle
	inputStyles.Blurred.Suggestion = valueStyle
	inputStyles.Cursor.Color = styles.PaletteYellow
	return inputStyles
}

// InlineListTextInput returns the text input styling used for editable rows
// embedded alongside selectable list options.
func InlineListTextInput() textinput.Styles {
	inputStyles := textinput.DefaultDarkStyles()
	valueStyle := styles.PanelText
	placeholderStyle := styles.PanelMuted
	inputStyles.Focused.Text = valueStyle
	inputStyles.Focused.Prompt = valueStyle
	inputStyles.Focused.Placeholder = placeholderStyle
	inputStyles.Focused.Suggestion = valueStyle
	inputStyles.Blurred.Text = valueStyle
	inputStyles.Blurred.Prompt = valueStyle
	inputStyles.Blurred.Placeholder = placeholderStyle
	inputStyles.Blurred.Suggestion = valueStyle
	inputStyles.Cursor.Color = styles.PaletteYellow
	return inputStyles
}

// Textarea returns the shared multi-line textarea styling.
func Textarea() textarea.Styles {
	s := textarea.DefaultStyles(true)
	textStyle := styles.FilterText
	s.Focused.Text = textStyle
	s.Focused.Prompt = lipgloss.NewStyle()
	s.Focused.Placeholder = styles.PanelMuted
	s.Focused.LineNumber = lipgloss.NewStyle()
	s.Focused.CursorLineNumber = lipgloss.NewStyle()
	s.Focused.CursorLine = lipgloss.NewStyle()
	s.Focused.EndOfBuffer = lipgloss.NewStyle()
	s.Blurred.Text = textStyle
	s.Blurred.Prompt = lipgloss.NewStyle()
	s.Blurred.Placeholder = styles.PanelMuted
	s.Blurred.LineNumber = lipgloss.NewStyle()
	s.Blurred.CursorLineNumber = lipgloss.NewStyle()
	s.Blurred.CursorLine = lipgloss.NewStyle()
	s.Blurred.EndOfBuffer = lipgloss.NewStyle()
	s.Cursor.Color = styles.PaletteYellow
	return s
}

// TextareaDescription returns textarea styling for the details description field.
func TextareaDescription() textarea.Styles {
	s := Textarea()
	textStyle := styles.FilterText
	s.Focused.CursorLine = textStyle
	return s
}

// NewTextInput returns a blurred single-line text input with shared styling.
func NewTextInput() textinput.Model {
	input := textinput.New()
	input.Prompt = ""
	input.SetStyles(TextInput())
	input.Blur()
	return input
}

// NewTextarea returns a blurred multi-line textarea with shared styling.
func NewTextarea() textarea.Model {
	input := textarea.New()
	input.Prompt = ""
	input.ShowLineNumbers = false
	input.EndOfBufferCharacter = ' '
	input.SetStyles(Textarea())
	input.Blur()
	return input
}

// NewDescriptionTextarea returns a blurred textarea for the details description field.
func NewDescriptionTextarea() textarea.Model {
	input := textarea.New()
	input.Prompt = ""
	input.ShowLineNumbers = false
	input.EndOfBufferCharacter = ' '
	input.SetVirtualCursor(false)
	input.SetStyles(TextareaDescription())
	input.Blur()
	return input
}
