package display

import (
	"strings"
	"time"

	"github.com/yumauri/fbrcm/core/firebase"
)

// FormatRemoteConfigAuthor formats the Firebase user attached to a published
// Remote Config version.
func FormatRemoteConfigAuthor(user firebase.RemoteConfigUser) string {
	name, email := strings.TrimSpace(user.Name), strings.TrimSpace(user.Email)
	switch {
	case name != "" && email != "":
		return name + " <" + email + ">"
	case email != "":
		return email
	case name != "":
		return name
	default:
		return "Unknown author"
	}
}

// FormatRemoteConfigVersionTime formats Firebase's RFC3339 timestamp in the
// local timezone, preserving an unrecognized value for diagnostics.
func FormatRemoteConfigVersionTime(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	timestamp, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return value
	}
	return FormatLocalDateTime(timestamp)
}

// FormatRemoteConfigVersionAttribution formats the author and publication time
// consistently across terminal interfaces.
func FormatRemoteConfigVersionAttribution(version firebase.RemoteConfigVersion) string {
	attribution := FormatRemoteConfigAuthor(version.UpdateUser)
	if published := FormatRemoteConfigVersionTime(version.UpdateTime); published != "" {
		attribution += " on " + published
	}
	return attribution
}
