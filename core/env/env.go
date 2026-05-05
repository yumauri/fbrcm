package env

import (
	"os"
	"strings"
)

const (
	LogLevel      = "FBRCM_LOG_LEVEL"
	NoColor       = "NO_COLOR"
	ConfigDir     = "FBRCM_CONFIG_DIR"
	CacheDir      = "FBRCM_CACHE_DIR"
	XDGConfigHome = "XDG_CONFIG_HOME"
)

func LookupTrimmed(name string) (string, bool) {
	value, ok := os.LookupEnv(name)
	if !ok {
		return "", false
	}

	value = strings.TrimSpace(value)
	if value == "" {
		return "", false
	}

	return value, true
}

func NoColorEnabled() bool {
	_, ok := LookupTrimmed(NoColor)
	return ok
}
