package app

import (
	"context"
	"errors"
	"fmt"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	rcdiff "github.com/yumauri/fbrcm/core/rc/diff"
	"github.com/yumauri/fbrcm/core/rc/display"
	dialogcmp "github.com/yumauri/fbrcm/tui/components/dialog"
	"github.com/yumauri/fbrcm/tui/messages"
)

type draftPublishPhase int

const (
	draftPublishPreparing draftPublishPhase = iota + 1
	draftPublishReviewing
	draftPublishPublishing
	draftPublishCompleted
)

type draftPublishItem struct {
	project  core.Project
	plan     *core.DraftPublishPlan
	prepare  error
	approved bool
	skipped  bool
}

type draftPublishResult struct {
	project   core.Project
	status    string
	err       error
	published bool
}

type draftPublishBatch struct {
	phase   draftPublishPhase
	items   []draftPublishItem
	results []draftPublishResult
	current int
}

type draftPublishPreparedMsg struct{ items []draftPublishItem }
type draftPublishDecisionMsg struct{ decision string }
type draftPublishExecutedMsg struct {
	item  draftPublishItem
	cache *core.ParametersCache
	tree  *core.ParametersTree
	err   error
}
type draftPublishRetryMsg struct{}
type draftPublishClosedMsg struct{}

func (m Model) beginDraftPublishBatch(projects []core.Project) (Model, tea.Cmd, bool) {
	if len(projects) == 0 || m.svc == nil {
		return m, nil, false
	}
	m.draftPublish = &draftPublishBatch{phase: draftPublishPreparing}
	m.dialog = m.dialog.Open(dialogcmp.Config{
		Title: "Preparing Draft Publishes",
		Body: []string{
			fmt.Sprintf("Preparing %s against fresh Firebase state.", display.FormatCount(len(projects), "draft publish plan", "draft publish plans")),
			"",
			"Remote Config publication is non-atomic across projects.",
		},
	})
	return m, func() tea.Msg {
		items := make([]draftPublishItem, 0, len(projects))
		for _, project := range projects {
			plan, err := m.svc.PrepareDraftPublish(context.Background(), project.ProjectID)
			items = append(items, draftPublishItem{project: project, plan: plan, prepare: err})
		}
		return draftPublishPreparedMsg{items: items}
	}, true
}

func (m Model) updateDraftPublishPrepared(msg draftPublishPreparedMsg) (Model, tea.Cmd, bool) {
	if m.draftPublish == nil {
		return m, nil, true
	}
	m.draftPublish.phase = draftPublishReviewing
	m.draftPublish.items = msg.items
	m.draftPublish.current = 0
	for _, item := range msg.items {
		if item.prepare != nil {
			m.draftPublish.results = append(m.draftPublish.results, draftPublishResult{
				project: item.project,
				status:  "preparation failed",
				err:     item.prepare,
			})
		}
	}
	return m.advanceDraftPublishReview()
}

func (m Model) advanceDraftPublishReview() (Model, tea.Cmd, bool) {
	if m.draftPublish == nil {
		return m, nil, true
	}
	for m.draftPublish.current < len(m.draftPublish.items) {
		item := m.draftPublish.items[m.draftPublish.current]
		if item.prepare != nil {
			m.draftPublish.current++
			continue
		}
		fromCfg, fromErr := firebase.ParseRemoteConfig(item.plan.Latest.RemoteConfig)
		toCfg, toErr := firebase.ParseRemoteConfig(item.plan.Candidate)
		if fromErr != nil || toErr != nil {
			err := fromErr
			if err == nil {
				err = toErr
			}
			m.draftPublish.results = append(m.draftPublish.results, draftPublishResult{project: item.project, status: "preparation failed", err: err})
			m.draftPublish.items[m.draftPublish.current].prepare = err
			m.draftPublish.current++
			continue
		}
		diffText, changed := rcdiff.RenderRemoteConfigDiff(fromCfg, toCfg)
		if !changed {
			diffText = "No changes; publishing will remove the already-applied draft."
		}
		body := []string{dialogProjectLine(item.project), ""}
		if len(m.draftPublish.items) > 1 {
			body = append(body,
				"Drafts are published independently for each project.",
				"Failures do not roll back projects already published.",
				"",
			)
		}
		body = append(body, dialogDiffLines(diffText)...)
		m.dialog = m.dialog.Open(dialogcmp.Config{
			Title: "Review Draft Publish",
			Body:  body,
			Buttons: []dialogcmp.Button{
				{Label: "Approve", Variant: dialogcmp.ButtonVariantDanger, OnPress: draftPublishDecisionCmd("approve")},
				{Label: "Skip", Variant: dialogcmp.ButtonVariantAccent, OnPress: draftPublishDecisionCmd("skip")},
				{Label: "Cancel All", Variant: dialogcmp.ButtonVariantAccent, OnPress: draftPublishDecisionCmd("cancel")},
			},
		})
		return m, nil, true
	}
	return m.startDraftPublishExecution()
}

func draftPublishDecisionCmd(decision string) tea.Cmd {
	return func() tea.Msg { return draftPublishDecisionMsg{decision: decision} }
}

func (m Model) updateDraftPublishDecision(msg draftPublishDecisionMsg) (Model, tea.Cmd, bool) {
	if m.draftPublish == nil || m.draftPublish.current >= len(m.draftPublish.items) {
		return m, nil, true
	}
	switch msg.decision {
	case "approve":
		m.draftPublish.items[m.draftPublish.current].approved = true
	case "skip":
		item := &m.draftPublish.items[m.draftPublish.current]
		item.skipped = true
		m.draftPublish.results = append(m.draftPublish.results, draftPublishResult{project: item.project, status: "skipped"})
	case "cancel":
		for index := range m.draftPublish.items {
			item := &m.draftPublish.items[index]
			if item.skipped || item.prepare != nil {
				continue
			}
			item.skipped = true
			m.draftPublish.results = append(m.draftPublish.results, draftPublishResult{project: item.project, status: "canceled"})
		}
		for index := range m.draftPublish.items {
			m.draftPublish.items[index].approved = false
		}
		return m.finishDraftPublishBatch()
	}
	m.draftPublish.current++
	return m.advanceDraftPublishReview()
}

func (m Model) startDraftPublishExecution() (Model, tea.Cmd, bool) {
	m.draftPublish.phase = draftPublishPublishing
	m.draftPublish.current = 0
	return m.publishNextDraftPlan()
}

func (m Model) publishNextDraftPlan() (Model, tea.Cmd, bool) {
	for m.draftPublish.current < len(m.draftPublish.items) {
		item := m.draftPublish.items[m.draftPublish.current]
		m.draftPublish.current++
		if !item.approved {
			continue
		}
		return m, func() tea.Msg {
			cache, tree, err := m.svc.ExecuteDraftPublish(context.Background(), item.project.ProjectID, item.plan)
			return draftPublishExecutedMsg{item: item, cache: cache, tree: tree, err: err}
		}, true
	}
	return m.finishDraftPublishBatch()
}

func (m Model) updateDraftPublishExecuted(msg draftPublishExecutedMsg) (Model, tea.Cmd, bool) {
	result := draftPublishResult{project: msg.item.project, status: "published", published: true}
	if !msg.item.plan.HasChanges && msg.err == nil {
		result.status = "already applied"
	}
	if msg.err != nil {
		result.err = msg.err
		result.published = msg.cache != nil
		var cleanupErr *core.DraftPublishedCleanupError
		var cacheErr *core.RemoteConfigPublishedCacheError
		switch {
		case errors.As(msg.err, &cleanupErr):
			result.status = "published; cleanup failed"
		case errors.As(msg.err, &cacheErr):
			result.status = "published; cache update failed"
		case result.published:
			result.status = "published; local update failed"
		default:
			result.status = "publish failed"
		}
	}
	m.draftPublish.results = append(m.draftPublish.results, result)

	var update tea.Cmd
	if msg.tree != nil {
		hasDraft := msg.err != nil
		loaded := messages.ParametersLoadedMsg{
			Project: msg.item.project, Tree: msg.tree, Source: "firebase", CacheSource: "firebase", HasDraft: hasDraft,
		}
		update = func() tea.Msg { return loaded }
	}
	next, nextCmd, _ := m.publishNextDraftPlan()
	return next, tea.Batch(update, nextCmd), true
}

func (m Model) finishDraftPublishBatch() (Model, tea.Cmd, bool) {
	if m.draftPublish == nil {
		return m, nil, true
	}
	m.draftPublish.phase = draftPublishCompleted
	body := []string{"Remote Config publication is non-atomic.", ""}
	retryable := false
	for _, result := range m.orderedDraftPublishResults() {
		line := fmt.Sprintf("%s: %s", result.project.ProjectID, result.status)
		if result.err != nil {
			line += ": " + result.err.Error()
			if !result.published {
				retryable = true
			}
		}
		body = append(body, line)
	}
	buttons := []dialogcmp.Button{{Label: "Close", Variant: dialogcmp.ButtonVariantAccent, OnPress: func() tea.Msg { return draftPublishClosedMsg{} }}}
	if retryable {
		buttons = append([]dialogcmp.Button{{Label: "Retry Failed", Variant: dialogcmp.ButtonVariantDanger, OnPress: func() tea.Msg { return draftPublishRetryMsg{} }}}, buttons...)
	}
	m.dialog = m.dialog.Open(dialogcmp.Config{Title: "Draft Publish Results", Body: body, Buttons: buttons})
	return m, nil, true
}

func (m Model) retryDraftPublishFailures() (Model, tea.Cmd, bool) {
	if m.draftPublish == nil {
		return m, nil, true
	}
	projects := make([]core.Project, 0)
	for _, result := range m.orderedDraftPublishResults() {
		if result.err != nil && !result.published {
			projects = append(projects, result.project)
		}
	}
	m.draftPublish = nil
	if len(projects) == 0 {
		return m, nil, true
	}
	return m.beginDraftPublishBatch(projects)
}

func (m Model) orderedDraftPublishResults() []draftPublishResult {
	if m.draftPublish == nil {
		return nil
	}
	ordered := make([]draftPublishResult, 0, len(m.draftPublish.results))
	for _, item := range m.draftPublish.items {
		for _, result := range m.draftPublish.results {
			if result.project.ProjectID == item.project.ProjectID {
				ordered = append(ordered, result)
				break
			}
		}
	}
	return ordered
}
