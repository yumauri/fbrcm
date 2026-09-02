package ops

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/spf13/pflag"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/ops/contract"
	"github.com/yumauri/fbrcm/ops/invocation"
	"github.com/yumauri/fbrcm/ops/launchflags"
	"github.com/yumauri/fbrcm/ops/machine"
	"github.com/yumauri/fbrcm/ops/shared"
	add "github.com/yumauri/fbrcm/ops/workflows/add"
	apply "github.com/yumauri/fbrcm/ops/workflows/apply"
	"github.com/yumauri/fbrcm/ops/workflows/conditions"
	deletecmd "github.com/yumauri/fbrcm/ops/workflows/delete"
	"github.com/yumauri/fbrcm/ops/workflows/doctor"
	"github.com/yumauri/fbrcm/ops/workflows/draft"
	"github.com/yumauri/fbrcm/ops/workflows/duplicate"
	"github.com/yumauri/fbrcm/ops/workflows/get"
	"github.com/yumauri/fbrcm/ops/workflows/groups"
	"github.com/yumauri/fbrcm/ops/workflows/managedfeatures"
	"github.com/yumauri/fbrcm/ops/workflows/plan"
	"github.com/yumauri/fbrcm/ops/workflows/project"
	"github.com/yumauri/fbrcm/ops/workflows/projects"
	"github.com/yumauri/fbrcm/ops/workflows/update"
	"github.com/yumauri/fbrcm/ops/workflows/versions"
	"github.com/yumauri/fbrcm/schemas"
)

type Registry struct {
	service      *core.Core
	factories    map[string]func() *invocation.Definition
	capabilities []contract.Capability
}

func NewRegistry(service *core.Core) (*Registry, error) {
	r := &Registry{service: service, factories: map[string]func() *invocation.Definition{}}
	for name, factory := range map[string]func(*core.Core) *invocation.Definition{
		"add": add.NewDefinition, "apply": apply.NewDefinition, "conditions": conditions.NewDefinition,
		"delete": deletecmd.NewDefinition, "doctor": doctor.NewDefinition, "draft": draft.NewDefinition,
		"duplicate": duplicatecmd.NewDefinition, "get": get.NewDefinition, "groups": groups.NewDefinition,
		"experiments": managedfeatures.NewExperimentsDefinition, "rollouts": managedfeatures.NewRolloutsDefinition,
		"personalizations": managedfeatures.NewPersonalizationsDefinition, "project": project.NewDefinition,
		"projects": projects.NewDefinition, "update": updatecmd.NewDefinition, "versions": versions.NewDefinition,
	} {
		r.factories[name] = func() *invocation.Definition { return factory(service) }
	}
	r.factories["plan"] = plancmd.NewDefinition
	if err := json.Unmarshal(schemas.CapabilitiesJSON, &r.capabilities); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *Registry) Capabilities() []contract.Capability {
	var result []contract.Capability
	for _, capability := range r.capabilities {
		if _, err := r.definition(capability.ID); err == nil {
			result = append(result, capability)
		}
	}
	return result
}

func (r *Registry) definition(id string) (*invocation.Definition, error) {
	path := strings.Split(id, ".")
	factory, ok := r.factories[path[0]]
	if !ok {
		return nil, fmt.Errorf("unknown operation %q", id)
	}
	definition, err := factory().Find(path[1:])
	if err != nil {
		return nil, err
	}
	if definition.RunE == nil {
		return nil, fmt.Errorf("operation %q is not executable", id)
	}
	return definition, nil
}

type Execution struct {
	Profile, Version, BuildVersion                              string
	Stateless, NoLocalConfig, AllowHooks, AllowOAuth, Confirmed bool
	AuthTimeout                                                 time.Duration
	OAuthObserver                                               func(core.OAuthAuthorizationEvent)
	Stderr                                                      io.Writer
}

// request is independent of Cobra: it cannot parse argv, select another
// command, run a pre/post hook, render usage, or exit the host process.
type request struct {
	ctx   context.Context
	id    string
	flags *pflag.FlagSet
	in    io.Reader
	out   bytes.Buffer
	err   io.Writer
}

func (r *request) Context() context.Context       { return r.ctx }
func (r *request) SetContext(ctx context.Context) { r.ctx = ctx }
func (r *request) Flags() *pflag.FlagSet          { return r.flags }
func (r *request) InheritedFlags() *pflag.FlagSet { return r.flags }
func (r *request) CommandPath() string            { return "fbrcm " + strings.ReplaceAll(r.id, ".", " ") }
func (r *request) InOrStdin() io.Reader           { return r.in }
func (r *request) OutOrStdout() io.Writer         { return &r.out }
func (r *request) ErrOrStderr() io.Writer         { return r.err }
func (r *request) HasResponseModel() bool         { return true }

func newRequest(ctx context.Context, id string, execution Execution) *request {
	ctx = invocation.WithVersion(shared.WithMachineState(ctx), execution.BuildVersion)
	if execution.Stateless {
		ctx = machine.WithProfileless(ctx)
	}
	stderr := execution.Stderr
	if stderr == nil {
		stderr = io.Discard
	}
	return &request{ctx: ctx, id: id, flags: pflag.NewFlagSet(id, pflag.ContinueOnError), in: shared.NonInteractiveInput(), err: stderr}
}

func Failure(ctx context.Context, id, version string, stateless bool, err error) contract.Envelope {
	return contract.BuildEnvelope(newRequest(ctx, id, Execution{Stateless: stateless}), version, nil, err)
}

// Execute invokes the shared operation directly. Callers serialize operations
// while process-scoped profile/cache configuration remains in use.
func (r *Registry) Execute(ctx context.Context, c contract.Capability, input Input, execution Execution) contract.Envelope {
	call := newRequest(ctx, c.ID, execution)
	err := r.run(call, c, input, execution)
	if err == nil {
		err = call.ctx.Err()
	}
	return contract.BuildEnvelope(call, execution.Version, call.out.Bytes(), err)
}

func (r *Registry) run(call *request, c contract.Capability, input Input, execution Execution) error {
	if err := input.Validate(c); err != nil {
		return shared.InvalidArgument(err)
	}
	definition, err := r.definition(c.ID)
	if err != nil {
		return shared.InvalidArgument(err)
	}
	call.flags.AddFlagSet(definition.Flags())
	globals := pflag.NewFlagSet("launch", pflag.ContinueOnError)
	launchflags.Add(globals)
	call.flags.AddFlagSet(globals)
	keys := make([]string, 0, len(input.Options))
	for key := range input.Options {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if err := bindOption(call.flags, key, input.Options[key]); err != nil {
			return shared.InvalidArgument(err)
		}
	}
	for name, value := range map[string]string{"json": "true", "stateless": fmt.Sprint(execution.Stateless), "no-local-config": fmt.Sprint(execution.NoLocalConfig || execution.Stateless)} {
		if err := call.flags.Set(name, value); err != nil {
			return shared.InvalidArgument(err)
		}
	}
	if !execution.Stateless {
		if err := call.flags.Set("profile", execution.Profile); err != nil {
			return shared.InvalidArgument(err)
		}
	}
	if execution.Confirmed && c.Supports.ConfirmationBypass {
		if err := call.flags.Set("yes", "true"); err != nil {
			return shared.InvalidArgument(err)
		}
	}
	if raw := bytes.TrimSpace(input.Stdin); len(raw) != 0 && !bytes.Equal(raw, []byte("null")) {
		call.in = shared.DocumentInput(bytes.NewReader(raw))
	}
	args, err := input.Positionals(c)
	if err != nil {
		return shared.InvalidArgument(err)
	}
	if err := shared.ValidateNonBlankInputs(call, args); err != nil {
		return err
	}
	if err := shared.ValidateQueryFlags(call); err != nil {
		return err
	}
	if b := definition.Args; len(args) < b.Min || b.Max >= 0 && len(args) > b.Max {
		return shared.InvalidArgument(fmt.Errorf("invalid argument count for %s", c.ID))
	}
	for _, name := range definition.Required {
		if !call.flags.Changed(name) {
			return shared.InvalidArgument(fmt.Errorf("required option --%s is missing", name))
		}
	}
	for _, names := range definition.Exclusive {
		count := 0
		for _, name := range names {
			if call.flags.Changed(name) {
				count++
			}
		}
		if count > 1 {
			return shared.InvalidArgument(fmt.Errorf("options %s are mutually exclusive", strings.Join(names, ", ")))
		}
	}
	if err := call.ctx.Err(); err != nil {
		return err
	}
	if err := PreparePolicy(call, execution.AllowHooks, execution.AllowOAuth); err != nil {
		return err
	}
	call.ctx = firebase.WithOAuthTimeout(call.ctx, execution.AuthTimeout)
	config.SetLocalConfigDisabled(execution.NoLocalConfig || execution.Stateless)
	if !execution.Stateless {
		if err := config.SetProfileOverride(execution.Profile); err != nil {
			return &shared.ValidationError{Code: "profile.invalid", Source: "profile", Stage: "selection", Err: err}
		}
	}
	if r.service != nil {
		r.service.ResetFirebaseClients()
		r.service.ConfigureOAuthAuthorization(false, execution.OAuthObserver)
		r.service.SetHookOutput(call.err)
	}
	if err := ConfigureRequests(call, r.service, execution.Stateless); err != nil {
		return err
	}
	firebase.InitOfflineModeContext(call.ctx, false)
	shared.MarkCommandRun(call)
	return definition.RunE(call, args)
}
