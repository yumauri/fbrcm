package mcp

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/yumauri/fbrcm/core"
	corelog "github.com/yumauri/fbrcm/core/log"
	"github.com/yumauri/fbrcm/internal/terminal/progress"
	"github.com/yumauri/fbrcm/ops/contract"
	"github.com/yumauri/fbrcm/ops/shared"
)

func Run(ctx context.Context, service *core.Core, version, commit, date string, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	progress.Configure(stderr, false)
	defer progress.Stop()
	root := rootCommand(service, version, commit, date)
	defer contract.UnregisterResponses(root)
	root.SetArgs(args)
	root.SetContext(shared.WithMachineState(ctx))
	root.SetIn(stdin)
	root.SetOut(stdout)
	root.SetErr(stderr)
	var captured bytes.Buffer
	jsonMode := contract.JSONRequested(args)
	if jsonMode {
		root.SetOut(&captured)
	}
	cmd, err := root.ExecuteC()
	if err == nil {
		err = ctx.Err()
	}
	if jsonMode {
		envelope := contract.BuildEnvelope(cmd, version, captured.Bytes(), err)
		if writeErr := contract.Write(stdout, envelope); writeErr != nil {
			_, _ = fmt.Fprintln(stderr, writeErr)
			return 13
		}
		return envelope.ExitCode
	}
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "error: %v\n", err)
		return contract.ExitCode(cmd, err)
	}
	return 0
}

func Init(service *core.Core, version, commit, date string) {
	corelog.ConfigureCLIOutput(os.Stderr, os.Stderr)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if code := Run(ctx, service, version, commit, date, os.Args[1:], os.Stdin, os.Stdout, os.Stderr); code != 0 {
		os.Exit(code)
	}
}
