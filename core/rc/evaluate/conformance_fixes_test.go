package evaluate

import (
	"strings"
	"testing"
)

func TestUnsupportedFieldOperatorPairsAndGrammarNeverPrompt(t *testing.T) {
	for _, expression := range []string{
		`app.id != 'a'`, `app.id.contains(['a'])`, `device.country == 'US'`, `device.language.contains(['en'])`,
		`app.firebaseInstallationId == 'a'`, `device.os in ['web']`, `percent < 10`, `percent > 101`,
		`percent('bad seed') <= 10`, `percent('') <= 10`, `percent['seed'] <= 10`,
		`true&& false`, `true &&false`, `true&&false`,
		`version(app.version) > '1.0'`, `version(app.customSignal['v']).contains(['1'])`,
		`false && device.language == 'en'`,
	} {
		t.Run(expression, func(t *testing.T) {
			result, err := Evaluate(expression, &Context{Resolve: func(need Need) (Fact, error) {
				t.Fatalf("invalid expression prompted for %s", need.Key)
				return Fact{}, nil
			}})
			if err != nil || result.State != Unknown || len(result.Steps) != 1 || result.Steps[0].Expression != expression {
				t.Fatalf("invalid expression was evaluated: %#v, %v", result, err)
			}
		})
	}
}

func TestFullRE2SupportsSingleByteMatchingInUTF8Values(t *testing.T) {
	for expression, want := range map[string]State{
		`app.customSignal['name'].matches(['^\\C$'])`:        False,
		`app.customSignal['name'].matches(['^\\C\\C$'])`:     True,
		`app.customSignal['name'].matches(['^\\p{Latin}$'])`: True,
	} {
		result, err := Evaluate(expression, &Context{Facts: facts(map[string]string{"app.customSignal['name']": "é"})})
		if err != nil || result.State != want {
			t.Fatalf("RE2 mismatch %s: %#v, %v", expression, result, err)
		}
	}
}

func TestLanguageNormalizationAndUnspecifiedRegions(t *testing.T) {
	for _, test := range []struct {
		actual, target string
		want           State
	}{
		{"pt_BR", "pt-BR", True}, {"EN_us", "en-US", True}, {"iw_IL", "he-IL", True},
		{"en-US", "en", Unknown}, {"en-US", "fr", False}, {"pt-BR", "pt-PT", Unknown},
		{"pt", "pt-BR", Unknown}, {"invalid_tag_!", "en", Unknown},
	} {
		expression := "device.language in [" + quote(test.target) + "]"
		result, err := Evaluate(expression, &Context{Facts: facts(map[string]string{"device.language": test.actual})})
		if err != nil || result.State != test.want {
			t.Fatalf("%s in %s: %#v, %v", test.actual, test.target, result, err)
		}
	}
}

func TestPercentileCertaintiesDoNotNeedAssignmentAndOtherValuesExplainWhy(t *testing.T) {
	for expression, want := range map[string]State{`percent <= 100`: True, `percent > 100`: False, `percent('rollout') <= 100`: True} {
		result, err := Evaluate(expression, &Context{Resolve: func(Need) (Fact, error) {
			t.Fatal("asked for a mathematically unnecessary assignment")
			return Fact{}, nil
		}})
		if err != nil || result.State != want {
			t.Fatalf("%s: %#v, %v", expression, result, err)
		}
	}
	result, _ := Evaluate(`percent('rollout') <= 10`, &Context{})
	if result.State != Unknown || !strings.Contains(result.Steps[0].Note, "not verified") {
		t.Fatalf("guessed an unverified client hash: %#v", result)
	}
	result, _ = Evaluate(`percent('rollout') <= 10`, &Context{Facts: facts(map[string]string{"percent['rollout']": "5"})})
	if result.State != True {
		t.Fatal("explicit verified percentile was ignored")
	}
}

func TestPlatformDerivationUsesAppIDResolvedDuringEvaluation(t *testing.T) {
	var asked []string
	result, err := Evaluate(`app.id == '1:123:web:abc' && device.os == 'web'`, &Context{Resolve: func(need Need) (Fact, error) {
		asked = append(asked, need.Key)
		return Fact{Known: true, Value: "1:123:web:abc", Source: "your answer"}, nil
	}})
	if err != nil || result.State != True || len(asked) != 1 || asked[0] != "app.id" {
		t.Fatalf("platform was not lazily derived: %#v, %v, prompts %v", result, err, asked)
	}
}

func TestConsoleNumericMethodsAgreeWithInfixComparisons(t *testing.T) {
	for _, operator := range numericOperators {
		for _, field := range []string{"app.build", "app.version", "app.userProperty['score']", "app.customSignal['score']"} {
			ctx := &Context{Facts: facts(map[string]string{field: "2"})}
			infix := field + " " + operator + " '3'"
			method := field + "." + operator + "(['3'])"
			a, err := Evaluate(infix, ctx)
			if err != nil || a.State == Unknown {
				t.Fatalf("invalid baseline %s: %#v, %v", infix, a, err)
			}
			b, err := Evaluate(method, ctx)
			if err != nil || b.State != a.State || b.Steps[0].Expression != method || b.Steps[0].Note != a.Steps[0].Note {
				t.Fatalf("console method %s differs from %s: %#v, %#v, %v", method, infix, a, b, err)
			}
		}
	}
	for expression, key := range map[string]string{
		`version(app.customSignal['release']).>=(['1.9'])`:                         "app.customSignal['release']",
		`app.browserAndVersion.inOne([browserName('Chrome').version.>=(['1.9'])])`: "app.browserAndVersion.version",
	} {
		result, err := Evaluate(expression, &Context{Facts: facts(map[string]string{key: "1.10", "app.browserAndVersion.name": "Chrome"})})
		if err != nil || result.State != True {
			t.Fatalf("nested console version syntax failed: %#v, %v", result, err)
		}
	}
}

func TestUncertainObservationAnswersAreMemoized(t *testing.T) {
	count := 0
	ctx := &Context{Facts: map[string]Fact{"device.country": {Value: "US", Source: "request locale", Reason: "location is not confirmed"}}, Resolve: func(Need) (Fact, error) { count++; return Fact{}, nil }}
	for range 3 {
		result, err := Evaluate(`device.country in ['US']`, ctx)
		if err != nil || result.State != Unknown || !strings.Contains(result.Steps[0].Note, "location is not confirmed") {
			t.Fatalf("lost observation reason: %#v, %v", result, err)
		}
	}
	if count != 1 {
		t.Fatalf("asked about the same uncertain observation %d times", count)
	}
}
