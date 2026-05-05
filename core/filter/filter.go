package filter

import "strings"

type Mode string

const (
	ModeFuzzy      Mode = "fuzzy"
	ModeStartsWith Mode = "starts-with"
	ModeIncludes   Mode = "includes"
	ModeExact      Mode = "exact"
)

func (m Mode) Label() string {
	switch m {
	case ModeStartsWith:
		return "^"
	case ModeIncludes:
		return "/"
	case ModeExact:
		return "="
	default:
		return "~"
	}
}

func ModeFromLabel(label string) (Mode, bool) {
	switch label {
	case "~":
		return ModeFuzzy, true
	case "^":
		return ModeStartsWith, true
	case "/":
		return ModeIncludes, true
	case "=":
		return ModeExact, true
	default:
		return "", false
	}
}

func Match(value, query string, mode Mode) (bool, []int) {
	if query == "" {
		return true, nil
	}

	switch mode {
	case ModeStartsWith:
		return startsWith(value, query)
	case ModeIncludes:
		return includes(value, query)
	case ModeExact:
		return exact(value, query)
	default:
		return fuzzy(value, query)
	}
}

func fuzzy(value, query string) (bool, []int) {
	valueLower := []rune(strings.ToLower(value))
	queryLower := strings.ToLower(query)

	indices := make([]int, 0, len(queryLower))
	next := 0
	for _, want := range queryLower { // range over string yields runes directly
		found := false
		for next < len(valueLower) {
			if valueLower[next] == want {
				indices = append(indices, next)
				next++
				found = true
				break
			}
			next++
		}
		if !found {
			return false, nil
		}
	}

	return true, indices
}

func startsWith(value, query string) (bool, []int) {
	valueLower := strings.ToLower(value)
	queryLower := strings.ToLower(query)
	if !strings.HasPrefix(valueLower, queryLower) {
		return false, nil
	}
	return true, span(0, len([]rune(query)))
}

func includes(value, query string) (bool, []int) {
	valueLower := []rune(strings.ToLower(value))
	queryLower := []rune(strings.ToLower(query))
	if len(queryLower) > len(valueLower) {
		return false, nil
	}

	for start := 0; start <= len(valueLower)-len(queryLower); start++ {
		matched := true
		for i := range queryLower {
			if valueLower[start+i] != queryLower[i] {
				matched = false
				break
			}
		}
		if matched {
			return true, span(start, len(queryLower))
		}
	}
	return false, nil
}

func exact(value, query string) (bool, []int) {
	if !strings.EqualFold(value, query) {
		return false, nil
	}
	return true, span(0, len([]rune(value)))
}

func span(start, length int) []int {
	indices := make([]int, length)
	for i := range length {
		indices[i] = start + i
	}
	return indices
}
