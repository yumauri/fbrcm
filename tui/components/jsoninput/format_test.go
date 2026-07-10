package jsoninput

import (
	"strings"
	"testing"
)

func TestPrettyJSON(t *testing.T) {
	got := PrettyJSON(`{"a":1}`)
	want := "{\n  \"a\": 1\n}"
	if got != want {
		t.Fatalf("PrettyJSON compact object = %q, want %q", got, want)
	}
	if PrettyJSON("not json") != "not json" {
		t.Fatal("invalid JSON should be returned unchanged")
	}
}

func TestIsValidJSON(t *testing.T) {
	if !IsValidJSON(`{"ok":true}`) {
		t.Fatal("expected valid JSON")
	}
	if IsValidJSON("{bad") {
		t.Fatal("expected invalid JSON")
	}
}

func TestCompactJSON(t *testing.T) {
	got, ok := CompactJSON("{\n  \"a\": 1\n}\n")
	if !ok {
		t.Fatal("expected compact to succeed")
	}
	if got != `{"a":1}` {
		t.Fatalf("CompactJSON = %q, want %q", got, `{"a":1}`)
	}
	if _, ok := CompactJSON("{"); ok {
		t.Fatal("expected compact to fail for invalid JSON")
	}
}

func TestModelCompactedValue(t *testing.T) {
	m := New()
	m.open = true
	m.area.SetValue("{\n  \"x\": \"y\"\n}\n")
	if !m.Valid() {
		t.Fatal("expected model value to be valid JSON")
	}
	got, ok := m.CompactedValue()
	if !ok {
		t.Fatal("expected CompactedValue to succeed")
	}
	if got != `{"x":"y"}` {
		t.Fatalf("CompactedValue = %q, want %q", got, `{"x":"y"}`)
	}
	m.area.SetValue("{")
	if _, ok := m.CompactedValue(); ok {
		t.Fatal("expected CompactedValue to fail for invalid JSON")
	}
}

func TestWrapPlainLine(t *testing.T) {
	parts := wrapPlainLine("hello world", 5)
	if len(parts) == 0 {
		t.Fatal("expected wrapped parts")
	}
	var combined strings.Builder
	combined.WriteString(strings.TrimSpace(parts[0].text))
	for i := 1; i < len(parts); i++ {
		combined.WriteString(strings.TrimSpace(parts[i].text))
	}
	if combined.String() != "helloworld" {
		t.Fatalf("wrapped text lost content: %q", combined.String())
	}
}
