package explain

import (
	"maps"
	"slices"
	"time"

	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/core/rc/evaluate"
	rctarget "github.com/yumauri/fbrcm/core/rc/target"
)

func optionalText(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
func optionalTime(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	return &value
}

func machineValue(value *firebase.RemoteConfigValue) valueJSON {
	if value == nil || value.UseInAppDefault {
		return valueJSON{Kind: "in_app_default"}
	}
	if value.IsOpaque() {
		kind := "unrecognized"
		switch {
		case len(value.PersonalizationValue) > 0:
			kind = "personalization"
		case len(value.ExperimentValue) > 0:
			kind = "experiment"
		case len(value.RolloutValue) > 0:
			kind = "rollout"
		}
		return valueJSON{Kind: "managed", ManagedKind: &kind}
	}
	return valueJSON{Kind: "literal", Value: new(value.Value)}
}

func outcomeCandidates(r report) ([]int, map[int]int) {
	results := make([]evaluate.Result, len(r.Conditions))
	for i, condition := range r.Conditions {
		results[i] = condition.Result
	}
	return evaluate.OutcomeCandidates(results)
}

func toJSON(r report) explanationJSON {
	target, _ := rctarget.Parse(r.Project)
	parameterType := r.Param.ValueType
	if parameterType == "" {
		parameterType = "STRING"
	}
	data := explanationJSON{ProjectID: target.ProjectID, Template: templateJSON{Kind: string(target.Kind), Version: r.Version, ETag: r.ETag, Source: r.Source, CachedAt: optionalTime(r.CachedAt)},
		Parameter: parameterJSON{Key: r.Parameter, Type: parameterType, Group: optionalText(r.Group), Description: r.Param.Description, DefaultValue: machineValue(r.Param.DefaultValue)},
		Inputs:    []inputJSON{}, Conditions: []conditionJSON{}, PossibleOutcomes: []outcomeJSON{}, NeededInputs: []string{}}
	eligible, blocked := outcomeCandidates(r)
	eligibleSet := make(map[int]bool)
	for _, i := range eligible {
		eligibleSet[i] = true
	}
	for i, c := range r.Conditions {
		item := conditionJSON{Name: c.Entry.Name, Priority: c.Entry.Priority, TagColor: c.Entry.TagColor, Expression: c.Entry.Expression, Value: machineValue(&c.Value), Evaluation: c.Result.State, Selection: "undetermined", ExclusionReason: "none", BlockedBy: []string{}, Rules: append([]evaluate.RuleEvaluation{}, c.Result.Rules...)}
		switch {
		case c.Result.State == evaluate.False:
			item.Selection, item.ExclusionReason = "not_selected", "not_matched"
		case !eligibleSet[i]:
			item.Selection, item.ExclusionReason = "not_selected", "subsumed"
			if prior, ok := blocked[i]; ok {
				item.BlockedBy = append(item.BlockedBy, r.Conditions[prior].Entry.Name)
				if r.Conditions[prior].Result.State == evaluate.True {
					item.ExclusionReason = "higher_priority_match"
				}
			}
		case !r.Unresolved && i == r.Winner:
			item.Selection = "selected"
		}
		data.Conditions = append(data.Conditions, item)
	}
	data.Result = resultJSON{Scope: "template", Selection: selectionJSON{State: "resolved", Source: new("default")}, Value: machineValue(r.Param.DefaultValue)}
	if r.Winner >= 0 {
		winner := r.Conditions[r.Winner]
		data.Result.Selection.Source = new("condition")
		data.Result.Selection.Condition = new(winner.Entry.Name)
		data.Result.Value = machineValue(&winner.Value)
	}
	if r.Unresolved {
		data.Result.Selection = selectionJSON{State: "unresolved"}
		data.Result.Value = valueJSON{Kind: "unknown"}
		guards := []guardJSON{}
		fallback := true
		for _, i := range eligible {
			c := r.Conditions[i]
			when := append([]guardJSON{}, guards...)
			when = append(when, guardJSON{Condition: c.Entry.Name, Matches: true})
			data.PossibleOutcomes = append(data.PossibleOutcomes, outcomeJSON{Source: "condition", Condition: new(c.Entry.Name), When: when, Value: machineValue(&c.Value)})
			if c.Result.State == evaluate.True {
				fallback = false
				break
			}
			guards = append(guards, guardJSON{Condition: c.Entry.Name, Matches: false})
		}
		if fallback {
			data.PossibleOutcomes = append(data.PossibleOutcomes, outcomeJSON{Source: "default", When: guards, Value: machineValue(r.Param.DefaultValue)})
		}
		// A value can be certain even while its supplying condition is unknown.
		if len(data.PossibleOutcomes) > 0 {
			first := data.PossibleOutcomes[0].Value
			if first.Kind == "literal" && !slices.ContainsFunc(data.PossibleOutcomes, func(o outcomeJSON) bool { return o.Value.Kind != "literal" || *o.Value.Value != *first.Value }) {
				data.Result.Value = first
			}
		}
	}
	data.Inputs, data.NeededInputs = machineInputs(r, eligibleSet)
	if r.Exchange != nil {
		data.Capture = machineCapture(r, data.Result)
	}
	return data
}

func machineInputs(r report, eligible map[int]bool) ([]inputJSON, []string) {
	facts := maps.Clone(r.Inputs)
	if facts == nil {
		facts = make(map[string]evaluate.Fact)
	}
	for _, fact := range r.Inputs {
		for _, parent := range fact.DerivedFrom {
			if value, ok := r.Facts[parent]; ok {
				facts[parent] = value
			}
		}
	}
	keys := make([]string, 0, len(facts))
	for key := range facts {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	inputs, needed := []inputJSON{}, []string{}
	for _, key := range keys {
		fact := facts[key]
		need, ok := evaluate.InputNeed(key)
		if !ok {
			continue
		}
		item := inputJSON{Field: key, Type: need.Kind, State: "known", Value: new(fact.Value), Source: factSource(fact), ReasonCode: "none", Message: fact.Reason, TimeMin: optionalTime(fact.TimeMin), TimeMax: optionalTime(fact.TimeMax), UsedBy: []inputUseJSON{}}
		switch {
		case !fact.Known:
			item.State, item.Value, item.ReasonCode = "unknown", nil, "missing_input"
			if fact.Reason != "" {
				item.ReasonCode = "input_unavailable"
			}
			if fact.Value != "" {
				item.ObservedValue = new(fact.Value)
			}
		case fact.Absent:
			item.State, item.Value = "absent", nil
		case item.TimeMin != nil || item.TimeMax != nil:
			item.State, item.Value, item.ObservedValue, item.ReasonCode = "bounded", nil, new(fact.Value), "bounded_observation"
		}
		for i, c := range r.Conditions {
			for _, rule := range c.Result.Rules {
				if slices.Contains(rule.Inputs, key) {
					item.UsedBy = append(item.UsedBy, inputUseJSON{Condition: c.Entry.Name, Rule: rule.ID})
					if eligible[i] && rule.Evaluation != nil && *rule.Evaluation == evaluate.Unknown && (item.State == "unknown" || item.State == "bounded") {
						item.RequiredForResolution = true
					}
				}
			}
		}
		if item.RequiredForResolution {
			needed = append(needed, key)
		}
		inputs = append(inputs, item)
	}
	return inputs, needed
}

func factSource(f evaluate.Fact) sourceJSON {
	s := sourceJSON{Kind: "unknown", Description: f.Source, DerivedFrom: append([]string{}, f.DerivedFrom...)}
	if len(f.DerivedFrom) > 0 {
		s.Kind = "derived"
		return s
	}
	switch f.Source {
	case "request", "request locale":
		s.Kind = "request_body"
	case "User-Agent header", "Client Hints header", "Accept-Language header":
		s.Kind = "request_header"
	case "response Date header", "capture timestamp":
		s.Kind = "capture_metadata"
	case "command line":
		s.Kind = "override"
	case "request Analytics _fot":
		s.Kind = "derived"
		s.DerivedFrom = []string{"app.userProperty['_fot']"}
	}
	return s
}

func machineCapture(r report, result resultJSON) *captureJSON {
	x := r.Exchange
	response := capturedResponseJSON{HTTPStatus: x.Status, ParameterState: "unavailable", TemplateVersion: optionalText(x.Version), State: optionalText(x.ResponseState), Assignments: append([]evaluate.Assignment{}, x.Assignments[r.Parameter]...), Experiments: append([]string{}, x.Experiments...), Comparison: "unavailable", ComparisonReason: "response_unavailable"}
	if x.HasResponse {
		response.ParameterState = "absent"
		response.Comparison = "inconclusive"
		response.ComparisonReason = "parameter_absent"
		if value, ok := x.Entries[r.Parameter]; ok {
			response.ParameterState, response.Value = "present", &value
			switch {
			case len(response.Assignments) > 0 || len(response.Experiments) > 0:
				response.ComparisonReason = "firebase_managed"
			case x.Version == "":
				response.ComparisonReason = "template_version_unknown"
			case x.Version != r.Version:
				response.ComparisonReason = "template_version_mismatch"
			case result.Value.Kind != "literal":
				response.ComparisonReason = "prediction_unresolved"
			default:
				response.ComparisonReason = "none"
				response.Comparison = "mismatch"
				if value == *result.Value.Value {
					response.Comparison = "match"
				}
			}
		}
	}
	return &captureJSON{Format: x.Format, Entry: x.Entry, CapturedAt: optionalTime(x.StartedAt), Response: response}
}
