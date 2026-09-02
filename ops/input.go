package ops

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/pflag"

	"github.com/yumauri/fbrcm/ops/contract"
)

// Input is the existing normalized machine contract. It is not converted to
// argv: names and values are bound directly to a fresh operation's options.
type Input struct {
	Arguments map[string]json.RawMessage `json:"arguments"`
	Options   map[string]json.RawMessage `json:"options"`
	Stdin     json.RawMessage            `json:"stdin"`
}

func BoundOption(name string) bool {
	return slices.Contains([]string{"profile", "stateless", "json", "timeout", "no-local-config", "help", "yes"}, strings.TrimPrefix(name, "--"))
}

func (in Input) Positionals(c contract.Capability) ([]string, error) {
	var result []string
	known := make(map[string]bool)
	missing := false
	for _, argument := range c.Arguments {
		known[argument.Name] = true
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
			result = append(result, values...)
		} else {
			var value string
			if err := json.Unmarshal(raw, &value); err != nil {
				return nil, err
			}
			result = append(result, value)
		}
	}
	for name := range in.Arguments {
		if !known[name] {
			return nil, fmt.Errorf("unknown argument %q", name)
		}
	}
	return result, nil
}

func (in Input) Validate(c contract.Capability) error {
	known := make(map[string]bool, len(c.Flags))
	for _, flag := range c.Flags {
		known[strings.TrimPrefix(flag.Name, "--")] = true
	}
	for name := range in.Options {
		if BoundOption(name) || !known[name] {
			return fmt.Errorf("option %q is not available to MCP callers", name)
		}
	}
	_, err := in.Positionals(c)
	return err
}

func bindOption(flags *pflag.FlagSet, name string, raw json.RawMessage) error {
	flag := flags.Lookup(name)
	if flag == nil {
		return fmt.Errorf("unknown option %q", name)
	}
	if values, ok := flag.Value.(pflag.SliceValue); ok {
		var input []string
		if err := json.Unmarshal(raw, &input); err != nil {
			return err
		}
		if err := values.Replace(input); err != nil {
			return err
		}
		flag.Changed = true
		return nil
	}
	var value any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return err
	}
	switch value.(type) {
	case string, bool, json.Number:
		return flags.Set(name, fmt.Sprint(value))
	default:
		return fmt.Errorf("option %q requires a scalar value", name)
	}
}
