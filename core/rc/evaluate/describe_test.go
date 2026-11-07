package evaluate

import (
	"strings"
	"testing"
)

func TestOperationDescriptionsExplainBothOutcomes(t *testing.T) {
	const signal = "app.customSignal['value']"
	for _, test := range []struct {
		name, expression, field, matching, other, yes, no string
	}{
		{"equal", `app.id == 'one'`, "app.id", "one", "two", "equals 'one'", "does not equal 'one'"},
		{"not equal", `device.os != 'ios'`, "device.os", "android", "ios", "does not equal 'ios'", "equals 'ios'"},
		{"less", signal + " < 3", signal, "2", "3", "is less than '3'", "is greater than or equal to '3'"},
		{"less or equal", signal + " <= 3", signal, "3", "4", "is less than or equal to '3'", "is greater than '3'"},
		{"greater", signal + " > 3", signal, "4", "3", "is greater than '3'", "is less than or equal to '3'"},
		{"greater or equal", signal + " >= 3", signal, "3", "2", "is greater than or equal to '3'", "is less than '3'"},
		{"exact", signal + ".exactlyMatches(['Tria'])", signal, "Tria", "Other", "exactly matches 'Tria'", "does not exactly match 'Tria'"},
		{"exact list", signal + ".exactlyMatches(['Tria', 'Demo'])", signal, "Demo", "Other", "exactly matches one of ['Tria', 'Demo']", "does not exactly match any of ['Tria', 'Demo']"},
		{"contains", signal + ".contains(['Tria'])", signal, "Tria QA", "Other", "contains 'Tria'", "does not contain 'Tria'"},
		{"contains list", signal + ".contains(['Tria', 'Demo'])", signal, "Demo QA", "Other", "contains at least one of ['Tria', 'Demo']", "does not contain any of ['Tria', 'Demo']"},
		{"not contains", signal + ".notContains(['Tria'])", signal, "Other", "Tria QA", "does not contain 'Tria'", "contains 'Tria'"},
		{"not contains list", signal + ".notContains(['Tria', 'Demo'])", signal, "Other", "Demo QA", "does not contain any of ['Tria', 'Demo']", "contains at least one of ['Tria', 'Demo']"},
		{"regex", signal + ".matches(['^Tria$'])", signal, "Tria", "Tria QA", "matches the regular expression '^Tria$'", "does not match the regular expression '^Tria$'"},
		{"regex list", signal + ".matches(['^Tria$', '^Demo$'])", signal, "Demo", "Other", "matches at least one regular expression in ['^Tria$', '^Demo$']", "does not match any regular expression in ['^Tria$', '^Demo$']"},
		{"membership", "device.language in ['en-US', 'en-GB']", "device.language", "en-US", "fr", "is in ['en-US', 'en-GB']", "is not in ['en-US', 'en-GB']"},
		{"semantic version", "version(" + signal + ") >= '1.9'", signal, "1.10", "1.8", "is greater than or equal to '1.9'", "is less than '1.9'"},
		{"range boundaries", "percent between 10 and 20", "percent", "20", "10", "is within the range from '10' (exclusive) to '20' (inclusive)", "is outside the range from '10' (exclusive) to '20' (inclusive)"},
		{"before", "device.dateTime < '2026-11-04T12:00:00Z'", "device.dateTime", "2026-11-04T11:00:00Z", "2026-11-04T12:00:00Z", "is before '2026-11-04T12:00:00Z'", "is on or after '2026-11-04T12:00:00Z'"},
		{"on or before", "device.dateTime <= '2026-11-04T12:00:00Z'", "device.dateTime", "2026-11-04T12:00:00Z", "2026-11-04T13:00:00Z", "is on or before '2026-11-04T12:00:00Z'", "is after '2026-11-04T12:00:00Z'"},
		{"after", "device.dateTime > '2026-11-04T12:00:00Z'", "device.dateTime", "2026-11-04T13:00:00Z", "2026-11-04T12:00:00Z", "is after '2026-11-04T12:00:00Z'", "is on or before '2026-11-04T12:00:00Z'"},
		{"on or after", "device.dateTime >= '2026-11-04T12:00:00Z'", "device.dateTime", "2026-11-04T12:00:00Z", "2026-11-04T11:00:00Z", "is on or after '2026-11-04T12:00:00Z'", "is before '2026-11-04T12:00:00Z'"},
	} {
		t.Run(test.name, func(t *testing.T) {
			for _, outcome := range []struct {
				state State
				value string
				text  string
			}{{True, test.matching, test.yes}, {False, test.other, test.no}} {
				t.Run(string(outcome.state), func(t *testing.T) {
					result, err := Evaluate(test.expression, &Context{Facts: facts(map[string]string{test.field: outcome.value})})
					if err != nil || result.State != outcome.state || len(result.Steps) != 1 {
						t.Fatalf("evaluation = %#v, %v", result, err)
					}
					step := result.Steps[0]
					want := test.field + " = " + quote(outcome.value) + " (request) " + outcome.text
					if step.Note != want || step.Expression != test.expression {
						t.Fatalf("note = %q, want %q; expression = %q", step.Note, want, step.Expression)
					}
				})
			}
		})
	}
}

func TestAudienceDescriptionsRespectNegativeOperatorsAndUnknownMembership(t *testing.T) {
	for _, test := range []struct {
		method, beta, staff string
		state               State
		phrase              string
	}{
		{"inAtLeastOne", "true", "false", True, "App belongs to at least one"},
		{"inAtLeastOne", "false", "false", False, "App does not belong to any"},
		{"notInAtLeastOne", "false", "false", True, "App does not belong to any"},
		{"notInAtLeastOne", "true", "false", Unknown, "Negative audience semantics need Firebase verification"},
		{"inAll", "true", "true", True, "App belongs to all"},
		{"inAll", "true", "false", False, "App does not belong to all"},
		{"notInAll", "true", "false", Unknown, "Negative audience semantics need Firebase verification"},
		{"notInAll", "true", "true", False, "App belongs to all"},
		{"inAtLeastOne", "unknown", "false", Unknown, "Cannot determine whether the app belongs to at least one"},
		{"notInAtLeastOne", "unknown", "false", Unknown, "Negative audience semantics need Firebase verification"},
		{"inAll", "unknown", "true", Unknown, "Cannot determine whether the app belongs to all"},
		{"notInAll", "unknown", "true", Unknown, "Negative audience semantics need Firebase verification"},
	} {
		t.Run(test.method+"/"+test.beta+"/"+test.staff, func(t *testing.T) {
			expression := "app.audiences." + test.method + "(['Beta', 'Staff'])"
			result, err := Evaluate(expression, &Context{Facts: facts(map[string]string{
				"app.audiences['Beta']": test.beta, "app.audiences['Staff']": test.staff,
			})})
			if err != nil || result.State != test.state || !strings.HasPrefix(result.Steps[0].Note, test.phrase) || !strings.Contains(result.Steps[0].Note, "'Beta'") || !strings.Contains(result.Steps[0].Note, "'Staff'") {
				t.Fatalf("audience description = %#v, %v", result, err)
			}
		})
	}
}

func TestClientVersionDescriptionsExplainNameAndVersionOutcomes(t *testing.T) {
	for _, test := range []struct {
		selector, name, version, phrase string
		state                           State
	}{
		{`browserName('Chrome').anyVersion`, "Chrome", "120", "exactly matches 'Chrome' and any version is accepted", True},
		{`browserName('Chrome').anyVersion`, "Firefox", "120", "is not in ['Chrome']", False},
		{`browserName('Chrome').version.>=('120')`, "Chrome", "121", "is greater than or equal to '120'", True},
		{`browserName('Chrome').version.>=('120')`, "Chrome", "119", "is less than '120'", False},
		{`browserName('Chrome').version.exactlyMatches(['120'])`, "Chrome", "119", "does not exactly match '120'", False},
		{`browserName('Chrome').version.notContains(['beta'])`, "Chrome", "120-beta", "contains 'beta'", False},
		{`browserName('Chrome').version.>=('120')`, "Chrome", "", "version is unavailable", Unknown},
	} {
		t.Run(test.selector+"/"+test.name+"/"+test.version, func(t *testing.T) {
			inputs := facts(map[string]string{"app.browserAndVersion.name": test.name})
			if test.version != "" {
				inputs["app.browserAndVersion.version"] = Fact{Known: true, Value: test.version, Source: "request"}
			}
			expression := "app.browserAndVersion.inOne([" + test.selector + "])"
			result, err := Evaluate(expression, &Context{Facts: inputs})
			if err != nil || result.State != test.state || !strings.Contains(result.Steps[0].Note, test.phrase) {
				t.Fatalf("client version description = %#v, %v", result, err)
			}
		})
	}
}
