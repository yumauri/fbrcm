package evaluate

import (
	"fmt"
	"strings"
	"time"

	"golang.org/x/text/language"
)

func compareLanguage(actual string, expected []string) (State, error) {
	actualTag, err := language.Parse(strings.ReplaceAll(actual, "_", "-"))
	if err != nil {
		return Unknown, fmt.Errorf("request language is not a recognized language tag")
	}
	actualBase, _ := actualTag.Base()
	uncertain := false
	for _, value := range expected {
		target, err := language.Parse(strings.ReplaceAll(value, "_", "-"))
		if err != nil {
			return Unknown, fmt.Errorf("condition language is not a recognized language tag")
		}
		if actualTag == target {
			return True, nil
		}
		targetBase, _ := target.Base()
		if actualBase != targetBase {
			continue
		}
		// Firebase has its own console locale groups. BCP-47 canonical
		// equivalence is certain, but region/script fallback is not.
		uncertain = true
	}
	if uncertain {
		return Unknown, fmt.Errorf("language grouping for request %s and target locales needs Firebase verification", quote(actual))
	}
	return False, nil
}

func compareObservedTime(r rule, fact Fact) (State, error) {
	var lower, upper bool
	var err error
	if !fact.TimeMin.IsZero() {
		lower, err = compareRule(r, fact.TimeMin.Format(time.RFC3339Nano))
		if err != nil {
			return Unknown, err
		}
	}
	if !fact.TimeMax.IsZero() {
		upper, err = compareRule(r, fact.TimeMax.Format(time.RFC3339Nano))
		if err != nil {
			return Unknown, err
		}
	}
	switch r.op {
	case "<", "<=":
		if !fact.TimeMax.IsZero() && upper {
			return True, nil
		}
		if !fact.TimeMin.IsZero() && !lower {
			return False, nil
		}
	case ">", ">=":
		if !fact.TimeMin.IsZero() && lower {
			return True, nil
		}
		if !fact.TimeMax.IsZero() && !upper {
			return False, nil
		}
	}
	return Unknown, fmt.Errorf("the condition boundary overlaps the captured fetch-time window, so Firebase's exact evaluation time is needed")
}

func knownPercentResult(r rule) (State, bool) {
	if r.field != "percent" && !strings.HasPrefix(r.field, "percent[") {
		return Unknown, false
	}
	if len(r.values) == 1 && r.values[0] != "" {
		threshold, err := finiteNumber(r.values[0])
		if err == nil && threshold == 100 {
			if r.op == "<=" {
				return True, true
			}
			if r.op == ">" {
				return False, true
			}
		}
	}
	return Unknown, false
}
