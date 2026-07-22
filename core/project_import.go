package core

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/yumauri/fbrcm/core/firebase"
	rcdiff "github.com/yumauri/fbrcm/core/rc/diff"
	"github.com/yumauri/fbrcm/core/rc/importer"
)

type (
	ProjectImportOptions    = importer.Options
	ProjectImportStrategy   = importer.Strategy
	ProjectImportResolution = importer.Resolution
	ProjectImportConflict   = importer.Conflict
	ProjectImportSummary    = importer.Summary
	ProjectConditionPolicy  = importer.ConditionPolicy
)

const (
	ProjectImportMerge   = importer.StrategyMerge
	ProjectImportReplace = importer.StrategyReplace

	ProjectImportKeepCurrent = importer.ResolutionCurrent
	ProjectImportUseImported = importer.ResolutionImport

	ProjectImportKeepConditions         = importer.ConditionPolicyKeep
	ProjectImportKeepPortableConditions = importer.ConditionPolicyKeepPortableOnly
	ProjectImportRemoveAllConditions    = importer.ConditionPolicyRemoveAll
)

type ProjectImportPlan struct {
	Project          Project
	SourcePath       string
	Options          ProjectImportOptions
	Source           *importer.ParsedSource
	Summary          ProjectImportSummary
	Conflicts        []ProjectImportConflict
	CurrentRaw       json.RawMessage
	DraftRaw         json.RawMessage
	PublishRaw       json.RawMessage
	Diff             string
	HasChanges       bool
	HasDraft         bool
	BaseETag         string
	BaseCachedAt     time.Time
	BaseDraftUpdated time.Time
}

type ProjectImportResult struct {
	Cache     *ParametersCache
	Tree      *ParametersTree
	Raw       json.RawMessage
	Drafted   bool
	Published bool
}

func (s *Core) PrepareProjectImport(ctx context.Context, project Project, sourceRaw []byte, opts ProjectImportOptions) (*ProjectImportPlan, error) {
	source, err := importer.ParseSource(sourceRaw)
	if err != nil {
		return nil, err
	}
	cache, _, err := s.GetParameters(ctx, project.ProjectID, false)
	if err != nil {
		return nil, err
	}
	currentRaw := append(json.RawMessage(nil), cache.RemoteConfig...)
	hasDraft := false
	var draftUpdated time.Time
	if record, ok, loadErr := s.LoadDraftRecord(project.ProjectID); loadErr != nil {
		return nil, loadErr
	} else if ok {
		hasDraft = true
		draftUpdated = record.UpdatedAt
		currentRaw = append(json.RawMessage(nil), record.RemoteConfig...)
	}
	currentCfg, err := firebase.ParseCloneRemoteConfig(currentRaw)
	if err != nil {
		return nil, fmt.Errorf("decode current remote config: %w", err)
	}
	currentVersion := currentCfg.Version
	currentCfg.Version = firebase.RemoteConfigVersion{}

	planned, err := importer.BuildPlan(project.ProjectID, project.Name, currentCfg, source, opts)
	if err != nil {
		return nil, err
	}
	planned.Final.Version = firebase.RemoteConfigVersion{}
	diffText, changed := rcdiff.RenderRemoteConfigDiff(currentCfg, planned.Final)
	publishRaw, err := firebase.MarshalRemoteConfigForUpdate(planned.Final)
	if err != nil {
		return nil, err
	}
	planned.Final.Version = currentVersion
	draftRaw, err := firebase.MarshalRemoteConfig(planned.Final)
	if err != nil {
		return nil, err
	}
	return &ProjectImportPlan{
		Project:          project,
		Options:          opts,
		Source:           source,
		Summary:          planned.Summary,
		Conflicts:        planned.Conflicts,
		CurrentRaw:       currentRaw,
		DraftRaw:         draftRaw,
		PublishRaw:       publishRaw,
		Diff:             diffText,
		HasChanges:       changed,
		HasDraft:         hasDraft,
		BaseETag:         cache.ETag,
		BaseCachedAt:     cache.CachedAt,
		BaseDraftUpdated: draftUpdated,
	}, nil
}

func (s *Core) ExecuteProjectImport(ctx context.Context, plan *ProjectImportPlan, publish bool) (*ProjectImportResult, error) {
	if plan == nil {
		return nil, fmt.Errorf("import plan is nil")
	}
	if !plan.HasChanges {
		return nil, fmt.Errorf("import has no changes")
	}
	if plan.HasDraft && publish {
		return nil, fmt.Errorf("project %s has an unpublished draft; update the draft or publish it separately", plan.Project.ProjectID)
	}
	if publish {
		latest, _, err := s.GetParameters(ctx, plan.Project.ProjectID, true)
		if err != nil {
			return nil, err
		}
		if latest.ETag != plan.BaseETag {
			return nil, fmt.Errorf("current Remote Config changed during import preview; review the import again")
		}
		if err := s.ValidateRemoteConfigWithETag(ctx, plan.Project.ProjectID, plan.PublishRaw, latest.ETag); err != nil {
			return nil, err
		}
		updatedRaw, nextETag, publishErr := s.PublishRemoteConfigWithETag(ctx, plan.Project.ProjectID, plan.PublishRaw, latest.ETag)
		if publishErr != nil && (len(updatedRaw) == 0 || nextETag == "") {
			return nil, publishErr
		}
		if len(updatedRaw) == 0 || nextETag == "" {
			return nil, fmt.Errorf("published remote config response is incomplete")
		}
		cache := &ParametersCache{ETag: nextETag, CachedAt: time.Now().UTC(), RemoteConfig: updatedRaw}
		tree, treeErr := s.BuildParametersTree(cache)
		result := &ProjectImportResult{Cache: cache, Tree: tree, Raw: updatedRaw, Published: true}
		if publishErr != nil || treeErr != nil {
			return result, errors.Join(publishErr, treeErr)
		}
		return result, nil
	}

	record, hasDraft, err := s.LoadDraftRecord(plan.Project.ProjectID)
	if err != nil {
		return nil, err
	}
	if hasDraft != plan.HasDraft || hasDraft && !record.UpdatedAt.Equal(plan.BaseDraftUpdated) {
		return nil, fmt.Errorf("local draft changed during import preview; review the import again")
	}
	if !hasDraft {
		cache, _, loadErr := s.GetParameters(ctx, plan.Project.ProjectID, false)
		if loadErr != nil {
			return nil, loadErr
		}
		if cache.ETag != plan.BaseETag || !bytes.Equal(cache.RemoteConfig, plan.CurrentRaw) {
			return nil, fmt.Errorf("current Remote Config changed during import preview; review the import again")
		}
	}
	if err := s.SaveDraft(plan.Project.ProjectID, plan.DraftRaw); err != nil {
		return nil, err
	}
	cache := &ParametersCache{ETag: plan.BaseETag, CachedAt: plan.BaseCachedAt, RemoteConfig: plan.DraftRaw}
	tree, err := s.BuildParametersTree(cache)
	return &ProjectImportResult{Cache: cache, Tree: tree, Raw: plan.DraftRaw, Drafted: true}, err
}
