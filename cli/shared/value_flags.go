package shared

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// ValueFlag is the selected typed Remote Config value source.
type ValueFlag struct {
	Value           string
	Type            string
	UseInAppDefault bool
}

// ReadValueFlag reads --type together with either --value or
// --use-in-app-default.
func ReadValueFlag(cmd *cobra.Command, required bool) (*ValueFlag, error) {
	valueChanged := cmd.Flags().Changed("value")
	useInAppDefault, err := cmd.Flags().GetBool("use-in-app-default")
	if err != nil {
		return nil, err
	}
	useInAppDefault = cmd.Flags().Changed("use-in-app-default") && useInAppDefault
	if valueChanged && useInAppDefault {
		return nil, fmt.Errorf("--value and --use-in-app-default are mutually exclusive")
	}
	if !valueChanged && !useInAppDefault {
		if cmd.Flags().Changed("type") {
			return nil, fmt.Errorf("--type requires --value or --use-in-app-default")
		}
		if required {
			return nil, fmt.Errorf("one of --value or --use-in-app-default is required")
		}
		return nil, nil
	}
	if !cmd.Flags().Changed("type") {
		return nil, fmt.Errorf("--type is required with %s", map[bool]string{true: "--use-in-app-default", false: "--value"}[useInAppDefault])
	}
	typeName, err := cmd.Flags().GetString("type")
	if err != nil {
		return nil, err
	}
	valueType, err := ParseValueType(typeName)
	if err != nil {
		return nil, err
	}
	if useInAppDefault {
		return &ValueFlag{Type: valueType, UseInAppDefault: true}, nil
	}
	value, err := cmd.Flags().GetString("value")
	if err != nil {
		return nil, err
	}
	if err := validateTypedValue(valueType, value); err != nil {
		return nil, err
	}
	return &ValueFlag{Value: value, Type: valueType}, nil
}

func validateTypedValue(valueType, value string) error {
	switch valueType {
	case "BOOLEAN":
		if value != "true" && value != "false" {
			return fmt.Errorf("--value must be true or false for boolean type")
		}
	case "NUMBER":
		if _, err := strconv.ParseFloat(value, 64); err != nil {
			return fmt.Errorf("--value must be a valid number for number type")
		}
	case "JSON":
		if !json.Valid([]byte(value)) {
			return fmt.Errorf("--value must be valid json for json type")
		}
	}
	return nil
}

// ParseValueType parses a user-facing Remote Config parameter type.
func ParseValueType(value string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "string":
		return "STRING", nil
	case "boolean", "bool":
		return "BOOLEAN", nil
	case "number":
		return "NUMBER", nil
	case "json":
		return "JSON", nil
	default:
		return "", fmt.Errorf("--type must be one of string, boolean, number, or json")
	}
}
