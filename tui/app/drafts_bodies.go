package app

import (
	"encoding/json"
	"fmt"

	"github.com/yumauri/fbrcm/core"
	rcdiff "github.com/yumauri/fbrcm/core/rc/diff"
)

type previewDialogFunc func() (*core.ParametersCache, json.RawMessage, error)
type currentRawFunc func(*core.ParametersCache) (json.RawMessage, error)

func (m Model) deleteDialogBody(project core.Project, groupKey, paramKey string) ([]string, bool) {
	cache, finalRaw, err := m.svc.PreviewDeleteParameter(project.ProjectID, groupKey, paramKey)
	if err != nil || cache == nil {
		return nil, false
	}

	currentCfg, finalCfg, err := parseRemoteConfigPair(cache.RemoteConfig, finalRaw)
	if err != nil {
		return nil, false
	}

	diffText, hasChanges := rcdiff.RenderRemoteConfigDiff(currentCfg, finalCfg)
	lines := []string{
		dialogProjectLine(project),
		"",
		"Delete parameter or draft changes?",
	}
	if !hasChanges {
		lines = append(lines, "", "Parameter: "+dialogParameterNameStyle.Render(paramKey))
		return lines, true
	}

	lines = append(lines, "")
	lines = append(lines, dialogDiffLines(diffText)...)
	return lines, true
}

func (m Model) deleteGroupDialogBody(project core.Project, groupKey string) ([]string, bool) {
	cache, finalRaw, err := m.svc.PreviewDeleteGroup(project.ProjectID, groupKey)
	if err != nil || cache == nil {
		return nil, false
	}

	currentCfg, finalCfg, err := parseRemoteConfigPair(cache.RemoteConfig, finalRaw)
	if err != nil {
		return nil, false
	}

	diffText, hasChanges := rcdiff.RenderRemoteConfigDiff(currentCfg, finalCfg)
	lines := []string{
		dialogProjectLine(project),
		"",
		"Delete group or draft changes?",
	}
	if !hasChanges {
		return lines, true
	}
	lines = append(lines, "")
	lines = append(lines, dialogDiffLines(diffText)...)
	return lines, true
}

func (m Model) deleteConditionalValueDialogBody(project core.Project, groupKey, paramKey, valueLabel string) ([]string, error) {
	cache, finalRaw, err := m.svc.PreviewDeleteConditionalValue(project.ProjectID, groupKey, paramKey, valueLabel)
	if err != nil || cache == nil {
		if err == nil {
			err = fmt.Errorf("parameters cache not found")
		}
		return nil, err
	}

	currentCfg, finalCfg, err := parseRemoteConfigPair(cache.RemoteConfig, finalRaw)
	if err != nil {
		return nil, err
	}

	diffText, hasChanges := rcdiff.RenderRemoteConfigDiff(currentCfg, finalCfg)
	lines := []string{
		dialogProjectLine(project),
		"",
		"Delete conditional value or draft changes?",
	}
	if !hasChanges {
		return nil, fmt.Errorf("conditional value not changed")
	}

	lines = append(lines, "")
	lines = append(lines, dialogDiffLines(diffText)...)
	return lines, nil
}

func (m Model) draftDialogBody(project core.Project, mode dialogMode) ([]string, bool) {
	cache, _, err := m.svc.InspectParametersCache(project.ProjectID)
	if err != nil || cache == nil {
		return nil, false
	}
	draftRaw, hasDraft, err := m.svc.LoadDraft(project.ProjectID)
	if err != nil || !hasDraft {
		return nil, false
	}

	currentCfg, draftCfg, err := parseRemoteConfigPair(cache.RemoteConfig, draftRaw)
	if err != nil {
		return nil, false
	}
	diffText, hasChanges := rcdiff.RenderRemoteConfigDiff(currentCfg, draftCfg)
	if !hasChanges {
		diffText = "\nNo changes.\n"
	}

	lines := []string{
		dialogProjectLine(project),
		"",
	}
	if mode == dialogModePublishDraft {
		lines = append(lines, "Publish draft changes?")
	} else {
		lines = append(lines, "Discard draft changes?")
	}
	lines = append(lines, "")
	lines = append(lines, dialogDiffLines(diffText)...)
	return lines, true
}

func (m Model) previewDialogBody(project core.Project, prompt, unchangedErr string, preview previewDialogFunc, currentRaw currentRawFunc) ([]string, error) {
	cache, finalRaw, err := preview()
	if err != nil || cache == nil {
		if err == nil {
			err = fmt.Errorf("parameters cache not found")
		}
		return nil, err
	}

	raw := cache.RemoteConfig
	if currentRaw != nil {
		raw, err = currentRaw(cache)
		if err != nil {
			return nil, err
		}
	}
	currentCfg, finalCfg, err := parseRemoteConfigPair(raw, finalRaw)
	if err != nil {
		return nil, err
	}

	diffText, hasChanges := rcdiff.RenderRemoteConfigDiff(currentCfg, finalCfg)
	if !hasChanges {
		return nil, fmt.Errorf("%s", unchangedErr)
	}

	lines := []string{
		dialogProjectLine(project),
		"",
		prompt,
		"",
	}
	lines = append(lines, dialogDiffLines(diffText)...)
	return lines, nil
}

func (m Model) renameDialogBody(project core.Project, groupKey, paramKey, nextParamKey string) ([]string, error) {
	return m.previewDialogBody(project, "Rename parameter or draft changes?", "parameter not changed", func() (*core.ParametersCache, json.RawMessage, error) {
		return m.svc.PreviewRenameParameter(project.ProjectID, groupKey, paramKey, nextParamKey)
	}, nil)
}

func (m Model) renameGroupDialogBody(project core.Project, groupKey, nextGroupKey string) ([]string, error) {
	return m.previewDialogBody(project, "Rename group or draft changes?", "group not changed", func() (*core.ParametersCache, json.RawMessage, error) {
		return m.svc.PreviewRenameGroup(project.ProjectID, groupKey, nextGroupKey)
	}, nil)
}

func (m Model) moveDialogBody(project core.Project, groupKey, paramKey, nextGroupKey string) ([]string, error) {
	return m.previewDialogBody(project, "Move parameter or draft changes?", "parameter not changed", func() (*core.ParametersCache, json.RawMessage, error) {
		return m.svc.PreviewMoveParameter(project.ProjectID, groupKey, paramKey, nextGroupKey)
	}, nil)
}

func (m Model) moveGroupDialogBody(project core.Project, groupKey, nextGroupKey string) ([]string, error) {
	return m.previewDialogBody(project, "Move group or draft changes?", "group not changed", func() (*core.ParametersCache, json.RawMessage, error) {
		return m.svc.PreviewMoveGroup(project.ProjectID, groupKey, nextGroupKey)
	}, nil)
}

func (m Model) duplicateDialogBody(project core.Project, groupKey, paramKey, nextParamKey string) ([]string, error) {
	return m.previewDialogBody(project, "Duplicate parameter or draft changes?", "parameter not changed", func() (*core.ParametersCache, json.RawMessage, error) {
		return m.svc.PreviewDuplicateParameter(project.ProjectID, groupKey, paramKey, nextParamKey)
	}, nil)
}

func (m Model) editDetailsDialogBody(project core.Project, edit core.ParameterDetailsEdit) ([]string, error) {
	return m.previewDialogBody(project, "Edit parameter or draft changes?", "parameter not changed", func() (*core.ParametersCache, json.RawMessage, error) {
		return m.svc.PreviewEditParameterDetails(project.ProjectID, edit)
	}, func(cache *core.ParametersCache) (json.RawMessage, error) {
		if draftRaw, hasDraft, err := m.svc.LoadDraft(project.ProjectID); err != nil {
			return nil, err
		} else if hasDraft {
			return draftRaw, nil
		}
		return cache.RemoteConfig, nil
	})
}

func (m Model) editGroupDetailsDialogBody(project core.Project, edit core.GroupDetailsEdit) ([]string, error) {
	return m.previewDialogBody(project, "Edit group or draft changes?", "group not changed", func() (*core.ParametersCache, json.RawMessage, error) {
		return m.svc.PreviewEditGroupDetails(project.ProjectID, edit)
	}, func(cache *core.ParametersCache) (json.RawMessage, error) {
		if draftRaw, hasDraft, err := m.svc.LoadDraft(project.ProjectID); err != nil {
			return nil, err
		} else if hasDraft {
			return draftRaw, nil
		}
		return cache.RemoteConfig, nil
	})
}
