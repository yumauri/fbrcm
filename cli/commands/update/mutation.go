package updatecmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/cli/shared/rc"
	coreconditions "github.com/yumauri/fbrcm/core/conditions"
	"github.com/yumauri/fbrcm/core/firebase"
	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
)

func confirmAndUpdateProject(cmd *cobra.Command, label string, cfg *firebase.RemoteConfig, matched []shared.ParamTarget, spec updateSpec, yes bool, diffOut io.Writer) ([]shared.ParamTarget, *firebase.RemoteConfig, error) {
	return shared.ConfirmParamTargets(cmd, label, cfg, matched, yes, diffOut, func(target shared.ParamTarget, finalCfg *firebase.RemoteConfig) (shared.ParamTargetMutationStep, error) {
		nextCfg, err := firebase.CloneRemoteConfig(finalCfg)
		if err != nil {
			return shared.ParamTargetMutationStep{}, err
		}
		if err := updateParamSlot(nextCfg, target, spec); err != nil {
			return shared.ParamTargetMutationStep{}, err
		}
		diffText, hasChanges := rc.RenderRemoteConfigDiff(finalCfg, nextCfg)
		if !hasChanges {
			return shared.ParamTargetMutationStep{Skip: true}, nil
		}
		return shared.ParamTargetMutationStep{DiffText: diffText, Prompt: fmt.Sprintf("Update %s in %s?", rcdisplay.FormatParameterHeader(target.Key, target.Group), label), Apply: func(_ *firebase.RemoteConfig) (*firebase.RemoteConfig, error) { return nextCfg, nil }}, nil
	})
}

func updateParamSlot(cfg *firebase.RemoteConfig, target shared.ParamTarget, spec updateSpec) error {
	param := target.Param
	if spec.value != nil {
		if spec.condition == "" {
			param.DefaultValue = &firebase.RemoteConfigValue{Value: spec.value.value}
		} else {
			condition, ok := coreconditions.ResolveName(cfg, spec.condition)
			if !ok {
				return fmt.Errorf("condition %q not found", spec.condition)
			}
			if param.ConditionalValues == nil {
				param.ConditionalValues = make(map[string]firebase.RemoteConfigValue)
			}
			param.ConditionalValues[condition] = firebase.RemoteConfigValue{Value: spec.value.value}
		}
		param.ValueType = spec.value.valueType
	}
	if spec.descriptionChanged {
		param.Description = spec.description
	}
	if spec.removeAllConditionalValues {
		param.ConditionalValues = nil
	} else if len(spec.removeConditionalValues) > 0 {
		for _, name := range spec.removeConditionalValues {
			delete(param.ConditionalValues, name)
		}
		if len(param.ConditionalValues) == 0 {
			param.ConditionalValues = nil
		}
	}
	nextGroup, nextKey := target.Group, target.Key
	if spec.groupChanged {
		nextGroup = spec.group
	}
	if spec.nameChanged {
		nextKey = spec.name
	}
	if (target.Key != nextKey || target.Group != nextGroup) && shared.ParamSlotExists(cfg, nextKey, nextGroup) {
		return fmt.Errorf("parameter %s already exists", rcdisplay.FormatParameterHeader(nextKey, nextGroup))
	}
	shared.RemoveParamSlot(cfg, target.Key, target.Group)
	shared.SetParamSlot(cfg, nextKey, nextGroup, param)
	return nil
}
