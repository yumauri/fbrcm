package progress

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/term"

	clistyles "github.com/yumauri/fbrcm/internal/terminal/styles"
)

func TestRendererUsesRequestedLineFrames(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	var out bytes.Buffer
	r := newRenderer(&out, true)

	r.start("Loading projects…")
	haltAnimation(r)
	r.mu.Lock()
	r.clearLocked()
	r.advanceLocked()
	r.renderLocked()
	r.clearLocked()
	r.advanceLocked()
	r.renderLocked()
	r.clearLocked()
	r.advanceLocked()
	r.renderLocked()
	r.mu.Unlock()
	r.stopProgress()

	got := ansi.Strip(out.String())
	for _, want := range []string{
		"/ Loading projects…",
		"- Loading projects…",
		"\\ Loading projects…",
		"| Loading projects…",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("progress output = %q, want frame %q", got, want)
		}
	}
}

func TestLogWriterClearsAndRedrawsProgress(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	var terminal bytes.Buffer
	var logs bytes.Buffer
	r := newRenderer(&terminal, true)
	r.start("Fetching Remote Config…")
	haltAnimation(r)

	if _, err := r.writeLog(&logs, []byte("INFO fetched project\n")); err != nil {
		t.Fatal(err)
	}
	r.stopProgress()

	if got := logs.String(); got != "INFO fetched project\n" {
		t.Fatalf("log output = %q", got)
	}
	got := ansi.Strip(terminal.String())
	if strings.Count(got, "/ Fetching Remote Config…") != 2 {
		t.Fatalf("terminal output = %q, want progress before and after log", got)
	}
}

func TestDisabledRendererWritesNothing(t *testing.T) {
	var out bytes.Buffer
	r := newRenderer(&out, false)
	r.start("Loading…")
	r.stopProgress()
	if out.Len() != 0 {
		t.Fatalf("disabled progress output = %q", out.String())
	}
}

func TestStopWriterErasesProgressBeforeOutput(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	var terminal bytes.Buffer
	var result bytes.Buffer
	r := newRenderer(&terminal, true)
	r.start("Loading Remote Config…")
	haltAnimation(r)

	if _, err := r.stopWriter(&result).Write([]byte("result\n")); err != nil {
		t.Fatal(err)
	}

	r.mu.Lock()
	active := r.active
	r.mu.Unlock()
	if active {
		t.Fatal("progress remains active after user-facing output")
	}
	if got := result.String(); got != "result\n" {
		t.Fatalf("result output = %q", got)
	}
	if got := terminal.String(); !strings.HasSuffix(got, "\r\x1b[2K") {
		t.Fatalf("terminal output = %q, want final progress erase", got)
	}
}

func TestStopWriterPreservesTerminalFileCapabilities(t *testing.T) {
	wrapped := StopWriter(os.Stderr)
	terminalFile, ok := wrapped.(term.File)
	if !ok {
		t.Fatalf("StopWriter type = %T, want term.File", wrapped)
	}
	if terminalFile.Fd() != os.Stderr.Fd() {
		t.Fatalf("wrapped fd = %d, want stderr fd %d", terminalFile.Fd(), os.Stderr.Fd())
	}
	if _, err := terminalFile.Read(nil); err != nil && err != io.EOF {
		t.Fatalf("wrapped stderr read = %v", err)
	}
	if err := terminalFile.Close(); err != nil {
		t.Fatalf("wrapped close = %v", err)
	}
	if _, err := os.Stderr.Stat(); err != nil {
		t.Fatalf("wrapped close closed stderr: %v", err)
	}
}

func TestStopWriterRemovesOnlyColorForAnyNonEmptyNoColorValue(t *testing.T) {
	for _, value := range []string{"1", "0", "false", "custom", " "} {
		t.Run(value, func(t *testing.T) {
			t.Setenv("NO_COLOR", value)
			var out bytes.Buffer
			r := newRenderer(io.Discard, false)
			input := "\x1b[1;38;2;255;0;0;48;5;22mstyled\x1b[0m"
			if _, err := r.stopWriter(&out).Write([]byte(input)); err != nil {
				t.Fatal(err)
			}

			if got, want := out.String(), "\x1b[1mstyled\x1b[m"; got != want {
				t.Fatalf("filtered output = %q, want %q", got, want)
			}
		})
	}
}

func TestStopWriterKeepsColorWhenNoColorIsEmpty(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	var out bytes.Buffer
	r := newRenderer(io.Discard, false)
	input := "\x1b[38;5;203mstyled\x1b[m"
	if _, err := r.stopWriter(&out).Write([]byte(input)); err != nil {
		t.Fatal(err)
	}
	if got := out.String(); got != input {
		t.Fatalf("output = %q, want unchanged %q", got, input)
	}
}

func TestRendererStylesWholeLineGray(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	var out bytes.Buffer
	r := newRenderer(&out, true)
	r.start("Loading…")
	haltAnimation(r)
	r.stopProgress()

	want := lipgloss.NewStyle().Foreground(clistyles.PaletteSlateDim).Render("/ Loading…")
	if !strings.Contains(out.String(), want) {
		t.Fatalf("progress output = %q, want gray line %q", out.String(), want)
	}
}

func TestFrameIntervalMatchesTUILineSpinner(t *testing.T) {
	if frameInterval.Milliseconds() != 100 {
		t.Fatalf("frame interval = %s, want 100ms", frameInterval)
	}
}

func haltAnimation(r *renderer) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.stop != nil {
		close(r.stop)
		r.stop = nil
	}
	r.generation++
}
