package log

import (
	"bytes"
	"io"
	"regexp"
	"strings"
	"testing"

	charmlog "charm.land/log/v2"
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

func TestLogLevelLabelsKeepStableCompactWidth(t *testing.T) {
	styles := loggerStyles()
	for _, level := range []charmlog.Level{charmlog.DebugLevel, charmlog.InfoLevel, charmlog.WarnLevel, charmlog.ErrorLevel, charmlog.FatalLevel} {
		style := styles.Levels[level]
		if width, maximum := style.GetWidth(), style.GetMaxWidth(); width != 0 || maximum != 4 {
			t.Errorf("%s level dimensions = width %d, maximum %d; want width 0, maximum 4", level, width, maximum)
		}
	}
}

func TestNoColorLoggerRemovesColorButKeepsTextDecoration(t *testing.T) {
	t.Setenv(env.LogPlain, "")
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

func TestPlainLogsAcrossModes(t *testing.T) {
	for _, mode := range []Mode{ModeCLI, ModeTUI, ModeMCP} {
		for _, noColor := range []string{"", "1"} {
			t.Run(string(mode)+"/NO_COLOR="+noColor, func(t *testing.T) {
				t.Setenv(env.NoColor, noColor)
				plain := "1"
				if mode == ModeMCP {
					plain = "" // MCP stays plain even without the environment opt-in.
				}
				t.Setenv(env.LogPlain, plain)
				t.Setenv("FORCE_COLOR", "1")
				t.Setenv(env.LogLevel, "debug")
				t.Setenv(env.LogNoTimestamp, "1")
				m := newManager()
				var logs bytes.Buffer
				m.configureCLIOutput(&logs, io.Discard)
				m.init(mode)
				logger := m.defaultLogger().With("component", "firebase")
				logger.Debug("http request", "url", "https://example.test/remoteConfig")
				logger.Info("http response", "status", "200 OK")
				logger.Warn("retrying request", "attempt", 2)
				logger.Error("request failed", "detail", "\x1b[31mnetwork failure\x1b[m")
				// Theme refresh and output reconfiguration must not restore ANSI.
				m.logger.SetStyles(loggerStyles())
				m.configureCLIOutput(&logs, io.Discard)
				m.defaultLogger().Info("styles refreshed", "url", "https://example.test")
				got := logs.String()
				if mode == ModeTUI {
					got = strings.Join(m.sink.snapshot(), "\n")
				}
				if strings.Contains(got, "\x1b") {
					t.Fatalf("plain logs contain escape sequences: %q", got)
				}
				for _, text := range []string{"logger initialized", "http request", "http response", "component=firebase", "https://example.test/remoteConfig", "200 OK", "retrying request", "request failed", "network failure", "styles refreshed"} {
					if !strings.Contains(got, text) {
						t.Errorf("plain logs lost %q: %q", text, got)
					}
				}
			})
		}
	}
}

func TestLoggingModeChangesRestoreTerminalStyling(t *testing.T) {
	t.Setenv(env.LogPlain, "")
	t.Setenv(env.NoColor, "")
	t.Setenv(env.LogLevel, "info")
	m := newManager()
	var logs bytes.Buffer
	m.configureCLIOutput(&logs, io.Discard)
	for _, mode := range []Mode{ModeCLI, ModeMCP, ModeTUI, ModeMCP, ModeCLI} {
		logs.Reset()
		m.init(mode)
		m.defaultLogger().Info("loaded project", "component", "config")
		got := logs.String()
		if mode == ModeTUI {
			lines := m.sink.snapshot()
			got = lines[len(lines)-1]
		}
		if styled := strings.Contains(got, "\x1b["); styled != (mode != ModeMCP) {
			t.Errorf("mode %s has unexpected styling: %q", mode, got)
		}
	}
}

func TestTUILogHyperlinksFollowPlainSetting(t *testing.T) {
	t.Setenv(env.NoColor, "")
	t.Setenv(env.LogLevel, "info")
	m := newManager()
	updates, unsubscribe := m.sink.subscribe()
	defer unsubscribe()
	for _, plain := range []string{"1", ""} {
		t.Setenv(env.LogPlain, plain)
		m.init(ModeTUI)
		m.defaultLogger().Info("http response", "url", "https://example.test")
		select {
		case line := <-updates:
			if linked := strings.Contains(line, "\x1b]8;;"); linked != (plain == "") {
				t.Fatalf("unexpected hyperlink styling with FBRCM_LOG_PLAIN=%q: %q", plain, line)
			}
			if plain != "" && strings.Contains(line, "\x1b") {
				t.Fatalf("plain log subscriber received escape sequences: %q", line)
			}
		default:
			t.Fatal("log subscriber received no line")
		}
	}
}
