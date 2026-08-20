package log

import (
	"bytes"
	"io"
	"regexp"
	"strings"
	"testing"

	"github.com/charmbracelet/colorprofile"

	"github.com/yumauri/fbrcm/core/env"
)

func TestCLIOutputsKeepLogsAndTerminalGuidanceSeparate(t *testing.T) {
	t.Setenv(env.LogLevel, "info")
	t.Setenv("NO_COLOR", "1")
	m := newManager()
	m.init(ModeCLI)

	var logs bytes.Buffer
	var terminal bytes.Buffer
	m.configureCLIOutput(&logs, &terminal)
	m.defaultLogger().Info("loading project")
	if !strings.Contains(logs.String(), "loading project") {
		t.Fatalf("log output = %q", logs.String())
	}

	if _, err := io.WriteString(m.terminalWriter(), "authorize here\n"); err != nil {
		t.Fatal(err)
	}
	if got := terminal.String(); got != "authorize here\n" {
		t.Fatalf("terminal output = %q", got)
	}

	before := logs.String()
	m.setLevel(SilentLevel)
	m.defaultLogger().Error("hidden error")
	if got := logs.String(); got != before {
		t.Fatalf("silent log output changed from %q to %q", before, got)
	}
	if _, err := io.WriteString(m.terminalWriter(), "still visible\n"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(terminal.String(), "still visible") {
		t.Fatalf("terminal guidance was suppressed with silent logs: %q", terminal.String())
	}
}

func TestLogTimestampCanBeDisabled(t *testing.T) {
	t.Setenv(env.LogNoTimestamp, "1")
	t.Setenv(env.NoColor, "1")
	m := newManager()
	m.init(ModeCLI)

	var logs bytes.Buffer
	m.configureCLIOutput(&logs, io.Discard)
	m.defaultLogger().Info("deterministic log")
	if regexp.MustCompile(`\d{2}:\d{2}:\d{2}`).MatchString(logs.String()) {
		t.Fatalf("log contains timestamp: %q", logs.String())
	}
	if !strings.Contains(logs.String(), "deterministic log") {
		t.Fatalf("log output = %q", logs.String())
	}
}

func TestNoColorLoggerRemovesColorButKeepsTextDecoration(t *testing.T) {
	t.Setenv(env.NoColor, "custom")
	m := newManager()
	m.init(ModeCLI)

	var logs bytes.Buffer
	m.configureCLIOutput(&logs, io.Discard)
	m.defaultLogger().Info("loaded project", "component", "config")
	got := logs.String()

	var filtered bytes.Buffer
	w := colorprofile.Writer{Forward: &filtered, Profile: colorprofile.ASCII}
	if _, err := w.Write([]byte(got)); err != nil {
		t.Fatal(err)
	}
	if filtered.String() != got {
		t.Fatalf("NO_COLOR log contains color sequences: %q", got)
	}
	if !strings.Contains(got, "\x1b[2m") {
		t.Fatalf("NO_COLOR log lost allowed faint styling: %q", got)
	}
}
