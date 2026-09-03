package machine

import (
	"regexp"
	"strings"
)

const MaxSafeTextRunes = 4096

var (
	jsonSecretPattern = regexp.MustCompile(`(?i)("(?:access_token|refresh_token|client_id|client_secret|private_key|password|token)"\s*:\s*")[^"]*(")`)
	bearerPattern     = regexp.MustCompile(`(?i)(bearer\s+)[A-Za-z0-9._~+/=-]+`)
	assignmentSecret  = regexp.MustCompile(`(?i)((?:client_id|token|secret|password|private_key)\s*[=:]\s*)\S+`)
)

// SafeText redacts common credential forms and bounds machine-visible text.
func SafeText(value string) string {
	value = jsonSecretPattern.ReplaceAllString(value, `${1}[REDACTED]${2}`)
	value = bearerPattern.ReplaceAllString(value, `${1}[REDACTED]`)
	value = assignmentSecret.ReplaceAllString(value, `${1}[REDACTED]`)
	runes := []rune(value)
	if len(runes) > MaxSafeTextRunes {
		value = string(runes[:MaxSafeTextRunes]) + "…"
	}
	return value
}

// SafeErrorText returns bounded, redacted text for an optional error.
func SafeErrorText(err error) string {
	if err == nil {
		return ""
	}
	return SafeText(strings.TrimSpace(err.Error()))
}
