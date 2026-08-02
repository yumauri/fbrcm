package core

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/core/parameters"
	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
	rcmutate "github.com/yumauri/fbrcm/core/rc/mutate"
	"github.com/yumauri/fbrcm/core/rootgroup"
	"github.com/yumauri/fbrcm/core/strfold"
)

type ManagedFeatureTemplate struct {
	Version string `json:"version"`
	Source  string `json:"source"`
}

type ManagedValueReference struct {
	Group          string                `json:"group"`
	Parameter      string                `json:"parameter"`
	Condition      string                `json:"condition,omitempty"`
	ConditionColor string                `json:"-"`
	Default        bool                  `json:"default,omitempty"`
	ValueType      string                `json:"value_type"`
	Value          *string               `json:"value,omitempty"`
	Percentage     *float64              `json:"percentage,omitempty"`
	Variants       []ManagedVariantValue `json:"variants,omitempty"`
}

type ManagedVariantValue struct {
	ID       string  `json:"id"`
	Value    *string `json:"value,omitempty"`
	NoChange *bool   `json:"no_change,omitempty"`
}

type ExperimentEntry struct {
	firebase.Experiment
	References []ManagedValueReference `json:"references"`
}

func (e ExperimentEntry) MarshalJSON() ([]byte, error) {
	return marshalManagedFeatureEntry(e.Experiment, e.References)
}

type ExperimentList struct {
	Template    ManagedFeatureTemplate `json:"template"`
	Experiments []ExperimentEntry      `json:"experiments"`
}

type RolloutEntry struct {
	firebase.Rollout
	References []ManagedValueReference `json:"references"`
}

func (r RolloutEntry) MarshalJSON() ([]byte, error) {
	return marshalManagedFeatureEntry(r.Rollout, r.References)
}

type RolloutList struct {
	Template ManagedFeatureTemplate `json:"template"`
	Rollouts []RolloutEntry         `json:"rollouts"`
}

type PersonalizationEntry struct {
	ID         string                  `json:"id"`
	References []ManagedValueReference `json:"references"`
}

type PersonalizationList struct {
	Template         ManagedFeatureTemplate `json:"template"`
	Personalizations []PersonalizationEntry `json:"personalizations"`
}

func marshalManagedFeatureEntry(metadata any, references []ManagedValueReference) ([]byte, error) {
	raw, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return nil, err
	}
	referencesJSON, err := json.Marshal(references)
	if err != nil {
		return nil, err
	}
	fields["references"] = referencesJSON
	return json.Marshal(fields)
}

func (s *Core) ListRemoteConfigExperiments(ctx context.Context, project Project, update bool) (ExperimentList, error) {
	fb, projectIdentifier, err := s.managedFeatureService(ctx, project)
	if err != nil {
		return ExperimentList{}, err
	}
	cfg, template, err := s.loadManagedFeatureTemplate(ctx, project.ProjectID, update)
	if err != nil {
		return ExperimentList{}, err
	}
	references, err := collectExperimentReferences(cfg)
	if err != nil {
		return ExperimentList{}, err
	}
	result := ExperimentList{Template: template}
	token := ""
	for {
		page, err := fb.ListExperiments(ctx, projectIdentifier, project.ProjectID, firebase.ListManagedFeaturesOptions{PageSize: 100, PageToken: token})
		if err != nil {
			return ExperimentList{}, fmt.Errorf("firebase error: %w", err)
		}
		for _, experiment := range page.Experiments {
			id := firebase.ManagedFeatureID(experiment.Name)
			result.Experiments = append(result.Experiments, ExperimentEntry{
				Experiment: experiment, References: references[id],
			})
		}
		if page.NextPageToken == "" {
			break
		}
		token = page.NextPageToken
	}
	return result, nil
}

func (s *Core) GetRemoteConfigExperiment(ctx context.Context, project Project, experimentID string, update bool) (ExperimentEntry, ManagedFeatureTemplate, error) {
	experiment, err := s.GetRemoteConfigExperimentMetadata(ctx, project, experimentID)
	if err != nil {
		return ExperimentEntry{}, ManagedFeatureTemplate{}, err
	}
	cfg, template, err := s.loadManagedFeatureTemplate(ctx, project.ProjectID, update)
	if err != nil {
		return ExperimentEntry{}, ManagedFeatureTemplate{}, err
	}
	references, err := collectExperimentReferences(cfg)
	if err != nil {
		return ExperimentEntry{}, ManagedFeatureTemplate{}, err
	}
	id := firebase.ManagedFeatureID(experiment.Name)
	return ExperimentEntry{Experiment: experiment, References: references[id]}, template, nil
}

func (s *Core) GetRemoteConfigExperimentMetadata(ctx context.Context, project Project, experimentID string) (firebase.Experiment, error) {
	fb, projectIdentifier, err := s.managedFeatureService(ctx, project)
	if err != nil {
		return firebase.Experiment{}, err
	}
	experiment, err := fb.GetExperiment(ctx, projectIdentifier, project.ProjectID, experimentID)
	if err != nil {
		return firebase.Experiment{}, fmt.Errorf("firebase error: %w", err)
	}
	return experiment, nil
}

func (s *Core) DeleteRemoteConfigExperiment(ctx context.Context, project Project, experimentID string) error {
	fb, projectIdentifier, err := s.managedFeatureService(ctx, project)
	if err != nil {
		return err
	}
	if err := fb.DeleteExperiment(ctx, projectIdentifier, project.ProjectID, experimentID); err != nil {
		return fmt.Errorf("firebase error: %w", err)
	}
	return nil
}

func (s *Core) ListRemoteConfigRollouts(ctx context.Context, project Project, update bool) (RolloutList, error) {
	fb, projectIdentifier, err := s.managedFeatureService(ctx, project)
	if err != nil {
		return RolloutList{}, err
	}
	cfg, template, err := s.loadManagedFeatureTemplate(ctx, project.ProjectID, update)
	if err != nil {
		return RolloutList{}, err
	}
	references, err := collectRolloutReferences(cfg)
	if err != nil {
		return RolloutList{}, err
	}
	result := RolloutList{Template: template}
	token := ""
	for {
		page, err := fb.ListRollouts(ctx, projectIdentifier, project.ProjectID, firebase.ListManagedFeaturesOptions{PageSize: 100, PageToken: token})
		if err != nil {
			return RolloutList{}, fmt.Errorf("firebase error: %w", err)
		}
		for _, rollout := range page.Rollouts {
			id := firebase.ManagedFeatureID(rollout.Name)
			result.Rollouts = append(result.Rollouts, RolloutEntry{Rollout: rollout, References: references[id]})
		}
		if page.NextPageToken == "" {
			break
		}
		token = page.NextPageToken
	}
	return result, nil
}

func (s *Core) GetRemoteConfigRollout(ctx context.Context, project Project, rolloutID string, update bool) (RolloutEntry, ManagedFeatureTemplate, error) {
	rollout, err := s.GetRemoteConfigRolloutMetadata(ctx, project, rolloutID)
	if err != nil {
		return RolloutEntry{}, ManagedFeatureTemplate{}, err
	}
	cfg, template, err := s.loadManagedFeatureTemplate(ctx, project.ProjectID, update)
	if err != nil {
		return RolloutEntry{}, ManagedFeatureTemplate{}, err
	}
	references, err := collectRolloutReferences(cfg)
	if err != nil {
		return RolloutEntry{}, ManagedFeatureTemplate{}, err
	}
	id := firebase.ManagedFeatureID(rollout.Name)
	return RolloutEntry{Rollout: rollout, References: references[id]}, template, nil
}

func (s *Core) GetRemoteConfigRolloutMetadata(ctx context.Context, project Project, rolloutID string) (firebase.Rollout, error) {
	fb, projectIdentifier, err := s.managedFeatureService(ctx, project)
	if err != nil {
		return firebase.Rollout{}, err
	}
	rollout, err := fb.GetRollout(ctx, projectIdentifier, project.ProjectID, rolloutID)
	if err != nil {
		return firebase.Rollout{}, fmt.Errorf("firebase error: %w", err)
	}
	return rollout, nil
}

func (s *Core) DeleteRemoteConfigRollout(ctx context.Context, project Project, rolloutID string) error {
	fb, projectIdentifier, err := s.managedFeatureService(ctx, project)
	if err != nil {
		return err
	}
	if err := fb.DeleteRollout(ctx, projectIdentifier, project.ProjectID, rolloutID); err != nil {
		return fmt.Errorf("firebase error: %w", err)
	}
	return nil
}

func (s *Core) ListRemoteConfigPersonalizations(ctx context.Context, project Project, update bool) (PersonalizationList, error) {
	cfg, template, err := s.loadManagedFeatureTemplate(ctx, project.ProjectID, update)
	if err != nil {
		return PersonalizationList{}, err
	}
	references, err := collectPersonalizationReferences(cfg)
	if err != nil {
		return PersonalizationList{}, err
	}
	ids := strfold.SortedKeys(references)
	result := PersonalizationList{
		Template:         template,
		Personalizations: make([]PersonalizationEntry, 0, len(ids)),
	}
	for _, id := range ids {
		result.Personalizations = append(result.Personalizations, PersonalizationEntry{ID: id, References: references[id]})
	}
	return result, nil
}

func (s *Core) GetRemoteConfigPersonalization(ctx context.Context, project Project, personalizationID string, update bool) (PersonalizationEntry, ManagedFeatureTemplate, error) {
	result, err := s.ListRemoteConfigPersonalizations(ctx, project, update)
	if err != nil {
		return PersonalizationEntry{}, ManagedFeatureTemplate{}, err
	}
	for _, personalization := range result.Personalizations {
		if strings.EqualFold(personalization.ID, personalizationID) {
			return personalization, result.Template, nil
		}
	}
	return PersonalizationEntry{}, result.Template, fmt.Errorf("personalization %q not found in project %s", personalizationID, project.ProjectID)
}

func (s *Core) managedFeatureService(ctx context.Context, project Project) (*firebase.Service, string, error) {
	projectIdentifier := strings.TrimSpace(project.ProjectNumber)
	if projectIdentifier == "" {
		projectIdentifier = strings.TrimSpace(project.ProjectID)
	}
	if projectIdentifier == "" {
		return nil, "", fmt.Errorf("project has no Firebase project identifier")
	}
	fb, err := s.firebaseServiceForProject(ctx, project.ProjectID)
	if err != nil {
		return nil, "", err
	}
	return fb, projectIdentifier, nil
}

func (s *Core) loadManagedFeatureTemplate(ctx context.Context, projectID string, update bool) (*firebase.RemoteConfig, ManagedFeatureTemplate, error) {
	var cache *ParametersCache
	var source string
	var err error
	if update {
		cache, source, err = s.RevalidateParameters(ctx, projectID)
	} else {
		cache, source, err = s.GetParameters(ctx, projectID, false)
	}
	if err != nil {
		return nil, ManagedFeatureTemplate{}, err
	}
	cfg, err := firebase.ParseRemoteConfig(cache.RemoteConfig)
	if err != nil {
		return nil, ManagedFeatureTemplate{}, err
	}
	return cfg, ManagedFeatureTemplate{Version: cfg.Version.VersionNumber, Source: source}, nil
}

type decodedManagedValue struct {
	ID         string
	Value      *string
	Percentage *float64
	Variants   []ManagedVariantValue
}

func collectExperimentReferences(cfg *firebase.RemoteConfig) (map[string][]ManagedValueReference, error) {
	return collectManagedValueReferences(cfg, func(raw json.RawMessage) (decodedManagedValue, error) {
		if len(raw) == 0 {
			return decodedManagedValue{}, nil
		}
		var value firebase.RemoteConfigExperimentValue
		if err := json.Unmarshal(raw, &value); err != nil {
			return decodedManagedValue{}, fmt.Errorf("decode experiment value: %w", err)
		}
		decoded := decodedManagedValue{
			ID: strings.TrimSpace(value.ExperimentID), Percentage: value.ExposurePercent,
			Variants: make([]ManagedVariantValue, 0, len(value.VariantValues)),
		}
		for _, variant := range value.VariantValues {
			decoded.Variants = append(decoded.Variants, ManagedVariantValue{
				ID: variant.VariantID, Value: variant.Value, NoChange: variant.NoChange,
			})
		}
		return decoded, nil
	}, func(value firebase.RemoteConfigValue) json.RawMessage {
		return value.ExperimentValue
	})
}

func collectRolloutReferences(cfg *firebase.RemoteConfig) (map[string][]ManagedValueReference, error) {
	return collectManagedValueReferences(cfg, func(raw json.RawMessage) (decodedManagedValue, error) {
		if len(raw) == 0 {
			return decodedManagedValue{}, nil
		}
		var value firebase.RemoteConfigRolloutValue
		if err := json.Unmarshal(raw, &value); err != nil {
			return decodedManagedValue{}, fmt.Errorf("decode rollout value: %w", err)
		}
		return decodedManagedValue{
			ID: strings.TrimSpace(value.RolloutID), Value: value.Value, Percentage: value.Percent,
		}, nil
	}, func(value firebase.RemoteConfigValue) json.RawMessage {
		return value.RolloutValue
	})
}

func collectPersonalizationReferences(cfg *firebase.RemoteConfig) (map[string][]ManagedValueReference, error) {
	return collectManagedValueReferences(cfg, func(raw json.RawMessage) (decodedManagedValue, error) {
		if len(raw) == 0 {
			return decodedManagedValue{}, nil
		}
		var value firebase.RemoteConfigPersonalizationValue
		if err := json.Unmarshal(raw, &value); err != nil {
			return decodedManagedValue{}, fmt.Errorf("decode personalization value: %w", err)
		}
		return decodedManagedValue{ID: strings.TrimSpace(value.PersonalizationID)}, nil
	}, func(value firebase.RemoteConfigValue) json.RawMessage {
		return value.PersonalizationValue
	})
}

func collectManagedValueReferences(
	cfg *firebase.RemoteConfig,
	decode func(json.RawMessage) (decodedManagedValue, error),
	rawValue func(firebase.RemoteConfigValue) json.RawMessage,
) (map[string][]ManagedValueReference, error) {
	result := make(map[string][]ManagedValueReference)
	conditionColors := firebase.ConditionTagColorsByName(cfg.Conditions)
	slots := rcmutate.CollectParamSlots(cfg)
	for _, slotKey := range strfold.SortedKeys(slots) {
		slot := slots[slotKey]
		group := slot.Group
		if group == "" {
			group = rootgroup.Label
		}
		parameter := rcmutate.SlotKeyParam(slotKey)
		if slot.Param.DefaultValue != nil {
			decoded, err := decode(rawValue(*slot.Param.DefaultValue))
			if err != nil {
				return nil, fmt.Errorf("%s default value: %w", rcmutate.SlotDisplayKey(slotKey), err)
			}
			if decoded.ID != "" {
				result[decoded.ID] = append(result[decoded.ID], ManagedValueReference{
					Group: group, Parameter: parameter, Default: true,
					ValueType: strings.ToUpper(rcdisplay.EmptyValueType(slot.Param.ValueType)),
					Value:     decoded.Value, Percentage: decoded.Percentage, Variants: decoded.Variants,
				})
			}
		}
		for _, condition := range parameters.OrderedConditionalKeys(slot.Param.ConditionalValues, cfg.Conditions) {
			remoteValue := slot.Param.ConditionalValues[condition]
			decoded, err := decode(rawValue(remoteValue))
			if err != nil {
				return nil, fmt.Errorf("%s condition %s: %w", rcmutate.SlotDisplayKey(slotKey), condition, err)
			}
			if decoded.ID != "" {
				result[decoded.ID] = append(result[decoded.ID], ManagedValueReference{
					Group: group, Parameter: parameter, Condition: condition, ConditionColor: conditionColors[condition],
					ValueType: strings.ToUpper(rcdisplay.EmptyValueType(slot.Param.ValueType)),
					Value:     decoded.Value, Percentage: decoded.Percentage, Variants: decoded.Variants,
				})
			}
		}
	}
	for id := range result {
		slices.SortFunc(result[id], compareManagedValueReferences)
	}
	return result, nil
}

func compareManagedValueReferences(left, right ManagedValueReference) int {
	if cmp := strfold.Compare(left.Group, right.Group); cmp != 0 {
		return cmp
	}
	if cmp := strfold.Compare(left.Parameter, right.Parameter); cmp != 0 {
		return cmp
	}
	if left.Default != right.Default {
		if left.Default {
			return 1
		}
		return -1
	}
	return strfold.Compare(left.Condition, right.Condition)
}
