package evaluate

import (
	"fmt"
	"slices"
	"strings"
)

// SourceSpan uses zero-based UTF-8 byte offsets in the original expression;
// End is exclusive. These positions never refer to a rendered staircase.
type SourceSpan struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

type VersionComparison struct {
	Operator string   `json:"operator" contract:"enum=<|<=|==|!=|>=|>|contains|notContains|exactlyMatches|matches"`
	Expected []string `json:"expected"`
}

type ClientTarget struct {
	Name    string             `json:"name"`
	Version *VersionComparison `json:"version"`
}

type Predicate struct {
	Kind       string         `json:"kind" contract:"enum=literal|comparison|client_targets|unsupported"`
	Field      *string        `json:"field"`
	Operator   *string        `json:"operator" contract:"enum=<|<=|==|!=|>=|>|contains|notContains|exactlyMatches|matches|in|between|inAtLeastOne|notInAtLeastOne|inAll|notInAll|inOne"`
	Comparison *string        `json:"comparison" contract:"enum=text|number|version|timestamp|percentile|audience|client"`
	Expected   []string       `json:"expected"`
	Literal    *bool          `json:"literal"`
	Targets    []ClientTarget `json:"targets"`
}

type RuleEvaluation struct {
	ID         string     `json:"id"`
	Expression string     `json:"expression"`
	Span       SourceSpan `json:"span"`
	Predicate  Predicate  `json:"predicate"`
	Status     string     `json:"status" contract:"enum=evaluated|skipped|unsupported"`
	Evaluation *State     `json:"evaluation" contract:"enum=true|false|unknown"`
	ReasonCode string     `json:"reason_code" contract:"enum=none|missing_input|input_unavailable|input_absent|comparison_unavailable|unsupported_expression|short_circuit"`
	Message    string     `json:"message"`
	Inputs     []string   `json:"inputs"`
}

func emptyPredicate(kind string) Predicate {
	return Predicate{Kind: kind, Expected: []string{}, Targets: []ClientTarget{}}
}

func clausePredicate(tokens []token) Predicate {
	if len(tokens) == 1 && (tokens[0].text == "true" || tokens[0].text == "false") {
		p := emptyPredicate("literal")
		v := tokens[0].text == "true"
		p.Literal = &v
		return p
	}
	if targets, err := parseClientTargets(tokens); err == nil {
		p := emptyPredicate("client_targets")
		field, op, comparison := strings.TrimSuffix(tokens[0].text, ".inOne"), "inOne", "client"
		p.Field, p.Operator, p.Comparison = &field, &op, &comparison
		for _, target := range targets {
			item := ClientTarget{Name: target.name}
			if target.version != nil {
				item.Version = &VersionComparison{Operator: target.version.op, Expected: append([]string{}, target.version.values...)}
			}
			p.Targets = append(p.Targets, item)
		}
		return p
	}
	r, err := parseRule(tokens)
	if err != nil {
		return emptyPredicate("unsupported")
	}
	p := emptyPredicate("comparison")
	p.Field, p.Operator = &r.field, &r.op
	p.Expected = append([]string{}, r.values...)
	comparison := "text"
	switch {
	case r.field == "device.dateTime" || r.field == "app.firstOpenTimestamp":
		comparison = "timestamp"
	case r.field == "percent" || strings.HasPrefix(r.field, "percent["):
		comparison = "percentile"
	case r.field == "app.audiences":
		comparison = "audience"
	case slices.Contains(numericOperators, r.op) && (r.semantic || r.field == "app.version"):
		comparison = "version"
	case slices.Contains(numericOperators, r.op) && r.field != "app.id" && r.field != "device.os":
		comparison = "number"
	}
	p.Comparison = &comparison
	return p
}

func structuredRule(source string, c clause, index int) RuleEvaluation {
	return RuleEvaluation{ID: fmt.Sprintf("r%d", index+1), Expression: source[c.start:c.end], Span: SourceSpan{Start: c.start, End: c.end}, Predicate: clausePredicate(c.tokens), Status: "skipped", ReasonCode: "short_circuit", Inputs: []string{}}
}

func (c *Context) ruleReason(state State) string {
	if state == Unknown {
		for _, key := range c.reads {
			fact := c.Used[key]
			if !fact.Known {
				if fact.Reason != "" {
					return "input_unavailable"
				}
				return "missing_input"
			}
		}
		return "comparison_unavailable"
	}
	if state == False {
		for _, key := range c.reads {
			if c.Used[key].Absent {
				return "input_absent"
			}
		}
	}
	return "none"
}
