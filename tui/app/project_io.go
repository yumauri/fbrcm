package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/cli/shared/fileoutput"
	"github.com/yumauri/fbrcm/cli/shared/rc"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	corelog "github.com/yumauri/fbrcm/core/log"
	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
	dialogcmp "github.com/yumauri/fbrcm/tui/components/dialog"
	"github.com/yumauri/fbrcm/tui/components/projectio"
	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/panels"
)

type projectImportPlanLoadedMsg struct {
	plan *core.ProjectImportPlan
	err  error
}

type projectImportCompletedMsg struct {
	plan   *core.ProjectImportPlan
	result *core.ProjectImportResult
	err    error
}

type projectExportSession struct {
	project core.Project
	path    string
	draft   bool
}

type projectExportCompletedMsg struct {
	session projectExportSession
	err     error
}

type projectExportOverwriteMsg struct{}
type projectExportBackMsg struct{}

type projectDefaultsSession struct {
	project   core.Project
	path      string
	format    firebase.DefaultsFormat
	overwrite bool
}

type projectDefaultsCompletedMsg struct {
	session projectDefaultsSession
	err     error
}

type projectDefaultsOverwriteMsg struct{}
type projectDefaultsBackMsg struct{}

func (m Model) openProjectImport() (Model, tea.Cmd, bool) {
	project, ok := m.projects.CurrentProject()
	if !ok {
		m.openErrorDialog("Import Unavailable", core.Project{}, "No project is selected under the Projects panel cursor.")
		return m, nil, true
	}
	if project.Disabled {
		return m, nil, true
	}
	if m.svc == nil {
		m.openErrorDialog("Import Unavailable", project, "Firebase service is unavailable.")
		return m, nil, true
	}
	var cmd tea.Cmd
	m.projectIO, cmd = m.projectIO.OpenImport(project)
	return m, cmd, true
}

func (m Model) openProjectExport() (Model, tea.Cmd, bool) {
	project, ok := m.projects.CurrentProject()
	if !ok {
		m.openErrorDialog("Export Unavailable", core.Project{}, "No project is selected under the Projects panel cursor.")
		return m, nil, true
	}
	if project.Disabled {
		return m, nil, true
	}
	if m.svc == nil {
		m.openErrorDialog("Export Unavailable", project, "Firebase service is unavailable.")
		return m, nil, true
	}
	hasDraft, err := m.svc.HasDraft(project.ProjectID)
	if err != nil {
		m.openErrorDialog("Export Failed", project, err.Error())
		return m, nil, true
	}
	var cmd tea.Cmd
	m.projectIO, cmd = m.projectIO.OpenExport(project, hasDraft)
	return m, cmd, true
}

func (m Model) openProjectDefaults() (Model, tea.Cmd, bool) {
	project, ok := m.projects.CurrentProject()
	if !ok {
		m.openErrorDialog("Defaults Unavailable", core.Project{}, "No project is selected under the Projects panel cursor.")
		return m, nil, true
	}
	if project.Disabled {
		return m, nil, true
	}
	if m.svc == nil {
		m.openErrorDialog("Defaults Unavailable", project, "Firebase service is unavailable.")
		return m, nil, true
	}
	var cmd tea.Cmd
	m.projectIO, cmd = m.projectIO.OpenDefaults(project)
	return m, cmd, true
}

func (m Model) prepareProjectImportCmd(request projectio.ImportPlanRequestedMsg) tea.Cmd {
	return func() tea.Msg {
		plan, err := m.svc.PrepareProjectImport(context.Background(), request.Project, request.Raw, request.Options)
		if plan != nil {
			plan.SourcePath = request.Path
		}
		return projectImportPlanLoadedMsg{plan: plan, err: err}
	}
}

func (m Model) updateProjectImportPlan(msg projectImportPlanLoadedMsg) (Model, tea.Cmd, bool) {
	if !m.projectIO.IsOpen() {
		return m, nil, true
	}
	if msg.err != nil {
		m.projectIO = m.projectIO.SetError(msg.err)
		return m, nil, true
	}
	if msg.plan == nil {
		m.projectIO = m.projectIO.SetError(fmt.Errorf("import plan is empty"))
		return m, nil, true
	}
	if len(msg.plan.Conflicts) > 0 && !m.projectIO.ConflictsOpen() {
		m.projectIO = m.projectIO.OpenConflicts(msg.plan.Conflicts)
		return m, nil, true
	}
	m.projectIO = m.projectIO.Close()
	if !msg.plan.HasChanges {
		m.openProjectImportNoChanges(msg.plan)
		return m, nil, true
	}
	m.projectImport = msg.plan
	m.openProjectImportReview(msg.plan)
	return m, nil, true
}

func (m *Model) openProjectImportNoChanges(plan *core.ProjectImportPlan) {
	m.dialog = m.dialog.Open(dialogcmp.Config{
		Title: "Import Has No Changes",
		Body: []string{
			dialogProjectLine(plan.Project),
			"",
			"The selected import produces no Remote Config changes.",
		},
		Buttons: []dialogcmp.Button{{Label: "Close", Variant: dialogcmp.ButtonVariantAccent, OnPress: dialogCanceledCmd()}},
	})
}

func (m *Model) openProjectImportReview(plan *core.ProjectImportPlan) {
	strategy := "Merge into current config"
	if plan.Options.Strategy == core.ProjectImportReplace {
		strategy = "Replace entire config"
	}
	body := []string{
		dialogProjectLine(plan.Project),
		"",
		"File: " + plan.SourcePath,
		"Source: " + rcdisplay.FormatCount(plan.Summary.Parameters(), "parameter", "parameters") +
			" · " + rcdisplay.FormatCount(plan.Summary.Groups, "group", "groups") +
			" · " + rcdisplay.FormatCount(plan.Summary.Conditions, "condition", "conditions"),
		"Strategy: " + strategy,
		fmt.Sprintf("Conflicts: %d", len(plan.Conflicts)),
		"",
		"Review Remote Config changes:",
		"",
	}
	body = append(body, dialogDiffLines(plan.Diff)...)
	buttons := []dialogcmp.Button{
		{Label: "Save Draft", Variant: dialogcmp.ButtonVariantAccent, OnPress: m.executeProjectImportCmd(false)},
		{Label: "Publish Now", Variant: dialogcmp.ButtonVariantDanger, OnPress: m.executeProjectImportCmd(true)},
		{Label: "Cancel", Variant: dialogcmp.ButtonVariantAccent, OnPress: dialogCanceledCmd()},
	}
	if plan.HasDraft {
		body = append(body, "", "An unpublished draft exists. This import will update that draft.")
		buttons = []dialogcmp.Button{
			{Label: "Update Draft", Variant: dialogcmp.ButtonVariantAccent, OnPress: m.executeProjectImportCmd(false)},
			{Label: "Cancel", Variant: dialogcmp.ButtonVariantAccent, OnPress: dialogCanceledCmd()},
		}
	}
	m.dialog = m.dialog.Open(dialogcmp.Config{Title: "Import Remote Config?", Body: body, Buttons: buttons})
}

func (m Model) executeProjectImportCmd(publish bool) tea.Cmd {
	plan := m.projectImport
	return func() tea.Msg {
		result, err := m.svc.ExecuteProjectImport(context.Background(), plan, publish)
		return projectImportCompletedMsg{plan: plan, result: result, err: err}
	}
}

func (m Model) updateProjectImportCompleted(msg projectImportCompletedMsg) (Model, tea.Cmd, bool) {
	if msg.plan == nil {
		return m, nil, true
	}
	if msg.err != nil {
		corelog.For("tui.import").Error("project import failed", "project_id", msg.plan.Project.ProjectID, "err", msg.err)
		m.openErrorDialog("Import Failed", msg.plan.Project, msg.err.Error())
		return m, nil, true
	}
	m.projectImport = nil
	status := "Saved the imported Remote Config as a local draft."
	source := "draft"
	hasDraft := true
	if msg.result.Published {
		status = "Published the imported Remote Config."
		source = "firebase"
		hasDraft = false
	}
	m.dialog = m.dialog.Open(dialogcmp.Config{
		Title: "Import Complete",
		Tone:  dialogcmp.ToneSuccess,
		Body: []string{
			dialogProjectLine(msg.plan.Project),
			"",
			status,
		},
		Buttons: []dialogcmp.Button{{Label: "Close", Variant: dialogcmp.ButtonVariantAccent, OnPress: dialogCanceledCmd()}},
	})
	selectionCmd := m.projects.SelectOnly(msg.plan.Project.ProjectID)
	m.setActive(panels.Parameters)
	loaded := messages.ParametersLoadedMsg{Project: msg.plan.Project, Tree: msg.result.Tree, Source: source, CacheSource: "cache", HasDraft: hasDraft}
	return m, tea.Sequence(selectionCmd, func() tea.Msg { return loaded }), true
}

func (m Model) handleProjectExportRequest(request projectio.ExportRequestedMsg) (Model, tea.Cmd, bool) {
	path, err := filepath.Abs(request.Path)
	if err != nil {
		m.projectIO = m.projectIO.SetError(err)
		return m, nil, true
	}
	session := projectExportSession{project: request.Project, path: path, draft: request.Draft}
	m.projectExport = &session
	if _, statErr := os.Stat(path); statErr == nil {
		m.projectIO = m.projectIO.Close()
		m.openProjectExportOverwrite(session)
		return m, nil, true
	} else if !os.IsNotExist(statErr) {
		m.projectIO = m.projectIO.SetError(statErr)
		return m, nil, true
	}
	m.projectIO = m.projectIO.Close()
	return m, m.exportProjectCmd(session), true
}

func (m *Model) openProjectExportOverwrite(session projectExportSession) {
	m.dialog = m.dialog.Open(dialogcmp.Config{
		Title: "Overwrite Export?",
		Body: []string{
			dialogProjectLine(session.project),
			"",
			"A file already exists at:",
			session.path,
			"",
			"Overwrite it with the selected Remote Config?",
		},
		Buttons: []dialogcmp.Button{
			{Label: "Overwrite", Variant: dialogcmp.ButtonVariantDanger, OnPress: func() tea.Msg { return projectExportOverwriteMsg{} }},
			{Label: "Back", Variant: dialogcmp.ButtonVariantAccent, OnPress: func() tea.Msg { return projectExportBackMsg{} }},
		},
	})
}

func (m Model) exportProjectCmd(session projectExportSession) tea.Cmd {
	return func() tea.Msg {
		var raw []byte
		var err error
		if session.draft {
			var ok bool
			raw, ok, err = m.svc.LoadDraft(session.project.ProjectID)
			if err == nil && !ok {
				err = fmt.Errorf("local draft no longer exists")
			}
		} else {
			raw, _, err = m.svc.ExportRemoteConfig(context.Background(), session.project.ProjectID)
		}
		if err == nil {
			err = rc.WriteRemoteConfigFile(session.path, raw)
		}
		return projectExportCompletedMsg{session: session, err: err}
	}
}

func (m Model) updateProjectExportCompleted(msg projectExportCompletedMsg) (Model, tea.Cmd, bool) {
	if msg.err != nil {
		corelog.For("tui.export").Error("project export failed", "project_id", msg.session.project.ProjectID, "path", msg.session.path, "err", msg.err)
		m.openErrorDialog("Export Failed", msg.session.project, msg.err.Error())
		return m, nil, true
	}
	m.projectExport = nil
	source := "Published Remote Config"
	if msg.session.draft {
		source = "Local draft"
	}
	m.dialog = m.dialog.Open(dialogcmp.Config{
		Title: "Export Complete",
		Tone:  dialogcmp.ToneSuccess,
		Body: []string{
			dialogProjectLine(msg.session.project),
			"",
			source + " exported to:",
			msg.session.path,
		},
		Buttons: []dialogcmp.Button{
			{Label: "Copy Path", Variant: dialogcmp.ButtonVariantAccent, OnPress: copyToClipboardCmd(msg.session.path)},
			{Label: "Close", Variant: dialogcmp.ButtonVariantAccent, OnPress: dialogCanceledCmd()},
		},
	})
	return m, nil, true
}

func (m Model) handleProjectDefaultsRequest(request projectio.DefaultsRequestedMsg) (Model, tea.Cmd, bool) {
	path, err := filepath.Abs(request.Path)
	if err != nil {
		m.projectIO = m.projectIO.SetError(err)
		return m, nil, true
	}
	session := projectDefaultsSession{project: request.Project, path: path, format: request.Format}
	m.projectDefaults = &session
	if _, statErr := os.Stat(path); statErr == nil {
		m.projectIO = m.projectIO.Close()
		m.openProjectDefaultsOverwrite(session)
		return m, nil, true
	} else if !os.IsNotExist(statErr) {
		m.projectIO = m.projectIO.SetError(statErr)
		return m, nil, true
	}
	m.projectIO = m.projectIO.Close()
	return m, m.downloadProjectDefaultsCmd(session), true
}

func (m *Model) openProjectDefaultsOverwrite(session projectDefaultsSession) {
	m.dialog = m.dialog.Open(dialogcmp.Config{
		Title: "Overwrite Defaults?",
		Body: []string{
			dialogProjectLine(session.project),
			"",
			"A file already exists at:",
			session.path,
			"",
			"Overwrite it with the downloaded application defaults?",
		},
		Buttons: []dialogcmp.Button{
			{Label: "Overwrite", Variant: dialogcmp.ButtonVariantDanger, OnPress: func() tea.Msg { return projectDefaultsOverwriteMsg{} }},
			{Label: "Back", Variant: dialogcmp.ButtonVariantAccent, OnPress: func() tea.Msg { return projectDefaultsBackMsg{} }},
		},
	})
}

func (m Model) downloadProjectDefaultsCmd(session projectDefaultsSession) tea.Cmd {
	return func() tea.Msg {
		defaults, err := m.svc.DownloadRemoteConfigDefaults(context.Background(), session.project.ProjectID, session.format)
		if err == nil {
			if session.overwrite {
				err = fileoutput.Write(session.path, defaults)
			} else {
				err = fileoutput.Create(session.path, defaults)
			}
		}
		return projectDefaultsCompletedMsg{session: session, err: err}
	}
}

func (m Model) updateProjectDefaultsCompleted(msg projectDefaultsCompletedMsg) (Model, tea.Cmd, bool) {
	if msg.err != nil {
		corelog.For("tui.defaults").Error(
			"application defaults download failed",
			"project_id", msg.session.project.ProjectID,
			"format", msg.session.format,
			"path", msg.session.path,
			"err", msg.err,
		)
		m.openErrorDialog("Defaults Download Failed", msg.session.project, msg.err.Error())
		return m, nil, true
	}
	m.projectDefaults = nil
	m.dialog = m.dialog.Open(dialogcmp.Config{
		Title: "Defaults Downloaded",
		Tone:  dialogcmp.ToneSuccess,
		Body: []string{
			dialogProjectLine(msg.session.project),
			"",
			strings.ToLower(string(msg.session.format)) + " application defaults downloaded to:",
			msg.session.path,
		},
		Buttons: []dialogcmp.Button{
			{Label: "Copy Path", Variant: dialogcmp.ButtonVariantAccent, OnPress: copyToClipboardCmd(msg.session.path)},
			{Label: "Close", Variant: dialogcmp.ButtonVariantAccent, OnPress: dialogCanceledCmd()},
		},
	})
	return m, nil, true
}
