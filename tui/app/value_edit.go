package app

import (
	"context"
	"fmt"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	rcdiff "github.com/yumauri/fbrcm/core/rc/diff"
	dialogcmp "github.com/yumauri/fbrcm/tui/components/dialog"
	"github.com/yumauri/fbrcm/tui/messages"
)

type previewValueEditFunc func() (*core.ParametersCache, []byte, error)
type valueDialogBodyFunc func() ([]string, error)
type setParameterValueFunc func(context.Context) (*core.ParametersTree, bool, error)

func (m Model) valueEditDialogBody(project core.Project, preview previewValueEditFunc) ([]string, error) {
	cache, finalRaw, err := preview()
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

func (m *Model) openValueEditDialog(project core.Project, bodyFn valueDialogBodyFunc, logErr func(error), applyCmd, draftCmd tea.Cmd) {
	body, err := bodyFn()
	if err != nil {
		logErr(err)
		m.openErrorDialog("Edit Value Failed", project, err.Error())
		return
	}
	m.dialog = m.dialog.Open(dialogcmp.Config{
		Title: "Edit Value?",
		Body:  body,
		Buttons: []dialogcmp.Button{
			{Label: "Apply", Variant: dialogcmp.ButtonVariantDanger, OnPress: applyCmd},
			{Label: "Draft", Variant: dialogcmp.ButtonVariantAccent, OnPress: draftCmd},
			{Label: "Cancel", Variant: dialogcmp.ButtonVariantAccent, OnPress: dialogCanceledCmd()},
		},
	})
}

func (m Model) runSetParameterValueCmd(project core.Project, groupKey, paramKey, valueLabel string, publish bool, set setParameterValueFunc) tea.Cmd {
	return func() tea.Msg {
		_, stale := m.parameters.ProjectDraftState(project.ProjectID)
		tree, hasDraft, err := set(context.Background())
		if err != nil {
			return messages.ParametersLoadedMsg{Project: project, Err: err, HasDraft: m.parameters.HasDraft(project.ProjectID), StaleDraft: stale}
		}
		return m.valueEditLoadedMsg(project, groupKey, paramKey, tree, hasDraft, stale, publish)
	}
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
