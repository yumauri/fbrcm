package shared

import (
	"io"
	"testing"
)

func TestResolveParameterArgFilters(t *testing.T) {
	got, err := ResolveParameterArgFilters([]string{"flag"}, nil)
	if err != nil {
		t.Fatalf("ResolveParameterArgFilters returned error: %v", err)
	}
	if len(got) != 1 || got[0] != "=flag" {
		t.Fatalf("filters = %#v, want =flag", got)
	}
}

func TestResolveParameterArgFiltersRejectsExistingFilter(t *testing.T) {
	_, err := ResolveParameterArgFilters([]string{"flag"}, []string{"foo"})
	if err == nil {
		t.Fatalf("ResolveParameterArgFilters accepted parameter arg with existing filter")
	}
}

func TestResolveParameterArgFiltersKeepsFiltersWithoutArg(t *testing.T) {
	in := []string{"foo"}
	got, err := ResolveParameterArgFilters(nil, in)
	if err != nil {
		t.Fatalf("ResolveParameterArgFilters returned error: %v", err)
	}
	if len(got) != 1 || got[0] != "foo" {
		t.Fatalf("filters = %#v, want original filters", got)
	}
}

func TestHasFiltersDropsEmptyQueries(t *testing.T) {
	if HasFilters([]string{"", "  "}) {
		t.Fatalf("HasFilters returned true for empty queries")
	}
	if !HasFilters([]string{"flag"}) {
		t.Fatalf("HasFilters returned false for non-empty query")
	}
}

func TestStdinAvailableRejectsNonFileReader(t *testing.T) {
	if StdinAvailable(io.NopCloser(nil)) {
		t.Fatalf("StdinAvailable returned true for non-file reader")
	}
}
