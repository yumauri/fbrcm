package conditions

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/yumauri/fbrcm/core/firebase"
	rcmutate "github.com/yumauri/fbrcm/core/rc/mutate"
	rcvalue "github.com/yumauri/fbrcm/core/rc/value"
	"github.com/yumauri/fbrcm/core/rootgroup"
)

const MaxNameLength = 100

var DisplayColors = []string{
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

type Definition struct {
	Name       string
	Expression string
	TagColor   string
}

type Edit struct {
	Expression *string
	TagColor   *string
}

func NormalizeName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("condition name cannot be empty")
	}
	if utf8.RuneCountInString(name) > MaxNameLength {
		return "", fmt.Errorf("condition name cannot exceed %d characters", MaxNameLength)
	}
	return name, nil
}

func NormalizeExpression(expression string) (string, error) {
	expression = strings.TrimSpace(expression)
	if expression == "" {
		return "", fmt.Errorf("condition expression cannot be empty")
	}
	return expression, nil
}

func NormalizeTagColor(color string) (string, error) {
	color = strings.ToUpper(strings.TrimSpace(color))
	if color == "" || color == "CONDITION_DISPLAY_COLOR_UNSPECIFIED" {
		return "", nil
	}
	if slices.Contains(DisplayColors, color) {
		return color, nil
	}
	return "", fmt.Errorf("unsupported condition color %q (allowed: %s)", color, strings.Join(DisplayColors, ", "))
}

// ResolveName returns the canonical condition name using exact, then
// case-insensitive matching.
func ResolveName(cfg *firebase.RemoteConfig, requested string) (string, bool) {
	if cfg == nil {
		return "", false
	}
	for _, condition := range cfg.Conditions {
		if condition.Name == requested {
			return condition.Name, true
		}
	}
	for _, condition := range cfg.Conditions {
		if strings.EqualFold(condition.Name, requested) {
			return condition.Name, true
		}
	}
	return "", false
}

func Add(cfg *firebase.RemoteConfig, definition Definition, priority int) error {
	if cfg == nil {
		return fmt.Errorf("remote config is nil")
	}
	name, err := NormalizeName(definition.Name)
	if err != nil {
		return err
	}
	expression, err := NormalizeExpression(definition.Expression)
	if err != nil {
		return err
	}
	color, err := NormalizeTagColor(definition.TagColor)
	if err != nil {
		return err
	}
	if _, ok := conditionIndex(cfg, name); ok {
		return fmt.Errorf("condition %q already exists", name)
	}
	if priority == 0 {
		priority = len(cfg.Conditions) + 1
	}
	if priority < 1 || priority > len(cfg.Conditions)+1 {
		return fmt.Errorf("condition priority must be between 1 and %d", len(cfg.Conditions)+1)
	}
	condition := firebase.RemoteConfigCondition{Name: name, Expression: expression, TagColor: color}
	index := priority - 1
	cfg.Conditions = append(cfg.Conditions, firebase.RemoteConfigCondition{})
	copy(cfg.Conditions[index+1:], cfg.Conditions[index:])
	cfg.Conditions[index] = condition
	return nil
}

func EditDefinition(cfg *firebase.RemoteConfig, name string, edit Edit) error {
	index, ok := conditionIndex(cfg, name)
	if !ok {
		return fmt.Errorf("condition %q not found", name)
	}
	condition := cfg.Conditions[index]
	if edit.Expression != nil {
		expression, err := NormalizeExpression(*edit.Expression)
		if err != nil {
			return err
		}
		condition.Expression = expression
	}
	if edit.TagColor != nil {
		color, err := NormalizeTagColor(*edit.TagColor)
		if err != nil {
			return err
		}
		condition.TagColor = color
	}
	if condition == cfg.Conditions[index] {
		return fmt.Errorf("condition not changed")
	}
	cfg.Conditions[index] = condition
	return nil
}

// EditDetails updates all editable condition details in one atomic mutation.
func EditDetails(cfg *firebase.RemoteConfig, edit DetailsEdit) error {
	if cfg == nil {
		return fmt.Errorf("remote config is nil")
	}
	index, ok := conditionIndex(cfg, edit.Name)
	if !ok {
		return fmt.Errorf("condition %q not found", edit.Name)
	}
	nextName, err := NormalizeName(edit.NextName)
	if err != nil {
		return err
	}
	nextExpression, err := NormalizeExpression(edit.NextExpression)
	if err != nil {
		return err
	}
	nextColor, err := NormalizeTagColor(edit.NextTagColor)
	if err != nil {
		return err
	}
	if edit.NextPriority < 1 || edit.NextPriority > len(cfg.Conditions) {
		return fmt.Errorf("condition priority must be between 1 and %d", len(cfg.Conditions))
	}
	if nextName != edit.Name {
		if _, exists := conditionIndex(cfg, nextName); exists {
			return fmt.Errorf("condition %q already exists", nextName)
		}
	}

	condition := cfg.Conditions[index]
	definitionChanged := condition.Name != nextName || condition.Expression != nextExpression || condition.TagColor != nextColor || index+1 != edit.NextPriority
	valuesChanged, err := editUsageValues(cfg, edit.Name, edit.ValueEdits)
	if err != nil {
		return err
	}
	if !definitionChanged && !valuesChanged {
		return fmt.Errorf("condition not changed")
	}
	previousName := condition.Name
	condition.Name = nextName
	condition.Expression = nextExpression
	condition.TagColor = nextColor

	target := edit.NextPriority - 1
	if index < target {
		copy(cfg.Conditions[index:target], cfg.Conditions[index+1:target+1])
	} else if index > target {
		copy(cfg.Conditions[target+1:index+1], cfg.Conditions[target:index])
	}
	cfg.Conditions[target] = condition
	if previousName != nextName {
		renameConditionalValues(cfg.Parameters, previousName, nextName)
		for groupName, group := range cfg.ParameterGroups {
			renameConditionalValues(group.Parameters, previousName, nextName)
			cfg.ParameterGroups[groupName] = group
		}
	}
	return nil
}

func editUsageValues(cfg *firebase.RemoteConfig, conditionName string, edits []UsageValueEdit) (bool, error) {
	changed := false
	slots := rcmutate.CollectParamSlots(cfg)
	pending := make(map[string]rcmutate.Slot)
	for _, edit := range edits {
		groupKey := edit.GroupKey
		if groupKey == rootgroup.TreeKey {
			groupKey = rootgroup.WireKey
		}
		slotKey := rcmutate.SlotKey(groupKey, edit.ParameterKey)
		slot, ok := pending[slotKey]
		if !ok {
			slot, ok = slots[slotKey]
			if ok {
				conditionalValues := make(map[string]firebase.RemoteConfigValue, len(slot.Param.ConditionalValues))
				maps.Copy(conditionalValues, slot.Param.ConditionalValues)
				slot.Param.ConditionalValues = conditionalValues
			}
		}
		if !ok {
			return false, fmt.Errorf("parameter %q not found", edit.ParameterKey)
		}
		value, ok := slot.Param.ConditionalValues[conditionName]
		if !ok {
			return false, fmt.Errorf("conditional value %q not found on parameter %q", conditionName, edit.ParameterKey)
		}
		if value.UseInAppDefault || len(value.PersonalizationValue) > 0 || len(value.RolloutValue) > 0 {
			return false, fmt.Errorf("value editor supports only plain values")
		}
		if err := rcvalue.ValidateRawValueForType(edit.NextValue, slot.Param.ValueType); err != nil {
			return false, err
		}
		if value.Value == edit.NextValue {
			continue
		}
		value.Value = edit.NextValue
		slot.Param.ConditionalValues[conditionName] = value
		pending[slotKey] = slot
		changed = true
	}
	for slotKey, slot := range pending {
		rcmutate.SetParamSlot(cfg, rcmutate.SlotKeyParam(slotKey), slot)
	}
	return changed, nil
}

func Rename(cfg *firebase.RemoteConfig, name, nextName string) error {
	index, ok := conditionIndex(cfg, name)
	if !ok {
		return fmt.Errorf("condition %q not found", name)
	}
	nextName, err := NormalizeName(nextName)
	if err != nil {
		return err
	}
	if cfg.Conditions[index].Name == nextName {
		return fmt.Errorf("condition not changed")
	}
	if _, exists := conditionIndex(cfg, nextName); exists {
		return fmt.Errorf("condition %q already exists", nextName)
	}
	previousName := cfg.Conditions[index].Name
	cfg.Conditions[index].Name = nextName
	renameConditionalValues(cfg.Parameters, previousName, nextName)
	for groupName, group := range cfg.ParameterGroups {
		renameConditionalValues(group.Parameters, previousName, nextName)
		cfg.ParameterGroups[groupName] = group
	}
	return nil
}

func Move(cfg *firebase.RemoteConfig, name string, priority int) error {
	index, ok := conditionIndex(cfg, name)
	if !ok {
		return fmt.Errorf("condition %q not found", name)
	}
	if priority < 1 || priority > len(cfg.Conditions) {
		return fmt.Errorf("condition priority must be between 1 and %d", len(cfg.Conditions))
	}
	target := priority - 1
	if index == target {
		return fmt.Errorf("condition not changed")
	}
	condition := cfg.Conditions[index]
	if index < target {
		copy(cfg.Conditions[index:target], cfg.Conditions[index+1:target+1])
	} else {
		copy(cfg.Conditions[target+1:index+1], cfg.Conditions[target:index])
	}
	cfg.Conditions[target] = condition
	return nil
}

func Delete(cfg *firebase.RemoteConfig, name string) error {
	index, ok := conditionIndex(cfg, name)
	if !ok {
		return fmt.Errorf("condition %q not found", name)
	}
	cfg.Conditions = append(cfg.Conditions[:index], cfg.Conditions[index+1:]...)
	if len(cfg.Conditions) == 0 {
		cfg.Conditions = nil
	}
	rcmutate.DropUnknownConditionReferences(cfg)
	return nil
}

func conditionIndex(cfg *firebase.RemoteConfig, name string) (int, bool) {
	if cfg == nil {
		return 0, false
	}
	for index, condition := range cfg.Conditions {
		if condition.Name == name {
			return index, true
		}
	}
	return 0, false
}

func renameConditionalValues(params map[string]firebase.RemoteConfigParam, name, nextName string) {
	for parameterName, parameter := range params {
		value, ok := parameter.ConditionalValues[name]
		if !ok {
			continue
		}
		delete(parameter.ConditionalValues, name)
		parameter.ConditionalValues[nextName] = value
		params[parameterName] = parameter
	}
}
