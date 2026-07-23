package app

import (
	"context"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	corelog "github.com/yumauri/fbrcm/core/log"
	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
	dialogcmp "github.com/yumauri/fbrcm/tui/components/dialog"
	promotecmp "github.com/yumauri/fbrcm/tui/components/promote"
	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/panels"
)

type promoteReturnState struct {
	active        panels.ID
	parametersTab panels.ID
	projectsMode  projectsPanelMode
	logsMode      logsPanelMode
	logsHeight    int
	logsSaved     int
}

type projectPromotionPreparedMsg struct {
	plan     *core.ProjectPromotionPlan
	verified bool
	err      error
}

type projectPromotionExecuteMsg struct{ publish bool }
type projectPromotionBackMsg struct{}

type projectPromotionCompletedMsg struct {
	preview *core.ProjectPromotionPreview
	result  *core.ProjectPromotionResult
	publish bool
	err     error
}

func (m Model) openPromote() (Model, tea.Cmd, bool) {
	project, ok := m.projects.CurrentProject()
	if !ok {
		m.openErrorDialog("Promotion Unavailable", core.Project{}, "No source project is selected under the Projects panel cursor.")
		return m, nil, true
	}
	if len(m.projects.AllProjects()) < 2 {
		m.openErrorDialog("Promotion Unavailable", project, "At least two projects are required for promotion.")
		return m, nil, true
	}
	if m.details.Dirty() {
		m.openErrorDialog("Promotion Unavailable", project, "Save or discard the unsaved Details changes before opening the Promote workspace.")
		return m, nil, true
	}
	if !m.promote.WorkspaceOpen() {
		m.promoteReturn = &promoteReturnState{
			active: m.active, parametersTab: m.parametersTab,
			projectsMode: m.projectsMode, logsMode: m.logsMode,
			logsHeight: m.logsHeight, logsSaved: m.logsSaved,
		}
		m.detailsVisible = false
	}
	m.promote = m.promote.OpenTargetPicker(project, m.projects.AllProjects())
	if row, anchorOK := m.projects.CurrentProjectScreenRow(); anchorOK {
		m.promote = m.promote.SetTargetRow(row)
	}
	return m, nil, true
}

func (m Model) updatePromoteMessage(msg tea.Msg) (Model, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case promotecmp.TargetSelectedMsg:
		return m.beginProjectPromotion(msg.Source, msg.Target, msg.Mode)
	case projectPromotionPreparedMsg:
		if msg.err != nil {
			m.promote = m.promote.SetError(msg.err)
			return m, nil, true
		}
		m.promote = m.promote.SetPlan(msg.plan, msg.verified)
		return m, nil, true
	case promotecmp.SwapRequestedMsg:
		source, target := m.promote.Target(), m.promote.Source()
		return m.beginProjectPromotion(source, target, core.ProjectPromotionEffective)
	case promotecmp.SourceModeRequestedMsg:
		return m.beginProjectPromotion(msg.Source, msg.Target, msg.Mode)
	case promotecmp.DiffRequestedMsg:
		m.openPromotionDiff(msg.Item)
		return m, nil, true
	case promotecmp.CloseRequestedMsg:
		return m.closePromote(), nil, true
	case promotecmp.SaveRequestedMsg:
		m.openProjectPromotionReview(msg.Preview, false)
		return m, nil, true
	case promotecmp.PublishRequestedMsg:
		m.openProjectPromotionReview(msg.Preview, true)
		return m, nil, true
	case projectPromotionBackMsg:
		m.promotionPreview = nil
		return m, nil, true
	case projectPromotionExecuteMsg:
		return m, m.executeProjectPromotionCmd(msg.publish), true
	case projectPromotionCompletedMsg:
		return m.updateProjectPromotionCompleted(msg)
	}
	return m, nil, false
}

func (m Model) beginProjectPromotion(source, target core.Project, mode core.ProjectPromotionSourceMode) (Model, tea.Cmd, bool) {
	m.promote = m.promote.SetLoading(source, target, mode)
	m.setActive(panels.Promote)
	force := !source.Disabled && !target.Disabled
	return m, func() tea.Msg {
		plan, err := m.svc.PrepareProjectPromotion(context.Background(), source, target, core.ProjectPromotionOptions{SourceMode: mode, Force: force})
		if err != nil && force {
			cached, cachedErr := m.svc.PrepareProjectPromotion(context.Background(), source, target, core.ProjectPromotionOptions{SourceMode: mode})
			if cachedErr == nil {
				return projectPromotionPreparedMsg{plan: cached, verified: false}
			}
		}
		return projectPromotionPreparedMsg{plan: plan, verified: force && err == nil, err: err}
	}, true
}

func (m *Model) openProjectPromotionReview(preview *core.ProjectPromotionPreview, publish bool) {
	if preview == nil || preview.Plan == nil {
		return
	}
	m.promotionPreview = preview
	plan := preview.Plan
	body := []string{
		"Source: " + plan.Source.Project.Name + " (" + plan.Source.Project.ProjectID + ") · " + plan.Source.Source,
		"Target: " + plan.Target.Project.Name + " (" + plan.Target.Project.ProjectID + ") · " + plan.Target.Source,
		"",
		rcdisplay.FormatCount(len(preview.Requested), "selected change", "selected changes") +
			" · " + rcdisplay.FormatCount(len(preview.Required), "automatic dependency", "automatic dependencies"),
		"",
	}
	title := "Save Promotion Draft?"
	label := "Save Draft"
	variant := dialogcmp.ButtonVariantAccent
	diffText := preview.CandidateDiffText
	if plan.Target.HasDraft {
		label = "Update Draft"
	}
	if publish {
		title = "Publish Promoted Remote Config?"
		label = "Publish"
		variant = dialogcmp.ButtonVariantDanger
		diffText = preview.PublishDiffText
		body = append(body,
			"The complete target draft is shown below, including pre-existing changes.",
			"If validation or publication fails, the promoted candidate remains as a draft.",
			"",
		)
	}
	body = append(body, dialogDiffLines(diffText)...)
	m.dialog = m.dialog.Open(dialogcmp.Config{
		Title: title,
		Body:  body,
		Buttons: []dialogcmp.Button{
			{Label: label, Variant: variant, OnPress: func() tea.Msg { return projectPromotionExecuteMsg{publish: publish} }},
			{Label: "Back", Variant: dialogcmp.ButtonVariantAccent, OnPress: func() tea.Msg { return projectPromotionBackMsg{} }},
		},
	})
}

func (m Model) executeProjectPromotionCmd(publish bool) tea.Cmd {
	preview := m.promotionPreview
	return func() tea.Msg {
		var result *core.ProjectPromotionResult
		var err error
		if publish {
			result, err = m.svc.PublishProjectPromotion(context.Background(), preview)
		} else {
			result, err = m.svc.SaveProjectPromotionDraft(preview)
		}
		return projectPromotionCompletedMsg{preview: preview, result: result, publish: publish, err: err}
	}
}

func (m Model) updateProjectPromotionCompleted(msg projectPromotionCompletedMsg) (Model, tea.Cmd, bool) {
	if msg.preview == nil || msg.preview.Plan == nil {
		return m, nil, true
	}
	target := msg.preview.Plan.Target.Project
	if msg.err != nil && (msg.result == nil || !msg.result.Published) {
		corelog.For("tui.promote").Error("project promotion failed", "source_project_id", msg.preview.Plan.Source.Project.ProjectID, "target_project_id", target.ProjectID, "publish", msg.publish, "err", msg.err)
		body := []string{dialogProjectLine(target), "", msg.err.Error()}
		if msg.result != nil && msg.result.HasDraft {
			body = append(body, "", "The promoted candidate remains saved as a local draft.")
		}
		m.dialog = m.dialog.Open(dialogcmp.Config{Title: "Promotion Failed", Body: body, Buttons: []dialogcmp.Button{{Label: "Close", Variant: dialogcmp.ButtonVariantAccent, OnPress: dialogCanceledCmd()}}})
		return m, nil, true
	}
	m.promotionPreview = nil
	status := "Saved selected promotion changes as a local draft."
	source := "draft"
	hasDraft := true
	if msg.result.Published {
		status = "Published selected promotion changes."
		source = "firebase"
		hasDraft = false
	}
	title := "Promotion Complete"
	if msg.err != nil {
		title = "Promotion Completed with Warning"
		status += " " + msg.err.Error()
	}
	m = m.closePromote()
	m.dialog = m.dialog.Open(dialogcmp.Config{Title: title, Tone: dialogcmp.ToneSuccess, Body: []string{dialogProjectLine(target), "", status}, Buttons: []dialogcmp.Button{{Label: "Close", Variant: dialogcmp.ButtonVariantAccent, OnPress: dialogCanceledCmd()}}})
	selectionCmd := m.projects.SelectOnly(target.ProjectID)
	m.setActive(panels.Parameters)
	loaded := messages.ParametersLoadedMsg{Project: target, Tree: msg.result.Tree, Source: source, CacheSource: "cache", HasDraft: hasDraft}
	return m, tea.Sequence(selectionCmd, func() tea.Msg { return loaded }), true
}

func (m Model) closePromote() Model {
	m.promote = m.promote.Close()
	m.promotionPreview = nil
	state := m.promoteReturn
	m.promoteReturn = nil
	if state == nil {
		m.setActive(panels.Projects)
		return m
	}
	m.projectsMode = state.projectsMode
	m.logsMode = state.logsMode
	m.logsHeight = state.logsHeight
	m.logsSaved = state.logsSaved
	m.parametersTab = state.parametersTab
	m.setActive(state.active)
	if m.width > 0 && m.height > 0 {
		m.applyLayout()
	}
	return m
}
