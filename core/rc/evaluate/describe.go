package evaluate

import "strings"

func describeRule(r rule, matched bool) string {
	values := make([]string, len(r.values))
	for i, value := range r.values {
		values[i] = quote(value)
	}
	list := "[" + strings.Join(values, ", ") + "]"
	choose := func(yes, no string) string {
		if matched {
			return yes
		}
		return no
	}
	switch r.op {
	case "==":
		return choose("equals ", "does not equal ") + values[0]
	case "!=":
		return choose("does not equal ", "equals ") + values[0]
	case "<", "<=", ">", ">=":
		phrases := map[string][2]string{
			"<": {"is less than ", "is greater than or equal to "}, "<=": {"is less than or equal to ", "is greater than "},
			">": {"is greater than ", "is less than or equal to "}, ">=": {"is greater than or equal to ", "is less than "},
		}
		if r.field == "device.dateTime" || r.field == "app.firstOpenTimestamp" {
			phrases = map[string][2]string{
				"<": {"is before ", "is on or after "}, "<=": {"is on or before ", "is after "},
				">": {"is after ", "is on or before "}, ">=": {"is on or after ", "is before "},
			}
		}
		phrase := phrases[r.op]
		return choose(phrase[0], phrase[1]) + values[0]
	case "exactlyMatches":
		if len(values) == 1 {
			return choose("exactly matches ", "does not exactly match ") + values[0]
		}
		return choose("exactly matches one of ", "does not exactly match any of ") + list
	case "contains", "notContains":
		if r.op == "notContains" {
			matched = !matched
		}
		if len(values) == 1 {
			return choose("contains ", "does not contain ") + values[0]
		}
		return choose("contains at least one of ", "does not contain any of ") + list
	case "matches":
		if len(values) == 1 {
			return choose("matches the regular expression ", "does not match the regular expression ") + values[0]
		}
		return choose("matches at least one regular expression in ", "does not match any regular expression in ") + list
	case "in":
		return choose("is in ", "is not in ") + list
	case "between":
		return choose("is within the range from ", "is outside the range from ") + values[0] + " (exclusive) to " + values[1] + " (inclusive)"
	case "inAtLeastOne", "notInAtLeastOne":
		if r.op == "notInAtLeastOne" {
			matched = !matched
		}
		return choose("belongs to at least one of the audiences ", "does not belong to any of the audiences ") + list
	case "inAll", "notInAll":
		if r.op == "notInAll" {
			matched = !matched
		}
		return choose("belongs to all of the audiences ", "does not belong to all of the audiences ") + list
	default:
		return choose("satisfies the comparison", "does not satisfy the comparison")
	}
}
