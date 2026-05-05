package tui

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"fbrcm/core"
	corelog "fbrcm/core/log"
	"fbrcm/tui/app"
)

func Init(s *core.Core) {
	corelog.For("tui").Debug("start tui")
	m := app.New(s)
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		corelog.For("tui").Error("tui exited with error", "err", err)
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
