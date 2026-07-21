package doctor

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/core"
)

type doctorFunc func(context.Context) core.DoctorReport
type notifyContextFunc func(context.Context) (context.Context, context.CancelFunc)

// New constructs the doctor command.
func New(svc *core.Core) *cobra.Command {
	return newCommand(svc.Doctor, func(parent context.Context) (context.Context, context.CancelFunc) {
		return signal.NotifyContext(parent, os.Interrupt)
	})
}

func newCommand(runDoctor doctorFunc, notifyContext notifyContextFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check profile, credentials, connectivity, APIs, permissions, and cache",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			timeout, err := cmd.Flags().GetDuration("timeout")
			if err != nil {
				return err
			}
			if cmd.Flags().Changed("timeout") && timeout <= 0 {
				return fmt.Errorf("--timeout must be greater than zero")
			}

			ctx, stopInterrupt := notifyContext(cmd.Context())
			defer stopInterrupt()
			if cmd.Flags().Changed("timeout") {
				var cancelTimeout context.CancelFunc
				ctx, cancelTimeout = context.WithTimeout(ctx, timeout)
				defer cancelTimeout()
			}
			report := runDoctor(ctx)
			interrupted := ctx.Err() != nil
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
			if report.Failed() || interrupted {
				cmd.Root().SilenceErrors = true
				cmd.Root().SilenceUsage = true
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
