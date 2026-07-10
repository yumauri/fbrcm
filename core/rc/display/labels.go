package display

import "strings"

// FormatProject formats a project name and ID for display.
func FormatProject(name, projectID string) string {
	if strings.TrimSpace(name) == "" {
		return projectID
	}
	return name + " (" + projectID + ")"
}

// FormatConditionLabel formats a condition slot label for display.
func FormatConditionLabel(label string) string {
	if label == "default" {
		return "Default value"
	}
	return label
}

// FormatParameterHeader formats a parameter key with its group when present.
func FormatParameterHeader(key, group string) string {
	if group == "" {
		return key
	}
	return key + " [" + group + "]"
}
