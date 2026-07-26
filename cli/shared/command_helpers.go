package shared

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"
)

// WriteJSON encodes v as indented JSON to the command's stdout. Callers wrap
// the returned error with their own context when needed.
func WriteJSON(cmd *cobra.Command, v any) error {
	encoder := json.NewEncoder(cmd.OutOrStdout())
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}

// HasFilters reports whether any filter query is non-empty after parsing.
func HasFilters(rawFilters []string) bool {
	return len(ParseFilters(rawFilters)) > 0
}

// ResolveParameterArgFilters turns an optional parameter argument into an exact filter.
func ResolveParameterArgFilters(args []string, rawFilters []string) ([]string, error) {
	if len(args) == 0 {
		return rawFilters, nil
	}
	if HasFilters(rawFilters) {
		return nil, fmt.Errorf("parameter argument cannot be used together with --filter")
	}
	return []string{"=" + args[0]}, nil
}

// StdinAvailable reports whether the given reader is a non-terminal file.
func StdinAvailable(in io.Reader) bool {
	file, ok := in.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) == 0
}

var openPromptTTY = tea.OpenTTY

// OpenPromptInput returns the command input for interactive prompts. When stdin
// is redirected, it opens the controlling terminal so prompts do not try to
// read from the redirected data stream.
func OpenPromptInput(in io.Reader) (io.Reader, func(), error) {
	if !StdinAvailable(in) {
		return in, func() {}, nil
	}

	ttyIn, ttyOut, err := openPromptTTY()
	if err != nil {
		return nil, func() {}, fmt.Errorf("open terminal for prompt: %w", err)
	}
	return ttyIn, func() {
		_ = ttyIn.Close()
		_ = ttyOut.Close()
	}, nil
}
