package evaluate

import "strings"

// The public reference contradicts the negative audience method names. Until
// backend truth tables are verified, only conclude results shared by both
// interpretations (NOT ANY and NOT ALL), rather than silently choosing one.
func evalNegativeAudience(r rule, ctx *Context) (State, string, error) {
	any, all := False, True
	var evidence []string
	for _, audience := range r.values {
		fact, err := ctx.fact(Need{Key: "app.audiences[" + quote(audience) + "]", Label: "Membership in Analytics audience " + quote(audience), Kind: "boolean"})
		if err != nil {
			return Unknown, "", err
		}
		member := Unknown
		if fact.Known && !fact.Absent && (fact.Value == "true" || fact.Value == "false") {
			member = State(fact.Value)
		}
		any, all = anyStates(any, member), allStates(all, member)
		evidence = append(evidence, quote(audience)+" = "+string(member)+" ("+fact.Source+")")
	}
	if any != all || any == Unknown {
		return Unknown, "Negative audience semantics need Firebase verification for mixed or unavailable memberships · " + strings.Join(evidence, ", "), nil
	}
	state := negate(any)
	return state, "App " + describeRule(r, state == True) + " · " + strings.Join(evidence, ", "), nil
}
