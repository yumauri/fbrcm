package mcpserver

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/yumauri/fbrcm/cli/contract"
)

// Invocation uses the existing published normalized CLI input shape. Values
// remain JSON until validated against the command's semantic schema.
type Invocation struct {
	Arguments map[string]json.RawMessage `json:"arguments"`
	Options   map[string]json.RawMessage `json:"options"`
	Stdin     json.RawMessage            `json:"stdin"`
}

// Argv converts validated structured input without invoking a shell. Positional
// values follow --, so even a value resembling a flag cannot change policy.
func (in Invocation) Argv(c contract.Capability, o Options, confirmed bool) ([]string, error) {
	argv := append([]string(nil), c.Path...)
	argv = append(argv, "--json", fmt.Sprintf("--stateless=%t", o.Stateless), fmt.Sprintf("--no-local-config=%t", o.NoLocalConfig || o.Stateless))
	if !o.Stateless {
		argv = append(argv, "--profile="+o.Profile)
	}
	flags := make(map[string]contract.FlagCapability)
	for _, flag := range c.Flags {
		flags[strings.TrimPrefix(flag.Name, "--")] = flag
	}
	keys := make([]string, 0, len(in.Options))
	for key := range in.Options {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		flag, ok := flags[key]
		if !ok || boundOption(key) {
			return nil, fmt.Errorf("option %q is not available to MCP callers", key)
		}
		var value any
		decoder := json.NewDecoder(bytes.NewReader(in.Options[key]))
		decoder.UseNumber()
		if err := decoder.Decode(&value); err != nil {
			return nil, err
		}
		if values, ok := value.([]any); ok {
			parts := make([]string, len(values))
			for i, item := range values {
				parts[i] = fmt.Sprint(item)
			}
			if flag.Type == "stringSlice" {
				var buffer bytes.Buffer
				writer := csv.NewWriter(&buffer)
				if err := writer.Write(parts); err != nil {
					return nil, err
				}
				writer.Flush()
				argv = append(argv, "--"+key+"="+strings.TrimSuffix(buffer.String(), "\n"))
			} else {
				for _, part := range parts {
					argv = append(argv, "--"+key+"="+part)
				}
			}
		} else {
			argv = append(argv, "--"+key+"="+fmt.Sprint(value))
		}
	}
	if confirmed && c.Supports.ConfirmationBypass {
		argv = append(argv, "--yes")
	}
	argv = append(argv, "--")
	missing := false
	for _, argument := range c.Arguments {
		raw, ok := in.Arguments[argument.Name]
		if !ok {
			missing = true
			continue
		}
		if missing {
			return nil, fmt.Errorf("cannot supply %s after an omitted positional argument", argument.Name)
		}
		if argument.Repeated {
			var values []string
			if err := json.Unmarshal(raw, &values); err != nil {
				return nil, err
			}
			argv = append(argv, values...)
		} else {
			var value string
			if err := json.Unmarshal(raw, &value); err != nil {
				return nil, err
			}
			argv = append(argv, value)
		}
	}
	return argv, nil
}
