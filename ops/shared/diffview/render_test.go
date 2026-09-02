package diffview

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"

	"github.com/yumauri/fbrcm/core/dictdiff"
)

func TestRenderUsesNaturalWidthAndCompleteStaticBody(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	result := compared(t, dictdiff.Input{
		EntityName: "Property: WEB / greeting",
		Left: dictdiff.NamedDictionary{
			Name:       "From: version 1",
			Properties: dictdiff.Dictionary{"value": dictdiff.String("hello")},
		},
		Right: dictdiff.NamedDictionary{
			Name:       "To: version 2",
			Properties: dictdiff.Dictionary{"value": dictdiff.String("hullo")},
		},
	})

	output := Render([]dictdiff.Result{result}, 100)
	for _, want := range []string{
		"Property: WEB / greeting",
		"value",
		"hello",
		"hullo",
		"│",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("Render() = %q, want substring %q", output, want)
		}
	}
	for _, unwanted := range []string{"From: version 1", "To: version 2", "┌", "┐", "└", "┘"} {
		if strings.Contains(output, unwanted) {
			t.Fatalf("Render() = %q, do not want substring %q", output, unwanted)
		}
	}
	for line := range strings.SplitSeq(output, "\n") {
		if got := lipgloss.Width(line); got >= 100 {
			t.Fatalf("natural-width line = %d columns, want less than terminal width: %q", got, line)
		}
	}
}

func TestRenderWrapsWithinNarrowTerminal(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	result := compared(t, dictdiff.Input{
		EntityName: "Property: a very long entity name",
		Left: dictdiff.NamedDictionary{
			Name:       "From: version 100",
			Properties: dictdiff.Dictionary{"value": dictdiff.String(strings.Repeat("left ", 20))},
		},
		Right: dictdiff.NamedDictionary{
			Name:       "To: version 200",
			Properties: dictdiff.Dictionary{"value": dictdiff.String(strings.Repeat("right ", 20))},
		},
	})

	const width = 42
	output := Render([]dictdiff.Result{result}, width)
	if !strings.Contains(output, "\n") {
		t.Fatalf("Render() = %q, want wrapped output", output)
	}
	for line := range strings.SplitSeq(output, "\n") {
		if got := lipgloss.Width(line); got > width {
			t.Fatalf("narrow line = %d columns, want <= %d: %q", got, width, line)
		}
	}
}

func compared(t *testing.T, input dictdiff.Input) dictdiff.Result {
	t.Helper()
	result, err := dictdiff.Compare(input)
	if err != nil {
		t.Fatal(err)
	}
	return result
}
