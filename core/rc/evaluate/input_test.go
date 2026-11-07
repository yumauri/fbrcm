package evaluate

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
)

const fetchURL = "https://firebaseremoteconfig.googleapis.com/v1/projects/123456789/namespaces/firebase:fetch?key=secret-api-key"

func TestCopiedCURLVariants(t *testing.T) {
	for _, command := range []string{
		`curl '` + fetchURL + `' -H 'Content-Type: application/json' --data-raw '{"app_id":"1:123456789:web:abc","language_code":"en-US","app_instance_id_token":"secret-token","custom_signals":{"score":3.5,"active":true}}'`,
		"curl '" + fetchURL + "' \\\n  --compressed -X POST --data-binary $'{\"app_id\":\"1:123456789:web:abc\",\"language_code\":\"en-US\",\"custom_signals\":{\"score\":3.5,\"active\":true}}'",
		`curl --url="` + fetchURL + `" --data "{\"app_id\":\"1:123456789:web:abc\",\"language_code\":\"en-US\",\"custom_signals\":{\"score\":3.5,\"active\":true}}"`,
	} {
		exchanges, err := ReadCapture(strings.NewReader(command))
		if err != nil || len(exchanges) != 1 {
			t.Fatalf("parse copied cURL: %v", err)
		}
		exchange := exchanges[0]
		if exchange.Project != "123456789" || exchange.Facts["app.id"].Value != "1:123456789:web:abc" || exchange.Facts["app.customSignal['score']"].Value != "3.5" {
			t.Fatalf("wrong normalized request: %#v", exchange)
		}
		encoded, _ := json.Marshal(exchange)
		if strings.Contains(string(encoded), "secret") {
			t.Fatalf("captured credentials leaked into evaluation context: %s", encoded)
		}
	}
}

func TestUnsafeOrIncompleteCURLIsRejectedWithoutEchoingBody(t *testing.T) {
	for _, command := range []string{
		`curl '` + fetchURL + `' --data-raw '{"secret":"sensitive"}'; touch /tmp/should-not-exist`,
		`curl '` + fetchURL + `' --data "$(cat /etc/passwd)"`,
		`curl '` + fetchURL + `' --data-binary @sensitive.json`,
		`curl '` + fetchURL + `' --data-raw 'sensitive'`,
		`curl 'https://example.com/firebase:fetch' --data '{}'`,
		`curl '` + fetchURL + `' -H`,
		`curl '` + fetchURL + `' --data 'unterminated`,
		`echo sensitive`,
	} {
		_, err := ReadCapture(strings.NewReader(command))
		if err == nil || strings.Contains(err.Error(), "sensitive") {
			t.Fatalf("expected safe parse failure: %v", err)
		}
	}
}

func TestHARFindsFetchesAndPairsResponseBodies(t *testing.T) {
	var compressed bytes.Buffer
	writer := gzip.NewWriter(&compressed)
	if _, err := writer.Write([]byte(`{"entries":{"new_player":"false"},"templateVersion":"142"}`)); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	archive := `{"log":{"version":"1.2","entries":[
	{"request":{"method":"POST","url":"https://example.com/other"}},
	{"startedDateTime":"2026-11-04T14:32:07Z","request":{"method":"POST","url":"` + fetchURL + `","postData":{"text":"{\"appId\":\"1:123456789:android:abc\",\"appVersion\":\"3.8.0\",\"analyticsUserProperties\":{\"tier\":\"paid\"}}"}},"response":{"status":200,"content":{"encoding":"base64","text":"` + base64.StdEncoding.EncodeToString(compressed.Bytes()) + `"}}},
	{"request":{"method":"POST","url":"` + fetchURL + `","postData":{"text":"{}"}},"response":{"status":304}}
	]}}`
	exchanges, err := ReadCapture(strings.NewReader(archive))
	if err != nil || len(exchanges) != 2 {
		t.Fatalf("HAR parse: %#v, %v", exchanges, err)
	}
	first := exchanges[0]
	if first.Entry != 2 || first.Facts["app.version"].Value != "3.8.0" || first.Facts["app.userProperty['tier']"].Value != "paid" || !first.HasResponse || first.Entries["new_player"] != "false" || first.Version != "142" || first.Facts["device.dateTime"].Value != "2026-11-04T14:32:07Z" {
		t.Fatalf("lost HAR evidence: %#v", first)
	}
	if exchanges[1].HasResponse || exchanges[1].Status != 304 || exchanges[1].Entry != 3 {
		t.Fatalf("304 incorrectly treated as a new value: %#v", exchanges[1])
	}
}

func TestMissingBodiesAndNonHARJSONFailClearly(t *testing.T) {
	for _, input := range []string{`{}`, `null`, `{"log":{"entries":[]}}`, `{"log":{"entries":[{"request":{"method":"POST","url":"` + fetchURL + `"}}]}}`} {
		if _, err := ReadCapture(strings.NewReader(input)); err == nil {
			t.Fatalf("accepted incomplete capture %s", input)
		}
	}
}

func TestAnswerMarkersAndCanonicalKeys(t *testing.T) {
	need, ok := InputNeed(`app.userProperty["tier"]`)
	if !ok || need.Key != "app.userProperty['tier']" {
		t.Fatalf("input normalization = %#v", need)
	}
	for _, answer := range []string{"", "/unknown", "/unset"} {
		fact, err := Answer(need, answer, "your answer")
		if err != nil || fact.Known != (answer != "/unknown") || fact.Absent != (answer == "/unset") {
			t.Fatalf("answer %q = %#v, %v", answer, fact, err)
		}
	}
}
