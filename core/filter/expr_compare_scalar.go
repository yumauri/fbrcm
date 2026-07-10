package filter

import (
	"fmt"
)

// compare reports whether any contained value satisfies the comparison.
func (v anyValue) compare(right any, compare func(any, any) bool) (bool, error) {
	if rightValues, ok := right.(anyValue); ok {
		for _, leftValue := range v.values {
			for _, rightValue := range rightValues.values {
				matched, err := exprCompareScalar(leftValue, rightValue, compare)
				if err != nil {
					continue
				}
				if matched {
					return true, nil
				}
			}
		}
		return false, nil
	}
	for _, value := range v.values {
		matched, err := exprCompareScalar(value, right, compare)
		if err != nil {
			continue
		}
		if matched {
			return true, nil
		}
	}
	return false, nil
}

func exprValuesCompare(left, right any, compare func(any, any) bool) (bool, error) {
	if leftValues, ok := left.(anyValue); ok {
		return leftValues.compare(right, compare)
	}
	if rightValues, ok := right.(anyValue); ok {
		for _, rightValue := range rightValues.values {
			matched, err := exprCompareScalar(left, rightValue, compare)
			if err != nil {
				continue
			}
			if matched {
				return true, nil
			}
		}
		return false, nil
	}
	matched, err := exprCompareScalar(left, right, compare)
	if err != nil {
		return false, nil
	}
	return matched, nil
}

// exprCompareScalar compares two scalar values and preserves expr runtime errors.
func exprCompareScalar(left, right any, compare func(any, any) bool) (matched bool, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("%v", recovered)
		}
	}()
	return compare(left, right), nil
}
