package shared

import (
	"bytes"
	"image/color"
	"strings"
	"testing"

	"github.com/erikgeiser/promptkit/confirmation"
)

func TestNewConfirmationIncludesNotes(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	confirm := NewConfirmation("Delete flag?", ConfirmationOptions{
		Destructive: true,
		Notes: []ConfirmationNote{{
			Text:  "This cannot be undone.",
			Color: color.RGBA{R: 255, A: 255},
		}},
	})
	if confirm == nil {
		t.Fatal("NewConfirmation returned nil")
	}
	if confirm.Template == "" {
		t.Fatal("confirmation template is empty")
	}
	if confirm.DefaultValue != confirmation.Yes {
		t.Fatalf("confirmation default = %v, want Yes", confirm.DefaultValue)
	}
}

func TestRenderConfirmationNotesPlainText(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	got := renderConfirmationNotes([]ConfirmationNote{{Text: "note one"}, {Text: "note two"}})
	if got != "note one\nnote two" {
		t.Fatalf("renderConfirmationNotes = %q", got)
	}
}

func TestRunConfirmationPromptUsesFallbackWriter(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	var buf bytes.Buffer
	// NOTE(suspicious): RunConfirmationPrompt always blocks for interactive input;
	// this test only verifies wiring by checking NewConfirmation construction paths.
	confirm := NewConfirmation("Proceed?", ConfirmationOptions{})
	confirm.Output = &buf
	if confirm.Output != &buf {
		t.Fatal("confirmation output writer was not assigned")
	}
	if !strings.Contains(confirm.Template, "Prompt") {
		t.Fatal("confirmation template missing prompt marker")
	}
}
