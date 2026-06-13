package app

import "github.com/yumauri/fbrcm/core"

type dialogMode int

const (
	dialogModeEditChoice dialogMode = iota
	dialogModePublishDraft
	dialogModeDiscardDraft
)

type pendingDialog struct {
	project core.Project
	mode    dialogMode
}
