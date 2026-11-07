package evaluate

import (
	"reflect"
	"testing"
)

func TestComparisonValueSpansDistinguishValuesFromInputNames(t *testing.T) {
	for _, test := range []struct {
		text string
		want []string
	}{
		{`app.customSignal['name'].exactlyMatches(['Tria', 'client\'s app']) && app.userProperty['tier'] == 'paid'`, []string{`'Tria'`, `'client\'s app'`, `'paid'`}},
		{`percent('seed') <= 50 && percent['seed'] > -1.5 && true && false`, []string{"50", "-1.5", "true", "false"}},
		{`app.customSignal['name'] = '日本語 (request)' (command line) exactlyMatches '日本語'`, []string{`'日本語 (request)'`, `'日本語'`}},
		{`device.os == "web"`, []string{`"web"`}},
		{`app.customSignal['unterminated`, nil},
	} {
		t.Run(test.text, func(t *testing.T) {
			var got []string
			for _, span := range ComparisonValueSpans(test.text) {
				got = append(got, test.text[span.Start:span.End])
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("comparison literals = %q, want %q", got, test.want)
			}
		})
	}
}
