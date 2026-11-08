package explain

import (
	"fmt"

	"github.com/yumauri/fbrcm/core/conditions"
	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/core/rc/evaluate"
	"github.com/yumauri/fbrcm/ops/shared"
)

// TemplateInput identifies the loaded template and the optional capture to compare.
type TemplateInput struct {
	Project, Parameter, Source, Capture string
	Config                              *firebase.RemoteConfig
	Exchange                            *evaluate.Exchange
	Assumptions                         bool
}

// PrepareHumanReport evaluates a loaded template and returns the CLI renderer
// for the resulting report. Rendering again at another width keeps the same facts.
func PrepareHumanReport(input TemplateInput, evaluation *evaluate.Context) (func(int) string, error) {
	r, err := newReport(input)
	if err != nil {
		return nil, err
	}
	if err := r.evaluate(input.Config, evaluation); err != nil {
		return nil, err
	}
	return func(width int) string { return render(r, width) }, nil
}

func newReport(input TemplateInput) (report, error) {
	for _, target := range shared.CollectParamTargets(input.Config) {
		if target.Key == input.Parameter {
			return report{
				Project: input.Project, Parameter: input.Parameter, Version: input.Config.Version.VersionNumber,
				Source: input.Source, Capture: input.Capture, Param: target.Param, Group: target.Group,
				Exchange: input.Exchange, Winner: -1, Assumptions: input.Assumptions,
			}, nil
		}
	}
	return report{}, &shared.SelectionError{
		Resource: "parameter", Kind: "not_found", Query: input.Parameter,
		Err: fmt.Errorf("parameter %q was not found in project %s", input.Parameter, input.Project),
	}
}

func (r *report) evaluate(cfg *firebase.RemoteConfig, evaluation *evaluate.Context) error {
	tree := conditions.BuildTree(cfg, r.CachedAt, r.ETag)
	for _, condition := range tree.Conditions {
		value, relevant := r.Param.ConditionalValues[condition.Name]
		if !relevant {
			continue
		}
		result, err := evaluate.Evaluate(condition.Expression, evaluation)
		if err != nil {
			return err
		}
		r.Conditions = append(r.Conditions, conditionResult{Entry: condition, Value: value, Result: result})
		if r.Winner < 0 {
			switch result.State {
			case evaluate.True:
				r.Winner = len(r.Conditions) - 1
			case evaluate.Unknown:
				r.Unresolved = true
			}
		}
	}
	if len(r.Conditions) != len(r.Param.ConditionalValues) {
		return &shared.ValidationError{
			Code: "remote_config.invalid", Source: "template", Stage: "evaluation",
			Err: fmt.Errorf("parameter %q references a condition missing from the template", r.Parameter),
		}
	}
	return nil
}
