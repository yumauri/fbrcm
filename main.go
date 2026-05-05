package main

import (
	"context"
	"fmt"
	"os"

	"fbrcm/cli"
	"fbrcm/core"
	corelog "fbrcm/core/log"
	"fbrcm/tui"
)

func main() {
	mode := corelog.ModeCLI
	if len(os.Args) == 1 {
		mode = corelog.ModeTUI
	}
	corelog.Init(mode)

	svc, err := core.NewService(context.Background())
	if err != nil {
		corelog.For("main").Error("application initialization failed", "err", err)
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	corelog.For("main").Debug("application start", "mode", mode, "arg_count", len(os.Args)-1)
	if mode == corelog.ModeTUI {
		tui.Init(svc)
	} else {
		cli.Init(svc)
	}
}
