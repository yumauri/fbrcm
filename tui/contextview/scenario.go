// Package contextview evaluates a TUI browsing scenario over a loaded template.
package contextview

import (
	"maps"
	"slices"

	"github.com/yumauri/fbrcm/core"
	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
	"github.com/yumauri/fbrcm/core/rc/evaluate"
	rctarget "github.com/yumauri/fbrcm/core/rc/target"
	clistyles "github.com/yumauri/fbrcm/internal/terminal/styles"
)

type Scenario struct {
	scope        rctarget.Target
	evaluation   evaluate.Context
	results      map[string]evaluate.Result
	showExcluded bool
}

func New(facts map[string]evaluate.Fact, capture *evaluate.Exchange, scope rctarget.Target, showExcluded bool) *Scenario {
	if len(facts) == 0 && scope.ProjectID == "" {
		return nil
	}
	return &Scenario{
		scope: scope, showExcluded: showExcluded,
		evaluation: evaluate.Context{Facts: maps.Clone(facts), Capture: capture},
		results:    make(map[string]evaluate.Result),
	}
}

func (s *Scenario) AppliesTo(project core.Project) bool {
	if s == nil {
		return false
	}
	if s.scope.ProjectID == "" {
		return true
	}
	target, err := rctarget.Parse(project.ProjectID)
	return err == nil && target.Kind == s.scope.Kind && s.scope.MatchesProject(target.ProjectID, project.ProjectNumber)
}

func (s *Scenario) Evaluate(expression string) evaluate.Result {
	if result, ok := s.results[expression]; ok {
		return result
	}
	// This context has no interactive resolver. Unsupported rules return Unknown.
	result, err := evaluate.Evaluate(expression, &s.evaluation)
	if err != nil {
		return evaluate.Result{State: evaluate.Unknown}
	}
	s.results[expression] = result
	return result
}

func (s *Scenario) Visible(state evaluate.State) bool {
	return s == nil || state != evaluate.False || s.showExcluded
}

func ConditionStatus(state evaluate.State) string {
	switch state {
	case evaluate.True:
		return "Matches"
	case evaluate.False:
		return "Excluded"
	case evaluate.Unknown:
		return "Needs input"
	default:
		return ""
	}
}

func ResultStatus(result evaluate.Result) string {
	if result.Unsupported {
		return "Evaluation unavailable"
	}
	return ConditionStatus(result.State)
}

type Value struct {
	Result     evaluate.Result
	Expression string
	Status     string
	Visible    bool
}

// ValueNote renders a parameter's evaluation outcome with the CLI note style.
func ValueNote(status string) string {
	if status == "" {
		return ""
	}
	return clistyles.PanelMuted.Render(" · ") + clistyles.OutcomeNoteStyle(status, clistyles.PanelMuted).Render(status)
}

// Values is indexed by the original parameter value indexes, including the
// default. Its caller can project rows without changing editor identities.
func (s *Scenario) Values(project core.Project, conditions []core.ParametersCondition, param core.ParametersEntry) []Value {
	if !s.AppliesTo(project) {
		return nil
	}
	values := make([]Value, len(param.Values))
	results := make([]evaluate.Result, 0, len(param.Values))
	indexes := make([]int, 0, len(param.Values))
	defaultIndex := -1
	for i, value := range param.Values {
		if value.Label == "default" {
			defaultIndex = i
			continue
		}
		result := evaluate.Result{State: evaluate.Unknown}
		expression := ""
		if index := slices.IndexFunc(conditions, func(c core.ParametersCondition) bool { return c.Name == value.Label }); index >= 0 {
			expression = conditions[index].Expression
			result = s.Evaluate(expression)
		}
		status := ResultStatus(result)
		if result.State == evaluate.False {
			status = rcdisplay.OutcomeDoesNotMatch
		}
		values[i] = Value{Result: result, Expression: expression, Status: status}
		results = append(results, result)
		indexes = append(indexes, i)
	}
	eligible, blocked := evaluate.OutcomeCandidates(results)
	for i := range blocked {
		if results[i].State == evaluate.True {
			values[indexes[i]].Status = rcdisplay.OutcomeLowerPriority
		}
	}
	selected := -1
	definite := len(eligible) > 0 && results[eligible[len(eligible)-1]].State == evaluate.True
	if definite && len(eligible) == 1 {
		selected = indexes[eligible[0]]
	}
	if defaultIndex >= 0 {
		switch {
		case definite:
			values[defaultIndex].Status = rcdisplay.OutcomeNotUsed
		case len(eligible) == 0:
			selected = defaultIndex
		default:
			values[defaultIndex].Status = "Possible fallback"
		}
	}
	if selected >= 0 {
		values[selected].Status = rcdisplay.OutcomeSelected
	}
	for i := range values {
		values[i].Visible = s.showExcluded || (values[i].Result.State != evaluate.False && (selected < 0 || i == selected))
	}
	return values
}
