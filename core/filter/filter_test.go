package filter

import "testing"

func TestParseModePrefixedQuery(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		wantMode  Mode
		wantQuery string
	}{
		{name: "empty", raw: "", wantMode: ModeFuzzy, wantQuery: ""},
		{name: "whitespace only", raw: "  \t ", wantMode: ModeFuzzy, wantQuery: ""},
		{name: "fuzzy implicit", raw: "alpha", wantMode: ModeFuzzy, wantQuery: "alpha"},
		{name: "fuzzy trimmed", raw: " alpha ", wantMode: ModeFuzzy, wantQuery: "alpha"},
		{name: "exact prefix", raw: "=alpha", wantMode: ModeExact, wantQuery: "alpha"},
		{name: "starts with prefix", raw: "^beta", wantMode: ModeStartsWith, wantQuery: "beta"},
		{name: "includes prefix", raw: "/gamma", wantMode: ModeIncludes, wantQuery: "gamma"},
		{name: "fuzzy explicit prefix", raw: "~delta", wantMode: ModeFuzzy, wantQuery: "delta"},
		{name: "exact prefix only", raw: "=", wantMode: ModeExact, wantQuery: ""},
		{name: "unknown prefix", raw: "@omega", wantMode: ModeFuzzy, wantQuery: "@omega"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMode, gotQuery := ParseModePrefixedQuery(tt.raw)
			if gotMode != tt.wantMode || gotQuery != tt.wantQuery {
				t.Fatalf("ParseModePrefixedQuery(%q) = %v/%q, want %v/%q", tt.raw, gotMode, gotQuery, tt.wantMode, tt.wantQuery)
			}
		})
	}
}
