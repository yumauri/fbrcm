package firebase

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestRemoteConfigValueRoundTripsEveryValueOption(t *testing.T) {
	tests := []struct {
		name string
		raw  string
	}{
		{name: "plain", raw: `{"value":"hello"}`},
		{name: "empty plain", raw: `{"value":""}`},
		{name: "in-app default", raw: `{"useInAppDefault":true}`},
		{name: "personalization", raw: `{"personalizationValue":{"personalizationId":"personalization-1","futureMetadata":{"enabled":true}}}`},
		{name: "experiment", raw: `{"experimentValue":{"experimentId":"experiment-1","exposurePercent":12.5,"variantValue":[{"variantId":"0","noChange":true},{"variantId":"1","value":""}]}}`},
		{name: "rollout", raw: `{"rolloutValue":{"rolloutId":"rollout-1","value":"enabled","percent":10,"futureMetadata":["a","b"]}}`},
		{name: "future object", raw: `{"someFutureValue":{"id":"future-1","settings":{"enabled":true}}}`},
		{name: "future scalar", raw: `{"someFutureScalar":"future"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var value RemoteConfigValue
			if err := json.Unmarshal([]byte(tt.raw), &value); err != nil {
				t.Fatalf("UnmarshalJSON returned error: %v", err)
			}
			got, err := json.Marshal(value)
			if err != nil {
				t.Fatalf("MarshalJSON returned error: %v", err)
			}
			assertEquivalentJSON(t, got, []byte(tt.raw))
		})
	}
}

func TestRemoteConfigValueRejectsInvalidUnion(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantErr string
	}{
		{name: "empty object", raw: `{}`, wantErr: "expected exactly one value option"},
		{name: "multiple options", raw: `{"value":"on","rolloutValue":{"rolloutId":"rollout-1"}}`, wantErr: "expected exactly one value option"},
		{name: "false in-app default", raw: `{"useInAppDefault":false}`, wantErr: "must be true"},
		{name: "managed value is not object", raw: `{"experimentValue":"experiment-1"}`, wantErr: "cannot unmarshal string"},
		{name: "managed value is null", raw: `{"rolloutValue":null}`, wantErr: "expected object"},
		{name: "plain value is null", raw: `{"value":null}`, wantErr: "cannot be null"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var value RemoteConfigValue
			err := json.Unmarshal([]byte(tt.raw), &value)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("UnmarshalJSON error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestRemoteConfigUnknownValueIsOpaque(t *testing.T) {
	var value RemoteConfigValue
	if err := json.Unmarshal([]byte(`{"someFutureValue":{"id":"future-1"}}`), &value); err != nil {
		t.Fatal(err)
	}
	if value.IsPlain() || value.IsManaged() || !value.IsOpaque() {
		t.Fatalf("future value classification = plain:%v managed:%v opaque:%v", value.IsPlain(), value.IsManaged(), value.IsOpaque())
	}
	if value.UnknownValueOption != "someFutureValue" {
		t.Fatalf("unknown option = %q", value.UnknownValueOption)
	}
}

func TestRemoteConfigValueMarshalRejectsMultipleOptions(t *testing.T) {
	value := RemoteConfigValue{
		PersonalizationValue: json.RawMessage(`{"personalizationId":"personalization-1"}`),
		RolloutValue:         json.RawMessage(`{"rolloutId":"rollout-1"}`),
	}
	if _, err := json.Marshal(value); err == nil || !strings.Contains(err.Error(), "expected exactly one value option") {
		t.Fatalf("MarshalJSON error = %v, want union error", err)
	}
}

func TestRemoteConfigValueMarshalRejectsIncompleteUnknownOption(t *testing.T) {
	for _, value := range []RemoteConfigValue{
		{UnknownValueOption: "someFutureValue"},
		{UnknownValue: json.RawMessage(`{"id":"future-1"}`)},
		{UnknownValueOption: "rolloutValue", UnknownValue: json.RawMessage(`{"rolloutId":"rollout-1"}`)},
	} {
		if _, err := json.Marshal(value); err == nil {
			t.Fatalf("MarshalJSON(%#v) returned nil error", value)
		}
	}
}

func TestRemoteConfigValueZeroValueEncodesEmptyString(t *testing.T) {
	got, err := json.Marshal(RemoteConfigValue{})
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"value":""}` {
		t.Fatalf("zero value = %s, want empty string value", got)
	}
}

func assertEquivalentJSON(t *testing.T, got, want []byte) {
	t.Helper()
	var gotValue any
	if err := json.Unmarshal(got, &gotValue); err != nil {
		t.Fatalf("decode got JSON: %v\n%s", err, got)
	}
	var wantValue any
	if err := json.Unmarshal(want, &wantValue); err != nil {
		t.Fatalf("decode want JSON: %v\n%s", err, want)
	}
	if !reflect.DeepEqual(gotValue, wantValue) {
		t.Fatalf("JSON mismatch\ngot:  %s\nwant: %s", got, want)
	}
}
