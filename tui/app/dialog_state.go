package app

import "github.com/yumauri/fbrcm/core"

type dialogMode int

const (
	dialogModeEditChoice dialogMode = iota
	dialogModePublishDraft
	dialogModeDiscardDraft
)

// pendingDialog holds pending dialog state used by the app package.
type pendingDialog struct {
	// project stores project for pendingDialog.
	project core.Project
	// mode stores mode for pendingDialog.
	mode dialogMode
}
