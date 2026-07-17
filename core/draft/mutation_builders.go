package draft

import (
	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/core/groups"
	rcmutate "github.com/yumauri/fbrcm/core/rc/mutate"
)

func DeleteParameter(groupKey, paramKey string) Mutation {
	group := NormalizeGroupKey(groupKey)
	return func(cfg *firebase.RemoteConfig) error {
		rcmutate.RemoveParamSlot(cfg, paramKey, group)
		return nil
	}
}

func DeleteGroup(groupKey string) Mutation {
	group := NormalizeGroupKey(groupKey)
	return func(cfg *firebase.RemoteConfig) error {
		return groups.Delete(cfg, group)
	}
}

func DeleteConditionalValue(groupKey, paramKey, valueLabel string) Mutation {
	group := NormalizeGroupKey(groupKey)
	return func(cfg *firebase.RemoteConfig) error {
		return deleteConditionalValueSlot(cfg, paramKey, group, valueLabel)
	}
}

func RenameParameter(groupKey, paramKey, nextParamKey string) Mutation {
	group := NormalizeGroupKey(groupKey)
	return func(cfg *firebase.RemoteConfig) error {
		return renameParamSlot(cfg, paramKey, nextParamKey, group)
	}
}

func RenameGroup(groupKey, nextGroupKey string) Mutation {
	return func(cfg *firebase.RemoteConfig) error {
		return groups.Rename(cfg, NormalizeGroupKey(groupKey), NormalizeGroupKey(nextGroupKey))
	}
}

func EditGroupDetails(edit groups.DetailsEdit) Mutation {
	return func(cfg *firebase.RemoteConfig) error { return groups.EditDetails(cfg, edit) }
}

func MoveParameter(groupKey, paramKey, nextGroupKey string) Mutation {
	return func(cfg *firebase.RemoteConfig) error {
		return moveParamSlot(cfg, paramKey, NormalizeGroupKey(groupKey), NormalizeGroupKey(nextGroupKey))
	}
}

func EditParameterDetails(edit ParameterDetailsEdit) Mutation {
	return func(cfg *firebase.RemoteConfig) error {
		return applyParameterDetailsEdit(cfg, edit)
	}
}

func MoveGroup(groupKey, nextGroupKey string) Mutation {
	return func(cfg *firebase.RemoteConfig) error {
		return moveGroupSlot(cfg, NormalizeGroupKey(groupKey), NormalizeGroupKey(nextGroupKey))
	}
}

func SetBooleanParameterValue(groupKey, paramKey, valueLabel string, nextValue bool) Mutation {
	group := NormalizeGroupKey(groupKey)
	return func(cfg *firebase.RemoteConfig) error {
		return setBooleanParamValueSlot(cfg, paramKey, group, valueLabel, nextValue)
	}
}

func SetNumberParameterValue(groupKey, paramKey, valueLabel, nextValue string) Mutation {
	group := NormalizeGroupKey(groupKey)
	return func(cfg *firebase.RemoteConfig) error {
		return setNumberParamValueSlot(cfg, paramKey, group, valueLabel, nextValue)
	}
}

func SetStringParameterValue(groupKey, paramKey, valueLabel, nextValue string) Mutation {
	group := NormalizeGroupKey(groupKey)
	return func(cfg *firebase.RemoteConfig) error {
		return setStringParamValueSlot(cfg, paramKey, group, valueLabel, nextValue)
	}
}

func SetJSONParameterValue(groupKey, paramKey, valueLabel, nextValue string) Mutation {
	group := NormalizeGroupKey(groupKey)
	return func(cfg *firebase.RemoteConfig) error {
		return setJSONParamValueSlot(cfg, paramKey, group, valueLabel, nextValue)
	}
}

func DuplicateParameterNamed(groupKey, paramKey, nextParamKey string) Mutation {
	group := NormalizeGroupKey(groupKey)
	return func(cfg *firebase.RemoteConfig) error {
		return duplicateParamSlotAs(cfg, paramKey, nextParamKey, group)
	}
}

// DuplicateParameterAutoNamed applies auto-naming duplication and returns the generated name.
func DuplicateParameterAutoNamed(groupKey, paramKey string) (Mutation, func() string) {
	group := NormalizeGroupKey(groupKey)
	var nextParamKey string
	return func(cfg *firebase.RemoteConfig) error {
		var err error
		nextParamKey, err = duplicateParamSlot(cfg, paramKey, group)
		return err
	}, func() string { return nextParamKey }
}
