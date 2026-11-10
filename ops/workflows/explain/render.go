package explain

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
	"github.com/yumauri/fbrcm/core/rc/evaluate"
	clistyles "github.com/yumauri/fbrcm/internal/terminal/styles"
)

type renderer struct {
	strings.Builder
	width        int
	flowIncoming bool
	flowOutgoing bool
}

func render(r report, width int) string {
	return renderHeader(r, width) + "\n" + renderEvaluation(r, width)
}

func renderHeader(r report, width int) string {
	out := renderer{width: max(width, 1)}
	parameter := parameterName(r.Parameter)
	out.line(r.Project+" / "+parameter, 0)
	out.line(fmt.Sprintf("Template v%s · %s", r.Version, r.Source), 0)
	if r.Exchange != nil {
		label := fmt.Sprintf("Request: %s · entry %d", r.Capture, r.Exchange.Entry)
		if !r.Exchange.StartedAt.IsZero() {
			stamp := r.Exchange.StartedAt
			if location, err := time.LoadLocation(r.Exchange.TimeZone); err == nil && r.Exchange.TimeZone != "" {
				stamp = stamp.In(location)
			}
			label += " · " + stamp.Format("15:04:05Z07:00")
		}
		out.line(label, 0)
	}
	return strings.TrimRight(out.String(), "\n")
}

func renderEvaluation(r report, width int) string {
	out := renderer{width: max(width, 1)}
	parameter := parameterName(r.Parameter)
	flowing := out.width >= 4
	for i, condition := range r.Conditions {
		out.flowIncoming = flowing && i > 0
		out.flowOutgoing = flowing && condition.Result.State == evaluate.False
		out.flowGap()
		out.flowIncoming = flowing
		out.condition(r, i, condition)
		flowing = out.flowOutgoing
	}
	out.flowIncoming = flowing && len(r.Conditions) > 0
	out.flowOutgoing = false
	out.flowGap()
	out.flowIncoming = flowing
	out.defaultValue(r)
	out.line("", 0)
	managed := r.Param.DefaultValue != nil && r.Param.DefaultValue.IsOpaque()
	if r.Winner >= 0 {
		managed = r.Conditions[r.Winner].Value.IsOpaque()
	}
	if r.Unresolved || managed {
		out.line("? "+parameter+" = unknown", 0)
		if managed && !r.Unresolved {
			if r.Winner >= 0 {
				winner := r.Conditions[r.Winner]
				out.line(fmt.Sprintf("├─ Template winner: %s %s", priorityText(winner.Entry.Priority), conditionName(winner.Entry)), 2)
			} else if len(r.Conditions) == 0 {
				out.line("├─ No conditions exist, so use the default value", 2)
			}
			out.line("╰─ The selected value requires Firebase-managed evaluation", 2)
		} else {
			out.possibleOutcomes(r)
		}
	} else {
		value := defaultValue(r.Param)
		if r.Winner >= 0 {
			winner := r.Conditions[r.Winner]
			value = valueText(winner.Value, r.Param.ValueType)
		}
		scope := ""
		if r.Exchange != nil && (len(r.Exchange.Experiments) > 0 || r.Exchange.Managed[r.Parameter] != "") {
			scope = " (template evaluation)"
		}
		out.line(statusSymbol("◆")+" "+parameter+" = "+value+scope, 0)
		if r.Winner >= 0 {
			winner := r.Conditions[r.Winner]
			out.line(fmt.Sprintf("├─ Winner: %s %s", priorityText(winner.Entry.Priority), conditionName(winner.Entry)), 2)
			for i, condition := range r.Conditions {
				if i != r.Winner && condition.Result.State == evaluate.True {
					out.line(fmt.Sprintf("├─ %s %s also matches, but lower priority", priorityText(condition.Entry.Priority), conditionName(condition.Entry)), 2)
				}
			}
			out.line("╰─ Default "+defaultValue(r.Param)+" is not used", 2)
		} else if len(r.Conditions) > 0 {
			out.line("╰─ No condition matches, so use the default value", 2)
		} else {
			out.line("╰─ No conditions exist, so use the default value", 2)
		}
	}
	if r.Assumptions {
		out.line("", 0)
		out.line(mutedText("note:")+" Evaluation assumes the inputs supplied by you or --set", 0)
	}
	if r.Exchange != nil && (r.Exchange.HasResponse || r.Exchange.Status != 0) {
		out.line("", 0)
		switch {
		case r.Exchange.HasResponse:
			if actual, ok := r.Exchange.Entries[r.Parameter]; ok {
				out.line("Captured response: "+styledValue(actual, r.Param.ValueType), 0)
			} else {
				out.line("Captured response: this parameter was not returned", 0)
			}
			if r.Exchange.ResponseState == "EMPTY_CONFIG" || r.Exchange.ResponseState == "NO_TEMPLATE" {
				out.line("╰─ Firebase returned "+r.Exchange.ResponseState, 2)
			}
			if assignment := r.Exchange.Managed[r.Parameter]; assignment != "" {
				out.line("╰─ Firebase assigned this parameter through "+assignment, 2)
			}
			if len(r.Exchange.Experiments) > 0 {
				out.line("", 0)
				out.line(mutedText("note:")+" The capture includes A/B experiment assignments, which may override template conditions", 0)
			}
			if r.Exchange.Version != "" && r.Exchange.Version != r.Version {
				out.line("", 0)
				out.line(mutedText("note:")+fmt.Sprintf(" Capture template v%s differs from evaluated template v%s", r.Exchange.Version, r.Version), 0)
			}
		case r.Exchange.Status == 304:
			out.line("Captured response: HTTP 304 with no new parameter values", 0)
		case r.Exchange.Status != 0:
			out.line(fmt.Sprintf("Captured response: HTTP %d, but parameter values are unavailable", r.Exchange.Status), 0)
		}
	}
	return strings.TrimRight(out.String(), "\n")
}

func (out *renderer) condition(r report, index int, condition conditionResult) {
	symbol, label := "×", rcdisplay.OutcomeDoesNotMatch
	switch condition.Result.State {
	case evaluate.True:
		symbol, label = "✓", rcdisplay.OutcomeLowerPriority
		if index == r.Winner {
			label = rcdisplay.OutcomeSelected
			if r.Unresolved {
				label = "matches · higher-priority condition unresolved"
			}
		}
	case evaluate.Unknown:
		symbol, label = "?", "unresolved"
	}
	out.valueHeader(symbol, priorityText(condition.Entry.Priority)+" "+conditionName(condition.Entry), valueText(condition.Value, r.Param.ValueType), label)
	if out.flowOutgoing {
		body := renderer{width: out.width - 2}
		body.conditionBody(r, index, condition, 0)
		for line := range strings.SplitSeq(strings.TrimSuffix(body.String(), "\n"), "\n") {
			prefix := mutedText("┆")
			if line != "" {
				prefix += " "
			}
			out.line(prefix+line, 0)
		}
		return
	}
	out.conditionBody(r, index, condition, 2)
}

func (out *renderer) conditionBody(r report, index int, condition conditionResult, bodyIndent int) {
	out.line("", 0)
	indent := out.indent(bodyIndent)
	for _, step := range condition.Result.Steps {
		indent = out.step(step, indent)
		out.line(boldSymbol("↓", evaluationStyle(step.Result)), indent)
	}
	noteStyle := clistyles.PanelMuted
	if condition.Result.State == evaluate.True && index == r.Winner && !r.Unresolved {
		noteStyle = lipgloss.NewStyle().Foreground(clistyles.PaletteSuccess)
	}
	out.line(mutedText("eval:")+" "+evaluationText(condition.Result.State), indent)
	switch condition.Result.State {
	case evaluate.False:
		out.conditionNote("This condition does not supply a value", "", noteStyle, bodyIndent)
	case evaluate.Unknown:
		out.conditionNote("Missing or unsupported information prevents a conclusion", "", noteStyle, bodyIndent)
	case evaluate.True:
		if index == r.Winner && !r.Unresolved {
			detail := "Supplies " + parameterName(r.Parameter) + " = " + valueText(condition.Value, r.Param.ValueType)
			if condition.Value.IsOpaque() {
				detail = "The actual value requires Firebase-managed evaluation"
			}
			out.conditionNote("Highest-priority matching condition for this parameter", detail, noteStyle, bodyIndent)
		} else if index != r.Winner && r.Winner >= 0 {
			winner := r.Conditions[r.Winner]
			out.conditionNote(fmt.Sprintf("Matches, but %s %s has higher priority", priorityText(winner.Entry.Priority), conditionName(winner.Entry)),
				"Conditional value "+valueText(condition.Value, r.Param.ValueType)+" is not selected", noteStyle, bodyIndent)
		} else {
			out.conditionNote("Matches, but a higher-priority condition is unresolved",
				"Conditional value "+valueText(condition.Value, r.Param.ValueType)+" may be selected", noteStyle, bodyIndent)
		}
	}
}

func (out *renderer) defaultValue(r report) {
	symbol, label := "✓", rcdisplay.OutcomeSelected
	if r.Winner >= 0 {
		symbol, label = "×", rcdisplay.OutcomeNotUsed
	} else if r.Unresolved {
		symbol, label = "?", "unresolved"
	}
	name := "Default value"
	if !clistyles.NoColorEnabled() {
		name = clistyles.DefaultValueLabel.Render(name)
	}
	out.valueHeader(symbol, name, defaultValue(r.Param), label)
}

func (out *renderer) valueHeader(symbol, name, value, label string) {
	divider := mutedText(strings.Repeat("─", max(out.width, 1)))
	if out.flowIncoming {
		divider = boldSymbol("↓", clistyles.PanelMuted) + " " + mutedText(strings.Repeat("─", out.width-2))
	}
	out.line(divider, 0)
	labelStyle := clistyles.OutcomeNoteStyle(label, clistyles.EvaluationStatusStyle(symbol))
	out.line(fmt.Sprintf("%s %s %s %s · %s", statusSymbol(symbol), name, boldSymbol("→", lipgloss.NewStyle()), value, labelStyle.Render(label)), 0)
}

func (out *renderer) flowGap() {
	text := ""
	if out.flowIncoming {
		text = mutedText("┆")
	}
	out.line(text, 0)
}

func (out *renderer) conditionNote(message, detail string, style lipgloss.Style, indent int) {
	if clistyles.NoColorEnabled() {
		style = lipgloss.NewStyle()
	}
	out.line("", 0)
	out.line(mutedText("note:")+" "+message, indent)
	if detail != "" {
		out.line(style.Render("╰── "+detail), indent+6)
	}
}

func mutedText(value string) string {
	if clistyles.NoColorEnabled() {
		return value
	}
	return clistyles.PanelMuted.Render(value)
}

func mutedItalicText(text string) string {
	if clistyles.NoColorEnabled() {
		return text
	}
	return clistyles.PanelMuted.Italic(true).Render(text)
}

func priorityText(priority int) string {
	return mutedItalicText(fmt.Sprintf("[%d]", priority))
}

func (out *renderer) indent(column int) int {
	return min(column, max(out.width/3, 0))
}

// step maps underline spans onto explicitly wrapped source lines, so pointers
// remain attached to their expression even in narrow terminals.
func (out *renderer) step(step evaluate.Step, indent int) int {
	room := max(out.width-indent, 1)
	plainLines := strings.Split(ansi.Hardwrap(step.Expression, room, true), "\n")
	styledLines := strings.Split(ansi.Hardwrap(stepExpressionText(step), room, true), "\n")
	cursor, pointer := 0, 0
	for i, plain := range plainLines {
		text := plain
		if i < len(styledLines) {
			text = styledLines[i]
		}
		out.line(text, indent)
		begin, end := max(step.Start-cursor, 0), min(step.End-cursor, len(plain))
		if begin < end {
			left := lipgloss.Width(plain[:begin])
			length := max(lipgloss.Width(plain[begin:end]), 1)
			pointer = min(left+2, left+length-1)
			underline := strings.Repeat(" ", left) + strings.Repeat("─", pointer-left) + "┬" + strings.Repeat("─", left+length-pointer-1)
			out.line(evaluationStyle(step.Result).Render(underline), indent)
		}
		cursor += len(plain)
		if cursor < len(step.Expression) && step.Expression[cursor] == '\n' {
			cursor++
		}
	}
	note := evaluationText(step.Result) + " · " + stepNoteText(step.Note)
	return out.annotation(note, indent+pointer, step.Result)
}

// annotation returns the result word's actual display column. The arrow and
// following expression use that same column, including after narrow wrapping.
func (out *renderer) annotation(note string, pointer int, state evaluate.State) int {
	textIndent := out.indent(pointer + 3) // "╰─ " precedes the result.
	connector := strings.Repeat(" ", pointer) + "╰"
	if pointer+1 < out.width {
		connector += "─"
	}
	connector = evaluationStyle(state).Render(connector)
	inline := textIndent == pointer+3
	if !inline {
		// Once the staircase reaches its width budget, place the note below
		// the elbow. Its result, arrow, and next expression still align.
		out.line(connector, 0)
	}
	for i, line := range strings.Split(ansi.Hardwrap(note, max(out.width-textIndent, 1), true), "\n") {
		if i == 0 && inline {
			out.line(connector+" "+line, 0)
		} else {
			out.line(line, textIndent)
		}
	}
	return textIndent
}

func (out *renderer) line(text string, indent int) {
	indent = min(indent, max(out.width-1, 0))
	for line := range strings.SplitSeq(ansi.Hardwrap(text, max(out.width-indent, 1), true), "\n") {
		out.WriteString(strings.Repeat(" ", indent))
		out.WriteString(line)
		out.WriteByte('\n')
	}
}

func statusSymbol(symbol string) string {
	if symbol == "×" || symbol == "✓" {
		return boldSymbol(symbol, clistyles.EvaluationStatusStyle(symbol))
	}
	return clistyles.EvaluationStatusStyle(symbol).Render(symbol)
}

func boldSymbol(symbol string, style lipgloss.Style) string {
	if clistyles.NoColorEnabled() {
		return symbol
	}
	return style.Bold(true).Render(symbol)
}

func parameterName(name string) string {
	if clistyles.NoColorEnabled() {
		return name
	}
	return lipgloss.NewStyle().Foreground(clistyles.PaletteBlueBright).Render(name)
}

func conditionName(entry core.ConditionEntry) string {
	if clistyles.NoColorEnabled() {
		return entry.Name
	}
	return conditionStyle(entry).Render(entry.Name)
}

func conditionStyle(entry core.ConditionEntry) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(clistyles.ConditionLipglossColor(entry.TagColor))
}

func styledValue(value, kind string) string {
	if clistyles.NoColorEnabled() {
		return value
	}
	return clistyles.RemoteConfigValueStyle(value, kind).Render(value)
}

func evaluationText(state evaluate.State) string {
	return evaluationStyle(state).Render(strings.ToUpper(string(state)))
}

func evaluationStyle(state evaluate.State) lipgloss.Style {
	if clistyles.NoColorEnabled() {
		return lipgloss.NewStyle()
	}
	return clistyles.RemoteConfigValueStyle(strings.ToUpper(string(state)), "BOOLEAN")
}

// Only the underlined rule participates in this step. Inactive fragments use
// the shared muted color, including any values or App IDs inside them.
func stepExpressionText(step evaluate.Step) string {
	if clistyles.NoColorEnabled() {
		return step.Expression
	}
	return clistyles.LogFieldKey.Render(step.Expression[:step.Start]) +
		expressionText(step.Expression[step.Start:step.End]) +
		clistyles.LogFieldKey.Render(step.Expression[step.End:])
}

// Quoted values can themselves contain source labels, so consume them whole
// before looking for the provenance added by the evaluator.
var noteSourcePattern = regexp.MustCompile(`'(?:\\.|[^'\\])*'|\((?:request|manual context|command line|your answer|capture timestamp|request locale|request Analytics _fot|response Date header|User-Agent header|Client Hints header|Accept-Language header)(?: app ID)?\)`)

func stepNoteText(value string) string {
	var out strings.Builder
	start := 0
	for _, span := range noteSourcePattern.FindAllStringIndex(value, -1) {
		if value[span[0]] != '(' {
			continue
		}
		out.WriteString(expressionText(value[start:span[0]]))
		out.WriteString(mutedItalicText(strings.ToLower(value[span[0]:span[1]])))
		start = span[1]
	}
	out.WriteString(expressionText(value[start:]))
	return out.String()
}

func expressionText(value string) string {
	if clistyles.NoColorEnabled() {
		return value
	}
	var out strings.Builder
	start := 0
	for _, span := range evaluate.ComparisonValueSpans(value) {
		out.WriteString(clistyles.RenderFirebaseAppPlatforms(value[start:span.Start], lipgloss.NewStyle()))
		style := lipgloss.NewStyle().Foreground(clistyles.PaletteGold)
		if span.Value == "true" || span.Value == "false" {
			style = clistyles.RemoteConfigValueStyle(span.Value, "BOOLEAN")
		}
		out.WriteString(clistyles.RenderFirebaseAppPlatforms(value[span.Start:span.End], style))
		start = span.End
	}
	out.WriteString(clistyles.RenderFirebaseAppPlatforms(value[start:], lipgloss.NewStyle()))
	return out.String()
}

func valueText(value firebase.RemoteConfigValue, kind string) string {
	return styledValue(rcdisplay.FormatSummary(value, kind), kind)
}

func defaultValue(param firebase.RemoteConfigParam) string {
	if param.DefaultValue == nil {
		return "(in-app default value unavailable)"
	}
	return valueText(*param.DefaultValue, param.ValueType)
}
