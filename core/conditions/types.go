package conditions

import "time"

// Tree is an order-aware, read-only view of the conditions in a Remote Config
// template. Conditions are kept in evaluation priority order.
type Tree struct {
	Version    string    `json:"version"`
	CachedAt   time.Time `json:"cached_at"`
	ETag       string    `json:"etag,omitempty"`
	Conditions []Entry   `json:"conditions"`

	parameters []parameterState
}

type Entry struct {
	Priority    int     `json:"priority"`
	Name        string  `json:"name"`
	Expression  string  `json:"expression"`
	Description string  `json:"description,omitempty"`
	TagColor    string  `json:"tag_color,omitempty"`
	Usages      []Usage `json:"usages"`
}

// DetailsEdit describes an atomic edit of a condition definition and its
// evaluation priority.
type DetailsEdit struct {
	Name           string
	NextName       string
	NextExpression string
	NextTagColor   string
	NextPriority   int
	ValueEdits     []UsageValueEdit
}

type UsageValueEdit struct {
	GroupKey     string
	ParameterKey string
	NextValue    string
}

type Usage struct {
	GroupKey     string `json:"group_key,omitempty"`
	GroupLabel   string `json:"group"`
	ParameterKey string `json:"parameter"`
	Value        string `json:"value"`
	RawValue     string `json:"raw_value,omitempty"`
	ValueType    string `json:"value_type"`
	Plain        bool   `json:"plain"`
}

type ParameterRef struct {
	GroupKey     string `json:"group_key,omitempty"`
	GroupLabel   string `json:"group"`
	ParameterKey string `json:"parameter"`
}

type DeleteImpact struct {
	Condition         Entry          `json:"condition"`
	Usages            []Usage        `json:"usages"`
	RemovedParameters []ParameterRef `json:"removed_parameters"`
}

type MoveImpact struct {
	Condition          Entry          `json:"condition"`
	FromPriority       int            `json:"from_priority"`
	ToPriority         int            `json:"to_priority"`
	CrossedConditions  []string       `json:"crossed_conditions"`
	AffectedParameters []ParameterRef `json:"affected_parameters"`
}

type parameterState struct {
	ref               ParameterRef
	hasDefault        bool
	conditionalValues map[string]struct{}
}
