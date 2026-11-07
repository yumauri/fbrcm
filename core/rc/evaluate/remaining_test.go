package evaluate

import (
	"reflect"
	"testing"
)

func TestRemainingRulesPreservesOnlyUnknownRequirements(t *testing.T) {
	for _, tc := range []struct {
		name, expression string
		want             []string
		unsupported      bool
	}{
		{"known prefix", `app.id == 'actual' && device.os == 'web'`, []string{`device.os == 'web'`}, false},
		{"known rule between unknowns", `device.os == 'web' && app.id == 'actual' && app.version >= '4.0'`, []string{`device.os == 'web'`, `app.version >= '4.0'`}, false},
		{"quoted separator and whitespace", `  app.customSignal['name'].contains(['a && b'])  &&   app.id == 'actual'  `, []string{`app.customSignal['name'].contains(['a && b'])`}, false},
		{"false after unknown", `device.os == 'web' && app.id == 'other'`, nil, false},
		{"short circuit", `false && device.os == 'web'`, nil, false},
		{"known match", `app.id == 'actual'`, nil, false},
		{"unsupported expression", `future.deviceCategory == 'tablet'`, []string{`future.deviceCategory == 'tablet'`}, true},
		{"unsupported boolean syntax", `device.os == 'web' || device.os == 'ios'`, []string{`device.os == 'web' || device.os == 'ios'`}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := &Context{Facts: facts(map[string]string{"app.id": "actual"})}
			result, err := Evaluate(tc.expression, ctx)
			if err != nil {
				t.Fatal(err)
			}
			ctx.Resolve = func(need Need) (Fact, error) {
				t.Fatalf("remaining requirements requested %s again", need.Key)
				return Fact{}, nil
			}
			if got := result.RemainingRules(); !reflect.DeepEqual(got, tc.want) || result.Unsupported != tc.unsupported {
				t.Fatalf("remaining rules = %q, unsupported = %v; want %q, %v", got, result.Unsupported, tc.want, tc.unsupported)
			}
		})
	}
}

func TestRecordedRequirementsProveImplicationConservatively(t *testing.T) {
	for _, tc := range []struct {
		name, candidate, earlier string
		want                     bool
	}{
		{"same requirement", `device.os == 'web'`, `device.os == 'web'`, true},
		{"whitespace and quote style", `device.os=="web"`, `device.os == 'web'`, true},
		{"equivalent escapes", `device.os == 'w\u0065b'`, `device.os == 'web'`, true},
		{"known prefix is irrelevant", `device.os == 'web'`, `app.id == 'actual' && device.os == 'web'`, true},
		{"timestamp aliases share a requirement", `dateTime >= dateTime('2026-11-01T00:00:00')`, `device.dateTime >= dateTime('2026-11-01T00:00:00')`, true},
		{"stricter conjunction", `device.os == 'web' && app.version >= '4.0'`, `device.os == 'web'`, true},
		{"reordered conjunction", `app.version >= '4.0' && device.os == 'web'`, `device.os == 'web' && app.version >= '4.0'`, true},
		{"weaker conjunction can still win", `device.os == 'web'`, `device.os == 'web' && app.version >= '4.0'`, false},
		{"different values", `device.os == 'ios'`, `device.os == 'web'`, false},
		{"different fields", `device.country in ['US']`, `device.language in ['en']`, false},
		{"spaces within values matter", `app.customSignal['name'].exactlyMatches(['a b'])`, `app.customSignal['name'].exactlyMatches(['ab'])`, false},
		{"operators within values are literal", `app.customSignal['name'].exactlyMatches(['a && b'])`, `app.customSignal['name'].exactlyMatches(['a', 'b'])`, false},
		{"unrelated operators are not assumed equivalent", `device.os != 'ios'`, `device.os == 'web'`, false},
		{"matching known condition", `device.os == 'web'`, `true`, true},
		{"false candidate cannot win", `device.os == 'web' && false`, `device.os == 'ios'`, true},
		{"false earlier condition does not shadow", `device.os == 'web'`, `false`, false},
		{"identical unsupported expression", `future.deviceCategory == 'tablet'`, `future.deviceCategory == "tablet"`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := &Context{Facts: facts(map[string]string{"app.id": "actual"})}
			candidate, err := Evaluate(tc.candidate, ctx)
			if err != nil {
				t.Fatal(err)
			}
			earlier, err := Evaluate(tc.earlier, ctx)
			if err != nil {
				t.Fatal(err)
			}
			if got := candidate.Implies(earlier); got != tc.want {
				t.Fatalf("%s implies %s = %v, want %v", tc.candidate, tc.earlier, got, tc.want)
			}
		})
	}
}
