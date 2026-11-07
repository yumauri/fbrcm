package evaluate

import (
	"fmt"
	"slices"
	"strings"
)

var numericOperators = []string{"<", "<=", "==", "!=", ">=", ">"}
var stringOperators = []string{"contains", "notContains", "exactlyMatches", "matches"}

func validateClause(tokens []token) error {
	if len(tokens) == 1 && !tokens[0].quoted && (tokens[0].text == "true" || tokens[0].text == "false") {
		return nil
	}
	if len(tokens) > 0 && (tokens[0].text == "app.browserAndVersion.inOne" || tokens[0].text == "app.operatingSystemAndVersion.inOne") {
		targets, err := parseClientTargets(tokens)
		if err != nil {
			return err
		}
		if len(targets) == 0 {
			return fmt.Errorf("expected at least one browser or operating-system target")
		}
		for _, target := range targets {
			if target.version != nil {
				if err := validateRule(*target.version); err != nil {
					return err
				}
			}
		}
		return nil
	}
	r, err := parseRule(tokens)
	if err != nil {
		return err
	}
	return validateRule(r)
}

func validateRule(r rule) error {
	if len(r.values) == 0 {
		return fmt.Errorf("expected at least one comparison value")
	}
	if r.numericMethod && (r.field != "app.version" && r.field != "app.build" && !strings.HasPrefix(r.field, "app.userProperty[") && !strings.HasPrefix(r.field, "app.customSignal[")) {
		return fmt.Errorf("numeric method comparisons are not supported for %s", r.field)
	}
	if r.semantic && (!strings.HasPrefix(r.field, "app.customSignal[") || !slices.Contains(numericOperators, r.op)) {
		return fmt.Errorf("version() supports numeric comparison operators on custom signals")
	}
	allowed := false
	switch r.field {
	case "app.id":
		allowed = r.op == "=="
	case "device.os":
		allowed = r.op == "==" || r.op == "!="
	case "device.category":
		allowed = r.op == "contains"
	case "device.country", "device.language", "app.firebaseInstallationId":
		allowed = r.op == "in"
	case "app.version", "app.build":
		allowed = slices.Contains(numericOperators, r.op) || slices.Contains(stringOperators, r.op)
	case "device.dateTime", "app.firstOpenTimestamp":
		allowed = slices.Contains([]string{"<", "<=", ">", ">="}, r.op)
	case "app.audiences":
		allowed = slices.Contains([]string{"inAll", "notInAll", "inAtLeastOne", "notInAtLeastOne"}, r.op)
	default:
		switch {
		case r.field == "percent" || strings.HasPrefix(r.field, "percent["):
			allowed = r.op == "<=" || r.op == ">" || r.op == "between"
			for _, value := range r.values {
				n, err := finiteNumber(value)
				if err != nil || n < 0 || n > 100 {
					return fmt.Errorf("percentile thresholds must be between 0 and 100")
				}
			}
			if r.op == "between" {
				if len(r.values) != 2 {
					return fmt.Errorf("expected two percentile boundaries")
				}
				lower, _ := finiteNumber(r.values[0])
				upper, _ := finiteNumber(r.values[1])
				if lower >= upper {
					return fmt.Errorf("percentile lower boundary must be less than its upper boundary")
				}
			}
		case strings.HasPrefix(r.field, "app.userProperty[") || strings.HasPrefix(r.field, "app.customSignal["):
			allowed = slices.Contains(numericOperators, r.op) || slices.Contains(stringOperators, r.op)
		}
	}
	if !allowed {
		return fmt.Errorf("operator %s is not supported for %s", r.op, r.field)
	}
	return nil
}
