package evaluate

import (
	"fmt"
	"strings"
	"testing"
)

func facts(values map[string]string) map[string]Fact {
	out := make(map[string]Fact, len(values))
	for key, value := range values {
		out[key] = Fact{Known: true, Value: value, Source: "request"}
	}
	return out
}

func TestShortCircuitDoesNotResolveSkippedInputs(t *testing.T) {
	for _, expression := range []string{
		`app.id == 'other' && app.audiences.inAtLeastOne(['Beta testers'])`,
		`app.id == 'actual' && false && app.userProperty['tier'].exactlyMatches(['paid'])`,
		`false && app.version >= '4.0.0' && app.audiences.inAtLeastOne(['Beta testers'])`,
	} {
		t.Run(expression, func(t *testing.T) {
			ctx := &Context{Facts: facts(map[string]string{"app.id": "actual"}), Resolve: func(need Need) (Fact, error) {
				t.Fatalf("asked for skipped input %s", need.Key)
				return Fact{}, nil
			}}
			result, err := Evaluate(expression, ctx)
			if err != nil || result.State == Unknown || result.Steps[0].Expression != expression {
				t.Fatalf("evaluation = %#v, %v", result, err)
			}
			for _, step := range result.Steps {
				if strings.Contains(step.Note, "Left side") || strings.Contains(step.Note, "skip the rest") {
					t.Fatalf("unnecessary control-flow step: %#v", step)
				}
				if step.Start < 0 || step.End > len(step.Expression) || step.Start > step.End {
					t.Fatalf("invalid annotation span: %#v", step)
				}
			}
		})
	}
}

func TestRulesAndConjunctions(t *testing.T) {
	ctx := &Context{Facts: facts(map[string]string{
		"app.id": "1:123:android:abc", "app.version": "3.10.0", "app.build": "42",
		"device.country": "DE", "device.language": "en-US", "device.os": "android",
		"app.userProperty['tier']": "premium", "app.customSignal['score']": "3.5",
		"app.customSignal['release']": "1.12.0", "app.firstOpenTimestamp": "2026-10-01T10:00:00Z",
		"device.dateTime": "2026-11-04T12:00:00Z", "percent['rollout']": "4.5",
	})}
	for _, test := range []struct {
		expression string
		want       State
	}{
		{`app.id == '1:123:android:abc' && app.version >= '3.9.0'`, True},
		{`app.version == '3.10'`, True},
		{`app.version.exactlyMatches(['3.9.0', '3.10.0'])`, True},
		{`app.version.matches(['^3\\.10\\.'])`, True},
		{`app.version.notContains(['-beta', '-rc'])`, True},
		{`app.build >= 40`, True},
		{`device.country in ['de', 'us'] && device.language in ['en-US']`, True},
		{`device.os != 'ios'`, True},
		{`app.userProperty['tier'].exactlyMatches(['premium'])`, True},
		{`app.customSignal['score'] > 3`, True},
		{`version(app.customSignal['release']) >= '1.9.0'`, True},
		{`device.dateTime > dateTime('2026-11-04T13:00:00', 'Europe/Moscow')`, True},
		{`app.firstOpenTimestamp >= ('2026-10-01T00:00:00')`, True},
		{`percent('rollout') <= 10`, True},
		{`percent('rollout') between 4 and 4.5`, True},
		{`percent('rollout') between 4.5 and 10`, False},
		{`true && true && true`, True},
		{`true && false && app.version >= '4.0.0'`, False},
		{`app.version >= '4.0.0'`, False},
		{`app.version >= '4.0-beta'`, Unknown},
		{`app.version.matches(['['])`, Unknown},
		{`app.customSignal['score'] > 'oops'`, Unknown},
		{`future.unknownMethod(['x'])`, Unknown},
		{`app.id == 'x' garbage`, Unknown},
		{`true &&`, Unknown},
	} {
		t.Run(test.expression, func(t *testing.T) {
			result, err := Evaluate(test.expression, ctx)
			if err != nil || result.State != test.want {
				t.Fatalf("got %#v, %v; want %s", result, err, test.want)
			}
		})
	}
}

func TestUnknownAndUnsetRemainDistinct(t *testing.T) {
	for _, test := range []struct {
		expression string
		fact       Fact
		want       State
	}{
		{`app.userProperty['x'].exactlyMatches([''])`, Fact{Known: true, Value: ""}, True},
		{`app.userProperty['x'].exactlyMatches([''])`, Fact{Known: true, Absent: true}, False},
		{`app.userProperty['x'].exactlyMatches([''])`, Fact{}, Unknown},
		{`app.userProperty['x'].exactlyMatches(['']) && false`, Fact{}, False},
		{`app.userProperty['x'].exactlyMatches(['']) && true`, Fact{}, Unknown},
	} {
		ctx := &Context{Facts: map[string]Fact{"app.userProperty['x']": test.fact}}
		result, err := Evaluate(test.expression, ctx)
		if err != nil || result.State != test.want {
			t.Fatalf("%s with %#v: %s, %v; want %s", test.expression, test.fact, result.State, err, test.want)
		}
	}
}

func TestRepeatedInputsAreResolvedOnceIncludingUnknown(t *testing.T) {
	for _, answer := range []Fact{{Known: true, Value: "paid", Source: "your answer"}, {}} {
		calls := 0
		ctx := &Context{Resolve: func(need Need) (Fact, error) {
			calls++
			return answer, nil
		}}
		for _, expression := range []string{`app.userProperty['tier'].exactlyMatches(['paid'])`, `app.userProperty["tier"].contains(['a'])`} {
			if _, err := Evaluate(expression, ctx); err != nil {
				t.Fatal(err)
			}
		}
		if calls != 1 {
			t.Fatalf("resolved the same input %d times", calls)
		}
	}
}

func TestAudienceMembershipUsesIndividualFactsAndShortCircuits(t *testing.T) {
	ctx := &Context{Resolve: func(need Need) (Fact, error) {
		if need.Key != "app.audiences['Beta testers']" {
			t.Fatalf("unnecessary membership question: %s", need.Key)
		}
		return Fact{Known: true, Value: "true", Source: "your answer"}, nil
	}}
	result, err := Evaluate(`app.audiences.inAtLeastOne(['Beta testers', 'Staff'])`, ctx)
	if err != nil || result.State != True || !strings.Contains(result.Steps[0].Note, "your answer") {
		t.Fatalf("audience result = %#v, %v", result, err)
	}
}

func TestResolverErrorsPropagate(t *testing.T) {
	want := fmt.Errorf("cancelled prompt")
	_, err := Evaluate(`app.id == 'x'`, &Context{Resolve: func(Need) (Fact, error) { return Fact{}, want }})
	if err != want {
		t.Fatalf("got %v, want resolver error", err)
	}
}

func TestUnsupportedBooleanSyntaxDoesNotResolveInputs(t *testing.T) {
	for _, expression := range []string{
		`app.id == 'a' || app.id == 'b'`,
		`(app.id == 'a')`,
		`(app.id == 'a' && app.version >= '4.0.0')`,
		`!true`,
		`!(app.id == 'a')`,
		`false && (app.id == 'a')`,
		`false && !true`,
		`false && app.id == 'a' || true`,
	} {
		t.Run(expression, func(t *testing.T) {
			result, err := Evaluate(expression, &Context{Resolve: func(need Need) (Fact, error) {
				t.Fatalf("resolved an input in unsupported syntax: %s", need.Key)
				return Fact{}, nil
			}})
			if err != nil || result.State != Unknown || len(result.Steps) != 1 || result.Steps[0].Expression != expression {
				t.Fatalf("unsupported expression = %#v, %v", result, err)
			}
		})
	}
}

func TestAudienceMethodsRetainAnyAllAndNegativeMembership(t *testing.T) {
	ctx := &Context{Facts: facts(map[string]string{
		"app.audiences['Beta testers']": "true", "app.audiences['Staff']": "false",
	})}
	for _, test := range []struct {
		method string
		want   State
	}{
		{"inAtLeastOne", True}, {"notInAtLeastOne", Unknown},
		{"inAll", False}, {"notInAll", Unknown},
	} {
		expression := "app.audiences." + test.method + "(['Beta testers', 'Staff'])"
		result, err := Evaluate(expression, ctx)
		if err != nil || result.State != test.want {
			t.Fatalf("%s = %#v, %v; want %s", expression, result, err, test.want)
		}
	}
}

func TestWebClientVersionTargets(t *testing.T) {
	ctx := &Context{Facts: facts(map[string]string{
		"app.browserAndVersion.name": "Chrome", "app.browserAndVersion.version": "120.0.0",
		"app.operatingSystemAndVersion.name": "Macintosh", "app.operatingSystemAndVersion.version": "10.15",
	})}
	for _, test := range []struct {
		expression string
		want       State
	}{
		{`app.browserAndVersion.inOne([browserName('Chrome').anyVersion])`, True},
		{`app.browserAndVersion.inOne([browserName('Safari').anyVersion, browserName('Chrome').version.>=('120')])`, True},
		{`app.operatingSystemAndVersion.inOne([operatingSystemName('Macintosh').version.==('10.15')])`, True},
		{`app.browserAndVersion.inOne([browserName('Chrome').version.exactlyMatches(['119.0'])])`, False},
	} {
		result, err := Evaluate(test.expression, ctx)
		if err != nil || result.State != test.want {
			t.Fatalf("%s: %#v, %v; want %s", test.expression, result, err, test.want)
		}
	}
	ctx = &Context{Facts: facts(map[string]string{"app.browserAndVersion.name": "Firefox"}), Resolve: func(need Need) (Fact, error) {
		t.Fatalf("irrelevant browser version prompt: %s", need.Key)
		return Fact{}, nil
	}}
	result, err := Evaluate(`app.browserAndVersion.inOne([browserName('Chrome').version.>=('120')])`, ctx)
	if err != nil || result.State != False {
		t.Fatalf("nonmatching browser: %#v, %v", result, err)
	}
}

func FuzzEvaluate(f *testing.F) {
	for _, source := range []string{"", "true", `app.id == 'a' && app.version.matches(['x'])`, `!(false || true)`, `'a && b'`} {
		f.Add(source)
	}
	f.Fuzz(func(t *testing.T, source string) {
		if len(source) > 8192 {
			t.Skip()
		}
		result, err := Evaluate(source, &Context{})
		if err != nil || !strings.Contains("true false unknown", string(result.State)) || result.State == "" {
			t.Fatalf("invalid evaluation: %#v, %v", result, err)
		}
		for _, step := range result.Steps {
			if step.Start < 0 || step.End < step.Start || step.End > len(step.Expression) {
				t.Fatalf("invalid annotation span: %#v", step)
			}
		}
		if len(result.Steps) == 0 || result.Steps[0].Expression != source {
			t.Fatalf("trace lost the original expression: %#v", result)
		}
	})
}
