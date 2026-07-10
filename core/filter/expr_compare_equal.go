package filter

import (
	"strings"

	exprruntime "github.com/expr-lang/expr/vm/runtime"
)

func exprValuesEqual(left, right any) bool {
	if _, ok := left.(rootGroup); ok {
		return right == nil || isRootGroupLabel(right) || exprruntime.Equal(left, right)
	}
	if _, ok := right.(rootGroup); ok {
		return left == nil || isRootGroupLabel(left) || exprruntime.Equal(left, right)
	}
	if leftValues, ok := left.(anyValue); ok {
		return leftValues.equal(right)
	}
	if rightValues, ok := right.(anyValue); ok {
		return rightValues.equal(left)
	}
	if exprruntime.Equal(left, right) {
		return true
	}
	return exprValuesCoerceEqual(left, right) || exprValuesCoerceEqual(right, left)
}

func isRootGroupLabel(value any) bool {
	text, ok := value.(string)
	return ok && text == rootGroupLabel
}

// exprValuesCoerceEqual compares typed values with string literals for backwards compatibility.
func exprValuesCoerceEqual(left, right any) bool {
	text, ok := left.(string)
	if !ok {
		return false
	}
	text = strings.TrimSpace(text)
	switch typed := right.(type) {
	case bool:
		return (text == "true" && typed) || (text == "false" && !typed)
	case int:
		return exprStringNumberEqual(text, float64(typed))
	case int8:
		return exprStringNumberEqual(text, float64(typed))
	case int16:
		return exprStringNumberEqual(text, float64(typed))
	case int32:
		return exprStringNumberEqual(text, float64(typed))
	case int64:
		return exprStringNumberEqual(text, float64(typed))
	case uint:
		return exprStringNumberEqual(text, float64(typed))
	case uint8:
		return exprStringNumberEqual(text, float64(typed))
	case uint16:
		return exprStringNumberEqual(text, float64(typed))
	case uint32:
		return exprStringNumberEqual(text, float64(typed))
	case uint64:
		return exprStringNumberEqual(text, float64(typed))
	case float32:
		return exprStringNumberEqual(text, float64(typed))
	case float64:
		return exprStringNumberEqual(text, typed)
	default:
		return false
	}
}

// equal reports whether any contained value equals right.
func (v anyValue) equal(right any) bool {
	if rightValues, ok := right.(anyValue); ok {
		for _, leftValue := range v.values {
			for _, rightValue := range rightValues.values {
				if exprValuesEqual(leftValue, rightValue) {
					return true
				}
			}
		}
		return false
	}
	for _, value := range v.values {
		if exprValuesEqual(value, right) {
			return true
		}
	}
	return false
}
