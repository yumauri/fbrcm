package evaluate

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"unicode"
)

type State string

const (
	True    State = "true"
	False   State = "false"
	Unknown State = "unknown"
)

type Need struct {
	Key   string
	Label string
	Kind  string // text, number, boolean, or timestamp
}

type Resolver func(Need) (Fact, error)

type Step struct {
	Expression string
	Start, End int // byte offsets of the evaluated portion in Expression
	Note       string
	Result     State
}

type Result struct {
	State       State
	Steps       []Step
	Unsupported bool
	Rules       []RuleEvaluation
}

// Context memoizes inputs (including unknown and unset answers) across all
// conditions. A resolver is called lazily, only for inputs actually visited.
type Context struct {
	Facts    map[string]Fact
	Resolve  Resolver
	Capture  *Exchange
	resolved map[string]bool
	Used     map[string]Fact
	reads    []string
}

func (c *Context) fact(need Need) (observed Fact, err error) {
	defer func() {
		if err == nil {
			if c.Used == nil {
				c.Used = make(map[string]Fact)
			}
			c.Used[need.Key] = observed
			if !slices.Contains(c.reads, need.Key) {
				c.reads = append(c.reads, need.Key)
			}
		}
	}()
	if fact, ok := c.Facts[need.Key]; ok {
		if fact.Known || fact.Reason == "" || c.Resolve == nil || c.resolved[need.Key] {
			return fact, nil
		}
		need.Label += " — " + fact.Reason
	} else if fact, ok := c.derive(need); ok {
		if c.Facts == nil {
			c.Facts = make(map[string]Fact)
		}
		c.Facts[need.Key] = fact
		return fact, nil
	}
	fact := Fact{}
	if c.Resolve != nil {
		fact, err = c.Resolve(need)
		if err != nil {
			return Fact{}, err
		}
		if c.resolved == nil {
			c.resolved = make(map[string]bool)
		}
		c.resolved[need.Key] = true
		if !fact.Known && fact.Reason == "" {
			fact.Reason = c.Facts[need.Key].Reason
		}
	}
	if !fact.Known && fact.Reason == "" && (need.Key == "percent" || strings.HasPrefix(need.Key, "percent[")) {
		fact.Reason = "Firebase's client-assigned percentile is not included in the capture and its assignment algorithm is not verified"
	}
	if c.Facts == nil {
		c.Facts = make(map[string]Fact)
	}
	c.Facts[need.Key] = fact
	return fact, nil
}

type token struct {
	text       string
	start, end int
	quoted     bool
}

type clause struct {
	start, end int
	tokens     []token
}

// Evaluate preserves source expressions and records actual short-circuit
// control flow. Unsupported syntax remains unknown instead of guessing false.
func Evaluate(expression string, ctx *Context) (Result, error) {
	tokens, err := lex(expression)
	if err != nil {
		return unsupported(expression, err.Error()), nil
	}
	clauses, err := parseClauses(expression, tokens)
	if err != nil {
		return unsupported(expression, err.Error()), nil
	}
	for _, clause := range clauses {
		if err := validateClause(clause.tokens); err != nil {
			return unsupported(expression, err.Error()), nil
		}
	}
	if ctx == nil {
		ctx = &Context{}
	}
	return evaluateClauses(expression, clauses, ctx)
}

func unsupported(expression, reason string) Result {
	state := Unknown
	return Result{State: Unknown, Steps: []Step{{Expression: expression, End: len(expression), Note: reason, Result: Unknown}}, Unsupported: true,
		Rules: []RuleEvaluation{{ID: "r1", Expression: expression, Span: SourceSpan{End: len(expression)}, Predicate: emptyPredicate("unsupported"), Status: "unsupported", Evaluation: &state, ReasonCode: "unsupported_expression", Message: reason, Inputs: []string{}}}}
}

// parseClauses accepts only top-level AND. Parentheses delimit method/function
// arguments and date-time tuples, never Boolean groups. Validate the entire
// expression before evaluating anything or asking for missing inputs.
func parseClauses(source string, tokens []token) ([]clause, error) {
	if len(tokens) == 0 {
		return nil, fmt.Errorf("expected a Remote Config rule")
	}
	var clauses []clause
	var delimiters []string
	begin := 0
	for i, t := range tokens {
		if t.quoted {
			continue
		}
		switch t.text {
		case "||", "|", "!":
			return nil, fmt.Errorf("unsupported Boolean operator %s; Remote Config rules are joined with &&", t.text)
		case "&&":
			if len(delimiters) != 0 || begin == i {
				return nil, fmt.Errorf("invalid && separator")
			}
			if i+1 >= len(tokens) || t.start == tokens[i-1].end || t.end == tokens[i+1].start {
				return nil, fmt.Errorf("spaces are required before and after &&")
			}
			if strings.TrimSpace(source[tokens[i-1].end:t.start]) != "" || strings.TrimSpace(source[t.end:tokens[i+1].start]) != "" {
				return nil, fmt.Errorf("invalid && separator")
			}
			clauses = append(clauses, clause{start: tokens[begin].start, end: tokens[i-1].end, tokens: tokens[begin:i]})
			begin = i + 1
		case "(", "[":
			if t.text == "(" && i == begin {
				return nil, fmt.Errorf("boolean grouping is not supported in Remote Config conditions")
			}
			delimiters = append(delimiters, t.text)
			if len(delimiters) > 64 {
				return nil, fmt.Errorf("expression nesting exceeds limit")
			}
		case ")", "]":
			if len(delimiters) == 0 {
				return nil, fmt.Errorf("unexpected closing delimiter")
			}
			open := delimiters[len(delimiters)-1]
			if open == "(" && t.text != ")" || open == "[" && t.text != "]" {
				return nil, fmt.Errorf("mismatched expression delimiters")
			}
			delimiters = delimiters[:len(delimiters)-1]
		}
	}
	if len(delimiters) != 0 || begin == len(tokens) {
		return nil, fmt.Errorf("incomplete Remote Config rule")
	}
	clauses = append(clauses, clause{start: tokens[begin].start, end: tokens[len(tokens)-1].end, tokens: tokens[begin:]})
	return clauses, nil
}

func lex(source string) ([]token, error) {
	var tokens []token
	for i := 0; i < len(source); {
		if unicode.IsSpace(rune(source[i])) {
			i++
			continue
		}
		start := i
		t := token{start: start}
		switch c := source[i]; {
		case c == '\'' || c == '"':
			i++
			var value strings.Builder
			for i < len(source) && source[i] != c {
				if source[i] == '\\' {
					r, _, tail, err := strconv.UnquoteChar(source[i:], c)
					if err != nil {
						return nil, fmt.Errorf("invalid quoted string")
					}
					value.WriteRune(r)
					i = len(source) - len(tail)
				} else {
					value.WriteByte(source[i])
					i++
				}
			}
			if i >= len(source) {
				return nil, fmt.Errorf("unterminated quoted string")
			}
			i++
			t.text, t.quoted = value.String(), true
		case strings.ContainsRune("()[],", rune(c)):
			i++
			t.text = source[start:i]
		case strings.ContainsRune("<>=!&|", rune(c)):
			i++
			if i < len(source) && (source[i] == '=' || source[i] == c && (c == '&' || c == '|')) {
				i++
			}
			t.text = source[start:i]
		default:
			for i < len(source) && !unicode.IsSpace(rune(source[i])) && !strings.ContainsRune("()[],<>=!&|'\"", rune(source[i])) {
				i++
			}
			if i == start {
				return nil, fmt.Errorf("unsupported expression token")
			}
			t.text = source[start:i]
		}
		t.end = i
		tokens = append(tokens, t)
		if len(tokens) > 2048 {
			return nil, fmt.Errorf("expression exceeds token limit")
		}
	}
	return tokens, nil
}

func quote(value string) string {
	var out strings.Builder
	out.WriteByte('\'')
	for _, r := range value {
		switch {
		case r == '\\' || r == '\'':
			out.WriteByte('\\')
			out.WriteRune(r)
		case unicode.IsControl(r):
			fmt.Fprintf(&out, `\u%04x`, r)
		default:
			out.WriteRune(r)
		}
	}
	out.WriteByte('\'')
	return out.String()
}
