package tui

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	corelog "github.com/yumauri/fbrcm/core/log"
	"github.com/yumauri/fbrcm/tui/app"
	tuiconfig "github.com/yumauri/fbrcm/tui/config"
)

// Init initializes init and returns the resulting value or error.
func Init(s *core.Core) {
	corelog.For("tui").Debug("start tui")
	if _, err := tuiconfig.Load(); err != nil {
		corelog.For("tui").Error("tui config load failed", "err", err)
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	m := app.New(s)
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		corelog.For("tui").Error("tui exited with error", "err", err)
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
