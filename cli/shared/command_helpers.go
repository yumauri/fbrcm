package shared

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"

	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"
)

// CommandContext returns Cobra's root-owned invocation context and falls back
// only for standalone command tests and embedders that did not set one.
func CommandContext(cmd *cobra.Command) context.Context {
	if cmd != nil && cmd.Context() != nil {
		return cmd.Context()
	}
	return context.Background()
}

// WriteJSON encodes v as indented JSON to the command's stdout. Callers wrap
// the returned error with their own context when needed.
func WriteJSON(cmd *cobra.Command, v any) error {
	value := reflect.ValueOf(v)
	if value.IsValid() && value.Kind() == reflect.Slice && value.IsNil() {
		v = reflect.MakeSlice(value.Type(), 0, 0).Interface()
	}
	encoder := json.NewEncoder(cmd.OutOrStdout())
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}

// HasFilters reports whether any filter query is non-empty after parsing.
func HasFilters(rawFilters []string) bool {
	return len(ParseFilters(rawFilters)) > 0
}

// ResolveParameterArgument returns an optional literal positional parameter
// selector and rejects combining it with query filters.
func ResolveParameterArgument(args []string, rawFilters []string) (*string, error) {
	if len(args) == 0 {
		return nil, nil
	}
	if HasFilters(rawFilters) {
		return nil, InvalidArgument(fmt.Errorf("parameter argument cannot be used together with --filter"))
	}
	value := args[0]
	return &value, nil
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
