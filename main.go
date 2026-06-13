package main

import (
	"context"
	"fmt"
	"os"

	"github.com/yumauri/fbrcm/cli"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/firebase"
	corelog "github.com/yumauri/fbrcm/core/log"
	"github.com/yumauri/fbrcm/tui"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// main handles main and returns the resulting value or error.
func main() {
	mode := corelog.ModeCLI
	if len(os.Args) == 1 {
		mode = corelog.ModeTUI
	}
	corelog.Init(mode)
	firebase.InitOfflineMode()

	svc, err := core.NewService(context.Background())
	if err != nil {
		corelog.For("main").Error("application initialization failed", "err", err)
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	corelog.For("main").Debug("application start", "mode", mode, "arg_count", len(os.Args)-1)
	if mode == corelog.ModeTUI {
		if err := config.EnsureActiveProfile(); err != nil {
			corelog.For("main").Error("application initialization failed", "err", err)
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		tui.Init(svc)
	} else {
		cli.Init(svc, version, commit, date)
	}
}
