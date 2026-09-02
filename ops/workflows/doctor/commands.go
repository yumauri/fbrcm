package doctor

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/ops/invocation"
	"github.com/yumauri/fbrcm/ops/shared"
)

type doctorFunc func(context.Context) core.DoctorReport
type notifyContextFunc func(context.Context) (context.Context, context.CancelFunc)

// New constructs the doctor command.
func NewDefinition(svc *core.Core) *invocation.Definition {
	cmd := newCommandDefinition(svc.Doctor, func(parent context.Context) (context.Context, context.CancelFunc) {
		return signal.NotifyContext(parent, os.Interrupt)
	})
	invocation.RegisterResponse(cmd, []doctorListItem{})
	return cmd
}

func newCommandDefinition(runDoctor doctorFunc, notifyContext notifyContextFunc) *invocation.Definition {
	cmd := &invocation.Definition{
		Use:   "doctor",
		Short: "Check profile, credentials, connectivity, APIs, permissions, and cache",
		Args:  invocation.NoArgs,
		RunE: func(cmd invocation.Call, args []string) error {
			timeout, err := cmd.Flags().GetDuration("timeout")
			if err != nil {
				return err
			}
			if cmd.Flags().Changed("timeout") && timeout <= 0 {
				return shared.InvalidArgument(fmt.Errorf("--timeout must be greater than zero"))
			}

			ctx, stopInterrupt := notifyContext(shared.CommandContext(cmd))
			defer stopInterrupt()
			if cmd.Flags().Changed("timeout") {
				var cancelTimeout context.CancelFunc
				ctx, cancelTimeout = context.WithTimeout(ctx, timeout)
				defer cancelTimeout()
			}
			report := runDoctor(ctx)
			contextErr := ctx.Err()
			jsonOut, err := cmd.Flags().GetBool("json")
			if err != nil {
				return err
			}
			if jsonOut {
				if err := shared.WriteJSON(cmd, newDoctorListItems(report)); err != nil {
					return err
				}
			} else {
				_, _ = cmd.OutOrStdout().Write([]byte(renderDoctorTable(report.Checks) + "\n"))
			}
			if contextErr != nil {
				return contextErr
			}
			if report.Failed() {

				return shared.WithExitCode(nil, 1)
			}
			return nil
		},
	}
	cmd.Flags().Bool("json", false, "Print diagnostics as JSON")
	cmd.Flags().Duration("timeout", 0, "Optional maximum time for the complete diagnostic run")
	return cmd
}

type doctorListItem struct {
	Profile   string `json:"profile"`
	ConfigDir string `json:"config_dir"`
	CacheDir  string `json:"cache_dir"`
	Offline   bool   `json:"offline"`
	core.DoctorCheck
}

func newDoctorListItems(report core.DoctorReport) []doctorListItem {
	items := make([]doctorListItem, len(report.Checks))
	for i, check := range report.Checks {
		items[i] = doctorListItem{
			Profile:     report.Profile,
			ConfigDir:   report.ConfigDir,
			CacheDir:    report.CacheDir,
			Offline:     report.Offline,
			DoctorCheck: check,
		}
	}
	return items
}
