package env

import (
	"os"
	"strings"
)

const (
	LogLevel                = "FBRCM_LOG_LEVEL"
	LogNoTimestamp          = "FBRCM_LOG_NO_TIMESTAMP"
	LogPlain                = "FBRCM_LOG_PLAIN"
	Offline                 = "FBRCM_OFFLINE"
	Profile                 = "FBRCM_PROFILE"
	GoogleAccessToken       = "FBRCM_GOOGLE_ACCESS_TOKEN"
	NoColor                 = "NO_COLOR"
	ConfigDir               = "FBRCM_CONFIG_DIR"
	CacheDir                = "FBRCM_CACHE_DIR"
	Editor                  = "FBRCM_EDITOR"
	NoLocalConfig           = "FBRCM_NO_LOCAL_CONFIG"
	GoogleCloudQuotaProject = "GOOGLE_CLOUD_QUOTA_PROJECT"
	TLSCertFile             = "SSL_CERT_FILE"
	XDGConfigHome           = "XDG_CONFIG_HOME"
)

// LookupNonEmpty returns an environment value without normalization. It is
// suitable for opaque values such as credentials, where trimming would hide
// invalid input or alter the value sent to another service.
func LookupNonEmpty(name string) (string, bool) {
	value, ok := os.LookupEnv(name)
	return value, ok && value != ""
}

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
	value, ok := os.LookupEnv(NoColor)
	return ok && value != ""
}

func LogTimestampDisabled() bool {
	value, ok := os.LookupEnv(LogNoTimestamp)
	return ok && value != ""
}

func LogPlainEnabled() bool {
	_, enabled := LookupNonEmpty(LogPlain)
	return enabled
}
