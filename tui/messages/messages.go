package messages

import (
	"fbrcm/core"
	"fbrcm/tui/panels"
)

type ProjectsLoadedMsg struct {
	Projects []core.Project
	Source   string
	Err      error
}

type ProjectsSelectionChangedMsg struct {
	Projects []core.Project
}

type ParametersLoadedMsg struct {
	Project    core.Project
	Tree       *core.ParametersTree
	Source     string
	Err        error
	Revalidate bool
}

type QuitMsg struct{}

type KeyboardCaptureMsg struct {
	Enabled bool
}

type SetActivePanelMsg struct {
	Panel panels.ID
}

type LogLineMsg struct {
	Line string
}
