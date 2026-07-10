package draft

import "github.com/yumauri/fbrcm/core/firebase"

// Mutation applies one in-memory change to a cloned Remote Config document.
type Mutation func(*firebase.RemoteConfig) error

// MutationSpec describes one draft mutation for mutate and preview paths.
type MutationSpec struct {
	UnchangedErr string
	Apply        Mutation
}

type ParameterDetailsEdit struct {
	Create          bool
	GroupKey        string
	ParamKey        string
	NextGroupKey    string
	NextParamKey    string
	NextValueType   string
	NextDescription string
	ValueEdits      []ParameterValueEdit
}

type ParameterValueEdit struct {
	Label     string
	NextValue string
}
