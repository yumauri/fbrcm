// Package explain explains parameter selection against captured client facts.
package explain

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/conditions"
	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/core/rc/evaluate"
	rctarget "github.com/yumauri/fbrcm/core/rc/target"
	"github.com/yumauri/fbrcm/internal/terminal/progress"
	"github.com/yumauri/fbrcm/ops/invocation"
	"github.com/yumauri/fbrcm/ops/shared"
	sharedrc "github.com/yumauri/fbrcm/ops/shared/rc"
)

func NewDefinition(svc *core.Core) *invocation.Definition {
	cmd := &invocation.Definition{
		Use: "explain <parameter>", Short: "Explain a parameter value using a captured cURL request or HAR",
		Args: invocation.ExactArgs(1),
		RunE: func(cmd invocation.Call, args []string) error {
			return run(cmd, svc, args[0], newPrompter(cmd))
		},
	}
	cmd.Flags().String("from", "", "Read a cURL command or HAR from this file, or - for stdin; piped stdin is detected automatically")
	cmd.Flags().StringP("project", "p", "", "Select a project; otherwise infer it from the captured request")
	cmd.Flags().Int("entry", 0, "Select a fetch by its original HAR entry number")
	cmd.Flags().StringArray("set", nil, "Supply an input as field=value; may be repeated (use /unknown or /unset explicitly)")
	cmd.Flags().Bool("no-input", false, "Leave missing inputs unknown instead of prompting")
	shared.AddParametersReadFlags(cmd, "Revalidate cached Remote Config before explaining")
	invocation.RegisterResponse(cmd, explanationJSON{})
	return cmd
}

type conditionResult struct {
	Entry  core.ConditionEntry
	Value  firebase.RemoteConfigValue
	Result evaluate.Result
}

type report struct {
	Project, Parameter, Version, Source, Capture string
	Param                                        firebase.RemoteConfigParam
	Exchange                                     *evaluate.Exchange
	Conditions                                   []conditionResult
	Winner                                       int
	Unresolved                                   bool
	Assumptions                                  bool
	Group, ETag                                  string
	CachedAt                                     time.Time
	Inputs, Facts                                map[string]evaluate.Fact
}

func run(cmd invocation.Call, svc *core.Core, key string, prompts prompter) error {
	machineMode := shared.MachineMode(cmd)
	opts, err := shared.ReadParametersReadOptions(cmd)
	if err != nil {
		return err
	}
	from, _ := cmd.Flags().GetString("from")
	query, _ := cmd.Flags().GetString("project")
	entry, _ := cmd.Flags().GetInt("entry")
	if entry < 0 {
		return shared.InvalidArgument(fmt.Errorf("--entry must be a positive HAR entry number"))
	}
	exchange, err := readExchange(cmd, from, entry, prompts)
	if err != nil {
		return err
	}
	if exchange == nil && query == "" {
		return shared.InvalidArgument(fmt.Errorf("provide a cURL command or HAR with --from or stdin, or select a project with --project"))
	}
	project, err := resolveProject(cmd, svc, query, exchange, opts.Cached)
	if err != nil {
		return err
	}
	ctx, err := shared.FirebaseServiceContextForExecution(shared.CommandContext(cmd), project.ProjectID)
	if err != nil {
		return err
	}
	cmd.SetContext(ctx)
	cache, source, err := shared.ReadParametersWithFallback(ctx, svc, project.ProjectID, opts)
	if err != nil {
		return err
	}
	cfg, err := sharedrc.ParseProjectRemoteConfig(project.ProjectID, cache.RemoteConfig)
	if err != nil {
		return err
	}
	var parameter *firebase.RemoteConfigParam
	var group string
	for _, target := range shared.CollectParamTargets(cfg) {
		if target.Key == key {
			parameter = &target.Param
			group = target.Group
			break
		}
	}
	if parameter == nil {
		return &shared.SelectionError{Resource: "parameter", Kind: "not_found", Query: key, Err: fmt.Errorf("parameter %q was not found in project %s", key, project.ProjectID)}
	}
	r := report{Project: project.ProjectID, Parameter: key, Version: cfg.Version.VersionNumber, Source: source, Capture: from, Param: *parameter, Exchange: exchange, Winner: -1}
	r.Group, r.ETag, r.CachedAt = group, cache.ETag, cache.CachedAt
	facts := make(map[string]evaluate.Fact)
	if exchange != nil {
		facts = exchange.Facts
		if r.Capture == "" || r.Capture == "-" {
			r.Capture = "stdin"
		}
	}
	values, _ := cmd.Flags().GetStringArray("set")
	for _, value := range values {
		field, answer, ok := strings.Cut(value, "=")
		need, valid := evaluate.InputNeed(field)
		if !ok || !valid {
			return shared.InvalidArgument(fmt.Errorf("--set requires a supported input field=value"))
		}
		fact, err := evaluate.Answer(need, answer, "command line")
		if err != nil {
			return shared.InvalidArgument(err)
		}
		facts[need.Key] = fact
		r.Assumptions = true
	}
	if !machineMode {
		progress.Stop()
		if _, err := fmt.Fprintln(cmd.OutOrStdout(), renderHeader(r, shared.TerminalWidth())); err != nil {
			return err
		}
	}
	noInput, _ := cmd.Flags().GetBool("no-input")
	evaluation := &evaluate.Context{Facts: facts, Capture: exchange, Resolve: func(need evaluate.Need) (evaluate.Fact, error) {
		if err := ctx.Err(); err != nil {
			return evaluate.Fact{}, err
		}
		if noInput || machineMode {
			return evaluate.Fact{}, nil
		}
		fact, err := prompts.Fact(need)
		if fact.Known {
			r.Assumptions = true
		}
		return fact, err
	}}
	tree := conditions.BuildTree(cfg, cache.CachedAt, cache.ETag)
	for _, condition := range tree.Conditions {
		value, relevant := parameter.ConditionalValues[condition.Name]
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
	if len(r.Conditions) != len(parameter.ConditionalValues) {
		return &shared.ValidationError{Code: "remote_config.invalid", Source: "template", Stage: "evaluation", Err: fmt.Errorf("parameter %q references a condition missing from the template", key)}
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	progress.Stop()
	if machineMode {
		r.Inputs, r.Facts = evaluation.Used, facts
		if r.Inputs == nil {
			r.Inputs = make(map[string]evaluate.Fact)
		}
		for _, value := range values {
			field, _, _ := strings.Cut(value, "=")
			need, _ := evaluate.InputNeed(field)
			r.Inputs[need.Key] = facts[need.Key]
		}
		if source == "cache-stale" {
			shared.AddMachineWarning(cmd, shared.MachineWarning{Code: "cache.stale", Message: "Evaluation uses a stale cached template", Target: project.ProjectID, Details: struct {
				Source string `json:"source"`
			}{source}})
		}
		if exchange != nil && exchange.Version != "" && exchange.Version != r.Version {
			shared.AddMachineWarning(cmd, shared.MachineWarning{Code: "explain.capture_version_mismatch", Message: "Captured response belongs to another template version", Target: project.ProjectID, Details: struct {
				Captured  string `json:"captured_version"`
				Evaluated string `json:"evaluated_version"`
			}{exchange.Version, r.Version}})
		}
		return shared.WriteJSON(cmd, toJSON(r))
	}
	_, err = fmt.Fprintln(cmd.OutOrStdout(), renderEvaluation(r, shared.TerminalWidth()))
	return err
}

func readExchange(cmd invocation.Call, from string, entry int, prompts prompter) (*evaluate.Exchange, error) {
	var input io.Reader
	switch {
	case from == "-" || from == "" && shared.StdinAvailable(cmd.InOrStdin()):
		input = cmd.InOrStdin()
	case from != "":
		file, err := os.Open(from)
		if err != nil {
			return nil, fmt.Errorf("open capture: %w", err)
		}
		defer func() { _ = file.Close() }()
		input = file
	default:
		if entry != 0 {
			return nil, shared.InvalidArgument(fmt.Errorf("--entry requires a capture"))
		}
		return nil, nil
	}
	exchanges, err := evaluate.ReadCapture(input)
	if err != nil {
		if _, ok := errors.AsType[*evaluate.CaptureError](err); ok {
			return nil, &shared.ValidationError{Code: "capture.invalid", Source: "capture", Stage: "parse", Err: err}
		}
		return nil, err
	}
	if entry != 0 {
		for _, exchange := range exchanges {
			if exchange.Entry == entry {
				return &exchange, nil
			}
		}
		return nil, shared.InvalidArgument(fmt.Errorf("HAR entry %d is not a Firebase fetch request", entry))
	}
	if len(exchanges) == 1 {
		return &exchanges[0], nil
	}
	noInput, _ := cmd.Flags().GetBool("no-input")
	if shared.MachineMode(cmd) || noInput {
		candidates := make([]shared.CaptureCandidate, 0, len(exchanges))
		for _, x := range exchanges {
			candidates = append(candidates, shared.CaptureCandidate{Entry: x.Entry, Project: x.Project, Namespace: x.Namespace, CapturedAt: optionalTime(x.StartedAt), AppID: optionalText(x.Facts["app.id"].Value), HTTPStatus: x.Status})
		}
		return nil, &shared.InteractionError{Message: "capture has multiple Firebase fetches; select an original HAR entry with --entry", Type: "capture_selection", RequiredOption: "--entry", CaptureCandidates: candidates}
	}
	selected, err := prompts.Exchange(exchanges)
	if err != nil {
		return nil, err
	}
	return &selected, nil
}

func resolveProject(cmd invocation.Call, svc *core.Core, query string, exchange *evaluate.Exchange, cached bool) (core.Project, error) {
	ctx := shared.CommandContext(cmd)
	var project core.Project
	var err error
	if query == "" && exchange != nil {
		query = exchange.Project
	}
	if _, numeric := strconv.ParseUint(query, 10, 64); numeric == nil {
		if cached {
			project, err = shared.ResolveCachedProjectNumber(cmd, query)
		} else {
			project, err = shared.ResolveProjectNumberForExecution(ctx, cmd, svc, query)
		}
	} else if cached {
		project, err = shared.ResolveCachedProjectTargetArg(cmd, query)
	} else {
		project, err = shared.ResolveProjectTargetForExecution(ctx, cmd, svc, query)
	}
	if err != nil {
		return core.Project{}, err
	}
	if exchange == nil {
		return project, nil
	}
	target, err := rctarget.Parse(project.ProjectID)
	if err != nil {
		return core.Project{}, err
	}
	if target.ProjectID != exchange.Project && project.ProjectNumber != exchange.Project {
		return core.Project{}, shared.InvalidArgument(fmt.Errorf("selected project does not match the captured fetch project"))
	}
	kind := rctarget.Client
	if exchange.Namespace == rctarget.ServerNamespace {
		kind = rctarget.Server
	}
	if _, explicit, _ := rctarget.ParseSelector(query); explicit && target.Kind != kind {
		return core.Project{}, shared.InvalidArgument(fmt.Errorf("selected template does not match the captured fetch namespace"))
	}
	project.ProjectID = (rctarget.Target{Kind: kind, ProjectID: target.ProjectID}).String()
	return project, nil
}

// Keep the workflow independently testable without opening terminal devices.
type prompter interface {
	Fact(evaluate.Need) (evaluate.Fact, error)
	Exchange([]evaluate.Exchange) (evaluate.Exchange, error)
}
