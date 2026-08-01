package tui

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/env"
	corelog "github.com/yumauri/fbrcm/core/log"
	"github.com/yumauri/fbrcm/tui/app"
	tuiconfig "github.com/yumauri/fbrcm/tui/config"
)

func Init(s *core.Core) {
	corelog.For("tui").Debug("start tui")
	if _, err := tuiconfig.Load(); err != nil {
		corelog.For("tui").Error("tui config load failed", "err", err)
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	m := app.New(s)
	p := tea.NewProgram(m, programOptions()...)
	if _, err := p.Run(); err != nil {
		corelog.For("tui").Error("tui exited with error", "err", err)
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func programOptions() []tea.ProgramOption {
	if !env.NoColorEnabled() {
		return nil
	}
	return []tea.ProgramOption{tea.WithColorProfile(colorprofile.ASCII)}
}
