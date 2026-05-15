package shared

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/erikgeiser/promptkit/confirmation"

	clistyles "github.com/yumauri/fbrcm/cli/styles"
)

// ConfirmationNote holds confirmation note state used by the shared package.
type ConfirmationNote struct {
	// Text stores text for ConfirmationNote.
	Text string
	// Color stores color for ConfirmationNote.
	Color color.Color
}

// ConfirmationOptions holds confirmation options state used by the shared package.
type ConfirmationOptions struct {
	// Destructive stores destructive for ConfirmationOptions.
	Destructive bool
	// Notes stores notes for ConfirmationOptions.
	Notes []ConfirmationNote
}

// NewConfirmation constructs confirmation and returns the resulting value or error.
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

// renderConfirmationNotes renders render confirmation notes and returns the resulting value or error.
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

// renderYesLabel renders render yes label and returns the resulting value or error.
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
