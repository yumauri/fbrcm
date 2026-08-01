package progress

import (
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"charm.land/bubbles/v2/spinner"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"

	clistyles "github.com/yumauri/fbrcm/cli/styles"
	"github.com/yumauri/fbrcm/core/env"
)

var (
	frameInterval = spinner.Line.FPS
	frames        = [...]string{
		spinner.Line.Frames[1],
		spinner.Line.Frames[2],
		spinner.Line.Frames[3],
		spinner.Line.Frames[0],
	}
)

type renderer struct {
	mu         sync.Mutex
	out        io.Writer
	enabled    bool
	active     bool
	message    string
	frame      int
	generation uint64
	stop       chan struct{}
}

var global = newRenderer(os.Stderr, false)

func newRenderer(out io.Writer, enabled bool) *renderer {
	return &renderer{out: out, enabled: enabled}
}

// Configure selects the terminal used by CLI progress. Progress remains
// disabled when enabled is false, such as when stderr is redirected.
func Configure(out io.Writer, enabled bool) {
	global.stopProgress()
	global.mu.Lock()
	defer global.mu.Unlock()
	global.out = out
	global.enabled = enabled
}

// Start begins progress or replaces the current phase. Calling Start after
// interactive or final output has stopped progress starts a new phase.
func Start(message string) {
	global.start(message)
}

// Stop erases active progress and stops its animation.
func Stop() {
	global.stopProgress()
}

// LogWriter coordinates durable log lines with progress without changing the
// logger's stream: progress is erased, the log is written, then progress is
// redrawn underneath.
func LogWriter(out io.Writer) io.Writer {
	return global.logWriter(out)
}

// StopWriter erases and stops progress before writing user-facing output.
// Use it for stdout data, diagnostics, diffs, and interactive prompts.
func StopWriter(out io.Writer) io.Writer {
	return global.stopWriter(out)
}

func (r *renderer) logWriter(out io.Writer) io.Writer {
	colorAwareOut := colorAwareWriter(out)
	return &coordinatedWriter{
		out: out,
		write: func(p []byte) (int, error) {
			return r.writeLog(colorAwareOut, p)
		},
	}
}

func (r *renderer) stopWriter(out io.Writer) io.Writer {
	colorAwareOut := colorAwareWriter(out)
	return &coordinatedWriter{
		out: out,
		write: func(p []byte) (int, error) {
			r.stopProgress()
			return colorAwareOut.Write(p)
		},
	}
}

func colorAwareWriter(out io.Writer) io.Writer {
	if !env.NoColorEnabled() {
		return out
	}
	return &colorprofile.Writer{Forward: out, Profile: colorprofile.ASCII}
}

func (r *renderer) start(message string) {
	message = strings.TrimSpace(message)
	if message == "" {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.enabled || r.out == nil {
		return
	}

	if r.active {
		r.clearLocked()
		r.message = message
		r.renderLocked()
		return
	}

	r.active = true
	r.message = message
	r.frame = 0
	r.generation++
	generation := r.generation
	stop := make(chan struct{})
	r.stop = stop
	r.renderLocked()
	go r.animate(generation, stop)
}

func (r *renderer) animate(generation uint64, stop <-chan struct{}) {
	ticker := time.NewTicker(frameInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			r.mu.Lock()
			if r.active && r.generation == generation {
				r.clearLocked()
				r.advanceLocked()
				r.renderLocked()
			}
			r.mu.Unlock()
		case <-stop:
			return
		}
	}
}

func (r *renderer) stopProgress() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.active {
		return
	}
	r.clearLocked()
	r.active = false
	r.message = ""
	if r.stop != nil {
		close(r.stop)
		r.stop = nil
	}
}

func (r *renderer) writeLog(out io.Writer, p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.active {
		r.clearLocked()
	}
	n, err := out.Write(p)
	if r.active {
		r.renderLocked()
	}
	return n, err
}

func (r *renderer) advanceLocked() {
	r.frame = (r.frame + 1) % len(frames)
}

func (r *renderer) renderLocked() {
	text := frames[r.frame] + " " + r.message
	if !clistyles.NoColorEnabled() {
		text = lipgloss.NewStyle().Foreground(clistyles.PaletteSlateDim).Render(text)
	}
	_, _ = io.WriteString(r.out, "\r"+text)
}

func (r *renderer) clearLocked() {
	_, _ = io.WriteString(r.out, "\r\x1b[2K")
}

type coordinatedWriter struct {
	out   io.Writer
	write func([]byte) (int, error)
}

func (w *coordinatedWriter) Read(p []byte) (int, error) {
	if reader, ok := w.out.(io.Reader); ok {
		return reader.Read(p)
	}
	return 0, os.ErrInvalid
}

func (w *coordinatedWriter) Write(p []byte) (int, error) {
	return w.write(p)
}

// Close intentionally leaves the underlying standard stream open. This method,
// together with Read and Fd, preserves the terminal-file capabilities expected
// by Bubble Tea when a prompt writes through the coordinated wrapper.
func (w *coordinatedWriter) Close() error {
	return nil
}

func (w *coordinatedWriter) Fd() uintptr {
	if file, ok := w.out.(interface{ Fd() uintptr }); ok {
		return file.Fd()
	}
	return ^uintptr(0)
}
