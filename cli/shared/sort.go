package shared

import "sort"

// SortedStringKeys returns the sorted keys from a string-keyed map.
func SortedStringKeys[V any](values map[string]V) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
