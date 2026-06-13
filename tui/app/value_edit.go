package app

import (
	"fmt"

	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/tui/messages"
)

type previewValueEditFunc func() (*core.ParametersCache, []byte, error)

func (m Model) valueEditDialogBody(project core.Project, preview previewValueEditFunc) ([]string, error) {
	cache, finalRaw, err := preview()
	if err != nil || cache == nil {
		if err == nil {
			err = fmt.Errorf("parameters cache not found")
		}
		return nil, err
	}

	currentCfg, err := firebase.ParseRemoteConfig(cache.RemoteConfig)
	if err != nil {
		return nil, err
	}
	finalCfg, err := firebase.ParseRemoteConfig(finalRaw)
	if err != nil {
		return nil, err
	}

	diffText, hasChanges := shared.RenderRemoteConfigDiff(currentCfg, finalCfg)
	if !hasChanges {
		return nil, fmt.Errorf("parameter value not changed")
	}

	lines := []string{
		"Project: " + dialogProjectNameStyle.Render(project.Name) + " (" + project.ProjectID + ")",
		"",
		"Edit value or draft changes?",
		"",
	}
	lines = append(lines, dialogDiffLines(diffText)...)
	return lines, nil
}

func (m Model) valueEditLoadedMsg(project core.Project, groupKey, paramKey string, tree *core.ParametersTree, hasDraft, stale, publish bool) messages.ParametersLoadedMsg {
	source := "draft"
	if publish {
		source = "firebase"
	}
	return messages.ParametersLoadedMsg{
		Project:        project,
		Tree:           tree,
		Source:         source,
		CacheSource:    "cache",
		Err:            nil,
		HasDraft:       hasDraft,
		StaleDraft:     !publish && hasDraft && stale,
		Revalidate:     false,
		SelectGroupKey: groupKey,
		SelectParamKey: paramKey,
	}
}
