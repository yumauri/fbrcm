package shared

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

// ValueFlag is the selected Remote Config value flag.
type ValueFlag struct {
	Value string
	Type  string
}

// ReadValueFlag reads one of --boolean, --number, --string, or --json.
func ReadValueFlag(cmd *cobra.Command, required bool) (*ValueFlag, error) {
	specs := []struct {
		name      string
		valueType string
		validate  func(string) error
	}{
		{name: "boolean", valueType: "BOOLEAN", validate: func(value string) error {
			switch value {
			case "true", "false":
				return nil
			default:
				return fmt.Errorf("--boolean must be true or false")
			}
		}},
		{name: "number", valueType: "NUMBER", validate: func(value string) error {
			if _, err := strconv.ParseFloat(value, 64); err != nil {
				return fmt.Errorf("--number must be valid number")
			}
			return nil
		}},
		{name: "string", valueType: "STRING", validate: func(string) error { return nil }},
		{name: "json", valueType: "JSON", validate: func(value string) error {
			if !json.Valid([]byte(value)) {
				return fmt.Errorf("--json must be valid json")
			}
			return nil
		}},
	}

	selected := make([]ValueFlag, 0, 1)
	for _, spec := range specs {
		value, err := cmd.Flags().GetString(spec.name)
		if err != nil {
			return nil, err
		}
		if !cmd.Flags().Changed(spec.name) {
			continue
		}
		if err := spec.validate(value); err != nil {
			return nil, err
		}
		selected = append(selected, ValueFlag{Value: value, Type: spec.valueType})
	}

	if len(selected) == 0 {
		if required {
			return nil, fmt.Errorf("exactly one of --boolean, --number, --string, or --json is required")
		}
		return nil, nil
	}
	if len(selected) > 1 {
		return nil, fmt.Errorf("only one of --boolean, --number, --string, or --json may be used")
	}
	return &selected[0], nil
}
