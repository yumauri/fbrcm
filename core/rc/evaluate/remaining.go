package evaluate

import (
	"slices"
	"strconv"
	"strings"
)

// RemainingRules returns the original rules that must still hold for an unknown
// condition to match. It uses the recorded evaluation, without resolving inputs
// again. A false condition has no possible match, even if earlier rules were
// unknown. An unsupported expression remains one opaque requirement.
func (r Result) RemainingRules() []string {
	if r.State != Unknown {
		return nil
	}
	var rules []string
	for _, step := range r.Steps {
		if step.Result == Unknown && step.Start >= 0 && step.End > step.Start && step.End <= len(step.Expression) {
			rules = append(rules, strings.TrimSpace(step.Expression[step.Start:step.End]))
		}
	}
	return rules
}

// Implies reports whether the recorded evaluation proves that matching r also
// matches other. For unknown conditions, every remaining conjunct in other must
// also occur in r. This conservative proof handles identical requirements and
// stricter conjunctions without assuming unrelated unknown rules are independent
// or attempting to solve arbitrary constraints.
func (r Result) Implies(other Result) bool {
	if r.State == False || other.State == True {
		return true
	}
	if r.State != Unknown || other.State != Unknown {
		return false
	}
	required := other.RemainingRules()
	if len(required) == 0 {
		return false
	}
	available := make(map[string]bool)
	for _, rule := range r.RemainingRules() {
		available[requirementKey(rule)] = true
	}
	for _, rule := range required {
		if !available[requirementKey(rule)] {
			return false
		}
	}
	return true
}

// Reuse the expression lexer so whitespace and quote style do not distinguish
// equivalent requirements, while spaces and escapes inside literals still do.
func requirementKey(source string) string {
	tokens, err := lex(source)
	if err != nil {
		return "raw:" + source
	}
	var key strings.Builder
	for i, token := range tokens {
		if token.quoted {
			key.WriteByte('q')
		} else {
			key.WriteByte('t')
			if i == 0 && len(tokens) > 1 && slices.Contains(numericOperators, tokens[1].text) {
				token.text = canonicalField(token.text)
			}
		}
		key.WriteString(strconv.Quote(token.text))
	}
	return key.String()
}
