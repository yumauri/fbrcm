package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/draft"
	"github.com/yumauri/fbrcm/core/firebase"
	rcdiff "github.com/yumauri/fbrcm/core/rc/diff"
	rcpromote "github.com/yumauri/fbrcm/core/rc/promote"
)

type ProjectPromotionSourceMode string

const (
	ProjectPromotionEffective ProjectPromotionSourceMode = "effective"
	ProjectPromotionPublished ProjectPromotionSourceMode = "published"
)

type ProjectPromotionOptions struct {
	SourceMode ProjectPromotionSourceMode
	Prune      bool
	Force      bool
}

type ProjectPromotionSnapshot struct {
	Project      Project
	Raw          json.RawMessage
	PublishedRaw json.RawMessage
	ETag         string
	Version      string
	DraftVersion string
	Source       string
	HasDraft     bool
	StaleDraft   bool
	DraftUpdated time.Time
	CachedAt     time.Time
}

type ProjectPromotionPlan struct {
	Source  ProjectPromotionSnapshot
	Target  ProjectPromotionSnapshot
	Options ProjectPromotionOptions
	Plan    rcpromote.Plan
}

type ProjectPromotionPreview struct {
	Plan              *ProjectPromotionPlan
	Requested         map[rcpromote.ItemID]bool
	Effective         map[rcpromote.ItemID]bool
	Required          []rcpromote.Item
	Applied           []rcpromote.Item
	CandidateRaw      json.RawMessage
	CandidateDiff     rcdiff.Result
	CandidateDiffText string
	PublishRaw        json.RawMessage
	PublishDiff       rcdiff.Result
	PublishDiffText   string
	HasChanges        bool
}

type ProjectPromotionResult struct {
	Cache     *ParametersCache
	Tree      *ParametersTree
	Raw       json.RawMessage
	Drafted   bool
	Published bool
	HasDraft  bool
}

func (s *Core) PrepareProjectPromotion(ctx context.Context, source, target Project, opts ProjectPromotionOptions) (*ProjectPromotionPlan, error) {
	if source.ProjectID == "" || target.ProjectID == "" {
		return nil, fmt.Errorf("source and target projects are required")
	}
	if source.ProjectID == target.ProjectID {
		return nil, fmt.Errorf("source and target projects must be different")
	}
	if opts.SourceMode == "" {
		opts.SourceMode = ProjectPromotionEffective
	}

	sourceSnapshot, err := s.loadPromotionSnapshot(ctx, source, opts.Force)
	if err != nil {
		return nil, fmt.Errorf("load source project %s: %w", source.ProjectID, err)
	}
	targetSnapshot, err := s.loadPromotionSnapshot(ctx, target, opts.Force)
	if err != nil {
		return nil, fmt.Errorf("load target project %s: %w", target.ProjectID, err)
	}
	if opts.SourceMode == ProjectPromotionPublished {
		sourceSnapshot.Raw = append(json.RawMessage(nil), sourceSnapshot.PublishedRaw...)
		sourceSnapshot.Source = "published"
	}

	sourceCfg, err := firebase.ParseCloneRemoteConfig(sourceSnapshot.Raw)
	if err != nil {
		return nil, fmt.Errorf("decode source Remote Config: %w", err)
	}
	targetCfg, err := firebase.ParseCloneRemoteConfig(targetSnapshot.Raw)
	if err != nil {
		return nil, fmt.Errorf("decode target Remote Config: %w", err)
	}
	sourceCfg.Version = firebase.RemoteConfigVersion{}
	targetCfg.Version = firebase.RemoteConfigVersion{}

	// Build with pruning enabled so target-only items remain visible. Preview
	// controls whether selected removals are eligible to apply.
	plan := rcpromote.BuildPlan(sourceCfg, targetCfg, rcpromote.Options{Prune: true})
	return &ProjectPromotionPlan{Source: sourceSnapshot, Target: targetSnapshot, Options: opts, Plan: plan}, nil
}

func (s *Core) loadPromotionSnapshot(ctx context.Context, project Project, force bool) (ProjectPromotionSnapshot, error) {
	var cache *ParametersCache
	var source string
	var err error
	if force {
		cache, source, err = s.GetParameters(ctx, project.ProjectID, true)
	} else {
		cache, _, err = s.InspectParametersCache(project.ProjectID)
		if err == nil && cache != nil {
			source = "cached"
		} else if err == nil {
			cache, source, err = s.GetParameters(ctx, project.ProjectID, false)
		}
	}
	if err != nil {
		return ProjectPromotionSnapshot{}, err
	}
	if cache == nil {
		return ProjectPromotionSnapshot{}, fmt.Errorf("parameters cache not found")
	}
	published, err := firebase.ParseRemoteConfig(cache.RemoteConfig)
	if err != nil {
		return ProjectPromotionSnapshot{}, err
	}
	out := ProjectPromotionSnapshot{
		Project:      project,
		Raw:          append(json.RawMessage(nil), cache.RemoteConfig...),
		PublishedRaw: append(json.RawMessage(nil), cache.RemoteConfig...),
		ETag:         cache.ETag,
		Version:      published.Version.VersionNumber,
		Source:       source,
		CachedAt:     cache.CachedAt,
	}
	record, ok, err := s.LoadDraftRecord(project.ProjectID)
	if err != nil {
		return ProjectPromotionSnapshot{}, err
	}
	if !ok {
		return out, nil
	}
	draftCfg, err := firebase.ParseRemoteConfig(record.RemoteConfig)
	if err != nil {
		return ProjectPromotionSnapshot{}, err
	}
	out.Raw = append(json.RawMessage(nil), record.RemoteConfig...)
	out.Source = "draft"
	out.HasDraft = true
	out.DraftVersion = draftCfg.Version.VersionNumber
	out.DraftUpdated = record.UpdatedAt
	out.StaleDraft = record.BaseVersion != "" && published.Version.VersionNumber != "" && record.BaseVersion != published.Version.VersionNumber
	return out, nil
}

func (s *Core) PreviewProjectPromotion(plan *ProjectPromotionPlan, requested map[rcpromote.ItemID]bool, prune bool) (*ProjectPromotionPreview, error) {
	if plan == nil {
		return nil, fmt.Errorf("promotion plan is nil")
	}
	selected := make(map[rcpromote.ItemID]bool, len(requested))
	for id, enabled := range requested {
		if !enabled {
			continue
		}
		item, ok := promotionItemByID(plan.Plan.Items, id)
		if !ok || item.Change == rcdiff.ChangeRemoved && !prune {
			continue
		}
		selected[id] = true
	}

	candidate, applied, err := rcpromote.Apply(plan.Plan, selected, rcpromote.Options{Prune: prune})
	if err != nil {
		return nil, err
	}
	targetVersion, err := remoteConfigVersionValue(plan.Target.Raw)
	if err != nil {
		return nil, err
	}
	candidate.Version = targetVersion
	candidateRaw, err := firebase.MarshalRemoteConfig(candidate)
	if err != nil {
		return nil, err
	}

	effective := make(map[rcpromote.ItemID]bool, len(applied))
	required := make([]rcpromote.Item, 0)
	for _, item := range applied {
		effective[item.ID] = true
		if !selected[item.ID] {
			item.Required = true
			required = append(required, item)
		}
	}
	candidateDiff := rcdiff.CompareRemoteConfigs(plan.Plan.Target, candidate)
	candidateText, changed := rcdiff.RenderResult(candidateDiff)
	publishRaw, publishDiff, publishText, err := projectPromotionPublishPreview(plan, candidateRaw)
	if err != nil {
		return nil, err
	}
	return &ProjectPromotionPreview{
		Plan:              plan,
		Requested:         selected,
		Effective:         effective,
		Required:          required,
		Applied:           applied,
		CandidateRaw:      candidateRaw,
		CandidateDiff:     candidateDiff,
		CandidateDiffText: candidateText,
		PublishRaw:        publishRaw,
		PublishDiff:       publishDiff,
		PublishDiffText:   publishText,
		HasChanges:        changed,
	}, nil
}

func projectPromotionPublishPreview(plan *ProjectPromotionPlan, candidateRaw json.RawMessage) (json.RawMessage, rcdiff.Result, string, error) {
	publishRaw := append(json.RawMessage(nil), candidateRaw...)
	if plan.Target.HasDraft {
		record, err := config.LoadDraft(plan.Target.Project.ProjectID)
		if err != nil {
			return nil, rcdiff.Result{}, "", err
		}
		if !record.UpdatedAt.Equal(plan.Target.DraftUpdated) {
			return nil, rcdiff.Result{}, "", fmt.Errorf("target draft changed during promotion; refresh the promotion")
		}
		var changed bool
		publishRaw, changed, err = draft.MergeWithLatest(record.BaseRemoteConfig, candidateRaw, plan.Target.PublishedRaw)
		if err != nil {
			return nil, rcdiff.Result{}, "", err
		}
		if !changed {
			publishRaw = append(json.RawMessage(nil), plan.Target.PublishedRaw...)
		}
	}
	current, err := firebase.ParseCloneRemoteConfig(plan.Target.PublishedRaw)
	if err != nil {
		return nil, rcdiff.Result{}, "", err
	}
	finalCfg, err := firebase.ParseCloneRemoteConfig(publishRaw)
	if err != nil {
		return nil, rcdiff.Result{}, "", err
	}
	current.Version = firebase.RemoteConfigVersion{}
	finalCfg.Version = firebase.RemoteConfigVersion{}
	result := rcdiff.CompareRemoteConfigs(current, finalCfg)
	text, _ := rcdiff.RenderResult(result)
	return publishRaw, result, text, nil
}

func (s *Core) SaveProjectPromotionDraft(preview *ProjectPromotionPreview) (*ProjectPromotionResult, error) {
	if err := validatePromotionPreview(preview); err != nil {
		return nil, err
	}
	if err := s.checkPromotionTarget(preview.Plan); err != nil {
		return nil, err
	}
	projectID := preview.Plan.Target.Project.ProjectID
	if preview.Plan.Target.HasDraft {
		if err := s.SaveDraft(projectID, preview.CandidateRaw); err != nil {
			return nil, err
		}
	} else {
		base := &config.ParametersCache{ETag: preview.Plan.Target.ETag, CachedAt: preview.Plan.Target.CachedAt, RemoteConfig: preview.Plan.Target.PublishedRaw}
		if err := draft.SaveWithBase(projectID, base, preview.CandidateRaw); err != nil {
			return nil, err
		}
	}
	tree, err := s.BuildParametersTreeFromRaw(preview.CandidateRaw, preview.Plan.Target.CachedAt, preview.Plan.Target.ETag)
	return &ProjectPromotionResult{Cache: &config.ParametersCache{ETag: preview.Plan.Target.ETag, CachedAt: preview.Plan.Target.CachedAt, RemoteConfig: preview.CandidateRaw}, Tree: tree, Raw: preview.CandidateRaw, Drafted: true, HasDraft: true}, err
}

func (s *Core) PublishProjectPromotion(ctx context.Context, preview *ProjectPromotionPreview) (*ProjectPromotionResult, error) {
	if err := validatePromotionPreview(preview); err != nil {
		return nil, err
	}
	latest, _, err := s.GetParameters(ctx, preview.Plan.Target.Project.ProjectID, true)
	if err != nil {
		return nil, err
	}
	if latest.ETag != preview.Plan.Target.ETag || !bytes.Equal(latest.RemoteConfig, preview.Plan.Target.PublishedRaw) {
		return nil, fmt.Errorf("target Remote Config changed during promotion; refresh the promotion")
	}
	if _, err := s.SaveProjectPromotionDraft(preview); err != nil {
		return nil, err
	}
	publishPlan, err := s.PrepareDraftPublish(ctx, preview.Plan.Target.Project.ProjectID)
	if err != nil {
		return &ProjectPromotionResult{Raw: preview.CandidateRaw, Drafted: true, HasDraft: true}, err
	}
	if !remoteConfigsEqual(publishPlan.Candidate, preview.PublishRaw) {
		return &ProjectPromotionResult{Raw: preview.CandidateRaw, Drafted: true, HasDraft: true}, fmt.Errorf("target Remote Config changed while preparing publication; the promotion remains saved as a draft")
	}
	cache, tree, publishErr := s.ExecuteDraftPublish(ctx, preview.Plan.Target.Project.ProjectID, publishPlan)
	result := &ProjectPromotionResult{Cache: cache, Tree: tree, Published: cache != nil, HasDraft: publishErr != nil}
	if cache != nil {
		result.Raw = cache.RemoteConfig
	}
	return result, publishErr
}

func validatePromotionPreview(preview *ProjectPromotionPreview) error {
	if preview == nil || preview.Plan == nil {
		return fmt.Errorf("promotion preview is incomplete")
	}
	if !preview.HasChanges || len(preview.Requested) == 0 {
		return fmt.Errorf("promotion has no selected changes")
	}
	return nil
}

func (s *Core) checkPromotionTarget(plan *ProjectPromotionPlan) error {
	record, hasDraft, err := s.LoadDraftRecord(plan.Target.Project.ProjectID)
	if err != nil {
		return err
	}
	if hasDraft != plan.Target.HasDraft || hasDraft && !record.UpdatedAt.Equal(plan.Target.DraftUpdated) {
		return fmt.Errorf("target draft changed during promotion; refresh the promotion")
	}
	cache, _, err := s.InspectParametersCache(plan.Target.Project.ProjectID)
	if err != nil {
		return err
	}
	if cache == nil || cache.ETag != plan.Target.ETag || !bytes.Equal(cache.RemoteConfig, plan.Target.PublishedRaw) {
		return fmt.Errorf("target Remote Config changed during promotion; refresh the promotion")
	}
	return nil
}

func promotionItemByID(items []rcpromote.Item, id rcpromote.ItemID) (rcpromote.Item, bool) {
	for _, item := range items {
		if item.ID == id {
			return item, true
		}
	}
	return rcpromote.Item{}, false
}

func remoteConfigVersionValue(raw json.RawMessage) (firebase.RemoteConfigVersion, error) {
	cfg, err := firebase.ParseRemoteConfig(raw)
	if err != nil {
		return firebase.RemoteConfigVersion{}, err
	}
	return cfg.Version, nil
}

func remoteConfigsEqual(left, right json.RawMessage) bool {
	l, err := firebase.ParseCloneRemoteConfig(left)
	if err != nil {
		return false
	}
	r, err := firebase.ParseCloneRemoteConfig(right)
	if err != nil {
		return false
	}
	l.Version = firebase.RemoteConfigVersion{}
	r.Version = firebase.RemoteConfigVersion{}
	lRaw, lErr := firebase.MarshalRemoteConfig(l)
	rRaw, rErr := firebase.MarshalRemoteConfig(r)
	return lErr == nil && rErr == nil && bytes.Equal(lRaw, rRaw)
}
