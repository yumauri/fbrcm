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
	SelectConditionName    string
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

type ConditionsLoadedMsg struct {
	Project             core.Project
	Tree                *core.ConditionsTree
	Source              string
	SelectConditionName string
	Err                 error
}

type HistoryLoadedMsg struct {
	Project                             core.Project
	PreviousTree, CurrentTree           *core.ParametersTree
	PreviousVersion, CurrentVersion     string
	PreviousPublished, CurrentPublished string
	Versions                            []core.RemoteConfigVersionEntry
	Unavailable                         bool
	Err                                 error
}

type HistoryRollbackRequestedMsg struct {
	Project                 core.Project
	Target                  core.RemoteConfigVersionEntry
	PickerLeft              bool
	LeftCursor, RightCursor int
}

type HistoryRollbackPreviewLoadedMsg struct {
	Project         core.Project
	Target, Current *core.ResolvedRemoteConfigVersion
	Diff            string
	Changed         bool
	Err             error
}

type HistoryRollbackConfirmedMsg struct{}
type HistoryRollbackCanceledMsg struct{}

type HistoryRollbackCompletedMsg struct {
	Project core.Project
	Result  core.VersionPublishResult
	Tree    *core.ParametersTree
	Err     error
}

type ParameterViewData struct {
	Project          core.Project
	GroupKey         string
	GroupLabel       string
	Groups           []ParameterGroupOption
	ParameterKeys    []string
	Conditions       []core.ParametersCondition
	Parameter        core.ParametersEntry
	SelectedValueIdx int
}

type GroupViewData struct {
	Project    core.Project
	Group      core.ParametersGroup
	GroupNames []string
}

type ParameterGroupOption struct {
	Key   string
	Label string
}

type ParameterSelectionChangedMsg struct {
	Data        *ParameterViewData
	GroupData   *GroupViewData
	Activate    bool
	ResetScroll bool
}

type ConditionViewData struct {
	Project        core.Project
	Condition      core.ConditionEntry
	ConditionNames []string
}

type ConditionSelectionChangedMsg struct {
	Data        *ConditionViewData
	Activate    bool
	ResetScroll bool
}

type KeyboardCaptureMsg struct {
	Enabled bool
}

type SetActivePanelMsg struct {
	Panel              panels.ID
	ResetParametersTab bool
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

type DetailsAddConditionalValueRequestedMsg struct{}
