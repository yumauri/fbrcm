package ops

import (
	"fmt"
	"strings"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/env"
	"github.com/yumauri/fbrcm/core/firebase"
	corelog "github.com/yumauri/fbrcm/core/log"
	"github.com/yumauri/fbrcm/ops/contract"
	"github.com/yumauri/fbrcm/ops/invocation"
	"github.com/yumauri/fbrcm/ops/machine"
	"github.com/yumauri/fbrcm/ops/shared"
)

// PreparePolicy is shared by CLI and structured callers. Interaction permission
// is explicit and separate from Firebase credentials and mutation approval.
func PreparePolicy(call invocation.Call, allowHooks, allowOAuth bool) error {
	ctx := shared.CommandContext(call)
	stateless, err := call.Flags().GetBool("stateless")
	if err != nil {
		return shared.InvalidArgument(err)
	}
	if stateless {
		ctx = machine.WithProfileless(core.WithExecutionPolicy(ctx, core.StatelessExecutionPolicy()))
		call.SetContext(ctx)
		id := contract.CommandID(call)
		if !contract.SupportsStatelessCommand(id) {
			return shared.InvalidArgument(fmt.Errorf("--stateless is not supported by %s", strings.TrimPrefix(call.CommandPath(), "fbrcm ")))
		}
		if call.Flags().Changed("profile") {
			return shared.InvalidArgument(fmt.Errorf("--profile cannot be used with --stateless"))
		}
		requiresToken := contract.StatelessCommandRequiresAccessToken(id)
		if id == "get" && shared.StdinAvailable(call.InOrStdin()) {
			requiresToken = false
		}
		if requiresToken {
			if _, ok := env.LookupNonEmpty(env.GoogleAccessToken); !ok {
				return &core.AuthError{Kind: "configuration", Err: fmt.Errorf("%s is required with --stateless", env.GoogleAccessToken)}
			}
		}
		corelog.For("cli.stateless").Info("stateless mode enabled", "command", id)
	} else {
		policy := core.StatefulExecutionPolicy()
		policy.RunHooks = allowHooks
		ctx = core.WithExecutionPolicy(ctx, policy)
	}
	if contract.Enabled(call) {
		ctx = firebase.WithOAuthInteractionAllowed(ctx, allowOAuth)
	}
	call.SetContext(ctx)
	return nil
}

func ConfigureRequests(call invocation.Call, service *core.Core, stateless bool) error {
	if service == nil {
		return nil
	}
	if stateless {
		service.ResetFirebaseRequestPolicy()
	} else if err := service.ConfigureFirebaseRequests(); err != nil {
		if contract.CommandID(call) != "doctor" {
			return err
		}
		service.ResetFirebaseRequestPolicy()
	}
	call.SetContext(service.WithFirebaseRequestController(shared.CommandContext(call)))
	return nil
}
