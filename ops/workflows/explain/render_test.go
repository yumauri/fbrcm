package explain

import (
	"fmt"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	charmlog "charm.land/log/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/core/rc/evaluate"
	clistyles "github.com/yumauri/fbrcm/internal/terminal/styles"
)

func exampleReport(t *testing.T) report {
	t.Helper()
	r := report{Project: "production", Parameter: "new_player", Version: "142", Source: "cache", Winner: 2, Param: firebase.RemoteConfigParam{ValueType: "BOOLEAN", DefaultValue: &firebase.RemoteConfigValue{Value: "false"}}}
	ctx := &evaluate.Context{Facts: map[string]evaluate.Fact{
		"app.id": {Known: true, Value: appID, Source: "request"}, "app.version": {Known: true, Value: "3.8.0", Source: "request"}, "device.language": {Known: true, Value: "en-US", Source: "request"},
	}}
	for _, item := range []struct {
		priority                       int
		name, expression, value, color string
	}{
		{8, "ios_beta", `app.id == '1:123456789:ios:abcdef' && app.audiences.inAtLeastOne(['Beta testers'])`, "true", "PURPLE"},
		{16, "android_v4", `app.id == '` + appID + `' && app.version.exactlyMatches(['4.0.0'])`, "true", "ORANGE"},
		{23, "android_english", `app.id == '` + appID + `' && device.language in ['en-US']`, "true", "GREEN"},
		{31, "android_users", `app.id == '` + appID + `'`, "false", "CYAN"},
	} {
		result, err := evaluate.Evaluate(item.expression, ctx)
		if err != nil {
			t.Fatal(err)
		}
		r.Conditions = append(r.Conditions, conditionResult{Entry: core.ConditionEntry{Priority: item.priority, Name: item.name, Expression: item.expression, TagColor: item.color}, Value: firebase.RemoteConfigValue{Value: item.value}, Result: result})
	}
	return r
}

func TestRenderNaturalAndNarrowTerminals(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	r := exampleReport(t)
	for _, width := range []int{200, 80, 48, 24, 12} {
		t.Run(fmt.Sprint(width), func(t *testing.T) {
			got := render(r, width)
			for line := range strings.SplitSeq(got, "\n") {
				if lipgloss.Width(line) > width {
					t.Fatalf("line overflows %d columns: %q", width, line)
				}
			}
			if strings.ContainsAny(got, "\x1b…") {
				t.Fatal("NO_COLOR or complete-expression output was not preserved")
			}
			// Wrapping changes layout, never the original expression content.
			joined := strings.Join(strings.Fields(got), "")
			for _, token := range []string{"inAtLeastOne(['Beta", "exactlyMatches(['4.0.0'])", "en-US", "eval:"} {
				// Underlines and arrows can interrupt wrapped words, so only
				// assert contiguous content where the natural width fits.
				if width == 200 && !strings.Contains(joined, token) {
					t.Fatalf("missing expression content %q:\n%s", token, got)
				}
			}
			if width == 200 {
				t.Log("\n" + got)
			}
		})
	}
}

func TestRendererReusesExistingValuePlatformAndConditionColors(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	r := exampleReport(t)
	r.Assumptions = true
	r.Exchange = &evaluate.Exchange{HasResponse: true, Entries: map[string]string{"new_player": "true"}, Version: "141", Experiments: []string{"e1"}}
	got := render(r, 200)
	for _, text := range []string{"×", "does not match"} {
		style := lipgloss.NewStyle().Foreground(clistyles.PaletteError)
		if text == "×" {
			style = style.Bold(true)
		}
		if !strings.Contains(got, style.Render(text)) {
			t.Fatalf("missing red status for %q", text)
		}
	}
	for _, text := range []string{"✓", "lower priority"} {
		style := lipgloss.NewStyle().Foreground(clistyles.PaletteSuccess)
		if text == "✓" {
			style = style.Bold(true)
		}
		if !strings.Contains(got, style.Render(text)) {
			t.Fatalf("missing green status for %q", text)
		}
	}
	selected := lipgloss.NewStyle().Foreground(clistyles.PaletteSlateBright).Background(clistyles.ColorHistoryAddedBackground).Render("selected")
	if strings.Count(got, selected) != 1 {
		t.Fatalf("selected status must use light text on a green background exactly once:\n%q", got)
	}
	for _, item := range []struct {
		symbol string
		styles []lipgloss.Style
	}{
		{"×", []lipgloss.Style{lipgloss.NewStyle().Foreground(clistyles.PaletteError)}},
		{"✓", []lipgloss.Style{lipgloss.NewStyle().Foreground(clistyles.PaletteSuccess)}},
		{"→", []lipgloss.Style{lipgloss.NewStyle()}},
		{"↓", []lipgloss.Style{clistyles.PanelMuted, evaluationStyle(evaluate.True), evaluationStyle(evaluate.False)}},
	} {
		boldCount := 0
		for _, style := range item.styles {
			boldCount += strings.Count(got, style.Bold(true).Render(item.symbol))
		}
		if boldCount != strings.Count(ansi.Strip(got), item.symbol) {
			t.Fatalf("not every %s symbol is bold", item.symbol)
		}
	}
	for _, boolean := range []string{"true", "false"} {
		want := clistyles.RemoteConfigValueStyle(boolean, "BOOLEAN").Render(boolean)
		if !strings.Contains(got, want) {
			t.Fatalf("missing shared value color for %s", boolean)
		}
		upper := strings.ToUpper(boolean)
		if !strings.Contains(got, clistyles.RemoteConfigValueStyle(upper, "BOOLEAN").Render(upper)) {
			t.Fatalf("missing shared evaluation color for %s", upper)
		}
	}
	for _, condition := range r.Conditions {
		want := lipgloss.NewStyle().Foreground(clistyles.ConditionLipglossColor(condition.Entry.TagColor)).Render(condition.Entry.Name)
		if !strings.Contains(got, want) {
			t.Fatalf("missing shared condition color for %s", condition.Entry.Name)
		}
		priority := fmt.Sprintf("[%d]", condition.Entry.Priority)
		styledPriority := clistyles.PanelMuted.Italic(true).Render(priority)
		if strings.Count(got, styledPriority) != strings.Count(ansi.Strip(got), priority) {
			t.Fatalf("not every %s priority is gray and italic:\n%q", priority, got)
		}
	}
	for _, platform := range []core.AppPlatform{core.AppPlatformIOS, core.AppPlatformAndroid} {
		want := lipgloss.NewStyle().Foreground(clistyles.FirebaseAppPlatformColor(platform)).Render(string(platform))
		if !strings.Contains(got, want) {
			t.Fatalf("missing shared app ID color for %s", platform)
		}
	}
	if plain := ansi.Strip(got); !strings.Contains(plain, "◆ new_player = true") {
		t.Fatalf("coloring changed content:\n%s", plain)
	}
	key := lipgloss.NewStyle().Foreground(clistyles.PaletteBlueBright).Render(r.Parameter)
	if strings.Count(got, key) != strings.Count(ansi.Strip(got), r.Parameter) {
		t.Fatalf("not every parameter name uses the shared Key color:\n%q", got)
	}
	if !strings.Contains(got, clistyles.PanelMuted.Italic(true).Render("(request)")) {
		t.Fatalf("value source is not gray and italic:\n%q", got)
	}
	for _, label := range []string{"note:", "eval:"} {
		if strings.Count(got, clistyles.PanelMuted.Render(label)) != strings.Count(ansi.Strip(got), label) {
			t.Fatalf("not every %s label is gray:\n%q", label, got)
		}
	}
}

func TestConditionHeaderHasFullWidthGrayDividerAndOnlySelectedBackground(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	r := exampleReport(t)
	for _, width := range []int{160, 24, 12} {
		for _, index := range []int{0, 2} {
			condition := r.Conditions[index]
			condition.Entry.Name += " · " + strings.Repeat("条件", 10)
			out := renderer{width: width}
			out.condition(r, index, condition)
			header, _, _ := strings.Cut(out.String(), "\n\n")
			divider, heading, _ := strings.Cut(header, "\n")
			if divider != clistyles.PanelMuted.Render(strings.Repeat("─", width)) {
				t.Fatalf("missing full-width gray divider: %q", header)
			}
			symbol := "×"
			if condition.Result.State == evaluate.True {
				symbol = "✓"
			}
			if !strings.HasPrefix(heading, statusSymbol(symbol)+" "+priorityText(condition.Entry.Priority)) {
				t.Fatalf("header retains the padded accent: %q", header)
			}
			selected := lipgloss.NewStyle().Foreground(clistyles.PaletteSlateBright).Background(clistyles.ColorHistoryAddedBackground).Render("selected")
			headerWithoutSelected := strings.ReplaceAll(header, selected, "selected")
			for _, background := range []lipgloss.Style{lipgloss.NewStyle().Background(clistyles.ColorHistoryAddedBackground), lipgloss.NewStyle().Background(clistyles.ColorHistoryRemovedBackground)} {
				code := strings.TrimSuffix(strings.TrimPrefix(background.Render("x"), "\x1b["), "mx\x1b[m")
				if strings.Contains(headerWithoutSelected, code) {
					t.Fatalf("header has a background outside the selected word: %q", header)
				}
			}
			for line := range strings.SplitSeq(header, "\n") {
				if lipgloss.Width(line) > width {
					t.Fatalf("header exceeds %d columns: %q", width, line)
				}
			}
		}
	}
}

func TestStepNoteSourceLabelsInsideValuesAreNotStyledAsSources(t *testing.T) {
	value := `app.customSignal['name'] = 'client\'s (request)' (command line) exactlyMatches 'client\'s (request)'`
	t.Setenv("NO_COLOR", "")
	literal := lipgloss.NewStyle().Foreground(clistyles.PaletteGold).Render(`'client\'s (request)'`)
	want := `app.customSignal['name'] = ` + literal + " " +
		clistyles.PanelMuted.Italic(true).Render("(command line)") + ` exactlyMatches ` + literal
	if got := stepNoteText(value); got != want || ansi.Strip(got) != value {
		t.Fatalf("source styling changed quoted values:\ngot %q\nwant %q", got, want)
	}
	t.Setenv("NO_COLOR", "1")
	if got := stepNoteText(value); got != value {
		t.Fatalf("NO_COLOR changed annotation text: %q", got)
	}
}

func TestInAppAndManagedValuesAreNotInvented(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	r := report{Project: "demo", Parameter: "flag", Winner: -1}
	if got := render(r, 80); !strings.Contains(got, "in-app default value unavailable") {
		t.Fatalf("invented missing default:\n%s", got)
	}
	r.Param.DefaultValue = &firebase.RemoteConfigValue{UseInAppDefault: true}
	if got := render(r, 80); !strings.Contains(got, "(in-app default)") {
		t.Fatalf("lost explicit in-app default:\n%s", got)
	}
}

func TestDefaultValueFollowsConditionsInEvaluationFlow(t *testing.T) {
	for _, tc := range []struct {
		name         string
		winner       int
		unresolved   bool
		noConditions bool
		header       string
	}{
		{"matching condition", 2, false, false, "× Default value → false · not used"},
		{"no matches", -1, false, false, "✓ Default value → false · selected"},
		{"no conditions", -1, false, true, "✓ Default value → false · selected"},
		{"unresolved without match", -1, true, false, "? Default value → false · unresolved"},
		{"unresolved with match", 2, true, false, "× Default value → false · not used"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := exampleReport(t)
			r.Winner, r.Unresolved = tc.winner, tc.unresolved
			if tc.noConditions {
				r.Conditions = nil
			} else if tc.winner < 0 {
				for i := range r.Conditions {
					r.Conditions[i].Result.State = evaluate.False
				}
				if tc.unresolved {
					r.Conditions[0].Result.State = evaluate.Unknown
				}
			}
			for _, noColor := range []string{"", "1"} {
				t.Setenv("NO_COLOR", noColor)
				got := render(r, 200)
				plain := ansi.Strip(got)
				if tc.noConditions && !strings.Contains(plain, "◆ new_player = false\n  ╰─ No conditions exist, so use the default value") {
					t.Fatalf("final result does not explain the absence of conditions:\n%s", plain)
				}
				divider := strings.Repeat("─", 200)
				if tc.winner < 0 && !tc.unresolved {
					divider = "↓ " + strings.Repeat("─", 198)
				}
				if strings.Count(plain, "Default value") != 1 || !strings.Contains(plain, divider+"\n"+tc.header+"\n\n") {
					t.Fatalf("default is not a separate flow header:\n%s", plain)
				}
				position := strings.Index(plain, tc.header)
				for _, condition := range r.Conditions {
					if strings.LastIndex(plain[:position], condition.Entry.Expression) < 0 {
						t.Fatalf("default appears before condition %s:\n%s", condition.Entry.Name, plain)
					}
				}
				if noColor != "" {
					if got != plain {
						t.Fatal("NO_COLOR output contains styling")
					}
					continue
				}
				if !strings.Contains(got, clistyles.DefaultValueLabel.Render("Default value")) {
					t.Fatal("default label does not reuse the get table style")
				}
				switch {
				case tc.winner >= 0:
					if !strings.Contains(got, clistyles.EvaluationStatusStyle("×").Render("not used")) {
						t.Fatal("unused default status is not red")
					}
				case !tc.unresolved:
					selected := lipgloss.NewStyle().Foreground(clistyles.PaletteSlateBright).Background(clistyles.ColorHistoryAddedBackground).Render("selected")
					if strings.Count(got, selected) != 1 {
						t.Fatal("selected default does not use the shared green highlight")
					}
				}
			}
		})
	}
}

func TestStaircaseIndentationAdvancesByStepRatherThanTreeDepth(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	expression := `app.customSignal['app_name'].notContains(['sandbox']) && app.customSignal['app_name'].notContains(['demo']) && app.customSignal['environment'].exactlyMatches(['by-dev'])`
	result, err := evaluate.Evaluate(expression, &evaluate.Context{Facts: map[string]evaluate.Fact{
		"app.customSignal['app_name']":    {Known: true, Value: "setplex", Source: "request"},
		"app.customSignal['environment']": {Known: true, Value: "by-dev", Source: "request"},
	}})
	if err != nil {
		t.Fatal(err)
	}
	for _, width := range []int{320, 48} {
		out := renderer{width: width}
		out.condition(report{Param: firebase.RemoteConfigParam{ValueType: "BOOLEAN"}, Winner: 0}, 0, conditionResult{
			Entry: core.ConditionEntry{Priority: 69, Name: "Tria Development"}, Value: firebase.RemoteConfigValue{Value: "true"}, Result: result,
		})
		if strings.Contains(out.String(), "Both sides must be true.") || strings.Contains(out.String(), "true && true") || strings.Contains(out.String(), "Left side") {
			t.Fatalf("rendered a redundant conjunction reduction:\n%s", out.String())
		}
		lines := strings.Split(out.String(), "\n")
		step, lastIndent := 0, -1
		for i, line := range lines {
			if strings.TrimSpace(line) != "↓" {
				continue
			}
			step++
			indent := len(line) - len(strings.TrimLeft(line, " "))
			if indent <= lastIndent && width == 320 || indent < lastIndent {
				t.Fatalf("width %d staircase moved left or stopped too early at step %d:\n%s", width, step, out.String())
			}
			if i+1 >= len(lines) || len(lines[i+1])-len(strings.TrimLeft(lines[i+1], " ")) != indent {
				t.Fatalf("connector does not align with the next expression:\n%s", out.String())
			}
			lastIndent = indent
		}
		if step != len(result.Steps) {
			t.Fatalf("width %d: got %d connectors for %d evaluation steps", width, step, len(result.Steps))
		}
		if width == 320 {
			t.Log("\n" + out.String())
		}
	}
}

func TestConjunctionTraceEndsImmediatelyAfterFirstFalseRule(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	expression := `app.customSignal['app_name'].notContains(['sandbox']) && app.customSignal['app_name'].notContains(['demo']) && app.customSignal['environment'].exactlyMatches(['by-dev'])`
	result, err := evaluate.Evaluate(expression, &evaluate.Context{Facts: map[string]evaluate.Fact{
		"app.customSignal['app_name']": {Known: true, Value: "sandbox", Source: "request"},
	}})
	if err != nil || result.State != evaluate.False {
		t.Fatalf("evaluation = %#v, %v", result, err)
	}
	out := renderer{width: 320}
	out.condition(report{Param: firebase.RemoteConfigParam{ValueType: "BOOLEAN"}, Winner: -1}, 0, conditionResult{
		Entry: core.ConditionEntry{Priority: 69, Name: "Tria Development"}, Value: firebase.RemoteConfigValue{Value: "true"}, Result: result,
	})
	if strings.Contains(out.String(), "skip the rest") || strings.Contains(out.String(), "Left side") || strings.Count(out.String(), "↓") != 1 {
		t.Fatalf("repeated short-circuit reductions:\n%s", out.String())
	}
	if strings.Contains(out.String(), "false && app.customSignal['environment']") || !strings.Contains(out.String(), "eval: FALSE") {
		t.Fatalf("continued evaluating a decided conjunction:\n%s", out.String())
	}
}

func TestShortAnnotationsAlignArrowWithWrappedResult(t *testing.T) {
	for _, test := range []struct {
		name                   string
		width, pointer, result int
	}{
		{"inline", 24, 4, 7},
		{"indented", 48, 9, 12},
		{"capped indent", 12, 6, 4},
		{"pointer at terminal edge", 12, 11, 4},
	} {
		t.Run(test.name, func(t *testing.T) {
			out := renderer{width: test.width}
			resultIndent := out.annotation("TRUE · a long explanation that wraps onto several lines", test.pointer, evaluate.True)
			out.line("↓", resultIndent)
			lines := strings.Split(strings.TrimRight(out.String(), "\n"), "\n")
			if resultIndent != test.result || lines[len(lines)-1] != strings.Repeat(" ", test.result)+"↓" {
				t.Fatalf("arrow does not use the result column %d:\n%s", test.result, out.String())
			}
			foundResult := false
			for _, line := range lines {
				if lipgloss.Width(line) > test.width {
					t.Fatalf("annotation overflows %d columns: %q", test.width, line)
				}
				if prefix, _, found := strings.Cut(line, "TRUE"); found {
					foundResult = true
					if lipgloss.Width(prefix) != test.result {
						t.Fatalf("result and arrow are misaligned:\n%s", out.String())
					}
				}
			}
			if !foundResult || strings.Count(out.String(), "↓") != 1 || strings.ContainsAny(out.String(), "┬├╮╭│┆") {
				t.Fatalf("unexpected annotation decoration:\n%s", out.String())
			}
		})
	}
}

func TestStaircaseArrowsAndNextExpressionsStartUnderResultWords(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	first := `device.os == 'web'`
	second := `app.customSignal['name'].exactlyMatches(['Tria'])`
	third := `app.customSignal['environment'].exactlyMatches(['us-qa'])`
	source := first + " && " + second + " && " + third
	result, err := evaluate.Evaluate(source, &evaluate.Context{Facts: map[string]evaluate.Fact{
		"device.os":                       {Known: true, Value: "web", Source: "command line"},
		"app.customSignal['name']":        {Known: true, Value: "Tria", Source: "request"},
		"app.customSignal['environment']": {Known: true, Value: "by-dev", Source: "request"},
	}})
	if err != nil || result.State != evaluate.False {
		t.Fatalf("evaluation = %#v, %v", result, err)
	}
	out := renderer{width: 320}
	out.condition(report{Param: firebase.RemoteConfigParam{ValueType: "BOOLEAN"}, Winner: -1}, 0, conditionResult{
		Entry: core.ConditionEntry{Priority: 70, Name: "Tria QA signal"}, Value: firebase.RemoteConfigValue{Value: "true"}, Result: result,
	})
	for _, line := range []string{
		"  " + source,
		"       " + second + " && " + third,
		"            " + third,
		"                 eval: FALSE",
	} {
		if !strings.Contains(out.String(), "\n"+line+"\n") {
			t.Fatalf("missing aligned expression %q:\n%s", line, out.String())
		}
	}
	annotations := 0
	lines := strings.Split(out.String(), "\n")
	for i, line := range lines {
		index := strings.Index(line, "╰─ ")
		if index < 0 {
			continue
		}
		column := lipgloss.Width(line[:index+len("╰─ ")])
		if column != 7+annotations*5 || i+1 >= len(lines) || lines[i+1] != strings.Repeat(" ", column)+"↓" {
			t.Fatalf("arrow does not start directly under result:\n%s", out.String())
		}
		annotations++
	}
	if annotations != 3 {
		t.Fatalf("got %d annotations, want 3:\n%s", annotations, out.String())
	}
	t.Log("\n" + out.String())
}

func TestConditionBlockSpacingAndEvaluationLabels(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	r := exampleReport(t)
	got := render(r, 320)
	lines := strings.Split(got, "\n")
	headers, notes := 0, 0
	for i, line := range lines {
		for _, condition := range r.Conditions {
			if !strings.Contains(line, "] "+condition.Entry.Name+" → ") {
				continue
			}
			headers++
			if i+2 >= len(lines) || strings.TrimPrefix(lines[i+1], "┆") != "" || strings.TrimSpace(strings.TrimPrefix(lines[i+2], "┆")) != condition.Entry.Expression {
				t.Fatalf("condition header is not separated from its raw expression:\n%s", got)
			}
		}
		if strings.HasPrefix(strings.TrimSpace(strings.TrimPrefix(line, "┆")), "note:") {
			notes++
			if i < 2 || strings.TrimPrefix(lines[i-1], "┆") != "" || !strings.Contains(lines[i-2], "eval:") {
				t.Fatalf("condition note is not separated from its evaluation:\n%s", got)
			}
		}
	}
	if headers != len(r.Conditions) || notes != len(r.Conditions) {
		t.Fatalf("got %d condition headers and %d notes, want %d each", headers, notes, len(r.Conditions))
	}
	for _, want := range []string{"╰─ TRUE ·", "╰─ FALSE ·", "eval: TRUE", "eval: FALSE", "× Default value → false · not used", "◆ new_player = true", "android_users → false",
		"note: Highest-priority matching condition for this parameter\n        ╰── Supplies new_player = true",
		"note: Matches, but [23] android_english has higher priority\n        ╰── Conditional value false is not selected",
		"◆ new_player = true\n  ├─ Winner: [23] android_english\n  ├─ [31] android_users also matches, but lower priority\n  ╰─ Default false is not used",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q:\n%s", want, got)
		}
	}
	for _, unwanted := range []string{"╰─ true ·", "╰─ false ·", "eval: true", "eval: false", "→ TRUE", "→ FALSE", "◆ new_player = TRUE"} {
		if strings.Contains(got, unwanted) {
			t.Fatalf("parameter values and evaluation labels are mixed up (%q):\n%s", unwanted, got)
		}
	}
}

func TestInactiveExpressionFragmentsUseLogKeyStyle(t *testing.T) {
	prefix := "unknown && "
	active := `app.id == '1:123:android:a1'`
	suffix := ` && app.id == '1:123:ios:b2' && app.customSignal['flag'].exactlyMatches(['true', 'false'])`
	step := evaluate.Step{Expression: prefix + active + suffix, Start: len(prefix), End: len(prefix) + len(active)}
	t.Run("colored", func(t *testing.T) {
		t.Setenv("NO_COLOR", "")
		got := stepExpressionText(step)
		keyStyle := charmlog.DefaultStyles().Key
		want := keyStyle.Render(prefix) +
			"app.id == " + clistyles.RenderFirebaseAppPlatforms("'1:123:android:a1'", lipgloss.NewStyle().Foreground(clistyles.PaletteGold)) +
			keyStyle.Render(suffix)
		if got != want || ansi.Strip(got) != step.Expression {
			t.Fatalf("inactive fragments are not gray or active styling was lost:\ngot %q\nwant %q", got, want)
		}
	})
	t.Run("no color", func(t *testing.T) {
		t.Setenv("NO_COLOR", "1")
		if got := stepExpressionText(step); got != step.Expression {
			t.Fatalf("NO_COLOR altered raw expression: %q", got)
		}
	})
}

func TestEvaluationCasingDoesNotChangeRawBooleanData(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	source := `app.customSignal['first'].exactlyMatches(['true']) && app.customSignal['second'].exactlyMatches(['false'])`
	result, err := evaluate.Evaluate(source, &evaluate.Context{Facts: map[string]evaluate.Fact{
		"app.customSignal['first']":  {Known: true, Value: "true", Source: "request"},
		"app.customSignal['second']": {Known: true, Value: "false", Source: "request"},
	}})
	if err != nil || result.State != evaluate.True {
		t.Fatalf("evaluation = %#v, %v", result, err)
	}
	out := renderer{width: 320}
	out.condition(report{Param: firebase.RemoteConfigParam{ValueType: "BOOLEAN"}, Winner: 0}, 0, conditionResult{
		Entry: core.ConditionEntry{Priority: 1, Name: "boolean_data"}, Value: firebase.RemoteConfigValue{Value: "true"}, Result: result,
	})
	got := out.String()
	if !strings.Contains(got, source) || strings.Contains(got, "'TRUE'") || strings.Contains(got, "'FALSE'") {
		t.Fatalf("evaluation labels changed raw input data:\n%s", got)
	}
	if strings.Count(got, "╰─ TRUE ·") != 2 || !strings.Contains(got, "eval: TRUE") || !strings.Contains(got, "→ true · selected") {
		t.Fatalf("evaluation results are not distinct from parameter values:\n%s", got)
	}
}

func TestReportNotesHaveBlankLines(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	got := render(report{
		Project: "demo", Parameter: "flag", Version: "142", Source: "cache", Capture: "capture.har", Winner: -1,
		Assumptions: true,
		Param:       firebase.RemoteConfigParam{DefaultValue: &firebase.RemoteConfigValue{Value: "false"}},
		Exchange:    &evaluate.Exchange{Entry: 1, HasResponse: true, Entries: map[string]string{"flag": "true"}, Version: "141"},
	}, 320)
	lines := strings.Split(got, "\n")
	notes := 0
	for i, line := range lines {
		if strings.HasPrefix(line, "note:") {
			notes++
			if i == 0 || lines[i-1] != "" {
				t.Fatalf("report note is missing its preceding blank line:\n%s", got)
			}
		}
	}
	if notes != 2 {
		t.Fatalf("got %d report notes, want 2:\n%s", notes, got)
	}
}
