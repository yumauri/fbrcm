package evaluate

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

func capturedCURL(t *testing.T, body string, headers ...string) Exchange {
	t.Helper()
	command := "curl '" + fetchURL + "'"
	for _, header := range headers {
		command += " -H '" + strings.ReplaceAll(header, "'", `'\''`) + "'"
	}
	command += " --data-raw '" + strings.ReplaceAll(body, "'", `'\''`) + "'"
	exchanges, err := ReadCapture(strings.NewReader(command))
	if err != nil {
		t.Fatal(err)
	}
	return exchanges[0]
}

func capturedHAR(t *testing.T, entry harEntry) Exchange {
	t.Helper()
	entry.Request.URL, entry.Request.Method = fetchURL, "POST"
	data, err := json.Marshal(struct{ Log struct{ Entries []harEntry } }{Log: struct{ Entries []harEntry }{Entries: []harEntry{entry}}})
	if err != nil {
		t.Fatal(err)
	}
	exchanges, err := ReadCapture(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	return exchanges[0]
}

func gzipCapture(t *testing.T, body string) []byte {
	t.Helper()
	var out bytes.Buffer
	w := gzip.NewWriter(&out)
	if _, err := w.Write([]byte(body)); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}

func TestAppleWireBodiesThroughCURLAndCompressedHAR(t *testing.T) {
	body := `{app_instance_id:'installation',app_instance_id_token:'secret-token',app_id:'1:123456789:ios:abc',language_code:'pt_BR',time_zone:'Europe/Lisbon',app_version:'4.0.0',app_build:'42',custom_signals:{name:"client's app"},analytics_user_properties:{tier:'paid',_fot:1762257600000}}`
	plain := capturedCURL(t, body)
	entry := harEntry{}
	entry.Request.PostData = harBody{Text: base64.StdEncoding.EncodeToString(gzipCapture(t, body)), Encoding: "base64"}
	compressed := capturedHAR(t, entry)
	var ansiBody strings.Builder
	for _, b := range gzipCapture(t, body) {
		fmt.Fprintf(&ansiBody, `\x%02x`, b)
	}
	command := "curl '" + fetchURL + "' --data-binary $'" + ansiBody.String() + "'"
	exchanges, err := ReadCapture(strings.NewReader(command))
	if err != nil {
		t.Fatal(err)
	}
	for _, exchange := range []Exchange{plain, compressed, exchanges[0]} {
		if exchange.Facts["app.version"].Value != "4.0.0" || exchange.Facts["app.customSignal['name']"].Value != "client's app" || !exchange.Facts["app.firstOpenTimestamp"].Known || exchange.TimeZone != "Europe/Lisbon" {
			t.Fatalf("lost Apple wire data: %#v", exchange)
		}
		ctx := &Context{Facts: exchange.Facts, Capture: &exchange, Resolve: func(need Need) (Fact, error) { t.Fatalf("unnecessary prompt: %s", need.Key); return Fact{}, nil }}
		for _, expression := range []string{`device.os == 'ios'`, `device.language in ['pt-BR']`, `app.userProperty['tier'].exactlyMatches(['paid'])`, `app.customSignal['name'].exactlyMatches(["client's app"])`} {
			result, err := Evaluate(expression, ctx)
			if err != nil || result.State != True {
				t.Fatalf("%s: %#v, %v", expression, result, err)
			}
		}
		encoded, _ := json.Marshal(exchange)
		if bytes.Contains(encoded, []byte("secret-token")) {
			t.Fatal("retained an installation credential")
		}
	}
}

func TestFetchBodyNormalizerNeverAcceptsExecutableExpressions(t *testing.T) {
	for _, body := range []string{`{app_id:alert('secret')}`, `{app_id:'a'+ 'b'}`, `{app_id:process.env.secret}`, `{app_id:undefined}`, `{app_id:'unterminated}`, `{app_id:NaN}`} {
		if _, err := requestFacts([]byte(body)); err == nil {
			t.Fatalf("accepted executable or invalid body %q", body)
		}
	}
}

func TestCapturedSignalsAreAbsentWhileWebAnalyticsRemainsUnknown(t *testing.T) {
	exchange := capturedCURL(t, `{"app_id":"1:123:web:abc","custom_signals":{"empty":""}}`)
	ctx := &Context{Facts: exchange.Facts, Capture: &exchange}
	for expression, want := range map[string]State{
		`app.customSignal['missing'].notContains(['x'])`:    False,
		`app.customSignal['missing'].exactlyMatches([''])`:  False,
		`app.customSignal['empty'].exactlyMatches([''])`:    True,
		`app.userProperty['missing'].exactlyMatches(['x'])`: Unknown,
	} {
		result, err := Evaluate(expression, ctx)
		if err != nil || result.State != want {
			t.Fatalf("%s = %#v, %v", expression, result, err)
		}
	}
	ctx.Facts["app.customSignal['missing']"] = Fact{Known: true, Value: "x", Source: "command line"}
	result, _ := Evaluate(`app.customSignal['missing'].exactlyMatches(['x'])`, ctx)
	if result.State != True {
		t.Fatal("explicit signal did not override known absence")
	}
	native := capturedCURL(t, `{"appId":"1:123:android:abc","analyticsUserProperties":{}}`)
	result, _ = Evaluate(`app.userProperty['missing'].exactlyMatches(['x'])`, &Context{Facts: native.Facts, Capture: &native})
	if result.State != False {
		t.Fatal("complete native property snapshot did not establish absence")
	}
}

func TestCountryObservationRequiresConfirmationAndDoesNotInferGeolocation(t *testing.T) {
	exchange := capturedCURL(t, `{"app_id":"1:123:web:abc","country_code":"US","language_code":"en-US","time_zone":"America/New_York"}`)
	ctx := &Context{Facts: exchange.Facts, Capture: &exchange}
	result, _ := Evaluate(`device.country in ['US']`, ctx)
	if result.State != Unknown || !strings.Contains(result.Steps[0].Note, "locale country") {
		t.Fatalf("locale country became geolocation: %#v", result)
	}
	ctx.Facts["device.country"] = Fact{Known: true, Value: "DE", Source: "command line"}
	result, _ = Evaluate(`device.country in ['DE']`, ctx)
	if result.State != True {
		t.Fatal("country override was lost")
	}
	noCountry := capturedCURL(t, `{"app_id":"1:123:web:abc","language_code":"en-US","time_zone":"America/New_York"}`)
	if noCountry.Facts["device.country"].Known {
		t.Fatal("guessed country from language or time zone")
	}
}

func TestFirstOpenDedicatedValueWinsOverAnalyticsFallback(t *testing.T) {
	exchange := capturedCURL(t, `{"appId":"1:123:android:abc","firstOpenTime":"2020-01-01T00:00:00Z","analyticsUserProperties":{"_fot":"1762257600000"}}`)
	if exchange.Facts["app.firstOpenTimestamp"].Value != "2020-01-01T00:00:00Z" {
		t.Fatal("fallback replaced explicit first-open timestamp")
	}
	for _, platform := range []string{"ios", "android"} {
		body := fmt.Sprintf(`{"appId":"1:123:%s:abc","analyticsUserProperties":{"_fot":"1762257600123"}}`, platform)
		exchange = capturedCURL(t, body)
		got, err := time.Parse(time.RFC3339Nano, exchange.Facts["app.firstOpenTimestamp"].Value)
		if err != nil || platform == "ios" && got.Nanosecond() != 0 || platform == "android" && got.Nanosecond() != 123000000 {
			t.Fatalf("SDK first-open precision lost: %s, %s, %v", platform, got, err)
		}
	}
}

func TestResponseStatesAndManagedAssignments(t *testing.T) {
	for _, test := range []struct {
		body        string
		status      int
		hasResponse bool
	}{
		{`{"state":"NO_CHANGE","templateVersion":"142"}`, 304, false},
		{`{"state":"EMPTY_CONFIG"}`, 200, true},
		{`{"state":"NO_TEMPLATE"}`, 200, true},
		{`{"state":"INSTANCE_STATE_UNSPECIFIED"}`, 500, false},
	} {
		entry := harEntry{}
		entry.Request.PostData.Text = `{}`
		entry.Response.Status, entry.Response.Content.Text = 200, test.body
		exchange := capturedHAR(t, entry)
		if exchange.Status != test.status || exchange.HasResponse != test.hasResponse {
			t.Fatalf("%s: %#v", test.body, exchange)
		}
	}
	entry := harEntry{}
	entry.Request.PostData.Text = `{}`
	entry.Response.Status = 200
	entry.Response.Content.Text = `{"entries":{"a":"true","b":"false"},"templateVersion":142,"personalizationMetadata":{"a":{"choiceId":"1"}},"rolloutMetadata":[{"rolloutId":"r1","variantId":"v1","affectedParameterKeys":["b"]}],"experimentDescriptions":[{"experimentId":"e1"}]}`
	exchange := capturedHAR(t, entry)
	if exchange.Managed["a"] != "personalization" || exchange.Managed["b"] != "rollout r1, variant v1" || len(exchange.Experiments) != 1 || exchange.Version != "142" {
		t.Fatalf("lost managed response metadata: %#v", exchange)
	}
}

func TestCaptureTimeUsesServerHeaderAndKeepsBoundariesUnresolved(t *testing.T) {
	entry := harEntry{}
	entry.Request.PostData.Text = `{}`
	entry.StartedAt = time.Date(2026, 11, 5, 10, 0, 0, 0, time.UTC)
	duration := 250.0
	entry.Time = &duration
	entry.Response.Headers = []harHeader{{Name: "date", Value: "Thu, 05 Nov 2026 12:00:00 GMT"}}
	exchange := capturedHAR(t, entry)
	if exchange.Facts["device.dateTime"].Source != "response Date header" {
		t.Fatal("did not use server time observation")
	}
	ctx := &Context{Facts: exchange.Facts, Capture: &exchange}
	for expression, want := range map[string]State{
		`device.dateTime < '2026-11-05T12:01:00Z'`:  True,
		`device.dateTime > '2026-11-05T11:59:00Z'`:  True,
		`device.dateTime <= '2026-11-05T12:00:00Z'`: Unknown,
		`dateTime < '2026-11-05T12:01:00Z'`:         True,
		`dateTime > '2026-11-05T11:59:00Z'`:         True,
		`dateTime <= '2026-11-05T12:00:00Z'`:        Unknown,
	} {
		result, err := Evaluate(expression, ctx)
		if err != nil || result.State != want {
			t.Fatalf("%s: %#v, %v", expression, result, err)
		}
	}
	if _, err := http.ParseTime(entry.Response.Headers[0].Value); err != nil {
		t.Fatal(err)
	}
}
