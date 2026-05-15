package messages

import (
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/tui/panels"
)

// ProjectsLoadedMsg holds projects loaded msg state used by the messages package.
type ProjectsLoadedMsg struct {
	// Projects stores projects for ProjectsLoadedMsg.
	Projects []core.Project
	// Source stores source for ProjectsLoadedMsg.
	Source string
	// Err stores err for ProjectsLoadedMsg.
	Err error
}

// ProjectsSelectionChangedMsg holds projects selection changed msg state used by the messages package.
type ProjectsSelectionChangedMsg struct {
	// Projects stores projects for ProjectsSelectionChangedMsg.
	Projects []core.Project
}

// ParametersLoadedMsg holds parameters loaded msg state used by the messages package.
type ParametersLoadedMsg struct {
	// Project stores project for ParametersLoadedMsg.
	Project core.Project
	// Tree stores tree for ParametersLoadedMsg.
	Tree *core.ParametersTree
	// Source stores source for ParametersLoadedMsg.
	Source string
	// CacheSource stores cache source for ParametersLoadedMsg.
	CacheSource string
	// CacheVersion stores cache version for ParametersLoadedMsg.
	CacheVersion string
	// DraftVersion stores draft version for ParametersLoadedMsg.
	DraftVersion string
	// SelectGroupKey stores select group key for ParametersLoadedMsg.
	SelectGroupKey string
	// SelectParamKey stores select param key for ParametersLoadedMsg.
	SelectParamKey string
	// TransientGroupKey stores transient group key for ParametersLoadedMsg.
	TransientGroupKey string
	// TransientAfterParamKey stores transient after param key for ParametersLoadedMsg.
	TransientAfterParamKey string
	// TransientParamKey stores transient param key for ParametersLoadedMsg.
	TransientParamKey string
	// TransientParamLabel stores transient param label for ParametersLoadedMsg.
	TransientParamLabel string
	// CloseDetails stores close details for ParametersLoadedMsg.
	CloseDetails bool
	// DetailsSaved stores details saved for ParametersLoadedMsg.
	DetailsSaved bool
	// Err stores err for ParametersLoadedMsg.
	Err error
	// Revalidate stores revalidate for ParametersLoadedMsg.
	Revalidate bool
	// HasDraft stores has draft for ParametersLoadedMsg.
	HasDraft bool
	// StaleDraft stores stale draft for ParametersLoadedMsg.
	StaleDraft bool
}

// ParameterViewData holds parameter view data state used by the messages package.
type ParameterViewData struct {
	// Project stores project for ParameterViewData.
	Project core.Project
	// GroupKey stores group key for ParameterViewData.
	GroupKey string
	// GroupLabel stores group label for ParameterViewData.
	GroupLabel string
	// Groups stores groups for ParameterViewData.
	Groups []ParameterGroupOption
	// ParameterKeys stores parameter keys for ParameterViewData.
	ParameterKeys []string
	// Parameter stores parameter for ParameterViewData.
	Parameter core.ParametersEntry
	// SelectedValueIdx stores selected value idx for ParameterViewData.
	SelectedValueIdx int
}

// ParameterGroupOption holds parameter group option state used by the messages package.
type ParameterGroupOption struct {
	// Key stores key for ParameterGroupOption.
	Key string
	// Label stores label for ParameterGroupOption.
	Label string
}

// ParameterSelectionChangedMsg holds parameter selection changed msg state used by the messages package.
type ParameterSelectionChangedMsg struct {
	// Data stores data for ParameterSelectionChangedMsg.
	Data *ParameterViewData
	// Activate stores activate for ParameterSelectionChangedMsg.
	Activate bool
}

// QuitMsg holds quit msg state used by the messages package.
type QuitMsg struct{}

// KeyboardCaptureMsg holds keyboard capture msg state used by the messages package.
type KeyboardCaptureMsg struct {
	// Enabled stores enabled for KeyboardCaptureMsg.
	Enabled bool
}

// SetActivePanelMsg holds set active panel msg state used by the messages package.
type SetActivePanelMsg struct {
	// Panel stores panel for SetActivePanelMsg.
	Panel panels.ID
}

// LogLineMsg holds log line msg state used by the messages package.
type LogLineMsg struct {
	// Line stores line for LogLineMsg.
	Line string
}

// DialogCanceledMsg holds dialog canceled msg state used by the messages package.
type DialogCanceledMsg struct{}

// DetailsEditCanceledMsg holds details edit canceled msg state used by the messages package.
type DetailsEditCanceledMsg struct {
	// CloseDetails stores close details for DetailsEditCanceledMsg.
	CloseDetails bool
}

// DetailsInvalidFixMsg holds details invalid fix msg state used by the messages package.
type DetailsInvalidFixMsg struct{}

// DetailsInvalidDiscardMsg holds details invalid discard msg state used by the messages package.
type DetailsInvalidDiscardMsg struct {
	// CloseDetails stores close details for DetailsInvalidDiscardMsg.
	CloseDetails bool
}

// DetailsValueEditRequestedMsg holds details value edit requested msg state used by the messages package.
type DetailsValueEditRequestedMsg struct{}
