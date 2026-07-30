package firebase

import (
	"bytes"
	"encoding/json"
	"fmt"
)

const (
	remoteConfigValueField                = "value"
	remoteConfigUseInAppDefaultField      = "useInAppDefault"
	remoteConfigPersonalizationValueField = "personalizationValue"
	remoteConfigExperimentValueField      = "experimentValue"
	remoteConfigRolloutValueField         = "rolloutValue"
)

// IsPlain reports whether the value is a normal string value rather than a
// Firebase-managed or in-app-default value.
func (v RemoteConfigValue) IsPlain() bool {
	return !v.UseInAppDefault &&
		len(v.PersonalizationValue) == 0 &&
		len(v.ExperimentValue) == 0 &&
		len(v.RolloutValue) == 0 &&
		v.UnknownValueOption == "" &&
		len(v.UnknownValue) == 0
}

// IsManaged reports whether Firebase manages the value through
// personalization, an experiment, or a rollout.
func (v RemoteConfigValue) IsManaged() bool {
	return len(v.PersonalizationValue) > 0 ||
		len(v.ExperimentValue) > 0 ||
		len(v.RolloutValue) > 0
}

// IsOpaque reports whether the value must be preserved without using a plain
// value editor. This includes known Firebase-managed values and unknown future
// union options.
func (v RemoteConfigValue) IsOpaque() bool {
	return v.IsManaged() || v.UnknownValueOption != "" || len(v.UnknownValue) > 0
}

// UnmarshalJSON decodes the documented RemoteConfigParameterValue union.
// Managed payloads remain opaque so nested fields survive read-modify-write
// cycles even when fbrcm does not interpret them.
func (v *RemoteConfigValue) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return fmt.Errorf("decode remote config parameter value: %w", err)
	}
	if len(fields) != 1 {
		return fmt.Errorf("decode remote config parameter value: expected exactly one value option, got %d", len(fields))
	}

	var decoded RemoteConfigValue
	for field, raw := range fields {
		switch field {
		case remoteConfigValueField:
			if err := json.Unmarshal(raw, &decoded.Value); err != nil {
				return fmt.Errorf("decode remote config parameter value %q: %w", field, err)
			}
		case remoteConfigUseInAppDefaultField:
			if err := json.Unmarshal(raw, &decoded.UseInAppDefault); err != nil {
				return fmt.Errorf("decode remote config parameter value %q: %w", field, err)
			}
			if !decoded.UseInAppDefault {
				return fmt.Errorf("decode remote config parameter value %q: must be true", field)
			}
		case remoteConfigPersonalizationValueField:
			value, err := decodeManagedRemoteConfigValue(field, raw)
			if err != nil {
				return err
			}
			decoded.PersonalizationValue = value
		case remoteConfigExperimentValueField:
			value, err := decodeManagedRemoteConfigValue(field, raw)
			if err != nil {
				return err
			}
			decoded.ExperimentValue = value
		case remoteConfigRolloutValueField:
			value, err := decodeManagedRemoteConfigValue(field, raw)
			if err != nil {
				return err
			}
			decoded.RolloutValue = value
		default:
			value, err := compactRemoteConfigValue(raw)
			if err != nil {
				return fmt.Errorf("decode remote config parameter value %q: %w", field, err)
			}
			decoded.UnknownValueOption = field
			decoded.UnknownValue = value
		}
	}

	*v = decoded
	return nil
}

// MarshalJSON encodes exactly one documented RemoteConfigParameterValue
// option. The zero Go value represents Firebase's valid empty string value.
func (v RemoteConfigValue) MarshalJSON() ([]byte, error) {
	managedCount := 0
	for _, raw := range []json.RawMessage{v.PersonalizationValue, v.ExperimentValue, v.RolloutValue} {
		if len(raw) > 0 {
			managedCount++
		}
	}

	optionCount := managedCount
	if v.UseInAppDefault {
		optionCount++
	}
	hasUnknownValue := v.UnknownValueOption != "" || len(v.UnknownValue) > 0
	if hasUnknownValue {
		optionCount++
	}
	if optionCount > 1 || (optionCount == 1 && v.Value != "") {
		return nil, fmt.Errorf("encode remote config parameter value: expected exactly one value option")
	}
	if hasUnknownValue && (v.UnknownValueOption == "" || len(v.UnknownValue) == 0) {
		return nil, fmt.Errorf("encode remote config parameter value: unknown value option requires both name and value")
	}
	if isKnownRemoteConfigValueField(v.UnknownValueOption) {
		return nil, fmt.Errorf("encode remote config parameter value: %q is a known value option", v.UnknownValueOption)
	}

	switch {
	case v.UseInAppDefault:
		return json.Marshal(struct {
			UseInAppDefault bool `json:"useInAppDefault"`
		}{UseInAppDefault: true})
	case len(v.PersonalizationValue) > 0:
		return marshalManagedRemoteConfigValue(remoteConfigPersonalizationValueField, v.PersonalizationValue)
	case len(v.ExperimentValue) > 0:
		return marshalManagedRemoteConfigValue(remoteConfigExperimentValueField, v.ExperimentValue)
	case len(v.RolloutValue) > 0:
		return marshalManagedRemoteConfigValue(remoteConfigRolloutValueField, v.RolloutValue)
	case hasUnknownValue:
		return marshalRemoteConfigRawValue(v.UnknownValueOption, v.UnknownValue)
	default:
		return json.Marshal(struct {
			Value string `json:"value"`
		}{Value: v.Value})
	}
}

func decodeManagedRemoteConfigValue(field string, raw json.RawMessage) (json.RawMessage, error) {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil {
		return nil, fmt.Errorf("decode remote config parameter value %q: %w", field, err)
	}
	if object == nil {
		return nil, fmt.Errorf("decode remote config parameter value %q: expected object", field)
	}
	return compactRemoteConfigValue(raw)
}

func compactRemoteConfigValue(raw json.RawMessage) (json.RawMessage, error) {
	var compact bytes.Buffer
	if err := json.Compact(&compact, raw); err != nil {
		return nil, err
	}
	return bytes.Clone(compact.Bytes()), nil
}

func marshalManagedRemoteConfigValue(field string, raw json.RawMessage) ([]byte, error) {
	if _, err := decodeManagedRemoteConfigValue(field, raw); err != nil {
		return nil, err
	}
	return marshalRemoteConfigRawValue(field, raw)
}

func marshalRemoteConfigRawValue(field string, raw json.RawMessage) ([]byte, error) {
	if _, err := compactRemoteConfigValue(raw); err != nil {
		return nil, fmt.Errorf("encode remote config parameter value %q: %w", field, err)
	}
	return json.Marshal(map[string]json.RawMessage{field: raw})
}

func isKnownRemoteConfigValueField(field string) bool {
	switch field {
	case remoteConfigValueField,
		remoteConfigUseInAppDefaultField,
		remoteConfigPersonalizationValueField,
		remoteConfigExperimentValueField,
		remoteConfigRolloutValueField:
		return true
	default:
		return false
	}
}
