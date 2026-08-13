package shared

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/spf13/cobra"
)

func TestWriteJSONNormalizesNilTopLevelSlice(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	var output bytes.Buffer
	cmd.SetOut(&output)
	var values []string
	if err := WriteJSON(cmd, values); err != nil {
		t.Fatal(err)
	}
	if got := output.String(); got != "[]\n" {
		t.Fatalf("output = %q, want an empty JSON array", got)
	}
}

func TestResolveParameterArgument(t *testing.T) {
	got, err := ResolveParameterArgument([]string{"Flag"}, nil)
	if err != nil {
		t.Fatalf("ResolveParameterArgument returned error: %v", err)
	}
	if got == nil || *got != "Flag" {
		t.Fatalf("argument = %#v, want literal Flag", got)
	}
}

func TestResolveParameterArgumentRejectsExistingFilter(t *testing.T) {
	_, err := ResolveParameterArgument([]string{"flag"}, []string{"foo"})
	if err == nil {
		t.Fatalf("ResolveParameterArgument accepted parameter arg with existing filter")
	}
}

func TestResolveParameterArgumentReturnsNilWithoutArg(t *testing.T) {
	got, err := ResolveParameterArgument(nil, []string{"foo"})
	if err != nil {
		t.Fatalf("ResolveParameterArgument returned error: %v", err)
	}
	if got != nil {
		t.Fatalf("argument = %#v, want nil", got)
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

func TestOpenPromptInputUsesControllingTerminalForRedirectedStdin(t *testing.T) {
	redirected, err := os.CreateTemp(t.TempDir(), "stdin")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = redirected.Close()
	}()

	ttyIn, ttyWriter, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = ttyWriter.Close()
	}()
	ttyReader, ttyOut, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = ttyReader.Close()
	}()

	originalOpenPromptTTY := openPromptTTY
	openPromptTTY = func() (*os.File, *os.File, error) {
		return ttyIn, ttyOut, nil
	}
	t.Cleanup(func() {
		openPromptTTY = originalOpenPromptTTY
	})

	input, closeInput, err := OpenPromptInput(redirected)
	if err != nil {
		t.Fatalf("OpenPromptInput returned error: %v", err)
	}
	if input != ttyIn {
		t.Fatalf("prompt input = %v, want controlling terminal input", input)
	}
	closeInput()
	if err := ttyIn.Close(); err == nil {
		t.Fatalf("prompt input was not closed")
	}
	if err := ttyOut.Close(); err == nil {
		t.Fatalf("prompt output was not closed")
	}
}

func TestOpenPromptInputKeepsInjectedInput(t *testing.T) {
	in := io.NopCloser(nil)
	input, closeInput, err := OpenPromptInput(in)
	if err != nil {
		t.Fatalf("OpenPromptInput returned error: %v", err)
	}
	defer closeInput()
	if input != in {
		t.Fatalf("prompt input = %v, want injected input", input)
	}
}
