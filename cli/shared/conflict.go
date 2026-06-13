package shared

import (
	"errors"
	"strings"
)

// IsRemoteConfigConflict reports whether err looks like an ETag/precondition conflict.
func IsRemoteConfigConflict(err error) bool {
	if err == nil {
		return false
	}

	target := err
	for target != nil {
		msg := strings.ToLower(target.Error())
		if strings.Contains(msg, "returned 412") ||
			strings.Contains(msg, "precondition failed") ||
			strings.Contains(msg, "conditionnotmet") ||
			strings.Contains(msg, "etag") ||
			strings.Contains(msg, "if-match") {
			return true
		}
		target = errors.Unwrap(target)
	}
	return false
}
