package explain

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/core/rc/evaluate"
)

func conditionTemplate(expression string) firebase.RemoteConfig {
	template := simpleTemplate()
	template.Conditions = []firebase.RemoteConfigCondition{{Name: "target", Expression: expression}}
	param := template.Parameters["flag"]
	param.ConditionalValues = map[string]firebase.RemoteConfigValue{"target": {Value: "true"}}
	template.Parameters["flag"] = param
	return template
}

func TestCapturedContextFixesWorkThroughTheCommandWithoutPrompts(t *testing.T) {
	for _, test := range []struct{ name, expression, body, headers, want string }{
		{"Apple console expression", `app.version.>=(['3.9']) && app.build.==(['42'])`, `{app_id:'1:123456789:ios:abc',app_version:'4.0.0',app_build:'42'}`, "", "◆ flag = true"},
		{"headers", `app.browserAndVersion.inOne([browserName('Chrome').anyVersion])`, `{"app_id":"1:123456789:web:abc"}`, ` -A 'Mozilla/5.0 Chrome/120.0.0.0'`, "◆ flag = true"},
		{"omitted signal", `app.customSignal['missing'].notContains(['x'])`, `{"app_id":"1:123456789:web:abc","custom_signals":{}}`, "", "◆ flag = false"},
		{"omitted native property", `app.userProperty['missing'].exactlyMatches(['x'])`, `{"appId":"1:123456789:android:abc","analyticsUserProperties":{}}`, "", "◆ flag = false"},
		{"known percentage", `percent <= 100`, `{}`, "", "◆ flag = true"},
		{"unverified percentage", `percent <= 10`, `{}`, "", "? flag = unknown"},
		{"country observation", `device.country in ['US']`, `{"app_id":"1:123456789:web:abc","country_code":"US"}`, "", "? flag = unknown"},
	} {
		t.Run(test.name, func(t *testing.T) {
			svc := serviceWithTemplate(t, conditionTemplate(test.expression))
			input := "curl '" + requestURL + "'" + test.headers + " --data-raw '" + strings.ReplaceAll(test.body, "'", `'\''`) + "'"
			prompts := &fakePrompter{}
			out, err := execute(t, svc, input, prompts, "flag", "--cached", "--no-input")
			if err != nil || !strings.Contains(out, test.want) || len(prompts.asked) != 0 {
				t.Fatalf("capture integration: %v, prompts %v\n%s", err, prompts.asked, out)
			}
		})
	}
}

func TestExplainDeviceCategoryUsesAnswersOrExplicitOverrides(t *testing.T) {
	const expression = `app.id == '1:123456789:web:abc' && device.category.contains(['Mobile'])`
	for _, tc := range []struct {
		name   string
		args   []string
		answer evaluate.Fact
		want   string
		asked  int
	}{
		{"supplied category matches", []string{"--set", "device.category=Mobile"}, evaluate.Fact{}, "◆ flag = true", 0},
		{"supplied category differs", []string{"--set", "device.category=Desktop"}, evaluate.Fact{}, "◆ flag = false", 0},
		{"interactive category", nil, evaluate.Fact{Known: true, Value: "Mobile", Source: "your answer"}, "◆ flag = true", 1},
		{"category unavailable", []string{"--no-input"}, evaluate.Fact{}, "? flag = unknown", 0},
		{"category explicitly absent", []string{"--set", "device.category=/unset"}, evaluate.Fact{}, "◆ flag = false", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc := serviceWithTemplate(t, conditionTemplate(expression))
			prompts := &fakePrompter{answers: map[string]evaluate.Fact{"device.category": tc.answer}}
			args := append([]string{"flag", "--cached"}, tc.args...)
			out, err := execute(t, svc, `curl '`+requestURL+`' --data '{"app_id":"1:123456789:web:abc"}'`, prompts, args...)
			if err != nil || !strings.Contains(out, tc.want) || len(prompts.asked) != tc.asked || strings.Contains(out, "evaluation unavailable") {
				t.Fatalf("category command result: %v, prompts %v\n%s", err, prompts.asked, out)
			}
			if tc.asked > 0 && prompts.asked[0] != "device.category" {
				t.Fatalf("asked for unrelated input: %v", prompts.asked)
			}
		})
	}
}

func TestExplainBareDateTimeUsesCapturedTimeAndEitherInputAlias(t *testing.T) {
	const expression = `dateTime >= dateTime('2026-11-01T00:00:00', 'Europe/Moscow') && dateTime < dateTime('2026-11-07T00:00:00', 'Europe/Moscow')`
	har := `{"log":{"entries":[{"startedDateTime":"2026-11-04T12:00:00Z","time":1000,"request":{"method":"POST","url":"` + requestURL + `","postData":{"text":"{}"}},"response":{"status":200,"headers":[{"name":"Date","value":"Wed, 04 Nov 2026 12:00:00 GMT"}],"content":{"text":"{\"entries\":{\"flag\":\"true\"}}"}}}]}}`
	for _, tc := range []struct {
		name, input, source string
		args                []string
	}{
		{"captured time", har, "response Date header", nil},
		{"bare input alias", `curl '` + requestURL + `' --data '{}'`, "command line", []string{"--set", "dateTime=2026-11-04T12:00:00Z"}},
		{"existing input alias", `curl '` + requestURL + `' --data '{}'`, "command line", []string{"--set", "device.dateTime=2026-11-04T12:00:00Z"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc := serviceWithTemplate(t, conditionTemplate(expression))
			prompts := &fakePrompter{}
			args := append([]string{"flag", "--cached", "--no-input"}, tc.args...)
			out, err := execute(t, svc, tc.input, prompts, args...)
			if err != nil || !strings.Contains(out, "◆ flag = true") || !strings.Contains(out, tc.source) || !strings.Contains(out, expression) || len(prompts.asked) != 0 {
				t.Fatalf("dateTime command result: %v, prompts %v\n%s", err, prompts.asked, out)
			}
		})
	}
}

func TestManagedSelectionsAndCapturedAssignmentsAreNotOrdinaryPredictions(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	r := report{Project: "demo", Parameter: "flag", Winner: -1, Param: firebase.RemoteConfigParam{ValueType: "BOOLEAN", DefaultValue: &firebase.RemoteConfigValue{PersonalizationValue: json.RawMessage(`{"personalizationId":"p1"}`)}}}
	got := render(r, 200)
	if !strings.Contains(got, "? flag = unknown") || !strings.Contains(got, "requires Firebase-managed evaluation") || strings.Contains(got, "◆ flag") {
		t.Fatalf("invented a managed value:\n%s", got)
	}
	if !strings.Contains(got, "? flag = unknown\n  ├─ No conditions exist, so use the default value\n  ╰─ The selected value requires Firebase-managed evaluation") {
		t.Fatalf("managed default does not explain the absence of conditions:\n%s", got)
	}
	r.Param.DefaultValue = &firebase.RemoteConfigValue{Value: "false"}
	r.Exchange = &evaluate.Exchange{HasResponse: true, Status: 200, Entries: map[string]string{"flag": "true"}, Managed: map[string]string{"flag": "rollout r1, variant v1"}, Experiments: []string{"e1"}}
	got = render(r, 200)
	for _, want := range []string{"◆ flag = false (template evaluation)", "Captured response: true", "assigned this parameter through rollout r1, variant v1", "A/B experiment assignments"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q:\n%s", want, got)
		}
	}
}

func TestNormalizedResponseStatesAndSpacing(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	for _, assumptions := range []bool{false, true} {
		r := report{Project: "demo", Parameter: "flag", Winner: -1, Assumptions: assumptions, Exchange: &evaluate.Exchange{Status: 304, ResponseState: "NO_CHANGE"}}
		got := render(r, 200)
		if !strings.Contains(got, "\n\nCaptured response: HTTP 304 with no new parameter values") || strings.Contains(got, "\n\n\nCaptured response:") {
			t.Fatalf("response state or spacing was lost:\n%s", got)
		}
	}
}
