package filter

import (
	"slices"
	"strings"
)

// exprStringOperator applies a string operator to any string value, skipping non-strings.
func exprStringOperator(left, right any, match func(string, string) bool) bool {
	pattern, ok := right.(string)
	if !ok {
		return false
	}
	return exprStringPredicate(left, func(text string) bool {
		return match(text, pattern)
	})
}

// exprStringPredicate applies a string predicate to any string value, skipping non-strings.
func exprStringPredicate(left any, match func(string) bool) bool {
	if values, ok := left.(anyValue); ok {
		for _, value := range values.values {
			text, ok := value.(string)
			if ok && match(text) {
				return true
			}
		}
		return false
	}
	text, ok := left.(string)
	return ok && match(text)
}

// exprValueTypeMatches reports whether value matches a Firebase value type.
func exprValueTypeMatches(value any, valueType string) bool {
	if values, ok := value.(anyValue); ok {
		return strings.EqualFold(values.valueType, valueType)
	}
	switch strings.ToUpper(strings.TrimSpace(valueType)) {
	case "NUMBER":
		return exprIsNumberScalar(value)
	case "STRING", "JSON":
		_, ok := value.(string)
		return ok
	case "BOOLEAN":
		_, ok := value.(bool)
		return ok
	default:
		return false
	}
}

// exprIsNumberScalar reports whether value is a numeric scalar.
func exprIsNumberScalar(value any) bool {
	switch value.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return true
	default:
		return false
	}
}

// exprValueIsEmpty reports whether value is empty.
func exprValueIsEmpty(value any) bool {
	if values, ok := value.(anyValue); ok {
		return len(values.values) == 0 || slices.ContainsFunc(values.values, exprValueIsEmpty)
	}
	switch typed := value.(type) {
	case nil:
		return true
	case string:
		return strings.TrimSpace(typed) == ""
	case []any:
		return len(typed) == 0
	case map[string]any:
		return len(typed) == 0
	default:
		return false
	}
}
