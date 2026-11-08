package evaluate

import (
	"encoding/json"
	"fmt"
	"strings"
)

// InputNeed validates and canonicalizes explicit CLI inputs using the same
// names as the Firebase expression grammar.
func InputNeed(field string) (Need, bool) {
	tokens, err := lex(field)
	if err != nil {
		return Need{}, false
	}
	if name, ok := indexedInputName(tokens); ok {
		field = IndexedInputKey(tokens[0].text, name)
		if tokens[0].text == "app.audiences" {
			return Need{Key: field, Label: "Membership in Analytics audience " + quote(name), Kind: "boolean"}, true
		}
	} else if len(tokens) != 1 {
		return Need{}, false
	}
	return fieldNeed(field)
}

// IndexedInputKey builds the canonical key for a named input.
func IndexedInputKey(field, name string) string {
	return field + "[" + quote(name) + "]"
}

// IndexedInputNames finds named inputs in an expression without evaluating the
// condition, so suggestions include rules that would be short-circuited.
func IndexedInputNames(expression, field string) []string {
	tokens, err := lex(expression)
	if err != nil {
		return nil
	}
	var names []string
	for i := 0; i+4 <= len(tokens); i++ {
		if tokens[i].text == field {
			if name, ok := indexedInputName(tokens[i : i+4]); ok {
				names = append(names, name)
			}
		}
	}
	return names
}

func indexedInputName(tokens []token) (string, bool) {
	if len(tokens) != 4 || tokens[0].quoted || tokens[1].quoted || tokens[1].text != "[" ||
		!tokens[2].quoted || tokens[3].quoted || tokens[3].text != "]" {
		return "", false
	}
	return tokens[2].text, true
}

// SplitInput validates the field of a field=value assignment.
func SplitInput(assignment string) (Need, string, bool) {
	field, answer, ok := strings.Cut(assignment, "=")
	need, valid := InputNeed(field)
	return need, answer, ok && valid
}

// Answer uses explicit slash-prefixed markers so unknown, unset, and an empty
// string remain distinct. Literal JSON strings can escape marker values.
func Answer(need Need, answer, source string) (Fact, error) {
	fact := Fact{Known: true, Source: source, Value: answer}
	switch answer {
	case "/unknown":
		fact.Known = false
		fact.Value = ""
		return fact, nil
	case "/unset":
		fact.Absent = true
		fact.Value = ""
		return fact, nil
	}
	if strings.HasPrefix(answer, `"`) {
		if err := json.Unmarshal([]byte(answer), &fact.Value); err != nil {
			return Fact{}, fmt.Errorf("invalid quoted input value")
		}
		answer = fact.Value
	}
	switch need.Kind {
	case "boolean":
		switch strings.ToLower(strings.TrimSpace(answer)) {
		case "true", "yes", "y":
			fact.Value = "true"
		case "false", "no", "n":
			fact.Value = "false"
		default:
			return Fact{}, fmt.Errorf("enter true, false, /unknown, or /unset")
		}
	case "number":
		number, err := finiteNumber(answer)
		if err != nil || number < 0 || number > 100 {
			return Fact{}, fmt.Errorf("enter the Firebase-assigned percentile from 0 to 100, /unknown, or /unset")
		}
	case "timestamp":
		if _, err := parseTimestamp(answer); err != nil {
			return Fact{}, fmt.Errorf("enter an RFC3339 timestamp, /unknown, or /unset")
		}
	}
	return fact, nil
}
