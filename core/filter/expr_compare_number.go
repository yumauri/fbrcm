package filter

import (
	"encoding/json"
	"math"
)

// exprStringNumberEqual compares a string number to a typed number.
func exprStringNumberEqual(text string, right float64) bool {
	left, ok := exprParseJSONNumber(text)
	return ok && left == right
}

// exprParseJSONNumber parses Firebase NUMBER values without accepting NaN or Infinity.
func exprParseJSONNumber(text string) (float64, bool) {
	var number float64
	if err := json.Unmarshal([]byte(text), &number); err != nil {
		return 0, false
	}
	if math.IsInf(number, 0) || math.IsNaN(number) {
		return 0, false
	}
	return number, true
}
