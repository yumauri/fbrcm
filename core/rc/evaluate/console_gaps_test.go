package evaluate

import (
	"strings"
	"testing"
)

func TestDeviceCategoryUsesConfirmedContainsRuleAndExplicitFacts(t *testing.T) {
	for _, tc := range []struct {
		name string
		fact Fact
		want State
	}{
		{"matching category", Fact{Known: true, Value: "Mobile", Source: "your answer"}, True},
		{"different category", Fact{Known: true, Value: "Desktop", Source: "your answer"}, False},
		{"absent category", Fact{Known: true, Absent: true, Source: "command line"}, False},
		{"unknown category", Fact{}, Unknown},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Evaluate(`device.category.contains(['Mobile'])`, &Context{Facts: map[string]Fact{"device.category": tc.fact}})
			if err != nil || result.Unsupported || result.State != tc.want {
				t.Fatalf("category evaluation: %#v, %v", result, err)
			}
			if tc.want == True && !strings.Contains(result.Steps[0].Note, "contains 'Mobile'") {
				t.Fatalf("category did not reuse human comparison wording: %s", result.Steps[0].Note)
			}
		})
	}
	for _, expression := range []string{`device.category == 'Mobile'`, `device.category.notContains(['Mobile'])`, `device.category.>(['Mobile'])`} {
		result, err := Evaluate(expression, nil)
		if err != nil || !result.Unsupported {
			t.Fatalf("accepted unverified category operator: %s, %v", expression, err)
		}
	}
}

func TestDeviceCategoryIsPromptedOnceWithoutGuessingFromHeaders(t *testing.T) {
	exchanges, err := ReadCapture(strings.NewReader(`curl '` + fetchURL + `' -A 'Mozilla/5.0 (iPhone; CPU iPhone OS 18_0 like Mac OS X) Mobile/15E148 Safari/604.1' -H 'Sec-CH-UA-Mobile: ?1' --data '{"app_id":"1:123456789:web:abc"}'`))
	if err != nil {
		t.Fatal(err)
	}
	for _, answer := range []Fact{{Known: true, Value: "Mobile", Source: "your answer"}, {}} {
		asked := 0
		ctx := &Context{Facts: exchanges[0].Facts, Capture: &exchanges[0], Resolve: func(need Need) (Fact, error) {
			asked++
			if need.Key != "device.category" || need.Kind != "text" {
				t.Fatalf("unexpected category input: %#v", need)
			}
			return answer, nil
		}}
		delete(ctx.Facts, "device.category")
		result, err := Evaluate(`device.category.contains(['Mobile']) && device.category.contains(['Mobile'])`, ctx)
		if err != nil || result.Unsupported || asked != 1 || (result.State == True) != answer.Known {
			t.Fatalf("category was inferred, not memoized, or unsupported: %#v, %v, asked %d", result, err, asked)
		}
	}
}

func TestBareDateTimeReusesTimestampAndTimezoneComparisons(t *testing.T) {
	expression := `dateTime >= dateTime('2026-11-01T00:00:00', 'Europe/Moscow') && dateTime < dateTime('2026-11-07T00:00:00', 'Europe/Moscow')`
	for _, tc := range []struct {
		stamp string
		want  State
	}{
		{"2026-10-31T20:59:59Z", False},
		{"2026-10-31T21:00:00Z", True},
		{"2026-11-04T12:00:00Z", True},
		{"2026-11-06T20:59:59Z", True},
		{"2026-11-06T21:00:00Z", False},
	} {
		t.Run(tc.stamp, func(t *testing.T) {
			result, err := Evaluate(expression, &Context{Facts: facts(map[string]string{"device.dateTime": tc.stamp})})
			if err != nil || result.Unsupported || result.State != tc.want || result.Steps[0].Expression != expression {
				t.Fatalf("bare dateTime or timezone boundary: %#v, %v", result, err)
			}
		})
	}
	result, err := Evaluate(`dateTime < dateTime('2026-11-07T00:00:00')`, &Context{Facts: facts(map[string]string{"device.dateTime": "2026-11-06T22:00:00Z"})})
	if err != nil || result.State != True {
		t.Fatalf("timezone-free dateTime is not UTC: %#v, %v", result, err)
	}
}

func TestDateTimeInputAliasesShareValidationAndOnePrompt(t *testing.T) {
	for _, field := range []string{"dateTime", "device.dateTime"} {
		need, ok := InputNeed(field)
		if !ok || need.Key != "device.dateTime" || need.Kind != "timestamp" {
			t.Fatalf("timestamp input alias not canonicalized: %#v", need)
		}
		if _, err := Answer(need, "not a timestamp", "command line"); err == nil {
			t.Fatalf("invalid timestamp accepted for %s", field)
		}
	}
	asked := 0
	ctx := &Context{Resolve: func(need Need) (Fact, error) {
		asked++
		if need.Key != "device.dateTime" {
			t.Fatalf("unexpected timestamp input: %#v", need)
		}
		return Answer(need, "2026-11-04T12:00:00Z", "your answer")
	}}
	result, err := Evaluate(`dateTime >= dateTime('2026-11-01T00:00:00') && device.dateTime < dateTime('2026-11-07T00:00:00')`, ctx)
	if err != nil || result.Unsupported || result.State != True || asked != 1 {
		t.Fatalf("timestamp aliases did not share their input: %#v, %v, asked %d", result, err, asked)
	}
}
