package evaluate

// evaluateClauses records one step per visited rule. Successful AND prefixes
// disappear silently; unresolved rules remain visible as unknown prefixes.
// The first false rule decides the condition without visiting later inputs.
func evaluateClauses(source string, clauses []clause, ctx *Context) (Result, error) {
	result := Result{State: True}
	for i, clause := range clauses {
		result.Rules = append(result.Rules, structuredRule(source, clause, i))
	}
	prefix := ""
	for i, clause := range clauses {
		ctx.reads = nil
		state, note, err := evalRule(clause.tokens, ctx)
		if err != nil {
			return Result{}, err
		}
		result.Rules[i].Status, result.Rules[i].Evaluation, result.Rules[i].Message = "evaluated", &state, note
		result.Rules[i].Inputs = append([]string{}, ctx.reads...)
		result.Rules[i].ReasonCode = ctx.ruleReason(state)
		expression := prefix + source[clause.start:]
		start := len(prefix)
		if i == 0 {
			expression = source
			start = clause.start
		}
		result.Steps = append(result.Steps, Step{
			Expression: expression, Start: start, End: start + clause.end - clause.start,
			Result: state, Note: note,
		})
		result.State = allStates(result.State, state)
		if result.State == False {
			return result, nil
		}
		if state == Unknown {
			prefix += "unknown && "
		}
	}
	return result, nil
}
