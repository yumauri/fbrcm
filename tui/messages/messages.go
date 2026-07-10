package messages

import (
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/tui/panels"
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
	Project                core.Project
	Tree                   *core.ParametersTree
	Source                 string
	CacheSource            string
	CacheVersion           string
	DraftVersion           string
	SelectGroupKey         string
	SelectParamKey         string
	TransientGroupKey      string
	TransientAfterParamKey string
	TransientParamKey      string
	TransientParamLabel    string
	CloseDetails           bool
	DetailsSaved           bool
	Err                    error
	Revalidate             bool
	RevalidateCache        *core.ParametersCache
	HasDraft               bool
	StaleDraft             bool
}

type ParameterViewData struct {
	Project          core.Project
	GroupKey         string
	GroupLabel       string
	Groups           []ParameterGroupOption
	ParameterKeys    []string
	Parameter        core.ParametersEntry
	SelectedValueIdx int
}

type ParameterGroupOption struct {
	Key   string
	Label string
}

type ParameterSelectionChangedMsg struct {
	Data        *ParameterViewData
	Activate    bool
	ResetScroll bool
}

type KeyboardCaptureMsg struct {
	Enabled bool
}

type SetActivePanelMsg struct {
	Panel panels.ID
}

type LogLineMsg struct {
	Line string
}

type DialogCanceledMsg struct{}

type DetailsEditCanceledMsg struct {
	CloseDetails bool
}

type DetailsInvalidFixMsg struct{}

type DetailsInvalidDiscardMsg struct {
	CloseDetails bool
}

type DetailsValueEditRequestedMsg struct{}
