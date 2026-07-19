package firebase

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
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

// NormalizeRemoteConfigForUpdate removes read-only metadata and
// validates condition fields against Firebase's v1 update schema.
func NormalizeRemoteConfigForUpdate(cfg *RemoteConfig) error {
	if cfg == nil {
		return nil
	}
	cfg.Version = RemoteConfigVersion{}
	for index := range cfg.Conditions {
		condition := &cfg.Conditions[index]
		color, err := NormalizeConditionTagColor(condition.TagColor)
		if err != nil {
			return fmt.Errorf("condition %q: %w", condition.Name, err)
		}
		condition.TagColor = color
	}
	return nil
}

// MarshalRemoteConfigForUpdate clones and encodes a Firebase-compatible v1
// update payload without mutating the caller's config.
func MarshalRemoteConfigForUpdate(cfg *RemoteConfig) ([]byte, error) {
	update, err := CloneRemoteConfig(cfg)
	if err != nil {
		return nil, err
	}
	if err := NormalizeRemoteConfigForUpdate(update); err != nil {
		return nil, err
	}
	return MarshalRemoteConfig(update)
}

// PrepareRemoteConfigUpdate parses arbitrary Remote Config JSON and returns a
// payload accepted by Firebase's v1 validate and update endpoints.
func PrepareRemoteConfigUpdate(raw json.RawMessage) ([]byte, error) {
	cfg, err := ParseCloneRemoteConfig(raw)
	if err != nil {
		return nil, err
	}
	return MarshalRemoteConfigForUpdate(cfg)
}
