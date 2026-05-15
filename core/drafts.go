package core

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/firebase"
	corelog "github.com/yumauri/fbrcm/core/log"
)

// paramSlot holds param slot state used by the core package.
type paramSlot struct {
	// group stores group for paramSlot.
	group string
	// param stores param for paramSlot.
	param firebase.RemoteConfigParam
}

// ParameterDetailsEdit holds parameter details edit state used by the core package.
type ParameterDetailsEdit struct {
	// Create stores create for ParameterDetailsEdit.
	Create bool
	// GroupKey stores group key for ParameterDetailsEdit.
	GroupKey string
	// ParamKey stores param key for ParameterDetailsEdit.
	ParamKey string
	// NextGroupKey stores next group key for ParameterDetailsEdit.
	NextGroupKey string
	// NextParamKey stores next param key for ParameterDetailsEdit.
	NextParamKey string
	// NextValueType stores next value type for ParameterDetailsEdit.
	NextValueType string
	// NextDescription stores next description for ParameterDetailsEdit.
	NextDescription string
	// ValueEdits stores value edits for ParameterDetailsEdit.
	ValueEdits []ParameterValueEdit
}

// ParameterValueEdit holds parameter value edit state used by the core package.
type ParameterValueEdit struct {
	// Label stores label for ParameterValueEdit.
	Label string
	// NextValue stores next value for ParameterValueEdit.
	NextValue string
}

var jsonNumberPattern = regexp.MustCompile(`^-?(0|[1-9]\d*)(\.\d+)?([eE][+-]?\d+)?$`)

// NormalizeRemoteConfigGroupKey handles normalize remote config group key and returns the resulting value or error.
func NormalizeRemoteConfigGroupKey(groupKey string) string {
	if groupKey == defaultParametersGroupKey {
		return ""
	}
	return groupKey
}

// ListDraftProjectIDs lists draft project ids for Core and returns the resulting state or error.
func (s *Core) ListDraftProjectIDs() ([]string, error) {
	return config.ListDraftProjectIDs()
}

// LoadDraft loads draft for Core and returns the resulting state or error.
func (s *Core) LoadDraft(projectID string) (json.RawMessage, bool, error) {
	raw, err := config.LoadDraft(projectID)
	if err != nil {
		var pathErr *os.PathError
		if errors.Is(err, os.ErrNotExist) || (errors.As(err, &pathErr) && errors.Is(pathErr.Err, os.ErrNotExist)) {
			return nil, false, nil
		}
		return nil, false, err
	}
	if _, err := firebase.ParseRemoteConfig(raw); err != nil {
		return nil, false, fmt.Errorf("decode draft: %w", err)
	}
	return raw, true, nil
}

// SaveDraft saves draft for Core and returns the resulting state or error.
func (s *Core) SaveDraft(projectID string, raw json.RawMessage) error {
	if _, err := firebase.ParseRemoteConfig(raw); err != nil {
		return fmt.Errorf("decode draft: %w", err)
	}
	return config.SaveDraft(projectID, raw)
}

// DeleteDraft removes draft for Core and returns the resulting state or error.
func (s *Core) DeleteDraft(projectID string) error {
	return config.DeleteDraft(projectID)
}

// BuildParametersTreeFromRaw handles build parameters tree from raw for Core and returns the resulting state or error.
func (s *Core) BuildParametersTreeFromRaw(raw json.RawMessage, cachedAt time.Time, etag string) (*ParametersTree, error) {
	return s.BuildParametersTree(&config.ParametersCache{
		ETag:         etag,
		CachedAt:     cachedAt,
		RemoteConfig: raw,
	})
}

// BuildDraftAwareParametersTree handles build draft aware parameters tree for Core and returns the resulting state or error.
func (s *Core) BuildDraftAwareParametersTree(projectID string, cache *ParametersCache) (*ParametersTree, bool, error) {
	draftRaw, hasDraft, err := s.LoadDraft(projectID)
	if err != nil {
		return nil, false, err
	}
	if hasDraft {
		tree, err := s.BuildParametersTreeFromRaw(draftRaw, cache.CachedAt, cache.ETag)
		return tree, true, err
	}
	tree, err := s.BuildParametersTree(cache)
	return tree, false, err
}

// DeleteParameter removes parameter for Core and returns the resulting state or error.
func (s *Core) DeleteParameter(ctx context.Context, projectID, groupKey, paramKey string, publish bool) (*ParametersCache, *ParametersTree, bool, error) {
	logger := corelog.For("core")
	cache, source, err := s.GetParameters(ctx, projectID, false)
	_ = source
	if err != nil {
		return nil, nil, false, err
	}

	currentRaw, hasDraft, err := s.LoadDraft(projectID)
	if err != nil {
		return nil, nil, false, err
	}
	if !hasDraft {
		currentRaw = cache.RemoteConfig
	}

	currentCfg, err := firebase.ParseRemoteConfig(currentRaw)
	if err != nil {
		return nil, nil, false, fmt.Errorf("decode current remote config: %w", err)
	}
	finalCfg := cloneRemoteConfig(currentCfg)
	removeParamSlot(finalCfg, paramKey, NormalizeRemoteConfigGroupKey(groupKey))
	pruneUnusedConditions(finalCfg)
	removeEmptyGroups(finalCfg)
	dropUnknownConditionReferences(finalCfg)

	if reflect.DeepEqual(currentCfg, finalCfg) {
		return nil, nil, hasDraft, fmt.Errorf("parameter not found")
	}

	finalRaw, err := marshalRemoteConfig(finalCfg)
	if err != nil {
		return nil, nil, hasDraft, err
	}

	if publish {
		updatedRaw, nextETag, err := s.PublishRemoteConfigWithETag(ctx, projectID, finalRaw, cache.ETag)
		if err != nil {
			return nil, nil, hasDraft, err
		}
		if err := config.DeleteDraft(projectID); err != nil {
			logger.Warn("remove draft after publish failed", "project_id", projectID, "err", err)
		}
		updatedCache := &config.ParametersCache{
			ETag:         nextETag,
			CachedAt:     time.Now().UTC(),
			RemoteConfig: updatedRaw,
		}
		tree, err := s.BuildParametersTree(updatedCache)
		return updatedCache, tree, false, err
	}

	if err := config.SaveDraft(projectID, finalRaw); err != nil {
		return nil, nil, hasDraft, err
	}
	tree, err := s.BuildParametersTreeFromRaw(finalRaw, cache.CachedAt, cache.ETag)
	return cache, tree, true, err
}

// DeleteGroup removes group for Core and returns the resulting state or error.
func (s *Core) DeleteGroup(ctx context.Context, projectID, groupKey string, publish bool) (*ParametersCache, *ParametersTree, bool, error) {
	logger := corelog.For("core")
	cache, source, err := s.GetParameters(ctx, projectID, false)
	_ = source
	if err != nil {
		return nil, nil, false, err
	}

	currentRaw, hasDraft, err := s.LoadDraft(projectID)
	if err != nil {
		return nil, nil, false, err
	}
	if !hasDraft {
		currentRaw = cache.RemoteConfig
	}

	currentCfg, err := firebase.ParseRemoteConfig(currentRaw)
	if err != nil {
		return nil, nil, false, fmt.Errorf("decode current remote config: %w", err)
	}
	finalCfg := cloneRemoteConfig(currentCfg)
	if err := removeGroupSlot(finalCfg, NormalizeRemoteConfigGroupKey(groupKey)); err != nil {
		logger.Error("delete group failed", "project_id", projectID, "group", groupKey, "publish", publish, "err", err)
		return nil, nil, hasDraft, err
	}
	pruneUnusedConditions(finalCfg)
	removeEmptyGroups(finalCfg)
	dropUnknownConditionReferences(finalCfg)

	if reflect.DeepEqual(currentCfg, finalCfg) {
		return nil, nil, hasDraft, fmt.Errorf("group not changed")
	}

	finalRaw, err := marshalRemoteConfig(finalCfg)
	if err != nil {
		return nil, nil, hasDraft, err
	}

	if publish {
		updatedRaw, nextETag, err := s.PublishRemoteConfigWithETag(ctx, projectID, finalRaw, cache.ETag)
		if err != nil {
			return nil, nil, hasDraft, err
		}
		if err := config.DeleteDraft(projectID); err != nil {
			logger.Warn("remove draft after publish failed", "project_id", projectID, "err", err)
		}
		updatedCache := &config.ParametersCache{
			ETag:         nextETag,
			CachedAt:     time.Now().UTC(),
			RemoteConfig: updatedRaw,
		}
		tree, err := s.BuildParametersTree(updatedCache)
		return updatedCache, tree, false, err
	}

	if err := config.SaveDraft(projectID, finalRaw); err != nil {
		return nil, nil, hasDraft, err
	}
	tree, err := s.BuildParametersTreeFromRaw(finalRaw, cache.CachedAt, cache.ETag)
	return cache, tree, true, err
}

// RenameParameter handles rename parameter for Core and returns the resulting state or error.
func (s *Core) RenameParameter(ctx context.Context, projectID, groupKey, paramKey, nextParamKey string, publish bool) (*ParametersCache, *ParametersTree, bool, error) {
	logger := corelog.For("core")
	cache, source, err := s.GetParameters(ctx, projectID, false)
	_ = source
	if err != nil {
		return nil, nil, false, err
	}

	currentRaw, hasDraft, err := s.LoadDraft(projectID)
	if err != nil {
		return nil, nil, false, err
	}
	if !hasDraft {
		currentRaw = cache.RemoteConfig
	}

	currentCfg, err := firebase.ParseRemoteConfig(currentRaw)
	if err != nil {
		return nil, nil, false, fmt.Errorf("decode current remote config: %w", err)
	}
	finalCfg := cloneRemoteConfig(currentCfg)
	if err := renameParamSlot(finalCfg, paramKey, nextParamKey, NormalizeRemoteConfigGroupKey(groupKey)); err != nil {
		logger.Error("rename parameter failed", "project_id", projectID, "group", groupKey, "param", paramKey, "next_param", nextParamKey, "publish", publish, "err", err)
		return nil, nil, hasDraft, err
	}
	pruneUnusedConditions(finalCfg)
	removeEmptyGroups(finalCfg)
	dropUnknownConditionReferences(finalCfg)

	if reflect.DeepEqual(currentCfg, finalCfg) {
		return nil, nil, hasDraft, fmt.Errorf("parameter not changed")
	}

	finalRaw, err := marshalRemoteConfig(finalCfg)
	if err != nil {
		return nil, nil, hasDraft, err
	}

	if publish {
		updatedRaw, nextETag, err := s.PublishRemoteConfigWithETag(ctx, projectID, finalRaw, cache.ETag)
		if err != nil {
			return nil, nil, hasDraft, err
		}
		if err := config.DeleteDraft(projectID); err != nil {
			logger.Warn("remove draft after publish failed", "project_id", projectID, "err", err)
		}
		updatedCache := &config.ParametersCache{
			ETag:         nextETag,
			CachedAt:     time.Now().UTC(),
			RemoteConfig: updatedRaw,
		}
		tree, err := s.BuildParametersTree(updatedCache)
		return updatedCache, tree, false, err
	}

	if err := config.SaveDraft(projectID, finalRaw); err != nil {
		return nil, nil, hasDraft, err
	}
	tree, err := s.BuildParametersTreeFromRaw(finalRaw, cache.CachedAt, cache.ETag)
	return cache, tree, true, err
}

// RenameGroup handles rename group for Core and returns the resulting state or error.
func (s *Core) RenameGroup(ctx context.Context, projectID, groupKey, nextGroupKey string, publish bool) (*ParametersCache, *ParametersTree, bool, error) {
	logger := corelog.For("core")
	cache, source, err := s.GetParameters(ctx, projectID, false)
	_ = source
	if err != nil {
		return nil, nil, false, err
	}

	currentRaw, hasDraft, err := s.LoadDraft(projectID)
	if err != nil {
		return nil, nil, false, err
	}
	if !hasDraft {
		currentRaw = cache.RemoteConfig
	}

	currentCfg, err := firebase.ParseRemoteConfig(currentRaw)
	if err != nil {
		return nil, nil, false, fmt.Errorf("decode current remote config: %w", err)
	}
	finalCfg := cloneRemoteConfig(currentCfg)
	if err := renameGroupSlot(finalCfg, NormalizeRemoteConfigGroupKey(groupKey), NormalizeRemoteConfigGroupKey(nextGroupKey)); err != nil {
		logger.Error("rename group failed", "project_id", projectID, "group", groupKey, "next_group", nextGroupKey, "publish", publish, "err", err)
		return nil, nil, hasDraft, err
	}
	pruneUnusedConditions(finalCfg)
	removeEmptyGroups(finalCfg)
	dropUnknownConditionReferences(finalCfg)

	if reflect.DeepEqual(currentCfg, finalCfg) {
		return nil, nil, hasDraft, fmt.Errorf("group not changed")
	}

	finalRaw, err := marshalRemoteConfig(finalCfg)
	if err != nil {
		return nil, nil, hasDraft, err
	}

	if publish {
		updatedRaw, nextETag, err := s.PublishRemoteConfigWithETag(ctx, projectID, finalRaw, cache.ETag)
		if err != nil {
			return nil, nil, hasDraft, err
		}
		if err := config.DeleteDraft(projectID); err != nil {
			logger.Warn("remove draft after publish failed", "project_id", projectID, "err", err)
		}
		updatedCache := &config.ParametersCache{
			ETag:         nextETag,
			CachedAt:     time.Now().UTC(),
			RemoteConfig: updatedRaw,
		}
		tree, err := s.BuildParametersTree(updatedCache)
		return updatedCache, tree, false, err
	}

	if err := config.SaveDraft(projectID, finalRaw); err != nil {
		return nil, nil, hasDraft, err
	}
	tree, err := s.BuildParametersTreeFromRaw(finalRaw, cache.CachedAt, cache.ETag)
	return cache, tree, true, err
}

// MoveParameter moves parameter for Core and returns the resulting state or error.
func (s *Core) MoveParameter(ctx context.Context, projectID, groupKey, paramKey, nextGroupKey string, publish bool) (*ParametersCache, *ParametersTree, bool, error) {
	logger := corelog.For("core")
	cache, source, err := s.GetParameters(ctx, projectID, false)
	_ = source
	if err != nil {
		return nil, nil, false, err
	}

	currentRaw, hasDraft, err := s.LoadDraft(projectID)
	if err != nil {
		return nil, nil, false, err
	}
	if !hasDraft {
		currentRaw = cache.RemoteConfig
	}

	currentCfg, err := firebase.ParseRemoteConfig(currentRaw)
	if err != nil {
		return nil, nil, false, fmt.Errorf("decode current remote config: %w", err)
	}
	finalCfg := cloneRemoteConfig(currentCfg)
	if err := moveParamSlot(finalCfg, paramKey, NormalizeRemoteConfigGroupKey(groupKey), NormalizeRemoteConfigGroupKey(nextGroupKey)); err != nil {
		logger.Error("move parameter failed", "project_id", projectID, "group", groupKey, "param", paramKey, "next_group", nextGroupKey, "publish", publish, "err", err)
		return nil, nil, hasDraft, err
	}
	pruneUnusedConditions(finalCfg)
	removeEmptyGroups(finalCfg)
	dropUnknownConditionReferences(finalCfg)

	if reflect.DeepEqual(currentCfg, finalCfg) {
		return nil, nil, hasDraft, fmt.Errorf("parameter not changed")
	}

	finalRaw, err := marshalRemoteConfig(finalCfg)
	if err != nil {
		return nil, nil, hasDraft, err
	}

	if publish {
		updatedRaw, nextETag, err := s.PublishRemoteConfigWithETag(ctx, projectID, finalRaw, cache.ETag)
		if err != nil {
			return nil, nil, hasDraft, err
		}
		if err := config.DeleteDraft(projectID); err != nil {
			logger.Warn("remove draft after publish failed", "project_id", projectID, "err", err)
		}
		updatedCache := &config.ParametersCache{
			ETag:         nextETag,
			CachedAt:     time.Now().UTC(),
			RemoteConfig: updatedRaw,
		}
		tree, err := s.BuildParametersTree(updatedCache)
		return updatedCache, tree, false, err
	}

	if err := config.SaveDraft(projectID, finalRaw); err != nil {
		return nil, nil, hasDraft, err
	}
	tree, err := s.BuildParametersTreeFromRaw(finalRaw, cache.CachedAt, cache.ETag)
	return cache, tree, true, err
}

// EditParameterDetails handles edit parameter details for Core and returns the resulting state or error.
func (s *Core) EditParameterDetails(ctx context.Context, projectID string, edit ParameterDetailsEdit, publish bool) (*ParametersCache, *ParametersTree, bool, error) {
	logger := corelog.For("core")
	cache, source, err := s.GetParameters(ctx, projectID, false)
	_ = source
	if err != nil {
		return nil, nil, false, err
	}

	currentRaw, hasDraft, err := s.LoadDraft(projectID)
	if err != nil {
		return nil, nil, false, err
	}
	if !hasDraft {
		currentRaw = cache.RemoteConfig
	}

	currentCfg, err := firebase.ParseRemoteConfig(currentRaw)
	if err != nil {
		return nil, nil, false, fmt.Errorf("decode current remote config: %w", err)
	}
	finalCfg := cloneRemoteConfig(currentCfg)
	if err := applyParameterDetailsEdit(finalCfg, edit); err != nil {
		logger.Error("edit parameter details failed", "project_id", projectID, "group", edit.GroupKey, "param", edit.ParamKey, "next_group", edit.NextGroupKey, "next_param", edit.NextParamKey, "next_type", edit.NextValueType, "publish", publish, "err", err)
		return nil, nil, hasDraft, err
	}
	pruneUnusedConditions(finalCfg)
	removeEmptyGroups(finalCfg)
	dropUnknownConditionReferences(finalCfg)

	if reflect.DeepEqual(currentCfg, finalCfg) {
		return nil, nil, hasDraft, fmt.Errorf("parameter not changed")
	}

	finalRaw, err := marshalRemoteConfig(finalCfg)
	if err != nil {
		return nil, nil, hasDraft, err
	}

	if publish {
		updatedRaw, nextETag, err := s.PublishRemoteConfigWithETag(ctx, projectID, finalRaw, cache.ETag)
		if err != nil {
			return nil, nil, hasDraft, err
		}
		if err := config.DeleteDraft(projectID); err != nil {
			logger.Warn("remove draft after publish failed", "project_id", projectID, "err", err)
		}
		updatedCache := &config.ParametersCache{
			ETag:         nextETag,
			CachedAt:     time.Now().UTC(),
			RemoteConfig: updatedRaw,
		}
		tree, err := s.BuildParametersTree(updatedCache)
		return updatedCache, tree, false, err
	}

	if err := config.SaveDraft(projectID, finalRaw); err != nil {
		return nil, nil, hasDraft, err
	}
	tree, err := s.BuildParametersTreeFromRaw(finalRaw, cache.CachedAt, cache.ETag)
	return cache, tree, true, err
}

// MoveGroup moves group for Core and returns the resulting state or error.
func (s *Core) MoveGroup(ctx context.Context, projectID, groupKey, nextGroupKey string, publish bool) (*ParametersCache, *ParametersTree, bool, error) {
	logger := corelog.For("core")
	cache, source, err := s.GetParameters(ctx, projectID, false)
	_ = source
	if err != nil {
		return nil, nil, false, err
	}

	currentRaw, hasDraft, err := s.LoadDraft(projectID)
	if err != nil {
		return nil, nil, false, err
	}
	if !hasDraft {
		currentRaw = cache.RemoteConfig
	}

	currentCfg, err := firebase.ParseRemoteConfig(currentRaw)
	if err != nil {
		return nil, nil, false, fmt.Errorf("decode current remote config: %w", err)
	}
	finalCfg := cloneRemoteConfig(currentCfg)
	if err := moveGroupSlot(finalCfg, NormalizeRemoteConfigGroupKey(groupKey), NormalizeRemoteConfigGroupKey(nextGroupKey)); err != nil {
		logger.Error("move group failed", "project_id", projectID, "group", groupKey, "next_group", nextGroupKey, "publish", publish, "err", err)
		return nil, nil, hasDraft, err
	}
	pruneUnusedConditions(finalCfg)
	removeEmptyGroups(finalCfg)
	dropUnknownConditionReferences(finalCfg)

	if reflect.DeepEqual(currentCfg, finalCfg) {
		return nil, nil, hasDraft, fmt.Errorf("group not changed")
	}

	finalRaw, err := marshalRemoteConfig(finalCfg)
	if err != nil {
		return nil, nil, hasDraft, err
	}

	if publish {
		updatedRaw, nextETag, err := s.PublishRemoteConfigWithETag(ctx, projectID, finalRaw, cache.ETag)
		if err != nil {
			return nil, nil, hasDraft, err
		}
		if err := config.DeleteDraft(projectID); err != nil {
			logger.Warn("remove draft after publish failed", "project_id", projectID, "err", err)
		}
		updatedCache := &config.ParametersCache{
			ETag:         nextETag,
			CachedAt:     time.Now().UTC(),
			RemoteConfig: updatedRaw,
		}
		tree, err := s.BuildParametersTree(updatedCache)
		return updatedCache, tree, false, err
	}

	if err := config.SaveDraft(projectID, finalRaw); err != nil {
		return nil, nil, hasDraft, err
	}
	tree, err := s.BuildParametersTreeFromRaw(finalRaw, cache.CachedAt, cache.ETag)
	return cache, tree, true, err
}

// SetBooleanParameterValue sets boolean parameter value for Core and returns the resulting state or error.
func (s *Core) SetBooleanParameterValue(ctx context.Context, projectID, groupKey, paramKey, valueLabel string, nextValue, publish bool) (*ParametersCache, *ParametersTree, bool, error) {
	logger := corelog.For("core")
	cache, source, err := s.GetParameters(ctx, projectID, false)
	_ = source
	if err != nil {
		return nil, nil, false, err
	}

	currentRaw, hasDraft, err := s.LoadDraft(projectID)
	if err != nil {
		return nil, nil, false, err
	}
	if !hasDraft {
		currentRaw = cache.RemoteConfig
	}

	currentCfg, err := firebase.ParseRemoteConfig(currentRaw)
	if err != nil {
		return nil, nil, false, fmt.Errorf("decode current remote config: %w", err)
	}
	finalCfg := cloneRemoteConfig(currentCfg)
	if err := setBooleanParamValueSlot(finalCfg, paramKey, NormalizeRemoteConfigGroupKey(groupKey), valueLabel, nextValue); err != nil {
		logger.Error("set boolean parameter value failed", "project_id", projectID, "group", groupKey, "param", paramKey, "value_label", valueLabel, "next_value", nextValue, "publish", publish, "err", err)
		return nil, nil, hasDraft, err
	}
	pruneUnusedConditions(finalCfg)
	removeEmptyGroups(finalCfg)
	dropUnknownConditionReferences(finalCfg)

	if reflect.DeepEqual(currentCfg, finalCfg) {
		return nil, nil, hasDraft, fmt.Errorf("parameter value not changed")
	}

	finalRaw, err := marshalRemoteConfig(finalCfg)
	if err != nil {
		return nil, nil, hasDraft, err
	}

	if publish {
		updatedRaw, nextETag, err := s.PublishRemoteConfigWithETag(ctx, projectID, finalRaw, cache.ETag)
		if err != nil {
			return nil, nil, hasDraft, err
		}
		if err := config.DeleteDraft(projectID); err != nil {
			logger.Warn("remove draft after publish failed", "project_id", projectID, "err", err)
		}
		updatedCache := &config.ParametersCache{
			ETag:         nextETag,
			CachedAt:     time.Now().UTC(),
			RemoteConfig: updatedRaw,
		}
		tree, err := s.BuildParametersTree(updatedCache)
		return updatedCache, tree, false, err
	}

	if err := config.SaveDraft(projectID, finalRaw); err != nil {
		return nil, nil, hasDraft, err
	}
	tree, err := s.BuildParametersTreeFromRaw(finalRaw, cache.CachedAt, cache.ETag)
	return cache, tree, true, err
}

// SetNumberParameterValue sets number parameter value for Core and returns the resulting state or error.
func (s *Core) SetNumberParameterValue(ctx context.Context, projectID, groupKey, paramKey, valueLabel, nextValue string, publish bool) (*ParametersCache, *ParametersTree, bool, error) {
	logger := corelog.For("core")
	cache, source, err := s.GetParameters(ctx, projectID, false)
	_ = source
	if err != nil {
		return nil, nil, false, err
	}

	currentRaw, hasDraft, err := s.LoadDraft(projectID)
	if err != nil {
		return nil, nil, false, err
	}
	if !hasDraft {
		currentRaw = cache.RemoteConfig
	}

	currentCfg, err := firebase.ParseRemoteConfig(currentRaw)
	if err != nil {
		return nil, nil, false, fmt.Errorf("decode current remote config: %w", err)
	}
	finalCfg := cloneRemoteConfig(currentCfg)
	if err := setNumberParamValueSlot(finalCfg, paramKey, NormalizeRemoteConfigGroupKey(groupKey), valueLabel, nextValue); err != nil {
		logger.Error("set number parameter value failed", "project_id", projectID, "group", groupKey, "param", paramKey, "value_label", valueLabel, "next_value", nextValue, "publish", publish, "err", err)
		return nil, nil, hasDraft, err
	}
	pruneUnusedConditions(finalCfg)
	removeEmptyGroups(finalCfg)
	dropUnknownConditionReferences(finalCfg)

	if reflect.DeepEqual(currentCfg, finalCfg) {
		return nil, nil, hasDraft, fmt.Errorf("parameter value not changed")
	}

	finalRaw, err := marshalRemoteConfig(finalCfg)
	if err != nil {
		return nil, nil, hasDraft, err
	}

	if publish {
		updatedRaw, nextETag, err := s.PublishRemoteConfigWithETag(ctx, projectID, finalRaw, cache.ETag)
		if err != nil {
			return nil, nil, hasDraft, err
		}
		if err := config.DeleteDraft(projectID); err != nil {
			logger.Warn("remove draft after publish failed", "project_id", projectID, "err", err)
		}
		updatedCache := &config.ParametersCache{
			ETag:         nextETag,
			CachedAt:     time.Now().UTC(),
			RemoteConfig: updatedRaw,
		}
		tree, err := s.BuildParametersTree(updatedCache)
		return updatedCache, tree, false, err
	}

	if err := config.SaveDraft(projectID, finalRaw); err != nil {
		return nil, nil, hasDraft, err
	}
	tree, err := s.BuildParametersTreeFromRaw(finalRaw, cache.CachedAt, cache.ETag)
	return cache, tree, true, err
}

// SetStringParameterValue sets string parameter value for Core and returns the resulting state or error.
func (s *Core) SetStringParameterValue(ctx context.Context, projectID, groupKey, paramKey, valueLabel, nextValue string, publish bool) (*ParametersCache, *ParametersTree, bool, error) {
	logger := corelog.For("core")
	cache, source, err := s.GetParameters(ctx, projectID, false)
	_ = source
	if err != nil {
		return nil, nil, false, err
	}

	currentRaw, hasDraft, err := s.LoadDraft(projectID)
	if err != nil {
		return nil, nil, false, err
	}
	if !hasDraft {
		currentRaw = cache.RemoteConfig
	}

	currentCfg, err := firebase.ParseRemoteConfig(currentRaw)
	if err != nil {
		return nil, nil, false, fmt.Errorf("decode current remote config: %w", err)
	}
	finalCfg := cloneRemoteConfig(currentCfg)
	if err := setStringParamValueSlot(finalCfg, paramKey, NormalizeRemoteConfigGroupKey(groupKey), valueLabel, nextValue); err != nil {
		logger.Error("set string parameter value failed", "project_id", projectID, "group", groupKey, "param", paramKey, "value_label", valueLabel, "publish", publish, "err", err)
		return nil, nil, hasDraft, err
	}
	pruneUnusedConditions(finalCfg)
	removeEmptyGroups(finalCfg)
	dropUnknownConditionReferences(finalCfg)

	if reflect.DeepEqual(currentCfg, finalCfg) {
		return nil, nil, hasDraft, fmt.Errorf("parameter value not changed")
	}

	finalRaw, err := marshalRemoteConfig(finalCfg)
	if err != nil {
		return nil, nil, hasDraft, err
	}

	if publish {
		updatedRaw, nextETag, err := s.PublishRemoteConfigWithETag(ctx, projectID, finalRaw, cache.ETag)
		if err != nil {
			return nil, nil, hasDraft, err
		}
		if err := config.DeleteDraft(projectID); err != nil {
			logger.Warn("remove draft after publish failed", "project_id", projectID, "err", err)
		}
		updatedCache := &config.ParametersCache{
			ETag:         nextETag,
			CachedAt:     time.Now().UTC(),
			RemoteConfig: updatedRaw,
		}
		tree, err := s.BuildParametersTree(updatedCache)
		return updatedCache, tree, false, err
	}

	if err := config.SaveDraft(projectID, finalRaw); err != nil {
		return nil, nil, hasDraft, err
	}
	tree, err := s.BuildParametersTreeFromRaw(finalRaw, cache.CachedAt, cache.ETag)
	return cache, tree, true, err
}

// SetJSONParameterValue sets jsonparameter value for Core and returns the resulting state or error.
func (s *Core) SetJSONParameterValue(ctx context.Context, projectID, groupKey, paramKey, valueLabel, nextValue string, publish bool) (*ParametersCache, *ParametersTree, bool, error) {
	logger := corelog.For("core")
	cache, source, err := s.GetParameters(ctx, projectID, false)
	_ = source
	if err != nil {
		return nil, nil, false, err
	}

	currentRaw, hasDraft, err := s.LoadDraft(projectID)
	if err != nil {
		return nil, nil, false, err
	}
	if !hasDraft {
		currentRaw = cache.RemoteConfig
	}

	currentCfg, err := firebase.ParseRemoteConfig(currentRaw)
	if err != nil {
		return nil, nil, false, fmt.Errorf("decode current remote config: %w", err)
	}
	finalCfg := cloneRemoteConfig(currentCfg)
	if err := setJSONParamValueSlot(finalCfg, paramKey, NormalizeRemoteConfigGroupKey(groupKey), valueLabel, nextValue); err != nil {
		logger.Error("set json parameter value failed", "project_id", projectID, "group", groupKey, "param", paramKey, "value_label", valueLabel, "publish", publish, "err", err)
		return nil, nil, hasDraft, err
	}
	pruneUnusedConditions(finalCfg)
	removeEmptyGroups(finalCfg)
	dropUnknownConditionReferences(finalCfg)

	if reflect.DeepEqual(currentCfg, finalCfg) {
		return nil, nil, hasDraft, fmt.Errorf("parameter value not changed")
	}

	finalRaw, err := marshalRemoteConfig(finalCfg)
	if err != nil {
		return nil, nil, hasDraft, err
	}

	if publish {
		updatedRaw, nextETag, err := s.PublishRemoteConfigWithETag(ctx, projectID, finalRaw, cache.ETag)
		if err != nil {
			return nil, nil, hasDraft, err
		}
		if err := config.DeleteDraft(projectID); err != nil {
			logger.Warn("remove draft after publish failed", "project_id", projectID, "err", err)
		}
		updatedCache := &config.ParametersCache{
			ETag:         nextETag,
			CachedAt:     time.Now().UTC(),
			RemoteConfig: updatedRaw,
		}
		tree, err := s.BuildParametersTree(updatedCache)
		return updatedCache, tree, false, err
	}

	if err := config.SaveDraft(projectID, finalRaw); err != nil {
		return nil, nil, hasDraft, err
	}
	tree, err := s.BuildParametersTreeFromRaw(finalRaw, cache.CachedAt, cache.ETag)
	return cache, tree, true, err
}

// DuplicateParameter handles duplicate parameter for Core and returns the resulting state or error.
func (s *Core) DuplicateParameter(ctx context.Context, projectID, groupKey, paramKey string) (*ParametersCache, *ParametersTree, bool, string, error) {
	cache, source, err := s.GetParameters(ctx, projectID, false)
	_ = source
	if err != nil {
		return nil, nil, false, "", err
	}

	currentRaw, hasDraft, err := s.LoadDraft(projectID)
	if err != nil {
		return nil, nil, false, "", err
	}
	if !hasDraft {
		currentRaw = cache.RemoteConfig
	}

	currentCfg, err := firebase.ParseRemoteConfig(currentRaw)
	if err != nil {
		return nil, nil, false, "", fmt.Errorf("decode current remote config: %w", err)
	}
	finalCfg := cloneRemoteConfig(currentCfg)
	nextParamKey, err := duplicateParamSlot(finalCfg, paramKey, NormalizeRemoteConfigGroupKey(groupKey))
	if err != nil {
		return nil, nil, hasDraft, "", err
	}
	pruneUnusedConditions(finalCfg)
	removeEmptyGroups(finalCfg)
	dropUnknownConditionReferences(finalCfg)

	finalRaw, err := marshalRemoteConfig(finalCfg)
	if err != nil {
		return nil, nil, hasDraft, "", err
	}
	if err := config.SaveDraft(projectID, finalRaw); err != nil {
		return nil, nil, hasDraft, "", err
	}
	tree, err := s.BuildParametersTreeFromRaw(finalRaw, cache.CachedAt, cache.ETag)
	return cache, tree, true, nextParamKey, err
}

// DuplicateParameterNamed handles duplicate parameter named for Core and returns the resulting state or error.
func (s *Core) DuplicateParameterNamed(ctx context.Context, projectID, groupKey, paramKey, nextParamKey string, publish bool) (*ParametersCache, *ParametersTree, bool, error) {
	logger := corelog.For("core")
	cache, source, err := s.GetParameters(ctx, projectID, false)
	_ = source
	if err != nil {
		return nil, nil, false, err
	}

	currentRaw, hasDraft, err := s.LoadDraft(projectID)
	if err != nil {
		return nil, nil, false, err
	}
	if !hasDraft {
		currentRaw = cache.RemoteConfig
	}

	currentCfg, err := firebase.ParseRemoteConfig(currentRaw)
	if err != nil {
		return nil, nil, false, fmt.Errorf("decode current remote config: %w", err)
	}
	finalCfg := cloneRemoteConfig(currentCfg)
	if err := duplicateParamSlotAs(finalCfg, paramKey, nextParamKey, NormalizeRemoteConfigGroupKey(groupKey)); err != nil {
		logger.Error("duplicate parameter failed", "project_id", projectID, "group", groupKey, "param", paramKey, "next_param", nextParamKey, "publish", publish, "err", err)
		return nil, nil, hasDraft, err
	}
	pruneUnusedConditions(finalCfg)
	removeEmptyGroups(finalCfg)
	dropUnknownConditionReferences(finalCfg)

	finalRaw, err := marshalRemoteConfig(finalCfg)
	if err != nil {
		return nil, nil, hasDraft, err
	}
	if publish {
		updatedRaw, nextETag, err := s.PublishRemoteConfigWithETag(ctx, projectID, finalRaw, cache.ETag)
		if err != nil {
			return nil, nil, hasDraft, err
		}
		if err := config.DeleteDraft(projectID); err != nil {
			logger.Warn("remove draft after publish failed", "project_id", projectID, "err", err)
		}
		updatedCache := &config.ParametersCache{
			ETag:         nextETag,
			CachedAt:     time.Now().UTC(),
			RemoteConfig: updatedRaw,
		}
		tree, err := s.BuildParametersTree(updatedCache)
		return updatedCache, tree, false, err
	}
	if err := config.SaveDraft(projectID, finalRaw); err != nil {
		return nil, nil, hasDraft, err
	}
	tree, err := s.BuildParametersTreeFromRaw(finalRaw, cache.CachedAt, cache.ETag)
	return cache, tree, true, err
}

// PreviewDeleteParameter handles preview delete parameter for Core and returns the resulting state or error.
func (s *Core) PreviewDeleteParameter(projectID, groupKey, paramKey string) (*ParametersCache, json.RawMessage, error) {
	cache, _, err := s.InspectParametersCache(projectID)
	if err != nil {
		return nil, nil, err
	}
	if cache == nil {
		return nil, nil, fmt.Errorf("parameters cache not found")
	}

	currentRaw := cache.RemoteConfig
	draftRaw, hasDraft, err := s.LoadDraft(projectID)
	if err != nil {
		return nil, nil, err
	}
	if hasDraft {
		currentRaw = draftRaw
	}

	currentCfg, err := firebase.ParseRemoteConfig(currentRaw)
	if err != nil {
		return nil, nil, fmt.Errorf("decode current remote config: %w", err)
	}

	finalCfg := cloneRemoteConfig(currentCfg)
	removeParamSlot(finalCfg, paramKey, NormalizeRemoteConfigGroupKey(groupKey))
	pruneUnusedConditions(finalCfg)
	removeEmptyGroups(finalCfg)
	dropUnknownConditionReferences(finalCfg)

	if reflect.DeepEqual(currentCfg, finalCfg) {
		return nil, nil, fmt.Errorf("parameter not found")
	}

	finalRaw, err := marshalRemoteConfig(finalCfg)
	if err != nil {
		return nil, nil, err
	}
	return cache, finalRaw, nil
}

// PreviewDeleteGroup handles preview delete group for Core and returns the resulting state or error.
func (s *Core) PreviewDeleteGroup(projectID, groupKey string) (*ParametersCache, json.RawMessage, error) {
	logger := corelog.For("core")
	cache, _, err := s.InspectParametersCache(projectID)
	if err != nil {
		return nil, nil, err
	}
	if cache == nil {
		return nil, nil, fmt.Errorf("parameters cache not found")
	}

	currentRaw := cache.RemoteConfig
	draftRaw, hasDraft, err := s.LoadDraft(projectID)
	if err != nil {
		return nil, nil, err
	}
	if hasDraft {
		currentRaw = draftRaw
	}

	currentCfg, err := firebase.ParseRemoteConfig(currentRaw)
	if err != nil {
		return nil, nil, fmt.Errorf("decode current remote config: %w", err)
	}

	finalCfg := cloneRemoteConfig(currentCfg)
	if err := removeGroupSlot(finalCfg, NormalizeRemoteConfigGroupKey(groupKey)); err != nil {
		logger.Error("preview delete group failed", "project_id", projectID, "group", groupKey, "err", err)
		return nil, nil, err
	}
	pruneUnusedConditions(finalCfg)
	removeEmptyGroups(finalCfg)
	dropUnknownConditionReferences(finalCfg)

	if reflect.DeepEqual(currentCfg, finalCfg) {
		return nil, nil, fmt.Errorf("group not changed")
	}

	finalRaw, err := marshalRemoteConfig(finalCfg)
	if err != nil {
		return nil, nil, err
	}
	return cache, finalRaw, nil
}

// PreviewRenameParameter handles preview rename parameter for Core and returns the resulting state or error.
func (s *Core) PreviewRenameParameter(projectID, groupKey, paramKey, nextParamKey string) (*ParametersCache, json.RawMessage, error) {
	logger := corelog.For("core")
	cache, _, err := s.InspectParametersCache(projectID)
	if err != nil {
		return nil, nil, err
	}
	if cache == nil {
		return nil, nil, fmt.Errorf("parameters cache not found")
	}

	currentRaw := cache.RemoteConfig
	draftRaw, hasDraft, err := s.LoadDraft(projectID)
	if err != nil {
		return nil, nil, err
	}
	if hasDraft {
		currentRaw = draftRaw
	}

	currentCfg, err := firebase.ParseRemoteConfig(currentRaw)
	if err != nil {
		return nil, nil, fmt.Errorf("decode current remote config: %w", err)
	}

	finalCfg := cloneRemoteConfig(currentCfg)
	if err := renameParamSlot(finalCfg, paramKey, nextParamKey, NormalizeRemoteConfigGroupKey(groupKey)); err != nil {
		logger.Error("preview rename parameter failed", "project_id", projectID, "group", groupKey, "param", paramKey, "next_param", nextParamKey, "err", err)
		return nil, nil, err
	}
	pruneUnusedConditions(finalCfg)
	removeEmptyGroups(finalCfg)
	dropUnknownConditionReferences(finalCfg)

	if reflect.DeepEqual(currentCfg, finalCfg) {
		return nil, nil, fmt.Errorf("parameter not changed")
	}

	finalRaw, err := marshalRemoteConfig(finalCfg)
	if err != nil {
		return nil, nil, err
	}
	return cache, finalRaw, nil
}

// PreviewRenameGroup handles preview rename group for Core and returns the resulting state or error.
func (s *Core) PreviewRenameGroup(projectID, groupKey, nextGroupKey string) (*ParametersCache, json.RawMessage, error) {
	logger := corelog.For("core")
	cache, _, err := s.InspectParametersCache(projectID)
	if err != nil {
		return nil, nil, err
	}
	if cache == nil {
		return nil, nil, fmt.Errorf("parameters cache not found")
	}

	currentRaw := cache.RemoteConfig
	draftRaw, hasDraft, err := s.LoadDraft(projectID)
	if err != nil {
		return nil, nil, err
	}
	if hasDraft {
		currentRaw = draftRaw
	}

	currentCfg, err := firebase.ParseRemoteConfig(currentRaw)
	if err != nil {
		return nil, nil, fmt.Errorf("decode current remote config: %w", err)
	}

	finalCfg := cloneRemoteConfig(currentCfg)
	if err := renameGroupSlot(finalCfg, NormalizeRemoteConfigGroupKey(groupKey), NormalizeRemoteConfigGroupKey(nextGroupKey)); err != nil {
		logger.Error("preview rename group failed", "project_id", projectID, "group", groupKey, "next_group", nextGroupKey, "err", err)
		return nil, nil, err
	}
	pruneUnusedConditions(finalCfg)
	removeEmptyGroups(finalCfg)
	dropUnknownConditionReferences(finalCfg)

	if reflect.DeepEqual(currentCfg, finalCfg) {
		return nil, nil, fmt.Errorf("group not changed")
	}

	finalRaw, err := marshalRemoteConfig(finalCfg)
	if err != nil {
		return nil, nil, err
	}
	return cache, finalRaw, nil
}

// PreviewMoveParameter handles preview move parameter for Core and returns the resulting state or error.
func (s *Core) PreviewMoveParameter(projectID, groupKey, paramKey, nextGroupKey string) (*ParametersCache, json.RawMessage, error) {
	logger := corelog.For("core")
	cache, _, err := s.InspectParametersCache(projectID)
	if err != nil {
		return nil, nil, err
	}
	if cache == nil {
		return nil, nil, fmt.Errorf("parameters cache not found")
	}

	currentRaw := cache.RemoteConfig
	draftRaw, hasDraft, err := s.LoadDraft(projectID)
	if err != nil {
		return nil, nil, err
	}
	if hasDraft {
		currentRaw = draftRaw
	}

	currentCfg, err := firebase.ParseRemoteConfig(currentRaw)
	if err != nil {
		return nil, nil, fmt.Errorf("decode current remote config: %w", err)
	}

	finalCfg := cloneRemoteConfig(currentCfg)
	if err := moveParamSlot(finalCfg, paramKey, NormalizeRemoteConfigGroupKey(groupKey), NormalizeRemoteConfigGroupKey(nextGroupKey)); err != nil {
		logger.Error("preview move parameter failed", "project_id", projectID, "group", groupKey, "param", paramKey, "next_group", nextGroupKey, "err", err)
		return nil, nil, err
	}
	pruneUnusedConditions(finalCfg)
	removeEmptyGroups(finalCfg)
	dropUnknownConditionReferences(finalCfg)

	if reflect.DeepEqual(currentCfg, finalCfg) {
		return nil, nil, fmt.Errorf("parameter not changed")
	}

	finalRaw, err := marshalRemoteConfig(finalCfg)
	if err != nil {
		return nil, nil, err
	}
	return cache, finalRaw, nil
}

// PreviewEditParameterDetails handles preview edit parameter details for Core and returns the resulting state or error.
func (s *Core) PreviewEditParameterDetails(projectID string, edit ParameterDetailsEdit) (*ParametersCache, json.RawMessage, error) {
	logger := corelog.For("core")
	cache, _, err := s.InspectParametersCache(projectID)
	if err != nil {
		return nil, nil, err
	}
	if cache == nil {
		return nil, nil, fmt.Errorf("parameters cache not found")
	}

	currentRaw := cache.RemoteConfig
	if draftRaw, hasDraft, err := s.LoadDraft(projectID); err != nil {
		return nil, nil, err
	} else if hasDraft {
		currentRaw = draftRaw
	}

	currentCfg, err := firebase.ParseRemoteConfig(currentRaw)
	if err != nil {
		return nil, nil, fmt.Errorf("decode current remote config: %w", err)
	}
	finalCfg := cloneRemoteConfig(currentCfg)
	if err := applyParameterDetailsEdit(finalCfg, edit); err != nil {
		logger.Error("preview edit parameter details failed", "project_id", projectID, "group", edit.GroupKey, "param", edit.ParamKey, "next_group", edit.NextGroupKey, "next_param", edit.NextParamKey, "next_type", edit.NextValueType, "err", err)
		return nil, nil, err
	}
	pruneUnusedConditions(finalCfg)
	removeEmptyGroups(finalCfg)
	dropUnknownConditionReferences(finalCfg)

	finalRaw, err := marshalRemoteConfig(finalCfg)
	if err != nil {
		return nil, nil, err
	}
	return cache, finalRaw, nil
}

// PreviewMoveGroup handles preview move group for Core and returns the resulting state or error.
func (s *Core) PreviewMoveGroup(projectID, groupKey, nextGroupKey string) (*ParametersCache, json.RawMessage, error) {
	logger := corelog.For("core")
	cache, _, err := s.InspectParametersCache(projectID)
	if err != nil {
		return nil, nil, err
	}
	if cache == nil {
		return nil, nil, fmt.Errorf("parameters cache not found")
	}

	currentRaw := cache.RemoteConfig
	draftRaw, hasDraft, err := s.LoadDraft(projectID)
	if err != nil {
		return nil, nil, err
	}
	if hasDraft {
		currentRaw = draftRaw
	}

	currentCfg, err := firebase.ParseRemoteConfig(currentRaw)
	if err != nil {
		return nil, nil, fmt.Errorf("decode current remote config: %w", err)
	}

	finalCfg := cloneRemoteConfig(currentCfg)
	if err := moveGroupSlot(finalCfg, NormalizeRemoteConfigGroupKey(groupKey), NormalizeRemoteConfigGroupKey(nextGroupKey)); err != nil {
		logger.Error("preview move group failed", "project_id", projectID, "group", groupKey, "next_group", nextGroupKey, "err", err)
		return nil, nil, err
	}
	pruneUnusedConditions(finalCfg)
	removeEmptyGroups(finalCfg)
	dropUnknownConditionReferences(finalCfg)

	if reflect.DeepEqual(currentCfg, finalCfg) {
		return nil, nil, fmt.Errorf("group not changed")
	}

	finalRaw, err := marshalRemoteConfig(finalCfg)
	if err != nil {
		return nil, nil, err
	}
	return cache, finalRaw, nil
}

// PreviewSetBooleanParameterValue handles preview set boolean parameter value for Core and returns the resulting state or error.
func (s *Core) PreviewSetBooleanParameterValue(projectID, groupKey, paramKey, valueLabel string, nextValue bool) (*ParametersCache, json.RawMessage, error) {
	logger := corelog.For("core")
	cache, _, err := s.InspectParametersCache(projectID)
	if err != nil {
		return nil, nil, err
	}
	if cache == nil {
		return nil, nil, fmt.Errorf("parameters cache not found")
	}

	currentRaw := cache.RemoteConfig
	draftRaw, hasDraft, err := s.LoadDraft(projectID)
	if err != nil {
		return nil, nil, err
	}
	if hasDraft {
		currentRaw = draftRaw
	}

	currentCfg, err := firebase.ParseRemoteConfig(currentRaw)
	if err != nil {
		return nil, nil, fmt.Errorf("decode current remote config: %w", err)
	}

	finalCfg := cloneRemoteConfig(currentCfg)
	if err := setBooleanParamValueSlot(finalCfg, paramKey, NormalizeRemoteConfigGroupKey(groupKey), valueLabel, nextValue); err != nil {
		logger.Error("preview set boolean parameter value failed", "project_id", projectID, "group", groupKey, "param", paramKey, "value_label", valueLabel, "next_value", nextValue, "err", err)
		return nil, nil, err
	}
	pruneUnusedConditions(finalCfg)
	removeEmptyGroups(finalCfg)
	dropUnknownConditionReferences(finalCfg)

	if reflect.DeepEqual(currentCfg, finalCfg) {
		return nil, nil, fmt.Errorf("parameter value not changed")
	}

	finalRaw, err := marshalRemoteConfig(finalCfg)
	if err != nil {
		return nil, nil, err
	}
	return cache, finalRaw, nil
}

// PreviewSetNumberParameterValue handles preview set number parameter value for Core and returns the resulting state or error.
func (s *Core) PreviewSetNumberParameterValue(projectID, groupKey, paramKey, valueLabel, nextValue string) (*ParametersCache, json.RawMessage, error) {
	logger := corelog.For("core")
	cache, _, err := s.InspectParametersCache(projectID)
	if err != nil {
		return nil, nil, err
	}
	if cache == nil {
		return nil, nil, fmt.Errorf("parameters cache not found")
	}

	currentRaw := cache.RemoteConfig
	draftRaw, hasDraft, err := s.LoadDraft(projectID)
	if err != nil {
		return nil, nil, err
	}
	if hasDraft {
		currentRaw = draftRaw
	}

	currentCfg, err := firebase.ParseRemoteConfig(currentRaw)
	if err != nil {
		return nil, nil, fmt.Errorf("decode current remote config: %w", err)
	}

	finalCfg := cloneRemoteConfig(currentCfg)
	if err := setNumberParamValueSlot(finalCfg, paramKey, NormalizeRemoteConfigGroupKey(groupKey), valueLabel, nextValue); err != nil {
		logger.Error("preview set number parameter value failed", "project_id", projectID, "group", groupKey, "param", paramKey, "value_label", valueLabel, "next_value", nextValue, "err", err)
		return nil, nil, err
	}
	pruneUnusedConditions(finalCfg)
	removeEmptyGroups(finalCfg)
	dropUnknownConditionReferences(finalCfg)

	if reflect.DeepEqual(currentCfg, finalCfg) {
		return nil, nil, fmt.Errorf("parameter value not changed")
	}

	finalRaw, err := marshalRemoteConfig(finalCfg)
	if err != nil {
		return nil, nil, err
	}
	return cache, finalRaw, nil
}

// PreviewSetStringParameterValue handles preview set string parameter value for Core and returns the resulting state or error.
func (s *Core) PreviewSetStringParameterValue(projectID, groupKey, paramKey, valueLabel, nextValue string) (*ParametersCache, json.RawMessage, error) {
	logger := corelog.For("core")
	cache, _, err := s.InspectParametersCache(projectID)
	if err != nil {
		return nil, nil, err
	}
	if cache == nil {
		return nil, nil, fmt.Errorf("parameters cache not found")
	}

	currentRaw := cache.RemoteConfig
	draftRaw, hasDraft, err := s.LoadDraft(projectID)
	if err != nil {
		return nil, nil, err
	}
	if hasDraft {
		currentRaw = draftRaw
	}

	currentCfg, err := firebase.ParseRemoteConfig(currentRaw)
	if err != nil {
		return nil, nil, fmt.Errorf("decode current remote config: %w", err)
	}

	finalCfg := cloneRemoteConfig(currentCfg)
	if err := setStringParamValueSlot(finalCfg, paramKey, NormalizeRemoteConfigGroupKey(groupKey), valueLabel, nextValue); err != nil {
		logger.Error("preview set string parameter value failed", "project_id", projectID, "group", groupKey, "param", paramKey, "value_label", valueLabel, "err", err)
		return nil, nil, err
	}
	pruneUnusedConditions(finalCfg)
	removeEmptyGroups(finalCfg)
	dropUnknownConditionReferences(finalCfg)

	if reflect.DeepEqual(currentCfg, finalCfg) {
		return nil, nil, fmt.Errorf("parameter value not changed")
	}

	finalRaw, err := marshalRemoteConfig(finalCfg)
	if err != nil {
		return nil, nil, err
	}
	return cache, finalRaw, nil
}

// PreviewSetJSONParameterValue handles preview set jsonparameter value for Core and returns the resulting state or error.
func (s *Core) PreviewSetJSONParameterValue(projectID, groupKey, paramKey, valueLabel, nextValue string) (*ParametersCache, json.RawMessage, error) {
	logger := corelog.For("core")
	cache, _, err := s.InspectParametersCache(projectID)
	if err != nil {
		return nil, nil, err
	}
	if cache == nil {
		return nil, nil, fmt.Errorf("parameters cache not found")
	}

	currentRaw := cache.RemoteConfig
	draftRaw, hasDraft, err := s.LoadDraft(projectID)
	if err != nil {
		return nil, nil, err
	}
	if hasDraft {
		currentRaw = draftRaw
	}

	currentCfg, err := firebase.ParseRemoteConfig(currentRaw)
	if err != nil {
		return nil, nil, fmt.Errorf("decode current remote config: %w", err)
	}

	finalCfg := cloneRemoteConfig(currentCfg)
	if err := setJSONParamValueSlot(finalCfg, paramKey, NormalizeRemoteConfigGroupKey(groupKey), valueLabel, nextValue); err != nil {
		logger.Error("preview set json parameter value failed", "project_id", projectID, "group", groupKey, "param", paramKey, "value_label", valueLabel, "err", err)
		return nil, nil, err
	}
	pruneUnusedConditions(finalCfg)
	removeEmptyGroups(finalCfg)
	dropUnknownConditionReferences(finalCfg)

	if reflect.DeepEqual(currentCfg, finalCfg) {
		return nil, nil, fmt.Errorf("parameter value not changed")
	}

	finalRaw, err := marshalRemoteConfig(finalCfg)
	if err != nil {
		return nil, nil, err
	}
	return cache, finalRaw, nil
}

// PreviewDuplicateParameter handles preview duplicate parameter for Core and returns the resulting state or error.
func (s *Core) PreviewDuplicateParameter(projectID, groupKey, paramKey, nextParamKey string) (*ParametersCache, json.RawMessage, error) {
	logger := corelog.For("core")
	cache, _, err := s.InspectParametersCache(projectID)
	if err != nil {
		return nil, nil, err
	}
	if cache == nil {
		return nil, nil, fmt.Errorf("parameters cache not found")
	}

	currentRaw := cache.RemoteConfig
	draftRaw, hasDraft, err := s.LoadDraft(projectID)
	if err != nil {
		return nil, nil, err
	}
	if hasDraft {
		currentRaw = draftRaw
	}

	currentCfg, err := firebase.ParseRemoteConfig(currentRaw)
	if err != nil {
		return nil, nil, fmt.Errorf("decode current remote config: %w", err)
	}

	finalCfg := cloneRemoteConfig(currentCfg)
	if err := duplicateParamSlotAs(finalCfg, paramKey, nextParamKey, NormalizeRemoteConfigGroupKey(groupKey)); err != nil {
		logger.Error("preview duplicate parameter failed", "project_id", projectID, "group", groupKey, "param", paramKey, "next_param", nextParamKey, "err", err)
		return nil, nil, err
	}
	pruneUnusedConditions(finalCfg)
	removeEmptyGroups(finalCfg)
	dropUnknownConditionReferences(finalCfg)

	if reflect.DeepEqual(currentCfg, finalCfg) {
		return nil, nil, fmt.Errorf("parameter not changed")
	}

	finalRaw, err := marshalRemoteConfig(finalCfg)
	if err != nil {
		return nil, nil, err
	}
	return cache, finalRaw, nil
}

// PublishDraft handles publish draft for Core and returns the resulting state or error.
func (s *Core) PublishDraft(ctx context.Context, projectID string) (*ParametersCache, *ParametersTree, error) {
	logger := corelog.For("core")
	cache, _, err := s.GetParameters(ctx, projectID, false)
	if err != nil {
		return nil, nil, err
	}
	draftRaw, hasDraft, err := s.LoadDraft(projectID)
	if err != nil {
		return nil, nil, err
	}
	if !hasDraft {
		return nil, nil, fmt.Errorf("draft not found")
	}

	updatedRaw, nextETag, err := s.PublishRemoteConfigWithETag(ctx, projectID, draftRaw, cache.ETag)
	if err != nil {
		return nil, nil, err
	}
	if err := config.DeleteDraft(projectID); err != nil {
		logger.Warn("remove draft after publish failed", "project_id", projectID, "err", err)
	}

	updatedCache := &config.ParametersCache{
		ETag:         nextETag,
		CachedAt:     time.Now().UTC(),
		RemoteConfig: updatedRaw,
	}
	tree, err := s.BuildParametersTree(updatedCache)
	return updatedCache, tree, err
}

// DiscardDraft handles discard draft for Core and returns the resulting state or error.
func (s *Core) DiscardDraft(ctx context.Context, projectID string) (*ParametersCache, *ParametersTree, error) {
	cache, _, err := s.GetParameters(ctx, projectID, false)
	if err != nil {
		return nil, nil, err
	}
	if err := config.DeleteDraft(projectID); err != nil {
		return nil, nil, err
	}
	tree, err := s.BuildParametersTree(cache)
	return cache, tree, err
}

// RefreshDraftAwareParameters handles refresh draft aware parameters for Core and returns the resulting state or error.
func (s *Core) RefreshDraftAwareParameters(ctx context.Context, projectID string) (*ParametersCache, *ParametersTree, string, bool, bool, error) {
	logger := corelog.For("core")
	previousCache, _, err := s.InspectParametersCache(projectID)
	if err != nil {
		return nil, nil, "", false, false, fmt.Errorf("inspect parameters cache: %w", err)
	}

	cache, source, err := s.GetParameters(ctx, projectID, true)
	if err != nil {
		return nil, nil, "", false, false, err
	}

	draftRaw, hasDraft, err := s.LoadDraft(projectID)
	if err != nil {
		return nil, nil, "", false, false, err
	}
	if !hasDraft {
		tree, err := s.BuildParametersTree(cache)
		return cache, tree, source, false, false, err
	}

	if previousCache == nil || bytes.Equal(previousCache.RemoteConfig, cache.RemoteConfig) {
		tree, err := s.BuildParametersTreeFromRaw(draftRaw, cache.CachedAt, cache.ETag)
		return cache, tree, "draft", true, false, err
	}

	mergedRaw, hasChanges, err := mergeDraftWithLatest(previousCache.RemoteConfig, draftRaw, cache.RemoteConfig)
	if err != nil {
		logger.Error("merge draft with latest failed", "project_id", projectID, "err", err)
		tree, treeErr := s.BuildParametersTreeFromRaw(draftRaw, previousCache.CachedAt, previousCache.ETag)
		return previousCache, tree, "draft-stale", true, true, treeErr
	}
	if !hasChanges {
		if err := config.DeleteDraft(projectID); err != nil {
			logger.Warn("remove obsolete draft failed", "project_id", projectID, "err", err)
		}
		tree, err := s.BuildParametersTree(cache)
		return cache, tree, source, false, false, err
	}
	if err := config.SaveDraft(projectID, mergedRaw); err != nil {
		return nil, nil, "", false, false, err
	}
	tree, err := s.BuildParametersTreeFromRaw(mergedRaw, cache.CachedAt, cache.ETag)
	return cache, tree, "draft", true, false, err
}

// mergeDraftWithLatest handles merge draft with latest and returns the resulting value or error.
func mergeDraftWithLatest(baseRaw, draftRaw, latestRaw json.RawMessage) (json.RawMessage, bool, error) {
	baseCfg, err := firebase.ParseRemoteConfig(baseRaw)
	if err != nil {
		return nil, false, fmt.Errorf("decode base remote config: %w", err)
	}
	draftCfg, err := firebase.ParseRemoteConfig(draftRaw)
	if err != nil {
		return nil, false, fmt.Errorf("decode draft remote config: %w", err)
	}
	latestCfg, err := firebase.ParseRemoteConfig(latestRaw)
	if err != nil {
		return nil, false, fmt.Errorf("decode latest remote config: %w", err)
	}

	merged := cloneRemoteConfig(latestCfg)
	baseSlots := collectParamSlots(baseCfg)
	draftSlots := collectParamSlots(draftCfg)
	latestSlots := collectParamSlots(latestCfg)

	for _, key := range sortedSlotKeys(baseSlots, draftSlots, latestSlots) {
		baseSlot, inBase := baseSlots[key]
		draftSlot, inDraft := draftSlots[key]
		latestSlot, inLatest := latestSlots[key]

		localChanged := !equalParamState(baseSlot, inBase, draftSlot, inDraft)
		if !localChanged {
			continue
		}
		remoteChanged := !equalParamState(baseSlot, inBase, latestSlot, inLatest)
		if !remoteChanged {
			applyMergedSlot(merged, key, baseSlot, inBase, draftSlot, inDraft)
			continue
		}
		if equalParamState(draftSlot, inDraft, latestSlot, inLatest) {
			continue
		}
		return nil, false, fmt.Errorf("draft conflict on %s", slotDisplayKey(key))
	}
	pruneUnusedConditions(merged)
	removeEmptyGroups(merged)
	dropUnknownConditionReferences(merged)

	if reflect.DeepEqual(latestCfg, merged) {
		return nil, false, nil
	}
	raw, err := marshalRemoteConfig(merged)
	if err != nil {
		return nil, false, err
	}
	return raw, true, nil
}

// collectParamSlots handles collect param slots and returns the resulting value or error.
func collectParamSlots(cfg *firebase.RemoteConfig) map[string]paramSlot {
	out := make(map[string]paramSlot)
	if cfg == nil {
		return out
	}
	for key, param := range cfg.Parameters {
		out[slotKey("", key)] = paramSlot{group: "", param: param}
	}
	for groupName, group := range cfg.ParameterGroups {
		for key, param := range group.Parameters {
			out[slotKey(groupName, key)] = paramSlot{group: groupName, param: param}
		}
	}
	return out
}

// slotKey handles slot key and returns the resulting value or error.
func slotKey(group, key string) string {
	return group + "\x00" + key
}

// slotKeyParam handles slot key param and returns the resulting value or error.
func slotKeyParam(key string) string {
	for i := 0; i < len(key); i++ {
		if key[i] == 0 {
			return key[i+1:]
		}
	}
	return key
}

// slotKeyGroup handles slot key group and returns the resulting value or error.
func slotKeyGroup(key string) string {
	for i := 0; i < len(key); i++ {
		if key[i] == 0 {
			return key[:i]
		}
	}
	return ""
}

// slotDisplayKey handles slot display key and returns the resulting value or error.
func slotDisplayKey(key string) string {
	group := slotKeyGroup(key)
	param := slotKeyParam(key)
	if group == "" {
		return param
	}
	return group + "/" + param
}

// sortedSlotKeys handles sorted slot keys and returns the resulting value or error.
func sortedSlotKeys(maps ...map[string]paramSlot) []string {
	seen := make(map[string]struct{})
	out := make([]string, 0)
	for _, items := range maps {
		for key := range items {
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, key)
		}
	}
	sort.Strings(out)
	return out
}

// equalParamState handles equal param state and returns the resulting value or error.
func equalParamState(left paramSlot, leftOK bool, right paramSlot, rightOK bool) bool {
	if leftOK != rightOK {
		return false
	}
	if !leftOK {
		return true
	}
	return reflect.DeepEqual(left, right)
}

// applyMergedSlot handles apply merged slot and returns the resulting value or error.
func applyMergedSlot(cfg *firebase.RemoteConfig, key string, baseSlot paramSlot, inBase bool, draftSlot paramSlot, inDraft bool) {
	paramKey := slotKeyParam(key)
	if !inDraft {
		group := slotKeyGroup(key)
		if inBase {
			group = baseSlot.group
		}
		removeParamSlot(cfg, paramKey, group)
		return
	}
	if inBase {
		removeParamSlot(cfg, paramKey, baseSlot.group)
	}
	setParamSlot(cfg, paramKey, draftSlot)
}

// cloneRemoteConfig handles clone remote config and returns the resulting value or error.
func cloneRemoteConfig(cfg *firebase.RemoteConfig) *firebase.RemoteConfig {
	if cfg == nil {
		return &firebase.RemoteConfig{}
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		return &firebase.RemoteConfig{}
	}
	var out firebase.RemoteConfig
	if err := json.Unmarshal(data, &out); err != nil {
		return &firebase.RemoteConfig{}
	}
	return &out
}

// marshalRemoteConfig handles marshal remote config and returns the resulting value or error.
func marshalRemoteConfig(cfg *firebase.RemoteConfig) ([]byte, error) {
	data, err := json.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("encode remote config: %w", err)
	}
	return data, nil
}

// removeParamSlot removes remove param slot and returns the resulting value or error.
func removeParamSlot(cfg *firebase.RemoteConfig, key, groupName string) {
	if groupName == "" {
		delete(cfg.Parameters, key)
		return
	}
	group, ok := cfg.ParameterGroups[groupName]
	if !ok {
		return
	}
	delete(group.Parameters, key)
	if len(group.Parameters) == 0 {
		delete(cfg.ParameterGroups, groupName)
		return
	}
	cfg.ParameterGroups[groupName] = group
}

// removeGroupSlot removes remove group slot and returns the resulting value or error.
func removeGroupSlot(cfg *firebase.RemoteConfig, groupName string) error {
	if groupName == "" {
		return fmt.Errorf("default group cannot be removed")
	}
	if _, ok := cfg.ParameterGroups[groupName]; !ok {
		return fmt.Errorf("group not found")
	}
	delete(cfg.ParameterGroups, groupName)
	return nil
}

// renameParamSlot handles rename param slot and returns the resulting value or error.
func renameParamSlot(cfg *firebase.RemoteConfig, key, nextKey, groupName string) error {
	nextKey = strings.TrimSpace(nextKey)
	if nextKey == "" {
		return fmt.Errorf("parameter name is empty")
	}
	if key == nextKey {
		return fmt.Errorf("parameter not changed")
	}
	slot, ok := lookupParamSlot(cfg, key, groupName)
	if !ok {
		return fmt.Errorf("parameter not found")
	}
	if _, exists := lookupAnyParamSlot(cfg, nextKey); exists {
		return fmt.Errorf("parameter %q already exists", nextKey)
	}
	removeParamSlot(cfg, key, groupName)
	setParamSlot(cfg, nextKey, slot)
	return nil
}

// renameGroupSlot handles rename group slot and returns the resulting value or error.
func renameGroupSlot(cfg *firebase.RemoteConfig, key, nextKey string) error {
	nextKey = strings.TrimSpace(nextKey)
	if key == "" {
		return fmt.Errorf("default group cannot be renamed")
	}
	if nextKey == "" {
		return fmt.Errorf("group name is empty")
	}
	if key == nextKey {
		return fmt.Errorf("group not changed")
	}
	group, ok := cfg.ParameterGroups[key]
	if !ok {
		return fmt.Errorf("group not found")
	}
	if _, exists := cfg.ParameterGroups[nextKey]; exists {
		return fmt.Errorf("group %q already exists", nextKey)
	}
	delete(cfg.ParameterGroups, key)
	cfg.ParameterGroups[nextKey] = group
	return nil
}

// moveParamSlot moves move param slot and returns the resulting value or error.
func moveParamSlot(cfg *firebase.RemoteConfig, key, currentGroup, nextGroup string) error {
	nextGroup = strings.TrimSpace(nextGroup)
	if currentGroup == nextGroup {
		return fmt.Errorf("parameter already in group %q", nextGroup)
	}
	slot, ok := lookupParamSlot(cfg, key, currentGroup)
	if !ok {
		return fmt.Errorf("parameter not found")
	}
	removeParamSlot(cfg, key, currentGroup)
	slot.group = nextGroup
	setParamSlot(cfg, key, slot)
	return nil
}

// applyParameterDetailsEdit handles apply parameter details edit and returns the resulting value or error.
func applyParameterDetailsEdit(cfg *firebase.RemoteConfig, edit ParameterDetailsEdit) error {
	if edit.Create {
		return createParameterDetailsSlot(cfg, edit)
	}
	return editParameterDetailsSlot(cfg, edit)
}

// createParameterDetailsSlot handles create parameter details slot and returns the resulting value or error.
func createParameterDetailsSlot(cfg *firebase.RemoteConfig, edit ParameterDetailsEdit) error {
	nextGroup := NormalizeRemoteConfigGroupKey(edit.NextGroupKey)
	nextKey := strings.TrimSpace(edit.NextParamKey)
	if nextKey == "" {
		return fmt.Errorf("parameter name is empty")
	}
	if _, exists := lookupAnyParamSlot(cfg, nextKey); exists {
		return fmt.Errorf("parameter %q already exists", nextKey)
	}
	param := firebase.RemoteConfigParam{
		Description:  strings.TrimSpace(edit.NextDescription),
		ValueType:    normalizeParameterValueType(edit.NextValueType),
		DefaultValue: &firebase.RemoteConfigValue{Value: ""},
	}
	slot := paramSlot{group: nextGroup, param: param}
	for _, valueEdit := range edit.ValueEdits {
		if err := setRawParamValue(&slot.param, valueEdit.Label, valueEdit.NextValue, slot.param.ValueType); err != nil {
			return err
		}
	}
	setParamSlot(cfg, nextKey, slot)
	return nil
}

// editParameterDetailsSlot handles edit parameter details slot and returns the resulting value or error.
func editParameterDetailsSlot(cfg *firebase.RemoteConfig, edit ParameterDetailsEdit) error {
	currentGroup := NormalizeRemoteConfigGroupKey(edit.GroupKey)
	nextGroup := NormalizeRemoteConfigGroupKey(edit.NextGroupKey)
	nextKey := strings.TrimSpace(edit.NextParamKey)
	if nextKey == "" {
		return fmt.Errorf("parameter name is empty")
	}
	slot, ok := lookupParamSlot(cfg, edit.ParamKey, currentGroup)
	if !ok {
		return fmt.Errorf("parameter not found")
	}
	if nextKey != edit.ParamKey {
		if _, exists := lookupAnyParamSlot(cfg, nextKey); exists {
			return fmt.Errorf("parameter %q already exists", nextKey)
		}
	}

	slot.param.Description = strings.TrimSpace(edit.NextDescription)
	slot.param.ValueType = normalizeParameterValueType(edit.NextValueType)
	for _, valueEdit := range edit.ValueEdits {
		if err := setRawParamValue(&slot.param, valueEdit.Label, valueEdit.NextValue, slot.param.ValueType); err != nil {
			return err
		}
	}
	slot.group = nextGroup
	removeParamSlot(cfg, edit.ParamKey, currentGroup)
	setParamSlot(cfg, nextKey, slot)
	return nil
}

// setRawParamValue sets set raw param value and returns the resulting value or error.
func setRawParamValue(param *firebase.RemoteConfigParam, valueLabel, nextValue, valueType string) error {
	if err := validateRawValueForType(nextValue, valueType); err != nil {
		return err
	}
	updateValue := func(value firebase.RemoteConfigValue) (firebase.RemoteConfigValue, error) {
		if value.UseInAppDefault || len(value.PersonalizationValue) > 0 || len(value.RolloutValue) > 0 {
			return firebase.RemoteConfigValue{}, fmt.Errorf("value editor supports only plain values")
		}
		value.Value = nextValue
		return value, nil
	}
	if valueLabel == "default" {
		if param.DefaultValue == nil {
			return fmt.Errorf("default value not found")
		}
		next, err := updateValue(*param.DefaultValue)
		if err != nil {
			return err
		}
		param.DefaultValue = &next
		return nil
	}
	if param.ConditionalValues == nil {
		return fmt.Errorf("conditional value %q not found", valueLabel)
	}
	current, ok := param.ConditionalValues[valueLabel]
	if !ok {
		return fmt.Errorf("conditional value %q not found", valueLabel)
	}
	next, err := updateValue(current)
	if err != nil {
		return err
	}
	param.ConditionalValues[valueLabel] = next
	return nil
}

// validateRawValueForType handles validate raw value for type and returns the resulting value or error.
func validateRawValueForType(value, valueType string) error {
	switch strings.ToUpper(strings.TrimSpace(valueType)) {
	case "", "STRING":
		return nil
	case "BOOLEAN":
		switch strings.TrimSpace(strings.ToLower(value)) {
		case "true", "false":
			return nil
		default:
			return fmt.Errorf("invalid boolean")
		}
	case "NUMBER":
		if !jsonNumberPattern.MatchString(strings.TrimSpace(value)) {
			return fmt.Errorf("invalid number")
		}
		return nil
	case "JSON":
		if !json.Valid([]byte(value)) {
			return fmt.Errorf("invalid json")
		}
		return nil
	default:
		return fmt.Errorf("invalid value type %q", valueType)
	}
}

// normalizeParameterValueType handles normalize parameter value type and returns the resulting value or error.
func normalizeParameterValueType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "string":
		return "STRING"
	case "boolean", "bool":
		return "BOOLEAN"
	case "number":
		return "NUMBER"
	case "json":
		return "JSON"
	default:
		return strings.TrimSpace(value)
	}
}

// moveGroupSlot moves move group slot and returns the resulting value or error.
func moveGroupSlot(cfg *firebase.RemoteConfig, currentGroup, nextGroup string) error {
	if currentGroup == nextGroup {
		return fmt.Errorf("group already moved to %q", nextGroup)
	}
	if currentGroup == "" {
		if nextGroup == "" {
			return fmt.Errorf("default group cannot be moved to default group")
		}
		destGroup := cfg.ParameterGroups[nextGroup]
		for key := range cfg.Parameters {
			if _, exists := destGroup.Parameters[key]; exists {
				return fmt.Errorf("parameter %q already exists", key)
			}
		}
		rootParams := make(map[string]firebase.RemoteConfigParam, len(cfg.Parameters))
		maps.Copy(rootParams, cfg.Parameters)
		for key, param := range rootParams {
			removeParamSlot(cfg, key, "")
			setParamSlot(cfg, key, paramSlot{group: nextGroup, param: param})
		}
		return nil
	}
	group, ok := cfg.ParameterGroups[currentGroup]
	if !ok {
		return fmt.Errorf("group not found")
	}
	for key := range group.Parameters {
		if nextGroup == "" {
			if _, exists := cfg.Parameters[key]; exists {
				return fmt.Errorf("parameter %q already exists", key)
			}
			continue
		}
		destGroup := cfg.ParameterGroups[nextGroup]
		if _, exists := destGroup.Parameters[key]; exists {
			return fmt.Errorf("parameter %q already exists", key)
		}
	}
	for key, param := range group.Parameters {
		removeParamSlot(cfg, key, currentGroup)
		setParamSlot(cfg, key, paramSlot{group: nextGroup, param: param})
	}
	delete(cfg.ParameterGroups, currentGroup)
	if nextGroup != "" {
		if group, ok := cfg.ParameterGroups[nextGroup]; ok {
			cfg.ParameterGroups[nextGroup] = group
		}
	}
	return nil
}

// setBooleanParamValueSlot sets set boolean param value slot and returns the resulting value or error.
func setBooleanParamValueSlot(cfg *firebase.RemoteConfig, key, groupName, valueLabel string, nextValue bool) error {
	slot, ok := lookupParamSlot(cfg, key, groupName)
	if !ok {
		return fmt.Errorf("parameter not found")
	}
	target := "false"
	if nextValue {
		target = "true"
	}
	updateValue := func(value firebase.RemoteConfigValue) (firebase.RemoteConfigValue, error) {
		if value.UseInAppDefault || len(value.PersonalizationValue) > 0 || len(value.RolloutValue) > 0 {
			return firebase.RemoteConfigValue{}, fmt.Errorf("boolean editor supports only plain values")
		}
		if strings.EqualFold(value.Value, target) {
			return firebase.RemoteConfigValue{}, fmt.Errorf("parameter value not changed")
		}
		value.Value = target
		return value, nil
	}

	if valueLabel == "default" {
		if slot.param.DefaultValue == nil {
			return fmt.Errorf("default value not found")
		}
		next, err := updateValue(*slot.param.DefaultValue)
		if err != nil {
			return err
		}
		slot.param.DefaultValue = &next
		setParamSlot(cfg, key, slot)
		return nil
	}

	if slot.param.ConditionalValues == nil {
		return fmt.Errorf("conditional value %q not found", valueLabel)
	}
	current, ok := slot.param.ConditionalValues[valueLabel]
	if !ok {
		return fmt.Errorf("conditional value %q not found", valueLabel)
	}
	next, err := updateValue(current)
	if err != nil {
		return err
	}
	slot.param.ConditionalValues[valueLabel] = next
	setParamSlot(cfg, key, slot)
	return nil
}

// setNumberParamValueSlot sets set number param value slot and returns the resulting value or error.
func setNumberParamValueSlot(cfg *firebase.RemoteConfig, key, groupName, valueLabel, nextValue string) error {
	nextValue = strings.TrimSpace(nextValue)
	if nextValue == "" || !jsonNumberPattern.MatchString(nextValue) {
		return fmt.Errorf("invalid number")
	}
	slot, ok := lookupParamSlot(cfg, key, groupName)
	if !ok {
		return fmt.Errorf("parameter not found")
	}
	updateValue := func(value firebase.RemoteConfigValue) (firebase.RemoteConfigValue, error) {
		if value.UseInAppDefault || len(value.PersonalizationValue) > 0 || len(value.RolloutValue) > 0 {
			return firebase.RemoteConfigValue{}, fmt.Errorf("number editor supports only plain values")
		}
		if strings.TrimSpace(value.Value) == nextValue {
			return firebase.RemoteConfigValue{}, fmt.Errorf("parameter value not changed")
		}
		value.Value = nextValue
		return value, nil
	}

	if valueLabel == "default" {
		if slot.param.DefaultValue == nil {
			return fmt.Errorf("default value not found")
		}
		next, err := updateValue(*slot.param.DefaultValue)
		if err != nil {
			return err
		}
		slot.param.DefaultValue = &next
		setParamSlot(cfg, key, slot)
		return nil
	}

	if slot.param.ConditionalValues == nil {
		return fmt.Errorf("conditional value %q not found", valueLabel)
	}
	current, ok := slot.param.ConditionalValues[valueLabel]
	if !ok {
		return fmt.Errorf("conditional value %q not found", valueLabel)
	}
	next, err := updateValue(current)
	if err != nil {
		return err
	}
	slot.param.ConditionalValues[valueLabel] = next
	setParamSlot(cfg, key, slot)
	return nil
}

// setStringParamValueSlot sets set string param value slot and returns the resulting value or error.
func setStringParamValueSlot(cfg *firebase.RemoteConfig, key, groupName, valueLabel, nextValue string) error {
	slot, ok := lookupParamSlot(cfg, key, groupName)
	if !ok {
		return fmt.Errorf("parameter not found")
	}
	updateValue := func(value firebase.RemoteConfigValue) (firebase.RemoteConfigValue, error) {
		if value.UseInAppDefault || len(value.PersonalizationValue) > 0 || len(value.RolloutValue) > 0 {
			return firebase.RemoteConfigValue{}, fmt.Errorf("string editor supports only plain values")
		}
		if value.Value == nextValue {
			return firebase.RemoteConfigValue{}, fmt.Errorf("parameter value not changed")
		}
		value.Value = nextValue
		return value, nil
	}

	if valueLabel == "default" {
		if slot.param.DefaultValue == nil {
			return fmt.Errorf("default value not found")
		}
		next, err := updateValue(*slot.param.DefaultValue)
		if err != nil {
			return err
		}
		slot.param.DefaultValue = &next
		setParamSlot(cfg, key, slot)
		return nil
	}

	if slot.param.ConditionalValues == nil {
		return fmt.Errorf("conditional value %q not found", valueLabel)
	}
	current, ok := slot.param.ConditionalValues[valueLabel]
	if !ok {
		return fmt.Errorf("conditional value %q not found", valueLabel)
	}
	next, err := updateValue(current)
	if err != nil {
		return err
	}
	slot.param.ConditionalValues[valueLabel] = next
	setParamSlot(cfg, key, slot)
	return nil
}

// setJSONParamValueSlot sets set jsonparam value slot and returns the resulting value or error.
func setJSONParamValueSlot(cfg *firebase.RemoteConfig, key, groupName, valueLabel, nextValue string) error {
	nextValue = strings.TrimSpace(nextValue)
	if !json.Valid([]byte(nextValue)) {
		return fmt.Errorf("invalid json")
	}
	var compact bytes.Buffer
	if err := json.Compact(&compact, []byte(nextValue)); err != nil {
		return fmt.Errorf("invalid json")
	}
	slot, ok := lookupParamSlot(cfg, key, groupName)
	if !ok {
		return fmt.Errorf("parameter not found")
	}
	updateValue := func(value firebase.RemoteConfigValue) (firebase.RemoteConfigValue, error) {
		if value.UseInAppDefault || len(value.PersonalizationValue) > 0 || len(value.RolloutValue) > 0 {
			return firebase.RemoteConfigValue{}, fmt.Errorf("json editor supports only plain values")
		}
		if value.Value == compact.String() {
			return firebase.RemoteConfigValue{}, fmt.Errorf("parameter value not changed")
		}
		value.Value = compact.String()
		return value, nil
	}

	if valueLabel == "default" {
		if slot.param.DefaultValue == nil {
			return fmt.Errorf("default value not found")
		}
		next, err := updateValue(*slot.param.DefaultValue)
		if err != nil {
			return err
		}
		slot.param.DefaultValue = &next
		setParamSlot(cfg, key, slot)
		return nil
	}

	if slot.param.ConditionalValues == nil {
		return fmt.Errorf("conditional value %q not found", valueLabel)
	}
	current, ok := slot.param.ConditionalValues[valueLabel]
	if !ok {
		return fmt.Errorf("conditional value %q not found", valueLabel)
	}
	next, err := updateValue(current)
	if err != nil {
		return err
	}
	slot.param.ConditionalValues[valueLabel] = next
	setParamSlot(cfg, key, slot)
	return nil
}

// duplicateParamSlot handles duplicate param slot and returns the resulting value or error.
func duplicateParamSlot(cfg *firebase.RemoteConfig, key, groupName string) (string, error) {
	slot, ok := lookupParamSlot(cfg, key, groupName)
	if !ok {
		return "", fmt.Errorf("parameter not found")
	}
	nextKey := nextDuplicateParamKey(cfg, key+"_copy")
	setParamSlot(cfg, nextKey, slot)
	return nextKey, nil
}

// duplicateParamSlotAs handles duplicate param slot as and returns the resulting value or error.
func duplicateParamSlotAs(cfg *firebase.RemoteConfig, key, nextKey, groupName string) error {
	nextKey = strings.TrimSpace(nextKey)
	if nextKey == "" {
		return fmt.Errorf("invalid name")
	}
	slot, ok := lookupParamSlot(cfg, key, groupName)
	if !ok {
		return fmt.Errorf("parameter not found")
	}
	if _, exists := lookupAnyParamSlot(cfg, nextKey); exists {
		return fmt.Errorf("parameter %q already exists", nextKey)
	}
	setParamSlot(cfg, nextKey, slot)
	return nil
}

// nextDuplicateParamKey handles next duplicate param key and returns the resulting value or error.
func nextDuplicateParamKey(cfg *firebase.RemoteConfig, base string) string {
	if _, exists := lookupAnyParamSlot(cfg, base); !exists {
		return base
	}
	for i := 2; ; i++ {
		next := fmt.Sprintf("%s__dup__%d", base, i)
		if _, exists := lookupAnyParamSlot(cfg, next); !exists {
			return next
		}
	}
}

// setParamSlot sets set param slot and returns the resulting value or error.
func setParamSlot(cfg *firebase.RemoteConfig, key string, slot paramSlot) {
	if slot.group == "" {
		if cfg.Parameters == nil {
			cfg.Parameters = map[string]firebase.RemoteConfigParam{}
		}
		cfg.Parameters[key] = slot.param
		return
	}

	if cfg.ParameterGroups == nil {
		cfg.ParameterGroups = map[string]firebase.RemoteConfigGroup{}
	}
	group := cfg.ParameterGroups[slot.group]
	if group.Parameters == nil {
		group.Parameters = map[string]firebase.RemoteConfigParam{}
	}
	group.Parameters[key] = slot.param
	cfg.ParameterGroups[slot.group] = group
}

// lookupParamSlot handles lookup param slot and returns the resulting value or error.
func lookupParamSlot(cfg *firebase.RemoteConfig, key, groupName string) (paramSlot, bool) {
	if groupName == "" {
		param, ok := cfg.Parameters[key]
		return paramSlot{group: "", param: param}, ok
	}
	group, ok := cfg.ParameterGroups[groupName]
	if !ok {
		return paramSlot{}, false
	}
	param, ok := group.Parameters[key]
	return paramSlot{group: groupName, param: param}, ok
}

// lookupAnyParamSlot handles lookup any param slot and returns the resulting value or error.
func lookupAnyParamSlot(cfg *firebase.RemoteConfig, key string) (paramSlot, bool) {
	if param, ok := cfg.Parameters[key]; ok {
		return paramSlot{group: "", param: param}, true
	}
	for groupName, group := range cfg.ParameterGroups {
		if param, ok := group.Parameters[key]; ok {
			return paramSlot{group: groupName, param: param}, true
		}
	}
	return paramSlot{}, false
}

// pruneUnusedConditions handles prune unused conditions and returns the resulting value or error.
func pruneUnusedConditions(cfg *firebase.RemoteConfig) {
	if cfg == nil || len(cfg.Conditions) == 0 {
		return
	}
	used := make(map[string]struct{})
	collectUsedConditions(used, cfg.Parameters)
	for _, group := range cfg.ParameterGroups {
		collectUsedConditions(used, group.Parameters)
	}
	kept := make([]firebase.RemoteConfigCondition, 0, len(cfg.Conditions))
	for _, condition := range cfg.Conditions {
		if _, ok := used[condition.Name]; ok {
			kept = append(kept, condition)
		}
	}
	cfg.Conditions = kept
}

// collectUsedConditions handles collect used conditions and returns the resulting value or error.
func collectUsedConditions(used map[string]struct{}, params map[string]firebase.RemoteConfigParam) {
	for _, param := range params {
		for condition := range param.ConditionalValues {
			used[condition] = struct{}{}
		}
	}
}

// dropUnknownConditionReferences handles drop unknown condition references and returns the resulting value or error.
func dropUnknownConditionReferences(cfg *firebase.RemoteConfig) {
	allowed := make(map[string]struct{}, len(cfg.Conditions))
	for _, condition := range cfg.Conditions {
		allowed[condition.Name] = struct{}{}
	}
	cfg.Parameters = stripUnknownConditionRefs(cfg.Parameters, allowed)
	for groupName, group := range cfg.ParameterGroups {
		group.Parameters = stripUnknownConditionRefs(group.Parameters, allowed)
		if len(group.Parameters) == 0 {
			delete(cfg.ParameterGroups, groupName)
			continue
		}
		cfg.ParameterGroups[groupName] = group
	}
}

// stripUnknownConditionRefs handles strip unknown condition refs and returns the resulting value or error.
func stripUnknownConditionRefs(params map[string]firebase.RemoteConfigParam, allowed map[string]struct{}) map[string]firebase.RemoteConfigParam {
	if len(params) == 0 {
		return nil
	}
	out := make(map[string]firebase.RemoteConfigParam, len(params))
	for key, param := range params {
		if len(param.ConditionalValues) > 0 {
			filtered := make(map[string]firebase.RemoteConfigValue, len(param.ConditionalValues))
			for cond, value := range param.ConditionalValues {
				if _, ok := allowed[cond]; !ok {
					continue
				}
				filtered[cond] = value
			}
			if len(filtered) > 0 {
				param.ConditionalValues = filtered
			} else {
				param.ConditionalValues = nil
			}
		}
		if param.DefaultValue == nil && len(param.ConditionalValues) == 0 {
			continue
		}
		out[key] = param
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// removeEmptyGroups removes remove empty groups and returns the resulting value or error.
func removeEmptyGroups(cfg *firebase.RemoteConfig) {
	for groupName, group := range cfg.ParameterGroups {
		if len(group.Parameters) == 0 {
			delete(cfg.ParameterGroups, groupName)
		}
	}
	if len(cfg.ParameterGroups) == 0 {
		cfg.ParameterGroups = nil
	}
	if len(cfg.Parameters) == 0 {
		cfg.Parameters = nil
	}
}
