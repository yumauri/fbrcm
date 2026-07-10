package log

import (
	"io"
	"os"
	"strings"
	"sync"

	"charm.land/lipgloss/v2"
	charmlog "charm.land/log/v2"
	"github.com/charmbracelet/colorprofile"

	"github.com/yumauri/fbrcm/core/env"
	corestyles "github.com/yumauri/fbrcm/core/styles"
)

type Mode string

const (
	ModeCLI Mode = "cli"
	ModeTUI Mode = "tui"
)

const SilentLevel charmlog.Level = 42

const (
	urlColor         = "117"
	silentLevelColor = "245"
)

type manager struct {
	mu     sync.RWMutex
	mode   Mode
	level  charmlog.Level
	logger *charmlog.Logger
	sink   *lineSink
}

var global = newManager()

func newManager() *manager {
	sink := newLineSink()
	logger := charmlog.NewWithOptions(io.Discard, charmlog.Options{
		Formatter:       charmlog.TextFormatter,
		Level:           charmlog.InfoLevel,
		ReportTimestamp: true,
		TimeFormat:      "15:04:05",
	})

	return &manager{
		logger: logger,
		sink:   sink,
	}
}

func Init(mode Mode) {
	global.init(mode)
}

func Default() *charmlog.Logger {
	return global.defaultLogger()
}

func For(component string) *charmlog.Logger {
	return Default().With("component", component)
}

func Snapshot() []string {
	return global.sink.snapshot()
}

func Subscribe() (<-chan string, func()) {
	return global.sink.subscribe()
}

func CurrentLevel() charmlog.Level {
	return global.currentLevel()
}

func SetLevel(level charmlog.Level) {
	global.setLevel(level)
}

func AvailableLevels() []charmlog.Level {
	return []charmlog.Level{
		charmlog.DebugLevel,
		charmlog.InfoLevel,
		charmlog.WarnLevel,
		charmlog.ErrorLevel,
		charmlog.FatalLevel,
		SilentLevel,
	}
}

func LevelColor(level charmlog.Level) string {
	if level == SilentLevel {
		return silentLevelColor
	}
	return corestyles.LogLevelColor(level)
}

func (m *manager) init(mode Mode) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.mode = mode
	m.logger.SetFormatter(charmlog.TextFormatter)
	m.logger.SetReportTimestamp(true)
	m.logger.SetTimeFormat("15:04:05")
	m.setLevelLocked(charmlog.InfoLevel)
	if env.NoColorEnabled() {
		m.logger.SetColorProfile(colorprofile.Ascii)
	} else {
		m.logger.SetColorProfile(colorprofile.ANSI256)
	}
	m.logger.SetStyles(loggerStyles())

	if raw, ok := env.LookupTrimmed(env.LogLevel); ok {
		level, err := parseLevel(raw)
		if err != nil {
			m.logger.Warn("invalid log level override; using default", "env", env.LogLevel, "value", raw, "default", charmlog.InfoLevel.String())
		} else {
			m.setLevelLocked(level)
		}
	}

	m.logger.Debug("logger initialized", "mode", mode, "log_level", levelLabel(m.level))
}

func (m *manager) defaultLogger() *charmlog.Logger {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.logger
}

func (m *manager) currentLevel() charmlog.Level {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.level
}

func (m *manager) setLevel(level charmlog.Level) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.setLevelLocked(level)
}

func (m *manager) setLevelLocked(level charmlog.Level) {
	m.level = level
	if level == SilentLevel {
		m.logger.SetLevel(charmlog.FatalLevel)
		m.logger.SetOutput(io.Discard)
		return
	}

	m.logger.SetLevel(level)
	if m.mode == ModeCLI {
		m.logger.SetOutput(os.Stderr)
		return
	}
	m.logger.SetOutput(io.Writer(m.sink))
}

func loggerStyles() *charmlog.Styles {
	styles := charmlog.DefaultStyles()
	styles.Values["url"] = lipgloss.NewStyle().Foreground(lipgloss.Color(urlColor))
	return styles
}

func parseLevel(raw string) (charmlog.Level, error) {
	if strings.EqualFold(strings.TrimSpace(raw), "silent") {
		return SilentLevel, nil
	}
	return charmlog.ParseLevel(raw)
}

func levelLabel(level charmlog.Level) string {
	if level == SilentLevel {
		return "SLNT"
	}
	return strings.ToUpper(level.String())
}
