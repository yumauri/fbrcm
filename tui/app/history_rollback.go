package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	corelog "github.com/yumauri/fbrcm/core/log"
	rcdiff "github.com/yumauri/fbrcm/core/rc/diff"
	dialogcmp "github.com/yumauri/fbrcm/tui/components/dialog"
	"github.com/yumauri/fbrcm/tui/messages"
)

type historyRollbackPhase int

const (
	historyRollbackPreparing historyRollbackPhase = iota
	historyRollbackConfirming
	historyRollbackPublishing
	historyRollbackFailed
)

type historyRollbackSession struct {
	request        messages.HistoryRollbackRequestedMsg
	phase          historyRollbackPhase
	currentVersion string
}

func (m Model) beginHistoryRollback(request messages.HistoryRollbackRequestedMsg) (Model, tea.Cmd, bool) {
	if m.historyRollback != nil {
		return m, nil, true
	}
	m.historyRollback = &historyRollbackSession{request: request, phase: historyRollbackPreparing}
	if m.parameters.HasDraft(request.Project.ProjectID) {
		m.openHistoryRollbackFailure("Rollback Unavailable", "This project has unpublished draft changes. Publish or discard the draft before rolling back a Remote Config version.")
		return m, nil, true
	}
	if m.svc == nil {
		m.openHistoryRollbackFailure("Rollback Unavailable", "Firebase service is unavailable.")
		return m, nil, true
	}
	m.openHistoryRollbackProgress("Preparing Rollback…", request.Project, request.Target.VersionNumber)
	return m, m.historyRollbackPreviewCmd(request), true
}

func (m Model) historyRollbackPreviewCmd(request messages.HistoryRollbackRequestedMsg) tea.Cmd {
	return func() tea.Msg {
		current, err := m.svc.GetRemoteConfigVersion(context.Background(), request.Project.ProjectID, "current", false)
		if err != nil {
			return messages.HistoryRollbackPreviewLoadedMsg{Project: request.Project, Err: err}
		}
		target, err := m.svc.GetRemoteConfigVersion(context.Background(), request.Project.ProjectID, request.Target.VersionNumber, false)
		if err != nil {
			return messages.HistoryRollbackPreviewLoadedMsg{Project: request.Project, Current: current, Err: err}
		}
		if current.Version.VersionNumber == target.Version.VersionNumber {
			return messages.HistoryRollbackPreviewLoadedMsg{Project: request.Project, Current: current, Target: target, Err: fmt.Errorf("version %s is already current", target.Version.VersionNumber)}
		}
		diffText, changed := rcdiff.RenderRemoteConfigDiff(current.Config, target.Config)
		return messages.HistoryRollbackPreviewLoadedMsg{Project: request.Project, Current: current, Target: target, Diff: diffText, Changed: changed}
	}
}

func (m Model) updateHistoryRollbackPreview(msg messages.HistoryRollbackPreviewLoadedMsg) (Model, tea.Cmd, bool) {
	session := m.historyRollback
	if session == nil || session.request.Project.ProjectID != msg.Project.ProjectID {
		return m, nil, true
	}
	if msg.Err != nil {
		corelog.For("tui.history").Error("rollback preview failed", "project_id", msg.Project.ProjectID, "version", session.request.Target.VersionNumber, "err", msg.Err)
		m.openHistoryRollbackFailure("Rollback Preview Failed", msg.Err.Error())
		return m, nil, true
	}
	if !msg.Changed {
		m.openHistoryRollbackFailure("Rollback Unavailable", "The selected version has no differences from the current Remote Config.")
		return m, nil, true
	}
	session.phase = historyRollbackConfirming
	session.currentVersion = msg.Current.Version.VersionNumber
	m.historyRollback = session
	m.openHistoryRollbackConfirmation(msg)
	return m, nil, true
}

func (m *Model) openHistoryRollbackConfirmation(msg messages.HistoryRollbackPreviewLoadedMsg) {
	target := msg.Target.Version
	body := []string{
		dialogProjectLine(msg.Project),
		"",
		"Current: " + rollbackVersionDescription(msg.Current.Version),
		"Target:  " + rollbackVersionDescription(target),
		"",
	}
	body = append(body, rollbackSummaryLines(msg.Diff)...)
	body = append(body,
		"",
		"Rollback publishes the selected historical template as a new Remote Config version.",
		"Existing version history remains available.",
	)
	m.dialog = m.dialog.Open(dialogcmp.Config{
		Title: "Rollback to v" + target.VersionNumber + "?",
		Body:  body,
		Buttons: []dialogcmp.Button{
			{Label: "Rollback", Variant: dialogcmp.ButtonVariantDanger, OnPress: historyRollbackConfirmedCmd()},
			{Label: "Back", Variant: dialogcmp.ButtonVariantAccent, OnPress: historyRollbackCanceledCmd()},
		},
	})
}

func (m Model) confirmHistoryRollback() (Model, tea.Cmd, bool) {
	session := m.historyRollback
	if session == nil || session.phase != historyRollbackConfirming {
		return m, nil, true
	}
	session.phase = historyRollbackPublishing
	m.historyRollback = session
	m.openHistoryRollbackProgress("Rolling Back…", session.request.Project, session.request.Target.VersionNumber)
	return m, m.historyRollbackCmd(*session), true
}

func (m Model) historyRollbackCmd(session historyRollbackSession) tea.Cmd {
	return func() tea.Msg {
		project := session.request.Project
		latest, err := m.svc.GetRemoteConfigVersion(context.Background(), project.ProjectID, "current", false)
		if err != nil {
			return messages.HistoryRollbackCompletedMsg{Project: project, Err: err}
		}
		if latest.Version.VersionNumber != session.currentVersion {
			return messages.HistoryRollbackCompletedMsg{Project: project, Err: fmt.Errorf("current Remote Config changed from v%s to v%s during preview; choose the rollback version again", session.currentVersion, latest.Version.VersionNumber)}
		}
		result, err := m.svc.RollbackRemoteConfig(context.Background(), project.ProjectID, session.request.Target.VersionNumber)
		var tree *core.ParametersTree
		if len(result.RemoteConfig) > 0 {
			cache := &core.ParametersCache{ETag: result.ETag, CachedAt: time.Now().UTC(), RemoteConfig: result.RemoteConfig}
			tree, _ = m.svc.BuildParametersTree(cache)
		}
		return messages.HistoryRollbackCompletedMsg{Project: project, Result: result, Tree: tree, Err: err}
	}
}

func (m Model) updateHistoryRollbackCompleted(msg messages.HistoryRollbackCompletedMsg) (Model, tea.Cmd, bool) {
	session := m.historyRollback
	if session == nil || session.request.Project.ProjectID != msg.Project.ProjectID {
		return m, nil, true
	}
	if msg.Err != nil && msg.Result.PublishedVersion == "" {
		corelog.For("tui.history").Error("rollback failed", "project_id", msg.Project.ProjectID, "version", session.request.Target.VersionNumber, "err", msg.Err)
		m.openHistoryRollbackFailure("Rollback Failed", msg.Err.Error())
		return m, nil, true
	}

	m.historyRollback = nil
	m.dialog = m.dialog.Close()
	title := "Rollback Complete"
	tone := dialogcmp.ToneSuccess
	body := []string{
		dialogProjectLine(msg.Project),
		"",
		fmt.Sprintf("Published v%s using v%s; previous current version was v%s.", msg.Result.PublishedVersion, msg.Result.SourceVersion, msg.Result.PreviousVersion),
	}
	if msg.Err != nil {
		title = "Rollback Completed with Warning"
		tone = dialogcmp.ToneDefault
		body = append(body, "", msg.Err.Error())
	}
	m.dialog = m.dialog.Open(dialogcmp.Config{Title: title, Body: body, Buttons: []dialogcmp.Button{{Label: "Close", Variant: dialogcmp.ButtonVariantAccent, OnPress: dialogCanceledCmd()}}, Tone: tone})
	m.parameters = m.parameters.PreferNextHistoryPair(msg.Project.ProjectID, msg.Result.SourceVersion, msg.Result.PublishedVersion)

	var loadErr error
	if msg.Tree == nil {
		loadErr = msg.Err
		if loadErr == nil {
			loadErr = fmt.Errorf("rollback succeeded, but the published Remote Config could not be rendered")
		}
	}
	loaded := messages.ParametersLoadedMsg{
		Project: msg.Project, Tree: msg.Tree, Source: "firebase", CacheSource: "firebase",
		CacheVersion: msg.Result.PublishedVersion, HasDraft: false, Err: loadErr,
	}
	return m, func() tea.Msg { return loaded }, true
}

func (m *Model) openHistoryRollbackProgress(title string, project core.Project, version string) {
	m.dialog = m.dialog.Open(dialogcmp.Config{Title: title, Body: []string{
		dialogProjectLine(project),
		"",
		"Target version: v" + version,
		"",
		"Please wait…",
	}})
}

func (m *Model) openHistoryRollbackFailure(title, text string) {
	if m.historyRollback != nil {
		m.historyRollback.phase = historyRollbackFailed
	}
	project := core.Project{}
	if m.historyRollback != nil {
		project = m.historyRollback.request.Project
	}
	m.dialog = m.dialog.Open(dialogcmp.Config{Title: title, Body: []string{
		dialogProjectLine(project),
		"",
		text,
	}, Buttons: []dialogcmp.Button{{Label: "Back", Variant: dialogcmp.ButtonVariantAccent, OnPress: historyRollbackCanceledCmd()}}})
}

func (m *Model) cancelHistoryRollback() {
	session := m.historyRollback
	m.dialog = m.dialog.Close()
	m.historyRollback = nil
	if session == nil {
		return
	}
	request := session.request
	m.parameters = m.parameters.RestoreHistoryPicker(request.Project.ProjectID, request.PickerLeft, request.LeftCursor, request.RightCursor)
}

func (m Model) historyRollbackModalLocked() bool {
	return m.historyRollback != nil && (m.historyRollback.phase == historyRollbackPreparing || m.historyRollback.phase == historyRollbackPublishing)
}

func historyRollbackConfirmedCmd() tea.Cmd {
	return func() tea.Msg { return messages.HistoryRollbackConfirmedMsg{} }
}

func historyRollbackCanceledCmd() tea.Cmd {
	return func() tea.Msg { return messages.HistoryRollbackCanceledMsg{} }
}

func rollbackVersionDescription(version firebase.RemoteConfigVersion) string {
	parts := []string{"v" + version.VersionNumber}
	if published := rollbackPublished(version.UpdateTime); published != "" {
		parts = append(parts, published)
	}
	author := strings.TrimSpace(version.UpdateUser.Email)
	if author == "" {
		author = strings.TrimSpace(version.UpdateUser.Name)
	}
	if author != "" {
		parts = append(parts, author)
	}
	return strings.Join(parts, "  ")
}

func rollbackPublished(raw string) string {
	parsed, err := time.Parse(time.RFC3339Nano, raw)
	if err != nil {
		return raw
	}
	return parsed.Local().Format("2006-01-02 15:04:05")
}

func rollbackSummaryLines(diffText string) []string {
	const marker = "\n\nSummary:\n"
	if _, summary, ok := strings.Cut(diffText, marker); ok {
		rawLines := strings.Split(strings.TrimSpace(summary), "\n")
		lines := make([]string, 0, len(rawLines)+1)
		lines = append(lines, "Summary:")
		for _, line := range rawLines {
			if line = strings.TrimSpace(line); line != "" {
				lines = append(lines, line)
			}
		}
		return lines
	}
	return []string{"Changes detected."}
}
