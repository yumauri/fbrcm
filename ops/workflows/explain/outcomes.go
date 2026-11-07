package explain

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/yumauri/fbrcm/core/rc/evaluate"
)

// The branches are an ordered decision list: "otherwise" means every earlier
// candidate failed. Keeping these guards avoids inventing inverse operators or
// assuming unknown rules on the same input are independent.
func (out *renderer) possibleOutcomes(r report) {
	out.line("╰─ Possible outcomes (first matching branch wins)", out.indent(2))
	indent := out.indent(5)
	first := true
	eligible, _ := outcomeCandidates(r)
	for _, i := range eligible {
		condition := r.Conditions[i]
		outcome := priorityText(condition.Entry.Priority) + " " + conditionName(condition.Entry) +
			" would win " + boldSymbol("→", lipgloss.NewStyle()) + " " + valueText(condition.Value, r.Param.ValueType)
		if condition.Value.IsOpaque() {
			outcome += mutedText(" (requires Firebase-managed evaluation)")
		}
		if condition.Result.State == evaluate.True {
			out.line("╰─ Otherwise: "+outcome, indent)
			return
		}
		rules := condition.Result.RemainingRules()
		if len(rules) == 0 {
			rules = []string{condition.Entry.Expression}
		}
		prefix := "├─ If "
		if !first {
			prefix = "├─ Otherwise, if "
		}
		line := prefix + expressionText(strings.Join(rules, " && ")) + ", then " + outcome
		if condition.Result.Unsupported {
			line += mutedText(" (evaluation unavailable)")
		}
		out.line(line, indent)
		first = false
	}
	fallback := "╰─ Otherwise: default " + boldSymbol("→", lipgloss.NewStyle()) + " " + defaultValue(r.Param)
	if r.Param.DefaultValue != nil && r.Param.DefaultValue.IsOpaque() {
		fallback += mutedText(" (requires Firebase-managed evaluation)")
	}
	out.line(fallback, indent)
}
