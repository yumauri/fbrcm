package shared

import (
	"sort"
	"strings"
	"unicode"

	"github.com/yumauri/fbrcm/core/firebase"
)

// ParameterSearch holds prepared search query variants.
type ParameterSearch struct {
	// Raw stores whitespace-collapsed, case-sensitive query for values and expressions.
	Raw string
	// Normalized stores normalized case-insensitive query for names and descriptions.
	Normalized string
}

// NewParameterSearch prepares a parameter search query.
func NewParameterSearch(raw string) ParameterSearch {
	return ParameterSearch{
		Raw:        collapseSpaces(raw),
		Normalized: normalizeSearchText(raw),
	}
}

// Empty reports whether search has no useful query.
func (s ParameterSearch) Empty() bool {
	return s.Raw == "" && s.Normalized == ""
}

// MatchParameterSearch matches a parameter against search query.
func MatchParameterSearch(name string, param firebase.RemoteConfigParam, cfg *firebase.RemoteConfig, search ParameterSearch) bool {
	if search.Empty() {
		return true
	}

	conditionByName := make(map[string]firebase.RemoteConfigCondition)
	if cfg != nil {
		for _, condition := range cfg.Conditions {
			conditionByName[condition.Name] = condition
		}
	}

	conditionNames := sortedStringKeys(param.ConditionalValues)
	normalizedParts := []string{name, param.Description}
	rawParts := make([]string, 0, 1+len(conditionNames)*2)

	if param.DefaultValue != nil {
		rawParts = append(rawParts, param.DefaultValue.Value)
	}
	for _, conditionName := range conditionNames {
		normalizedParts = append(normalizedParts, conditionName)
		rawParts = append(rawParts, param.ConditionalValues[conditionName].Value)
		if condition, ok := conditionByName[conditionName]; ok {
			rawParts = append(rawParts, condition.Expression)
		}
	}

	normalizedBlob := normalizeSearchText(strings.Join(normalizedParts, " "))
	rawBlob := strings.Join(rawParts, " ")
	return search.Normalized != "" && strings.Contains(normalizedBlob, search.Normalized) ||
		search.Raw != "" && strings.Contains(rawBlob, search.Raw)
}

func normalizeSearchText(value string) string {
	var b strings.Builder
	for _, r := range value {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(unicode.ToLower(r))
		default:
			b.WriteRune(' ')
		}
	}
	return collapseSpaces(b.String())
}

func collapseSpaces(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func sortedStringKeys[V any](items map[string]V) []string {
	keys := make([]string, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		left := strings.ToLower(keys[i])
		right := strings.ToLower(keys[j])
		if left == right {
			return keys[i] < keys[j]
		}
		return left < right
	})
	return keys
}
