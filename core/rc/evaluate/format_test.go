package evaluate

import "testing"

func TestFormatExpressionPreservesQuotedConjunctionsAndUnsupportedSyntax(t *testing.T) {
	for _, test := range []struct {
		expression string
		want       string
	}{
		{`app.id == '1:123:android:abc' && app.version.contains(['2.10.261.0', '3.11.272'])`, "app.id == '1:123:android:abc'\napp.version.contains(['2.10.261.0', '3.11.272'])"},
		{`app.customSignal['text'].exactlyMatches(['a && b']) && device.os == 'web'`, "app.customSignal['text'].exactlyMatches(['a && b'])\ndevice.os == 'web'"},
		{"device.os == 'web' &&\n  app.version == '2' && true", "device.os == 'web'\napp.version == '2'\ntrue"},
		{`app.id == 'x' || app.id == 'y' && true`, `app.id == 'x' || app.id == 'y' && true`},
		{`app.id == 'unterminated && true`, `app.id == 'unterminated && true`},
	} {
		t.Run(test.expression, func(t *testing.T) {
			if got := FormatExpression(test.expression); got != test.want {
				t.Fatalf("FormatExpression = %q, want %q", got, test.want)
			}
		})
	}
}
