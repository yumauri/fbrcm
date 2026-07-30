package log

import (
	"bytes"
	"io"
	"strings"
	"testing"

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
