// Package groups owns Remote Config parameter-group metadata mutations.
package groups

import (
	"fmt"
	"strings"

	"github.com/yumauri/fbrcm/core/firebase"
)

type Definition struct {
	Name        string
	Description string
}

type Edit struct {
	Description *string
}

type DetailsEdit struct {
	Name            string
	NextName        string
	NextDescription string
}

func NormalizeName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("group name cannot be empty")
	}
	return name, nil
}

func ResolveName(cfg *firebase.RemoteConfig, requested string) (string, bool) {
	if cfg == nil {
		return "", false
	}
	requested = strings.TrimSpace(requested)
	if _, ok := cfg.ParameterGroups[requested]; ok {
		return requested, true
	}
	for name := range cfg.ParameterGroups {
		if strings.EqualFold(name, requested) {
			return name, true
		}
	}
	return "", false
}

// Add creates a group even when it has no description or parameters.
func Add(cfg *firebase.RemoteConfig, definition Definition) error {
	if cfg == nil {
		return fmt.Errorf("remote config is nil")
	}
	name, err := NormalizeName(definition.Name)
	if err != nil {
		return err
	}
	if _, exists := cfg.ParameterGroups[name]; exists {
		return fmt.Errorf("group %q already exists", name)
	}
	if cfg.ParameterGroups == nil {
		cfg.ParameterGroups = make(map[string]firebase.RemoteConfigGroup)
	}
	cfg.ParameterGroups[name] = firebase.RemoteConfigGroup{Description: strings.TrimSpace(definition.Description)}
	return nil
}

func EditMetadata(cfg *firebase.RemoteConfig, name string, edit Edit) error {
	if cfg == nil {
		return fmt.Errorf("remote config is nil")
	}
	group, ok := cfg.ParameterGroups[name]
	if !ok {
		return fmt.Errorf("group %q not found", name)
	}
	if edit.Description == nil {
		return fmt.Errorf("group not changed")
	}
	next := strings.TrimSpace(*edit.Description)
	if group.Description == next {
		return fmt.Errorf("group not changed")
	}
	group.Description = next
	cfg.ParameterGroups[name] = group
	return nil
}

func Rename(cfg *firebase.RemoteConfig, name, nextName string) error {
	if cfg == nil {
		return fmt.Errorf("remote config is nil")
	}
	nextName, err := NormalizeName(nextName)
	if err != nil {
		return err
	}
	if name == nextName {
		return fmt.Errorf("group not changed")
	}
	group, ok := cfg.ParameterGroups[name]
	if !ok {
		return fmt.Errorf("group %q not found", name)
	}
	if _, exists := cfg.ParameterGroups[nextName]; exists {
		return fmt.Errorf("group %q already exists", nextName)
	}
	delete(cfg.ParameterGroups, name)
	cfg.ParameterGroups[nextName] = group
	return nil
}

func EditDetails(cfg *firebase.RemoteConfig, edit DetailsEdit) error {
	if cfg == nil {
		return fmt.Errorf("remote config is nil")
	}
	group, ok := cfg.ParameterGroups[edit.Name]
	if !ok {
		return fmt.Errorf("group %q not found", edit.Name)
	}
	nextName, err := NormalizeName(edit.NextName)
	if err != nil {
		return err
	}
	nextDescription := strings.TrimSpace(edit.NextDescription)
	if edit.Name == nextName && group.Description == nextDescription {
		return fmt.Errorf("group not changed")
	}
	if edit.Name != nextName {
		if _, exists := cfg.ParameterGroups[nextName]; exists {
			return fmt.Errorf("group %q already exists", nextName)
		}
		delete(cfg.ParameterGroups, edit.Name)
	}
	group.Description = nextDescription
	cfg.ParameterGroups[nextName] = group
	return nil
}

func Delete(cfg *firebase.RemoteConfig, name string) error {
	if cfg == nil {
		return fmt.Errorf("remote config is nil")
	}
	if _, ok := cfg.ParameterGroups[name]; !ok {
		return fmt.Errorf("group %q not found", name)
	}
	delete(cfg.ParameterGroups, name)
	return nil
}
