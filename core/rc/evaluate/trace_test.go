package evaluate

import (
	"reflect"
	"strings"
	"testing"
)

func TestTraceVisitsOnlyRulesAndKeepsFullRemainingExpression(t *testing.T) {
	first := `app.customSignal['app_name'].notContains(['sandbox'])`
	second := `app.customSignal['app_name'].notContains(['demo'])`
	third := `app.customSignal['environment'].exactlyMatches(['by-dev'])`
	source := first + " && " + second + " && " + third
	result, err := Evaluate(source, &Context{Facts: facts(map[string]string{
		"app.customSignal['app_name']": "setplex", "app.customSignal['environment']": "by-dev",
	})})
	if err != nil || result.State != True {
		t.Fatalf("evaluation = %#v, %v", result, err)
	}
	wantExpressions := []string{source, second + " && " + third, third}
	wantHighlights := []string{first, second, third}
	var expressions, highlights []string
	for _, step := range result.Steps {
		expressions = append(expressions, step.Expression)
		highlights = append(highlights, step.Expression[step.Start:step.End])
	}
	if !reflect.DeepEqual(expressions, wantExpressions) || !reflect.DeepEqual(highlights, wantHighlights) {
		t.Fatalf("rule-only trace:\nexpressions %#v\nhighlights %#v", expressions, highlights)
	}
}

func TestTracePreservesIdenticalClausesAndQuotedOperatorCharacters(t *testing.T) {
	clause := `app.customSignal['text'].exactlyMatches(['(x || y) && !z'])`
	source := "  " + clause + " && " + clause + "  "
	result, err := Evaluate(source, &Context{Facts: facts(map[string]string{
		"app.customSignal['text']": "(x || y) && !z",
	})})
	if err != nil || result.State != True || len(result.Steps) != 2 {
		t.Fatalf("evaluation = %#v, %v", result, err)
	}
	if result.Steps[0].Expression != source || result.Steps[1].Expression != clause+"  " {
		t.Fatalf("lost source text or replaced both identical clauses: %#v", result.Steps)
	}
	for _, step := range result.Steps {
		if step.Expression[step.Start:step.End] != clause {
			t.Fatalf("incorrect rule annotation: %#v", step)
		}
	}
}

func TestTraceUnknownPrefixesDoNotPreventDecisiveFalse(t *testing.T) {
	missing := `app.userProperty['tier'].exactlyMatches(['premium'])`
	app := `app.id == 'a'`
	version := `app.version >= '4.0.0'`
	skipped := `app.customSignal['unused'].exactlyMatches(['x'])`
	source := missing + " && " + app + " && " + version + " && " + skipped
	result, err := Evaluate(source, &Context{
		Facts: map[string]Fact{
			"app.userProperty['tier']": {},
			"app.id":                   {Known: true, Value: "a", Source: "request"},
			"app.version":              {Known: true, Value: "3.8.0", Source: "request"},
		},
		Resolve: func(need Need) (Fact, error) {
			t.Fatalf("asked for a skipped input: %s", need.Key)
			return Fact{}, nil
		},
	})
	if err != nil || result.State != False {
		t.Fatalf("evaluation = %#v, %v", result, err)
	}
	want := []string{source, "unknown && " + app + " && " + version + " && " + skipped, "unknown && " + version + " && " + skipped}
	var got []string
	for _, step := range result.Steps {
		got = append(got, step.Expression)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unknown-prefix trace:\ngot %#v\nwant %#v", got, want)
	}
	if last := result.Steps[2]; last.Result != False || last.Expression[last.Start:last.End] != version {
		t.Fatalf("trace did not end on the decisive rule: %#v", last)
	}
}

func TestTraceKeepsUnknownResultsWhenAllOtherRulesMatch(t *testing.T) {
	missing := `app.userProperty['tier'].exactlyMatches(['premium'])`
	source := "true && " + missing + " && true && " + missing
	result, err := Evaluate(source, &Context{})
	if err != nil || result.State != Unknown || len(result.Steps) != 4 {
		t.Fatalf("evaluation = %#v, %v", result, err)
	}
	want := []string{source, missing + " && true && " + missing, "unknown && true && " + missing, "unknown && " + missing}
	for i, step := range result.Steps {
		if step.Expression != want[i] {
			t.Fatalf("step %d expression = %q, want %q", i, step.Expression, want[i])
		}
	}
}

func TestTraceStopsOnFirstFalseWithoutControlFlowSteps(t *testing.T) {
	missing := `app.userProperty['tier'].exactlyMatches(['premium'])`
	for _, test := range []struct {
		source string
		steps  int
	}{
		{"false && " + missing + " && " + missing, 1},
		{"true && false && " + missing, 2},
		{"true && true && false && " + missing, 3},
	} {
		t.Run(test.source, func(t *testing.T) {
			result, err := Evaluate(test.source, &Context{Resolve: func(need Need) (Fact, error) {
				t.Fatalf("asked for a short-circuited input: %s", need.Key)
				return Fact{}, nil
			}})
			if err != nil || result.State != False || len(result.Steps) != test.steps {
				t.Fatalf("evaluation = %#v, %v; want false in %d rule steps", result, err, test.steps)
			}
			for _, step := range result.Steps {
				if strings.Contains(step.Note, "Left side") || strings.Contains(step.Note, "skip the rest") {
					t.Fatalf("unexpected control-flow step: %#v", step)
				}
			}
			if result.Steps[len(result.Steps)-1].Result != False {
				t.Fatal("continued the trace after the condition was decided")
			}
		})
	}
}
