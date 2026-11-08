package evaluate

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"

	regexp "github.com/wasilibs/go-re2"
)

type rule struct {
	field         string
	op            string
	values        []string
	semantic      bool
	numericMethod bool
}

var methods = []string{"exactlyMatches", "contains", "notContains", "matches", "inAtLeastOne", "notInAtLeastOne", "inAll", "notInAll"}

func evalRule(tokens []token, ctx *Context) (State, string, error) {
	if len(tokens) == 1 && !tokens[0].quoted && (tokens[0].text == "true" || tokens[0].text == "false") {
		note := "literal " + tokens[0].text
		if tokens[0].text == "true" {
			note = "Targets all users within this project"
		}
		return State(tokens[0].text), note, nil
	}
	if len(tokens) > 0 && (tokens[0].text == "app.browserAndVersion.inOne" || tokens[0].text == "app.operatingSystemAndVersion.inOne") {
		return evalClientVersion(tokens, ctx)
	}
	r, err := parseRule(tokens)
	if err != nil {
		return Unknown, "Cannot evaluate: " + err.Error(), nil
	}
	if r.field == "app.audiences" {
		return evalAudience(r, ctx)
	}
	if state, known := knownPercentResult(r); known {
		if state == True {
			return state, "Every assigned percentile is at most '100'", nil
		}
		return state, "No assigned percentile is greater than '100'", nil
	}
	need, supported := fieldNeed(r.field)
	if !supported {
		return Unknown, "Unsupported input: " + r.field, nil
	}
	fact, err := ctx.fact(need)
	if err != nil {
		return Unknown, "", err
	}
	if !fact.Known {
		if fact.Reason != "" {
			return Unknown, fact.Reason, nil
		}
		return Unknown, need.Label + " is unavailable", nil
	}
	if fact.Absent {
		return False, need.Label + " is unset (" + fact.Source + ") and undefined inputs do not match", nil
	}
	match, err := compareRule(r, fact.Value)
	if err != nil {
		return Unknown, err.Error(), nil
	}
	state := False
	if match {
		state = True
	}
	if r.field == "device.language" {
		state, err = compareLanguage(fact.Value, r.values)
		if err != nil {
			return Unknown, err.Error(), nil
		}
		match = state == True
	}
	if (r.field == "device.dateTime" || r.field == "app.firstOpenTimestamp") && (!fact.TimeMin.IsZero() || !fact.TimeMax.IsZero()) {
		state, err = compareObservedTime(r, fact)
		if err != nil {
			return Unknown, err.Error(), nil
		}
		match = state == True
	}
	note := fmt.Sprintf("%s = %s (%s) %s", r.field, quote(fact.Value), fact.Source, describeRule(r, match))
	if !fact.TimeMin.IsZero() || !fact.TimeMax.IsZero() {
		note += " throughout the captured fetch-time window"
	}
	return state, note, nil
}

func parseRule(tokens []token) (rule, error) {
	r := rule{}
	if len(tokens) < 3 || tokens[0].quoted {
		return r, fmt.Errorf("unsupported rule syntax")
	}
	i := 0
	if tokens[i].text == "version" {
		if len(tokens) < 6 || tokens[1].text != "(" {
			return r, fmt.Errorf("invalid version comparison")
		}
		r.semantic, i = true, 2
	}
	r.field = canonicalField(tokens[i].text)
	i++
	if i < len(tokens) && tokens[i].text == "[" {
		if r.field != "app.customSignal" && r.field != "app.userProperty" {
			return r, fmt.Errorf("indexed input is not supported for %s", r.field)
		}
		if i+2 >= len(tokens) {
			return r, fmt.Errorf("invalid indexed input")
		}
		name, ok := indexedInputName(tokens[i-1 : i+3])
		if !ok {
			return r, fmt.Errorf("invalid indexed input")
		}
		r.field = IndexedInputKey(r.field, name)
		i += 3
	}
	if r.semantic {
		if i >= len(tokens) || tokens[i].text != ")" {
			return r, fmt.Errorf("invalid version input")
		}
		i++
	}
	methodIndex := i
	if i < len(tokens) && tokens[i].text == "." {
		methodIndex++
	}
	if methodIndex < len(tokens) && slices.Contains(numericOperators, tokens[methodIndex].text) && (methodIndex != i || strings.HasSuffix(r.field, ".")) {
		r.field = strings.TrimSuffix(r.field, ".")
		r.op, r.numericMethod = tokens[methodIndex].text, true
		values, err := comparisonArguments(tokens[methodIndex+1:])
		r.values = values
		return r, err
	}
	for _, method := range methods {
		if field, ok := strings.CutSuffix(r.field, "."+method); ok {
			r.field, r.op = field, method
			break
		}
		if i < len(tokens) && tokens[i].text == "."+method {
			r.op = method
			i++
			break
		}
	}
	if r.field == "percent" && i < len(tokens) && tokens[i].text == "(" {
		if i+2 >= len(tokens) || !tokens[i+1].quoted || tokens[i+2].text != ")" {
			return r, fmt.Errorf("invalid percentile seed")
		}
		seed := tokens[i+1].text
		if len(seed) < 1 || len(seed) > 32 || strings.IndexFunc(seed, func(c rune) bool {
			return !strings.ContainsRune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789._-", c)
		}) >= 0 {
			return r, fmt.Errorf("percentile seed must contain 1 to 32 letters, digits, dots, hyphens, or underscores")
		}
		r.field += "[" + quote(tokens[i+1].text) + "]"
		i += 3
	}
	if r.op != "" {
		if i >= len(tokens) || tokens[i].text != "(" || tokens[len(tokens)-1].text != ")" {
			return r, fmt.Errorf("invalid method arguments")
		}
		values, err := listValues(tokens[i+1 : len(tokens)-1])
		r.values = values
		return r, err
	}
	if i >= len(tokens) {
		return r, fmt.Errorf("missing comparison operator")
	}
	r.op = tokens[i].text
	i++
	if r.op == "between" && (r.field == "percent" || strings.HasPrefix(r.field, "percent[")) {
		if i+3 != len(tokens) || tokens[i+1].text != "and" {
			return r, fmt.Errorf("invalid percentile range")
		}
		r.values = []string{tokens[i].text, tokens[i+2].text}
		return r, nil
	}
	if r.op == "in" {
		values, err := listValues(tokens[i:])
		r.values = values
		return r, err
	}
	if !slices.Contains([]string{"==", "!=", "<", "<=", ">", ">="}, r.op) || i >= len(tokens) {
		return r, fmt.Errorf("unsupported comparison operator %s", r.op)
	}
	if i+1 == len(tokens) {
		r.values = []string{tokens[i].text}
		return r, nil
	}
	if tokens[i].text == "dateTime" || tokens[i].text == "(" {
		value, err := dateLiteral(tokens[i:])
		r.values = []string{value}
		return r, err
	}
	return r, fmt.Errorf("unsupported comparison value")
}

func comparisonArguments(tokens []token) ([]string, error) {
	if len(tokens) < 3 || tokens[0].text != "(" || tokens[len(tokens)-1].text != ")" {
		return nil, fmt.Errorf("invalid numeric method arguments")
	}
	arguments := tokens[1 : len(tokens)-1]
	if len(arguments) == 1 {
		return []string{arguments[0].text}, nil
	}
	values, err := listValues(arguments)
	if err != nil {
		return nil, err
	}
	if len(values) != 1 {
		return nil, fmt.Errorf("numeric comparisons require exactly one target value")
	}
	return values, nil
}

func listValues(tokens []token) ([]string, error) {
	if len(tokens) < 2 || tokens[0].text != "[" || tokens[len(tokens)-1].text != "]" {
		return nil, fmt.Errorf("expected a literal list")
	}
	var values []string
	for i := 1; i < len(tokens)-1; i++ {
		if i%2 == 0 {
			if tokens[i].text != "," {
				return nil, fmt.Errorf("invalid list separator")
			}
			continue
		}
		if !tokens[i].quoted {
			if _, err := strconv.ParseFloat(tokens[i].text, 64); err != nil {
				return nil, fmt.Errorf("unsupported nonliteral list item")
			}
		}
		values = append(values, tokens[i].text)
	}
	if len(tokens) > 2 && (len(tokens)-2)%2 == 0 {
		return nil, fmt.Errorf("incomplete list")
	}
	return values, nil
}

func canonicalField(field string) string {
	if field == "dateTime" {
		return "device.dateTime"
	}
	return field
}

func fieldNeed(field string) (Need, bool) {
	field = canonicalField(field)
	need := Need{Key: field, Label: field, Kind: "text"}
	switch field {
	case "app.id", "app.version", "app.build", "device.os", "device.language", "device.country", "app.firebaseInstallationId":
		return need, true
	case "app.browserAndVersion.name", "app.browserAndVersion.version", "app.operatingSystemAndVersion.name", "app.operatingSystemAndVersion.version":
		return need, true
	case "app.firstOpenTimestamp", "device.dateTime":
		need.Kind = "timestamp"
		return need, true
	case "device.category":
		need.Label = "Firebase device category (for example Mobile)"
		return need, true
	case "percent":
		need.Kind, need.Label = "number", "Firebase-assigned percentile for the default seed (0–100)"
		return need, true
	default:
		if strings.HasPrefix(field, "app.userProperty[") || strings.HasPrefix(field, "app.customSignal[") {
			return need, true
		}
		if strings.HasPrefix(field, "percent[") {
			need.Kind, need.Label = "number", "Firebase-assigned percentile for "+field+" (0–100)"
			return need, true
		}
		return need, false
	}
}

func evalAudience(r rule, ctx *Context) (State, string, error) {
	if strings.HasPrefix(r.op, "not") {
		return evalNegativeAudience(r, ctx)
	}
	all := r.op == "inAll" || r.op == "notInAll"
	if !slices.Contains([]string{"inAll", "notInAll", "inAtLeastOne", "notInAtLeastOne"}, r.op) {
		return Unknown, "unsupported audience operator", nil
	}
	state := False
	if all {
		state = True
	}
	var evidence []string
	for _, audience := range r.values {
		fact, err := ctx.fact(Need{Key: "app.audiences[" + quote(audience) + "]", Label: "Membership in Analytics audience " + quote(audience), Kind: "boolean"})
		if err != nil {
			return Unknown, "", err
		}
		member := Unknown
		if fact.Known && !fact.Absent {
			if fact.Value == "true" || fact.Value == "false" {
				member = State(fact.Value)
			}
		}
		evidence = append(evidence, quote(audience)+" = "+string(member)+" ("+fact.Source+")")
		if all {
			state = allStates(state, member)
		} else {
			state = anyStates(state, member)
		}
		if all && state == False || !all && state == True {
			break
		}
	}
	if strings.HasPrefix(r.op, "not") {
		state = negate(state)
	}
	description := "App " + describeRule(r, state == True)
	if state == Unknown {
		description = "Cannot determine whether the app " + describeRule(r, true)
	}
	return state, description + " · " + strings.Join(evidence, ", "), nil
}

func compareRule(r rule, actual string) (bool, error) {
	if len(r.values) == 0 {
		return false, fmt.Errorf("empty comparison list")
	}
	if r.op == "between" {
		value, verr := finiteNumber(actual)
		lower, lerr := finiteNumber(r.values[0])
		upper, uerr := finiteNumber(r.values[1])
		if verr != nil || lerr != nil || uerr != nil || lower < 0 || upper > 100 || lower >= upper {
			return false, fmt.Errorf("invalid percentile range")
		}
		// Firebase's lower bound is exclusive; its upper bound is inclusive.
		return value > lower && value <= upper, nil
	}
	switch r.op {
	case "exactlyMatches", "contains", "notContains", "matches", "in":
		match := false
		for _, expected := range r.values {
			one := false
			switch r.op {
			case "contains", "notContains":
				one = strings.Contains(actual, expected)
			case "matches":
				expression, err := regexp.Compile(expected)
				if err != nil {
					return false, fmt.Errorf("invalid or unsupported RE2 regular expression")
				}
				one = expression.MatchString(actual)
			default:
				one = actual == expected
				if r.field == "device.country" || r.field == "device.os" {
					one = strings.EqualFold(actual, expected)
				}
				if r.field == "device.language" {
					one = strings.EqualFold(actual, expected)
				}
			}
			match = match || one
		}
		if r.op == "notContains" {
			match = !match
		}
		return match, nil
	}
	expected := r.values[0]
	cmp := 0
	switch {
	case r.field == "device.dateTime" || r.field == "app.firstOpenTimestamp":
		left, lerr := parseTimestamp(actual)
		right, rerr := parseTimestamp(expected)
		if lerr != nil || rerr != nil {
			return false, fmt.Errorf("timestamp must be RFC3339 or an ISO date-time in UTC")
		}
		cmp = left.Compare(right)
	case r.semantic || r.field == "app.version":
		var err error
		cmp, err = compareVersions(actual, expected)
		if err != nil {
			return false, err
		}
	case r.field == "app.id" || r.field == "device.os" || r.field == "app.firebaseInstallationId":
		cmp = strings.Compare(actual, expected)
		if r.field == "device.os" {
			cmp = strings.Compare(strings.ToLower(actual), strings.ToLower(expected))
		}
	default:
		left, lerr := finiteNumber(actual)
		right, rerr := finiteNumber(expected)
		if lerr != nil || rerr != nil {
			return false, fmt.Errorf("numeric comparison requires numeric inputs; use exactlyMatches for strings")
		}
		if left < right {
			cmp = -1
		} else if left > right {
			cmp = 1
		}
	}
	switch r.op {
	case "==":
		return cmp == 0, nil
	case "!=":
		return cmp != 0, nil
	case "<":
		return cmp < 0, nil
	case "<=":
		return cmp <= 0, nil
	case ">":
		return cmp > 0, nil
	case ">=":
		return cmp >= 0, nil
	default:
		return false, fmt.Errorf("unsupported comparison")
	}
}

func finiteNumber(value string) (float64, error) {
	number, err := strconv.ParseFloat(value, 64)
	if err != nil || math.IsNaN(number) || math.IsInf(number, 0) {
		return 0, fmt.Errorf("expected a finite number")
	}
	return number, nil
}

func compareVersions(left, right string) (int, error) {
	l, r := strings.Split(left, "."), strings.Split(right, ".")
	if len(l) > 5 || len(r) > 5 {
		return 0, fmt.Errorf("numeric versions may contain at most five segments")
	}
	cmp := 0
	for i := range max(len(l), len(r)) {
		var a, b uint64
		for _, item := range []struct {
			parts []string
			out   *uint64
		}{{l, &a}, {r, &b}} {
			if i >= len(item.parts) {
				continue
			}
			part := item.parts[i]
			if part == "" || strings.IndexFunc(part, func(c rune) bool { return c < '0' || c > '9' }) >= 0 {
				return 0, fmt.Errorf("version comparison requires numeric dot-separated versions")
			}
			number, err := strconv.ParseUint(part, 10, 64)
			if err != nil {
				return 0, fmt.Errorf("version component exceeds the supported range")
			}
			*item.out = number
		}
		if cmp == 0 {
			if a < b {
				cmp = -1
			} else if a > b {
				cmp = 1
			}
		}
	}
	return cmp, nil
}

func parseTimestamp(value string) (time.Time, error) {
	if timestamp, err := time.Parse(time.RFC3339Nano, value); err == nil {
		return timestamp, nil
	}
	return time.Parse("2006-01-02T15:04:05", value)
}

func dateLiteral(tokens []token) (string, error) {
	if tokens[0].text == "dateTime" {
		tokens = tokens[1:]
	}
	if len(tokens) != 3 && len(tokens) != 5 || tokens[0].text != "(" || tokens[len(tokens)-1].text != ")" || !tokens[1].quoted {
		return "", fmt.Errorf("unsupported date-time literal")
	}
	location := time.UTC
	if len(tokens) == 5 {
		if tokens[2].text != "," || !tokens[3].quoted {
			return "", fmt.Errorf("invalid date-time timezone")
		}
		var err error
		location, err = time.LoadLocation(tokens[3].text)
		if err != nil {
			return "", fmt.Errorf("unknown date-time timezone")
		}
	}
	stamp, err := time.ParseInLocation("2006-01-02T15:04:05", tokens[1].text, location)
	if err != nil {
		return "", fmt.Errorf("invalid ISO date-time literal")
	}
	return stamp.Format(time.RFC3339Nano), nil
}
