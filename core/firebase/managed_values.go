package firebase

// RemoteConfigExperimentValue is the template payload for a value managed by
// Firebase A/B Testing.
type RemoteConfigExperimentValue struct {
	ExperimentID    string                               `json:"experimentId"`
	VariantValues   []RemoteConfigExperimentVariantValue `json:"variantValue"`
	ExposurePercent *float64                             `json:"exposurePercent"`
}

// RemoteConfigExperimentVariantValue is one experiment variant's template
// value. Firebase supplies either Value or NoChange.
type RemoteConfigExperimentVariantValue struct {
	VariantID string  `json:"variantId"`
	Value     *string `json:"value"`
	NoChange  *bool   `json:"noChange"`
}

// RemoteConfigRolloutValue is the template payload for a rollout-managed
// value.
type RemoteConfigRolloutValue struct {
	RolloutID string   `json:"rolloutId"`
	Value     *string  `json:"value"`
	Percent   *float64 `json:"percent"`
}

// RemoteConfigPersonalizationValue is the template payload for a personalized
// value.
type RemoteConfigPersonalizationValue struct {
	PersonalizationID string `json:"personalizationId"`
}
