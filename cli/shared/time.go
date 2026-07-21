package shared

import "time"

// FormatDateTime renders an RFC3339 timestamp in local time for human output.
// Empty and unparseable values are returned unchanged.
func FormatDateTime(value string) string {
	if value == "" {
		return ""
	}

	timestamp, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return value
	}

	return timestamp.Local().Format("2006-01-02 15:04:05")
}
