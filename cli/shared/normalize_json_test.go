package shared

import "testing"

func TestNormalizeExportJSON(t *testing.T) {
	in := []byte(`{"parameters":{"flag":{"conditionalValues":{"beta":{"value":"b"},"1":{"value":"one"},"prod":{"value":"p"},"2":{"value":"two"}},"defaultValue":{"value":"\u003cok\u003e \u0026 more"}}}}`)
	want := `{"parameters":{"flag":{"conditionalValues":{"1":{"value":"one"},"2":{"value":"two"},"beta":{"value":"b"},"prod":{"value":"p"}},"defaultValue":{"value":"<ok> & more"}}}}`

	if got := string(NormalizeExportJSON(in)); got != want {
		t.Fatalf("NormalizeExportJSON() = %s, want %s", got, want)
	}
}

func TestNormalizeJSONEscapes(t *testing.T) {
	got := string(NormalizeJSONEscapes([]byte(`"\u003cok\u003e \u0026 more"`)))
	want := `"<ok> & more"`
	if got != want {
		t.Fatalf("NormalizeJSONEscapes() = %s, want %s", got, want)
	}
}

func TestTrimTrailingLineBreaks(t *testing.T) {
	got := string(TrimTrailingLineBreaks([]byte("body\r\n\n")))
	if got != "body" {
		t.Fatalf("TrimTrailingLineBreaks() = %q, want %q", got, "body")
	}
}
