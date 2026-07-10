// Package strfold provides case-insensitive string ordering helpers shared
// across the core packages.
package strfold

import (
	"slices"
	"strings"
)

// Compare orders strings case-insensitively, breaking ties by case-sensitive
// order so that distinct strings never compare equal.
func Compare(left, right string) int {
	if cmp := CompareFolded(left, right); cmp != 0 {
		return cmp
	}
	switch {
	case left < right:
		return -1
	case left > right:
		return 1
	default:
		return 0
	}
}

// CompareFolded orders strings case-insensitively only.
func CompareFolded(left, right string) int {
	leftFolded := strings.ToLower(left)
	rightFolded := strings.ToLower(right)
	switch {
	case leftFolded < rightFolded:
		return -1
	case leftFolded > rightFolded:
		return 1
	default:
		return 0
	}
}

// Sort sorts values in place using Compare.
func Sort(values []string) {
	slices.SortFunc(values, Compare)
}

// SortedKeys returns the sorted keys from a string-keyed map.
func SortedKeys[V any](values map[string]V) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	Sort(keys)
	return keys
}

// CompareProjects orders two projects by display name (case-insensitive),
// falling back to project ID when the name is empty, then by project ID.
func CompareProjects(leftName, leftID, rightName, rightID string) int {
	leftName = strings.TrimSpace(leftName)
	rightName = strings.TrimSpace(rightName)
	if leftName == "" {
		leftName = leftID
	}
	if rightName == "" {
		rightName = rightID
	}
	if cmp := CompareFolded(leftName, rightName); cmp != 0 {
		return cmp
	}
	return Compare(leftID, rightID)
}

// SortProjects orders projects in place using CompareProjects.
func SortProjects[T any](projects []T, name func(T) string, id func(T) string) {
	slices.SortFunc(projects, func(left, right T) int {
		return CompareProjects(name(left), id(left), name(right), id(right))
	})
}
