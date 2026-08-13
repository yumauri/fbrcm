package firebase

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"
)

// ConditionDisplayColors contains the values accepted by Firebase's v1
// RemoteConfigCondition.tagColor field.
var ConditionDisplayColors = []string{
	"BLUE",
	"BROWN",
	"CYAN",
	"DEEP_ORANGE",
	"GREEN",
	"INDIGO",
	"LIME",
	"ORANGE",
	"PINK",
	"PURPLE",
	"TEAL",
}

// NormalizeConditionTagColor returns a Firebase v1 condition display color.
func NormalizeConditionTagColor(color string) (string, error) {
	color = strings.ToUpper(strings.TrimSpace(color))
	if color == "" || color == "CONDITION_DISPLAY_COLOR_UNSPECIFIED" {
		return "", nil
	}
	if slices.Contains(ConditionDisplayColors, color) {
		return color, nil
	}
	return "", fmt.Errorf("unsupported condition color %q (allowed: %s)", color, strings.Join(ConditionDisplayColors, ", "))
}

// NormalizeRemoteConfigForUpdate removes read-only metadata, preserves an
// explicitly supplied change note, and validates condition fields against
// Firebase's v1 update schema.
func NormalizeRemoteConfigForUpdate(cfg *RemoteConfig, changeNote ...string) error {
	if cfg == nil {
		return nil
	}
	cfg.Version = RemoteConfigVersion{}
	if len(changeNote) > 1 {
		return fmt.Errorf("several change notes were supplied")
	}
	if len(changeNote) == 1 {
		note, err := NormalizeChangeNote(changeNote[0])
		if err != nil {
			return err
		}
		cfg.Version.ChangeNote = note
	}
	conditionNames := make(map[string]bool, len(cfg.Conditions))
	for index := range cfg.Conditions {
		condition := &cfg.Conditions[index]
		if strings.TrimSpace(condition.Name) == "" {
			return fmt.Errorf("condition %d has an empty name", index+1)
		}
		if strings.TrimSpace(condition.Expression) == "" {
			return fmt.Errorf("condition %q has an empty expression", condition.Name)
		}
		if conditionNames[condition.Name] {
			return fmt.Errorf("condition name %q is duplicated", condition.Name)
		}
		conditionNames[condition.Name] = true
		color, err := NormalizeConditionTagColor(condition.TagColor)
		if err != nil {
			return fmt.Errorf("condition %q: %w", condition.Name, err)
		}
		condition.TagColor = color
	}
	parameterNames := make(map[string]bool, len(cfg.Parameters))
	if err := validateRemoteConfigParameters(cfg.Parameters, parameterNames); err != nil {
		return err
	}
	for groupName, group := range cfg.ParameterGroups {
		if err := validateRemoteConfigName("parameter group", groupName); err != nil {
			return err
		}
		if utf8.RuneCountInString(group.Description) > 256 {
			return fmt.Errorf("parameter group %q description exceeds 256 characters", groupName)
		}
		if err := validateRemoteConfigParameters(group.Parameters, parameterNames); err != nil {
			return fmt.Errorf("parameter group %q: %w", groupName, err)
		}
	}
	return nil
}

func validateRemoteConfigParameters(parameters map[string]RemoteConfigParam, seen map[string]bool) error {
	for name, parameter := range parameters {
		if err := validateRemoteConfigName("parameter", name); err != nil {
			return err
		}
		if seen[name] {
			return fmt.Errorf("parameter %q appears more than once in the template", name)
		}
		seen[name] = true
		if utf8.RuneCountInString(parameter.Description) > 256 {
			return fmt.Errorf("parameter %q description exceeds 256 characters", name)
		}
		if parameter.ValueType != "" && !slices.Contains([]string{"PARAMETER_VALUE_TYPE_UNSPECIFIED", "STRING", "BOOLEAN", "NUMBER", "JSON"}, parameter.ValueType) {
			return fmt.Errorf("parameter %q has unsupported valueType %q", name, parameter.ValueType)
		}
	}
	return nil
}

func validateRemoteConfigName(kind, name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("%s name is empty", kind)
	}
	if utf8.RuneCountInString(name) > 256 {
		return fmt.Errorf("%s name %q exceeds 256 characters", kind, name)
	}
	return nil
}

// MarshalRemoteConfigForUpdate clones and encodes a Firebase-compatible v1
// update payload without mutating the caller's config.
func MarshalRemoteConfigForUpdate(cfg *RemoteConfig, changeNote ...string) ([]byte, error) {
	update, err := CloneRemoteConfig(cfg)
	if err != nil {
		return nil, err
	}
	if err := NormalizeRemoteConfigForUpdate(update, changeNote...); err != nil {
		return nil, err
	}
	return MarshalRemoteConfig(update)
}

// PrepareRemoteConfigUpdate parses arbitrary Remote Config JSON and returns a
// payload accepted by Firebase's v1 validate and update endpoints.
func PrepareRemoteConfigUpdate(raw json.RawMessage, changeNote ...string) ([]byte, error) {
	cfg, err := ParseCloneRemoteConfig(raw)
	if err != nil {
		return nil, err
	}
	return MarshalRemoteConfigForUpdate(cfg, changeNote...)
}
