package shared

import (
	"image/color"
	"io"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/erikgeiser/promptkit/confirmation"

	clistyles "github.com/yumauri/fbrcm/cli/styles"
)

type ConfirmationNote struct {
	Text  string
	Color color.Color
}

type ConfirmationOptions struct {
	Destructive bool
	Notes       []ConfirmationNote
}

// RunConfirmationPrompt shows a yes/no prompt and returns the user's choice.
// When destructive is true the affirmative option is styled as destructive.
// A non-nil fallbackOut overrides the prompt's output writer.
func RunConfirmationPrompt(prompt string, destructive bool, fallbackOut io.Writer) (bool, error) {
	confirm := NewConfirmation(prompt, confirmation.Yes, ConfirmationOptions{Destructive: destructive})
	if fallbackOut != nil {
		confirm.Output = fallbackOut
	}
	return confirm.RunPrompt()
}

func NewConfirmation(prompt string, defaultValue confirmation.Value, options ConfirmationOptions) *confirmation.Confirmation {
	confirm := confirmation.New(prompt, defaultValue)
	hint := renderConfirmationNotes(options.Notes)

	confirm.Template = `
{{- Bold .Prompt -}}
{{- "\n" -}}
{{- if HasHint -}}
{{- Hint -}}
{{- "\n" -}}
{{- end -}}
{{ if .YesSelected -}}
	{{- print (YesSelectedLabel) " No" -}}
{{- else if .NoSelected -}}
	{{- print (YesLabel) (Bold "▸No") -}}
{{- else -}}
	{{- print (YesLabel) " No" -}}
{{- end -}}
`
	confirm.ExtendedTemplateFuncs["HasHint"] = func() bool {
		return hint != ""
	}
	confirm.ExtendedTemplateFuncs["Hint"] = func() string {
		return hint
	}
	confirm.ExtendedTemplateFuncs["YesLabel"] = func() string {
		return renderYesLabel("  Yes ", options.Destructive, false)
	}
	confirm.ExtendedTemplateFuncs["YesSelectedLabel"] = func() string {
		return renderYesLabel(" ▸Yes ", options.Destructive, true)
	}
	return confirm
}

func renderConfirmationNotes(notes []ConfirmationNote) string {
	lines := make([]string, 0, len(notes))
	for _, note := range notes {
		if note.Text == "" {
			continue
		}
		if clistyles.NoColorEnabled() {
			lines = append(lines, note.Text)
			continue
		}
		lines = append(lines, lipgloss.NewStyle().Foreground(note.Color).Render(note.Text))
	}
	return strings.Join(lines, "\n")
}

func renderYesLabel(label string, destructive, selected bool) string {
	if clistyles.NoColorEnabled() {
		if selected {
			return lipgloss.NewStyle().Bold(true).Render(label)
		}
		return label
	}

	yesColor := clistyles.PaletteBlueBright
	if destructive {
		yesColor = clistyles.PaletteError
	}

	style := lipgloss.NewStyle().Foreground(yesColor)
	if selected {
		style = style.Bold(true)
	}
	return style.Render(label)
}
