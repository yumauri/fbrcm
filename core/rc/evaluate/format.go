package evaluate

import "strings"

// FormatExpression places each top-level AND clause on its own display line,
// omitting the && separators. Quoted values are preserved;
// expressions the parser cannot split safely are returned unchanged.
func FormatExpression(expression string) string {
	tokens, err := lex(expression)
	if err != nil {
		return expression
	}
	clauses, err := parseClauses(expression, tokens)
	if err != nil || len(clauses) < 2 {
		return expression
	}
	parts := make([]string, len(clauses))
	for i, clause := range clauses {
		parts[i] = expression[clause.start:clause.end]
	}
	return strings.Join(parts, "\n")
}
