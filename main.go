package main

import (
	"context"
	"fmt"
	"os"

	charmlog "charm.land/log/v2"

	"github.com/yumauri/fbrcm/cli"
	"github.com/yumauri/fbrcm/cli/contract"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
	"github.com/yumauri/fbrcm/core/firebase"
	corelog "github.com/yumauri/fbrcm/core/log"
	"github.com/yumauri/fbrcm/tui"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	mode := corelog.ModeCLI
	if len(os.Args) == 1 {
		mode = corelog.ModeTUI
	}
	corelog.InitWithDefault(mode, defaultLogLevel(mode, os.Args[1:]))
	if mode == corelog.ModeTUI {
		firebase.InitOfflineMode()
	}

	svc, err := core.NewService(context.Background())
	if err != nil {
		corelog.For("main").Error("application initialization failed", "err", err)
		if mode == corelog.ModeCLI && contract.JSONRequested(os.Args[1:]) {
			envelope := contract.BuildEnvelope(nil, version, nil, err)
			if writeErr := contract.Write(os.Stdout, envelope); writeErr != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", writeErr)
				os.Exit(13)
			}
			os.Exit(envelope.ExitCode)
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(applicationInitializationExitCode(mode, err))
	}

	corelog.For("main").Debug("application start", "mode", mode, "arg_count", len(os.Args)-1)
	if mode == corelog.ModeTUI {
		if err := config.EnsureActiveProfile(); err != nil {
			corelog.For("main").Error("application initialization failed", "err", err)
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		tui.Init(svc, version, commit, date)
	} else {
		cli.Init(svc, version, commit, date)
	}
}

func applicationInitializationExitCode(mode corelog.Mode, err error) int {
	if mode == corelog.ModeCLI {
		return contract.ExitCode(nil, err)
	}
	return 1
}

func defaultLogLevel(mode corelog.Mode, args []string) charmlog.Level {
	if mode == corelog.ModeCLI && contract.JSONRequested(args) {
		if _, overridden := env.LookupTrimmed(env.LogLevel); !overridden {
			return corelog.SilentLevel
		}
	}
	return charmlog.InfoLevel
}
