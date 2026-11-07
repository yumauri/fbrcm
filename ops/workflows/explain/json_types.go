package explain

import (
	"github.com/yumauri/fbrcm/core/rc/evaluate"
	"time"
)

type explanationJSON struct {
	ProjectID        string          `json:"project_id"`
	Template         templateJSON    `json:"template"`
	Parameter        parameterJSON   `json:"parameter"`
	Capture          *captureJSON    `json:"capture"`
	Inputs           []inputJSON     `json:"inputs"`
	Conditions       []conditionJSON `json:"conditions"`
	Result           resultJSON      `json:"result"`
	PossibleOutcomes []outcomeJSON   `json:"possible_outcomes"`
	NeededInputs     []string        `json:"needed_inputs"`
}
type templateJSON struct {
	Kind     string     `json:"kind" contract:"enum=client|server"`
	Version  string     `json:"version"`
	ETag     string     `json:"etag"`
	Source   string     `json:"source" contract:"enum=cache|cache-verified|cache-stale|firebase"`
	CachedAt *time.Time `json:"cached_at"`
}
type parameterJSON struct {
	Key          string    `json:"key"`
	Type         string    `json:"type" contract:"enum=STRING|BOOLEAN|NUMBER|JSON"`
	Group        *string   `json:"group"`
	Description  string    `json:"description"`
	DefaultValue valueJSON `json:"default_value"`
}
type valueJSON struct {
	Kind        string  `json:"kind" contract:"enum=literal|in_app_default|managed|unknown"`
	Value       *string `json:"value"`
	ManagedKind *string `json:"managed_kind" contract:"enum=personalization|experiment|rollout|unrecognized"`
}
type selectionJSON struct {
	State     string  `json:"state" contract:"enum=resolved|unresolved"`
	Source    *string `json:"source" contract:"enum=condition|default"`
	Condition *string `json:"condition"`
}
type resultJSON struct {
	Scope     string        `json:"scope" contract:"enum=template"`
	Selection selectionJSON `json:"selection"`
	Value     valueJSON     `json:"value"`
}
type conditionJSON struct {
	Name            string                    `json:"name"`
	Priority        int                       `json:"priority"`
	TagColor        string                    `json:"tag_color"`
	Expression      string                    `json:"expression"`
	Value           valueJSON                 `json:"value"`
	Evaluation      evaluate.State            `json:"evaluation" contract:"enum=true|false|unknown"`
	Selection       string                    `json:"selection" contract:"enum=selected|not_selected|undetermined"`
	ExclusionReason string                    `json:"exclusion_reason" contract:"enum=none|not_matched|higher_priority_match|subsumed"`
	BlockedBy       []string                  `json:"blocked_by"`
	Rules           []evaluate.RuleEvaluation `json:"rules"`
}
type guardJSON struct {
	Condition string `json:"condition"`
	Matches   bool   `json:"matches"`
}
type outcomeJSON struct {
	Source    string      `json:"source" contract:"enum=condition|default"`
	Condition *string     `json:"condition"`
	When      []guardJSON `json:"when"`
	Value     valueJSON   `json:"value"`
}
type sourceJSON struct {
	Kind        string   `json:"kind" contract:"enum=request_body|request_header|capture_metadata|override|derived|unknown"`
	Description string   `json:"description"`
	DerivedFrom []string `json:"derived_from"`
}
type inputUseJSON struct {
	Condition string `json:"condition"`
	Rule      string `json:"rule"`
}
type inputJSON struct {
	Field                 string         `json:"field"`
	Type                  string         `json:"type" contract:"enum=text|number|boolean|timestamp"`
	State                 string         `json:"state" contract:"enum=known|absent|unknown|bounded"`
	Value                 *string        `json:"value"`
	ObservedValue         *string        `json:"observed_value"`
	Source                sourceJSON     `json:"source"`
	ReasonCode            string         `json:"reason_code" contract:"enum=none|missing_input|input_unavailable|bounded_observation"`
	Message               string         `json:"message"`
	TimeMin               *time.Time     `json:"time_min"`
	TimeMax               *time.Time     `json:"time_max"`
	UsedBy                []inputUseJSON `json:"used_by"`
	RequiredForResolution bool           `json:"required_for_resolution"`
}
type captureJSON struct {
	Format     string               `json:"format" contract:"enum=curl|har"`
	Entry      int                  `json:"entry"`
	CapturedAt *time.Time           `json:"captured_at"`
	Response   capturedResponseJSON `json:"response"`
}
type capturedResponseJSON struct {
	HTTPStatus       int                   `json:"http_status"`
	ParameterState   string                `json:"parameter_state" contract:"enum=present|absent|unavailable"`
	Value            *string               `json:"value"`
	TemplateVersion  *string               `json:"template_version"`
	State            *string               `json:"state"`
	Assignments      []evaluate.Assignment `json:"assignments"`
	Experiments      []string              `json:"experiments"`
	Comparison       string                `json:"comparison" contract:"enum=match|mismatch|inconclusive|unavailable"`
	ComparisonReason string                `json:"comparison_reason" contract:"enum=none|response_unavailable|parameter_absent|template_version_unknown|template_version_mismatch|prediction_unresolved|firebase_managed"`
}
