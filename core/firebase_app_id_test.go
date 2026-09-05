package core

import (
	"reflect"
	"testing"
)

func TestParseFirebaseAppID(t *testing.T) {
	got, ok := ParseFirebaseAppID("1:849939820737:web:03250aec28d4830d4d2440")
	if !ok {
		t.Fatal("ParseFirebaseAppID() rejected a complete App ID")
	}
	want := FirebaseAppID{Version: "1", ProjectNumber: "849939820737", Platform: AppPlatformWeb, Hash: "03250aec28d4830d4d2440"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseFirebaseAppID() = %#v, want %#v", got, want)
	}
}

func TestParseFirebaseAppIDRequiresEntireValue(t *testing.T) {
	for _, value := range []string{"web", "1:123:web:hash suffix", "prefix 1:123:web:hash", "1:123:desktop:hash"} {
		if got, ok := ParseFirebaseAppID(value); ok {
			t.Errorf("ParseFirebaseAppID(%q) = %#v, true; want false", value, got)
		}
	}
}

func TestFirebaseAppIDPlatformSpansFindsEverySupportedPlatform(t *testing.T) {
	value := "app.id in ['1:123:android:a1', '1:123:ios:B2', '2:456:web:c3']"
	spans := FirebaseAppIDPlatformSpans(value)
	if len(spans) != 3 {
		t.Fatalf("FirebaseAppIDPlatformSpans() returned %d spans, want 3: %#v", len(spans), spans)
	}
	wantPlatforms := []AppPlatform{AppPlatformAndroid, AppPlatformIOS, AppPlatformWeb}
	for index, span := range spans {
		if got := value[span.Start:span.End]; got != string(wantPlatforms[index]) {
			t.Errorf("span %d text = %q, want %q", index, got, wantPlatforms[index])
		}
		if span.Platform != wantPlatforms[index] {
			t.Errorf("span %d platform = %q, want %q", index, span.Platform, wantPlatforms[index])
		}
	}
}

func TestFirebaseAppIDPlatformSpansRequiresCompleteAppID(t *testing.T) {
	tests := []string{
		"android",
		"1:123:android",
		"1:123:android:",
		"version:123:web:hash",
		"1:project:web:hash",
		"1:123:desktop:hash",
		"prefix1:123:web:hash",
		"1:123:web:hash_suffix",
	}
	for _, value := range tests {
		if got := FirebaseAppIDPlatformSpans(value); len(got) != 0 {
			t.Errorf("FirebaseAppIDPlatformSpans(%q) = %#v, want no spans", value, got)
		}
	}
}

func TestFirebaseAppIDPlatformSpansReturnsEmptyNonNilSlice(t *testing.T) {
	if got := FirebaseAppIDPlatformSpans("true"); !reflect.DeepEqual(got, []FirebaseAppIDPlatformSpan{}) {
		t.Fatalf("FirebaseAppIDPlatformSpans(true) = %#v, want empty non-nil slice", got)
	}
}
