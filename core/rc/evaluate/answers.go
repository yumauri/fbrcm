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
	if len(tokens) == 4 && tokens[1].text == "[" && tokens[2].quoted && tokens[3].text == "]" {
		field = tokens[0].text + "[" + quote(tokens[2].text) + "]"
		if tokens[0].text == "app.audiences" {
			return Need{Key: field, Label: "Membership in Analytics audience " + quote(tokens[2].text), Kind: "boolean"}, true
		}
	} else if len(tokens) != 1 {
		return Need{}, false
	}
	return fieldNeed(field)
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
